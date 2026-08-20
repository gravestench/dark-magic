package host

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestManagerEnablesDependenciesAndRejectsUnsafeDisable protects dependency ordering and teardown safety.
func TestManagerEnablesDependenciesAndRejectsUnsafeDisable(t *testing.T) {
	t.Parallel()

	var calls []string

	m := NewManager()
	registerManaged(t, m, "assets", nil, &calls)
	registerManaged(t, m, "scripts", []string{"assets"}, &calls)
	registerManaged(t, m, "game", []string{"scripts"}, &calls)

	if err := m.Enable(context.Background(), "game"); err != nil {
		t.Fatal(err)
	}

	err := m.Disable(context.Background(), "assets")
	if err == nil || !strings.Contains(err.Error(), "active dependents") {
		t.Fatalf("Disable error = %v", err)
	}

	// Cascading is explicit because it expands a local disable request to every active dependent.
	if err := m.DisableCascade(context.Background(), "assets"); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"start assets", "start scripts", "start game",
		"stop game", "stop scripts", "stop assets",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}

	for _, id := range []string{"assets", "scripts", "game"} {
		status, _ := m.Status(id)
		if status.Desired || status.State != StateDisabled {
			t.Fatalf("status %q = %#v", id, status)
		}
	}
}

// TestManagerFailedEnableRollsBackImplicitDependencies verifies failed requests do not leak support services.
func TestManagerFailedEnableRollsBackImplicitDependencies(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")

	var calls []string

	m := NewManager()
	registerManaged(t, m, "assets", nil, &calls)

	if err := m.Register(ManagedDefinition{
		ID:        "game",
		DependsOn: []string{"assets"},
		New: func(context.Context) (Component, error) {
			return componentFunc{start: func(context.Context) error { return boom }}, nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	err := m.Enable(context.Background(), "game")
	if !errors.Is(err, boom) {
		t.Fatalf("Enable error = %v", err)
	}

	// Assets was enabled only to satisfy game, so the failed request must restore its original disabled state.
	assets, _ := m.Status("assets")
	if assets.State != StateDisabled || assets.Desired {
		t.Fatalf("assets status = %#v", assets)
	}

	game, _ := m.Status("game")
	if game.State != StateFailed || !game.Desired {
		t.Fatalf("game status = %#v", game)
	}

	if !reflect.DeepEqual(calls, []string{"start assets", "stop assets"}) {
		t.Fatalf("calls = %#v", calls)
	}
}

// TestManagerPreservesExplicitDependencyAfterFailedEnable distinguishes user intent from temporary dependency work.
func TestManagerPreservesExplicitDependencyAfterFailedEnable(t *testing.T) {
	t.Parallel()

	m := NewManager()
	registerManaged(t, m, "assets", nil, nil)

	if err := m.Enable(context.Background(), "assets"); err != nil {
		t.Fatal(err)
	}

	if err := m.Register(ManagedDefinition{
		ID:        "broken",
		DependsOn: []string{"assets"},
		New:       func(context.Context) (Component, error) { return nil, errors.New("broken") },
	}); err != nil {
		t.Fatal(err)
	}

	if err := m.Enable(context.Background(), "broken"); err == nil {
		t.Fatal("expected enable failure")
	}

	// Assets remains live because its own desired bit predates the failed dependent request.
	status, _ := m.Status("assets")
	if status.State != StateEnabled || !status.Desired {
		t.Fatalf("assets status = %#v", status)
	}
}

// TestManagerRestartAndEvents verifies a restart is observable as one complete disable-enable transition.
func TestManagerRestartAndEvents(t *testing.T) {
	t.Parallel()

	var calls []string

	m := NewManager()
	registerManaged(t, m, "scripts", nil, &calls)

	events, cancel := m.Subscribe(8)
	defer cancel()

	if err := m.Enable(context.Background(), "scripts"); err != nil {
		t.Fatal(err)
	}

	if err := m.Restart(context.Background(), "scripts"); err != nil {
		t.Fatal(err)
	}

	wantCalls := []string{"start scripts", "stop scripts", "start scripts"}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", calls, wantCalls)
	}

	wantStates := []State{
		StateEnabling,
		StateEnabled,
		StateDisabling,
		StateDisabled,
		StateEnabling,
		StateEnabled,
	}
	for _, want := range wantStates {
		select {
		case event := <-events:
			if event.ID != "scripts" || event.State != want {
				t.Fatalf("event = %#v, want state %q", event, want)
			}
		default:
			t.Fatalf("missing event for state %q", want)
		}
	}
}

// TestManagerRejectsCyclesAndMissingDependencies ensures invalid graphs fail before component construction.
func TestManagerRejectsCyclesAndMissingDependencies(t *testing.T) {
	t.Parallel()

	m := NewManager()
	noop := func(context.Context) (Component, error) { return componentFunc{}, nil }

	if err := m.Register(ManagedDefinition{
		ID:        "a",
		DependsOn: []string{"b"},
		New:       noop,
	}); err != nil {
		t.Fatal(err)
	}

	if err := m.Enable(context.Background(), "a"); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Fatalf("Enable error = %v", err)
	}

	if err := m.Register(ManagedDefinition{
		ID:        "b",
		DependsOn: []string{"a"},
		New:       noop,
	}); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("Register error = %v", err)
	}
}

