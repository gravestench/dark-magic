package acceptance

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// TestGenericHostBootsAnAlternateModWithoutD2Legacy proves the dependency
// boundary in executable form. This tiny mod uses the generic Lua runtime, ECS,
// deterministic RNG, and checkpointed state without mounting or importing any
// first-party d2legacy package.
func TestGenericHostBootsAnAlternateModWithoutD2Legacy(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	runtime := modruntime.New()
	state := simulation.NewStateStore()
	streams := simulation.NewRandomStreams(41)
	if err := streams.Register("example.boot"); err != nil {
		t.Fatal(err)
	}
	content := fstest.MapFS{
		"lua/example/boot.lua": {Data: []byte(`
local ecs = require("engine.ecs/v1")
local state = require("engine.authority_state/v1")
local random = require("engine.authority_random/v1")
state.register("example.counter", "example.counter/v1", {value=0})
local value = state.read("example.counter")
value.value = value.value + random.integer("example.boot", 10)
state.replace("example.counter", "example.counter/v1", value)
ecs.component({name="example.identity", fields={{name="name",type="string"}}})
local entity = ecs.create({["example.identity"]={name="independent"}})
return {entity=entity:id()}
`)},
	}
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(content, "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []modruntime.Module{
		modruntime.DeterministicModule(), modruntime.AuthorityStateModule(state),
		modruntime.AuthorityRandomModule(streams), modruntime.NewECSCapability(runtime, engine).Module(),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)
	if err := runtime.Run(ctx, func(vm *lua.LState) error {
		if err := vm.CallByParam(lua.P{Fn: vm.GetGlobal("require"), NRet: 1, Protect: true}, lua.LString("example.boot")); err != nil {
			return err
		}
		vm.Pop(1)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, found := state.Read("example.counter"); !found {
		t.Fatal("alternate mod state was not registered")
	}
	if _, found := akara.GetDynamicStore(engine.World(), "example.identity"); !found {
		t.Fatal("alternate mod ECS component was not registered")
	}
	if got := runtime.ModuleNames(); containsString(got, "d2legacy") {
		t.Fatalf("generic host loaded d2legacy capability: %v", got)
	}
}

func containsString(values []string, fragment string) bool {
	for _, value := range values {
		if value == fragment || len(value) >= len(fragment) && value[:len(fragment)] == fragment {
			return true
		}
	}
	return false
}
