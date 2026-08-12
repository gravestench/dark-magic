package d2legacy

import (
	"context"
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

func TestConfigureRuntimePreservesClientCatalogOverrides(t *testing.T) {
	runtime := modruntime.New()
	for _, name := range []string{"d2legacy.quest_catalog/v1", "d2legacy.map_catalog/v1"} {
		if err := runtime.RegisterModule(modruntime.Module{
			Name: name,
			Help: modruntime.ModuleHelp{Summary: "client override"},
			Loader: func(state *lua.LState) int {
				state.Push(state.NewTable())
				return 1
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := ConfigureRuntime(runtime, content.D2Legacy(), runtimeFixtureRecords{}, engine, session,
		simulation.NewStateStore(), simulation.NewRandomStreams(1), nil); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"d2legacy.quest_catalog/v1", "d2legacy.map_catalog/v1"} {
		if got := runtime.ModuleHelp()[name].Summary; got != "client override" {
			t.Fatalf("%s summary = %q, want client override", name, got)
		}
		count := 0
		for _, registered := range runtime.ModuleNames() {
			if registered == name {
				count++
			}
		}
		if count != 1 {
			t.Fatalf("%s registered %d times", name, count)
		}
	}
}

func TestAuthorityMaterializesPlayerEntryThroughLua(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)
	payload, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 20,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "",
		"palette": "data/global/Palette/units/pal.dat", "direction": 0, "mode": "NU",
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
	if left != int64(36) {
		t.Fatalf("initial left skill = %v, want Lua-selected Fire Bolt 36", left)
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
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 99)
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

func TestMonsterMeleeReachIncludesBothActorFootprints(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 17)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)

	player, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 20,
		"health": 500, "max_health": 500, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": 0, "y": 0,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 1,
	})
	monster, _ := json.Marshal(map[string]any{
		"spawn_id": "spacing-fallen", "seed": 9, "x": 2.5, "y": 0, "act": 1, "level_id": 1,
		"definition": map[string]any{
			"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen",
			"ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
			"life_min": 256, "life_max": 256, "level": 1, "defense": 0, "attack_rating": 0,
			"physical_min": 256, "physical_max": 256, "experience": 5, "treasure_class": "",
			"collider_radius": 1, "select_radius": 1, "velocity": 5, "think_interval": 1,
			"aggro_radius": 20, "attack_range": 1,
		},
	})
	for _, command := range []simulation.Command{
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.player.enter", Payload: player},
		{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.monster.spawn", Payload: monster},
	} {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	for range 3 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}

	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	selectables, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable")
	var playerX, playerY, monsterX, monsterY float64
	for _, entity := range selectables.Entities() {
		selectable, _ := selectables.Get(entity)
		id, _ := selectable.Get("id")
		position, _ := positions.Get(entity)
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		if id == "player:alice" {
			playerX, playerY = x.(float64), y.(float64)
		} else if id == "monster:spacing-fallen" {
			monsterX, monsterY = x.(float64), y.(float64)
		}
	}
	distance := math.Hypot(monsterX-playerX, monsterY-playerY)
	if distance < 2 {
		t.Fatalf("actor centers separated by %v, want at least combined collider radii 2", distance)
	}
	brains, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.ai")
	brain, _ := brains.Get(brains.Entities()[0])
	state, _ := brain.Get("state")
	if state != "attack" {
		t.Fatalf("monster state = %v at distance %v, want attack within footprint-aware reach 3", state, distance)
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
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 3)
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
	authority, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, Config{Seed: 4, InitialData: bootstrap})
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
	authority, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, Config{Seed: 4, InitialData: bootstrap})
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

type runtimeFixtureRecords struct{}

func (runtimeFixtureRecords) Invalidate(string)  {}
func (runtimeFixtureRecords) Loaded(string) bool { return true }
func (runtimeFixtureRecords) Load(path string) ([]map[string]string, error) {
	switch path {
	case "data/global/excel/charstats.txt":
		return []map[string]string{{"class": "Amazon", "StartSkill": "Fire Bolt"}}, nil
	case "data/global/excel/skilldesc.txt":
		return []map[string]string{{"skilldesc": "firebolt", "ListRow": "0", "IconCel": "0"}}, nil
	case "data/global/excel/skills.txt":
		return []map[string]string{{"Id": "36", "skill": "Fire Bolt", "skilldesc": "firebolt", "leftskill": "1", "general": "0", "passive": "0", "srvmissile": "firebolt", "etype": "fire", "interrupt": "1", "srvstfunc": "", "srvdofunc": "", "mana": "5", "manashift": "7", "emin": "3", "emax": "6", "HitShift": "8"}}, nil
	case "data/global/excel/Missiles.txt":
		return []map[string]string{{"Missile": "firebolt", "Skill": "Fire Bolt", "pSrvDoFunc": "1", "CollideType": "3", "CollideKill": "1", "Vel": "20", "Range": "40", "Size": "2", "CelFile": "firebolt", "AnimSpeed": "16", "NumDirections": "16", "LoopAnim": "1"}}, nil
	}
	return nil, nil
}

