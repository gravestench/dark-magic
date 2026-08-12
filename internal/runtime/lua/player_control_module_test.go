package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
)

func TestPlayerControlModuleQueuesMovementIntent(t *testing.T) {
	controller := &d2movement.MovementController{}
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
	if !controller.Running() {
		t.Fatal("Lua run intent did not reach the fixed-tick command mailbox")
	}
	if !controller.HasMoveTarget() {
		t.Fatal("Lua move target did not reach the movement mailbox")
	}
}
