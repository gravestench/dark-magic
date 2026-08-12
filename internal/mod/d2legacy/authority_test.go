package d2legacy

import (
	"context"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

// authorityFixture boots exactly the same renderer-free d2legacy composition
// root used by the standalone server. Tests may substitute tiny immutable
// record tables, but they do not redeclare components, systems, or commands.
type authorityFixture struct {
	engine    *gameecs.Engine
	session   *gamesession.Session
	authority *Authority
}

func newAuthorityFixture(
	t *testing.T,
	records Records,
	initial map[string]any,
) *authorityFixture {
	t.Helper()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{
		CheckpointInterval: 1,
	})
	if err != nil {
		engine.Close()
		t.Fatal(err)
	}
	authority, err := StartWithConfig(
		t.Context(),
		content.D2Legacy(),
		records,
		engine,
		session,
		Config{Seed: 42, InitialData: initial},
	)
	if err != nil {
		session.Close()
		engine.Close()
		t.Fatal(err)
	}
	fixture := &authorityFixture{
		engine:    engine,
		session:   session,
		authority: authority,
	}
	t.Cleanup(func() {
		_ = authority.Stop(context.Background())
		_ = session.Close()
		_ = engine.Close()
	})
	return fixture
}

func (fixture *authorityFixture) submit(
	t *testing.T,
	tick uint64,
	sequence uint64,
	player string,
	kind string,
	payload string,
) {
	t.Helper()
	err := fixture.session.Submit(simulation.Command{
		Tick:      tick,
		Player:    player,
		Authority: simulation.AuthorityPlayer,
		Sequence:  sequence,
		Kind:      kind,
		Payload:   []byte(payload),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func (fixture *authorityFixture) submitSystem(
	t *testing.T,
	tick uint64,
	sequence uint64,
	kind string,
	payload string,
) {
	t.Helper()
	err := fixture.session.Submit(simulation.Command{
		Tick:      tick,
		Player:    "system:test",
		Authority: simulation.AuthoritySystem,
		Sequence:  sequence,
		Kind:      kind,
		Payload:   []byte(payload),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func (fixture *authorityFixture) step(t *testing.T) {
	t.Helper()
	if err := fixture.session.Step(); err != nil {
		t.Fatal(err)
	}
}

func (fixture *authorityFixture) run(t *testing.T, source string) {
	t.Helper()
	err := fixture.authority.Runtime.Run(
		t.Context(),
		func(state *lua.LState) error { return state.DoString(source) },
	)
	if err != nil {
		t.Fatal(err)
	}
}

// fixtureRecords is a small policy-neutral record reader. Unmentioned files
// return an empty table, which is enough for tests that do not exercise them.
type fixtureRecords map[string][]map[string]string

func (records fixtureRecords) Load(path string) ([]map[string]string, error) {
	if rows, found := records[path]; found {
		return rows, nil
	}
	switch path {
	case "data/global/excel/charstats.txt":
		return []map[string]string{{"class": "Amazon", "StartSkill": "Fire Bolt"}}, nil
	case "data/global/excel/skills.txt":
		return []map[string]string{{
			"Id": "36", "skill": "Fire Bolt", "srvmissile": "firebolt",
			"skilldesc": "firebolt", "leftskill": "1", "general": "0", "passive": "0",
			"etype": "fire", "interrupt": "1", "srvstfunc": "",
			"srvdofunc": "", "mana": "5", "manashift": "7",
			"emin": "3", "emax": "6", "HitShift": "8",
		}}, nil
	case "data/global/excel/skilldesc.txt":
		return []map[string]string{{"skilldesc": "firebolt", "ListRow": "0", "IconCel": "0"}}, nil
	case "data/global/excel/Missiles.txt":
		return []map[string]string{{
			"Missile": "firebolt", "Skill": "Fire Bolt",
			"pSrvDoFunc": "1", "CollideType": "3", "CollideKill": "1",
			"Vel": "20", "Range": "40", "Size": "2",
			"CelFile": "firebolt",
		}}, nil
	default:
		return nil, nil
	}
}

func (fixtureRecords) Invalidate(string)  {}
func (fixtureRecords) Loaded(string) bool { return true }

func TestHeadlessD2LegacyBootsWithoutRendererOrClientStartup(t *testing.T) {
	fixture := newAuthorityFixture(t, fixtureRecords{}, nil)
	fixture.run(t, `
local ecs = require("engine.ecs/v1")
assert(type(require("d2legacy.authoritative")) == "table")
assert(type(ecs.query({all={"d2legacy.player.identity"}})) == "table")
`)
}

func TestD2LegacyItemMovementAndHeldSwapAreAuthoritative(t *testing.T) {
	initial := map[string]any{
		"d2legacy.items": map[string]any{
			"owner":            "alice",
			"inventory_width":  4.0,
			"inventory_height": 4.0,
			"belt_capacity":    4.0,
			"items": []any{
				map[string]any{
					"id": "held", "code": "ssd",
					"width": 2.0, "height": 1.0,
					"container": "held",
				},
				map[string]any{
					"id": "placed", "code": "cap",
					"width": 2.0, "height": 1.0,
					"container": "inventory",
					"x":         1.0, "y": 1.0,
				},
			},
		},
		"d2legacy.interactions": map[string]any{"owner": "alice"},
	}
	fixture := newAuthorityFixture(t, fixtureRecords{}, initial)
	fixture.submit(
		t,
		1,
		1,
		"alice",
		"item.move",
		`{"item_id":"held","place_held":true,`+
			`"destination":{"container":"inventory","x":1,"y":1}}`,
	)
	fixture.step(t)
	fixture.run(t, `
local ecs = require("engine.ecs/v1")
local found = {}
for _, entity in ipairs(ecs.query({all={
    "d2legacy.item.identity",
    "d2legacy.item.placement",
}})) do
    local item = ecs.get(entity, "d2legacy.item.identity")
    local placed = ecs.get(entity, "d2legacy.item.placement")
    found[item:get("id")] = placed:get("container")
end
assert(found.held == "inventory")
assert(found.placed == "held")
`)
}

func TestD2LegacyPlayerEntryAndMovementRunThroughLuaAuthority(t *testing.T) {
	fixture := newAuthorityFixture(t, fixtureRecords{}, nil)
	fixture.submitSystem(t, 1, 1, "system.player.enter", `{
"character_id":"hero","player":"alice","name":"Hero","class":"Amazon",
"level":1,"experience":0,"dexterity":20,"defense":0,
"health":50,"max_health":50,"mana":20,"max_mana":20,
"expansion":true,"hardcore":false,"cof":"","palette":"units",
"direction":0,"mode":"NU","x":10,"y":12,
"world_width":100,"world_height":80,"act":1,"level_id":1,"skills":[]
}`)
	fixture.step(t)
	fixture.submit(
		t,
		2,
		1,
		"alice",
		"player.move",
		`{"x":1,"y":1,"running":true}`,
	)
	fixture.step(t)
	fixture.run(t, `
local ecs = require("engine.ecs/v1")
local players = ecs.query({all={
    "d2legacy.player.identity",
    "d2legacy.world.velocity",
    "d2legacy.player.animation",
}})
assert(#players == 1)
local velocity = ecs.get(players[1], "d2legacy.world.velocity")
local animation = ecs.get(players[1], "d2legacy.player.animation")
local facing = ecs.get(players[1], "d2legacy.world.facing")
local mode = ecs.get(players[1], "d2legacy.player.movement_mode")
local expected = 15 * 0.7071067811865476
assert(math.abs(velocity:get("x") - expected) < 0.000000001)
assert(math.abs(velocity:get("y") - expected) < 0.000000001)
assert(animation:get("mode") == "RN" and facing:get("direction") == 4)
assert(mode:get("running") == true)
`)
	fixture.submit(
		t,
		3,
		2,
		"alice",
		"player.move",
		`{"x":0,"y":0,"running":false,"target":{"x":20,"y":16}}`,
	)
	fixture.step(t)
	fixture.run(t, `
local ecs = require("engine.ecs/v1")
local player = ecs.query({all={"d2legacy.player.identity"}})[1]
local animation = ecs.get(player, "d2legacy.player.animation")
local facing = ecs.get(player, "d2legacy.world.facing")
assert(animation:get("mode") == "WL" and facing:get("direction") == 15)
`)
}

func TestD2LegacyDerivedMitigationAndLethalDamageVectors(t *testing.T) {
	fixture := newAuthorityFixture(t, fixtureRecords{}, nil)
	fixture.submitSystem(t, 1, 1, "system.player.enter", `{
"character_id":"hero","player":"alice","name":"Hero","class":"Amazon",
"level":1,"experience":0,"dexterity":20,"defense":0,"fire_resistance":20,
"health":10,"max_health":10,"mana":20,"max_mana":20,
"expansion":true,"hardcore":false,"cof":"","palette":"units",
"direction":0,"mode":"NU","x":10,"y":12,
"world_width":100,"world_height":80,"act":1,"level_id":1,"skills":[]
}`)
	fixture.step(t)
	fixture.run(t, `
local ecs=require("engine.ecs/v1")
local player=ecs.query({all={"d2legacy.player.identity"}})[1]
ecs.create({["d2legacy.stat.source"]={target=player,source_id="shield:fire",stat="fire_resist",value=30}})
ecs.create({["d2legacy.stat.source"]={target=player,source_id="armor:physical",stat="physical_resist",value=25}})
ecs.create({["d2legacy.stat.source"]={target=player,source_id="armor:flat",stat="physical_reduction_raw",value=20}})
`)
	fixture.step(t)
	fixture.run(t, `
local ecs=require("engine.ecs/v1")
local mitigation=require("d2legacy.policy.mitigation")
local damage=require("d2legacy.policy.damage")
local player=ecs.query({all={"d2legacy.player.identity"}})[1]
local defense=ecs.get(player,"d2legacy.combat.defense")
assert(defense:get("fire_resist")==50 and defense:get("physical_resist")==25)
assert(mitigation.apply(1000,"fire",defense)==500)
assert(mitigation.apply(1000,"physical",defense)==730)
local remaining,lethal,applied=damage.apply(player,4096,ecs,"fire")
assert(applied==2048 and remaining==512 and not lethal)
remaining,lethal,applied=damage.apply(player,1024,ecs,"fire")
assert(applied==512 and remaining==0 and lethal)
`)
}

func TestD2LegacyWorldTransitionRunsThroughLuaSystem(t *testing.T) {
	initial := map[string]any{
		"d2legacy.world_transitions": map[string]any{"seams": []any{
			map[string]any{
				"source_level": 1.0, "destination_level": 2.0,
				"source_x": 10.0, "source_y": 12.0,
				"arrival_x": 6.0, "arrival_y": 40.0,
				"world_width": 400.0, "world_height": 400.0,
			},
		}},
	}
	fixture := newAuthorityFixture(t, fixtureRecords{}, initial)
	fixture.submitSystem(t, 1, 1, "system.player.enter", `{
"character_id":"hero","player":"alice","name":"Hero","class":"Amazon",
"level":1,"experience":0,"dexterity":20,"defense":0,
"health":50,"max_health":50,"mana":20,"max_mana":20,
"expansion":true,"hardcore":false,"cof":"","palette":"units",
"direction":0,"mode":"NU","x":10,"y":12,
"world_width":100,"world_height":80,"act":1,"level_id":1,"skills":[]
}`)
	fixture.step(t)
	fixture.run(t, `
local ecs = require("engine.ecs/v1")
local player = ecs.query({all={"d2legacy.world.player_control"}})[1]
local location = ecs.get(player, "d2legacy.world.location")
local position = ecs.get(player, "d2legacy.world.position")
local bounds = ecs.get(player, "d2legacy.world.bounds")
assert(location:get("level_id") == 2)
assert(position:get("x") == 6 and position:get("y") == 40)
assert(bounds:get("width") == 400 and bounds:get("height") == 400)
`)
}

func TestD2LegacyOwnedUnitLimitsAndAttributionAreAuthoritative(t *testing.T) {
	fixture := newAuthorityFixture(t, fixtureRecords{}, nil)
	fixture.run(t, `
local ecs = require("engine.ecs/v1")
local function actor(id, kind)
    return ecs.create({
        ["d2legacy.world.selectable"]={
            id=id,kind=kind,label=id,owner="",radius=1,priority=1,
        },
    })
end
actor("player:hero", "player")
actor("monster:skeleton-one", "friendly")
actor("monster:skeleton-two", "friendly")
`)
	category := `"category":{"id":"skeleton","group":1,"base_max":1,` +
		`"replacement":"replace_oldest","warp_with_owner":true}`
	fixture.submitSystem(t, 1, 1, "system.owned_unit.attach", `{
"unit_id":"monster:skeleton-one","owner_id":"player:hero",
"ultimate_owner_id":"player:hero",`+category+`}`)
	fixture.step(t)
	fixture.submitSystem(t, 2, 2, "system.owned_unit.attach", `{
"unit_id":"monster:skeleton-two","owner_id":"player:hero",
"ultimate_owner_id":"player:hero",`+category+`}`)
	fixture.step(t)
	fixture.run(t, `
local ecs = require("engine.ecs/v1")
local attribution = require("d2legacy.owned_units.attribution")
local limits = require("d2legacy.owned_units.limits")
local active, inactive = 0, 0
for _, entity in ipairs(ecs.query({all={"d2legacy.owned_unit"}})) do
    local relation = ecs.get(entity, "d2legacy.owned_unit")
    if relation:get("active") then active = active + 1 else inactive = inactive + 1 end
    assert(relation:get("owner_id") == "player:hero")
    assert(relation:get("warp_with_owner") == true)
end
assert(active == 1 and inactive == 1)
local entities = ecs.query({all={"d2legacy.world.selectable"}})
local credit = attribution.resolve(entities, "monster:skeleton-two")
assert(credit.source_id == "monster:skeleton-two")
assert(credit.immediate_owner_id == "player:hero")
assert(credit.ultimate_owner_id == "player:hero")

-- The pure policy helper also proves the less common group-conflict and
-- newest-replacement branches without constructing unrelated world entities.
local function fake(id) return {id=function() return id end} end
local candidates = {
    {entity=fake(7),category="wolf",group=2,active=true,created_tick=1},
    {entity=fake(8),category="skeleton",group=1,active=true,created_tick=2},
    {entity=fake(9),category="skeleton",group=1,active=true,created_tick=3},
}
local victims = limits.victims(candidates, {
    id="skeleton",group=2,base_max=2,replacement="replace_newest",
})
assert(#victims == 2 and victims[1].entity:id() == 7 and victims[2].entity:id() == 9)
local ok = pcall(function()
    limits.victims(candidates, {
        id="skeleton",group=1,base_max=1,replacement="reject",
    })
end)
assert(not ok)
`)
	replay, err := fixture.session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	restored, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	store, found := akara.GetDynamicStore(restored.World(), "d2legacy.owned_unit")
	if !found || store.Len() != 2 {
		t.Fatalf("restored owned-unit relations = %d, want 2", store.Len())
	}
	active := 0
	for _, entity := range store.Entities() {
		value, _ := store.Get(entity)
		isActive, _ := value.Get("active")
		if isActive == true {
			active++
		}
	}
	if active != 1 {
		t.Fatalf("restored active owned units = %d, want 1", active)
	}
}

func TestNestedTreasureClassesKeepStrongestQualityModifiers(t *testing.T) {
	records := fixtureRecords{
		"data/global/excel/treasureclassex.txt": {
			{
				"Treasure Class": "root", "Picks": "-1",
				"Item1": "child", "Prob1": "1", "Unique": "200",
			},
			{
				"Treasure Class": "child", "Picks": "-1",
				"Item1": "rin", "Prob1": "1", "Unique": "100",
				"Set": "300",
			},
		},
	}
	fixture := newAuthorityFixture(t, records, nil)
	fixture.run(t, `
local treasure = require("d2legacy.loot.treasure_class")
local drops = treasure.roll("root")
assert(#drops == 1 and drops[1].code == "rin")
assert(drops[1].quality.unique == 200)
assert(drops[1].quality.set == 300)
local generate = require("d2legacy.loot.generate")
assert(generate.encode({}) == "[]")
`)
}

func TestLegacyLootQualityAffixAndPropertyVectors(t *testing.T) {
	records := fixtureRecords{
		"data/global/excel/itemratio.txt": {{
			"Version": "100", "Uber": "0", "Class Specific": "0",
			"Unique": "1000000000", "Set": "1000000000", "Rare": "1000000000",
			"Magic": "1000000000", "HiQuality": "1000000000", "Normal": "1000000000",
		}},
		"data/global/excel/magicprefix.txt": {{
			"Name": "Strong", "spawnable": "1", "frequency": "1", "level": "1",
			"version": "100", "itype1": "swor", "mod1code": "dmg",
			"mod1min": "7", "mod1max": "7",
		}},
		"data/global/excel/properties.txt": {{
			"code": "dmg", "func1": "1", "stat1": "item_maxdamage",
		}},
	}
	fixture := newAuthorityFixture(t, records, nil)
	fixture.run(t, `
local quality = require("d2legacy.loot.quality")
local affixes = require("d2legacy.loot.affixes")
local properties = require("d2legacy.loot.properties")
local sword = {type="swor",type2="",level=1,uber=false,class_specific=false}
local rolled = quality.roll(sword, {
    version=100,monster_level=1,magic_find=0,
}, {unique=0,set=0,rare=0,magic=1024})
assert(rolled == "magic")

local prefixes
for _ = 1, 16 do
    prefixes = affixes.roll(sword, rolled, 1, 100)
    if #prefixes == 1 then break end
end
assert(#prefixes == 1 and prefixes[1].name == "Strong")
assert(prefixes[1].modifiers[1].value == 7)
local stats, effects, unsupported = properties.apply(prefixes, {})
assert(#stats == 1 and stats[1].code == "item_maxdamage")
assert(stats[1].value == 7 and stats[1].fn == 1)
assert(#effects == 0 and #unsupported == 0)
`)
}

func TestGenericHostCanBootWithoutD2Legacy(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if engine.Tick() != 0 {
		t.Fatalf("fresh generic engine tick = %d", engine.Tick())
	}
}
