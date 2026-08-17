package d2legacy

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// TestGeneratedHostileLifecycleRestoresIdentically closes the last M21.14.7
// acceptance gap. The monster is selected and materialized by the production
// Blood Moor population policy, acquires the player through production AI, and
// dies through the production cast/projectile/damage/death systems. A newly
// constructed Lua runtime must continue that live hostile to the same corpse,
// credit, XP, loot-event surface, and authoritative checksum.
func TestGeneratedHostileLifecycleRestoresIdentically(t *testing.T) {
	ctx := context.Background()
	records := generatedHostileRecords()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := Start(ctx, content.D2Legacy(), records, engine, session, 314)
	if err != nil {
		t.Fatal(err)
	}

	playerPayload, _ := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 0,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": 0, "y": 0,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 2,
	})
	populationPayload, _ := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": 0,
		"links": []map[string]any{},
		"rooms": []map[string]any{{
			"id": "blood-moor-a", "populate": true,
			"x": 0, "y": 0, "width": 10, "height": 10,
			"points": []map[string]any{{"x": 4, "y": 0}},
		}},
	})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.population.bootstrap", Payload: populationPayload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	if identities.Len() != 0 {
		t.Fatalf("population bootstrap eagerly created %d monsters without an active player room", identities.Len())
	}
	if err := session.Submit(simulation.Command{Tick: 2, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: playerPayload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	assertMonsterPlayerCount(t, engine, "level:2:room:blood-moor-a:monster:1", 1)
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]

	castPayload, _ := json.Marshal(map[string]any{
		"side": "left", "target_x": 4, "target_y": 0,
		"target_id": "monster:level:2:room:blood-moor-a:monster:1",
	})
	cast := simulation.Command{Tick: 3, Player: "alice", Authority: simulation.AuthorityPlayer, Sequence: 1, Kind: "player.use_skill", Payload: castPayload}
	if err := session.Submit(cast); err != nil {
		t.Fatal(err)
	}
	stepSession(t, session, 8)
	originalReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	original := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1]
	assertCompletedHostileLifecycle(t, engine)
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if err := engine.Close(); err != nil {
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
	restored, err := StartWithConfig(ctx, content.D2Legacy(), records, restoredEngine, restoredSession, Config{Seed: 314, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)
	if err := restoredSession.Submit(cast); err != nil {
		t.Fatal(err)
	}
	stepSession(t, restoredSession, 8)
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	continued := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1]
	if continued.Checksum != original.Checksum {
		t.Fatalf("restored hostile lifecycle checksum = %s, want %s", continued.Checksum, original.Checksum)
	}
	assertCompletedHostileLifecycle(t, restoredEngine)
}

