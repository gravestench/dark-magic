package clientapp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	d2legacy "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type networkController struct {
	mu                            sync.Mutex
	app                           *application
	phase, mode, address, failure string
	host                          *gameserver.Host
	server                        *sessionquic.Server
	client                        *clientsession.Session
	cancel                        context.CancelFunc
	sequence                      uint64
}

func newNetworkController(app *application) *networkController {
	return &networkController{app: app, phase: "idle"}
}

func (controller *networkController) Host() error {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.phase != "idle" && controller.phase != "failed" {
		return controller.rejectLocked("host", errors.New("network operation already active"))
	}
	controller.phase, controller.mode, controller.address, controller.failure = "selecting", "host", "", ""
	return nil
}

func (controller *networkController) StartSelected() error {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.phase != "selecting" || (controller.mode != "host" && controller.mode != "join") {
		return controller.rejectLocked(controller.mode, errors.New("no network operation is awaiting character selection"))
	}
	if _, selected := controller.app.saves.Selected(); !selected {
		return controller.rejectLocked(controller.mode, errors.New("select a character before continuing"))
	}
	controller.phase, controller.failure = "starting", ""
	if controller.mode == "host" {
		go controller.startHost()
	} else {
		go controller.startJoin(controller.address)
	}
	return nil
}

func (controller *networkController) Cancel() {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.phase == "selecting" || controller.phase == "failed" {
		controller.phase, controller.mode, controller.address, controller.failure = "idle", "", "", ""
	}
}

func (controller *networkController) Join(address string) error {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	address = strings.TrimSpace(address)
	if address == "" {
		return controller.rejectLocked("join", errors.New("server address is required"))
	}
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = net.JoinHostPort(address, "6112")
	}
	controller.phase, controller.mode, controller.address, controller.failure = "selecting", "join", address, ""
	return nil
}

func (controller *networkController) startJoin(address string) {
	ctx, cancel := context.WithCancel(controller.app.ctx)
	identity, err := d2legacy.Identity(controller.app.options.Content)
	if err != nil {
		controller.fail(err)
		return
	}
	clientTLS, err := controller.app.networkTrust.ClientTLS(address)
	if err != nil {
		controller.fail(err)
		return
	}
	client, err := clientsession.ConnectSelfHosted(ctx, clientsession.SelfHostedAssignment{
		GameID: "listen-local", Endpoint: realm.GameEndpoint{Address: address}, Runtime: identity,
	}, clientTLS, controller.app.saves)
	if err != nil {
		cancel()
		controller.fail(err)
		return
	}
	controller.mu.Lock()
	controller.client, controller.cancel, controller.phase = client, cancel, "connected"
	controller.mu.Unlock()
	controller.watch(ctx, client)
}

func (controller *networkController) fail(err error) {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	controller.phase, controller.failure = "failed", err.Error()
}

func (controller *networkController) rejectLocked(mode string, err error) error {
	controller.phase, controller.mode, controller.failure = "failed", mode, err.Error()
	return err
}

func (controller *networkController) Status() map[string]any {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return map[string]any{"phase": controller.phase, "mode": controller.mode, "address": controller.address, "error": controller.failure}
}

