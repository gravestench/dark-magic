package d2legacy

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	lua "github.com/yuin/gopher-lua"
)

func TestAuthorityMaterializesPlayerEntryThroughLua(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(ctx, content.D2Legacy(), fixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)
	payload, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "attack_rating": 100, "defense": 20,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "", "token": "AM",
		"palette": "data/global/Palette/units/pal.dat", "direction": 0, "mode": "NU", "weapon_class": "HTH",
		"melee_range": 2, "physical_min_raw": 256, "physical_max_raw": 512,
		"x": 10, "y": 20, "world_width": 100, "world_height": 100, "act": 1, "level_id": 1,
		"skills": []map[string]any{{"id": 0, "level": 1, "list_row": 0, "left_allowed": true, "right_allowed": true}},
	})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	identities, found := akara.GetDynamicStore(engine.World(), "d2legacy.player.identity")
	if !found || len(identities.Entities()) != 1 {
		t.Fatalf("Lua entry created %d players", len(identities.Entities()))
	}
	player := identities.Entities()[0]
	assignments, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.skill_assignment")
	assignment, _ := assignments.Get(player)
	left, _ := assignment.Get("left")
	if left != int64(0) {
		t.Fatalf("initial left skill = %v, want basic Attack 0", left)
	}
	learned, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.learned_skill")
	if len(learned.Entities()) != 1 {
		t.Fatalf("Lua entry created %d learned skills", len(learned.Entities()))
	}
}

func TestAuthorityMonsterSpawnUsesCheckpointedLuaRandomStream(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(ctx, content.D2Legacy(), fixtureRecords{}, engine, session, 99)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)
	payload, _ := json.Marshal(map[string]any{
		"spawn_id": "fallen-1", "seed": 123, "x": 10, "y": 20, "act": 1, "level_id": 2,
		"definition": map[string]any{
			"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen",
			"ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
			"life_min": 256, "life_max": 768, "level": 1, "defense": 0, "attack_rating": 0,
			"physical_min": 256, "physical_max": 256, "experience": 5, "treasure_class": "Act 1 H2H A",
			"collider_radius": 1, "select_radius": 1, "velocity": 5, "think_interval": 1,
			"aggro_radius": 20, "attack_range": 1,
		},
	})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.monster.spawn", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	stats, found := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")
	if !found || stats.Len() != 1 {
		t.Fatalf("Lua spawn created %d monster stats", stats.Len())
	}
	value, _ := stats.Get(stats.Entities()[0])
	health, _ := value.Get("health")
	if health != int64(256) && health != int64(512) && health != int64(768) {
		t.Fatalf("spawned health = %v, want authored whole point", health)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if len(replay.InitialParticipants) != 3 {
		t.Fatalf("participant states = %d, want identity, Lua state, and RNG", len(replay.InitialParticipants))
	}
}

