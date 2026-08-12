package modruntime

import (
	"testing"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
)

func TestCommandIntentModuleQueuesSerializableModCommand(t *testing.T) {
	controller := &gamesession.IntentController{}
	runtime := New()
	if err := runtime.RegisterModule(CommandIntentModule(controller)); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())

	if _, err := runtime.Eval(t.Context(), `
local intents = require("engine.command_intent/v1")
intents.submit("example.command", { value = 42, nested = { ok = true } })
return true
`); err != nil {
		t.Fatal(err)
	}

	source, err := gamesession.NewIntentSource(controller, "alice")
	if err != nil {
		t.Fatal(err)
	}
	commands := source.Commands(9)
	if len(commands) != 1 {
		t.Fatalf("commands = %d, want 1", len(commands))
	}
	command := commands[0]
	if command.Kind != "example.command" || command.Player != "alice" || command.Tick != 9 {
		t.Fatalf("unexpected command: %+v", command)
	}
}
