package clientapp

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/app/networkclock"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

func TestPresentationBufferInterpolatesBetweenKnownSnapshots(t *testing.T) {
	buffer := newPresentationBuffer()
	buffer.Push(worldView(10, worldEntity("peer", 0, 4)))
	buffer.Push(worldView(12, worldEntity("peer", 4, 8)))

	sample, found := buffer.Sample(networkclock.Moment{Tick: 11, Fraction: 0.5})
	if !found {
		t.Fatal("expected a presentation sample")
	}
	if sample.discreteTick != 10 || len(sample.entities) != 1 {
		t.Fatalf("sample = %+v", sample)
	}
	if got := sample.entities[0].Position; got.X != 3 || got.Y != 7 {
		t.Fatalf("interpolated position = %+v, want {3 7}", got)
	}
}

func TestPresentationBufferAppliesEntityLifecycleAtSnapshotBoundary(t *testing.T) {
	buffer := newPresentationBuffer()
	buffer.Push(worldView(10, worldEntity("departing", 0, 0)))
	buffer.Push(worldView(12, worldEntity("joining", 2, 2)))

	before, _ := buffer.Sample(networkclock.Moment{Tick: 11, Fraction: .99})
	if got := entityIDs(before.entities); len(got) != 1 || got[0] != "departing" {
		t.Fatalf("entities before boundary = %v", got)
	}
	at, _ := buffer.Sample(networkclock.Moment{Tick: 12})
	if got := entityIDs(at.entities); len(got) != 1 || got[0] != "joining" {
		t.Fatalf("entities at boundary = %v", got)
	}
}

func TestPresentationBufferUpsertsReliableMetadataAtNewestTransformTick(t *testing.T) {
	buffer := newPresentationBuffer()
	buffer.Push(worldView(10, worldEntity("old", 1, 2)))
	revision := buffer.snapshots[0].revision
	if !buffer.Upsert(worldView(10, worldEntity("new", 3, 4))) {
		t.Fatal("same-tick metadata upsert was rejected")
	}
	if len(buffer.snapshots) != 1 || buffer.snapshots[0].revision <= revision {
		t.Fatalf("upserted snapshots = %#v", buffer.snapshots)
	}
	sample, _ := buffer.Sample(networkclock.Moment{Tick: 10})
	if got := entityIDs(sample.entities); len(got) != 1 || got[0] != "new" {
		t.Fatalf("upserted entities = %v", got)
	}
}

func TestPresentationBufferBoundsExtrapolation(t *testing.T) {
	buffer := newPresentationBuffer()
	buffer.Push(worldView(10, worldEntity("peer", 0, 0)))
	buffer.Push(worldView(12, worldEntity("peer", 4, 0)))

	sample, _ := buffer.Sample(networkclock.Moment{Tick: 30})
	if !sample.extrapolated {
		t.Fatal("expected bounded extrapolation")
	}
	if got := sample.entities[0].Position.X; got != 9 {
		t.Fatalf("extrapolated x = %v, want 9", got)
	}
}

func TestPresentationBufferIsBoundedOrderedAndImmutable(t *testing.T) {
	buffer := newPresentationBuffer()
	health := int64(10)
	first := worldEntity("peer", 1, 2)
	first.Health = &health
	if !buffer.Push(worldView(1, first)) {
		t.Fatal("initial snapshot was rejected")
	}
	health = 99
	if buffer.Push(worldView(1, worldEntity("duplicate", 0, 0))) {
		t.Fatal("duplicate tick was accepted")
	}
	for tick := uint64(2); tick <= presentationSnapshotCapacity+4; tick++ {
		buffer.Push(worldView(tick, worldEntity("peer", float64(tick), 0)))
	}
	if len(buffer.snapshots) != presentationSnapshotCapacity {
		t.Fatalf("snapshot count = %d, want %d", len(buffer.snapshots), presentationSnapshotCapacity)
	}
	if buffer.snapshots[0].tick != 5 {
		t.Fatalf("oldest retained tick = %d, want 5", buffer.snapshots[0].tick)
	}

	immutable := newPresentationBuffer()
	immutableHealth := int64(42)
	immutableEntity := worldEntity("peer", 1, 2)
	immutableEntity.Health = &immutableHealth
	immutable.Push(worldView(1, immutableEntity))
	immutableHealth = 99
	sample, _ := immutable.Sample(networkclock.Moment{Tick: 1})
	if sample.entities[0].Health == nil || *sample.entities[0].Health != 42 {
		t.Fatalf("copied health = %v, want 42", sample.entities[0].Health)
	}
	*sample.entities[0].Health = 7
	again, _ := immutable.Sample(networkclock.Moment{Tick: 1})
	if *again.entities[0].Health != 42 {
		t.Fatalf("sample mutation leaked into history: %d", *again.entities[0].Health)
	}
}

func worldView(tick uint64, entities ...playeradapter.WorldEntity) playeradapter.WorldView {
	return playeradapter.WorldView{Version: playeradapter.WorldViewVersion, Tick: tick, Entities: entities}
}

func worldEntity(id string, x, y float64) playeradapter.WorldEntity {
	return playeradapter.WorldEntity{ID: id, Kind: "player", Position: playeradapter.HUDPosition{X: x, Y: y}}
}

func entityIDs(entities []playeradapter.WorldEntity) []string {
	result := make([]string, len(entities))
	for index := range entities {
		result[index] = entities[index].ID
	}
	return result
}
