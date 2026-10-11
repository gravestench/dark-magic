package modruntime

import (
	"context"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

// TestAuthorityStateModuleRunsHeadlesslyAndUsesEngineOwnedState protects the authority state module runs headlessly
// and uses engine owned state contract, including its observable ordering and failure behavior.
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
	defer func() { _ = runtime.Stop(ctx) }()

	script := `
local state = require("engine.authority_state/v1")

-- The script decides that a counter should advance. The engine owns the bytes.
state.register("example.counter", "counter/v1", { value = 1 })
local current = state.read("example.counter")
state.replace("example.counter", "counter/v1", { value = current.value + 1 })
`

	if err := runtime.Run(
		ctx,
		func(state *lua.LState) error { return state.DoString(script) },
	); err != nil {
		t.Fatal(err)
	}

	got, found := stores.Read("example.counter")
	if !found || string(got.Data) != `{"value":2}` {
		t.Fatalf("registered state = %s, %v", got.Data, found)
	}
}

func TestAuthorityStateReadsStayIndependentAcrossReplaceAndRestore(t *testing.T) {
	stores := simulation.NewStateStore()
	runtime := New()
	if err := runtime.RegisterModule(AuthorityStateModule(stores)); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { runtime.Stop(context.Background()) })
	run := func(script string) {
		t.Helper()
		if err := runtime.Run(t.Context(), func(state *lua.LState) error { return state.DoString(script) }); err != nil {
			t.Fatal(err)
		}
	}
	run(`
local state = require("engine.authority_state/v1")
state.register("tree", "tree/v1", {nested={value=1}, array={1,2,3}, enabled=true, label="original"})
local first = state.read("tree")
first.nested.value, first.array[2], first.enabled, first.label = 99, 99, false, "changed"
local second = state.read("tree")
assert(second.nested.value == 1 and second.array[2] == 2 and second.enabled and second.label == "original")
assert(state.read("missing") == nil)
state.register("null", "scalar/v1", nil)
state.register("number", "scalar/v1", 17)
assert(state.read("null") == nil and state.read("number") == 17)
`)
	archive, err := stores.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	run(`local state=require("engine.authority_state/v1"); state.replace("tree", "tree/v1", {nested={value=2}}); assert(state.read("tree").nested.value == 2)`)
	if err := stores.RestoreState(archive); err != nil {
		t.Fatal(err)
	}
	run(`assert(require("engine.authority_state/v1").read("tree").nested.value == 1)`)
	if err := stores.Replace("tree", "tree/v1", []byte("invalid json")); err != nil {
		t.Fatal(err)
	}
	run(`assert(not pcall(require("engine.authority_state/v1").read, "tree"))`)
	if err := stores.RestoreState(archive); err != nil {
		t.Fatal(err)
	}
	run(`assert(require("engine.authority_state/v1").read("tree").array[3] == 3)`)
	if err := stores.Replace("tree", "tree/v1", []byte(strings.Repeat("[", 66)+"0"+strings.Repeat("]", 66))); err != nil {
		t.Fatal(err)
	}
	run(`local ok, err = pcall(require("engine.authority_state/v1").read, "tree"); assert(not ok and string.find(err, "convert registered state: state exceeds maximum nesting depth", 1, true))`)
	if err := stores.RestoreState(archive); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Stop(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	run(`assert(require("engine.authority_state/v1").read("tree").label == "original")`)
}

func BenchmarkAuthorityStateRead(b *testing.B) {
	for _, changed := range []bool{false, true} {
		name := "unchanged"
		if changed {
			name = "changed"
		}
		b.Run(name, func(b *testing.B) {
			entries := strings.TrimSuffix(strings.Repeat(`{"id":1,"position":{"x":12,"y":34},"tags":["alive","resident"]},`, 64), ",")
			payloads := [2][]byte{[]byte(`{"version":0,"entries":[` + entries + `]}`), []byte(`{"version":1,"entries":[` + entries + `]}`)}
			stores := simulation.NewStateStore()
			if err := stores.Register("benchmark", "benchmark/v1", payloads[0]); err != nil {
				b.Fatal(err)
			}
			runtime := New()
			if err := runtime.SetExecutionBudget(0); err != nil {
				b.Fatal(err)
			}
			if err := runtime.RegisterModule(AuthorityStateModule(stores)); err != nil {
				b.Fatal(err)
			}
			if err := runtime.Start(b.Context()); err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() { _ = runtime.Stop(context.Background()) })
			if err := runtime.Run(b.Context(), func(state *lua.LState) error {
				if err := state.DoString(`read_benchmark = require("engine.authority_state/v1").read`); err != nil {
					return err
				}
				read := lua.P{Fn: state.GetGlobal("read_benchmark"), NRet: 1, Protect: true}
				b.ReportAllocs()
				b.ResetTimer()
				for index := 0; b.Loop(); index++ {
					if changed {
						if err := stores.Replace("benchmark", "benchmark/v1", payloads[(index+1)%2]); err != nil {
							return err
						}
					}
					if err := state.CallByParam(read, lua.LString("benchmark")); err != nil {
						return err
					}
					state.Pop(1)
				}
				return nil
			}); err != nil {
				b.Fatal(err)
			}
		})
	}
}
