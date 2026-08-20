package navigation

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type testScene struct {
	id       string
	blocks   bool
	enterErr error
	calls    *[]string
}

// Create records initialization so lifecycle ordering can be asserted precisely.
func (s *testScene) Create(context.Context) error {
	*s.calls = append(*s.calls, "create "+s.id)
	return nil
}

// Enter records activation and optionally injects the failure used by rollback tests.
func (s *testScene) Enter(context.Context) error {
	*s.calls = append(*s.calls, "enter "+s.id)
	return s.enterErr
}

// Update records simulation work so blockers can prove which stack suffix ran.
func (s *testScene) Update(context.Context, time.Duration) error {
	*s.calls = append(*s.calls, "update "+s.id)
	return nil
}

// Render records drawing order without coupling tests to a renderer.
func (s *testScene) Render(context.Context) error {
	*s.calls = append(*s.calls, "render "+s.id)
	return nil
}

// Exit records deactivation so cleanup order remains observable.
func (s *testScene) Exit(context.Context) error {
	*s.calls = append(*s.calls, "exit "+s.id)
	return nil
}

// Destroy records final cleanup independently from Exit.
func (s *testScene) Destroy(context.Context) error {
	*s.calls = append(*s.calls, "destroy "+s.id)
	return nil
}

// BlocksUpdateBelow exposes each fixture's modal-update policy.
func (s *testScene) BlocksUpdateBelow() bool { return s.blocks }

// TestNavigationStackUpdateRenderAndCleanup protects bottom-to-top rendering,
// blocker-aware updates, and top-to-bottom cleanup as one lifecycle contract.
func TestNavigationStackUpdateRenderAndCleanup(t *testing.T) {
	t.Parallel()

	var calls []string

	manager := New()
	registerScene(t, manager, "world", false, nil, &calls)
	registerScene(t, manager, "inventory", true, nil, &calls)
	registerScene(t, manager, "tooltip", false, nil, &calls)

	if err := manager.Replace(context.Background(), "world"); err != nil {
		t.Fatal(err)
	}

	if err := manager.Push(context.Background(), "inventory"); err != nil {
		t.Fatal(err)
	}

	if err := manager.Push(context.Background(), "tooltip"); err != nil {
		t.Fatal(err)
	}

	if err := manager.Update(context.Background(), time.Second/60); err != nil {
		t.Fatal(err)
	}

	if err := manager.Render(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := manager.Close(context.Background()); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"create world", "enter world", "create inventory", "enter inventory", "create tooltip", "enter tooltip",
		"update inventory", "update tooltip",
		"render world", "render inventory", "render tooltip",
		"exit tooltip", "destroy tooltip", "exit inventory", "destroy inventory", "exit world", "destroy world",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %#v, want %#v", calls, want)
	}
}

