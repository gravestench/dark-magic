package world

import (
	"encoding/json"
	"fmt"
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

// levelMovementFixture uses real collision maps; selecting the wrong map changes the observed position.
func levelMovementFixture(
	t *testing.T,
	levels map[int64]VelocityPathFinder,
	field string,
	kind akara.FieldKind,
) (*gameecs.Engine, VelocityComponents, *akara.DynamicStore) {
	t.Helper()

	engine := gameecs.NewWithClock(40*time.Millisecond, 1)

	t.Cleanup(func() { _ = engine.Close() })

	location, err := akara.RegisterSchema(
		engine.World(),
		akara.Schema{Name: "test.location", Version: 1, Fields: []akara.Field{{Name: field, Kind: kind}}},
	)
	if err != nil {
		t.Fatal(err)
	}

	components := VelocityComponents{Position: "test.position", Velocity: "test.velocity", Collider: "test.collider"}
	if err := RegisterVelocityMovementByLevel(
		engine,
		levels,
		components,
		VelocityLevelLocation{Component: "test.location", Field: "level"},
	); err != nil {
		t.Fatal(err)
	}

	return engine, components, location
}

// Resolve the current store on each write so restore tests never retain a stale world reference.
func setLevelMovementValues(
	t *testing.T,
	engine *gameecs.Engine,
	name string,
	entity akara.Entity,
	values map[string]any,
) {
	t.Helper()

	store, found := akara.GetDynamicStore(engine.World(), name)
	if !found {
		t.Fatalf("missing store %s", name)
	}

	if _, err := store.Set(entity, values); err != nil {
		t.Fatal(err)
	}
}

// Create identical spatial state with optional location to isolate selection and missing-location failures.
func addLevelMover(
	t *testing.T,
	engine *gameecs.Engine,
	components VelocityComponents,
	level any,
	speed, radius float64,
) akara.Entity {
	t.Helper()
	entity := addVelocityRestoreMover(t, engine, components)
	setLevelMovementValues(t, engine, components.Position, entity, map[string]any{"x": 9.4, "y": 12.0})
	setLevelMovementValues(t, engine, components.Velocity, entity, map[string]any{"x": speed, "y": 0.0})
	setLevelMovementValues(t, engine, components.Collider, entity, map[string]any{"radius": radius})

	if level != nil {
		setLevelMovementValues(t, engine, "test.location", entity, map[string]any{"level": level})
	}

	return entity
}

// Check both axes so rejecting a step must consume the entire requested velocity.
func assertLevelVelocity(
	t *testing.T,
	engine *gameecs.Engine,
	components VelocityComponents,
	entity akara.Entity,
	want float64,
) {
	t.Helper()

	store, found := akara.GetDynamicStore(engine.World(), components.Velocity)
	if !found {
		t.Fatal("velocity store absent")
	}

	component, found := store.Get(entity)
	if !found {
		t.Fatalf("entity %d velocity absent", entity)
	}

	for _, field := range []string{"x", "y"} {
		value, err := component.Get(field)
		if err != nil {
			t.Fatal(err)
		}

		expected := 0.0
		if field == "x" {
			expected = want
		}

		if value != expected {
			t.Fatalf("velocity %s = %v, want %v", field, value, expected)
		}
	}
}

// Opposing maps and a subsequent level change catch fixed-map selection and retained location values.
func TestVelocityMovementByLevelUsesCurrentMoverLocation(t *testing.T) {
	for _, blockedLevel := range []int64{2, 3} {
		t.Run(fmt.Sprint(blockedLevel), func(t *testing.T) {
			open, blocked := openTestMap(40, 40), openTestMap(40, 40)

			for y := 10; y < 15; y++ {
				for x := 10; x < 15; x++ {
					blocked.flags[y*40+x].BlockWalk = true
				}
			}

			levels := map[int64]VelocityPathFinder{2: open, 3: open}
			levels[blockedLevel] = blocked
			engine, components, _ := levelMovementFixture(t, levels, "level", akara.FieldInt64)
			// Replacing the caller's lookup must not replace registered geometry.
			levels[blockedLevel] = open

			movers := map[int64]akara.Entity{}
			for _, level := range []int64{2, 3} {
				movers[level] = addLevelMover(t, engine, components, level, 25, 0.1)
				if level == blockedLevel {
					setLevelMovementValues(
						t,
						engine,
						components.Velocity,
						movers[level],
						map[string]any{"x": 25.0, "y": 5.0},
					)
				}
			}

			if err := engine.Update(40 * time.Millisecond); err != nil {
				t.Fatal(err)
			}

			for _, level := range []int64{2, 3} {
				want := 10.4
				if level == blockedLevel {
					want = 9.4

					assertLevelVelocity(t, engine, components, movers[level], 0)
				}

				assertVelocityRestoreX(t, engine, components.Position, movers[level], want)
			}

			entity := movers[blockedLevel]
			setLevelMovementValues(t, engine, "test.location", entity, map[string]any{"level": int64(5) - blockedLevel})
			setLevelMovementValues(t, engine, components.Velocity, entity, map[string]any{"x": 25.0, "y": 0.0})

			if err := engine.Update(40 * time.Millisecond); err != nil {
				t.Fatal(err)
			}

			assertVelocityRestoreX(t, engine, components.Position, entity, 10.4)
		})
	}
}

// Long steps must use the selected map and full collider footprint rather than point-only destination checks.
func TestVelocityMovementByLevelLongStepUsesSelectedMap(t *testing.T) {
	// At radius .6 the goal footprint includes (14,12); a point mover would incorrectly pass.
	for _, level := range []int64{2, 3} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			open, blocked := openTestMap(40, 40), openTestMap(40, 40)
			blocked.flags[12*40+14].BlockWalk = true
			engine, components, _ := levelMovementFixture(
				t,
				map[int64]VelocityPathFinder{2: open, 3: blocked},
				"level",
				akara.FieldInt64,
			)

			entity := addLevelMover(t, engine, components, level, 100, .6)
			if err := engine.Update(40 * time.Millisecond); err != nil {
				t.Fatal(err)
			}

			want := 13.4
			if level == 3 {
				want = 9.4

				assertLevelVelocity(t, engine, components, entity, 0)
			}

			assertVelocityRestoreX(t, engine, components.Position, entity, want)
		})
	}
}

