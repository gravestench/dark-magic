package modruntime

import (
	"context"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameplayer "github.com/gravestench/dark-magic/internal/game/player"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/persistence"
	lua "github.com/yuin/gopher-lua"
)

func materializeGameplayPlayer(t *testing.T, engine *gameecs.Engine, step time.Duration) *gamesession.Session {
	t.Helper()
	session, err := gamesession.New(engine, gamesession.Config{Step: step})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	if err := gameplayer.Register(session); err != nil {
		t.Fatal(err)
	}
	if err := gamesession.RegisterMovement(session); err != nil {
		t.Fatal(err)
	}
	entry := gameplayer.EntryFromCharacter(persistence.Character{
		ID: "test-hero", Name: "Hero", Class: "Amazon", Level: 2,
		Stats: &persistence.Stats{Experience: 125, Health: 70, MaxHealth: 80, Mana: 30, MaxMana: 40},
	}, "test-player", 50, 40, 100, 80)
	command, err := gameplayer.Command(entry, "test-system", 1, 1, simulation.AuthoritySystem)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	return session
}

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
	session := materializeGameplayPlayer(t, engine, 500*time.Millisecond)
	scope := &Scope{}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`local world=require("darkmagic.gameplay.world"); gameplay=world.create(100,80,nil,"test-player")`)
	}); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{"right": {Down: true}}, Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})
	source, err := gamesession.NewMovementSource(engine, &input, "test-player", "game_world")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.AdvanceWithSource(500*time.Millisecond, source.Commands); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`
local world=require("darkmagic.gameplay.world")
hero_x,hero_y=world.position(gameplay.hero)
camera_x,camera_y=world.position(gameplay.camera)
hud=world.hud_snapshot(gameplay.hero,{next_level_experience=250,stamina=44,max_stamina=60})
composite=world.composite_snapshot(gameplay.hero)
`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		if state.GetGlobal("hero_x") != lua.LNumber(55) || state.GetGlobal("hero_y") != lua.LNumber(40) || state.GetGlobal("camera_x") != lua.LNumber(55) || state.GetGlobal("camera_y") != lua.LNumber(40) {
			t.Fatalf("hero/camera = %s,%s / %s,%s", state.GetGlobal("hero_x"), state.GetGlobal("hero_y"), state.GetGlobal("camera_x"), state.GetGlobal("camera_y"))
		}
		hud := state.GetGlobal("hud").(*lua.LTable)
		for name, want := range map[string]lua.LNumber{
			"health": 70, "max_health": 80, "mana": 30, "max_mana": 40,
			"experience": 125, "next_level_experience": 250, "stamina": 44, "max_stamina": 60,
		} {
			if got := hud.RawGetString(name); got != want {
				t.Fatalf("HUD %s = %s, want %s", name, got, want)
			}
		}
		if hud.RawGetString("running") != lua.LFalse {
			t.Fatalf("HUD running = %s, want false", hud.RawGetString("running"))
		}
		if hud.RawGetString("left_skill") != lua.LNumber(0) || hud.RawGetString("right_skill") != lua.LNumber(0) {
			t.Fatalf("HUD skills = %s/%s, want 0/0", hud.RawGetString("left_skill"), hud.RawGetString("right_skill"))
		}
		if hud.RawGetString("belt_capacity") != lua.LNumber(4) {
			t.Fatalf("HUD belt capacity = %s, want 4", hud.RawGetString("belt_capacity"))
		}
		composite := state.GetGlobal("composite").(*lua.LTable)
		if composite.RawGetString("token") != lua.LString("AM") || composite.RawGetString("mode") != lua.LString("WL") || composite.RawGetString("direction") != lua.LNumber(3) || composite.RawGetString("weapon_class") != lua.LString("HTH") {
			t.Fatalf("composite snapshot = %#v", composite)
		}
		beltSlots := hud.RawGetString("belt_slots").(*lua.LTable)
		if beltSlots.Len() != 16 || beltSlots.RawGetInt(16) != lua.LString("") {
			t.Fatalf("HUD belt slots = len %d, slot 16 %s", beltSlots.Len(), beltSlots.RawGetInt(16))
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
	session := materializeGameplayPlayer(t, engine, 30*time.Millisecond)
	scope := &Scope{}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`
local world=require("darkmagic.gameplay.world")
-- The center advances only to x=50.3, but its radius reaches collision cell
-- 51. A point-only query would clip into the wall; the footprint must stop X.
local collision={blocked=function(_,x,_) return x >= 51 end}
gameplay=world.create(100,80,collision,"test-player")
`)
	}); err != nil {
		t.Fatal(err)
	}
	input.Publish(inputstate.Frame{Actions: map[string]inputstate.ActionState{
		"right": {Down: true},
	}, Owner: inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "game_world"}})
	source, err := gamesession.NewMovementSource(engine, &input, "test-player", "game_world")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.AdvanceWithSource(time.Second, source.Commands); err != nil {
		t.Fatal(err)
	}
	if err := runtime.RunScoped(context.Background(), scope, func(state *lua.LState) error {
		return state.DoString(`local world=require("darkmagic.gameplay.world"); hero_x,hero_y=world.position(gameplay.hero)`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Run(context.Background(), func(state *lua.LState) error {
		x := float64(state.GetGlobal("hero_x").(lua.LNumber))
		y := float64(state.GetGlobal("hero_y").(lua.LNumber))
		if x != 50 || y != 40 {
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
