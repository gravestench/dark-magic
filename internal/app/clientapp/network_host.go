package clientapp

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

const (
	listenGameID        = "listen-local"
	listenBindAddress   = ":6112"
	listenDialAddress   = "127.0.0.1:6112"
	listenPrincipalID   = "listen-local-user"
	listenPlayerID      = "player"
	listenCredentialTTL = time.Minute
)

// listenHostRuntime is the temporary owner of a listen server until commitListenHost succeeds.
// Keeping partial construction outside networkController prevents half-ready resources from being
// visible as a connected session.
type listenHostRuntime struct {
	host              *gameserver.Host
	server            *sessionquic.Server
	client            *clientsession.Session
	tickets           *gameserver.TicketAuthority
	clientTLS         *tls.Config
	fingerprint       string
	profileCredential string
}

// close releases any subset produced by failed preparation. The client and transport close before
// authority so their shutdown paths can still make final calls while authority exists.
func (runtime *listenHostRuntime) close() {
	if runtime.client != nil {
		_ = runtime.client.Close(context.Background())
	}

	if runtime.server != nil {
		_ = runtime.server.Close()
	}

	if runtime.host != nil {
		_ = runtime.host.Close(context.Background())
	}
}

// startHost builds the authority, exposes it through the normal authenticated protocol, and joins
// the selected local player as a real client. This keeps hosted behavior aligned with remote play
// instead of adding an in-process gameplay shortcut.
func (controller *networkController) startHost(ctx context.Context, generation uint64) {
	slog.Debug("starting listen server", "address", listenBindAddress)

	runtime := &listenHostRuntime{}
	if err := controller.prepareListenHost(ctx, runtime); err != nil {
		runtime.close()
		controller.fail(generation, err)

		return
	}

	controller.startListenServices(ctx, generation, runtime)

	if err := controller.connectListenPlayer(ctx, runtime); err != nil {
		runtime.close()
		controller.fail(generation, err)

		return
	}

	if !controller.commitListenHost(generation, runtime) {
		runtime.close()

		return
	}

	controller.logListenPlayer(runtime)
	controller.startClientLoops(ctx, runtime.client)
}

// prepareListenHost completes deterministic authority, transport, world, and profile setup before
// service goroutines run. No request can therefore observe a partially populated session.
func (controller *networkController) prepareListenHost(
	ctx context.Context,
	runtime *listenHostRuntime,
) error {
	if err := controller.startListenAuthority(ctx, runtime); err != nil {
		return err
	}

	if err := controller.bindListenTransport(runtime); err != nil {
		return err
	}

	if err := controller.installListenWorlds(ctx, runtime.host); err != nil {
		return err
	}

	if err := controller.configureListenProfile(runtime); err != nil {
		return err
	}

	return nil
}

// startListenAuthority creates the sole gameplay authority for hosted play. The client renderer
// receives projections from this host and must not mutate its simulation directly.
func (controller *networkController) startListenAuthority(
	ctx context.Context,
	runtime *listenHostRuntime,
) error {
	d2legacySource, err := controller.app.modSource("d2legacy")
	if err != nil {
		return err
	}

	runtime.host, err = gameserver.Start(
		ctx,
		d2legacySource,
		recordstore.New(controller.app.options.Content),
		gameserver.Config{
			Mode:        gameserver.ModeListen,
			SessionID:   listenGameID,
			Prediction:  gamesession.PredictionLimited,
			InitialData: controller.app.sessionInitialData(),
			Packages:    controller.app.options.Packages,
			Content:     controller.app.options.Content,
			Mods:        controller.app.options.Mods,
			AssetSetID:  controller.app.options.AssetSetID,
		},
	)
	if err != nil {
		return err
	}

	slog.Debug("listen authority started")

	return nil
}

