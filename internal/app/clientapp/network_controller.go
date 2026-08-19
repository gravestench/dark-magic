package clientapp

import (
	"context"
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
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// networkController is the single state machine for local, direct, hosted, and Realm play.
// Its generation counters prevent asynchronous work from an abandoned attempt from taking
// ownership of a newer session.
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

// networkResources is an ownership bundle detached while the controller mutex is held.
// Callers close the bundle after unlocking so shutdown callbacks cannot deadlock the state
// machine.
type networkResources struct {
	cancel  context.CancelFunc
	client  *clientsession.Session
	server  *sessionquic.Server
	host    *gameserver.Host
	mode    string
	address string
}

// newNetworkController starts without session authority; the frontend must explicitly select a
// character before local or network play can begin.
func newNetworkController(app *application) *networkController {
	return &networkController{
		app:         app,
		phase:       "frontend",
		submissions: make(chan gameserver.CommandIntent, 64),
	}
}

// Host records intent but defers authority creation until character selection. That boundary
// keeps canceling from the selection screen cheap and prevents an unselected save from being
// admitted accidentally.
func (controller *networkController) Host() error {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.phase != "frontend" && controller.phase != "failed" {
		return controller.rejectLocked("host", errors.New("network operation already active"))
	}

	controller.phase = "selecting"
	controller.mode = "host"
	controller.address = ""
	controller.failure = ""

	slog.Debug("network host requested; awaiting character selection")

	return nil
}

// StartSelected turns the selected save into the identity for the requested session. Local play
// changes phase synchronously, while network startup receives a generation and runs in the
// background so the frontend remains responsive.
func (controller *networkController) StartSelected() error {
	controller.mu.Lock()

	if err := controller.validateSelectedStartLocked(); err != nil {
		mode := controller.mode
		err = controller.rejectLocked(mode, err)
		controller.mu.Unlock()

		return err
	}

	character, selected := controller.app.saves.Selected()
	if !selected {
		mode := controller.mode
		err := controller.rejectLocked(mode, errors.New("select a character before continuing"))
		controller.mu.Unlock()

		return err
	}

	if controller.phase == "frontend" {
		controller.phase = "local"
		controller.mode = "local"
		controller.failure = ""
		controller.mu.Unlock()

		slog.Debug(
			"local game session activated",
			"character_id", character.ID,
			"character_name", character.Name,
			"character_class", character.Class,
		)

		return nil
	}

	generation, ctx, mode, address := controller.beginSelectedNetworkLocked()
	controller.mu.Unlock()

	slog.Debug(
		"network operation starting",
		"mode", mode,
		"address", address,
		"character_id", character.ID,
		"character_name", character.Name,
		"character_class", character.Class,
	)

	controller.launchSelectedNetwork(ctx, generation, mode, address)

	return nil
}

// validateSelectedStartLocked allows selection to advance only from the frontend or a pending
// host/join request. The caller holds mu, making validation and the following transition atomic.
func (controller *networkController) validateSelectedStartLocked() error {
	if controller.phase == "frontend" {
		return nil
	}

	validMode := controller.mode == "host" || controller.mode == "join"
	if controller.phase == "selecting" && validMode {
		return nil
	}

	return errors.New("no network operation is awaiting character selection")
}

// beginSelectedNetworkLocked gives this attempt a cancellation scope and a unique generation.
// Any completion carrying an older generation must discard its prepared resources.
func (controller *networkController) beginSelectedNetworkLocked() (
	uint64,
	context.Context,
	string,
	string,
) {
	controller.generation++
	ctx, cancel := context.WithCancel(controller.app.ctx)
	controller.cancel = cancel
	controller.phase = "starting"
	controller.failure = ""

	return controller.generation, ctx, controller.mode, controller.address
}

// launchSelectedNetwork transfers startup to a goroutine after all state needed by that goroutine
// has been copied out from under the controller lock.
func (controller *networkController) launchSelectedNetwork(
	ctx context.Context,
	generation uint64,
	mode string,
	address string,
) {
	if mode == "host" {
		go controller.startHost(ctx, generation)

		return
	}

	go controller.startJoin(ctx, generation, address)
}

// Cancel abandons setup phases without silently tearing down an established game. Incrementing
// generation invalidates workers that may observe cancellation only after completing more work.
func (controller *networkController) Cancel() {
	controller.mu.Lock()

	if !cancelableNetworkPhase(controller.phase) {
		controller.mu.Unlock()

		return
	}

	controller.generation++
	cancel := controller.cancel
	controller.cancel = nil
	controller.phase = "frontend"
	controller.mode = ""
	controller.address = ""
	controller.failure = ""
	controller.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// cancelableNetworkPhase is an explicit allowlist: connected and reconnecting sessions require
// Close so their leave and persistence obligations cannot be skipped.
func cancelableNetworkPhase(phase string) bool {
	return phase == "selecting" || phase == "starting" || phase == "failed"
}

// Join normalizes a direct-server address and waits for selection. Supplying the default port here
// gives every later trust, dialing, and status path one canonical endpoint string.
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

	controller.phase = "selecting"
	controller.mode = "join"
	controller.address = address
	controller.failure = ""

	slog.Debug("network join requested; awaiting character selection", "address", address)

	return nil
}

// fail publishes failure only for the active generation, then closes its detached resources.
// Realm departure precedes transport teardown because it is the last opportunity to commit the
// authority's canonical character state.
func (controller *networkController) fail(generation uint64, cause error) {
	controller.mu.Lock()

	if !controller.canFailLocked(generation) {
		controller.mu.Unlock()

		return
	}

	resources := controller.detachResourcesLocked()
	controller.phase = "failed"
	controller.failure = cause.Error()
	controller.mu.Unlock()

	if resources.cancel != nil {
		resources.cancel()
	}

	// Realm departure commits canonical state before transport teardown.
	cleanupContext, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()

	if resources.mode == "realm" && controller.app != nil && controller.app.realm != nil {
		cause = errors.Join(cause, controller.app.realm.LeaveConnectedGame(cleanupContext))
	}

	if resources.client != nil {
		_ = resources.client.Close(cleanupContext)
	}

	if resources.server != nil {
		_ = resources.server.Close()
	}

	if resources.host != nil {
		_ = resources.host.Close(cleanupContext)
	}

	slog.Debug(
		"network operation failed",
		"mode", resources.mode,
		"address", resources.address,
		"error", cause,
	)
}

// canFailLocked prevents stale startup or recovery goroutines from replacing a newer phase with
// their late error.
func (controller *networkController) canFailLocked(generation uint64) bool {
	return generation == controller.generation &&
		controller.phase != "closed" &&
		controller.phase != "failed"
}

// detachResourcesLocked transfers every live network owner out of the controller exactly once.
// Clearing the fields makes repeated failure or shutdown paths harmless.
func (controller *networkController) detachResourcesLocked() networkResources {
	resources := networkResources{
		cancel:  controller.cancel,
		client:  controller.client,
		server:  controller.server,
		host:    controller.host,
		mode:    controller.mode,
		address: controller.address,
	}

	controller.cancel = nil
	controller.client = nil
	controller.server = nil
	controller.host = nil

	return resources
}

// rejectLocked converts a synchronous validation error into frontend-visible state as well as
// returning it to the immediate caller.
func (controller *networkController) rejectLocked(mode string, err error) error {
	controller.phase = "failed"
	controller.mode = mode
	controller.failure = err.Error()

	return err
}

// Status returns a presentation-safe snapshot without exposing tickets, TLS material, or mutable
// session objects. The HUD lookup happens after unlocking because it may perform its own locking.
func (controller *networkController) Status() map[string]any {
	controller.mu.Lock()
	phase := controller.phase
	mode := controller.mode
	address := controller.address
	failure := controller.failure
	client := controller.client
	controller.mu.Unlock()

	playerID := "local-player"

	if client != nil {
		hud, _ := client.View()
		if hud.Player.PlayerID != "" {
			playerID = hud.Player.PlayerID
		}
	}

	return map[string]any{
		"phase":     phase,
		"mode":      mode,
		"address":   address,
		"error":     failure,
		"player_id": playerID,
	}
}

// hasSelectedCharacter uses the signed admission identity rather than transient HUD state. This
// prevents the loading flow from treating an unauthenticated or partially connected client as an
// admitted Realm character.
func (controller *networkController) hasSelectedCharacter() bool {
	if controller == nil {
		return false
	}

	controller.mu.Lock()
	phase := controller.phase
	mode := controller.mode
	client := controller.client
	controller.mu.Unlock()

	if phase != "connected" || mode != "realm" || client == nil {
		return false
	}

	return strings.TrimSpace(client.Admission.Admission.CharacterID) != ""
}

// Connected requires both the connected phase and a committed client, so presentation code never
// mistakes an in-progress transition for a usable remote session.
func (controller *networkController) Connected() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return controller.phase == "connected" && controller.client != nil
}

