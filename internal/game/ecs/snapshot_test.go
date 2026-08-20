package ecs

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
)

// canonicalSnapshotFixture retains the mutable position component so tests can prove checksums respond to state
// changes after first establishing that independently built worlds serialize identically.
type canonicalSnapshotFixture struct {
	engine   *Engine
	position *akara.DynamicComponent
}

// TestSnapshotIsCanonicalAndDetectsStateChanges verifies equal worlds produce equal, schema-sorted checksums while a
// component mutation changes the checksum. Replay divergence detection depends on both guarantees.
func TestSnapshotIsCanonicalAndDetectsStateChanges(t *testing.T) {
	first := newCanonicalSnapshotFixture(t)
	second := newCanonicalSnapshotFixture(t)

	if err := first.engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if err := second.engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	left, err := first.engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	right, err := second.engine.Snapshot()
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

	if err := first.position.Set("label", "moved"); err != nil {
		t.Fatal(err)
	}

	changed, err := first.engine.Snapshot()
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

// TestSnapshotRestoresEntityIdentityTickAndAllocator verifies a JSON round trip preserves canonical state and allocator
// progress. Restoring only live entities must not allow a destroyed identity to be reused.
func TestSnapshotRestoresEntityIdentityTickAndAllocator(t *testing.T) {
	engine := New()
	closeTestEngine(t, engine)

	position := registerTestSchema(t, engine, akara.Schema{
		Name:    "world.position",
		Version: 1,
		Fields:  []akara.Field{{Name: "x", Kind: akara.FieldFloat64}},
	})

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

	closeTestEngine(t, restored)

	again, err := restored.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	want, err := snapshot.Checksum()
	if err != nil {
		t.Fatal(err)
	}

	got, err := again.Checksum()
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Fatalf("restored checksum = %s, want %s", got, want)
	}

	if next := restored.World().MustCreateEntity(); next != hero+1 {
		t.Fatalf("next entity = %d, want %d", next, hero+1)
	}
}

// TestRestorePreservesRegisteredSystemQueries confirms Restore replaces world-bound component handles and
// subscriptions before the next update, allowing an existing system to observe restored entities.
func TestRestorePreservesRegisteredSystemQueries(t *testing.T) {
	engine := New()
	closeTestEngine(t, engine)

	store := registerTestSchema(t, engine, akara.Schema{Name: "example.marker", Version: 1})

	entity := engine.World().MustCreateEntity()
	if _, err := store.Set(entity, map[string]any{}); err != nil {
		t.Fatal(err)
	}

	seen := 0

	registerTestSystem(t, engine, Definition{
		ID:    "observe",
		Phase: PhaseInput,
		All:   []akara.ComponentType{store},
		Update: func(_ Context, entities []akara.Entity, _ *StructuralCommands) error {
			seen = len(entities)

			return nil
		},
	})

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

// newCanonicalSnapshotFixture builds one world with deliberately reverse schema registration. The test relies on this
// ordering to prove that snapshots follow Akara's canonical store order rather than registration order.
func newCanonicalSnapshotFixture(t *testing.T) canonicalSnapshotFixture {
	t.Helper()

	engine := New()
	closeTestEngine(t, engine)

	velocity := registerTestSchema(t, engine, akara.Schema{
		Name:    "world.velocity",
		Version: 1,
		Fields:  []akara.Field{{Name: "x", Kind: akara.FieldFloat64}},
	})
	position := registerTestSchema(t, engine, akara.Schema{
		Name:    "world.position",
		Version: 2,
		Fields: []akara.Field{
			{Name: "label", Kind: akara.FieldString},
			{Name: "target", Kind: akara.FieldEntity},
		},
	})

	entity := engine.World().MustCreateEntity()
	if _, err := velocity.Set(entity, map[string]any{"x": 1.25}); err != nil {
		t.Fatal(err)
	}

	component, err := position.Set(entity, map[string]any{"label": "hero", "target": entity})
	if err != nil {
		t.Fatal(err)
	}

	return canonicalSnapshotFixture{engine: engine, position: component}
}

// registerTestSchema centralizes fatal schema setup while returning the world-bound store needed by each scenario.
func registerTestSchema(t *testing.T, engine *Engine, schema akara.Schema) *akara.DynamicStore {
	t.Helper()

	store, err := akara.RegisterSchema(engine.World(), schema)
	if err != nil {
		t.Fatal(err)
	}

	return store
}
