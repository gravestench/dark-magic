package modruntime

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

func TestAuthorityStateModuleRunsHeadlesslyAndUsesEngineOwnedState(t *testing.T) {
	stores := simulation.NewStateStore()
	runtime := New()
	if err := runtime.RegisterModule(AuthorityStateModule(stores)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)

	script := `
local state = require("dm.authority_state/v1")

-- The script decides that a counter should advance. The engine owns the bytes.
state.register("d2legacy.example", "counter/v1", { value = 1 })
local current = state.read("d2legacy.example")
state.replace("d2legacy.example", "counter/v1", { value = current.value + 1 })
`
	if err := runtime.Run(ctx, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}
	got, found := stores.Read("d2legacy.example")
	if !found || string(got.Data) != `{"value":2}` {
		t.Fatalf("registered state = %s, %v", got.Data, found)
	}
}