// Invalid location or geometry must fail before the same-cell fast path can permit movement.
func TestVelocityMovementByLevelRejectsUnknownLocationBeforeSameCell(t *testing.T) {
	for _, test := range []struct {
		name, field string
		kind        akara.FieldKind
		level       any
		paths       VelocityPathFinder
		known       bool
	}{
		{"missing component", "level", akara.FieldInt64, nil, openTestMap(40, 40), true},
		{"missing field", "other", akara.FieldInt64, nil, openTestMap(40, 40), true},
		{"wrong type", "level", akara.FieldString, "2", openTestMap(40, 40), true},
		{"zero level", "level", akara.FieldInt64, int64(0), openTestMap(40, 40), true},
		{"negative level", "level", akara.FieldInt64, int64(-2), openTestMap(40, 40), true},
		{"absent map", "level", akara.FieldInt64, int64(2), nil, false},
		{"nil map", "level", akara.FieldInt64, int64(2), nil, true},
		{"typed nil map", "level", akara.FieldInt64, int64(2), (*Map)(nil), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			levels := map[int64]VelocityPathFinder{}
			if test.known {
				levels[2] = test.paths
			}

			engine, components, location := levelMovementFixture(t, levels, test.field, test.kind)

			entity := addLevelMover(t, engine, components, nil, 1, .1)
			if test.name == "missing field" {
				if _, err := location.Set(entity, map[string]any{"other": int64(2)}); err != nil {
					t.Fatal(err)
				}
			} else if test.level != nil {
				setLevelMovementValues(t, engine, "test.location", entity, map[string]any{"level": test.level})
			}

			if err := engine.Update(40 * time.Millisecond); err == nil {
				t.Fatal("invalid location/map allowed same-cell movement")
			}

			assertVelocityRestoreX(t, engine, components.Position, entity, 9.4)
			assertLevelVelocity(t, engine, components, entity, 1)
		})
	}
}

// Restore and late-input rollback replace stores; movement must read authoritative location from the new world.
func TestVelocityMovementByLevelUsesRestoredLocation(t *testing.T) {
	open, blocked := openTestMap(40, 40), openTestMap(40, 40)
	blocked.flags[12*40+10].BlockWalk = true
	engine, components, _ := levelMovementFixture(
		t,
		map[int64]VelocityPathFinder{2: blocked, 3: open},
		"level",
		akara.FieldInt64,
	)
	entity := addLevelMover(t, engine, components, int64(2), 25, .1)

	snapshot, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if err := engine.Restore(snapshot); err != nil {
		t.Fatal(err)
	}

	setLevelMovementValues(t, engine, "test.location", entity, map[string]any{"level": int64(3)})

	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	assertVelocityRestoreX(t, engine, components.Position, entity, 10.4)
	// Late admitted input replaces the world and changes authoritative location during replay.
	session, err := gamesession.New(
		engine,
		gamesession.Config{Step: 40 * time.Millisecond, CheckpointInterval: 1, RollbackWindow: 8},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = session.Close() }()

	if err := session.Register("test.level", gamesession.CommandHandler{
		Validate: func(simulation.Command) error { return nil },
		Apply: func(engine *gameecs.Engine, _ simulation.Command) error {
			setLevelMovementValues(t, engine, "test.location", entity, map[string]any{"level": int64(2)})
			setLevelMovementValues(t, engine, components.Position, entity, map[string]any{"x": 9.4, "y": 12.0})

			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}

	before := engine.World()

	command := simulation.Command{
		Tick: 2, Player: "player", Sequence: 1, Kind: "test.level", Payload: json.RawMessage(`{}`),
	}
	if _, err := session.SubmitNetwork(command); err != nil {
		t.Fatal(err)
	}

	if engine.World() == before {
		t.Fatal("late command did not replace world")
	}

	assertVelocityRestoreX(t, engine, components.Position, entity, 9.4)
	assertLevelVelocity(t, engine, components, entity, 0)
}