// Local reports the distinct offline-authority phase; hosted play is intentionally not local even
// though its server happens to run in this process.
func (controller *networkController) Local() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return controller.phase == "local"
}

// currentGeneration snapshots the startup identity used to reject late asynchronous completions.
func (controller *networkController) currentGeneration() uint64 {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return controller.generation
}

// currentConnectionEpoch snapshots the committed transport revision. Recovery work must still
// match this value before it can replace the active endpoint.
func (controller *networkController) currentConnectionEpoch() uint64 {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return controller.connectionEpoch
}

// resetInputLocked starts sequence numbers, fixed-step sampling, and backpressure from a clean
// connection boundary; carrying any of them across servers could duplicate or suppress input.
func (controller *networkController) resetInputLocked() {
	controller.sequence = 0
	controller.lastMovementTick = 0
	controller.lastMovementActive = false
	controller.inputLag = 0
	controller.submissions = make(chan gameserver.CommandIntent, 64)
}

// Close first removes resource ownership from the state machine, then fulfills leave/persistence
// obligations before tearing down transports and authority. This ordering preserves the last
// canonical character state while preventing new work from observing a half-closed session.
func (controller *networkController) Close() error {
	slog.Debug("closing network controller")

	controller.mu.Lock()
	resources := controller.detachResourcesLocked()
	controller.phase = "closed"
	controller.mu.Unlock()

	ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()

	var err error

	if resources.client != nil {
		err = errors.Join(err, controller.leaveConnectedSession(ctx, resources))
		err = errors.Join(err, resources.client.Close(ctx))
	}

	if resources.cancel != nil {
		resources.cancel()
	}

	if resources.server != nil {
		err = errors.Join(err, resources.server.Close())
	}

	if resources.host != nil {
		err = errors.Join(err, resources.host.Close(ctx))
	}

	return err
}

