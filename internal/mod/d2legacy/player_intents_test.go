package d2legacy_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	. "github.com/gravestench/dark-magic/internal/runtime/lua"
)

func TestD2LegacyBuildsSkillCommandsThroughGenericIntentMailbox(t *testing.T) {
	controller := &gamesession.IntentController{}
	runtime := New()
	if err := runtime.RegisterModule(CommandIntentModule(controller)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(PlayerControlModule(&d2movement.MovementController{})); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	script := fstest.MapFS{"test.lua": {Data: []byte(`
local player=require("d2legacy.player_intents")
player.assign_skill("right",17)
player.request_skill("right",20.5,30.25,"fallen:7")
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
	source, err := gamesession.NewIntentSource(controller, "alice")
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(9)
	if len(commands) != 2 || commands[0].Kind != "player.assign_skills" || commands[1].Kind != "player.use_skill" {
		t.Fatalf("d2legacy skill requests = %#v", commands)
	}
}
