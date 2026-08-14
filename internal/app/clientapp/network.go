package clientapp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/logging"
	d2legacy "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

type networkController struct {
	mu                            sync.Mutex
	reconnectMu                   sync.Mutex
	app                           *application
	phase, mode, address, failure string
	generation                    uint64
	host                          *gameserver.Host
	server                        *sessionquic.Server
	client                        *clientsession.Session
	cancel                        context.CancelFunc
	sequence                      uint64
	lastMovementTick              uint64
	lastMovementActive            bool
	inputLag                      time.Duration
	connectionEpoch               uint64
	submissions                   chan gameserver.CommandIntent
}

func newNetworkController(app *application) *networkController {
	return &networkController{app: app, phase: "frontend", submissions: make(chan gameserver.CommandIntent, 64)}
}

func (controller *networkController) Host() error {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.phase != "frontend" && controller.phase != "failed" {
		return controller.rejectLocked("host", errors.New("network operation already active"))
	}
	controller.phase, controller.mode, controller.address, controller.failure = "selecting", "host", "", ""
	slog.Debug("network host requested; awaiting character selection")
	return nil
}

func (controller *networkController) StartSelected() error {
	controller.mu.Lock()
	if controller.phase != "frontend" && (controller.phase != "selecting" || (controller.mode != "host" && controller.mode != "join")) {
		defer controller.mu.Unlock()
		return controller.rejectLocked(controller.mode, errors.New("no network operation is awaiting character selection"))
	}
	character, selected := controller.app.saves.Selected()
	if !selected {
		defer controller.mu.Unlock()
		return controller.rejectLocked(controller.mode, errors.New("select a character before continuing"))
	}
	if controller.phase == "frontend" {
		controller.phase, controller.mode, controller.failure = "local", "local", ""
		controller.mu.Unlock()
		slog.Debug("local game session activated", "character_id", character.ID,
			"character_name", character.Name, "character_class", character.Class)
		return nil
	}
	controller.generation++
	generation := controller.generation
	ctx, cancel := context.WithCancel(controller.app.ctx)
	controller.cancel = cancel
	controller.phase, controller.failure = "starting", ""
	mode, address := controller.mode, controller.address
	slog.Debug("network operation starting", "mode", controller.mode, "address", controller.address,
		"character_id", character.ID, "character_name", character.Name, "character_class", character.Class)
	controller.mu.Unlock()
	if mode == "host" {
		go controller.startHost(ctx, generation)
	} else {
		go controller.startJoin(ctx, generation, address)
	}
	return nil
}