func TestPopulationActivatesAdjacentRoomsAndPinsCurrentPlayerCount(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	defer engine.Close()
	authority, err := Start(ctx, content.D2Legacy(), generatedHostileRecords(), engine, session, 314)
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(ctx)

	population, _ := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": 0,
		"links": []map[string]any{{"from": "a", "to": "b"}},
		"rooms": []map[string]any{
			{"id": "a", "populate": true, "x": 0, "y": 0, "width": 10, "height": 10,
				"points": []map[string]any{{"x": 4, "y": 0}}},
			{"id": "b", "populate": true, "x": 10, "y": 0, "width": 10, "height": 10,
				"points": []map[string]any{{"x": 14, "y": 0}}},
			{"id": "c", "populate": true, "x": 20, "y": 0, "width": 10, "height": 10,
				"points": []map[string]any{{"x": 24, "y": 0}}},
		},
	})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.population.bootstrap", Payload: population}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if got := monsterCount(engine); got != 0 {
		t.Fatalf("monster count before room activation = %d, want 0", got)
	}

	first := generatedPlayerPayload(t, "hero", "alice", 1, 1)
	if err := session.Submit(simulation.Command{Tick: 2, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: first}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if got := monsterCount(engine); got != 2 {
		t.Fatalf("monster count after activating room and neighbor = %d, want 2", got)
	}
	assertMonsterPlayerCount(t, engine, "level:2:room:a:monster:1", 1)
	assertMonsterPlayerCount(t, engine, "level:2:room:b:monster:1", 1)

	second := generatedPlayerPayload(t, "hero-2", "bob", 21, 1)
	if err := session.Submit(simulation.Command{Tick: 3, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 2, Kind: "system.player.enter", Payload: second}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if got := monsterCount(engine); got != 3 {
		t.Fatalf("monster count after activating remote room = %d, want 3", got)
	}
	assertMonsterPlayerCount(t, engine, "level:2:room:a:monster:1", 1)
	assertMonsterPlayerCount(t, engine, "level:2:room:c:monster:1", 2)
}

func TestPopulationArchivesInactiveMonsterAndRestoresCheckpointParity(t *testing.T) {
	ctx := context.Background()
	authority, engine, session := startPartyFixture(t, nil)
	t.Cleanup(func() {
		_ = authority.Stop(ctx)
		_ = session.Close()
		_ = engine.Close()
	})

	population, _ := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": 0,
		"links": []map[string]any{{"from": "a", "to": "b"}, {"from": "b", "to": "c"}},
		"rooms": []map[string]any{
			{"id": "a", "populate": true, "x": 0, "y": 0, "width": 10, "height": 10,
				"points": []map[string]any{{"x": 8, "y": 1}}},
			{"id": "b", "populate": false, "x": 10, "y": 0, "width": 10, "height": 10,
				"points": []map[string]any{}},
			{"id": "c", "populate": false, "x": 20, "y": 0, "width": 10, "height": 10,
				"points": []map[string]any{}},
		},
	})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.population.bootstrap", Payload: population}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(simulation.Command{Tick: 2, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: generatedPlayerPayload(t, "hero", "alice", 1, 1)}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	const spawnID = "level:2:room:a:monster:1"
	assertMonsterPlayerCount(t, engine, spawnID, 1)
	submitMoveCommand(t, session, 3, "alice", 1, 1)
	for playerPositionX(t, engine, "alice") < 20 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
		if engine.Tick() > 100 {
			t.Fatal("player did not reach the remote room")
		}
	}
	beforeArchive := monsterArchiveSnapshot(t, engine, spawnID)
	beforeAI := beforeArchive["d2legacy.monster.ai"]
	delete(beforeArchive, "d2legacy.monster.ai")
	archiveTick := engine.Tick() + 1
	submitMoveCommand(t, session, archiveTick, "alice", 2, 0)
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	if got := monsterCount(engine); got != 0 {
		t.Fatalf("monster count after room deactivation = %d, want 0", got)
	}
	assertArchivedMonster(t, authority, "a", spawnID, beforeArchive)
	assertArchivedMonsterAI(t, authority, "a", beforeAI)

	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	restored, restoredEngine, restoredSession := startPartyFixture(t, &checkpoint)
	t.Cleanup(func() {
		_ = restored.Stop(ctx)
		_ = restoredSession.Close()
		_ = restoredEngine.Close()
	})
	if got := monsterCount(restoredEngine); got != 0 {
		t.Fatalf("restored inactive monster count = %d, want 0", got)
	}
	assertArchivedMonster(t, restored, "a", spawnID, beforeArchive)
	assertArchivedMonsterAI(t, restored, "a", beforeAI)

	returnTick := checkpoint.Tick + 1
	submitMoveCommand(t, session, returnTick, "alice", 3, -1)
	submitMoveCommand(t, restoredSession, returnTick, "alice", 1, -1)
	stepSession(t, session, 55)
	stepSession(t, restoredSession, 55)
	stopTick := returnTick + 55
	submitMoveCommand(t, session, stopTick, "alice", 4, 0)
	submitMoveCommand(t, restoredSession, stopTick, "alice", 2, 0)
	stepSession(t, session, 2)
	stepSession(t, restoredSession, 2)

	originalReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	originalCheckpoint := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1]
	restoredCheckpoint := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1]
	if restoredCheckpoint.Checksum != originalCheckpoint.Checksum {
		t.Fatalf("inactive monster restore checksum = %s, want %s", restoredCheckpoint.Checksum, originalCheckpoint.Checksum)
	}
	originalMonster := monsterArchiveSnapshot(t, engine, spawnID)
	restoredMonster := monsterArchiveSnapshot(t, restoredEngine, spawnID)
	if !reflect.DeepEqual(restoredMonster, originalMonster) {
		t.Fatalf("restored monster components = %#v, want %#v", restoredMonster, originalMonster)
	}
	assertRoomActiveWithoutArchive(t, authority, "a")
	assertRoomActiveWithoutArchive(t, restored, "a")
}

