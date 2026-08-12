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
		"CelFile": "firebolt", "AnimSpeed": "16", "NumDirections": "16",
		"LoopAnim": "1", "TravelSound": "firebolt", "HitSound": "fireboltimpact",
		"Xoffset": "1", "Yoffset": "2", "Zoffset": "3",
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

	streams := simulation.NewRandomStreams(42)
	if err := streams.Register("d2legacy.combat.fire_bolt.damage"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"d2legacy.combat.basic_melee.hit", "d2legacy.combat.basic_melee.damage"} {
		if err := streams.Register(name); err != nil {
			t.Fatal(err)
		}
	}
	if err := session.RegisterStateParticipant(streams); err != nil {
		t.Fatal(err)
	}

	runtime := New()
	if err := runtime.RegisterInstaller(ContentRequire(content.D2Legacy(), "lua")); err != nil {
		t.Fatal(err)
	}
	for _, module := range []Module{
		NewECSCapability(runtime, engine).Module(),
		RecordsModule(fireBoltRecords{}),
		AuthorityCommandModule(runtime, session),
		AuthorityRandomModule(streams),
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
local ecs = require("engine.ecs/v1")
ecs.component({name="d2legacy.player.identity",fields={
    {name="character_id",type="string"},{name="player",type="string"},
    {name="name",type="string"},{name="class",type="string"},
}})
ecs.component({name="d2legacy.player.vitals",fields={
    {name="health",type="i64"},{name="max_health",type="i64"},
    {name="mana",type="i64"},{name="max_mana",type="i64"},
    {name="mana_raw",type="i64"},{name="max_mana_raw",type="i64"},
}})
ecs.component({name="d2legacy.player.learned_skill",fields={
    {name="owner",type="entity"},{name="skill_id",type="i64"},
    {name="level",type="i64"},{name="list_row",type="i64"},
    {name="left_allowed",type="bool"},{name="right_allowed",type="bool"},
}})
ecs.component({name="d2legacy.player.skill_assignment",fields={{name="left",type="i64"},{name="right",type="i64"}}})
ecs.component({name="d2legacy.world.position",fields={{name="x",type="f64"},{name="y",type="f64"}}})
ecs.component({name="d2legacy.world.location",fields={{name="act",type="i64"},{name="level_id",type="i64"}}})
ecs.component({name="d2legacy.world.collider",fields={{name="radius",type="f64"}}})
ecs.component({name="d2legacy.world.selectable",fields={
    {name="id",type="string"},{name="kind",type="string"},{name="label",type="string"},
    {name="owner",type="string"},{name="radius",type="f64"},{name="priority",type="i64"},
}})
ecs.component({name="d2legacy.monster.stats",fields={
    {name="level",type="i64"},{name="health",type="i64"},{name="max_health",type="i64"},
    {name="defense",type="i64"},{name="attack_rating",type="i64"},
    {name="physical_min",type="i64"},{name="physical_max",type="i64"},{name="experience",type="i64"},
}})
player = ecs.create({
    ["d2legacy.player.identity"]={character_id="hero",player="alice",name="Hero",class="Sorceress"},
    ["d2legacy.player.vitals"]={health=50,max_health=50,mana=10,max_mana=10,mana_raw=2560,max_mana_raw=2560},
    ["d2legacy.player.skill_assignment"]={left=36,right=36},
    ["d2legacy.world.position"]={x=0,y=0}, ["d2legacy.world.location"]={act=1,level_id=1},
    ["d2legacy.world.collider"]={radius=0.5},
    ["d2legacy.world.selectable"]={id="player:alice",kind="player",label="Hero",owner="alice",radius=0.5,priority=1},
})
ecs.create({["d2legacy.player.learned_skill"]={owner=player,skill_id=36,level=1,list_row=0,left_allowed=true,right_allowed=true}})
monster = ecs.create({
    ["d2legacy.monster.stats"]={level=1,health=4096,max_health=4096,defense=0,attack_rating=0,physical_min=0,physical_max=0,experience=0},
    ["d2legacy.world.position"]={x=4,y=0}, ["d2legacy.world.location"]={act=1,level_id=1},
    ["d2legacy.world.collider"]={radius=0.5},
    ["d2legacy.world.selectable"]={id="monster:fallen",kind="hostile",label="Fallen",owner="",radius=0.5,priority=1},
})
require("d2legacy.authoritative").start()
ecs.set(player,"d2legacy.combat.melee_profile",{range=5,physical_min=256,physical_max=256})
require("d2legacy.policy.melee").temporary_hit_chance=100
`
	if err := runtime.RunScoped(ctx, scope, func(state *lua.LState) error { return state.DoString(script) }); err != nil {
		t.Fatal(err)
	}

	payload, _ := json.Marshal(map[string]any{"side": "left", "target_x": 8, "target_y": 0})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "alice", Sequence: 1, Kind: "player.use_skill", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	for range 6 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	if err := runtime.Run(ctx, func(state *lua.LState) error {
		return state.DoString(`local ecs=require("engine.ecs/v1"); ecs.set(player,"d2legacy.combat.basic_attack_request",{target_id="monster:fallen",request_tick=7})`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Run(ctx, func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("engine.ecs/v1")
local vitals=assert(ecs.get(player,"d2legacy.player.vitals"))
assert(vitals:get("mana_raw")==1920 and vitals:get("mana")==7)
local monster_stats=assert(ecs.get(monster,"d2legacy.monster.stats"))
assert(monster_stats:get("health") < 4096 and monster_stats:get("health") >= 2560)
assert(#ecs.query({all={"d2legacy.missile.projectile"}})==0)
local events=ecs.query({all={"d2legacy.combat.event"}})
assert(#events==1)
local event=ecs.get(events[1],"d2legacy.combat.event")
assert(event:get("target_id")=="monster:fallen")
assert(event:get("damage_channel")=="fire")
local melee=ecs.query({all={"d2legacy.combat.melee_event"}})
assert(#melee==1)
local melee_event=ecs.get(melee[1],"d2legacy.combat.melee_event")
assert(melee_event:get("hit") and melee_event:get("damage_raw")==256)
`)
	}); err != nil {
		t.Fatal(err)
	}
}
