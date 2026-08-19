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

// networkController coordinates one local, direct, hosted, or Realm session.
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

// networkResources groups resources detached during failure or shutdown.
type networkResources struct {
	cancel  context.CancelFunc
	client  *clientsession.Session
	server  *sessionquic.Server
	host    *gameserver.Host
	mode    string
	address string
}

// newNetworkController creates a controller in the frontend selection phase.
func newNetworkController(app *application) *networkController {
	return &networkController{
		app:         app,
		phase:       "frontend",
		submissions: make(chan gameserver.CommandIntent, 64),
	}
}

// Host records a host request and waits for explicit character selection.
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

// StartSelected activates local play or launches the pending network request.
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

// validateSelectedStartLocked checks whether selection can advance the phase.
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

// beginSelectedNetworkLocked allocates one cancelable startup generation.
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

// launchSelectedNetwork starts the requested host or direct-join workflow.
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

// Cancel returns cancelable setup phases to the frontend.
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

// cancelableNetworkPhase reports whether Cancel may return to the frontend.
func cancelableNetworkPhase(phase string) bool {
	return phase == "selecting" || phase == "starting" || phase == "failed"
}

// Join records a normalized direct-server address and waits for selection.
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

// fail rejects one active generation and releases all resources it owned.
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

// canFailLocked prevents stale goroutines from overwriting a newer phase.
func (controller *networkController) canFailLocked(generation uint64) bool {
	return generation == controller.generation &&
		controller.phase != "closed" &&
		controller.phase != "failed"
}

// detachResourcesLocked transfers active resources out of the controller.
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

// rejectLocked records a synchronous request failure for frontend display.
func (controller *networkController) rejectLocked(mode string, err error) error {
	controller.phase = "failed"
	controller.mode = mode
	controller.failure = err.Error()

	return err
}

// Status returns a presentation-safe snapshot of controller state.
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

// hasSelectedCharacter reports whether Realm admitted a character for loading.
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

// Connected reports whether a remote session is ready.
func (controller *networkController) Connected() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return controller.phase == "connected" && controller.client != nil
}

// Local reports whether the selected character is playing offline.
func (controller *networkController) Local() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return controller.phase == "local"
}

// currentGeneration returns the identity of the active startup attempt.
func (controller *networkController) currentGeneration() uint64 {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return controller.generation
}

// currentConnectionEpoch returns the transport revision used by recovery.
func (controller *networkController) currentConnectionEpoch() uint64 {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return controller.connectionEpoch
}

// resetInputLocked clears sequencing state for a newly committed connection.
func (controller *networkController) resetInputLocked() {
	controller.sequence = 0
	controller.lastMovementTick = 0
	controller.lastMovementActive = false
	controller.inputLag = 0
	controller.submissions = make(chan gameserver.CommandIntent, 64)
}

// Close leaves the active game and releases every network resource.
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

// leaveConnectedSession persists or departs according to connection ownership.
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

// persistSelfHostedCharacter refreshes and stores the selected offline save.
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

// updateSelectedCharacter merges canonical HUD state into the selected save.
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
