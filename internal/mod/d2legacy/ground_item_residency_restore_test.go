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
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	lua "github.com/yuin/gopher-lua"
)

// TestGroundItemResidencyRestoresAcrossInactiveRoom verifies inactive-room
// serialization preserves item ownership, placement, and checksum parity.
func TestGroundItemResidencyRestoresAcrossInactiveRoom(t *testing.T) {
	ctx := context.Background()
	authority, engine, session := startGroundItemFixture(t, nil, "ground-potion", "world")
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
	groundTarget := componentSnapshot(t, engine, item, "d2legacy.world.selectable")["id"].(string)
	assertResidentRoom(t, dynamicStore(t, engine, "d2legacy.world.room_resident"), item, 2, "a")
	assertGroundItemWorldState(t, engine, item, "world", true)
	checkpointView, err := session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	public, err := playeradapter.ProjectWorldView("alice", checkpointView)
	if err != nil {
		t.Fatal(err)
	}
	var view playeradapter.WorldView
	if err := json.Unmarshal(public, &view); err != nil {
		t.Fatal(err)
	}
	var visible bool
	for _, entity := range view.Entities {
		if entity.ID == groundTarget {
			visible = entity.Kind == "item" && entity.WorldDC6 == "data/global/items/flphp1.dc6" && entity.WorldAnimated
		}
	}
	if !visible {
		t.Fatalf("ground-item art absent from public view: %+v", view.Entities)
	}

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
	assertInactiveRoom(t, authority, "a", groundTarget)
	inactive := groundItemResidencySnapshot(t, engine, item)

	recovery, err := session.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	checkpoint := recovery.State

	private, err := playeradapter.ProjectPrivateView("alice", checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	if len(private.Items.Items) != 0 {
		t.Fatalf("inactive ground item leaked into private presentation: %#v", private.Items.Items)
	}

	restored, restoredEngine, restoredSession := startGroundItemFixture(t, &checkpoint, "ground-potion", "world")
	if err := restoredSession.RestoreRecoveryCheckpoint(recovery); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = restored.Stop(ctx)
		_ = restoredSession.Close()
		_ = restoredEngine.Close()
	})

	if got := groundItemResidencySnapshot(t, restoredEngine, item); !reflect.DeepEqual(got, inactive) {
		t.Fatalf("restored inactive ground item = %#v, want %#v", got, inactive)
	}

	assertInactiveRoom(t, restored, "a", groundTarget)

	returnTick := checkpoint.Tick + 1
	submitMoveCommand(t, session, returnTick, "alice", 3, -1)
	submitGroundContinuation(t, restoredSession, returnTick, 3, "player.move", map[string]any{"x": -1, "y": 0, "running": false})
	stepSession(t, session, 75)
	stepSession(t, restoredSession, 75)

	stopTick = returnTick + 75
	submitMoveCommand(t, session, stopTick, "alice", 4, 0)
	submitGroundContinuation(t, restoredSession, stopTick, 4, "player.move", map[string]any{"x": 0, "y": 0, "running": false})
	stepSession(t, session, 2)
	stepSession(t, restoredSession, 2)
	assertResidentActivation(t, engine, item, true, false)
	assertResidentActivation(t, restoredEngine, item, true, false)
	assertEqualLatestChecksums(t, session, restoredSession, "reactivated ground item")
	if x := playerPositionX(t, engine, "alice"); x > 4.5 {
		t.Fatalf("player has not approached ground item: x=%g", x)
	}

	moveItem(t, session, engine.Tick()+1, 5, "held", 0, 0)
	submitGroundContinuation(t, restoredSession, restoredEngine.Tick()+1, 5, "item.move", map[string]any{
		"item_id": "ground-potion", "destination": map[string]any{"container": "held", "x": 0, "y": 0},
	})
	stepSession(t, session, 1)
	stepSession(t, restoredSession, 1)
	assertGroundItemWorldState(t, engine, item, "held", false)
	assertGroundItemWorldState(t, restoredEngine, item, "held", false)
	assertEqualLatestChecksums(t, session, restoredSession, "picked-up ground item")

	moveItem(t, session, engine.Tick()+1, 6, "world", 2, 1)
	submitGroundContinuation(t, restoredSession, restoredEngine.Tick()+1, 6, "item.move", map[string]any{
		"item_id": "ground-potion", "destination": map[string]any{"container": "world", "x": 2, "y": 1},
	})
	stepSession(t, session, 1)
	stepSession(t, restoredSession, 1)
	assertGroundItemWorldState(t, engine, item, "world", true)
	assertGroundItemWorldState(t, restoredEngine, item, "world", true)
	assertResidentRoom(t, dynamicStore(t, engine, "d2legacy.world.room_resident"), item, 2, "a")
	assertResidentRoom(t, dynamicStore(t, restoredEngine, "d2legacy.world.room_resident"), item, 2, "a")
	assertEqualLatestChecksums(t, session, restoredSession, "redropped ground item")
}

