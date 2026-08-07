package modruntime

import (
	"context"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/inputstate"
	lua "github.com/yuin/gopher-lua"
)

func TestShimWorldGameplayUsesLuaDefinedECSSystems(t *testing.T) {
	var input inputstate.Store
	engine := gameecs.New()
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(content.Shim(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(InputModule(&input)); err != nil {
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
		return state.DoString(`local world=require("darkmagic.gameplay.world"); gameplay=world.create(100,80)`)
	}); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"right": {Down: true}}})
	if err := engine.Update(500 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`local world=require("darkmagic.gameplay.world"); hero_x,hero_y=world.position(gameplay.hero); camera_x,camera_y=world.position(gameplay.camera)`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("hero_x") != lua.LNumber(55) || state.GetGlobal("hero_y") != lua.LNumber(40) || state.GetGlobal("camera_x") != lua.LNumber(55) || state.GetGlobal("camera_y") != lua.LNumber(40) {
			t.Fatalf("hero/camera = %s,%s / %s,%s", state.GetGlobal("hero_x"), state.GetGlobal("hero_y"), state.GetGlobal("camera_x"), state.GetGlobal("camera_y"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`require("darkmagic.gameplay.world").destroy(gameplay)`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}
	if len(engine.Systems()) != 0 {
		t.Fatalf("systems leaked after scope close: %v", engine.Systems())
	}
}

func TestShimWorldGameplayRejectsBlockedMovement(t *testing.T) {
	var input inputstate.Store
	engine := gameecs.New()
	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(content.Shim(), "lua")); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterModule(InputModule(&input)); err != nil {
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
local world=require("darkmagic.gameplay.world")
local collision={blocked=function(_,x,_) return x >= 60 end}
gameplay=world.create(100,80,collision)
`)
	}); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{
		"right": {Down: true},
		"down":  {Down: true},
	}})
	if err := engine.Update(time.Second); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`local world=require("darkmagic.gameplay.world"); hero_x,hero_y=world.position(gameplay.hero)`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("hero_x") != lua.LNumber(50) || state.GetGlobal("hero_y") != lua.LNumber(50) {
			t.Fatalf("hero = %s,%s", state.GetGlobal("hero_x"), state.GetGlobal("hero_y"))
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`require("darkmagic.gameplay.world").destroy(gameplay)`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := scope.Close(); err != nil {
		t.Fatal(err)
	}
}
