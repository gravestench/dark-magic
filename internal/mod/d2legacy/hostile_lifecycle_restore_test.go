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
		"level": 1, "experience": 0, "dexterity": 20, "vitality": 20, "defense": 0,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"stamina": 84, "max_stamina": 84,
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

	cast := simulation.Command{
		Tick: 3, Player: "alice", Authority: simulation.AuthorityPlayer,
		Sequence: 1, Kind: "player.use_skill", Payload: castPayload,
	}
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
	defer func() { _ = restoredEngine.Close() }()

	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = restoredSession.Close() }()

	restored, err := StartWithConfig(
		ctx,
		content.D2Legacy(),
		records,
		restoredEngine,
		restoredSession,
		Config{Seed: 314, Restore: checkpoint.Participants},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = restored.Stop(ctx) }()

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

// TestPopulationActivatesAdjacentRoomsAndPinsCurrentPlayerCount verifies room
// activation and spawn scaling use the same authoritative player snapshot.
func TestPopulationActivatesAdjacentRoomsAndPinsCurrentPlayerCount(t *testing.T) {
	ctx := context.Background()
	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()
	defer func() { _ = engine.Close() }()

	authority, err := Start(ctx, content.D2Legacy(), generatedHostileRecords(), engine, session, 314)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = authority.Stop(ctx) }()

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