func TestGroundItemMaximumIdentityDropPickupAndRestore(t *testing.T) {
	id := strings.Repeat("i", 128)
	authority, engine, session := startGroundItemFixture(t, nil, id, "inventory")
	t.Cleanup(func() { _ = authority.Stop(t.Context()); _ = session.Close(); _ = engine.Close() })
	admitGroundItemPlayer(t, session, "alice", 1, 1)
	checkedGroundItemView(t, session, "alice")
	submitPartyCommand(t, session, engine.Tick()+1, "alice", 1, "item.move", map[string]any{
		"item_id": id, "destination": map[string]any{"container": "world", "x": 1, "y": 1},
	})
	stepSession(t, session, 1)
	item := groundItemEntity(t, engine, id)
	selector := componentSnapshot(t, engine, item, "d2legacy.world.selectable")["id"].(string)
	if len(selector) > 128 {
		t.Fatalf("maximum durable ID created %d-byte selector", len(selector))
	}
	checkedGroundItemView(t, session, "alice")
	callGroundIdentityFixture(t, authority, "assert_selected", selector)
	recovery, err := session.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := recovery.State
	restored, restoredEngine, restoredSession := startGroundItemFixture(t, &checkpoint, id, "inventory")
	if err := restoredSession.RestoreRecoveryCheckpoint(recovery); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = restored.Stop(t.Context()); _ = restoredSession.Close(); _ = restoredEngine.Close() })
	for _, timeline := range []struct {
		engine  *gameecs.Engine
		session *gamesession.Session
	}{{engine, session}, {restoredEngine, restoredSession}} {
		sequence := uint64(1)
		if got := componentSnapshot(t, timeline.engine, item, "d2legacy.world.selectable")["id"]; got != selector {
			t.Fatalf("restored selector = %v, want %s", got, selector)
		}
		submitGroundContinuation(t, timeline.session, timeline.engine.Tick()+1, sequence+1, "item.pickup", map[string]any{"target_id": selector})
		stepSession(t, timeline.session, 1)
		view := checkedGroundItemView(t, timeline.session, "alice")
		if len(view.Private.Items.Items) != 1 || view.Private.Items.Items[0].ID != id {
			t.Fatal("pickup shortened or lost the durable item identity")
		}
		submitGroundContinuation(t, timeline.session, timeline.engine.Tick()+1, sequence+2, "item.move", map[string]any{
			"item_id": id, "destination": map[string]any{"container": "world", "x": 1, "y": 1},
		})
		stepSession(t, timeline.session, 1)
		next := componentSnapshot(t, timeline.engine, item, "d2legacy.world.selectable")["id"].(string)
		if next == selector || len(next) > 128 {
			t.Fatal("redrop reused the retired selector or exceeded the public bound")
		}
		before := groundItemResidencySnapshot(t, timeline.engine, item)
		submitGroundContinuation(t, timeline.session, timeline.engine.Tick()+1, sequence+3, "item.pickup", map[string]any{"target_id": selector})
		stepSession(t, timeline.session, 1)
		if !reflect.DeepEqual(before, groundItemResidencySnapshot(t, timeline.engine, item)) {
			t.Fatal("stale selector picked up a later ground incarnation")
		}
		submitGroundContinuation(t, timeline.session, timeline.engine.Tick()+1, sequence+4, "item.pickup", map[string]any{"target_id": next})
		stepSession(t, timeline.session, 1)
		checkedGroundItemView(t, timeline.session, "alice")
		if componentSnapshot(t, timeline.engine, item, "d2legacy.item.placement")["container"] != "inventory" {
			t.Fatal("fresh redrop selector did not collect the exact item")
		}
	}
	assertEqualLatestChecksums(t, session, restoredSession, "maximum ID pickup and redrop")
}

