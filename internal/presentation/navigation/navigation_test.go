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

func (s *testScene) Create(context.Context) error {
	*s.calls = append(*s.calls, "create "+s.id)
	return nil
}

func (s *testScene) Enter(context.Context) error {
	*s.calls = append(*s.calls, "enter "+s.id)
	return s.enterErr
}
func (s *testScene) Update(context.Context, time.Duration) error {
	*s.calls = append(*s.calls, "update "+s.id)
	return nil
}
func (s *testScene) Render(context.Context) error {
	*s.calls = append(*s.calls, "render "+s.id)
	return nil
}
func (s *testScene) Exit(context.Context) error {
	*s.calls = append(*s.calls, "exit "+s.id)
	return nil
}
func (s *testScene) Destroy(context.Context) error {
	*s.calls = append(*s.calls, "destroy "+s.id)
	return nil
}
func (s *testScene) BlocksUpdateBelow() bool { return s.blocks }

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

func (s *routedScene) PassesInputBelow() bool { return s.passes }
func (s *routedScene) UpdateRoutedInput(_ context.Context, _ time.Duration, _ bool, mode, view string) error {
	s.modes[s.id] = mode + ":" + view
	return nil
}

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
	if modes["world"] != "gameplay_pointer:none" || modes["inventory"] != "pointer:none" || modes["character"] != "focused:none" {
		t.Fatalf("spatial input modes = %#v", modes)
	}
	allowed, view := manager.InputPolicy("world")
	if !allowed || view != "none" {
		t.Fatalf("world policy with both halves covered = %v,%q", allowed, view)
	}
}

func (s *passthroughScene) PassesInputBelow() bool { return true }
func (s *passthroughScene) WorldView() string      { return s.view }
func (s *passthroughScene) UpdateInputFocused(_ context.Context, _ time.Duration, _ bool, _ bool, view string) error {
	*s.receivedView = view
	return nil
}

func TestOverlaySeparatesUIFocusFromGameplayInputAndWorldFraming(t *testing.T) {
	t.Parallel()
	var calls []string
	var receivedView string
	manager := New()
	registerScene(t, manager, "world", false, nil, &calls)
	if err := manager.Register("inventory", func(context.Context) (Scene, error) {
		return &passthroughScene{testScene: &testScene{id: "inventory", calls: &calls}, view: "left", receivedView: &receivedView}, nil
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

func registerScene(t *testing.T, manager *Manager, id string, blocks bool, enterErr error, calls *[]string) {
	t.Helper()
	if err := manager.Register(id, func(context.Context) (Scene, error) {
		return &testScene{id: id, blocks: blocks, enterErr: enterErr, calls: calls}, nil
	}); err != nil {
		t.Fatal(err)
	}
}