func TestAuthorityRunsTimedStateLifecycleThroughLua(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(ctx, content.D2Legacy(), fixtureRecords{}, engine, session, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)
	if err := authority.Runtime.Run(ctx, func(state *lua.LState) error {
		return state.DoString(`
local ecs=require("engine.ecs/v1")
timed_target=ecs.create()
ecs.create({["d2legacy.state.request"]={operation="apply",target=timed_target,
    state_id="poison",source_id="monster:fallen",duration=2,policy="refresh_same_source"}})
`)
	}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	instances, _ := akara.GetDynamicStore(engine.World(), "d2legacy.state.instance")
	if instances.Len() != 1 {
		t.Fatalf("instances after apply = %d", instances.Len())
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if instances.Len() != 0 {
		t.Fatalf("instances after expiration = %d", instances.Len())
	}
	events, _ := akara.GetDynamicStore(engine.World(), "d2legacy.state.event")
	if events.Len() != 2 {
		t.Fatalf("timed-state events = %d, want apply and expire", events.Len())
	}
}

func TestAuthorityMaterializesInitialItemsIntoLuaOwnedECS(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	bootstrap := map[string]any{"d2legacy.items": map[string]any{
		"owner": "alice", "inventory_width": float64(10), "inventory_height": float64(4),
		"stash_width": float64(6), "stash_height": float64(8), "cube_width": float64(3), "cube_height": float64(4),
		"belt_capacity": float64(4), "active_weapon_set": float64(0), "vendor_width": float64(10), "vendor_height": float64(10),
		"carried_gold": float64(100), "stashed_gold": float64(200),
		"items": []any{map[string]any{"id": "sword", "code": "ssd", "width": float64(1), "height": float64(3),
			"body_slots": "rarm,larm", "belt_eligible": false, "base_cost": float64(100),
			"inventory_dc6": "sword.dc6", "world_dc6": "flpsword.dc6", "world_animated": true,
			"container": "inventory", "x": float64(0), "y": float64(0), "slot": "", "belt_slot": float64(0),
			"weapon_set": float64(0), "page": float64(0), "melee_range": float64(2),
			"physical_min": float64(512), "physical_max": float64(1024), "melee_weapon_class": "1HS"}},
	}}
	authority, err := StartWithConfig(ctx, content.D2Legacy(), fixtureRecords{}, engine, session, Config{Seed: 4, InitialData: bootstrap})
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)
	layouts, _ := akara.GetDynamicStore(engine.World(), "d2legacy.items.layout")
	items, _ := akara.GetDynamicStore(engine.World(), "d2legacy.item.identity")
	if layouts.Len() != 1 || items.Len() != 1 {
		t.Fatalf("bootstrapped layouts/items = %d/%d", layouts.Len(), items.Len())
	}
}

func TestAuthorityMovesItemsThroughLuaOwnedPolicy(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	bootstrap := map[string]any{"d2legacy.items": map[string]any{
		"owner": "alice", "inventory_width": float64(10), "inventory_height": float64(4),
		"stash_width": float64(6), "stash_height": float64(8), "cube_width": float64(3), "cube_height": float64(4),
		"belt_capacity": float64(4), "vendor_width": float64(10), "vendor_height": float64(10),
		"items": []any{map[string]any{"id": "sword", "code": "ssd", "width": float64(1), "height": float64(3),
			"body_slots": "rarm,larm", "container": "inventory"}},
	}}
	authority, err := StartWithConfig(ctx, content.D2Legacy(), fixtureRecords{}, engine, session, Config{Seed: 4, InitialData: bootstrap})
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)
	payload, _ := json.Marshal(map[string]any{"item_id": "sword", "destination": map[string]any{"container": "held"}})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "alice", Authority: simulation.AuthorityPlayer, Sequence: 1, Kind: "item.move", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	placements, _ := akara.GetDynamicStore(engine.World(), "d2legacy.item.placement")
	value, _ := placements.Get(placements.Entities()[0])
	container, _ := value.Get("container")
	if container != "held" {
		t.Fatalf("container = %v, want held", container)
	}
}

type fixtureRecords struct{}

func (fixtureRecords) Invalidate(string)  {}
func (fixtureRecords) Loaded(string) bool { return true }
func (fixtureRecords) Load(path string) ([]map[string]string, error) {
	if path == "data/global/excel/skills.txt" {
		return []map[string]string{{"Id": "36", "skill": "Fire Bolt", "srvmissile": "firebolt", "etype": "fire", "interrupt": "1", "srvstfunc": "", "srvdofunc": "", "mana": "5", "manashift": "7", "emin": "3", "emax": "6", "HitShift": "8"}}, nil
	}
	return []map[string]string{{"Missile": "firebolt", "Skill": "Fire Bolt", "pSrvDoFunc": "1", "CollideType": "3", "CollideKill": "1", "Vel": "20", "Range": "40", "Size": "2", "CelFile": "firebolt", "AnimSpeed": "16", "NumDirections": "16", "LoopAnim": "1"}}, nil
}

func TestAuthorityRestoresAllDeterministicParticipantsBeforeFirstTick(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), fixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := authority.Random.Uint64n("d2legacy.combat.fire_bolt.damage", 100); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	session.Close()

	checkpoint := replay.Checkpoints[0]
	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), fixtureRecords{}, restoredEngine, restoredSession, Config{Seed: 7, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(checkpoint.Participants, restoredReplay.InitialParticipants) {
		t.Fatalf("restored participants differ\nwant: %#v\n got: %#v", checkpoint.Participants, restoredReplay.InitialParticipants)
	}
	assertParticipantIDs(t, restoredReplay.InitialParticipants,
		"engine.authoritative_rng/v1", "engine.authoritative_runtime/v1", "engine.authoritative_state/v1")
}

func assertParticipantIDs(t *testing.T, states []simulation.ParticipantState, expected ...string) {
	t.Helper()
	if len(states) != len(expected) {
		t.Fatalf("participant count = %d, want %d", len(states), len(expected))
	}
	for index, id := range expected {
		if states[index].ID != id {
			t.Fatalf("participant %d = %q, want %q", index, states[index].ID, id)
		}
	}
}

func TestAuthorityBootsWithoutClientOrRenderer(t *testing.T) {
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(context.Background(), content.D2Legacy(), fixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
