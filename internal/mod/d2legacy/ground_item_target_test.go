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
