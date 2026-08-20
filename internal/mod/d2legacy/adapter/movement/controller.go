package movement

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
)

// MoveTarget is a wire value consumed by d2legacy Lua. Explicit JSON names preserve the Lua table contract.
type MoveTarget struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	StopRadius float64 `json:"stop_radius"`
}

// MovementController is the thread-safe intent mailbox shared by Lua UI and fixed-tick command generation.
// It deliberately owns no ECS state, so UI writes cannot bypass authoritative simulation.
type MovementController struct {
	running  atomic.Bool
	sequence atomic.Uint64
	mu       sync.Mutex
	target   *MoveTarget
}

// SetRunning records the run mode that subsequent commands must carry across the simulation boundary.
func (controller *MovementController) SetRunning(running bool) {
	controller.running.Store(running)
}

// Running returns the latest run mode without coupling readers to target mailbox locking.
func (controller *MovementController) Running() bool {
	return controller.running.Load()
}

// HasMoveTarget reports whether pointer movement remains active while keeping the target private.
func (controller *MovementController) HasMoveTarget() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	return controller.target != nil
}

// nextSequence assigns monotonically increasing command order even when input producers are concurrent.
func (controller *MovementController) nextSequence() uint64 {
	return controller.sequence.Add(1)
}

// SetMoveTarget records an exact world position with no arrival-radius override.
func (controller *MovementController) SetMoveTarget(x, y float64) error {
	return controller.SetMoveTargetWithRadius(x, y, 0)
}

// SetMoveTargetWithRadius validates pointer intent before it enters pathfinding or serialized commands.
func (controller *MovementController) SetMoveTargetWithRadius(x, y, stopRadius float64) error {
	if !finiteCoordinate(x) || !finiteCoordinate(y) {
		return fmt.Errorf("d2legacy movement: target must be finite")
	}

	if stopRadius < 0 || !finiteCoordinate(stopRadius) {
		return fmt.Errorf("d2legacy movement: stop radius must be non-negative and finite")
	}

	controller.mu.Lock()
	defer controller.mu.Unlock()

	controller.target = &MoveTarget{X: x, Y: y, StopRadius: stopRadius}

	return nil
}

// finiteCoordinate keeps NaN and infinity from poisoning collision-cell and distance calculations.
func finiteCoordinate(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// moveTarget returns a defensive copy so path updates cannot race with UI replacement of the mailbox value.
func (controller *MovementController) moveTarget() *MoveTarget {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.target == nil {
		return nil
	}

	result := *controller.target

	return &result
}

// clearMoveTarget consumes pointer intent when keyboard input, arrival, or a world transition supersedes it.
func (controller *MovementController) clearMoveTarget() {
	controller.mu.Lock()
	controller.target = nil
	controller.mu.Unlock()
}

// restoreMoveTarget replaces a rejected click with the last accepted target using independent storage.
func (controller *MovementController) restoreMoveTarget(target *MoveTarget) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if target == nil {
		controller.target = nil
		return
	}

	copyTarget := *target
	controller.target = &copyTarget
}