// TestPopulationTracksMovingResidentCurrentRoom ensures residency follows the
// entity's current room instead of retaining its original spawn location.
func TestPopulationTracksMovingResidentCurrentRoom(t *testing.T) {
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
			{"id": "a", "populate": false, "x": 0, "y": 0, "width": 10, "height": 10, "points": []map[string]any{}},
			{"id": "b", "populate": false, "x": 10, "y": 0, "width": 10, "height": 10, "points": []map[string]any{}},
			{"id": "c", "populate": false, "x": 20, "y": 0, "width": 10, "height": 10, "points": []map[string]any{}},
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

	residentStore, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.room_resident")
	positionStore, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	locationStore, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.location")

	objectID := engine.World().MustCreateEntity()
	if _, err := residentStore.Set(objectID, map[string]any{
		"id": "object:moving", "level_id": int64(2), "room_id": "a",
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := positionStore.Set(objectID, map[string]any{"x": float64(1), "y": float64(1)}); err != nil {
		t.Fatal(err)
	}

	if _, err := locationStore.Set(objectID, map[string]any{"act": int64(1), "level_id": int64(2)}); err != nil {
		t.Fatal(err)
	}

	setEntityPosition(t, positionStore, objectID, 11, 1)

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	assertResidentRoom(t, residentStore, objectID, 2, "b")

	playerID := playerEntity(t, engine, "alice")
	setEntityPosition(t, positionStore, playerID, 21, 1)

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	assertResidentRoom(t, residentStore, objectID, 2, "b")
	assertResidentActivation(t, engine, objectID, true, false)
	assertInactiveRoom(t, authority, "a")
	assertRoomActiveWithoutInactiveResidents(t, authority, "b")
}

// TestPopulationInactivatesMonsterAndRestoresCheckpointParity verifies a live
// monster's complete state graph survives room eviction and checkpoint restore.
func TestPopulationInactivatesMonsterAndRestoresCheckpointParity(t *testing.T) {
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
	monsterID := monsterEntity(t, engine, spawnID)
	ownerID := playerEntity(t, engine, "alice")
	submitOwnedUnitAttach(t, session, engine.Tick()+1, spawnID, "alice")

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	ownedRelation := assertOwnedUnitRelation(t, engine, monsterID, ownerID)
	graph := installMonsterTimedState(t, session, engine, monsterID)
	assertMonsterStateGraph(t, engine, graph)
	residentStore, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.room_resident")

	objectID := engine.World().MustCreateEntity()
	if _, err := residentStore.Set(objectID, map[string]any{
		"id": "object:room-a", "level_id": int64(2), "room_id": "a",
	}); err != nil {
		t.Fatal(err)
	}

	townObjectID := engine.World().MustCreateEntity()
	if _, err := residentStore.Set(townObjectID, map[string]any{
		"id": "object:town-a", "level_id": int64(1), "room_id": "a",
	}); err != nil {
		t.Fatal(err)
	}

	assertResidentActivation(t, engine, objectID, true, false)
	assertResidentActivation(t, engine, townObjectID, true, false)
	submitMoveCommand(t, session, engine.Tick()+1, "alice", 1, 1)

	for playerPositionX(t, engine, "alice") < 20 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}

		if engine.Tick() > 100 {
			t.Fatal("player did not reach the remote room")
		}
	}

	archiveTick := engine.Tick() + 1
	submitMoveCommand(t, session, archiveTick, "alice", 2, 0)

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	if got := monsterCount(engine); got != 1 {
		t.Fatalf("monster existence count after room deactivation = %d, want 1", got)
	}

	assertResidentActivation(t, engine, monsterID, false, true)
	assertResidentActivation(t, engine, objectID, false, false)
	assertResidentActivation(t, engine, townObjectID, true, false)
	assertMonsterStateGraph(t, engine, graph)

	if got := assertOwnedUnitRelation(t, engine, monsterID, ownerID); !reflect.DeepEqual(got, ownedRelation) {
		t.Fatalf("inactive owned-unit relation = %#v, want %#v", got, ownedRelation)
	}

	assertInactiveRoom(t, authority, "a", spawnID, "object:room-a")
	inactiveAI := componentSnapshot(t, engine, monsterID, "d2legacy.monster.ai")
	stepSession(t, session, 5)

	if got := componentSnapshot(t, engine, monsterID, "d2legacy.monster.ai"); !reflect.DeepEqual(got, inactiveAI) {
		t.Fatalf("inactive monster AI advanced: got %#v, want %#v", got, inactiveAI)
	}

	assertMonsterStateGraph(t, engine, graph)

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

	if got := monsterCount(restoredEngine); got != 1 {
		t.Fatalf("restored inactive monster existence count = %d, want 1", got)
	}

	assertResidentActivation(t, restoredEngine, monsterID, false, true)
	assertResidentActivation(t, restoredEngine, objectID, false, false)
	assertResidentActivation(t, restoredEngine, townObjectID, true, false)
	assertMonsterStateGraph(t, restoredEngine, graph)

	if restoredOwner := playerEntity(t, restoredEngine, "alice"); restoredOwner != ownerID {
		t.Fatalf("restored owned-unit owner entity = %d, want %d", restoredOwner, ownerID)
	}

	if got := assertOwnedUnitRelation(t, restoredEngine, monsterID, ownerID); !reflect.DeepEqual(got, ownedRelation) {
		t.Fatalf("restored owned-unit relation = %#v, want %#v", got, ownedRelation)
	}

	assertInactiveRoom(t, restored, "a", spawnID, "object:room-a")

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

	originalMonster := monsterComponentSnapshot(t, engine, spawnID)

	restoredMonster := monsterComponentSnapshot(t, restoredEngine, spawnID)
	if !reflect.DeepEqual(restoredMonster, originalMonster) {
		t.Fatalf("restored monster components = %#v, want %#v", restoredMonster, originalMonster)
	}

	assertResidentActivation(t, engine, monsterID, true, true)
	assertResidentActivation(t, restoredEngine, monsterID, true, true)
	assertResidentActivation(t, engine, objectID, true, false)
	assertResidentActivation(t, restoredEngine, objectID, true, false)
	assertResidentActivation(t, engine, townObjectID, true, false)
	assertResidentActivation(t, restoredEngine, townObjectID, true, false)
	assertMonsterStateGraph(t, engine, graph)
	assertMonsterStateGraph(t, restoredEngine, graph)
	assertRoomActiveWithoutInactiveResidents(t, authority, "a")
	assertRoomActiveWithoutInactiveResidents(t, restored, "a")

	if got := assertOwnedUnitRelation(t, engine, monsterID, ownerID); !reflect.DeepEqual(got, ownedRelation) {
		t.Fatalf("reactivated owned-unit relation = %#v, want %#v", got, ownedRelation)
	}

	if got := assertOwnedUnitRelation(t, restoredEngine, monsterID, ownerID); !reflect.DeepEqual(got, ownedRelation) {
		t.Fatalf("restored reactivated owned-unit relation = %#v, want %#v", got, ownedRelation)
	}
}

