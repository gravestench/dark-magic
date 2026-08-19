package clientapp

import (
	"context"
	"crypto/tls"
	"errors"
	"io/fs"
	"log/slog"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2legacy "github.com/gravestench/dark-magic/internal/mod/d2legacy"
)

// realmConnectionPlan carries both the prepared client and the rollback obligation created when
// package recomposition mutates live application state.
type realmConnectionPlan struct {
	client     *clientsession.Session
	recomposed bool
}

// selfHostedJoinPlan separates authenticated preparation from live recomposition. Keeping these
// values together ensures the dial uses the exact recipe and TLS policy that were verified.
type selfHostedJoinPlan struct {
	address   string
	clientTLS *tls.Config
	recipe    simulation.RuntimeRecipe
	identity  simulation.RuntimeIdentity
}

// ConnectRealm consumes a private Realm assignment that must never pass through Lua or frontend
// state. A failed handoff is routed through the generation-aware failure path for safe cleanup.
func (controller *networkController) ConnectRealm(
	ctx context.Context,
	assignment realm.JoinAssignment,
) error {
	if err := validateRealmAssignment(ctx, assignment); err != nil {
		return err
	}

	generation, runCtx, err := controller.beginRealmConnection(ctx, assignment)
	if err != nil {
		return err
	}

	if err := controller.connectRealm(runCtx, generation, assignment); err != nil {
		controller.fail(generation, err)

		return err
	}

	return nil
}

// validateRealmAssignment rejects incomplete private handoffs before they can mutate package or
// controller state. The game ID and ticket bind this client to one Realm-created session.
func validateRealmAssignment(ctx context.Context, assignment realm.JoinAssignment) error {
	if ctx == nil || strings.TrimSpace(assignment.GameID) == "" || strings.TrimSpace(assignment.Ticket) == "" {
		return errors.New("network: invalid Realm assignment")
	}

	return nil
}

// beginRealmConnection atomically reserves the controller and derives cancellation from the
// caller's authenticated login flow. Its generation protects a later session from this work.
func (controller *networkController) beginRealmConnection(
	ctx context.Context,
	assignment realm.JoinAssignment,
) (uint64, context.Context, error) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.phase != "frontend" && controller.phase != "failed" {
		return 0, nil, errors.New("network operation already active")
	}

	controller.generation++
	runCtx, cancel := context.WithCancel(ctx)
	controller.cancel = cancel
	controller.phase = "starting"
	controller.mode = "realm"
	controller.address = assignment.Endpoint.Address
	controller.failure = ""

	return controller.generation, runCtx, nil
}

// ReconnectRealm serializes reassignment and pins the replacement worker's TLS fingerprint before
// the existing logical session is retargeted. The player and game identities do not change.
func (controller *networkController) ReconnectRealm(
	ctx context.Context,
	assignment realm.JoinAssignment,
) error {
	if err := validateRealmReconnect(ctx, assignment); err != nil {
		return err
	}

	controller.reconnectMu.Lock()
	defer controller.reconnectMu.Unlock()

	client, err := controller.beginRealmReconnect()
	if err != nil {
		return err
	}

	tlsConfig, err := networktrust.PinnedTLSFingerprint(assignment.Endpoint.TLSFingerprint)
	if err == nil {
		err = client.Reassign(ctx, assignment, tlsConfig)
	}

	return controller.finishRealmReconnect(client, assignment, err)
}

// validateRealmReconnect requires the durable game identity and a fresh ticket; an endpoint alone
// is not authority to resume a Realm session.
func validateRealmReconnect(ctx context.Context, assignment realm.JoinAssignment) error {
	if ctx == nil || strings.TrimSpace(assignment.GameID) == "" || strings.TrimSpace(assignment.Ticket) == "" {
		return errors.New("network: invalid Realm reconnect assignment")
	}

	return nil
}

// beginRealmReconnect moves only a committed Realm client into the reconnecting phase. Capturing
// its pointer lets the commit path detect if Close or another transition replaced it.
func (controller *networkController) beginRealmReconnect() (*clientsession.Session, error) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	client := controller.client
	if controller.phase != "connected" || controller.mode != "realm" || client == nil {
		return nil, errors.New("network: no connected Realm session")
	}

	controller.phase = "reconnecting"

	return client, nil
}

