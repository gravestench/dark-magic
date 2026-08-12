package modruntime

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

func TestAuthorityRandomModuleUsesCheckpointedNamedStreams(t *testing.T) {
	streams := simulation.NewRandomStreams(99)
	if err := streams.Register("d2legacy.combat.hit"); err != nil {
		t.Fatal(err)
	}
	checkpoint, err := streams.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}

	runtime := New()
	if err := runtime.RegisterModule(AuthorityRandomModule(streams)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)

	roll := func() lua.LNumber {
		var result lua.LNumber
		err := runtime.Run(ctx, func(state *lua.LState) error {
			return state.DoString(`
local random = require("dm.authority_random/v1")
result = random.integer("d2legacy.combat.hit", 100)
`)
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := runtime.Run(ctx, func(state *lua.LState) error {
			result = state.GetGlobal("result").(lua.LNumber)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		return result
	}

	want := roll()
	if err := streams.RestoreState(checkpoint); err != nil {
		t.Fatal(err)
	}
	if got := roll(); got != want {
		t.Fatalf("restored Lua roll = %v, want %v", got, want)
	}
}

func TestAuthorityRandomModuleRejectsUnregisteredPurpose(t *testing.T) {
	streams := simulation.NewRandomStreams(1)
	runtime := New()
	if err := runtime.RegisterModule(AuthorityRandomModule(streams)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	err := runtime.Run(ctx, func(state *lua.LState) error {
		return state.DoString(`require("dm.authority_random/v1").integer("typo", 10)`)
	})
	if err == nil {
		t.Fatal("unregistered random purpose was accepted")
	}
}