// TestPopulationInactivatesCorpseAndRestoresCheckpointParity verifies corpse
// ownership and timed state survive room eviction and checkpoint restoration.
func TestPopulationInactivatesCorpseAndRestoresCheckpointParity(t *testing.T) {
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

	monsterID := monsterEntity(t, engine, spawnID)
	setMonsterHealth(t, engine, monsterID, 0)

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	assertCorpseActivation(t, engine, monsterID, true)
	corpse := corpseComponentSnapshot(t, engine, monsterID)

	submitMoveCommand(t, session, engine.Tick()+1, "alice", 1, 1)

	for playerPositionX(t, engine, "alice") < 20 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}

		if engine.Tick() > 100 {
			t.Fatal("player did not leave the corpse room")
		}
	}

	stopTick := engine.Tick() + 1
	submitMoveCommand(t, session, stopTick, "alice", 2, 0)

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	assertCorpseActivation(t, engine, monsterID, false)

	if got := corpseComponentSnapshot(t, engine, monsterID); !reflect.DeepEqual(got, corpse) {
		t.Fatalf("inactive corpse components = %#v, want %#v", got, corpse)
	}

	assertInactiveRoom(t, authority, "a", spawnID)
	assertInactiveResidentMover(t, authority, "a", spawnID, false)

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
	assertCorpseActivation(t, restoredEngine, monsterID, false)

	if got := corpseComponentSnapshot(t, restoredEngine, monsterID); !reflect.DeepEqual(got, corpse) {
		t.Fatalf("restored inactive corpse components = %#v, want %#v", got, corpse)
	}

	assertInactiveRoom(t, restored, "a", spawnID)
	assertInactiveResidentMover(t, restored, "a", spawnID, false)

	returnTick := checkpoint.Tick + 1
	submitMoveCommand(t, session, returnTick, "alice", 3, -1)
	submitMoveCommand(t, restoredSession, returnTick, "alice", 1, -1)
	stepSession(t, session, 55)
	stepSession(t, restoredSession, 55)

	stopTick = returnTick + 55
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
		t.Fatalf("inactive corpse restore checksum = %s, want %s", restoredCheckpoint.Checksum, originalCheckpoint.Checksum)
	}

	assertCorpseActivation(t, engine, monsterID, true)
	assertCorpseActivation(t, restoredEngine, monsterID, true)

	if got := corpseComponentSnapshot(t, engine, monsterID); !reflect.DeepEqual(got, corpse) {
		t.Fatalf("reactivated corpse components = %#v, want %#v", got, corpse)
	}

	if got := corpseComponentSnapshot(t, restoredEngine, monsterID); !reflect.DeepEqual(got, corpse) {
		t.Fatalf("restored reactivated corpse components = %#v, want %#v", got, corpse)
	}
}

var retainedMonsterComponents = []string{
	"d2legacy.world.room_resident",
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
	"d2legacy.owned_unit",
}

var retainedCorpseComponents = []string{
	"d2legacy.world.room_resident",
	"d2legacy.monster.identity",
	"d2legacy.monster.stats",
	"d2legacy.combat.melee_profile",
	"d2legacy.monster.appearance",
	"d2legacy.monster.death",
	"d2legacy.monster.revivable",
	"d2legacy.world.position",
	"d2legacy.world.velocity",
	"d2legacy.world.facing",
	"d2legacy.world.location",
	"d2legacy.world.occupancy",
	"d2legacy.world.selectable",
}

type populationPlanFixture struct {
	Rooms []populationRoomFixture `json:"rooms"`
}

type populationRoomFixture struct {
	ID                string          `json:"id"`
	Active            bool            `json:"active"`
	Activated         bool            `json:"activated"`
	InactiveResidents json.RawMessage `json:"inactive_residents"`
}

type inactiveResidentFixture struct {
	ID            string `json:"id"`
	VelocityMover bool   `json:"velocity_mover"`
}

