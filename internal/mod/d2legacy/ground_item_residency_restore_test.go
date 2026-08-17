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
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

func TestGroundItemResidencyRestoresAcrossInactiveRoom(t *testing.T) {
	ctx := context.Background()
	authority, engine, session := startGroundItemFixture(t, nil)
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

	item := groundItemEntity(t, engine, "ground-potion")
	assertResidentRoom(t, dynamicStore(t, engine, "d2legacy.world.room_resident"), item, 2, "a")
	assertGroundItemWorldState(t, engine, item, "world", true)

	submitMoveCommand(t, session, engine.Tick()+1, "alice", 1, 1)
	for playerPositionX(t, engine, "alice") < 20 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
		if engine.Tick() > 110 {
			t.Fatal("player did not leave the ground-item room")
		}
	}
	stopTick := engine.Tick() + 1
	submitMoveCommand(t, session, stopTick, "alice", 2, 0)
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	assertResidentActivation(t, engine, item, false, false)
	assertInactiveRoom(t, authority, "a", "item:alice:ground-potion")
	inactive := groundItemResidencySnapshot(t, engine, item)

	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	private, err := playeradapter.ProjectPrivateView("alice", checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	if len(private.Items.Items) != 0 {
		t.Fatalf("inactive ground item leaked into private presentation: %#v", private.Items.Items)
	}

	restored, restoredEngine, restoredSession := startGroundItemFixture(t, &checkpoint)
	t.Cleanup(func() {
		_ = restored.Stop(ctx)
		_ = restoredSession.Close()
		_ = restoredEngine.Close()
	})
	if got := groundItemResidencySnapshot(t, restoredEngine, item); !reflect.DeepEqual(got, inactive) {
		t.Fatalf("restored inactive ground item = %#v, want %#v", got, inactive)
	}
	assertInactiveRoom(t, restored, "a", "item:alice:ground-potion")

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
	assertResidentActivation(t, engine, item, true, false)
	assertResidentActivation(t, restoredEngine, item, true, false)
	assertEqualLatestChecksums(t, session, restoredSession, "reactivated ground item")

	moveItem(t, session, engine.Tick()+1, 5, "held", 0, 0)
	moveItem(t, restoredSession, restoredEngine.Tick()+1, 3, "held", 0, 0)
	stepSession(t, session, 1)
	stepSession(t, restoredSession, 1)
	assertGroundItemWorldState(t, engine, item, "held", false)
	assertGroundItemWorldState(t, restoredEngine, item, "held", false)
	assertEqualLatestChecksums(t, session, restoredSession, "picked-up ground item")

	moveItem(t, session, engine.Tick()+1, 6, "world", 2, 1)
	moveItem(t, restoredSession, restoredEngine.Tick()+1, 4, "world", 2, 1)
	stepSession(t, session, 1)
	stepSession(t, restoredSession, 1)
	assertGroundItemWorldState(t, engine, item, "world", true)
	assertGroundItemWorldState(t, restoredEngine, item, "world", true)
	assertResidentRoom(t, dynamicStore(t, engine, "d2legacy.world.room_resident"), item, 2, "a")
	assertResidentRoom(t, dynamicStore(t, restoredEngine, "d2legacy.world.room_resident"), item, 2, "a")
	assertEqualLatestChecksums(t, session, restoredSession, "redropped ground item")
}

func startGroundItemFixture(t *testing.T, checkpoint *simulation.Checkpoint) (*Authority, *gameecs.Engine, *gamesession.Session) {
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
	initial := map[string]any{"d2legacy.items": map[string]any{
		"owner": "alice", "inventory_width": 10, "inventory_height": 4,
		"stash_width": 6, "stash_height": 8, "cube_width": 3, "cube_height": 4,
		"belt_capacity": 4, "vendor_width": 10, "vendor_height": 10,
		"items": []any{map[string]any{
			"id": "ground-potion", "code": "hp1", "width": 1, "height": 1,
			"container": "world", "x": 1, "y": 1, "act": 1, "level_id": 2,
			"world_dc6": "data/global/items/flphp1.dc6", "world_animated": true,
		}},
	}}
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), generatedHostileRecords(), engine, session,
		Config{Seed: 315, InitialData: initial, Restore: restore})
	if err != nil {
		_ = session.Close()
		_ = engine.Close()
		t.Fatal(err)
	}
	return authority, engine, session
}

func groundItemEntity(t *testing.T, engine *gameecs.Engine, itemID string) akara.Entity {
	t.Helper()
	identities := dynamicStore(t, engine, "d2legacy.item.identity")
	for _, entity := range identities.Entities() {
		identity, _ := identities.Get(entity)
		value, _ := identity.Get("id")
		if value == itemID {
			return entity
		}
	}
	t.Fatalf("ground item %q was not found", itemID)
	return 0
}

func groundItemResidencySnapshot(t *testing.T, engine *gameecs.Engine, item akara.Entity) map[string]map[string]any {
	t.Helper()
	result := make(map[string]map[string]any, 7)
	for _, name := range []string{
		"d2legacy.item.identity",
		"d2legacy.item.placement",
		"d2legacy.item.presentation",
		"d2legacy.world.position",
		"d2legacy.world.location",
		"d2legacy.world.room_resident",
		"d2legacy.world.inactive",
	} {
		result[name] = componentSnapshot(t, engine, item, name)
	}
	return result
}

func assertGroundItemWorldState(t *testing.T, engine *gameecs.Engine, item akara.Entity, container string, spatial bool) {
	t.Helper()
	placement := componentSnapshot(t, engine, item, "d2legacy.item.placement")
	if placement["container"] != container {
		t.Fatalf("item container = %v, want %s", placement["container"], container)
	}
	for _, name := range []string{
		"d2legacy.world.position", "d2legacy.world.location", "d2legacy.world.room_resident",
	} {
		_, present := dynamicStore(t, engine, name).Get(item)
		if present != spatial {
			t.Fatalf("item component %s present = %t, want %t", name, present, spatial)
		}
	}
}

func moveItem(t *testing.T, session *gamesession.Session, tick, sequence uint64, container string, x, y int) {
	t.Helper()
	submitPartyCommand(t, session, tick, "alice", sequence, "item.move", map[string]any{
		"item_id":     "ground-potion",
		"destination": map[string]any{"container": container, "x": x, "y": y},
	})
}

func assertEqualLatestChecksums(t *testing.T, original, restored *gamesession.Session, label string) {
	t.Helper()
	originalReplay, err := original.Replay()
	if err != nil {
		t.Fatal(err)
	}
	restoredReplay, err := restored.Replay()
	if err != nil {
		t.Fatal(err)
	}
	got := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1].Checksum
	want := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1].Checksum
	if got != want {
		t.Fatalf("%s checksum = %s, want %s", label, got, want)
	}
}