var archivedMonsterComponents = []string{
	"d2legacy.population.room_resident",
	"d2legacy.monster.identity",
	"d2legacy.monster.stats",
	"d2legacy.combat.melee_profile",
	"d2legacy.monster.appearance",
	"d2legacy.monster.ai",
	"d2legacy.world.position",
	"d2legacy.world.velocity",
	"d2legacy.world.facing",
	"d2legacy.world.location",
	"d2legacy.world.collider",
	"engine.world.velocity_mover",
	"d2legacy.world.selectable",
}

type populationPlanFixture struct {
	Rooms []populationRoomFixture `json:"rooms"`
}

type populationRoomFixture struct {
	ID        string          `json:"id"`
	Active    bool            `json:"active"`
	Activated bool            `json:"activated"`
	Archived  json.RawMessage `json:"archived"`
}

type archivedMonsterFixture struct {
	SpawnID    string                     `json:"spawn_id"`
	Components map[string]json.RawMessage `json:"components"`
}

func submitMoveCommand(t *testing.T, session *gamesession.Session, tick uint64, player string, sequence uint64, x int) {
	t.Helper()
	payload, err := json.Marshal(map[string]any{"x": x, "y": 0, "running": false})
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(simulation.Command{Tick: tick, Player: player, Authority: simulation.AuthorityPlayer,
		Sequence: sequence, Kind: "player.move", Payload: payload}); err != nil {
		t.Fatal(err)
	}
}

func playerPositionX(t *testing.T, engine *gameecs.Engine, player string) float64 {
	t.Helper()
	controls, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.player_control")
	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		current, _ := control.Get("player")
		if current != player {
			continue
		}
		position, present := positions.Get(entity)
		if !present {
			t.Fatalf("player %s has no position", player)
		}
		x, _ := position.Get("x")
		return x.(float64)
	}
	t.Fatalf("player %s was not found", player)
	return 0
}

func monsterArchiveSnapshot(t *testing.T, engine *gameecs.Engine, spawnID string) map[string]map[string]any {
	t.Helper()
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	var monster akara.Entity
	found := false
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		current, _ := identity.Get("spawn_id")
		if current == spawnID {
			monster, found = entity, true
			break
		}
	}
	if !found {
		t.Fatalf("monster %s was not found", spawnID)
	}
	result := make(map[string]map[string]any, len(archivedMonsterComponents))
	for _, name := range archivedMonsterComponents {
		store, _ := akara.GetDynamicStore(engine.World(), name)
		component, present := store.Get(monster)
		if !present {
			t.Fatalf("monster %s has no %s", spawnID, name)
		}
		snapshot, err := component.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		result[name] = snapshot
	}
	return result
}

func readPopulationPlan(t *testing.T, authority *Authority) populationPlanFixture {
	t.Helper()
	registered, found := authority.State.Read("d2legacy.population.plan")
	if !found {
		t.Fatal("population plan authority state is missing")
	}
	var plan populationPlanFixture
	if err := json.Unmarshal(registered.Data, &plan); err != nil {
		t.Fatalf("decode population plan %s: %v", registered.Data, err)
	}
	return plan
}