// submitMoveCommand keeps movement ordering explicit so activation transitions
// can be attributed to a precise authority tick.
func submitMoveCommand(
	t *testing.T,
	session *gamesession.Session,
	tick uint64,
	player string,
	sequence uint64,
	x int,
) {
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

// submitOwnedUnitAttach establishes durable ownership through the same command
// path used by gameplay, ensuring restore tests cover real relation creation.
func submitOwnedUnitAttach(
	t *testing.T,
	session *gamesession.Session,
	tick uint64,
	spawnID string,
	owner string,
) {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"unit_id":           "monster:" + spawnID,
		"owner_id":          "player:" + owner,
		"ultimate_owner_id": "player:" + owner,
		"durable_id":        "companion:" + spawnID,
		"category": map[string]any{
			"id": "synthetic-room-companion", "group": 1, "base_max": 1,
			"replacement": "replace_oldest", "durable": true, "warp_with_owner": true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := session.Submit(simulation.Command{
		Tick: tick, Player: "owned-unit-fixture", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.owned_unit.attach", Payload: payload,
	}); err != nil {
		t.Fatal(err)
	}
}

// playerPositionX reads the authoritative position by stable player identity,
// avoiding assumptions about ECS allocation order.
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

// playerEntity resolves a player ID to its ECS entity and fails early when the
// lifecycle did not materialize the expected resident.
func playerEntity(t *testing.T, engine *gameecs.Engine, player string) akara.Entity {
	t.Helper()

	controls, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.player_control")
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)

		current, _ := control.Get("player")
		if current == player {
			return entity
		}
	}

	t.Fatalf("player %s was not found", player)

	return 0
}

// setEntityPosition mutates both coordinates as one fixture operation so room
// assignment never observes an intentionally half-updated position.
func setEntityPosition(
	t *testing.T,
	positions *akara.DynamicStore,
	entity akara.Entity,
	x float64,
	y float64,
) {
	t.Helper()

	position, present := positions.Get(entity)
	if !present {
		t.Fatalf("entity %d has no position", entity)
	}

	if err := position.Set("x", x); err != nil {
		t.Fatal(err)
	}

	if err := position.Set("y", y); err != nil {
		t.Fatal(err)
	}
}

// assertResidentRoom checks level and room together because either stale field
// would route activation or serialization to the wrong population partition.
func assertResidentRoom(
	t *testing.T,
	residents *akara.DynamicStore,
	entity akara.Entity,
	levelID int64,
	roomID string,
) {
	t.Helper()

	resident, present := residents.Get(entity)
	if !present {
		t.Fatalf("entity %d has no room resident", entity)
	}

	level, _ := resident.Get("level_id")

	room, _ := resident.Get("room_id")
	if level != levelID || room != roomID {
		t.Fatalf("entity %d resident level/room = %v/%v, want %d/%s", entity, level, room, levelID, roomID)
	}
}