func TestGroundItemPickupPreservesOwnerLocalIdentities(t *testing.T) {
	for _, existingID := range []string{"sword", "other"} {
		t.Run(existingID, func(t *testing.T) {
			authority, engine, session := startGroundItemFixture(t, nil, "sword", "world")
			t.Cleanup(func() { _ = authority.Stop(t.Context()); _ = session.Close(); _ = engine.Close() })
			admitGroundItemPlayer(t, session, "alice", 1, 1)
			admitGroundItemPlayer(t, session, "bob", 2, 4)
			callGroundIdentityFixture(t, authority, "create_destination_item", existingID)
			checkedGroundItemView(t, session, "alice")
			checkedGroundItemView(t, session, "bob")
			item := groundItemEntity(t, engine, "sword")
			var unrelated akara.Entity
			for _, entity := range dynamicStore(t, engine, "d2legacy.item.identity").Entities() {
				if entity != item {
					unrelated = entity
				}
			}
			if unrelated == 0 {
				t.Fatal("destination fixture item is missing")
			}
			unrelatedBefore := groundItemResidencySnapshot(t, engine, unrelated)
			before := groundItemResidencySnapshot(t, engine, item)
			selector := before["d2legacy.world.selectable"]["id"].(string)
			submitPartyCommand(t, session, engine.Tick()+1, "bob", 1, "item.pickup", map[string]any{"target_id": selector})
			stepSession(t, session, 1)
			view := checkedGroundItemView(t, session, "bob")
			after := groundItemResidencySnapshot(t, engine, item)
			if existingID == "sword" {
				if !reflect.DeepEqual(before, after) || len(view.Private.Items.Items) != 1 {
					t.Fatal("destination ID collision changed the ground item or duplicated an owner-local ID")
				}
			} else if len(view.Private.Items.Items) != 2 || after["d2legacy.item.placement"]["container"] != "inventory" {
				t.Fatal("distinct owner-local identity did not transfer")
			}
			if !reflect.DeepEqual(before["d2legacy.item.generation"], after["d2legacy.item.generation"]) {
				t.Fatal("pickup changed rolled generation facts")
			}
			submitPartyCommand(t, session, engine.Tick()+1, "bob", 2, "item.pickup", map[string]any{"target_id": selector})
			stepSession(t, session, 1)
			checkedGroundItemView(t, session, "bob")
			if !reflect.DeepEqual(after, groundItemResidencySnapshot(t, engine, item)) {
				t.Fatal("repeated pickup changed item identity or generation")
			}
			if !reflect.DeepEqual(unrelatedBefore, groundItemResidencySnapshot(t, engine, unrelated)) {
				t.Fatal("pickup mutated the unrelated destination item")
			}
			if existingID == "other" {
				submitPartyCommand(t, session, engine.Tick()+1, "bob", 3, "item.move", map[string]any{
					"item_id": "sword", "destination": map[string]any{"container": "world", "x": 4, "y": 1},
				})
				stepSession(t, session, 1)
				redropped := groundItemResidencySnapshot(t, engine, item)
				fresh := redropped["d2legacy.world.selectable"]["id"].(string)
				if fresh == selector {
					t.Fatal("ownership transfer and redrop reused the old target")
				}
				submitPartyCommand(t, session, engine.Tick()+1, "alice", 1, "item.pickup", map[string]any{"target_id": selector})
				stepSession(t, session, 1)
				if !reflect.DeepEqual(redropped, groundItemResidencySnapshot(t, engine, item)) {
					t.Fatal("retired source-owner target collected the transferred redrop")
				}
				submitPartyCommand(t, session, engine.Tick()+1, "bob", 4, "item.pickup", map[string]any{"target_id": fresh})
				stepSession(t, session, 1)
				checkedGroundItemView(t, session, "bob")
				if componentSnapshot(t, engine, item, "d2legacy.item.placement")["container"] != "inventory" {
					t.Fatal("current owner could not pick up the transferred redrop")
				}
			}
		})
	}
}

func callGroundIdentityFixture(t *testing.T, authority *Authority, action, value string) {
	t.Helper()
	if err := authority.Runtime.Run(t.Context(), func(state *lua.LState) error {
		if err := state.CallByParam(lua.P{Fn: state.GetGlobal("require"), NRet: 1, Protect: true},
			lua.LString("d2legacy.tests.integration.item_identity")); err != nil {
			return err
		}
		module := state.Get(-1)
		state.Pop(1)
		return state.CallByParam(lua.P{Fn: state.GetField(module, action), Protect: true}, lua.LString(value))
	}); err != nil {
		t.Fatal(err)
	}
}

