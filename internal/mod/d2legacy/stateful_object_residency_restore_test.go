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
)

func TestStatefulObjectAndPendingActionRestoreAcrossInactiveRoom(t *testing.T) {
	ctx := context.Background()
	authority, engine, session := startStatefulObjectFixture(t, nil)
	t.Cleanup(func() {
		_ = authority.Stop(ctx)
		_ = session.Close()
		_ = engine.Close()
	})

	population, _ := json.Marshal(map[string]any{
		"act": 1, "level_id": 2, "difficulty": 0,
		"links": []map[string]any{{"from": "a", "to": "b"}, {"from": "b", "to": "c"}},
		"rooms": []map[string]any{
			{"id": "a", "populate": false, "x": 0, "y": 0, "width": 10, "height": 10,
				"points": []map[string]any{}},
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

	object := interactionTargetEntity(t, engine, "fixture-object")
	action := pendingObjectActionEntity(t, engine, "delayed-fixture")
	assertResidentRoom(t, dynamicStore(t, engine, "d2legacy.world.room_resident"), object, 2, "a")
	assertResidentRoom(t, dynamicStore(t, engine, "d2legacy.world.room_resident"), action, 2, "a")
	assertPendingActionTarget(t, engine, action, object)

	openObject(t, session, 3, 1)
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	assertObjectState(t, engine, object, "open", true, 8)

	submitMoveCommand(t, session, engine.Tick()+1, "alice", 2, 1)
	for playerPositionX(t, engine, "alice") < 20 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
		if engine.Tick() > 115 {
			t.Fatal("player did not leave the stateful-object room")
		}
	}
	stopTick := engine.Tick() + 1
	submitMoveCommand(t, session, stopTick, "alice", 3, 0)
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	assertResidentActivation(t, engine, object, false, false)
	assertResidentActivation(t, engine, action, false, false)
	assertInactiveRoom(t, authority, "a",
		"level:2:fixture-object", "object-action:fixture-object:delayed-fixture")
	objectInactive := statefulObjectSnapshot(t, engine, object)
	actionInactive := pendingObjectActionSnapshot(t, engine, action)

	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	restored, restoredEngine, restoredSession := startStatefulObjectFixture(t, &checkpoint)
	t.Cleanup(func() {
		_ = restored.Stop(ctx)
		_ = restoredSession.Close()
		_ = restoredEngine.Close()
	})
	if got := statefulObjectSnapshot(t, restoredEngine, object); !reflect.DeepEqual(got, objectInactive) {
		t.Fatalf("restored inactive object = %#v, want %#v", got, objectInactive)
	}
	if got := pendingObjectActionSnapshot(t, restoredEngine, action); !reflect.DeepEqual(got, actionInactive) {
		t.Fatalf("restored inactive object action = %#v, want %#v", got, actionInactive)
	}
	assertPendingActionTarget(t, restoredEngine, action, object)
	assertInactiveRoom(t, restored, "a",
		"level:2:fixture-object", "object-action:fixture-object:delayed-fixture")

	returnTick := checkpoint.Tick + 1
	submitMoveCommand(t, session, returnTick, "alice", 4, -1)
	submitMoveCommand(t, restoredSession, returnTick, "alice", 1, -1)
	stepSession(t, session, 55)
	stepSession(t, restoredSession, 55)
	stopTick = returnTick + 55
	submitMoveCommand(t, session, stopTick, "alice", 5, 0)
	submitMoveCommand(t, restoredSession, stopTick, "alice", 2, 0)
	stepSession(t, session, 2)
	stepSession(t, restoredSession, 2)
	assertResidentActivation(t, engine, object, true, false)
	assertResidentActivation(t, engine, action, true, false)
	assertResidentActivation(t, restoredEngine, object, true, false)
	assertResidentActivation(t, restoredEngine, action, true, false)
	assertEqualLatestChecksums(t, session, restoredSession, "reactivated object graph")

	// The fixture behavior is one-shot. Re-operating after reactivation must
	// preserve the committed mode/revision on both original and restored runs.
	openObject(t, session, engine.Tick()+1, 6)
	openObject(t, restoredSession, restoredEngine.Tick()+1, 3)
	stepSession(t, session, 1)
	stepSession(t, restoredSession, 1)
	assertObjectState(t, engine, object, "open", true, 8)
	assertObjectState(t, restoredEngine, object, "open", true, 8)
	assertEqualLatestChecksums(t, session, restoredSession, "reoperated object graph")
}

func startStatefulObjectFixture(t *testing.T, checkpoint *simulation.Checkpoint) (*Authority, *gameecs.Engine, *gamesession.Session) {
	t.Helper()
	var engine *gameecs.Engine
	var restore []simulation.ParticipantState
	if checkpoint == nil {
		engine = gameecs.New()
	} else {
		var err error
		engine, err = gameecs.RestoreSnapshot(*checkpoint.Snapshot)
		if err != nil {
			t.Fatal(err)
		}
		restore = checkpoint.Participants
	}
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	initial := map[string]any{"d2legacy.interactions": map[string]any{
		"owner": "alice",
		"targets": []any{map[string]any{
			"id": "fixture-object", "npc": "Fixture Object", "x": 1, "y": 1, "radius": 4,
			"resident_id": "level:2:fixture-object", "level_id": 2, "room_id": "a",
			"object_state": map[string]any{
				"definition_id": "synthetic-once", "mode": "closed", "seed": 73,
				"revision": 7, "once_result_mode": "open",
			},
			"pending_actions": []any{map[string]any{
				"id": "delayed-fixture", "kind": "synthetic-event", "due_tick": 1000,
				"sequence": 2, "active": true,
			}},
		}},
	}}
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), generatedHostileRecords(), engine, session,
		Config{Seed: 316, InitialData: initial, Restore: restore})
	if err != nil {
		_ = session.Close()
		_ = engine.Close()
		t.Fatal(err)
	}
	return authority, engine, session
}