// monsterComponentSnapshot captures the durable monster graph used for parity
// comparisons without depending on entity allocation details.
func monsterComponentSnapshot(
	t *testing.T,
	engine *gameecs.Engine,
	spawnID string,
) map[string]map[string]any {
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

	result := make(map[string]map[string]any, len(retainedMonsterComponents))
	for _, name := range retainedMonsterComponents {
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

// readPopulationPlan crosses the Lua boundary once and decodes a typed plan,
// keeping activation assertions independent of Lua table mechanics.
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

// assertInactiveRoom verifies both room status and the exact serialized resident
// inventory, catching omissions and duplicate inactive snapshots.
func assertInactiveRoom(t *testing.T, authority *Authority, roomID string, residentIDs ...string) {
	t.Helper()

	for _, room := range readPopulationPlan(t, authority).Rooms {
		if room.ID != roomID {
			continue
		}

		residents := decodeInactiveResidents(t, room.InactiveResidents)
		if room.Active || !room.Activated || len(residents) != len(residentIDs) {
			t.Fatalf("inactive room residents = %#v", room)
		}

		want := make(map[string]bool, len(residentIDs))
		for _, id := range residentIDs {
			want[id] = true
		}

		for _, resident := range residents {
			if !want[resident.ID] {
				t.Fatalf("unexpected inactive resident %#v", resident)
			}
		}

		return
	}

	t.Fatalf("population room %s was not found", roomID)
}

// assertInactiveResidentMover checks whether velocity behavior was serialized;
// this determines whether a restored resident resumes movement.
func assertInactiveResidentMover(
	t *testing.T,
	authority *Authority,
	roomID string,
	residentID string,
	want bool,
) {
	t.Helper()

	for _, room := range readPopulationPlan(t, authority).Rooms {
		if room.ID != roomID {
			continue
		}

		for _, resident := range decodeInactiveResidents(t, room.InactiveResidents) {
			if resident.ID == residentID {
				if resident.VelocityMover != want {
					t.Fatalf("inactive resident %s velocity mover = %v, want %v", residentID, resident.VelocityMover, want)
				}

				return
			}
		}

		t.Fatalf("inactive resident %s was not found in room %s", residentID, roomID)
	}

	t.Fatalf("population room %s was not found", roomID)
}

// assertRoomActiveWithoutInactiveResidents ensures reactivation consumes the
// serialized inventory instead of leaving duplicate dormant residents behind.
func assertRoomActiveWithoutInactiveResidents(t *testing.T, authority *Authority, roomID string) {
	t.Helper()

	for _, room := range readPopulationPlan(t, authority).Rooms {
		if room.ID == roomID {
			if !room.Active || !room.Activated || len(decodeInactiveResidents(t, room.InactiveResidents)) != 0 {
				t.Fatalf("restored room state = %#v", room)
			}

			return
		}
	}

	t.Fatalf("population room %s was not found", roomID)
}

// decodeInactiveResidents turns the durable JSON boundary into typed fixtures,
// failing locally when checkpoint schema changes.
func decodeInactiveResidents(t *testing.T, encoded json.RawMessage) []inactiveResidentFixture {
	t.Helper()

	if len(encoded) == 0 || string(encoded) == "{}" {
		return nil
	}

	var residents []inactiveResidentFixture
	if err := json.Unmarshal(encoded, &residents); err != nil {
		t.Fatal(err)
	}

	return residents
}

// generatedPlayerPayload builds a deterministic player-entry command with only
// identity and position exposed as scenario variables.
func generatedPlayerPayload(
	t *testing.T,
	characterID string,
	player string,
	x float64,
	y float64,
) json.RawMessage {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"character_id": characterID, "player": player, "name": characterID, "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "vitality": 20, "defense": 0,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"stamina": 84, "max_stamina": 84,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": x, "y": y,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	return payload
}

// monsterCount reports materialized monsters only, distinguishing active ECS
// residents from serialized inactive-room records.
func monsterCount(engine *gameecs.Engine) int {
	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	return identities.Len()
}

// monsterEntity resolves a stable spawn ID to its current ECS allocation so
// restored instances can be compared without reusing entity numbers.
func monsterEntity(t *testing.T, engine *gameecs.Engine, spawnID string) akara.Entity {
	t.Helper()

	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)

		current, _ := identity.Get("spawn_id")
		if current == spawnID {
			return entity
		}
	}

	t.Fatalf("monster %s was not found", spawnID)

	return 0
}

// componentSnapshot copies one dynamic component into map form, creating a
// stable value boundary before later fixture mutations occur.
func componentSnapshot(
	t *testing.T,
	engine *gameecs.Engine,
	entity akara.Entity,
	name string,
) map[string]any {
	t.Helper()

	store, found := akara.GetDynamicStore(engine.World(), name)
	if !found {
		t.Fatalf("component store %s was not found", name)
	}

	component, present := store.Get(entity)
	if !present {
		t.Fatalf("entity %d has no %s", entity, name)
	}

	values, err := component.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	return values
}

// setMonsterHealth establishes a non-default durable value so restore parity
// cannot pass by merely recreating a fresh monster definition.
func setMonsterHealth(t *testing.T, engine *gameecs.Engine, monster akara.Entity, health int64) {
	t.Helper()

	stats, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")

	component, present := stats.Get(monster)
	if !present {
		t.Fatalf("monster %d has no stats", monster)
	}

	if err := component.Set("health", health); err != nil {
		t.Fatal(err)
	}
}

// corpseComponentSnapshot captures the components that preserve corpse identity,
// ownership, position, and decay behavior across room eviction.
func corpseComponentSnapshot(
	t *testing.T,
	engine *gameecs.Engine,
	corpse akara.Entity,
) map[string]map[string]any {
	t.Helper()

	result := make(map[string]map[string]any, len(retainedCorpseComponents))
	for _, name := range retainedCorpseComponents {
		result[name] = componentSnapshot(t, engine, corpse, name)
	}

	return result
}

