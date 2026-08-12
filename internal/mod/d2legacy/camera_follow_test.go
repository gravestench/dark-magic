package d2legacy_test

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

func TestCameraFollowDefaultsInstantAndSupportsParameterizedEasing(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	runtime := modruntime.New()
	if err := runtime.RegisterInstaller(modruntime.ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(modruntime.NewECSCapability(runtime, engine).Module()); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(t.Context())
	script := `local ecs=require("engine.ecs/v1")
require("d2legacy.gameplay.components.world_spatial").register()
require("d2legacy.gameplay.systems.camera_follow").register()
target=ecs.create({["d2legacy.world.position"]={x=0,y=0}})
instant=ecs.create({["d2legacy.world.position"]={x=0,y=0},["d2legacy.world.camera_follow"]={
 target=target,strategy="instant",duration=0,param_1=0,param_2=0,param_3=0,
 origin_x=0,origin_y=0,destination_x=0,destination_y=0,elapsed=0}})
eased=ecs.create({["d2legacy.world.position"]={x=0,y=0},["d2legacy.world.camera_follow"]={
 target=target,strategy="cubic_out",duration=1,param_1=0,param_2=0,param_3=0,
 origin_x=0,origin_y=0,destination_x=0,destination_y=0,elapsed=0}})
ecs.get(target,"d2legacy.world.position"):set("x",10)`
	scope := &modruntime.Scope{}
	defer scope.Close()
	if err := runtime.RunScoped(t.Context(), scope, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(500 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	instant, _ := positions.Get(akara.Entity(2))
	eased, _ := positions.Get(akara.Entity(3))
	instantX, _ := instant.Get("x")
	easedX, _ := eased.Get("x")
	if instantX != float64(10) || easedX != float64(8.75) {
		t.Fatalf("camera x instant/eased = %v/%v, want 10/8.75", instantX, easedX)
	}
	if err := engine.Update(500 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	easedX, _ = eased.Get("x")
	if easedX != float64(10) {
		t.Fatalf("completed eased camera x = %v, want 10", easedX)
	}
}
