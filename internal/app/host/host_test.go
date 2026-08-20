package host

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type componentFunc struct {
	start func(context.Context) error
	stop  func(context.Context) error
}

// Start delegates to the supplied hook, allowing tests to describe lifecycle behavior without defining a named type.
func (c componentFunc) Start(ctx context.Context) error {
	if c.start == nil {
		return nil
	}

	return c.start(ctx)
}

// Stop delegates to the supplied hook; an absent hook models a component that requires no shutdown work.
func (c componentFunc) Stop(ctx context.Context) error {
	if c.stop == nil {
		return nil
	}

	return c.stop(ctx)
}

// TestHostStartsDependenciesAndStopsInReverse protects the ordering contract that makes dependent teardown safe.
func TestHostStartsDependenciesAndStopsInReverse(t *testing.T) {
	t.Parallel()

	var calls []string

	h := New()

	registerRecordingComponent(t, h, "game", &calls, "scripts", "renderer")
	registerRecordingComponent(t, h, "renderer", &calls, "assets")
	registerRecordingComponent(t, h, "scripts", &calls, "assets")
	registerRecordingComponent(t, h, "assets", &calls)

	if err := h.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := h.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"start assets", "start scripts", "start renderer", "start game",
		"stop game", "stop renderer", "stop scripts", "stop assets",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

// TestHostRollsBackFailedStartAndCanRetry verifies that failed startup leaves no stale running state behind.
func TestHostRollsBackFailedStartAndCanRetry(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	fail := true

	var calls []string

	h := New()

	if err := h.Register(Definition{
		ID: "one",
		Component: componentFunc{
			start: func(context.Context) error {
				calls = append(calls, "start one")
				return nil
			},
			stop: func(context.Context) error {
				calls = append(calls, "stop one")
				return nil
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if err := h.Register(Definition{
		ID:        "two",
		DependsOn: []string{"one"},
		Component: componentFunc{
			start: func(context.Context) error {
				calls = append(calls, "start two")

				if fail {
					return boom
				}

				return nil
			},
			stop: func(context.Context) error {
				calls = append(calls, "stop two")
				return nil
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	// The first component must be stopped when its dependent fails, otherwise a retry would duplicate live state.
	if err := h.Start(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("Start error = %v, want %v", err, boom)
	}

	if got := h.Started(); len(got) != 0 {
		t.Fatalf("started after rollback = %v", got)
	}

	fail = false

	if err := h.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := h.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := []string{"start one", "start two", "stop one", "start one", "start two", "stop two", "stop one"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

// TestHostValidatesDependencyGraph confirms invalid graphs fail before any component startup can cause side effects.
func TestHostValidatesDependencyGraph(t *testing.T) {
	t.Parallel()

	t.Run("missing", func(t *testing.T) {
		h := New()
		if err := h.Register(Definition{
			ID:        "game",
			DependsOn: []string{"renderer"},
			Component: componentFunc{},
		}); err != nil {
			t.Fatal(err)
		}

		err := h.Start(context.Background())
		if err == nil || !strings.Contains(err.Error(), "unregistered component") {
			t.Fatalf("Start error = %v", err)
		}
	})

	t.Run("cycle", func(t *testing.T) {
		h := New()
		for _, def := range []Definition{
			{ID: "a", DependsOn: []string{"b"}, Component: componentFunc{}},
			{ID: "b", DependsOn: []string{"a"}, Component: componentFunc{}},
		} {
			if err := h.Register(def); err != nil {
				t.Fatal(err)
			}
		}

		err := h.Start(context.Background())
		if err == nil || !strings.Contains(err.Error(), "dependency cycle") {
			t.Fatalf("Start error = %v", err)
		}
	})
}

// TestHostHonorsCancellationAndJoinsStopErrors verifies shutdown attempts every component despite individual failures.
func TestHostHonorsCancellationAndJoinsStopErrors(t *testing.T) {
	t.Parallel()

	firstErr := errors.New("first stop")
	secondErr := errors.New("second stop")

	h := New()
	for _, def := range []Definition{
		{
			ID: "one",
			Component: componentFunc{
				stop: func(context.Context) error { return firstErr },
			},
		},
		{
			ID:        "two",
			DependsOn: []string{"one"},
			Component: componentFunc{
				stop: func(context.Context) error { return secondErr },
			},
		},
	} {
		if err := h.Register(def); err != nil {
			t.Fatal(err)
		}
	}

	if err := h.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Joined errors let callers identify every failed cleanup operation after the host has attempted them all.
	err := h.Stop(context.Background())
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatalf("Stop error = %v", err)
	}

	if err := h.Stop(context.Background()); err != nil {
		t.Fatalf("second Stop error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := h.Start(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Start error = %v", err)
	}
}

// TestHostRejectsInvalidRegistration protects the host's identity and lifecycle invariants at its public boundary.
func TestHostRejectsInvalidRegistration(t *testing.T) {
	t.Parallel()

	h := New()
	if err := h.Register(Definition{}); err == nil {
		t.Fatal("expected empty ID to fail")
	}

	if err := h.Register(Definition{ID: "nil"}); err == nil {
		t.Fatal("expected nil component to fail")
	}

	if err := h.Register(Definition{ID: "valid", Component: componentFunc{}}); err != nil {
		t.Fatal(err)
	}

	if err := h.Register(Definition{ID: "valid", Component: componentFunc{}}); err == nil {
		t.Fatal("expected duplicate ID to fail")
	}

	if err := h.Start(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := h.Register(Definition{ID: "late", Component: componentFunc{}}); err == nil {
		t.Fatal("expected registration while running to fail")
	}
}

// registerRecordingComponent adds a component whose hooks expose exact lifecycle order through calls.
func registerRecordingComponent(
	t *testing.T,
	h *Host,
	id string,
	calls *[]string,
	dependencies ...string,
) {
	t.Helper()

	err := h.Register(Definition{
		ID:        id,
		DependsOn: dependencies,
		Component: componentFunc{
			start: func(context.Context) error {
				*calls = append(*calls, "start "+id)
				return nil
			},
			stop: func(context.Context) error {
				*calls = append(*calls, "stop "+id)
				return nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