func (controller *networkController) startHost() {
	ctx, cancel := context.WithCancel(controller.app.ctx)
	fail := func(err error) {
		cancel()
		controller.mu.Lock()
		controller.phase, controller.failure = "failed", err.Error()
		controller.mu.Unlock()
	}
	host, err := gameserver.Start(ctx, controller.app.options.Content, recordstore.New(controller.app.options.Content), gameserver.Config{
		Mode: gameserver.ModeListen, SessionID: "listen-local", Prediction: gamesession.PredictionLimited,
	})
	if err != nil {
		fail(err)
		return
	}
	secret, err := randomBytes(32)
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	tickets, err := gameserver.NewTicketAuthority(secret, "listen-local")
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	endpoint, err := gameserver.NewEndpoint(host, tickets, playeradapter.ProjectClientView)
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	endpoint.SetSnapshotPending(func(err error) bool { return errors.Is(err, playeradapter.ErrHUDPlayer) })
	serverTLS, clientTLS, fingerprint, err := controller.app.networkTrust.HostTLS()
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	server, err := sessionquic.Listen(":6112", serverTLS, endpoint)
	if err != nil {
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	level := controller.app.activeWorldLevel
	world, zone := controller.app.gameWorlds[level], controller.app.gameWorldZones[level]
	spawn, found := controller.app.gameWorldSpawns[level]
	if world == nil || zone == nil || !found {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(errors.New("active world has no trusted host destination"))
		return
	}
	request := zone.Request()
	destination, err := playeradapter.NewDestination(spawn[0], spawn[1], float64(world.WidthSubtiles), float64(world.HeightSubtiles), int64(request.Act), int64(request.LevelID))
	if err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	credentialBytes, err := randomBytes(32)
	if err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	profileCredential := hex.EncodeToString(credentialBytes)
	profiles, err := serverapp.NewRemoteProfileAdmissions(host, tickets, serverapp.RemoteProfileConfig{
		Credential: profileCredential, AllowDirect: true, PrincipalID: "listen-local-user", PlayerID: "player", Destination: destination, Lifetime: time.Minute,
	})
	if err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	server.SetProfileAdmissions(profiles)
	go host.Session.Run(ctx)
	go server.Serve(ctx)
	client, err := clientsession.ConnectSelfHosted(ctx, clientsession.SelfHostedAssignment{
		GameID: "listen-local", Endpoint: realm.GameEndpoint{Address: "127.0.0.1:6112", TLSFingerprint: fingerprint},
		Runtime: host.Authority.Identity, ProfileCredential: profileCredential,
	}, clientTLS, controller.app.saves)
	if err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	controller.mu.Lock()
	controller.host, controller.server, controller.client, controller.cancel = host, server, client, cancel
	controller.phase, controller.address = "connected", server.Addr()
	controller.mu.Unlock()
	controller.watch(ctx, client)
}

func (controller *networkController) watch(ctx context.Context, client *clientsession.Session) {
	deltas, failures, err := client.Watch(ctx)
	if err != nil {
		controller.fail(err)
		return
	}
	go func() {
		for deltas != nil || failures != nil {
			select {
			case _, open := <-deltas:
				if !open {
					deltas = nil
				}
			case err, open := <-failures:
				if !open {
					failures = nil
					continue
				}
				if err != nil && ctx.Err() == nil {
					controller.fail(err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (controller *networkController) Advance(ctx context.Context) error {
	controller.mu.Lock()
	client := controller.client
	controller.mu.Unlock()
	if client == nil {
		return nil
	}
	_, world := client.View()
	if controller.app.movementSource != nil {
		for _, command := range controller.app.movementSource.Commands(world.Tick + 2) {
			if err := controller.submit(ctx, client, command.Tick, command.Kind, command.Payload); err != nil {
				return err
			}
		}
	}
	for _, intent := range controller.app.commandIntents.Drain() {
		payload, err := json.Marshal(intent.Payload)
		if err != nil {
			return err
		}
		if err := controller.submit(ctx, client, world.Tick+2, intent.Kind, payload); err != nil {
			return err
		}
	}
	return controller.app.installRemoteView(client)
}

func (controller *networkController) submit(ctx context.Context, client *clientsession.Session, tick uint64, kind string, payload json.RawMessage) error {
	controller.mu.Lock()
	controller.sequence++
	sequence := controller.sequence
	controller.mu.Unlock()
	return client.Submit(ctx, gameserver.CommandIntent{Tick: tick, Sequence: sequence, Kind: kind, Payload: payload})
}

func (controller *networkController) Connected() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.phase == "connected" && controller.client != nil
}

func (controller *networkController) Close() error {
	controller.mu.Lock()
	cancel, client, server, host := controller.cancel, controller.client, controller.server, controller.host
	controller.cancel, controller.client, controller.server, controller.host = nil, nil, nil, nil
	controller.phase = "closed"
	controller.mu.Unlock()
	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	var err error
	if client != nil {
		err = errors.Join(err, client.Close(ctx))
	}
	if cancel != nil {
		cancel()
	}
	if server != nil {
		err = errors.Join(err, server.Close())
	}
	if host != nil {
		err = errors.Join(err, host.Close(ctx))
	}
	return err
}

func randomBytes(size int) ([]byte, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return nil, err
	}
	return value, nil
}
