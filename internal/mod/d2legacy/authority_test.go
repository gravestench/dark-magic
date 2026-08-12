package d2legacy

import (
	"context"
	"testing"

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
local mode = ecs.get(players[1], "d2legacy.player.movement_mode")
local expected = 15 * 0.7071067811865476
assert(math.abs(velocity:get("x") - expected) < 0.000000001)
assert(math.abs(velocity:get("y") - expected) < 0.000000001)
assert(animation:get("mode") == "RN" and animation:get("direction") == 4)
assert(mode:get("running") == true)
`)
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
