package d2legacy_test

import (
	"context"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// This test protects the ownership boundary as well as the behavior.  The only
// native capability installed here is generic ECS access; all component
// vocabulary and pointer-selection policy are loaded from the d2legacy mod.
func TestD2LegacyLuaSelectsLiveEntitiesByPriorityDistanceAndStableID(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
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

	script := `
local ecs = require("engine.ecs/v1")
ecs.component({name="d2legacy.world.position", fields={
  {name="x",type="f64"},{name="y",type="f64"},
}})
ecs.component({name="d2legacy.world.selectable", fields={
  {name="id",type="string"},{name="kind",type="string"},{name="label",type="string"},
  {name="owner",type="string"},{name="radius",type="f64"},{name="priority",type="i64"},
}})
local function spawn(id, kind, x, priority)
  ecs.create({
    ["d2legacy.world.position"]={x=x,y=10},
    ["d2legacy.world.selectable"]={id=id,kind=kind,label=id,owner="owner:"..id,radius=3,priority=priority},
  })
end
spawn("near", "npc", 10, 0)
spawn("far", "npc", 12, 0)
spawn("item", "item", 12, 5)
spawn("distance-near", "npc", 20, 0)
spawn("distance-far", "npc", 22, 0)
spawn("stable-b", "npc", 30, 0)
spawn("stable-a", "npc", 30, 0)
local targeting=require("d2legacy.gameplay.targeting")
local hit=targeting.selectable_at(10,10)
assert(hit and hit.id=="item" and hit.kind=="item" and hit.label=="item")
assert(hit.owner=="owner:item" and hit.x==12 and hit.y==10 and hit.radius==3)
assert(targeting.selectable_at(20,10).id=="distance-near")
assert(targeting.selectable_at(30,10).id=="stable-a")
assert(targeting.selectable_at(100,100)==nil)
assert(targeting.selectable_at(0/0,10)==nil)
`
	if err := runtime.RunScoped(ctx, &modruntime.Scope{}, func(state *lua.LState) error {
		return state.DoString(script)
	}); err != nil {
		t.Fatal(err)
	}

	// A test-only native lookup proves the script registered schemas into the
	// same authoritative ECS rather than evaluating a disconnected Lua fixture.
	if _, found := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable"); !found {
		t.Fatal("d2legacy selectable schema was not registered in authoritative ECS")
	}
}
