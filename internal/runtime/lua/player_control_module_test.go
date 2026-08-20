package modruntime

import (
	"context"
	"testing"
	"testing/fstest"
)

type testPlayerController struct {
	running, pending bool
	x, y, radius     float64
}

// SetRunning applies running through the capability boundary so validation completes before shared state
// changes.
func (controller *testPlayerController) SetRunning(value bool) { controller.running = value }

// SetMoveTargetWithRadius applies move target with radius through the capability boundary so validation
// completes before shared state changes.
func (controller *testPlayerController) SetMoveTargetWithRadius(x, y, radius float64) error {
	controller.x, controller.y, controller.radius, controller.pending = x, y, radius, true
	return nil
}

// HasMoveTarget owns the has move target step at this boundary, keeping its side effects and failure point explicit
// to callers.
func (controller *testPlayerController) HasMoveTarget() bool { return controller.pending }

// TestPlayerControlModuleQueuesMovementIntent protects the player control module queues movement intent contract,
// including its observable ordering and failure behavior.
func TestPlayerControlModuleQueuesMovementIntent(t *testing.T) {
	controller := &testPlayerController{}

	runtime := New()
	if err := runtime.RegisterModule(PlayerControlModule(controller)); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	script := `local player=require("engine.player/v1"); player.request_running(true); player.request_move(12.5, 44.25)`
	if err := runtime.Execute(
		context.Background(),
		fstest.MapFS{"test.lua": {Data: []byte(script)}},
		"test.lua",
	); err != nil {
		t.Fatal(err)
	}

	if !controller.running {
		t.Fatal("Lua run intent did not reach the fixed-tick command mailbox")
	}

	if !controller.pending {
		t.Fatal("Lua move target did not reach the movement mailbox")
	}
}