// assertCorpseActivation checks active and inactive markers as a complementary
// pair, preventing invalid states where a corpse is both or neither.
func assertCorpseActivation(t *testing.T, engine *gameecs.Engine, corpse akara.Entity, active bool) {
	t.Helper()
	assertResidentActivation(t, engine, corpse, active, false)

	for _, name := range []string{
		"d2legacy.monster.ai",
		"d2legacy.world.collider",
		"engine.world.velocity_mover",
	} {
		store, _ := akara.GetDynamicStore(engine.World(), name)
		if _, present := store.Get(corpse); present {
			t.Fatalf("corpse %d retained %s", corpse, name)
		}
	}

	selectable := componentSnapshot(t, engine, corpse, "d2legacy.world.selectable")
	if selectable["kind"] != "corpse" {
		t.Fatalf("corpse %d selectable = %#v, want corpse kind", corpse, selectable)
	}

	death := componentSnapshot(t, engine, corpse, "d2legacy.monster.death")
	appearance := componentSnapshot(t, engine, corpse, "d2legacy.monster.appearance")

	velocity := componentSnapshot(t, engine, corpse, "d2legacy.world.velocity")
	if death["active"] != false || death["corpse_usable"] != true || appearance["mode"] != "DT" ||
		velocity["x"] != float64(0) || velocity["y"] != float64(0) {
		t.Fatalf("corpse semantic state = death %#v appearance %#v velocity %#v", death, appearance, velocity)
	}
}

// assertOwnedUnitRelation validates both relation endpoints and returns the
// snapshot used for parity checks after re-materialization.
func assertOwnedUnitRelation(
	t *testing.T,
	engine *gameecs.Engine,
	unit akara.Entity,
	owner akara.Entity,
) map[string]any {
	t.Helper()

	relation := componentSnapshot(t, engine, unit, "d2legacy.owned_unit")
	if relation["owner"] != owner || relation["owner_id"] != "player:alice" ||
		relation["ultimate_owner_id"] != "player:alice" || relation["category"] != "synthetic-room-companion" ||
		relation["durable_id"] != "companion:level:2:room:a:monster:1" || relation["durable"] != true ||
		relation["warp_with_owner"] != true || relation["active"] != true {
		t.Fatalf("owned-unit relation = %#v", relation)
	}

	return relation
}

// assertResidentActivation verifies activation markers and optional movement
// scheduling together because rehydration must restore both concerns atomically.
func assertResidentActivation(
	t *testing.T,
	engine *gameecs.Engine,
	entity akara.Entity,
	active bool,
	velocityMover bool,
) {
	t.Helper()

	inactive, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.inactive")
	movers, _ := akara.GetDynamicStore(engine.World(), "engine.world.velocity_mover")
	_, dormant := inactive.Get(entity)

	_, moving := movers.Get(entity)
	if active && (dormant || moving != velocityMover) {
		t.Fatalf("active resident %d inactive/mover = %v/%v, want false/%v", entity, dormant, moving, velocityMover)
	}

	if !active && (!dormant || moving) {
		t.Fatalf("inactive resident %d inactive/mover = %v/%v", entity, dormant, moving)
	}
}

type monsterStateGraph struct {
	Monster  akara.Entity
	Instance akara.Entity
	Source   akara.Entity
	Event    akara.Entity
	Expires  int64
}

