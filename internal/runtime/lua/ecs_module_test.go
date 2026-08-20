package modruntime

import (
	"context"
	"strings"
	"testing"
	"time"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	lua "github.com/yuin/gopher-lua"
)

// newECSRuntime constructs ecsruntime with its dependencies in one place, keeping ownership and cleanup
// explicit.
func newECSRuntime(t *testing.T) (*Runtime, *gameecs.Engine) {
	t.Helper()

	runtime := New()

	engine := gameecs.New()
	if err := runtime.RegisterModule(NewECSCapability(runtime, engine).Module()); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = runtime.Stop(context.Background())
		_ = engine.Close()
	})

	return runtime, engine
}

// TestLuaDefinesComponentsEntitiesAndMovementSystem verifies schema registration, entity access, deterministic
// updates, and removal of scope-owned systems.
func TestLuaDefinesComponentsEntitiesAndMovementSystem(t *testing.T) {
	runtime, engine := newECSRuntime(t)
	scope := &Scope{}

	t.Cleanup(func() { _ = scope.Close() })

	err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`
local ecs = require("engine.ecs/v1")
ecs.component{name="position", version=1, fields={{name="x",type="f64"},{name="y",type="f64"}}}
ecs.component{name="velocity", version=1, fields={{name="x",type="f64"},{name="y",type="f64"}}}
player = ecs.create{position={x=10,y=5}, velocity={x=4,y=-2}}
ecs.system{
  id="movement.integrate", phase="movement",
  query={all={"position","velocity"}},
  read={"velocity"}, write={"position"},
  update=function(ctx, entities, commands)
    for _, entity in ipairs(entities) do
      local p, v = ecs.get(entity,"position"), ecs.get(entity,"velocity")
      p:set("x", p:get("x") + v:get("x") * ctx.delta_seconds)
      p:set("y", p:get("y") + v:get("y") * ctx.delta_seconds)
    end
  end,
}
`)
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.Update(500 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	err = runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(
			`local ecs=require("engine.ecs/v1"); local p=ecs.get(player,"position"); assert(p:get("x")==12 and p:get("y")==4)`,
		)
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}

	if len(engine.Systems()) != 0 {
		t.Fatalf("scope leaked systems: %v", engine.Systems())
	}
}

// TestLuaCanQueryAnOrderedEntitySnapshotOutsideSystems verifies filtering and stable entity order for ad hoc Lua
// queries outside engine-managed system iteration.
func TestLuaCanQueryAnOrderedEntitySnapshotOutsideSystems(t *testing.T) {
	runtime, _ := newECSRuntime(t)

	err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("engine.ecs/v1")
ecs.component{name="position",fields={}}
ecs.component{name="hidden",fields={}}
local first=ecs.create{position={}}
ecs.create{position={},hidden={}}
local entities=ecs.query{all={"position"},none={"hidden"}}
assert(#entities==1 and entities[1]:id()==first:id())
`)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestLuaStructuralCommandsApplyBetweenSystems verifies that intent-phase structural writes become visible to the
// following movement phase in the same update.
func TestLuaStructuralCommandsApplyBetweenSystems(t *testing.T) {
	runtime, engine := newECSRuntime(t)
	scope := &Scope{}

	t.Cleanup(func() { _ = scope.Close() })

	err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("engine.ecs/v1")
ecs.component{name="spawn",fields={}}
entity=ecs.create()
ecs.system{id="add",phase="intent",write={"spawn"},update=function(ctx,entities,commands) commands` +
			`:set(entity,"spawn",{}) end}
ecs.system{id="observe",phase="movement",query={all={"spawn"}},read={"spawn"},` +
			`update=function(ctx,entities,commands) observed=#entities end}
`)
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}

	err = runtime.Run(
		context.Background(),
		func(state *lua.LState) error { return state.DoString(`assert(observed==1)`) },
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestLuaSystemCanSpawnThroughCommandBuffer verifies that command-buffer creation becomes queryable by a later
// phase without mutating ECS structure during iteration.
func TestLuaSystemCanSpawnThroughCommandBuffer(t *testing.T) {
	runtime, engine := newECSRuntime(t)
	scope := &Scope{}

	t.Cleanup(func() { _ = scope.Close() })

	err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("engine.ecs/v1")
ecs.component{name="monster",fields={{name="kind",type="string"}}}
ecs.system{id="spawn",phase="intent",write={"monster"},update=function(ctx,entities,commands) commands` +
			`:create{monster={kind="fallen"}} end}
ecs.system{id="count",phase="movement",query={all={"monster"}},read={"monster"},` +
			`update=function(ctx,entities,commands) monster_count=#entities end}
`)
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}

	err = runtime.Run(
		context.Background(),
		func(state *lua.LState) error { return state.DoString(`assert(monster_count==1)`) },
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestLuaSystemAccessDeclarationsAreEnforced keeps undeclared component reads from bypassing system access policy.
func TestLuaSystemAccessDeclarationsAreEnforced(t *testing.T) {
	runtime, engine := newECSRuntime(t)
	scope := &Scope{}

	t.Cleanup(func() { _ = scope.Close() })

	err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("engine.ecs/v1")
ecs.component{name="secret",fields={{name="value",type="string"}}}
entity=ecs.create{secret={value="hidden"}}
ecs.system{id="bad",phase="effects",update=function(ctx,entities,commands) ecs.get(entity,"secret") end}
`)
	})
	if err != nil {
		t.Fatal(err)
	}

	err = engine.Update(time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "did not declare read access") {
		t.Fatalf("access error = %v", err)
	}
}

// TestLuaComponentSchemaMigrationPreservesEntities verifies that upgrading a schema migrates existing values and
// supplies the new field without replacing the entity.
func TestLuaComponentSchemaMigrationPreservesEntities(t *testing.T) {
	runtime, _ := newECSRuntime(t)
	scope := &Scope{}

	t.Cleanup(func() { _ = scope.Close() })

	err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("engine.ecs/v1")
ecs.component{name="counter",version=1,fields={{name="value",type="i64"}}}
entity=ecs.create{counter={value=7}}
ecs.component{name="counter",version=2,fields={{name="value",type="i64"},` +
			`{name="step",type="i64",default=1}},migrate=function(old,entity) ` +
			`return {value=old.value,step=2} end}
local c=ecs.get(entity,"counter"); assert(c:get("value")==7 and c:get("step")==2)
`)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestLuaRejectsEntityFromAnotherWorld keeps handles tied to the engine that created them after an engine rebind.
func TestLuaRejectsEntityFromAnotherWorld(t *testing.T) {
	state := lua.NewState()
	defer state.Close()

	registerECSTypes(state)

	first := NewECSCapability(nil, gameecs.New())
	second := NewECSCapability(nil, gameecs.New())

	defer func() { _ = first.engine.Close() }()
	defer func() { _ = second.engine.Close() }()

	entity := first.engine.World().MustCreateEntity()
	function := state.NewFunction(second.destroyEntity)

	err := state.CallByParam(
		lua.P{Fn: function, NRet: 0, Protect: true},
		first.entityValue(state, entity),
	)
	if err == nil || !strings.Contains(err.Error(), "different engine.ecs/v1 world") {
		t.Fatalf("cross-world entity error = %v", err)
	}
}
