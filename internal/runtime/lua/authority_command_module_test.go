package modruntime

import (
	"context"
	"encoding/json"
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

func TestAuthorityCommandModuleKeepsAdmissionInGoAndPolicyInLua(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	stores := simulation.NewStateStore()
	if err := stores.Register("d2legacy.test.commands", "counter/v1", []byte(`{"count":0}`)); err != nil {
		t.Fatal(err)
	}

	runtime := New()
	if err := runtime.RegisterModule(AuthorityStateModule(stores)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(AuthorityCommandModule(runtime, session)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)

	script := `
local commands = require("engine.authority_command/v1")
local state = require("engine.authority_state/v1")

commands.register({
    kind = "d2legacy.test.increment",
    authorities = { "player" },
    validate = function(command)
        if command.payload.amount < 1 then
            error("amount must be positive")
        end
    end,
    apply = function(command)
        local counter = state.read("d2legacy.test.commands")
        counter.count = counter.count + command.payload.amount
        state.replace("d2legacy.test.commands", "counter/v1", counter)
    end,
})
`
	if err := runtime.Run(ctx, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}

	payload, _ := json.Marshal(map[string]int{"amount": 3})
	command := simulation.Command{Tick: 1, Player: "alice", Sequence: 1, Kind: "d2legacy.test.increment", Payload: payload}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	got, _ := stores.Read("d2legacy.test.commands")
	if string(got.Data) != `{"count":3}` {
		t.Fatalf("Lua-applied state = %s", got.Data)
	}
}

func TestAuthorityCommandLuaValidatorRejectsBeforeQueueing(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	runtime := New()
	if err := runtime.RegisterModule(AuthorityCommandModule(runtime, session)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	if err := runtime.Run(ctx, func(state *lua.LState) error {
		return state.DoString(`require("engine.authority_command/v1").register({
            kind = "d2legacy.test.reject",
            validate = function() error("no") end,
            apply = function() end,
        })`)
	}); err != nil {
		t.Fatal(err)
	}
	command := simulation.Command{Tick: 1, Player: "alice", Sequence: 1, Kind: "d2legacy.test.reject", Payload: json.RawMessage(`{}`)}
	if err := session.Submit(command); err == nil {
		t.Fatal("Lua validator accepted rejected command")
	}
	if status := session.Status(); status.Pending != 0 {
		t.Fatalf("rejected command entered queue: %+v", status)
	}
}
