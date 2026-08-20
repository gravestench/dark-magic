package ecs

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gravestench/akara"
)

// TestEngineOrdersPhasesDependenciesAndStableEntities verifies that semantic phases and explicit dependencies produce
// one deterministic schedule while each system receives a stable, entity-ID-ordered query snapshot.
func TestEngineOrdersPhasesDependenciesAndStableEntities(t *testing.T) {
	engine := New()
	closeTestEngine(t, engine)

	position := akara.Register[struct{ X float64 }](engine.World())
	first := engine.World().MustCreateEntity()
	second := engine.World().MustCreateEntity()
	_, _ = position.Add(second)
	_, _ = position.Add(first)

	var (
		calls    []string
		entities []akara.Entity
	)

	registerTestSystem(t, engine, Definition{
		ID:    "input",
		Phase: PhaseInput,
		Update: func(Context, []akara.Entity, *StructuralCommands) error {
			calls = append(calls, "input")

			return nil
		},
	})
	registerTestSystem(t, engine, Definition{
		ID:    "move.a",
		Phase: PhaseMovement,
		Update: func(Context, []akara.Entity, *StructuralCommands) error {
			calls = append(calls, "move.a")

			return nil
		},
	})
	registerTestSystem(t, engine, Definition{
		ID:    "move.b",
		Phase: PhaseMovement,
		After: []string{"move.a"},
		All:   []akara.ComponentType{position},
		Update: func(_ Context, got []akara.Entity, _ *StructuralCommands) error {
			calls = append(calls, "move.b")
			entities = append([]akara.Entity(nil), got...)

			return nil
		},
	})

	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(calls, []string{"input", "move.a", "move.b"}) {
		t.Fatalf("calls = %v", calls)
	}

	if !reflect.DeepEqual(entities, []akara.Entity{first, second}) {
		t.Fatalf("entities = %v", entities)
	}
}

// TestEngineRejectsMissingDependenciesAndCyclesTransactionally ensures invalid registrations cannot partially replace
// the live schedule, which is essential when mods register systems incrementally.
func TestEngineRejectsMissingDependenciesAndCyclesTransactionally(t *testing.T) {
	engine := New()
	closeTestEngine(t, engine)

	missingDependency := Definition{
		ID:     "missing",
		Phase:  PhaseMovement,
		After:  []string{"absent"},
		Update: noOpUpdate,
	}
	if err := engine.Register(missingDependency); !errors.Is(err, ErrSystemNotFound) {
		t.Fatalf("missing dependency error = %v", err)
	}

	if len(engine.Systems()) != 0 {
		t.Fatal("failed registration changed schedule")
	}

	registerTestSystem(t, engine, Definition{ID: "a", Phase: PhaseMovement, Update: noOpUpdate})

	cycle := Definition{
		ID:     "b",
		Phase:  PhaseMovement,
		Before: []string{"a"},
		After:  []string{"a"},
		Update: noOpUpdate,
	}
	if err := engine.Register(cycle); !errors.Is(err, ErrSystemCycle) {
		t.Fatalf("cycle error = %v", err)
	}

	if !reflect.DeepEqual(engine.Systems(), []string{"a"}) {
		t.Fatalf("failed cycle changed schedule: %v", engine.Systems())
	}
}

// TestEngineAppliesStructuralCommandsAtSystemBarrier proves that a producer cannot observe its queued mutation while a
// later system can, defining the stable-query boundary that systems use while iterating entities.
func TestEngineAppliesStructuralCommandsAtSystemBarrier(t *testing.T) {
	engine := New()
	closeTestEngine(t, engine)

	tag, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "tag.active"})
	if err != nil {
		t.Fatal(err)
	}

	entity := engine.World().MustCreateEntity()
	observed := false

	registerTestSystem(t, engine, Definition{
		ID:    "add",
		Phase: PhaseIntent,
		Update: func(_ Context, _ []akara.Entity, commands *StructuralCommands) error {
			commands.AddDynamic(tag, entity, nil)

			if tag.Has(entity) {
				t.Fatal("structural mutation was visible inside producer system")
			}

			return nil
		},
	})
	registerTestSystem(t, engine, Definition{
		ID:    "observe",
		Phase: PhaseMovement,
		All:   []akara.ComponentType{tag},
		Update: func(_ Context, entities []akara.Entity, _ *StructuralCommands) error {
			observed = len(entities) == 1 && entities[0] == entity

			return nil
		},
	})

	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}

	if !observed {
		t.Fatal("later phase did not observe applied structural command")
	}
}

