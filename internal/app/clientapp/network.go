package clientapp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
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
	"github.com/gravestench/dark-magic/internal/logging"
	d2legacy "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
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
	lastMovementTick              uint64
	lastMovementActive            bool
	submissions                   chan gameserver.CommandIntent
}

func newNetworkController(app *application) *networkController {
	return &networkController{app: app, phase: "idle", submissions: make(chan gameserver.CommandIntent, 64)}
}

func (controller *networkController) Host() error {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.phase != "idle" && controller.phase != "failed" {
		return controller.rejectLocked("host", errors.New("network operation already active"))
	}
	controller.phase, controller.mode, controller.address, controller.failure = "selecting", "host", "", ""
	slog.Debug("network host requested; awaiting character selection")
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
	slog.Debug("network operation starting", "mode", controller.mode, "address", controller.address)
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
	slog.Debug("network join requested; awaiting character selection", "address", address)
	return nil
}

func (controller *networkController) startJoin(address string) {
	slog.Debug("dialing self-hosted game", "address", address)
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
	hud, _ := client.View()
	slog.Debug("joined self-hosted game", "address", address, "player_id", hud.Player.PlayerID, "character_id", hud.Player.CharacterID)
	go controller.send(ctx, client)
	controller.watch(ctx, client)
}

func (controller *networkController) fail(err error) {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	controller.phase, controller.failure = "failed", err.Error()
	slog.Debug("network operation failed", "mode", controller.mode, "address", controller.address, "error", err)
}

func (controller *networkController) rejectLocked(mode string, err error) error {
	controller.phase, controller.mode, controller.failure = "failed", mode, err.Error()
	return err
}

func (controller *networkController) Status() map[string]any {
	controller.mu.Lock()
	phase, mode, address, failure, client := controller.phase, controller.mode, controller.address, controller.failure, controller.client
	controller.mu.Unlock()
	playerID := "local-player"
	if client != nil {
		hud, _ := client.View()
		if hud.Player.PlayerID != "" {
			playerID = hud.Player.PlayerID
		}
	}
	return map[string]any{"phase": phase, "mode": mode, "address": address, "error": failure, "player_id": playerID}
}

func (controller *networkController) startHost() {
	slog.Debug("starting listen server", "address", ":6112")
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
	if err := modruntime.SetWorldMap(ctx, host.Authority.Runtime, "d2legacy.gameplay.systems.init", "set_collision", world); err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
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
	hud, _ := client.View()
	slog.Debug("listen server connected local player", "address", server.Addr(), "player_id", hud.Player.PlayerID, "character_id", hud.Player.CharacterID)
	go controller.send(ctx, client)
	controller.watch(ctx, client)
}

func (controller *networkController) send(ctx context.Context, client *clientsession.Session) {
	for {
		select {
		case intent := <-controller.submissions:
			logging.Trace(slog.Default(), "sending network command", "sequence", intent.Sequence, "kind", intent.Kind)
			if err := client.Submit(ctx, intent); err != nil {
				if ctx.Err() == nil {
					controller.fail(err)
				}
				return
			}
		case <-ctx.Done():
			return
		}
	}
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
			case delta, open := <-deltas:
				if !open {
					deltas = nil
				}
				logging.Trace(slog.Default(), "received network correction", "tick", delta.Tick, "upserts", len(delta.Upserts), "removed", len(delta.Removed))
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
	targetTick := world.Tick + 2
	sampleMovement := controller.sampleMovement(targetTick)
	if sampleMovement && controller.app.movementSource != nil {
		for _, command := range controller.app.movementSource.Commands(targetTick) {
			var movementPayload movement.MovePayload
			if command.Kind == movement.MoveCommand && json.Unmarshal(command.Payload, &movementPayload) == nil {
				active := movementPayload.X != 0 || movementPayload.Y != 0 || movementPayload.Target != nil
				if !controller.movementRequired(active) {
					continue
				}
				if err := controller.submit(command.Kind, command.Payload); err != nil {
					return err
				}
				controller.markMovement(active)
				continue
			}
			if err := controller.submit(command.Kind, command.Payload); err != nil {
				return err
			}
		}
	}
	for _, intent := range controller.app.commandIntents.Drain() {
		payload, err := json.Marshal(intent.Payload)
		if err != nil {
			return err
		}
		if err := controller.submit(intent.Kind, payload); err != nil {
			return err
		}
	}
	return controller.app.installRemoteView(client)
}

func (controller *networkController) sampleMovement(targetTick uint64) bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if targetTick <= controller.lastMovementTick {
		return false
	}
	controller.lastMovementTick = targetTick
	return true
}

func (controller *networkController) movementRequired(active bool) bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return active || controller.lastMovementActive
}

func (controller *networkController) markMovement(active bool) {
	controller.mu.Lock()
	controller.lastMovementActive = active
	controller.mu.Unlock()
}

func (controller *networkController) submit(kind string, payload json.RawMessage) error {
	controller.mu.Lock()
	intent := gameserver.CommandIntent{Sequence: controller.sequence + 1, Kind: kind, Payload: append(json.RawMessage(nil), payload...)}
	select {
	case controller.submissions <- intent:
		controller.sequence++
		controller.mu.Unlock()
		return nil
	default:
		controller.mu.Unlock()
		return errors.New("network input queue is full")
	}
}

func (controller *networkController) Connected() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.phase == "connected" && controller.client != nil
}

func (controller *networkController) Close() error {
	slog.Debug("closing network controller")
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
