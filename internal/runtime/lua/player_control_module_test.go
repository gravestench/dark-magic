package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
)

func TestPlayerControlModuleQueuesMovementIntent(t *testing.T) {
	controller := &gamesession.MovementController{}
	runtime := New()
	if err := runtime.RegisterModule(PlayerControlModule(controller)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(context.Background())
	script := `require("dm.player/v1").request_running(true)`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": {Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if !controller.Running() {
		t.Fatal("Lua run intent did not reach the fixed-tick command mailbox")
	}
}
