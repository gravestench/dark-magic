package ecs

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
)

func TestSnapshotIsCanonicalAndDetectsStateChanges(t *testing.T) {
	build := func() (*Engine, *akara.DynamicComponent) {
		engine := New()
		velocity, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "world.velocity", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}}})
		if err != nil {
			t.Fatal(err)
		}
		position, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "world.position", Version: 2, Fields: []akara.Field{{Name: "label", Kind: akara.FieldString}, {Name: "target", Kind: akara.FieldEntity}}})
		if err != nil {
			t.Fatal(err)
		}
		entity := engine.World().MustCreateEntity()
		if _, err := velocity.Set(entity, map[string]any{"x": 1.25}); err != nil {
			t.Fatal(err)
		}
		component, err := position.Set(entity, map[string]any{"label": "hero", "target": entity})
		if err != nil {
			t.Fatal(err)
		}
		return engine, component
	}
	first, component := build()
	defer first.Close()
	second, _ := build()
	defer second.Close()
	if err := first.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := second.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	left, err := first.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	right, err := second.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	leftChecksum, err := left.Checksum()
	if err != nil {
		t.Fatal(err)
	}
	rightChecksum, err := right.Checksum()
	if err != nil {
		t.Fatal(err)
	}
	if leftChecksum != rightChecksum {
		t.Fatalf("identical worlds differ: %s != %s", leftChecksum, rightChecksum)
	}
	if left.Components[0].Name != "world.position" || left.Components[1].Name != "world.velocity" {
		t.Fatalf("components not sorted: %#v", left.Components)
	}
	if err := component.Set("label", "moved"); err != nil {
		t.Fatal(err)
	}
	changed, err := first.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	changedChecksum, err := changed.Checksum()
	if err != nil {
		t.Fatal(err)
	}
	if changedChecksum == leftChecksum {
		t.Fatal("state mutation did not change checksum")
	}
}

func TestSnapshotRestoresEntityIdentityTickAndAllocator(t *testing.T) {
	engine := New()
	defer engine.Close()
	position, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}}})
	if err != nil {
		t.Fatal(err)
	}
	destroyed := engine.World().MustCreateEntity()
	hero := engine.World().MustCreateEntity()
	if !engine.World().DestroyEntity(destroyed) {
		t.Fatal("failed to create identity gap")
	}
	if _, err := position.Set(hero, map[string]any{"x": 3.5}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	snapshot, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := snapshot.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := UnmarshalSnapshot(encoded)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreSnapshot(decoded)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	again, err := restored.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	want, _ := snapshot.Checksum()
	got, _ := again.Checksum()
	if got != want {
		t.Fatalf("restored checksum = %s, want %s", got, want)
	}
	if next := restored.World().MustCreateEntity(); next != hero+1 {
		t.Fatalf("next entity = %d, want %d", next, hero+1)
	}
}

func TestRestorePreservesRegisteredSystemQueries(t *testing.T) {
	engine := New()
	defer engine.Close()
	store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "example.marker", Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	entity := engine.World().MustCreateEntity()
	if _, err := store.Set(entity, map[string]any{}); err != nil {
		t.Fatal(err)
	}
	seen := 0
	if err := engine.Register(Definition{ID: "observe", Phase: PhaseInput, All: []akara.ComponentType{store}, Update: func(_ Context, entities []akara.Entity, _ *akara.CommandBuffer) error {
		seen = len(entities)
		return nil
	}}); err != nil {
		t.Fatal(err)
	}
	before, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Restore(before); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(DefaultStep); err != nil {
		t.Fatal(err)
	}
	if seen != 1 {
		t.Fatalf("restored system query saw %d entities", seen)
	}
}