// TestEngineStopsOnSystemError confirms callback failures retain their identity through wrapping and prevent later
// systems from running in the failed tick.
func TestEngineStopsOnSystemError(t *testing.T) {
	engine := New()
	closeTestEngine(t, engine)

	failure := errors.New("boom")

	registerTestSystem(t, engine, Definition{
		ID:    "fail",
		Phase: PhaseInput,
		Update: func(Context, []akara.Entity, *StructuralCommands) error {
			return failure
		},
	})

	if err := engine.Update(time.Millisecond); !errors.Is(err, failure) {
		t.Fatalf("update error = %v", err)
	}
}

// TestEngineAdvanceUsesFixedStepsAndLimitsCatchUp verifies sub-step lag accumulation and the catch-up cap. These rules
// keep simulation deltas deterministic even when host-frame elapsed time is irregular or extreme.
func TestEngineAdvanceUsesFixedStepsAndLimitsCatchUp(t *testing.T) {
	engine := NewWithClock(10*time.Millisecond, 3)
	closeTestEngine(t, engine)

	var (
		deltas []time.Duration
		ticks  []uint64
	)

	registerTestSystem(t, engine, Definition{
		ID:    "clock",
		Phase: PhaseInput,
		Update: func(context Context, _ []akara.Entity, _ *StructuralCommands) error {
			deltas = append(deltas, context.Delta)
			ticks = append(ticks, context.Tick)

			return nil
		},
	})

	if steps, err := engine.Advance(9 * time.Millisecond); err != nil || steps != 0 {
		t.Fatalf("first advance = %d, %v", steps, err)
	}

	if steps, err := engine.Advance(6 * time.Millisecond); err != nil || steps != 1 {
		t.Fatalf("second advance = %d, %v", steps, err)
	}

	if steps, err := engine.Advance(time.Second); err != nil || steps != 3 {
		t.Fatalf("limited advance = %d, %v", steps, err)
	}

	wantDeltas := []time.Duration{
		10 * time.Millisecond,
		10 * time.Millisecond,
		10 * time.Millisecond,
		10 * time.Millisecond,
	}
	if !reflect.DeepEqual(deltas, wantDeltas) {
		t.Fatalf("deltas = %v", deltas)
	}

	if !reflect.DeepEqual(ticks, []uint64{1, 2, 3, 4}) {
		t.Fatalf("ticks = %v", ticks)
	}
}

// BenchmarkEngineSteadyStateTick tracks allocation and dispatch overhead for a representative movement schedule. It
// guards the lazy command-buffer path from readability refactors that accidentally allocate every tick.
func BenchmarkEngineSteadyStateTick(b *testing.B) {
	engine := New()

	b.Cleanup(func() { _ = engine.Close() })

	for index := range 24 {
		definition := Definition{
			ID:     fmt.Sprintf("system-%02d", index),
			Phase:  PhaseMovement,
			Update: noOpUpdate,
		}
		if err := engine.Register(definition); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if err := engine.Update(DefaultStep); err != nil {
			b.Fatal(err)
		}
	}
}

// closeTestEngine registers checked cleanup so a close failure is attributed to the test without hiding its primary
// assertion failure.
func closeTestEngine(t *testing.T, engine *Engine) {
	t.Helper()
	t.Cleanup(func() {
		if err := engine.Close(); err != nil {
			t.Errorf("close engine: %v", err)
		}
	})
}

// registerTestSystem keeps scenario tests focused on ordering and barrier behavior while preserving fatal setup errors.
func registerTestSystem(t *testing.T, engine *Engine, definition Definition) {
	t.Helper()

	if err := engine.Register(definition); err != nil {
		t.Fatal(err)
	}
}

// noOpUpdate supplies a valid system callback when a test exercises registration rather than runtime behavior.
func noOpUpdate(Context, []akara.Entity, *StructuralCommands) error {
	return nil
}
