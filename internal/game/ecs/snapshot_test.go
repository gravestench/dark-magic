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
