package modruntime

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

type fireBoltRecords struct{}

func (fireBoltRecords) Invalidate(string)  {}
func (fireBoltRecords) Loaded(string) bool { return true }
func (fireBoltRecords) Load(path string) ([]map[string]string, error) {
	if path == "data/global/excel/skills.txt" {
		return []map[string]string{{
			"Id": "36", "skill": "Fire Bolt", "srvmissile": "firebolt",
			"etype": "fire", "interrupt": "1", "srvstfunc": "", "srvdofunc": "",
			"mana": "5", "manashift": "7", "emin": "3", "emax": "6", "HitShift": "8",
		}}, nil
	}
	return []map[string]string{{
		"Missile": "firebolt", "Skill": "Fire Bolt", "pSrvDoFunc": "1",
		"CollideType": "3", "CollideKill": "1", "Vel": "20", "Range": "40", "Size": "2",
	}}, nil
}

func TestD2LegacyFireBoltCastRunsHeadlesslyThroughLua(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(content.Shim(), "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []Module{
		NewECSCapability(runtime, engine).Module(),
		RecordsModule(fireBoltRecords{}),
		AuthorityCommandModule(runtime, session),
	} {
		if err := runtime.RegisterModule(module); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer runtime.Stop(ctx)

	scope := &Scope{}
	defer scope.Close()
	script := `
local ecs = require("dm.ecs/v1")
ecs.component({name="dm.player.identity",fields={
    {name="character_id",type="string"},{name="player",type="string"},
    {name="name",type="string"},{name="class",type="string"},
}})
ecs.component({name="dm.player.vitals",fields={
    {name="health",type="i64"},{name="max_health",type="i64"},
    {name="mana",type="i64"},{name="max_mana",type="i64"},
    {name="mana_raw",type="i64"},{name="max_mana_raw",type="i64"},
}})
ecs.component({name="dm.player.learned_skill",fields={
    {name="owner",type="entity"},{name="skill_id",type="i64"},
    {name="level",type="i64"},{name="list_row",type="i64"},
    {name="left_allowed",type="bool"},{name="right_allowed",type="bool"},
}})
player = ecs.create({
    ["dm.player.identity"]={character_id="hero",player="alice",name="Hero",class="Sorceress"},
    ["dm.player.vitals"]={health=50,max_health=50,mana=10,max_mana=10,mana_raw=2560,max_mana_raw=2560},
})
ecs.create({["dm.player.learned_skill"]={owner=player,skill_id=36,level=1,list_row=0,left_allowed=true,right_allowed=true}})
require("d2legacy.bootstrap.authoritative").start()
`
	if err := runtime.RunScoped(ctx, scope, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}

	payload, _ := json.Marshal(map[string]any{"skill_id": 36, "skill_level": 1, "target_x": 8, "target_y": 4})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "alice", Sequence: 1, Kind: "d2legacy.skill.cast", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Run(ctx, func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("dm.ecs/v1")
local cast=assert(ecs.get(player,"d2legacy.skill.cast"))
local vitals=assert(ecs.get(player,"dm.player.vitals"))
assert(cast:get("skill_id")==36 and cast:get("effect_tick")==2)
assert(vitals:get("mana_raw")==1920 and vitals:get("mana")==7)
`)
	}); err != nil {
		t.Fatal(err)
	}
}