func (controller *networkController) Cancel() {
	controller.mu.Lock()
	if controller.phase == "selecting" || controller.phase == "starting" || controller.phase == "failed" {
		controller.generation++
		cancel := controller.cancel
		controller.cancel = nil
		controller.phase, controller.mode, controller.address, controller.failure = "frontend", "", "", ""
		controller.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		return
	}
	controller.mu.Unlock()
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

func (controller *networkController) startJoin(ctx context.Context, generation uint64, address string) {
	slog.Debug("dialing self-hosted game", "address", address)
	recomposed := false
	fail := func(err error) {
		if recomposed {
			restoreContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			restoreErr := controller.app.restoreConfiguredPackages(restoreContext)
			cancel()
			err = errors.Join(err, restoreErr)
		}
		controller.fail(generation, err)
	}
	d2legacySource, err := controller.app.modSource("d2legacy")
	if err != nil {
		fail(err)
		return
	}
	identity, err := d2legacy.IdentityForPackages(d2legacySource, controller.app.options.Packages, controller.app.sessionInitialData())
	if err != nil {
		fail(err)
		return
	}
	clientTLS, err := controller.app.networkTrust.ClientTLS(address)
	if err != nil {
		fail(err)
		return
	}
	store, err := controller.app.ensureModCache()
	if err != nil {
		fail(err)
		return
	}
	assignment := clientsession.SelfHostedAssignment{
		GameID: "listen-local", Endpoint: realm.GameEndpoint{Address: address}, Runtime: identity,
	}
	recipe, err := clientsession.PrepareSelfHostedExtensions(ctx, assignment, clientTLS, store, controller.app.options.Packages.Base)
	if err != nil {
		fail(err)
		return
	}
	// Package acquisition verifies exact extension archives without mutating the
	// live VFS. Reconstruct the deterministic recipe from the local built-in and
	// verified descriptors before any selected character is offered to the host.
	identity, err = d2legacy.IdentityForPackages(d2legacySource, recipe.Packages, controller.app.sessionInitialData())
	if err != nil {
		fail(err)
		return
	}
	if err := sameRuntimeRecipe(identity, recipe); err != nil {
		fail(err)
		return
	}
	// Recomposition can fail after stopping an old component or swapping a VFS
	// layer. Mark the attempt first so the failure path always restores the
	// configured startup recipe with a fresh cleanup context.
	recomposed = true
	if err := controller.app.recomposeForNetworkRecipe(ctx, recipe); err != nil {
		fail(err)
		return
	}
	d2legacySource, err = controller.app.modSource("d2legacy")
	if err != nil {
		fail(err)
		return
	}
	identity, err = d2legacy.IdentityForPackages(d2legacySource, controller.app.options.Packages, controller.app.sessionInitialData())
	if err != nil {
		fail(err)
		return
	}
	if err := sameRuntimeRecipe(identity, recipe); err != nil {
		fail(err)
		return
	}
	client, err := clientsession.ConnectSelfHosted(ctx, clientsession.SelfHostedAssignment{
		GameID: "listen-local", Endpoint: realm.GameEndpoint{Address: address}, Runtime: identity,
	}, clientTLS, controller.app.saves)
	if err != nil {
		fail(err)
		return
	}
	if err := controller.app.prepareConnectedWorld(ctx); err != nil {
		_ = client.Close(context.Background())
		fail(err)
		return
	}
	controller.mu.Lock()
	if controller.generation != generation || controller.phase != "starting" {
		controller.mu.Unlock()
		_ = client.Close(context.Background())
		restoreContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = controller.app.restoreConfiguredPackages(restoreContext)
		cancel()
		return
	}
	controller.client, controller.phase = client, "connected"
	controller.connectionEpoch++
	controller.resetInputLocked()
	controller.mu.Unlock()
	hud, _ := client.View()
	slog.Debug("joined self-hosted game", "address", address, "player_id", hud.Player.PlayerID, "character_id", hud.Player.CharacterID)
	go controller.send(ctx, client)
	go controller.watch(ctx, client)
}

func sameRuntimeRecipe(identity simulation.RuntimeIdentity, recipe simulation.RuntimeRecipe) error {
	localDigest, err := identity.Digest()
	if err != nil {
		return err
	}
	serverDigest, err := (simulation.RuntimeIdentity{Recipe: recipe}).Digest()
	if err != nil {
		return err
	}
	if localDigest != serverDigest {
		return errors.New("network recipe differs from the locally composed deterministic runtime")
	}
	return nil
}

func (controller *networkController) fail(generation uint64, err error) {
	controller.mu.Lock()
	if generation != controller.generation || controller.phase == "closed" || controller.phase == "failed" {
		controller.mu.Unlock()
		return
	}
	cancel, client, server, host := controller.cancel, controller.client, controller.server, controller.host
	mode, address := controller.mode, controller.address
	controller.cancel, controller.client, controller.server, controller.host = nil, nil, nil, nil
	controller.phase, controller.failure = "failed", err.Error()
	controller.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	cleanupContext, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if client != nil {
		_ = client.Close(cleanupContext)
	}
	if server != nil {
		_ = server.Close()
	}
	if host != nil {
		_ = host.Close(cleanupContext)
	}
	slog.Debug("network operation failed", "mode", mode, "address", address, "error", err)
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

func (controller *networkController) startHost(ctx context.Context, generation uint64) {
	slog.Debug("starting listen server", "address", ":6112")
	fail := func(err error) {
		controller.fail(generation, err)
	}
	d2legacySource, err := controller.app.modSource("d2legacy")
	if err != nil {
		fail(err)
		return
	}
	host, err := gameserver.Start(ctx, d2legacySource, recordstore.New(controller.app.options.Content), gameserver.Config{
		Mode: gameserver.ModeListen, SessionID: "listen-local", Prediction: gamesession.PredictionLimited,
		InitialData: controller.app.sessionInitialData(), Packages: controller.app.options.Packages,
		Content: controller.app.options.Content, Mods: controller.app.options.Mods,
	})
	if err != nil {
		fail(err)
		return
	}
	slog.Debug("listen authority started")
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
	departures := &playeradapter.DepartureQueue{}
	endpoint.SetLeave(func(principal gameserver.Principal) error {
		return departures.Submit(host.Session, principal.PlayerID)
	})
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
	packages, err := serverapp.NewPackageProvider(host.Allocation.Identity.Recipe, controller.app.options.ModCache)
	if err != nil {
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	server.SetPackageProvider(packages)
	slog.Debug("listen transport bound", "address", server.Addr())
	for levelID, collision := range controller.app.gameWorlds {
		if err := modruntime.SetWorldMapForLevel(ctx, host.Authority.Runtime,
			"d2legacy.gameplay.systems.init", "set_collision", levelID, collision); err != nil {
			_ = server.Close()
			_ = host.Close(context.Background())
			fail(err)
			return
		}
	}
	population, err := controller.app.populationBootstrapCommand()
	if err == nil {
		err = host.Session.Submit(population)
	}
	if err != nil {
		_ = server.Close()
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
	slog.Debug("listen authority worlds installed", "levels", len(controller.app.gameWorlds))
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
	slog.Debug("listen profile admission configured")
	go func() {
		if runErr := host.Session.Run(ctx); runErr != nil && ctx.Err() == nil {
			controller.fail(generation, runErr)
		}
	}()
	go func() {
		if serveErr := server.Serve(ctx); serveErr != nil && ctx.Err() == nil {
			controller.fail(generation, serveErr)
		}
	}()
	slog.Debug("connecting host player to listen authority")
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
	if err := controller.app.activateNetworkClientComponents(ctx); err != nil {
		_ = client.Close(context.Background())
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	if err := controller.app.prepareConnectedWorld(ctx); err != nil {
		_ = client.Close(context.Background())
		_ = server.Close()
		_ = host.Close(context.Background())
		fail(err)
		return
	}
	controller.mu.Lock()
	if controller.generation != generation || controller.phase != "starting" {
		controller.mu.Unlock()
		_ = client.Close(context.Background())
		_ = server.Close()
		_ = host.Close(context.Background())
		return
	}
	controller.host, controller.server, controller.client = host, server, client
	controller.phase, controller.address = "connected", server.Addr()
	controller.connectionEpoch++
	controller.resetInputLocked()
	controller.mu.Unlock()
	hud, _ := client.View()
	slog.Debug("listen server connected local player", "address", server.Addr(), "player_id", hud.Player.PlayerID, "character_id", hud.Player.CharacterID)
	go controller.send(ctx, client)
	go controller.watch(ctx, client)
}

func (controller *networkController) send(ctx context.Context, client *clientsession.Session) {
	const maximumInFlight = 8
	inFlight := make(chan struct{}, maximumInFlight)
	var submissions sync.WaitGroup
	defer submissions.Wait()
	for {
		select {
		case intent := <-controller.submissions:
			select {
			case inFlight <- struct{}{}:
			case <-ctx.Done():
				return
			}
			submissions.Add(1)
			go func(intent gameserver.CommandIntent) {
				defer submissions.Done()
				defer func() { <-inFlight }()
				controller.sendIntent(ctx, client, intent)
			}(intent)
		case <-ctx.Done():
			return
		}
	}
}

func (controller *networkController) sendIntent(ctx context.Context, client *clientsession.Session, intent gameserver.CommandIntent) {
	logging.Trace(slog.Default(), "sending network command", "sequence", intent.Sequence, "kind", intent.Kind)
	for {
		epoch := controller.currentConnectionEpoch()
		if err := client.Submit(ctx, intent); err == nil {
			return
		} else if ctx.Err() != nil {
			return
		} else if isRemoteProtocolError(err) {
			client.DiscardInput(intent.Sequence)
			slog.Debug("network command rejected", "sequence", intent.Sequence, "kind", intent.Kind, "error", err)
			return
		} else if recoverErr := controller.recover(ctx, client, epoch, err); recoverErr != nil {
			controller.fail(controller.currentGeneration(), recoverErr)
			return
		}
	}
}

func (controller *networkController) watch(ctx context.Context, client *clientsession.Session) {
	for ctx.Err() == nil {
		epoch := controller.currentConnectionEpoch()
		deltas, failures, err := client.Watch(ctx)
		if err != nil {
			if isRemoteProtocolError(err) {
				controller.fail(controller.currentGeneration(), err)
				return
			}
			if recoverErr := controller.recover(ctx, client, epoch, err); recoverErr != nil && ctx.Err() == nil {
				controller.fail(controller.currentGeneration(), recoverErr)
			}
			continue
		}
		var streamErr error
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
					streamErr = err
					deltas, failures = nil, nil
				}
			case <-ctx.Done():
				return
			}
		}
		if ctx.Err() != nil {
			return
		}
		if streamErr == nil {
			streamErr = errors.New("network correction stream ended")
		}
		if isRemoteProtocolError(streamErr) {
			controller.fail(controller.currentGeneration(), streamErr)
			return
		}
		if err := controller.recover(ctx, client, epoch, streamErr); err != nil {
			if ctx.Err() == nil {
				controller.fail(controller.currentGeneration(), err)
			}
			return
		}
	}
}

func isRemoteProtocolError(err error) bool {
	var remote *sessionquic.RemoteError
	return errors.As(err, &remote)
}

func (controller *networkController) recover(ctx context.Context, client *clientsession.Session, observedEpoch uint64, cause error) error {
	controller.reconnectMu.Lock()
	defer controller.reconnectMu.Unlock()
	controller.mu.Lock()
	if controller.phase == "closed" || controller.client != client {
		controller.mu.Unlock()
		return context.Canceled
	}
	if controller.connectionEpoch != observedEpoch && controller.phase == "connected" {
		controller.mu.Unlock()
		return nil
	}
	controller.phase = "reconnecting"
	mode, address := controller.mode, controller.address
	controller.mu.Unlock()
	slog.Debug("network connection interrupted; reconnecting", "mode", mode, "address", address, "error", cause)

	recoveryContext, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	delay := time.Duration(0)
	var lastErr error = cause
	for attempt := 1; recoveryContext.Err() == nil; attempt++ {
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-recoveryContext.Done():
				timer.Stop()
				break
			case <-timer.C:
			}
		}
		attemptContext, stop := context.WithTimeout(recoveryContext, 2*time.Second)
		lastErr = client.Reconnect(attemptContext)
		stop()
		if lastErr == nil {
			controller.mu.Lock()
			if controller.client != client || controller.phase == "closed" {
				controller.mu.Unlock()
				return context.Canceled
			}
			controller.connectionEpoch++
			controller.phase, controller.failure = "connected", ""
			controller.mu.Unlock()
			slog.Debug("network session reconnected", "attempt", attempt, "mode", mode, "address", address)
			return nil
		}
		delay = min(1600*time.Millisecond, max(200*time.Millisecond, delay*2))
	}
	return fmt.Errorf("network reconnect lease expired: %w", lastErr)
}