// TestReplacePreservesStackWhenNewSceneCannotEnter verifies replacement is
// transactional until the candidate completes both startup callbacks.
func TestReplacePreservesStackWhenNewSceneCannotEnter(t *testing.T) {
	t.Parallel()

	var calls []string

	manager := New()
	registerScene(t, manager, "menu", false, nil, &calls)
	registerScene(t, manager, "broken", false, errors.New("broken"), &calls)

	if err := manager.Replace(context.Background(), "menu"); err != nil {
		t.Fatal(err)
	}

	if err := manager.Replace(context.Background(), "broken"); err == nil {
		t.Fatal("expected replacement to fail")
	}

	if got := manager.Stack(); !reflect.DeepEqual(got, []string{"menu"}) {
		t.Fatalf("stack = %v", got)
	}

	want := []string{"create menu", "enter menu", "create broken", "enter broken", "destroy broken"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

// TestFocusedReportsTopScene confirms focus is absent for an empty stack and
// moves to the most recently pushed scene.
func TestFocusedReportsTopScene(t *testing.T) {
	t.Parallel()

	var calls []string

	manager := New()
	if _, ok := manager.Focused(); ok {
		t.Fatal("empty manager reported a focused scene")
	}

	registerScene(t, manager, "world", false, nil, &calls)
	registerScene(t, manager, "inventory", true, nil, &calls)

	if err := manager.Replace(context.Background(), "world"); err != nil {
		t.Fatal(err)
	}

	if err := manager.Push(context.Background(), "inventory"); err != nil {
		t.Fatal(err)
	}

	if got, ok := manager.Focused(); !ok || got != "inventory" {
		t.Fatalf("focused = %q, %v; want inventory, true", got, ok)
	}
}

// TestOverlaySlotsToggleReplaceAndExclude protects the spatial invariants that
// allow opposite side overlays to coexist while full overlays remain exclusive.
func TestOverlaySlotsToggleReplaceAndExclude(t *testing.T) {
	t.Parallel()

	var calls []string

	manager := New()
	for _, id := range []string{"world", "inventory", "skills", "character", "help"} {
		registerScene(t, manager, id, false, nil, &calls)
	}

	ctx := context.Background()
	if err := manager.Replace(ctx, "world"); err != nil {
		t.Fatal(err)
	}

	if err := manager.ToggleOverlay(ctx, "inventory", OverlayRight); err != nil {
		t.Fatal(err)
	}

	if err := manager.ToggleOverlay(ctx, "character", OverlayLeft); err != nil {
		t.Fatal(err)
	}

	if got := manager.Stack(); !reflect.DeepEqual(got, []string{"world", "inventory", "character"}) {
		t.Fatalf("coexisting side overlays = %v", got)
	}

	if err := manager.ToggleOverlay(ctx, "skills", OverlayRight); err != nil {
		t.Fatal(err)
	}

	if got := manager.Stack(); !reflect.DeepEqual(got, []string{"world", "character", "skills"}) {
		t.Fatalf("same-slot replacement = %v", got)
	}

	if err := manager.ToggleOverlay(ctx, "skills", OverlayRight); err != nil {
		t.Fatal(err)
	}

	if got := manager.Stack(); !reflect.DeepEqual(got, []string{"world", "character"}) {
		t.Fatalf("same-overlay toggle close = %v", got)
	}

	if err := manager.ToggleOverlay(ctx, "help", OverlayFull); err != nil {
		t.Fatal(err)
	}

	if got := manager.Stack(); !reflect.DeepEqual(got, []string{"world", "help"}) {
		t.Fatalf("full overlay did not evict side slots = %v", got)
	}

	if err := manager.ToggleOverlay(ctx, "inventory", OverlayRight); err != nil {
		t.Fatal(err)
	}

	if got := manager.Stack(); !reflect.DeepEqual(got, []string{"world", "inventory"}) {
		t.Fatalf("side overlay did not evict full slot = %v", got)
	}
}

type passthroughScene struct {
	*testScene
	view         string
	receivedView *string
}

type routedScene struct {
	*testScene
	passes bool
	modes  map[string]string
}

// PassesInputBelow supplies the fixture's modal or passthrough behavior.
func (s *routedScene) PassesInputBelow() bool { return s.passes }

// UpdateRoutedInput captures the combined input mode and world view assigned to a scene.
func (s *routedScene) UpdateRoutedInput(_ context.Context, _ time.Duration, _ bool, mode, view string) error {
	s.modes[s.id] = mode + ":" + view
	return nil
}

// TestSpatialOverlayInputRoutesHUDAndBothSidePanels verifies that simultaneous
// side panels consume the world view while retaining their distinct input modes.
func TestSpatialOverlayInputRoutesHUDAndBothSidePanels(t *testing.T) {
	t.Parallel()

	var calls []string

	modes := make(map[string]string)
	manager := New()

	for _, id := range []string{"world", "inventory", "character"} {
		id := id
		if err := manager.Register(id, func(context.Context) (Scene, error) {
			return &routedScene{testScene: &testScene{id: id, calls: &calls}, passes: id != "world", modes: modes}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}

	ctx := context.Background()
	if err := manager.Replace(ctx, "world"); err != nil {
		t.Fatal(err)
	}

	if err := manager.ToggleOverlay(ctx, "inventory", OverlayRight); err != nil {
		t.Fatal(err)
	}

	if err := manager.ToggleOverlay(ctx, "character", OverlayLeft); err != nil {
		t.Fatal(err)
	}

	if err := manager.Update(ctx, time.Second/60); err != nil {
		t.Fatal(err)
	}

	wantModes := map[string]string{
		"world":     "gameplay_pointer:none",
		"inventory": "pointer:none",
		"character": "focused:none",
	}
	if !reflect.DeepEqual(modes, wantModes) {
		t.Fatalf("spatial input modes = %#v", modes)
	}

	allowed, view := manager.InputPolicy("world")
	if !allowed || view != "none" {
		t.Fatalf("world policy with both halves covered = %v,%q", allowed, view)
	}
}

// PassesInputBelow keeps gameplay available to the world beneath this fixture.
func (s *passthroughScene) PassesInputBelow() bool { return true }

// WorldView supplies the visible world region authored by this fixture.
func (s *passthroughScene) WorldView() string { return s.view }

// UpdateInputFocused captures framing independently from the fixture's UI focus.
func (s *passthroughScene) UpdateInputFocused(_ context.Context, _ time.Duration, _ bool, _ bool, view string) error {
	*s.receivedView = view
	return nil
}

// TestOverlaySeparatesUIFocusFromGameplayInputAndWorldFraming proves a
// passthrough overlay can own UI focus without starving the world of input.
func TestOverlaySeparatesUIFocusFromGameplayInputAndWorldFraming(t *testing.T) {
	t.Parallel()

	var (
		calls        []string
		receivedView string
	)

	manager := New()
	registerScene(t, manager, "world", false, nil, &calls)

	if err := manager.Register("inventory", func(context.Context) (Scene, error) {
		return &passthroughScene{
			testScene:    &testScene{id: "inventory", calls: &calls},
			view:         "left",
			receivedView: &receivedView,
		}, nil
	}); err != nil {
		t.Fatal(err)
	}

	if err := manager.Replace(context.Background(), "world"); err != nil {
		t.Fatal(err)
	}

	if err := manager.Push(context.Background(), "inventory"); err != nil {
		t.Fatal(err)
	}

	if focused, ok := manager.Focused(); !ok || focused != "inventory" {
		t.Fatalf("focused = %q, %v", focused, ok)
	}

	allowed, view := manager.InputPolicy("world")
	if !allowed {
		t.Fatal("inventory did not pass gameplay input to world")
	}

	if view != "left" {
		t.Fatalf("input policy world view = %q, want left", view)
	}

	if err := manager.Update(context.Background(), time.Second); err != nil {
		t.Fatal(err)
	}

	if receivedView != "left" {
		t.Fatalf("world view = %q, want left", receivedView)
	}
}

// registerScene installs a lifecycle-recording fixture while keeping individual
// tests focused on the navigation behavior they exercise.
func registerScene(t *testing.T, manager *Manager, id string, blocks bool, enterErr error, calls *[]string) {
	t.Helper()

	if err := manager.Register(id, func(context.Context) (Scene, error) {
		return &testScene{id: id, blocks: blocks, enterErr: enterErr, calls: calls}, nil
	}); err != nil {
		t.Fatal(err)
	}
}