func assertArchivedMonster(t *testing.T, authority *Authority, roomID, spawnID string, before map[string]map[string]any) {
	t.Helper()
	for _, room := range readPopulationPlan(t, authority).Rooms {
		if room.ID != roomID {
			continue
		}
		archived := decodeArchivedMonsters(t, room.Archived)
		if room.Active || !room.Activated || len(archived) != 1 || archived[0].SpawnID != spawnID {
			t.Fatalf("inactive room archive = %#v", room)
		}
		for name, want := range before {
			encoded, present := archived[0].Components[name]
			if !present {
				t.Fatalf("archived monster has no %s", name)
			}
			var got map[string]any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			wantJSON, _ := json.Marshal(want)
			var normalizedWant map[string]any
			if err := json.Unmarshal(wantJSON, &normalizedWant); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, normalizedWant) {
				t.Fatalf("archived %s = %#v, want %#v", name, got, normalizedWant)
			}
		}
		return
	}
	t.Fatalf("population room %s was not found", roomID)
}

func assertRoomActiveWithoutArchive(t *testing.T, authority *Authority, roomID string) {
	t.Helper()
	for _, room := range readPopulationPlan(t, authority).Rooms {
		if room.ID == roomID {
			if !room.Active || !room.Activated || len(decodeArchivedMonsters(t, room.Archived)) != 0 {
				t.Fatalf("restored room state = %#v", room)
			}
			return
		}
	}
	t.Fatalf("population room %s was not found", roomID)
}

func assertArchivedMonsterAI(t *testing.T, authority *Authority, roomID string, before map[string]any) {
	t.Helper()
	for _, room := range readPopulationPlan(t, authority).Rooms {
		if room.ID != roomID {
			continue
		}
		archived := decodeArchivedMonsters(t, room.Archived)
		if len(archived) != 1 {
			t.Fatalf("room %s archived monsters = %d, want 1", roomID, len(archived))
		}
		encoded, present := archived[0].Components["d2legacy.monster.ai"]
		if !present {
			t.Fatal("archived monster has no AI state")
		}
		var got map[string]any
		if err := json.Unmarshal(encoded, &got); err != nil {
			t.Fatal(err)
		}
		for _, field := range []string{"behavior", "state", "target_id", "think_interval", "aggro_radius", "attack_range", "speed"} {
			wantJSON, _ := json.Marshal(before[field])
			var want any
			if err := json.Unmarshal(wantJSON, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got[field], want) {
				t.Fatalf("archived monster AI %s = %#v, want %#v", field, got[field], want)
			}
		}
		if got["next_think_tick"].(float64) < float64(before["next_think_tick"].(int64)) {
			t.Fatalf("archived next think tick = %v, before %v", got["next_think_tick"], before["next_think_tick"])
		}
		return
	}
	t.Fatalf("population room %s was not found", roomID)
}

func decodeArchivedMonsters(t *testing.T, encoded json.RawMessage) []archivedMonsterFixture {
	t.Helper()
	if len(encoded) == 0 || string(encoded) == "{}" {
		return nil
	}
	var archived []archivedMonsterFixture
	if err := json.Unmarshal(encoded, &archived); err != nil {
		t.Fatal(err)
	}
	return archived
}

func generatedPlayerPayload(t *testing.T, characterID, player string, x, y float64) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"character_id": characterID, "player": player, "name": characterID, "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 0,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": x, "y": y,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func monsterCount(engine *gameecs.Engine) int {
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	return identities.Len()
}

func assertMonsterPlayerCount(t *testing.T, engine *gameecs.Engine, spawnID string, want int64) {
	t.Helper()
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	stats, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		current, _ := identity.Get("spawn_id")
		if current != spawnID {
			continue
		}
		values, present := stats.Get(entity)
		if !present {
			t.Fatalf("monster %s has no stats", spawnID)
		}
		count, _ := values.Get("spawn_player_count")
		if count != want {
			t.Fatalf("monster %s player count = %v, want %d", spawnID, count, want)
		}
		return
	}
	t.Fatalf("monster %s was not created", spawnID)
}

