package player

import (
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"testing"
)

func TestPrivateViewCannotExposeWorldSourceAsEmptyOwner(t *testing.T) {
	snapshot := gameecs.Snapshot{Version: gameecs.SnapshotVersion, Tick: 4, Components: []gameecs.ComponentSnapshot{
		stringsComponent("d2legacy.items.layout", []string{"owner"}, []any{uint64(10), ""}, []any{uint64(20), "alice"}),
		entityAndStringsComponent("d2legacy.item.identity", 11, 10, map[string]string{"id": "world-drop", "code": "ssd"}),
	}}
	for _, owner := range []string{"", " \t"} {
		view, err := ProjectPrivateView(owner, simulation.Checkpoint{Tick: 4, Snapshot: &snapshot})
		if err == nil || len(view.Items.Items) != 0 {
			t.Fatalf("empty principal received world source: %#v error=%v", view, err)
		}
	}
	view, err := ProjectPrivateView("alice", simulation.Checkpoint{Tick: 4, Snapshot: &snapshot})
	if err != nil || len(view.Items.Items) != 0 {
		t.Fatalf("real owner received world drop in private inventory: %#v error=%v", view, err)
	}
}