// bindListenTransport places the in-process authority behind the same ticket, TLS, QUIC, and
// package-distribution boundaries used by external clients. Loopback is a location, not a trust
// exemption.
func (controller *networkController) bindListenTransport(runtime *listenHostRuntime) error {
	endpoint, err := controller.listenEndpoint(runtime)
	if err != nil {
		return err
	}

	serverTLS, clientTLS, fingerprint, err := controller.app.networkTrust.HostTLS()
	if err != nil {
		return err
	}

	runtime.clientTLS = clientTLS
	runtime.fingerprint = fingerprint

	runtime.server, err = sessionquic.Listen(listenBindAddress, serverTLS, endpoint)
	if err != nil {
		return err
	}

	packages, err := serverapp.NewPackageProvider(
		runtime.host.Allocation.Identity.Recipe,
		controller.app.options.ModCache,
	)
	if err != nil {
		return err
	}

	runtime.server.SetPackageProvider(packages)
	slog.Debug("listen transport bound", "address", runtime.server.Addr())

	return nil
}

// listenEndpoint uses a fresh per-session ticket secret and queues departures through authority.
// The leave hook must not write saves directly because the authoritative simulation owns the final
// player state.
func (controller *networkController) listenEndpoint(
	runtime *listenHostRuntime,
) (*gameserver.Endpoint, error) {
	secret, err := randomBytes(32)
	if err != nil {
		return nil, err
	}

	runtime.tickets, err = gameserver.NewTicketAuthority(secret, listenGameID)
	if err != nil {
		return nil, err
	}

	endpoint, err := gameserver.NewEndpoint(
		runtime.host,
		runtime.tickets,
		playeradapter.ProjectClientView,
	)
	if err != nil {
		return nil, err
	}

	endpoint.SetSnapshotPending(func(err error) bool {
		return errors.Is(err, playeradapter.ErrHUDPlayer)
	})

	departures := &playeradapter.DepartureQueue{}

	endpoint.SetLeave(func(principal gameserver.Principal) error {
		return departures.Submit(runtime.host.Session, principal.PlayerID)
	})

	return endpoint, nil
}

// installListenWorlds supplies trusted collision maps before submitting population bootstrap.
// Reversing that order could spawn entities into a world whose movement bounds do not yet exist.
func (controller *networkController) installListenWorlds(
	ctx context.Context,
	host *gameserver.Host,
) error {
	for levelID, collision := range controller.app.gameWorlds {
		if err := modruntime.SetWorldMapForLevel(
			ctx,
			host.Authority.Runtime,
			"d2legacy.gameplay.systems.init",
			"set_collision",
			levelID,
			collision,
		); err != nil {
			return err
		}
	}

	population, err := controller.app.populationBootstrapCommand()
	if err != nil {
		return err
	}

	if err := host.Session.Submit(population); err != nil {
		return err
	}

	slog.Debug("listen authority worlds installed", "levels", len(controller.app.gameWorlds))

	return nil
}

// configureListenProfile gives the selected local save a short-lived opaque credential. Even the
// host player crosses profile admission so direct and Realm sessions share identity rules.
func (controller *networkController) configureListenProfile(runtime *listenHostRuntime) error {
	destination, err := controller.listenDestination()
	if err != nil {
		return err
	}

	credential, err := randomBytes(32)
	if err != nil {
		return err
	}

	runtime.profileCredential = hex.EncodeToString(credential)

	profiles, err := serverapp.NewRemoteProfileAdmissions(
		runtime.host,
		runtime.tickets,
		serverapp.RemoteProfileConfig{
			Credential:  runtime.profileCredential,
			AllowDirect: true,
			PrincipalID: listenPrincipalID,
			PlayerID:    listenPlayerID,
			Destination: destination,
			Lifetime:    listenCredentialTTL,
		},
	)
	if err != nil {
		return err
	}

	runtime.server.SetProfileAdmissions(profiles)
	slog.Debug("listen profile admission configured")

	return nil
}

