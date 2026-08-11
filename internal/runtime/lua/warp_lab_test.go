package modruntime

import (
	"context"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	lua "github.com/yuin/gopher-lua"
)

// This test deliberately excludes rendering. It proves that a pointer-facing
// fixture request becomes a fixed-tick ECS intent, approaches the selected
// portal, resolves its explicit entity pair, and publishes one atomic arrival.
func TestWarpLabIntentWalksToPortalAndArrivesAtPair(t *testing.T) {
	engine := gameecs.New()
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(content.Shim(), "lua")); err != nil {
		t.Fatal(err)
	}
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

	scope := &Scope{}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`
local fixture=require("darkmagic.dev.warp_lab.fixture")
warp_fixture_module=fixture
warp_fixture=fixture.create({x=10,y=0},{x=100,y=50},{x=0,y=0})
fixture.intent(warp_fixture,"warp-lab:a")
`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(time.Second); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("dm.ecs/v1")
local position=ecs.get(warp_fixture.player,"dm.world.position")
local status=ecs.get(warp_fixture.player,"dm.lab.warp.state")
local actor=ecs.get(warp_fixture.player,"dm.lab.warp.actor")
assert(position:get("x")==102 and position:get("y")==52)
assert(status:get("warp_count")==1)
assert(actor:get("direction")==3)
assert(ecs.get(warp_fixture.player,"dm.lab.warp.intent")==nil)
`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`warp_fixture_module.move(warp_fixture,110,52)`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(time.Second); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("dm.ecs/v1")
local position=ecs.get(warp_fixture.player,"dm.world.position")
local actor=ecs.get(warp_fixture.player,"dm.lab.warp.actor")
assert(position:get("x")==110 and position:get("y")==52)
assert(actor:get("direction")==3)
assert(ecs.get(warp_fixture.player,"dm.lab.warp.move_intent")==nil)
`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}
	if len(engine.Systems()) != 0 {
		t.Fatalf("Warp Lab leaked systems after scope close: %v", engine.Systems())
	}
}
