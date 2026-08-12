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

func (controller *testPlayerController) SetRunning(value bool) { controller.running = value }
func (controller *testPlayerController) SetMoveTargetWithRadius(x, y, radius float64) error {
	controller.x, controller.y, controller.radius, controller.pending = x, y, radius, true
	return nil
}
func (controller *testPlayerController) HasMoveTarget() bool { return controller.pending }

func TestPlayerControlModuleQueuesMovementIntent(t *testing.T) {
	controller := &testPlayerController{}
	runtime := New()
	if err := runtime.RegisterModule(PlayerControlModule(controller)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `local player=require("engine.player/v1"); player.request_running(true); player.request_move(12.5, 44.25)`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": {Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if !controller.running {
		t.Fatal("Lua run intent did not reach the fixed-tick command mailbox")
	}
	if !controller.pending {
		t.Fatal("Lua move target did not reach the movement mailbox")
	}
}
