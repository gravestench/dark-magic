package d2legacy

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// TestProjectileResidencyRestoresAcrossInactiveRoom verifies moving projectile
// state pauses, serializes, and resumes identically across checkpoint restore.
func TestProjectileResidencyRestoresAcrossInactiveRoom(t *testing.T) {
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

	cast, _ := json.Marshal(map[string]any{
		"side": "left", "target_x": 8, "target_y": 1, "target_id": "",
	})
	if err := session.Submit(simulation.Command{Tick: 3, Player: "alice", Authority: simulation.AuthorityPlayer,
		Sequence: 1, Kind: "player.use_skill", Payload: cast}); err != nil {
		t.Fatal(err)
	}

	projectile := waitForProjectile(t, session, engine, 10)
	assertResidentRoom(t, dynamicStore(t, engine, "d2legacy.world.room_resident"), projectile, 2, "a")
	resident := componentSnapshot(t, engine, projectile, "d2legacy.world.room_resident")

	residentID, _ := resident["id"].(string)
	if !strings.HasPrefix(residentID, "projectile:player:alice:skill:36:effect:") {
		t.Fatalf("projectile resident ID = %q", residentID)
	}

	setProjectileMotion(t, engine, projectile, 0, 0, 1000)
	submitMoveCommand(t, session, engine.Tick()+1, "alice", 2, 1)

	for playerPositionX(t, engine, "alice") < 20 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}

		if engine.Tick() > 110 {
			t.Fatal("player did not leave the projectile room")
		}
	}

	stopTick := engine.Tick() + 1
	submitMoveCommand(t, session, stopTick, "alice", 3, 0)

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	assertResidentActivation(t, engine, projectile, false, false)
	assertInactiveRoom(t, authority, "a", residentID)

	setProjectileMotion(t, engine, projectile, 1, 0, 1000)
	positionBefore := componentSnapshot(t, engine, projectile, "d2legacy.world.position")
	projectileBefore := componentSnapshot(t, engine, projectile, "d2legacy.missile.projectile")
	stepSession(t, session, 5)

	gotPosition := componentSnapshot(t, engine, projectile, "d2legacy.world.position")
	if !reflect.DeepEqual(gotPosition, positionBefore) {
		t.Fatalf("inactive projectile moved: got %#v, want %#v", gotPosition, positionBefore)
	}

	gotProjectile := componentSnapshot(t, engine, projectile, "d2legacy.missile.projectile")
	if !reflect.DeepEqual(gotProjectile, projectileBefore) {
		t.Fatalf("inactive projectile lifetime advanced: got %#v, want %#v", gotProjectile, projectileBefore)
	}

	setProjectileMotion(t, engine, projectile, 0, 0, 1000)

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	inactive := projectileResidencySnapshot(t, engine, projectile)

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
	assertResidentActivation(t, restoredEngine, projectile, false, false)

	if got := projectileResidencySnapshot(t, restoredEngine, projectile); !reflect.DeepEqual(got, inactive) {
		t.Fatalf("restored inactive projectile = %#v, want %#v", got, inactive)
	}

	assertInactiveRoom(t, restored, "a", residentID)

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
		t.Fatalf(
			"inactive projectile restore checksum = %s, want %s",
			restoredCheckpoint.Checksum,
			originalCheckpoint.Checksum,
		)
	}

	assertResidentActivation(t, engine, projectile, true, false)
	assertResidentActivation(t, restoredEngine, projectile, true, false)
	got := projectileResidencySnapshot(t, restoredEngine, projectile)

	want := projectileResidencySnapshot(t, engine, projectile)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("restored reactivated projectile = %#v, want %#v", got, want)
	}
}

// waitForProjectile advances a bounded number of deterministic ticks until the
// projectile exists; the limit prevents a broken spawn from hanging the suite.
func waitForProjectile(
	t *testing.T,
	session *gamesession.Session,
	engine *gameecs.Engine,
	limit int,
) akara.Entity {
	t.Helper()
	projectiles := dynamicStore(t, engine, "d2legacy.missile.projectile")

	for range limit {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}

		if entities := projectiles.Entities(); len(entities) == 1 {
			return entities[0]
		}
	}

	t.Fatal("straight-missile cast did not materialize a projectile")

	return 0
}

// dynamicStore resolves a required component store and fails immediately so
// later fixture mutations cannot accidentally operate on a missing schema.
func dynamicStore(t *testing.T, engine *gameecs.Engine, name string) *akara.DynamicStore {
	t.Helper()

	store, found := akara.GetDynamicStore(engine.World(), name)
	if !found {
		t.Fatalf("component store %s was not found", name)
	}

	return store
}

// setProjectileMotion installs a known trajectory before inactivation, making
// resumed position and remaining-lifetime comparisons deterministic.
func setProjectileMotion(
	t *testing.T,
	engine *gameecs.Engine,
	projectile akara.Entity,
	velocityX float64,
	velocityY float64,
	ticks int64,
) {
	t.Helper()

	component, present := dynamicStore(t, engine, "d2legacy.missile.projectile").Get(projectile)
	if !present {
		t.Fatalf("projectile %d was not found", projectile)
	}

	for name, value := range map[string]any{
		"velocity_x": velocityX, "velocity_y": velocityY, "remaining_ticks": ticks,
	} {
		if err := component.Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
}

// projectileResidencySnapshot captures the components that jointly define
// projectile identity, location, movement, and expiration across inactivation.
func projectileResidencySnapshot(
	t *testing.T,
	engine *gameecs.Engine,
	projectile akara.Entity,
) map[string]map[string]any {
	t.Helper()

	result := make(map[string]map[string]any, 4)
	for _, name := range []string{
		"d2legacy.missile.projectile",
		"d2legacy.world.position",
		"d2legacy.world.location",
		"d2legacy.world.room_resident",
	} {
		result[name] = componentSnapshot(t, engine, projectile, name)
	}

	return result
}