func (controller *networkController) Advance(ctx context.Context, elapsed time.Duration) error {
	controller.mu.Lock()
	client := controller.client
	controller.mu.Unlock()
	if client == nil {
		return nil
	}
	// Install the newest correction before sampling input. Point-and-click path
	// selection and camera-relative pointer projection then start from the latest
	// authenticated local position instead of the frozen frontend authority.
	if err := controller.app.clientWorld.reconcile(controller.app, client, elapsed); err != nil {
		return err
	}
	now := time.Now()
	for _, targetTick := range controller.inputTicks(client, elapsed, now) {
		if controller.app.movementSource == nil {
			continue
		}
		for _, command := range controller.app.movementSource.Commands(targetTick) {
			var movementPayload movement.MovePayload
			if command.Kind == movement.MoveCommand && json.Unmarshal(command.Payload, &movementPayload) == nil {
				active := movementPayload.X != 0 || movementPayload.Y != 0 || movementPayload.Target != nil
				if !controller.movementRequired(active) {
					continue
				}
				if err := controller.submit(targetTick, command.Kind, command.Payload); err != nil {
					return err
				}
				controller.markMovement(active)
				continue
			}
			if err := controller.submit(targetTick, command.Kind, command.Payload); err != nil {
				return err
			}
		}
	}
	return controller.submitPendingIntents(client, now)
}