func interactionTargetEntity(t *testing.T, engine *gameecs.Engine, targetID string) akara.Entity {
	t.Helper()
	targets := dynamicStore(t, engine, "d2legacy.interaction.target")
	for _, entity := range targets.Entities() {
		target, _ := targets.Get(entity)
		value, _ := target.Get("id")
		if value == targetID {
			return entity
		}
	}
	t.Fatalf("interaction target %q was not found", targetID)
	return 0
}

func pendingObjectActionEntity(t *testing.T, engine *gameecs.Engine, actionID string) akara.Entity {
	t.Helper()
	actions := dynamicStore(t, engine, "d2legacy.object.pending_action")
	for _, entity := range actions.Entities() {
		action, _ := actions.Get(entity)
		value, _ := action.Get("id")
		if value == actionID {
			return entity
		}
	}
	t.Fatalf("pending object action %q was not found", actionID)
	return 0
}

func openObject(t *testing.T, session *gamesession.Session, tick, sequence uint64) {
	t.Helper()
	submitPartyCommand(t, session, tick, "alice", sequence, "interaction.open", map[string]any{
		"target": "fixture-object",
	})
}

func assertObjectState(t *testing.T, engine *gameecs.Engine, object akara.Entity, mode string, used bool, revision int64) {
	t.Helper()
	state := componentSnapshot(t, engine, object, "d2legacy.object.state")
	if state["mode"] != mode || state["used"] != used || state["revision"] != revision {
		t.Fatalf("object state = %#v, want mode=%s used=%t revision=%d", state, mode, used, revision)
	}
}

func assertPendingActionTarget(t *testing.T, engine *gameecs.Engine, action, object akara.Entity) {
	t.Helper()
	component, _ := dynamicStore(t, engine, "d2legacy.object.pending_action").Get(action)
	value, err := component.Get("target")
	if err != nil {
		t.Fatal(err)
	}
	if value != object {
		t.Fatalf("pending action target = %v, want %v", value, object)
	}
}

func statefulObjectSnapshot(t *testing.T, engine *gameecs.Engine, object akara.Entity) map[string]map[string]any {
	t.Helper()
	result := make(map[string]map[string]any, 5)
	for _, name := range []string{
		"d2legacy.interaction.target",
		"d2legacy.object.state",
		"d2legacy.object.once_operation",
		"d2legacy.world.room_resident",
		"d2legacy.world.inactive",
	} {
		result[name] = componentSnapshot(t, engine, object, name)
	}
	return result
}

func pendingObjectActionSnapshot(t *testing.T, engine *gameecs.Engine, action akara.Entity) map[string]map[string]any {
	t.Helper()
	result := make(map[string]map[string]any, 3)
	for _, name := range []string{
		"d2legacy.object.pending_action",
		"d2legacy.world.room_resident",
		"d2legacy.world.inactive",
	} {
		result[name] = componentSnapshot(t, engine, action, name)
	}
	return result
}
