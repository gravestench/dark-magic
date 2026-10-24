package world

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

type countingVelocityPaths struct {
	pathCalls int
	stepCalls int
	allowed   bool
}

// FindPath records expensive route requests so the test can detect unnecessary fixed-tick allocations.
func (paths *countingVelocityPaths) FindPath(request PathRequest) ([]Point, error) {
	paths.pathCalls++
	return []Point{request.Start, request.Goal}, nil
}

// WalkableStep records local collision validation and returns the fixture's configured decision.
func (paths *countingVelocityPaths) WalkableStep(_, _ Point, _ float64) bool {
	paths.stepCalls++
	return paths.allowed
}

// TestVelocityMovementSkipsPathSearchInsideCollisionCell protects the fast path and adjacent-step validator selection.
func TestVelocityMovementSkipsPathSearchInsideCollisionCell(t *testing.T) {
	engine := gameecs.NewWithClock(40*time.Millisecond, 1)
	// The engine owns subscriptions and world resources even when an assertion stops the test early.
	defer func() { _ = engine.Close() }()

	paths := &countingVelocityPaths{allowed: true}

	components := VelocityComponents{
		Position: "test.position", Velocity: "test.velocity", Collider: "test.collider",
	}
	if err := RegisterVelocityMovement(engine, paths, components); err != nil {
		t.Fatal(err)
	}

	position, _ := akara.GetDynamicStore(engine.World(), components.Position)
	velocity, _ := akara.GetDynamicStore(engine.World(), components.Velocity)
	collider, _ := akara.GetDynamicStore(engine.World(), components.Collider)
	marker, _ := akara.GetDynamicStore(engine.World(), "engine.world.velocity_mover")

	entity := engine.World().MustCreateEntity()
	if _, err := position.Set(entity, map[string]any{"x": 10.0, "y": 10.0}); err != nil {
		t.Fatal(err)
	}

	if _, err := velocity.Set(entity, map[string]any{"x": 1.0, "y": 0.0}); err != nil {
		t.Fatal(err)
	}

	if _, err := collider.Set(entity, map[string]any{"radius": 0.5}); err != nil {
		t.Fatal(err)
	}

	if _, err := marker.Set(entity, map[string]any{}); err != nil {
		t.Fatal(err)
	}

	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if paths.pathCalls != 0 || paths.stepCalls != 0 {
		t.Fatalf("same-cell velocity performed path/step checks %d/%d, want 0/0", paths.pathCalls, paths.stepCalls)
	}

	if _, err := velocity.Set(entity, map[string]any{"x": 20.0, "y": 0.0}); err != nil {
		t.Fatal(err)
	}

	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if paths.pathCalls != 0 || paths.stepCalls != 1 {
		t.Fatalf("adjacent velocity performed path/step checks %d/%d, want 0/1", paths.pathCalls, paths.stepCalls)
	}
}

// velocityRestoreFixture registers the native mover before snapshots so restore retains its system definition.
func velocityRestoreFixture(t *testing.T) (*gameecs.Engine, VelocityComponents) {
	t.Helper()

	engine := gameecs.NewWithClock(40*time.Millisecond, 1)
	components := VelocityComponents{
		Position: "test.restore.position",
		Velocity: "test.restore.velocity",
		Collider: "test.restore.collider",
	}
	if err := RegisterVelocityMovement(engine, nil, components); err != nil {
		t.Fatal(err)
	}

	return engine, components
}

// addVelocityRestoreMover gives one entity every schema required by the native movement filter.
func addVelocityRestoreMover(t *testing.T, engine *gameecs.Engine, components VelocityComponents) akara.Entity {
	t.Helper()

	entity := engine.World().MustCreateEntity()
	values := []struct {
		name   string
		fields map[string]any
	}{
		{components.Position, map[string]any{"x": 10.0, "y": 10.0}},
		{components.Velocity, map[string]any{"x": 1.0, "y": 0.0}},
		{components.Collider, map[string]any{"radius": 0.5}},
		{"engine.world.velocity_mover", map[string]any{}},
	}
	for _, value := range values {
		store, found := akara.GetDynamicStore(engine.World(), value.name)
		if !found {
			t.Fatalf("missing fixture store %q", value.name)
		}

		if _, err := store.Set(entity, value.fields); err != nil {
			t.Fatal(err)
		}
	}

	return entity
}