// finishRealmReconnect restores the usable connected phase after a failed attempt, but commits a
// successful endpoint only if the same client is still owned. Advancing the epoch invalidates
// recovery work started against the previous transport.
func (controller *networkController) finishRealmReconnect(
	client *clientsession.Session,
	assignment realm.JoinAssignment,
	reassignErr error,
) error {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if reassignErr != nil {
		if controller.client == client && controller.phase == "reconnecting" {
			controller.phase = "connected"
		}

		return reassignErr
	}

	if controller.client != client || controller.phase == "closed" {
		return context.Canceled
	}

	controller.address = assignment.Endpoint.Address
	controller.phase = "connected"
	controller.failure = ""
	controller.connectionEpoch++

	slog.Debug(
		"Realm session moved to replacement worker",
		"game_id", assignment.GameID,
		"address", assignment.Endpoint.Address,
	)

	return nil
}

// connectRealm prepares all trust-sensitive inputs before installing the presentation world, then
// commits ownership only if the startup generation is still active. Every later failure repairs
// live package composition before returning.
func (controller *networkController) connectRealm(
	ctx context.Context,
	generation uint64,
	assignment realm.JoinAssignment,
) error {
	plan, err := controller.prepareRealmConnection(ctx, assignment)
	if err != nil {
		return controller.restoreAfterRecomposition(plan.recomposed, err)
	}

	if err := controller.app.prepareConnectedWorld(ctx); err != nil {
		_ = plan.client.Close(context.Background())

		return controller.restoreAfterRecomposition(plan.recomposed, err)
	}

	if !controller.commitConnectedClient(generation, plan.client) {
		_ = plan.client.Close(context.Background())

		return controller.restoreAfterRecomposition(plan.recomposed, context.Canceled)
	}

	hud, _ := plan.client.View()
	slog.Debug(
		"joined Realm game",
		"game_id", assignment.GameID,
		"address", assignment.Endpoint.Address,
		"player_id", hud.Player.PlayerID,
		"character_id", hud.Player.CharacterID,
	)

	controller.startClientLoops(ctx, plan.client)

	return nil
}

// prepareRealmConnection pins transport trust, obtains the authenticated package recipe, and
// composes it before dialing gameplay. Once recomposition begins, callers inherit a rollback
// obligation even when the failing step appears unrelated.
func (controller *networkController) prepareRealmConnection(
	ctx context.Context,
	assignment realm.JoinAssignment,
) (*realmConnectionPlan, error) {
	plan := &realmConnectionPlan{}

	tlsConfig, err := networktrust.PinnedTLSFingerprint(assignment.Endpoint.TLSFingerprint)
	if err != nil {
		return plan, err
	}

	store, err := controller.app.ensureModCache()
	if err != nil {
		return plan, err
	}

	recipe, err := clientsession.PrepareExtensions(
		ctx,
		assignment,
		tlsConfig,
		store,
		controller.app.options.Packages.Base,
	)
	if err != nil {
		return plan, err
	}

	// Recomposition may partially mutate live state, so all later errors repair it.
	plan.recomposed = true

	if err := controller.app.recomposeForNetworkRecipe(ctx, recipe); err != nil {
		return plan, err
	}

	// Realm owns the complete authority identity; the client only verifies it.
	if err := sameRuntimeRecipe(assignment.Runtime, recipe); err != nil {
		return plan, err
	}

	plan.client, err = clientsession.Connect(ctx, assignment, tlsConfig)
	if err != nil {
		return plan, err
	}

	return plan, nil
}

// startJoin performs direct-server preparation, live recomposition, dialing, and commit in that
// order. Preparing first avoids leaving the offline application half-mutated for failures that can
// be detected without touching live state.
func (controller *networkController) startJoin(
	ctx context.Context,
	generation uint64,
	address string,
) {
	slog.Debug("dialing self-hosted game", "address", address)

	plan, err := controller.prepareSelfHostedJoin(ctx, address)
	if err != nil {
		controller.fail(generation, err)

		return
	}

	if err := controller.composeSelfHostedJoin(ctx, plan); err != nil {
		controller.fail(generation, controller.restoreAfterRecomposition(true, err))

		return
	}

	client, err := controller.connectSelfHostedJoin(ctx, plan)
	if err != nil {
		controller.fail(generation, controller.restoreAfterRecomposition(true, err))

		return
	}

	if !controller.commitConnectedClient(generation, client) {
		_ = client.Close(context.Background())
		_ = controller.restoreAfterRecomposition(true, nil)

		return
	}

	hud, _ := client.View()
	slog.Debug(
		"joined self-hosted game",
		"address", address,
		"player_id", hud.Player.PlayerID,
		"character_id", hud.Player.CharacterID,
	)

	controller.startClientLoops(ctx, client)
}

