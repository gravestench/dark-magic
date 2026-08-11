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
	script := `local player=require("dm.player/v1"); player.request_running(true); player.request_move(12.5, 44.25); player.assign_skill("right", 17); player.request_skill("right", 20.5, 30.25, "fallen:7")`
	if err := runtime.Execute(context.Background(), fstest.MapFS{"test.lua": {Data: []byte(script)}}, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if !controller.Running() {
		t.Fatal("Lua run intent did not reach the fixed-tick command mailbox")
	}
	source, err := gamesession.NewSkillSource(controller, "local-player")
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(1)
	if len(commands) != 2 || commands[0].Kind != gamesession.AssignSkillsCommand || commands[1].Kind != gamesession.UseSkillCommand {
		t.Fatalf("Lua skill intent commands = %#v", commands)
	}
}