// installMonsterTimedState creates linked effect, overlay, and stat-source data
// so inactivation is tested against a realistic multi-entity state graph.
func installMonsterTimedState(
	t *testing.T,
	session *gamesession.Session,
	engine *gameecs.Engine,
	monster akara.Entity,
) monsterStateGraph {
	t.Helper()

	requests, _ := akara.GetDynamicStore(engine.World(), "d2legacy.state.request")

	request := engine.World().MustCreateEntity()
	if _, err := requests.Set(request, map[string]any{
		"operation": "apply", "target": monster, "state_id": "probe-state", "source_id": "probe-source",
		"duration": int64(10000), "policy": "refresh_same_source", "stat": "physical_resist",
		"stat_operation": "add", "stat_value": int64(7), "stat_order": int64(1),
	}); err != nil {
		t.Fatal(err)
	}

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	graph := monsterStateGraph{Monster: monster}

	instances, _ := akara.GetDynamicStore(engine.World(), "d2legacy.state.instance")
	for _, entity := range instances.Entities() {
		instance, _ := instances.Get(entity)
		target, _ := instance.Get("target")

		source, _ := instance.Get("source_id")
		if target == monster && source == "probe-source" {
			graph.Instance = entity
			value, _ := instance.Get("expires_tick")
			graph.Expires = value.(int64)
		}
	}

	sources, _ := akara.GetDynamicStore(engine.World(), "d2legacy.stat.source")
	for _, entity := range sources.Entities() {
		source, _ := sources.Get(entity)
		target, _ := source.Get("target")

		id, _ := source.Get("source_id")
		if target == monster && id == "probe-source" {
			graph.Source = entity
		}
	}

	events, _ := akara.GetDynamicStore(engine.World(), "d2legacy.state.event")
	for _, entity := range events.Entities() {
		event, _ := events.Get(entity)
		target, _ := event.Get("target")

		source, _ := event.Get("source_id")
		if target == monster && source == "probe-source" {
			graph.Event = entity
		}
	}

	if graph.Instance == 0 || graph.Source == 0 || graph.Event == 0 {
		t.Fatalf("incomplete timed-state graph: %#v", graph)
	}

	return graph
}

// assertMonsterStateGraph verifies every linked entity and component survived,
// catching dangling references that a root-monster snapshot would miss.
func assertMonsterStateGraph(t *testing.T, engine *gameecs.Engine, graph monsterStateGraph) {
	t.Helper()

	for name, entity := range map[string]akara.Entity{
		"d2legacy.state.instance": graph.Instance,
		"d2legacy.stat.source":    graph.Source,
		"d2legacy.state.event":    graph.Event,
	} {
		store, _ := akara.GetDynamicStore(engine.World(), name)

		component, present := store.Get(entity)
		if !present {
			t.Fatalf("state graph entity %d lost %s", entity, name)
		}

		target, _ := component.Get("target")
		if target != graph.Monster {
			t.Fatalf("%s target = %v, want monster %d", name, target, graph.Monster)
		}
	}

	instance := componentSnapshot(t, engine, graph.Instance, "d2legacy.state.instance")
	if instance["expires_tick"] != graph.Expires {
		t.Fatalf("timed state expiration = %v, want %d", instance["expires_tick"], graph.Expires)
	}
}

// assertMonsterPlayerCount checks the value pinned at spawn time rather than the
// current population, protecting deterministic scaling across later joins.
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

// stepSession advances an exact bounded number of ticks and fails immediately,
// preserving the intended timing of lifecycle transitions.
func stepSession(t *testing.T, session *gamesession.Session, count int) {
	t.Helper()

	for range count {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
}

// assertCompletedHostileLifecycle checks the final combat and corpse state after
// both original and restored timelines complete equivalent work.
func assertCompletedHostileLifecycle(t *testing.T, engine *gameecs.Engine) {
	t.Helper()

	identities, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.identity")
	deaths, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death")
	events, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.death_event")
	selectables, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.selectable")
	brains, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.ai")
	movers, _ := akara.GetDynamicStore(engine.World(), "engine.world.velocity_mover")
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

	if _, present := movers.Get(monster); present {
		t.Fatal("dead monster retained velocity-mover opt-in")
	}

	foundCorpseTarget := false

	for _, entity := range selectables.Entities() {
		value, _ := selectables.Get(entity)

		id, _ := value.Get("id")
		if id == "monster:level:2:room:blood-moor-a:monster:1" {
			kind, _ := value.Get("kind")
			if kind != "corpse" {
				t.Fatalf("dead monster selectable kind = %v, want corpse", kind)
			}

			foundCorpseTarget = true
		}
	}

	if !foundCorpseTarget {
		t.Fatal("consumable corpse lost its corpse-skill target")
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

// generatedHostileRecords supplies the minimal authoritative data needed by
// lifecycle tests, keeping unrelated archive behavior out of the fixture.
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
		"Id": "fallen", "BaseW": "HTH", "SizeX": "1", "SizeY": "1", "MeleeRng": "1", "corpseSel": "1", "revive": "1",
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