func admitGroundItemPlayer(t *testing.T, session *gamesession.Session, owner string, sequence uint64, x float64) {
	t.Helper()
	checkpoint, err := session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(simulation.Command{Tick: checkpoint.Tick + 1, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: sequence, Kind: "system.player.enter", Payload: generatedPlayerPayload(t, owner, owner, x, 1)}); err != nil {
		t.Fatal(err)
	}
	stepSession(t, session, 1)
}

func checkedGroundItemView(t *testing.T, session *gamesession.Session, owner string) playeradapter.ClientView {
	t.Helper()
	checkpoint, err := session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := playeradapter.ProjectClientView(owner, checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	var view playeradapter.ClientView
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	if err := playeradapter.ValidateClientView(view, checkpoint.Tick); err != nil {
		t.Fatal(err)
	}
	return view
}

// startGroundItemFixture boots either a fresh or restored authority with the
// same seed, making checkpoint state the only difference between both runs.
func startGroundItemFixture(
	t *testing.T,
	checkpoint *simulation.Checkpoint,
	itemID, container string,
) (*Authority, *gameecs.Engine, *gamesession.Session) {
	t.Helper()

	var (
		engine  *gameecs.Engine
		restore []simulation.ParticipantState
	)

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
			"id": itemID, "code": "hp1", "width": 1, "height": 1,
			"container": container, "x": 1, "y": 1, "act": 1, "level_id": 2,
			"world_dc6": "data/global/items/flphp1.dc6", "world_animated": true,
			"generation": map[string]any{"quality": "normal", "level": 1, "facts": "fixture-generation"},
		}},
	}}

	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), generatedHostileRecords(), engine, session,
		Config{Seed: 315, InitialData: initial, Restore: restore})
	if err != nil {
		_ = session.Close()
		_ = engine.Close()

		t.Fatal(err)
	}
	geometry, err := gameworld.NewOpenMap(30, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := modruntime.SetWorldMapForLevel(t.Context(), authority.Runtime,
		"d2legacy.gameplay.systems.init", "set_collision", 2, geometry); err != nil {
		t.Fatal(err)
	}

	return authority, engine, session
}

// groundItemEntity resolves a stable item identifier to its ECS entity so
// assertions do not depend on allocation order.
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

// groundItemResidencySnapshot captures every component that must survive room
// inactivation, providing one parity boundary for original and restored runs.
func groundItemResidencySnapshot(
	t *testing.T,
	engine *gameecs.Engine,
	item akara.Entity,
) map[string]map[string]any {
	t.Helper()

	result := make(map[string]map[string]any, 7)
	for _, name := range []string{
		"d2legacy.item.identity",
		"d2legacy.item.placement",
		"d2legacy.item.presentation",
		"d2legacy.item.generation",
		"d2legacy.item.origin",
		"d2legacy.world.selectable",
		"d2legacy.world.position",
		"d2legacy.world.location",
		"d2legacy.world.room_resident",
		"d2legacy.world.inactive",
	} {
		if _, present := dynamicStore(t, engine, name).Get(item); present {
			result[name] = componentSnapshot(t, engine, item, name)
		}
	}

	return result
}

// assertGroundItemWorldState checks the mutually dependent container and
// spatial markers; testing them together catches impossible hybrid placement.
func assertGroundItemWorldState(
	t *testing.T,
	engine *gameecs.Engine,
	item akara.Entity,
	container string,
	spatial bool,
) {
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

// moveItem submits a deterministic authority command and fails at the call site
// so ordering mistakes remain easy to diagnose.
func moveItem(
	t *testing.T,
	session *gamesession.Session,
	tick uint64,
	sequence uint64,
	container string,
	x int,
	y int,
) {
	t.Helper()
	submitPartyCommand(t, session, tick, "alice", sequence, "item.move", map[string]any{
		"item_id":     "ground-potion",
		"destination": map[string]any{"container": container, "x": x, "y": y},
	})
}

// assertEqualLatestChecksums proves both timelines converged after equivalent
// commands, covering state that individual component assertions may omit.
func assertEqualLatestChecksums(
	t *testing.T,
	original *gamesession.Session,
	restored *gamesession.Session,
	label string,
) {
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