func TestAuthorityRestoresAllDeterministicParticipantsBeforeFirstTick(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 7)
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
	restored, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, restoredEngine, restoredSession, Config{Seed: 7, Restore: checkpoint.Participants})
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

func TestAuthorityCheckpointRestoreContinuesWithIdenticalOutcome(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 77)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]

	spawnPayload, _ := json.Marshal(map[string]any{"spawn_id": "continued-fallen", "seed": float64(9), "x": float64(8), "y": float64(9), "act": float64(1), "level_id": float64(2), "definition": map[string]any{
		"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen", "ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
		"life_min": float64(256), "life_max": float64(768), "level": float64(1), "defense": float64(0), "attack_rating": float64(0), "physical_min": float64(256), "physical_max": float64(256), "experience": float64(5), "treasure_class": "Act 1 H2H A", "collider_radius": float64(1), "select_radius": float64(1), "velocity": float64(5), "think_interval": float64(1), "aggro_radius": float64(20), "attack_range": float64(1)}})
	command := simulation.Command{Tick: 2, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.monster.spawn", Payload: spawnPayload}
	if err := session.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	originalReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	originalChecksum := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1].Checksum
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	session.Close()

	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, restoredEngine, restoredSession, Config{Seed: 77, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	if err := restoredSession.Submit(command); err != nil {
		t.Fatal(err)
	}
	if err := restoredSession.Step(); err != nil {
		t.Fatal(err)
	}
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	restoredChecksum := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1].Checksum
	if restoredChecksum != originalChecksum {
		t.Fatalf("continued checksum = %s, want %s", restoredChecksum, originalChecksum)
	}
}

// TestFireBoltCheckpointRestoreParity covers the complete migrated vertical
// path rather than a synthetic state mutation: Lua admits the player command,
// pays mana, advances cast timing, creates and moves a missile, resolves swept
// contact, applies damage, and emits the combat result. A newly constructed Lua
// runtime must continue the in-flight cast to the identical session checksum.
func TestFireBoltCheckpointRestoreParity(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 123)
	if err != nil {
		t.Fatal(err)
	}

	playerPayload, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 0,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": 0, "y": 0,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 1,
	})
	monsterPayload, _ := json.Marshal(map[string]any{
		"spawn_id": "fallen-firebolt", "seed": 9, "x": 4, "y": 0, "act": 1, "level_id": 1,
		"definition": map[string]any{
			"id": "fallen", "base_id": "fallen", "graphics_id": "fallen", "name_key": "Fallen",
			"ai": "fallen", "token": "FA", "weapon_class": "HTH", "components": map[string]string{},
			"life_min": 4096, "life_max": 4096, "level": 1, "defense": 0, "attack_rating": 0,
			"physical_min": 256, "physical_max": 256, "experience": 5, "treasure_class": "",
			"collider_radius": 0.5, "select_radius": 0.5, "velocity": 0,
			"think_interval": 100, "aggro_radius": 0, "attack_range": 1,
		},
	})
	for _, command := range []simulation.Command{
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.player.enter", Payload: playerPayload},
		{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem, Sequence: 1, Kind: "system.monster.spawn", Payload: monsterPayload},
	} {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	// No unit target is supplied: this is the ground-targeted missile vector.
	castPayload, _ := json.Marshal(map[string]any{"side": "left", "target_x": 8, "target_y": 0})
	if err := session.Submit(simulation.Command{Tick: 2, Player: "alice", Authority: simulation.AuthorityPlayer, Sequence: 1, Kind: "player.use_skill", Payload: castPayload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]

	for range 6 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	originalReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	original := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1]
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}

	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer restoredEngine.Close()
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), runtimeFixtureRecords{}, restoredEngine, restoredSession, Config{Seed: 123, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	for range 6 {
		if err := restoredSession.Step(); err != nil {
			t.Fatal(err)
		}
	}
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	continued := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1]
	if continued.Checksum != original.Checksum {
		t.Fatalf("Fire Bolt continuation checksum = %s, want %s", continued.Checksum, original.Checksum)
	}
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
	authority, err := Start(context.Background(), content.D2Legacy(), runtimeFixtureRecords{}, engine, session, 7)
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
}