// listenDestination derives spawn coordinates and bounds from already loaded world data. Treating
// frontend-provided coordinates as authoritative would permit invalid or out-of-bounds admission.
func (controller *networkController) listenDestination() (playeradapter.Destination, error) {
	level := controller.app.activeWorldLevel
	world := controller.app.gameWorlds[level]
	zone := controller.app.gameWorldZones[level]
	spawn, found := controller.app.gameWorldSpawns[level]

	if world == nil || zone == nil || !found {
		return playeradapter.Destination{}, errors.New("active world has no trusted host destination")
	}

	request := zone.Request()

	return playeradapter.NewDestination(
		spawn[0],
		spawn[1],
		float64(world.WidthSubtiles),
		float64(world.HeightSubtiles),
		int64(request.Act),
		int64(request.LevelID),
	)
}

// startListenServices starts background work only after all prerequisites are installed. Either
// unexpected service failure invalidates the same startup generation and tears down the pair.
func (controller *networkController) startListenServices(
	ctx context.Context,
	generation uint64,
	runtime *listenHostRuntime,
) {
	go func() {
		if err := runtime.host.Session.Run(ctx); err != nil && ctx.Err() == nil {
			controller.fail(generation, err)
		}
	}()

	go func() {
		if err := runtime.server.Serve(ctx); err != nil && ctx.Err() == nil {
			controller.fail(generation, err)
		}
	}()
}

// connectListenPlayer dials loopback through pinned TLS and profile admission, then activates only
// network-safe components and an entity-empty presentation replica.
func (controller *networkController) connectListenPlayer(
	ctx context.Context,
	runtime *listenHostRuntime,
) error {
	slog.Debug("connecting host player to listen authority")

	assignment := clientsession.SelfHostedAssignment{
		GameID: listenGameID,
		Endpoint: realm.GameEndpoint{
			Address:        listenDialAddress,
			TLSFingerprint: runtime.fingerprint,
		},
		Runtime:           runtime.host.Authority.Identity,
		ProfileCredential: runtime.profileCredential,
	}

	client, err := clientsession.ConnectSelfHosted(
		ctx,
		assignment,
		runtime.clientTLS,
		controller.app.saves,
	)
	if err != nil {
		return err
	}

	runtime.client = client

	if err := controller.app.activateNetworkClientComponents(ctx); err != nil {
		return err
	}

	return controller.app.prepareConnectedWorld(ctx)
}

// commitListenHost is the all-or-nothing ownership transfer. A canceled or superseded generation
// rejects the runtime, leaving its temporary owner responsible for cleanup.
func (controller *networkController) commitListenHost(
	generation uint64,
	runtime *listenHostRuntime,
) bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.generation != generation || controller.phase != "starting" {
		return false
	}

	controller.host = runtime.host
	controller.server = runtime.server
	controller.client = runtime.client
	controller.phase = "connected"
	controller.address = runtime.server.Addr()
	controller.connectionEpoch++
	controller.resetInputLocked()

	// Resource ownership now belongs to controller.Close and controller.fail.

	return true
}

// startClientLoops begins transport work only after client ownership is committed, so either loop
// can safely route terminal failure through the controller.
func (controller *networkController) startClientLoops(
	ctx context.Context,
	client *clientsession.Session,
) {
	go controller.send(ctx, client)
	go controller.watch(ctx, client)
}

// logListenPlayer records authority-issued identities rather than the requested local save values,
// making admission mismatches visible during diagnosis.
func (controller *networkController) logListenPlayer(runtime *listenHostRuntime) {
	hud, _ := runtime.client.View()
	slog.Debug(
		"listen server connected local player",
		"address",
		runtime.server.Addr(),
		"player_id",
		hud.Player.PlayerID,
		"character_id",
		hud.Player.CharacterID,
	)
}

// randomBytes uses the operating system CSPRNG because ticket secrets and profile credentials are
// authentication material, not merely collision-resistant identifiers.
func randomBytes(size int) ([]byte, error) {
	value := make([]byte, size)

	if _, err := rand.Read(value); err != nil {
		return nil, err
	}

	return value, nil
}