func (controller *networkController) submitPendingIntents(client *clientsession.Session, now time.Time) error {
	intents := controller.app.commandIntents.Drain()
	if len(intents) == 0 {
		return nil
	}
	targetTick := client.NextInputTick(now)
	for _, intent := range intents {
		payload, err := json.Marshal(intent.Payload)
		if err != nil {
			return err
		}
		if err := controller.submit(targetTick, intent.Kind, payload); err != nil {
			return err
		}
	}
	return nil
}

const networkInputStep = 40 * time.Millisecond

func (controller *networkController) inputTicks(client *clientsession.Session, elapsed time.Duration, now time.Time) []uint64 {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if elapsed < 0 {
		elapsed = 0
	}
	// The authority uses the same bounded catch-up policy. A renderer hitch may
	// produce several fixed input samples, but never an unbounded burst that can
	// overflow home-network queues or the server's admission window.
	controller.inputLag = min(controller.inputLag+elapsed, 5*networkInputStep)
	result := make([]uint64, 0, 5)
	for controller.inputLag >= networkInputStep {
		target := client.NextInputTick(now)
		if target > controller.lastMovementTick {
			controller.lastMovementTick = target
			result = append(result, target)
		}
		controller.inputLag -= networkInputStep
	}
	return result
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

func (controller *networkController) submit(targetTick uint64, kind string, payload json.RawMessage) error {
	controller.mu.Lock()
	intent := gameserver.CommandIntent{TargetTick: targetTick, Sequence: controller.sequence + 1, Kind: kind, Payload: append(json.RawMessage(nil), payload...)}
	client := controller.client
	if client == nil {
		controller.mu.Unlock()
		return errors.New("network client is unavailable")
	}
	if err := client.StageInput(intent); err != nil {
		controller.mu.Unlock()
		return err
	}
	select {
	case controller.submissions <- intent:
		controller.sequence++
		controller.mu.Unlock()
		return nil
	default:
		client.DiscardInput(intent.Sequence)
		controller.mu.Unlock()
		return errors.New("network input queue is full")
	}
}

func (controller *networkController) Connected() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.phase == "connected" && controller.client != nil
}