// TestManagerApplyDesired verifies reconciliation starts dependencies first and removes dependents first.
func TestManagerApplyDesired(t *testing.T) {
	t.Parallel()

	var calls []string

	m := NewManager()
	registerManaged(t, m, "assets", nil, &calls)
	registerManaged(t, m, "game", []string{"assets"}, &calls)

	if err := m.ApplyDesired(context.Background(), map[string]bool{"game": true}); err != nil {
		t.Fatal(err)
	}

	if err := m.ApplyDesired(context.Background(), map[string]bool{}); err != nil {
		t.Fatal(err)
	}

	want := []string{"start assets", "start game", "stop game", "stop assets"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

// TestManagerTransactionallyReplacesEnabledComponent protects state transfer and zero-downtime replacement order.
func TestManagerTransactionallyReplacesEnabledComponent(t *testing.T) {
	t.Parallel()

	var calls []string

	m := NewManager()
	old := &statefulComponent{name: "old", state: "saved", calls: &calls}

	registerManagedInstance(t, m, "scripts", old)

	if err := m.Enable(context.Background(), "scripts"); err != nil {
		t.Fatal(err)
	}

	newComponent := &statefulComponent{name: "new", calls: &calls}
	if err := m.Replace(context.Background(), managedInstanceDefinition("scripts", newComponent)); err != nil {
		t.Fatal(err)
	}

	// Import precedes startup so the replacement never serves traffic with empty state.
	if newComponent.state != "saved" {
		t.Fatalf("replacement state = %q", newComponent.state)
	}

	want := []string{"start old", "export old", "import new", "start new", "stop old"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

// TestManagerKeepsExistingInstanceWhenReplacementFails verifies replacement failure does not cause an outage.
func TestManagerKeepsExistingInstanceWhenReplacementFails(t *testing.T) {
	t.Parallel()

	var calls []string

	m := NewManager()
	old := &statefulComponent{name: "old", calls: &calls}
	registerManagedInstance(t, m, "scripts", old)

	if err := m.Enable(context.Background(), "scripts"); err != nil {
		t.Fatal(err)
	}

	broken := &statefulComponent{name: "broken", startErr: errors.New("compile failure"), calls: &calls}
	if err := m.Replace(context.Background(), managedInstanceDefinition("scripts", broken)); err == nil {
		t.Fatal("expected replacement failure")
	}

	// The old instance remains authoritative even though the candidate was constructed and partially started.
	status, _ := m.Status("scripts")
	if status.State != StateEnabled {
		t.Fatalf("status = %#v", status)
	}

	if err := m.Disable(context.Background(), "scripts"); err != nil {
		t.Fatal(err)
	}

	want := []string{"start old", "export old", "import broken", "start broken", "stop broken", "stop old"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

type statefulComponent struct {
	name     string
	state    string
	startErr error
	calls    *[]string
}

// Start records candidate ordering and optionally simulates a startup failure after construction.
func (c *statefulComponent) Start(context.Context) error {
	*c.calls = append(*c.calls, "start "+c.name)
	return c.startErr
}

// Stop records which instance the manager retired or cleaned up after a failed replacement.
func (c *statefulComponent) Stop(context.Context) error {
	*c.calls = append(*c.calls, "stop "+c.name)
	return nil
}

// ExportState records that the live instance supplied its state before the candidate started.
func (c *statefulComponent) ExportState(context.Context) (any, error) {
	*c.calls = append(*c.calls, "export "+c.name)
	return c.state, nil
}

// ImportState records that the candidate received state before becoming the authoritative instance.
func (c *statefulComponent) ImportState(_ context.Context, state any) error {
	*c.calls = append(*c.calls, "import "+c.name)
	c.state, _ = state.(string)

	return nil
}

// registerManaged installs a fresh recording component for dependency and lifecycle tests.
func registerManaged(t *testing.T, m *Manager, id string, dependencies []string, calls *[]string) {
	t.Helper()

	err := m.Register(ManagedDefinition{
		ID:        id,
		DependsOn: dependencies,
		New: func(context.Context) (Component, error) {
			return componentFunc{
				start: func(context.Context) error {
					if calls != nil {
						*calls = append(*calls, "start "+id)
					}

					return nil
				},
				stop: func(context.Context) error {
					if calls != nil {
						*calls = append(*calls, "stop "+id)
					}

					return nil
				},
			}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// registerManagedInstance installs one fixed instance so replacement tests can observe its exact identity.
func registerManagedInstance(t *testing.T, m *Manager, id string, component Component) {
	t.Helper()

	if err := m.Register(managedInstanceDefinition(id, component)); err != nil {
		t.Fatal(err)
	}
}

// managedInstanceDefinition wraps a fixed test instance in the manager's construction contract.
func managedInstanceDefinition(id string, component Component) ManagedDefinition {
	return ManagedDefinition{
		ID: id,
		New: func(context.Context) (Component, error) {
			return component, nil
		},
	}
}