// assertVelocityRestoreX reads the engine's current world so writes to stale pre-restore stores cannot pass.
func assertVelocityRestoreX(t *testing.T, engine *gameecs.Engine, name string, entity akara.Entity, want float64) {
	t.Helper()

	store, found := akara.GetDynamicStore(engine.World(), name)
	if !found {
		t.Fatalf("missing position store %q", name)
	}

	component, found := store.Get(entity)
	if !found {
		t.Fatalf("missing restored position for entity %d", entity)
	}

	value, err := component.Get("x")
	if err != nil {
		t.Fatal(err)
	}

	x := value.(float64)
	if math.Abs(x-want) > 1e-9 {
		t.Fatalf("restored position x = %.12f, want %.12f", x, want)
	}
}

// TestVelocityMovementExistingMoverAdvancesAfterRestore catches stale writes through two world replacements.
func TestVelocityMovementExistingMoverAdvancesAfterRestore(t *testing.T) {
	engine, components := velocityRestoreFixture(t)
	defer func() { _ = engine.Close() }()

	entity := addVelocityRestoreMover(t, engine, components)
	for _, want := range []float64{10.04, 10.08} {
		snapshot, err := engine.Snapshot()
		if err != nil {
			t.Fatal(err)
		}

		if err := engine.Restore(snapshot); err != nil {
			t.Fatal(err)
		}

		if err := engine.Update(40 * time.Millisecond); err != nil {
			t.Fatal(err)
		}

		assertVelocityRestoreX(t, engine, components.Position, entity, want)
	}
}

// TestVelocityMovementNewMoverAdvancesAfterRestore catches an old store's absent entity read after restore.
func TestVelocityMovementNewMoverAdvancesAfterRestore(t *testing.T) {
	engine, components := velocityRestoreFixture(t)
	defer func() { _ = engine.Close() }()

	snapshot, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.Restore(snapshot); err != nil {
		t.Fatal(err)
	}

	entity := addVelocityRestoreMover(t, engine, components)
	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	assertVelocityRestoreX(t, engine, components.Position, entity, 10.04)
}

// TestVelocityMovementMissingPositionReturnsError exercises the guard directly because the query excludes it.
func TestVelocityMovementMissingPositionReturnsError(t *testing.T) {
	engine, components := velocityRestoreFixture(t)
	defer func() { _ = engine.Close() }()

	entity := engine.World().MustCreateEntity()
	position, _ := akara.GetDynamicStore(engine.World(), components.Position)
	velocity, _ := akara.GetDynamicStore(engine.World(), components.Velocity)
	collider, _ := akara.GetDynamicStore(engine.World(), components.Collider)

	stores := velocityStores{position: position, velocity: velocity, collider: collider}
	err := moveVelocityEntity(nil, stores, entity, 0.04)
	if err == nil || !strings.Contains(err.Error(), "position") {
		t.Fatalf("missing position error = %v", err)
	}
}

// TestVelocityMovementContinuesAfterSessionRollback makes a late network command replace the authoritative world.
func TestVelocityMovementContinuesAfterSessionRollback(t *testing.T) {
	engine, components := velocityRestoreFixture(t)
	entity := addVelocityRestoreMover(t, engine, components)

	session, err := gamesession.New(engine, gamesession.Config{
		Step: 40 * time.Millisecond, CheckpointInterval: 1, RollbackWindow: 8,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()

	if err := session.Register("test.noop", gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply:    func(*gameecs.Engine, simulation.Command) error { return nil },
	}); err != nil {
		t.Fatal(err)
	}

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	beforeRollback := engine.World()
	if _, err := session.SubmitNetwork(simulation.Command{
		Tick: 1, Player: "player", Sequence: 1, Kind: "test.noop", Payload: json.RawMessage(`{}`),
	}); err != nil {
		t.Fatal(err)
	}

	if engine.World() == beforeRollback {
		t.Fatal("late network input did not replace the ECS world")
	}

	if err := session.Step(); err != nil {
		t.Fatal(err)
	}

	assertVelocityRestoreX(t, engine, components.Position, entity, 10.12)
}
