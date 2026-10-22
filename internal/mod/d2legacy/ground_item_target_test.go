package d2legacy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func submitGroundContinuation(t *testing.T, session *gamesession.Session, tick, sequence uint64, kind string, payload map[string]any) {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.SubmitNetwork(simulation.Command{Tick: tick, Player: "alice", Authority: simulation.AuthorityPlayer,
		Sequence: sequence, Kind: kind, Payload: encoded}); err != nil {
		t.Fatal(err)
	}
}

func TestGroundItemTargetRefusesReassignmentAfterLateOwnerDrop(t *testing.T) {
	authority, engine, session := startGroundItemFixture(t, nil, "first", "inventory")
	t.Cleanup(func() { _ = authority.Stop(t.Context()); _ = session.Close(); _ = engine.Close() })
	admitGroundItemPlayer(t, session, "alice", 1, 1)
	admitGroundItemPlayer(t, session, "aaron", 2, 4)
	callGroundIdentityFixture(t, authority, "create_late_drop_item", "second")
	tick := engine.Tick() + 1
	submitPartyCommand(t, session, tick, "alice", 1, "item.move", map[string]any{
		"item_id": "first", "destination": map[string]any{"container": "world", "x": 1, "y": 1},
	})
	stepSession(t, session, 1)
	first := groundItemEntity(t, engine, "first")
	published := componentSnapshot(t, engine, first, "d2legacy.world.selectable")["id"].(string)
	checkedGroundItemView(t, session, "alice")
	payload, err := json.Marshal(map[string]any{
		"item_id": "second", "destination": map[string]any{"container": "world", "x": 4, "y": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := session.SubmitNetwork(simulation.Command{Tick: tick, Player: "aaron", Authority: simulation.AuthorityPlayer,
		Sequence: 1, Kind: "item.move", Payload: payload}); err != nil {
		t.Fatal(err)
	}
	first = groundItemEntity(t, engine, "first")
	second := groundItemEntity(t, engine, "second")
	checkedGroundItemView(t, session, "alice")
	checkedGroundItemView(t, session, "aaron")
	firstBefore := groundItemResidencySnapshot(t, engine, first)
	secondBefore := groundItemResidencySnapshot(t, engine, second)
	if firstBefore["d2legacy.world.selectable"]["id"] == published {
		t.Fatal("fixture did not reorder the first allocation")
	}
	if secondBefore["d2legacy.world.selectable"]["id"] == published {
		t.Fatal("published first target was reassigned to another owner's item")
	}
	submitPartyCommand(t, session, engine.Tick()+1, "alice", 2, "item.pickup", map[string]any{"target_id": published})
	stepSession(t, session, 1)
	if !reflect.DeepEqual(firstBefore, groundItemResidencySnapshot(t, engine, first)) ||
		!reflect.DeepEqual(secondBefore, groundItemResidencySnapshot(t, engine, second)) {
		t.Fatal("pre-replay target mutated an item after the successful late replay")
	}
	checkedGroundItemView(t, session, "alice")
	submitPartyCommand(t, session, engine.Tick()+1, "alice", 3, "item.pickup", map[string]any{
		"target_id": firstBefore["d2legacy.world.selectable"]["id"],
	})
	stepSession(t, session, 1)
	if componentSnapshot(t, engine, first, "d2legacy.item.placement")["container"] != "inventory" {
		t.Fatal("refreshed target did not collect the intended item")
	}
	checkedGroundItemView(t, session, "alice")
}

func TestGroundLootTargetBindsSourceAcrossReallocatedMonster(t *testing.T) {
	authority, engine, session := startGroundItemFixture(t, nil, "bootstrap", "inventory")
	t.Cleanup(func() { _ = authority.Stop(t.Context()); _ = session.Close(); _ = engine.Close() })
	admitGroundItemPlayer(t, session, "alice", 1, 1)
	tick := engine.Tick() + 1
	spawn := func(player, source string) simulation.Command {
		payload, err := json.Marshal(map[string]any{
			"spawn_id": source, "seed": 1, "x": 1, "y": 1, "act": 1, "level_id": 2,
			"definition": map[string]any{
				"id": "fallen", "token": "FA", "weapon_class": "HTH", "name_key": "Fallen", "ai": "fallen",
				"level": 1, "life_min": 0, "life_max": 0, "defense": 0, "attack_rating": 0,
				"physical_min": 0, "physical_max": 0, "experience": 0, "treasure_class": "fallen-drop",
				"think_interval": 10, "aggro_radius": 0, "attack_range": 0, "velocity": 0,
				"collider_radius": 0.4, "select_radius": 0.5,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		return simulation.Command{Tick: tick, Player: player, Authority: simulation.AuthoritySystem,
			Sequence: 1, Kind: "system.monster.spawn", Payload: payload}
	}
	if err := session.Submit(spawn("zzz", "published-source")); err != nil {
		t.Fatal(err)
	}
	stepSession(t, session, 1)
	var durableID string
	for _, entity := range dynamicStore(t, engine, "d2legacy.item.generation").Entities() {
		id := componentSnapshot(t, engine, entity, "d2legacy.item.identity")["id"].(string)
		if strings.HasPrefix(id, "loot:") {
			durableID = id
		}
	}
	if durableID == "" {
		t.Fatal("production death did not materialize the fixture roll")
	}
	item := groundItemEntity(t, engine, durableID)
	before := groundItemResidencySnapshot(t, engine, item)
	published := before["d2legacy.world.selectable"]["id"].(string)
	checkedGroundItemView(t, session, "alice")
	if _, err := session.SubmitNetwork(spawn("aaa", "earlier-source")); err != nil {
		t.Fatal(err)
	}
	item = groundItemEntity(t, engine, durableID)
	after := groundItemResidencySnapshot(t, engine, item)
	if !reflect.DeepEqual(before["d2legacy.item.generation"], after["d2legacy.item.generation"]) {
		t.Fatal("fixture must reuse the same durable ID and rolled facts to distinguish source binding")
	}
	if reflect.DeepEqual(before["d2legacy.item.origin"], after["d2legacy.item.origin"]) {
		t.Fatal("fixture did not reassign the old monster allocation to another source")
	}
	if after["d2legacy.world.selectable"]["id"] == published {
		t.Fatal("another monster's identical rolled item inherited the published target")
	}
	checkedGroundItemView(t, session, "alice")
	submitPartyCommand(t, session, engine.Tick()+1, "alice", 1, "item.pickup", map[string]any{"target_id": published})
	stepSession(t, session, 1)
	for _, viewItem := range checkedGroundItemView(t, session, "alice").Private.Items.Items {
		if strings.HasPrefix(viewItem.ID, "loot:") {
			t.Fatal("stale source target collected a generated item")
		}
	}
}

func TestGroundTargetBindingRollsBackOnFailureAndRefusesExhaustion(t *testing.T) {
	for _, exhausted := range []bool{false, true} {
		t.Run(fmt.Sprint(exhausted), func(t *testing.T) {
			authority, engine, session := startGroundItemFixture(t, nil, "probe", "inventory")
			t.Cleanup(func() { _ = authority.Stop(t.Context()); _ = session.Close(); _ = engine.Close() })
			admitGroundItemPlayer(t, session, "alice", 1, 1)
			last := uint64(0)
			sequence := uint64(1)
			if exhausted {
				last = 9007199254740991
				if err := authority.State.Replace("d2legacy.items.ground_targets", "d2legacy.items.ground_targets/v1",
					[]byte(fmt.Sprintf(`{"last":%d}`, last-1))); err != nil {
					t.Fatal(err)
				}
				submitPartyCommand(t, session, engine.Tick()+1, "alice", 1, "item.move", map[string]any{
					"item_id": "probe", "destination": map[string]any{"container": "world", "x": 1, "y": 1},
				})
				stepSession(t, session, 1)
				item := groundItemEntity(t, engine, "probe")
				selector := componentSnapshot(t, engine, item, "d2legacy.world.selectable")["id"].(string)
				if !strings.HasPrefix(selector, "item:9007199254740991:") || len(selector) != 86 {
					t.Fatal("final exact counter target violated the bounded full-digest format")
				}
				checkedGroundItemView(t, session, "alice")
				submitPartyCommand(t, session, engine.Tick()+1, "alice", 2, "item.pickup", map[string]any{"target_id": selector})
				stepSession(t, session, 1)
				sequence = 3
			}
			item := groundItemEntity(t, engine, "probe")
			before := groundItemResidencySnapshot(t, engine, item)
			tick := engine.Tick()
			containers := []string{"world", "invalid-container"}
			if exhausted {
				containers = containers[:1]
			}
			for index, container := range containers {
				submitPartyCommand(t, session, tick+1, "alice", sequence+uint64(index), "item.move", map[string]any{
					"item_id": "probe", "destination": map[string]any{"container": container, "x": 1, "y": 1},
				})
			}
			if err := session.Step(); err == nil {
				t.Fatal("invalid tick unexpectedly succeeded")
			}
			value, _ := authority.State.Read("d2legacy.items.ground_targets")
			var counter struct {
				Last uint64 `json:"last"`
			}
			if err := json.Unmarshal(value.Data, &counter); err != nil {
				t.Fatal(err)
			}
			if engine.Tick() != tick || counter.Last != last ||
				!reflect.DeepEqual(before, groundItemResidencySnapshot(t, engine, item)) {
				t.Fatal("failed tick retained target, item or counter mutation")
			}
			checkedGroundItemView(t, session, "alice")
		})
	}
}
