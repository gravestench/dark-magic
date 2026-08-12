package ecs

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gravestench/akara"
)

func TestEngineOrdersPhasesDependenciesAndStableEntities(t *testing.T) {
	engine := New()
	defer engine.Close()
	position := akara.Register[struct{ X float64 }](engine.World())
	first := engine.World().MustCreateEntity()
	second := engine.World().MustCreateEntity()
	_, _ = position.Add(second)
	_, _ = position.Add(first)
	var calls []string
	var entities []akara.Entity
	register := func(definition Definition) {
		t.Helper()
		if err := engine.Register(definition); err != nil {
			t.Fatal(err)
		}
	}
	register(Definition{ID: "input", Phase: PhaseInput, Update: func(Context, []akara.Entity, *akara.CommandBuffer) error {
		calls = append(calls, "input")
		return nil
	}})
	register(Definition{ID: "move.a", Phase: PhaseMovement, Update: func(Context, []akara.Entity, *akara.CommandBuffer) error {
		calls = append(calls, "move.a")
		return nil
	}})
	register(Definition{ID: "move.b", Phase: PhaseMovement, After: []string{"move.a"}, All: []akara.ComponentType{position}, Update: func(_ Context, got []akara.Entity, _ *akara.CommandBuffer) error {
		calls = append(calls, "move.b")
		entities = append([]akara.Entity(nil), got...)
		return nil
	}})
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

func TestEngineRejectsMissingDependenciesAndCyclesTransactionally(t *testing.T) {
	engine := New()
	defer engine.Close()
	update := func(Context, []akara.Entity, *akara.CommandBuffer) error { return nil }
	if err := engine.Register(Definition{ID: "missing", Phase: PhaseMovement, After: []string{"absent"}, Update: update}); !errors.Is(err, ErrSystemNotFound) {
		t.Fatalf("missing dependency error = %v", err)
	}
	if len(engine.Systems()) != 0 {
		t.Fatal("failed registration changed schedule")
	}
	if err := engine.Register(Definition{ID: "a", Phase: PhaseMovement, Update: update}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Register(Definition{ID: "b", Phase: PhaseMovement, Before: []string{"a"}, After: []string{"a"}, Update: update}); !errors.Is(err, ErrSystemCycle) {
		t.Fatalf("cycle error = %v", err)
	}
	if !reflect.DeepEqual(engine.Systems(), []string{"a"}) {
		t.Fatalf("failed cycle changed schedule: %v", engine.Systems())
	}
}

func TestEngineAppliesStructuralCommandsAtSystemBarrier(t *testing.T) {
	engine := New()
	defer engine.Close()
	tag, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "tag.active"})
	if err != nil {
		t.Fatal(err)
	}
	entity := engine.World().MustCreateEntity()
	var observed bool
	if err := engine.Register(Definition{ID: "add", Phase: PhaseIntent, Update: func(_ Context, _ []akara.Entity, commands *akara.CommandBuffer) error {
		commands.AddDynamic(tag, entity, nil)
		if tag.Has(entity) {
			t.Fatal("structural mutation was visible inside producer system")
		}
		return nil
	}}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Register(Definition{ID: "observe", Phase: PhaseMovement, All: []akara.ComponentType{tag}, Update: func(_ Context, entities []akara.Entity, _ *akara.CommandBuffer) error {
		observed = len(entities) == 1 && entities[0] == entity
		return nil
	}}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	if !observed {
		t.Fatal("later phase did not observe applied structural command")
	}
}

func TestEngineStopsOnSystemError(t *testing.T) {
	engine := New()
	defer engine.Close()
	failure := errors.New("boom")
	if err := engine.Register(Definition{ID: "fail", Phase: PhaseInput, Update: func(Context, []akara.Entity, *akara.CommandBuffer) error { return failure }}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(time.Millisecond); !errors.Is(err, failure) {
		t.Fatalf("update error = %v", err)
	}
}

func TestEngineAdvanceUsesFixedStepsAndLimitsCatchUp(t *testing.T) {
	engine := NewWithClock(10*time.Millisecond, 3)
	defer engine.Close()
	var deltas []time.Duration
	var ticks []uint64
	if err := engine.Register(Definition{ID: "clock", Phase: PhaseInput, Update: func(context Context, _ []akara.Entity, _ *akara.CommandBuffer) error {
		deltas = append(deltas, context.Delta)
		ticks = append(ticks, context.Tick)
		return nil
	}}); err != nil {
		t.Fatal(err)
	}
	if steps, err := engine.Advance(9 * time.Millisecond); err != nil || steps != 0 {
		t.Fatalf("first advance = %d, %v", steps, err)
	}
	if steps, err := engine.Advance(6 * time.Millisecond); err != nil || steps != 1 {
		t.Fatalf("second advance = %d, %v", steps, err)
	}
	if steps, err := engine.Advance(time.Second); err != nil || steps != 3 {
		t.Fatalf("limited advance = %d, %v", steps, err)
	}
	if !reflect.DeepEqual(deltas, []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond}) {
		t.Fatalf("deltas = %v", deltas)
	}
	if !reflect.DeepEqual(ticks, []uint64{1, 2, 3, 4}) {
		t.Fatalf("ticks = %v", ticks)
	}
}

func BenchmarkEngineSteadyStateTick(b *testing.B) {
	engine := New()
	b.Cleanup(func() { _ = engine.Close() })
	update := func(Context, []akara.Entity, *akara.CommandBuffer) error { return nil }
	for index := range 24 {
		if err := engine.Register(Definition{ID: fmt.Sprintf("system-%02d", index), Phase: PhaseMovement, Update: update}); err != nil {
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