// prepareSelfHostedJoin verifies TLS and the server's recipe against locally derivable runtime
// identity without mutating the VFS. This makes early trust or download failures rollback-free.
func (controller *networkController) prepareSelfHostedJoin(
	ctx context.Context,
	address string,
) (*selfHostedJoinPlan, error) {
	source, err := controller.app.modSource("d2legacy")
	if err != nil {
		return nil, err
	}

	identity, err := controller.runtimeIdentity(source, controller.app.options.Packages)
	if err != nil {
		return nil, err
	}

	clientTLS, err := controller.app.networkTrust.ClientTLS(address)
	if err != nil {
		return nil, err
	}

	store, err := controller.app.ensureModCache()
	if err != nil {
		return nil, err
	}

	assignment := clientsession.SelfHostedAssignment{
		GameID:   listenGameID,
		Endpoint: realm.GameEndpoint{Address: address},
		Runtime:  identity,
	}

	recipe, err := clientsession.PrepareSelfHostedExtensions(
		ctx,
		assignment,
		clientTLS,
		store,
		controller.app.options.Packages.Base,
	)
	if err != nil {
		return nil, err
	}

	identity, err = controller.runtimeIdentity(source, recipe.Packages)
	if err != nil {
		return nil, err
	}

	if err := sameRuntimeRecipe(identity, recipe); err != nil {
		return nil, err
	}

	return &selfHostedJoinPlan{
		address:   address,
		clientTLS: clientTLS,
		recipe:    recipe,
		identity:  identity,
	}, nil
}

// composeSelfHostedJoin installs the verified recipe, then reopens the mod source because VFS
// replacement invalidates the previous view. A second digest check proves the live composition is
// the one that was authenticated.
func (controller *networkController) composeSelfHostedJoin(
	ctx context.Context,
	plan *selfHostedJoinPlan,
) error {
	if err := controller.app.recomposeForNetworkRecipe(ctx, plan.recipe); err != nil {
		return err
	}

	// Reopen the source after VFS replacement so identity reflects live content.
	source, err := controller.app.modSource("d2legacy")
	if err != nil {
		return err
	}

	plan.identity, err = controller.runtimeIdentity(source, controller.app.options.Packages)
	if err != nil {
		return err
	}

	return sameRuntimeRecipe(plan.identity, plan.recipe)
}

// connectSelfHostedJoin authenticates the selected local save and builds a presentation-only world
// before returning ownership. A presentation failure closes the otherwise valid transport.
func (controller *networkController) connectSelfHostedJoin(
	ctx context.Context,
	plan *selfHostedJoinPlan,
) (*clientsession.Session, error) {
	client, err := clientsession.ConnectSelfHosted(
		ctx,
		clientsession.SelfHostedAssignment{
			GameID:   listenGameID,
			Endpoint: realm.GameEndpoint{Address: plan.address},
			Runtime:  plan.identity,
		},
		plan.clientTLS,
		controller.app.saves,
	)
	if err != nil {
		return nil, err
	}

	if err := controller.app.prepareConnectedWorld(ctx); err != nil {
		_ = client.Close(context.Background())

		return nil, err
	}

	return client, nil
}

// runtimeIdentity binds packages, assets, generated data, and bootstrap state into the digest used
// by both peers. Any difference can affect simulation and therefore must reject the connection.
func (controller *networkController) runtimeIdentity(
	source fs.FS,
	packages simulation.RuntimePackageSet,
) (simulation.RuntimeIdentity, error) {
	return d2legacy.IdentityForPackagesAndData(
		source,
		packages,
		controller.app.options.AssetSetID,
		controller.app.gameDataGenerationID(),
		controller.app.sessionInitialData(),
	)
}

// commitConnectedClient is the ownership handoff point. It rejects canceled generations and
// resets all connection-scoped input state before exposing the connected phase.
func (controller *networkController) commitConnectedClient(
	generation uint64,
	client *clientsession.Session,
) bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.generation != generation || controller.phase != "starting" {
		return false
	}

	controller.client = client
	controller.phase = "connected"
	controller.connectionEpoch++
	controller.resetInputLocked()

	return true
}

// restoreAfterRecomposition returns the application to its configured offline package set after a
// failed network attempt. The bounded background context keeps rollback possible after the
// original connection context has already been canceled.
func (controller *networkController) restoreAfterRecomposition(
	recomposed bool,
	cause error,
) error {
	if !recomposed {
		return cause
	}

	restoreContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	restoreErr := controller.app.restoreConfiguredPackages(restoreContext)

	cancel()

	return errors.Join(cause, restoreErr)
}

// sameRuntimeRecipe compares deterministic digests rather than selected fields so newly added
// simulation inputs cannot silently escape compatibility checks.
func sameRuntimeRecipe(
	identity simulation.RuntimeIdentity,
	recipe simulation.RuntimeRecipe,
) error {
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
