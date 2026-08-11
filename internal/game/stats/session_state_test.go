package stats_test

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/game/stats"
)

// Keep the replay boundary honest without making the stats package import the
// simulation package merely to name an interface it already satisfies.
var _ simulation.StateParticipant = (*stats.Authority)(nil)

func TestStatStateChangesCompositeCheckpointChecksum(t *testing.T) {
	engine := ecs.New()
	defer engine.Close()
	world, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	authority := stats.NewAuthority()
	before, err := authority.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	beforeChecksum, err := simulation.CompositeChecksum(world, []simulation.ParticipantState{{ID: authority.StateID(), Data: before}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = authority.Apply("hero", stats.Mutation{Replace: []stats.Source{{
		ID: "base", Kind: stats.SourceBase, Lifetime: stats.LifetimeDurable,
		Owner:   stats.OwnerRef{Kind: "player", ID: "hero"},
		Entries: []stats.Entry{{Key: stats.Key{ID: 1}, Value: 20}},
	}}})
	if err != nil {
		t.Fatal(err)
	}
	after, err := authority.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	afterChecksum, err := simulation.CompositeChecksum(world, []simulation.ParticipantState{{ID: authority.StateID(), Data: after}})
	if err != nil {
		t.Fatal(err)
	}
	if beforeChecksum == afterChecksum {
		t.Fatal("stat source mutation did not change the composite checkpoint checksum")
	}
}