// leaveConnectedSession distinguishes Realm-owned persistence from self-hosted persistence. Realm
// state is committed through its control plane, while direct and hosted games must merge the
// authority snapshot back into the local save store.
func (controller *networkController) leaveConnectedSession(
	ctx context.Context,
	resources networkResources,
) error {
	if resources.mode == "realm" && controller.app != nil && controller.app.realm != nil {
		return controller.app.realm.LeaveConnectedGame(ctx)
	}

	if resources.mode == "host" || resources.mode == "join" {
		return controller.persistSelfHostedCharacter(ctx, resources.client)
	}

	return nil
}

// persistSelfHostedCharacter asks the authority for a final snapshot before touching disk. Using a
// cached HUD could lose changes accepted immediately before disconnect.
func (controller *networkController) persistSelfHostedCharacter(
	ctx context.Context,
	client *clientsession.Session,
) error {
	if controller.app == nil || controller.app.saves == nil || client == nil {
		return nil
	}

	if _, err := client.Refresh(ctx); err != nil {
		return fmt.Errorf("refresh selected self-hosted character: %w", err)
	}

	hud, _ := client.View()

	return updateSelectedCharacter(controller.app.saves, hud)
}

// updateSelectedCharacter preserves the selected save's durable identity and replaces only fields
// represented by the canonical HUD. The store performs the final atomic selected-save update.
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
