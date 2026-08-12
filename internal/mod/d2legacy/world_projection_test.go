package d2legacy_test

import (
	"context"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

func TestSemanticCuesTolerateUnregisteredOptionalEventSchemas(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	runtime := modruntime.New()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(modruntime.NewECSCapability(runtime, engine).Module()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)

	// Combat Lab can observe a melee event before missile/death event schemas
	// are installed by an authority composition. That is a valid empty optional
	// projection, not an unknown-component scene failure.
	script := `
local ecs = require("engine.ecs/v1")
ecs.component({name="d2legacy.combat.melee_event",fields={
  {name="kind",type="string"},
}})
ecs.create({["d2legacy.combat.melee_event"]={kind="hit_resolved"}})
local world = require("d2legacy.gameplay.world")
local cues = world.semantic_cues()
assert(#cues == 1)
assert(cues[1].cue_type == "combat")
assert(cues[1].kind == "hit_resolved")
`
	if err := runtime.RunScoped(ctx, &modruntime.Scope{}, func(state *lua.LState) error {
		return state.DoString(script)
	}); err != nil {
		t.Fatal(err)
	}
}