func (controller *networkController) Local() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.phase == "local"
}

func (controller *networkController) currentGeneration() uint64 {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.generation
}

func (controller *networkController) currentConnectionEpoch() uint64 {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.connectionEpoch
}

func (controller *networkController) resetInputLocked() {
	controller.sequence = 0
	controller.lastMovementTick = 0
	controller.lastMovementActive = false
	controller.inputLag = 0
	controller.submissions = make(chan gameserver.CommandIntent, 64)
}

func (controller *networkController) Close() error {
	slog.Debug("closing network controller")
	controller.mu.Lock()
	cancel, client, server, host, mode := controller.cancel, controller.client, controller.server, controller.host, controller.mode
	controller.cancel, controller.client, controller.server, controller.host = nil, nil, nil, nil
	controller.phase = "closed"
	controller.mu.Unlock()
	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	var err error
	if client != nil {
		if mode == "host" || mode == "join" {
			err = errors.Join(err, controller.persistSelfHostedCharacter(ctx, client))
		}
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

func (controller *networkController) persistSelfHostedCharacter(ctx context.Context, client *clientsession.Session) error {
	if controller.app == nil || controller.app.saves == nil || client == nil {
		return nil
	}
	if _, err := client.Refresh(ctx); err != nil {
		return fmt.Errorf("refresh selected self-hosted character: %w", err)
	}
	hud, _ := client.View()
	return updateSelectedCharacter(controller.app.saves, hud)
}

func updateSelectedCharacter(saves *d2save.Store, hud playeradapter.HUD) error {
	baseline, selected := saves.Selected()
	if !selected {
		return errors.New("persist self-hosted character: no selected character")
	}
	updated, err := playeradapter.MergeCharacter(baseline, hud)
	if err != nil {
		return fmt.Errorf("persist self-hosted character: %w", err)
	}
	if err := saves.UpdateSelected(updated); err != nil {
		return fmt.Errorf("persist self-hosted character: %w", err)
	}
	return nil
}

func randomBytes(size int) ([]byte, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return nil, err
	}
	return value, nil
}