func stepSession(t *testing.T, session *gamesession.Session, count int) {
	t.Helper()
	for range count {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
}

func assertCompletedHostileLifecycle(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	deaths, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death")
	events, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death_event")
	selectables, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable")
	brains, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.ai")
	progress, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.progress")
	if identities.Len() != 1 || deaths.Len() != 1 {
		t.Fatalf("monster identities/deaths = %d/%d, want 1/1", identities.Len(), deaths.Len())
	}
	monster := identities.Entities()[0]
	death, present := deaths.Get(monster)
	if !present {
		t.Fatal("generated monster has no durable death state")
	}
	credited, _ := death.Get("credited_id")
	active, _ := death.Get("active")
	corpse, _ := death.Get("corpse_usable")
	drops, _ := death.Get("drops")
	if credited != "player:alice" || active != false || corpse != true {
		t.Fatalf("death credit/active/corpse = %v/%v/%v", credited, active, corpse)
	}
	if !strings.Contains(drops.(string), `"code":"rin"`) ||
		!strings.Contains(drops.(string), `"quality":"unique"`) {
		t.Fatalf("checkpointed monster drops = %s, want unique ring", drops)
	}
	if events.Len() != 4 {
		t.Fatalf("death events = %d, want kill, loot, quest, and presentation", events.Len())
	}
	if _, present := brains.Get(monster); present {
		t.Fatal("dead monster retained active AI")
	}
	for _, entity := range selectables.Entities() {
		value, _ := selectables.Get(entity)
		id, _ := value.Get("id")
		if id == "monster:level:2:room:blood-moor-a:monster:1" {
			t.Fatal("dead monster remained targetable")
		}
	}
	if progress.Len() != 1 {
		t.Fatalf("player progress records = %d, want 1", progress.Len())
	}
	value, _ := progress.Get(progress.Entities()[0])
	experience, _ := value.Get("experience")
	if experience != int64(5) {
		t.Fatalf("credited experience = %v, want 5", experience)
	}
}

func generatedHostileRecords() fixtureRecords {
	records := fixtureRecords{}
	records["data/global/excel/levels.txt"] = []map[string]string{{
		"Id": "2", "MonDen": "100000", "NumMon": "1", "mon1": "fallen",
	}}
	records["data/global/excel/monstats.txt"] = []map[string]string{{
		"Id": "fallen", "BaseId": "fallen", "NameStr": "Fallen", "AI": "fallen", "Code": "FA",
		"enabled": "1", "isSpawn": "1", "npc": "0", "noRatio": "1", "Level": "1",
		"minHP": "3", "maxHP": "3", "AC": "0", "A1TH": "0", "A1MinD": "1", "A1MaxD": "1",
		"Exp": "5", "Velocity": "0", "aidel": "1", "aidist": "20", "MinGrp": "1", "MaxGrp": "1",
		"Rarity": "1", "TreasureClass1": "fallen-drop",
	}}
	records["data/global/excel/monstats2.txt"] = []map[string]string{{
		"Id": "fallen", "BaseW": "HTH", "SizeX": "1", "SizeY": "1", "MeleeRng": "1",
	}}
	records["data/global/excel/monlvl.txt"] = []map[string]string{{"Level": "1"}}
	records["data/global/excel/treasureclassex.txt"] = []map[string]string{{
		"Treasure Class": "fallen-drop", "Picks": "-1", "Item1": "rin", "Prob1": "1", "Unique": "1024",
	}}
	records["data/global/excel/misc.txt"] = []map[string]string{{
		"code": "rin", "namestr": "Ring", "type": "ring", "level": "1", "invwidth": "1", "invheight": "1",
	}}
	records["data/global/excel/itemratio.txt"] = []map[string]string{{
		"Version": "100", "Uber": "0", "Class Specific": "0",
		"Unique": "4000", "Set": "4000", "Rare": "4000", "Magic": "4000",
		"HiQuality": "4000", "Normal": "4000",
	}}
	return records
}
