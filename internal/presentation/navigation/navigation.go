// Package navigation manages short-lived screens and overlays independently of
// the application component lifecycle.
package navigation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Scene is one screen or overlay instance.
type Scene interface {
	Create(context.Context) error
	Enter(context.Context) error
	Update(context.Context, time.Duration) error
	Render(context.Context) error
	Exit(context.Context) error
	Destroy(context.Context) error
}

// UpdateBlocker lets an overlay pause scenes below it.
type UpdateBlocker interface {
	BlocksUpdateBelow() bool
}

// FocusedUpdater distinguishes simulation updates from input focus. Only the
// top scene is focused even when transparent overlays allow updates below.
type FocusedUpdater interface {
	UpdateFocused(context.Context, time.Duration, bool) error
}

// InputAwareUpdater receives UI focus and routed gameplay-input access
// separately. A panel may own UI focus while deliberately allowing the world
// below it to continue consuming gameplay actions.
type InputAwareUpdater interface {
	UpdateInputFocused(context.Context, time.Duration, bool, bool, string) error
}

// RoutedInputUpdater receives the precise input channel assigned to each
// visible scene. Modes are focused, gameplay, pointer, gameplay_pointer, none.
type RoutedInputUpdater interface {
	UpdateRoutedInput(context.Context, time.Duration, bool, string, string) error
}

// InputPassthrough declares that a scene does not consume gameplay input meant
// for scenes beneath it. Modal and pause scenes do not implement this policy.
type InputPassthrough interface {
	PassesInputBelow() bool
}

// WorldViewer describes the unobscured region an overlay leaves for world
// presentation. Values are authored for the fixed logical viewport.
type WorldViewer interface{ WorldView() string }

// Factory creates a fresh scene instance.
type Factory func(context.Context) (Scene, error)

// entry keeps a registered ID, live instance, and optional spatial slot together while stack operations reorder them.
type entry struct {
	id    string
	slot  string
	scene Scene
}

const (
	// Overlay view values describe which logical half remains available to world
	// presentation and routed pointer input.
	OverlayLeft  = "left"
	OverlayRight = "right"
	OverlayFull  = "full"
)

// Manager owns a deterministic bottom-to-top scene stack. All operations are
// serialized because scene callbacks may touch single-threaded engine APIs.
type Manager struct {
	mu        sync.Mutex
	factories map[string]Factory
	stack     []entry
}

// New constructs an empty scene manager.
func New() *Manager { return &Manager{factories: make(map[string]Factory)} }

// Register makes a scene kind available for navigation.
func (m *Manager) Register(id string, factory Factory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" || factory == nil {
		return errors.New("navigation: scene ID and factory are required")
	}

	if _, exists := m.factories[id]; exists {
		return fmt.Errorf("navigation: scene %q is already registered", id)
	}

	m.factories[id] = factory

	return nil
}

// Unregister removes an inactive scene kind.
func (m *Manager) Unregister(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.factories[id]; !exists {
		return fmt.Errorf("navigation: scene %q is not registered", id)
	}

	for _, current := range m.stack {
		if current.id == id {
			return fmt.Errorf("navigation: scene %q is active", id)
		}
	}

	delete(m.factories, id)

	return nil
}

// Replace transactionally enters a new root screen before exiting and
// destroying the existing stack.
func (m *Manager) Replace(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Entering the replacement first makes failure transactional: callers keep
	// the prior stack when the new scene cannot initialize.
	next, err := m.createAndEnter(ctx, id)
	if err != nil {
		return err
	}

	old := m.stack
	m.stack = []entry{next}

	if err := closeEntries(ctx, old); err != nil {
		return fmt.Errorf("navigation: replace with %q: %w", id, err)
	}

	return nil
}

// Push enters an overlay above the current stack.
func (m *Manager) Push(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	next, err := m.createAndEnter(ctx, id)
	if err != nil {
		return err
	}

	m.stack = append(m.stack, next)

	return nil
}

// ToggleOverlay opens an overlay in a spatial slot, closes it when already active, and publishes replacements before
// cleaning up conflicting overlays. Left and right slots may coexist; a full overlay excludes both side slots.
func (m *Manager) ToggleOverlay(ctx context.Context, id, slot string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !validOverlaySlot(slot) {
		return fmt.Errorf("navigation: invalid overlay slot %q", slot)
	}

	if index, ok := activeOverlayIndex(m.stack, id, slot); ok {
		current := m.stack[index]
		m.stack = append(m.stack[:index], m.stack[index+1:]...)

		return closeEntries(ctx, []entry{current})
	}

	next, err := m.createAndEnter(ctx, id)
	if err != nil {
		return err
	}

	next.slot = slot

	retained := make([]entry, 0, len(m.stack)+1)
	removed := make([]entry, 0, 2)

	for _, current := range m.stack {
		if overlaySlotsConflict(current.slot, slot) {
			removed = append(removed, current)
			continue
		}

		retained = append(retained, current)
	}

	// Publish the new stack before cleanup, matching Replace: navigation state
	// moves forward even if an outgoing scene reports an Exit or Destroy error.
	m.stack = append(retained, next)

	if err := closeEntries(ctx, removed); err != nil {
		return fmt.Errorf("navigation: replace %s overlay with %q: %w", slot, id, err)
	}

	return nil
}

// validOverlaySlot limits spatial policy to the three layouts understood by
// world framing and input routing.
func validOverlaySlot(slot string) bool {
	return slot == OverlayLeft || slot == OverlayRight || slot == OverlayFull
}

// activeOverlayIndex searches from the top so toggling duplicate historical
// entries, if any exist, closes the one users can currently see.
func activeOverlayIndex(stack []entry, id, slot string) (int, bool) {
	for index := len(stack) - 1; index >= 0; index-- {
		if stack[index].id == id && stack[index].slot == slot {
			return index, true
		}
	}

	return 0, false
}

// overlaySlotsConflict encodes the layout invariant: matching side slots
// exclude each other, and a full overlay excludes either side.
func overlaySlotsConflict(current, requested string) bool {
	if current == requested || current == OverlayFull {
		return true
	}

	return requested == OverlayFull && (current == OverlayLeft || current == OverlayRight)
}

// Pop exits and destroys the top scene.
func (m *Manager) Pop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.stack) == 0 {
		return errors.New("navigation: scene stack is empty")
	}

	index := len(m.stack) - 1
	current := m.stack[index]
	m.stack = m.stack[:index]

	return closeEntries(ctx, []entry{current})
}

// Update updates the visible stack from the highest update-blocking overlay to
// the focused top scene.
func (m *Manager) Update(ctx context.Context, elapsed time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	start := firstUpdatedEntry(m.stack)

	view := worldView(m.stack)
	for index := start; index < len(m.stack); index++ {
		focused := index == len(m.stack)-1
		inputAllowed := focused || passesInputFrom(m.stack, index+1)
		mode := routedInputMode(m.stack[index], index, focused, inputAllowed)

		if err := updateEntry(ctx, elapsed, m.stack[index], focused, inputAllowed, mode, view); err != nil {
			return fmt.Errorf("navigation: update %q: %w", m.stack[index].id, err)
		}
	}

	return nil
}

// firstUpdatedEntry finds the highest blocking scene. Starting there pauses
// lower simulations while still updating overlays stacked above the blocker.
func firstUpdatedEntry(stack []entry) int {
	for index := len(stack) - 1; index >= 0; index-- {
		blocker, ok := stack[index].scene.(UpdateBlocker)
		if ok && blocker.BlocksUpdateBelow() {
			return index
		}
	}

	return 0
}

// routedInputMode translates focus, passthrough, and placement into the stable strings consumed by RoutedInputUpdater.
func routedInputMode(current entry, index int, focused, inputAllowed bool) string {
	if focused {
		return "focused"
	}

	if !inputAllowed {
		return "none"
	}

	if current.slot != "" {
		return "pointer"
	}

	if index == 0 {
		return "gameplay_pointer"
	}

	return "gameplay"
}

// updateEntry selects the richest update contract a scene implements. The
// assertion order is part of the API: a scene implementing several contracts
// receives exactly one callback with the most detailed routing information.
func updateEntry(
	ctx context.Context,
	elapsed time.Duration,
	current entry,
	focused bool,
	inputAllowed bool,
	mode string,
	view string,
) error {
	if updater, ok := current.scene.(RoutedInputUpdater); ok {
		return updater.UpdateRoutedInput(ctx, elapsed, focused, mode, view)
	}

	if updater, ok := current.scene.(InputAwareUpdater); ok {
		return updater.UpdateInputFocused(ctx, elapsed, focused, inputAllowed, view)
	}

	if updater, ok := current.scene.(FocusedUpdater); ok {
		return updater.UpdateFocused(ctx, elapsed, focused)
	}

	return current.scene.Update(ctx, elapsed)
}

// InputPolicy reports whether input passes through every higher scene and returns the current unobscured world region.
func (m *Manager) InputPolicy(id string) (bool, string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for index := len(m.stack) - 1; index >= 0; index-- {
		if m.stack[index].id == id {
			return index == len(m.stack)-1 || passesInputFrom(m.stack, index+1), worldView(m.stack)
		}
	}

	return false, "center"
}

// worldView reduces all active overlay slots to the remaining world region.
// Explicit side slots take precedence over a custom view on the top scene.
func worldView(stack []entry) string {
	if len(stack) == 0 {
		return "center"
	}

	left, right := false, false

	for _, current := range stack {
		switch current.slot {
		case OverlayFull:
			return "none"
		case OverlayLeft:
			left = true
		case OverlayRight:
			right = true
		}
	}

	if left && right {
		return "none"
	}

	if left {
		return "right"
	}

	if right {
		return "left"
	}

	if viewer, ok := stack[len(stack)-1].scene.(WorldViewer); ok {
		return viewer.WorldView()
	}

	return "center"
}

// passesInputFrom requires every scene above start to opt into passthrough;
// one modal scene therefore closes the gameplay route for all lower scenes.
func passesInputFrom(stack []entry, start int) bool {
	for index := start; index < len(stack); index++ {
		passthrough, ok := stack[index].scene.(InputPassthrough)
		if !ok || !passthrough.PassesInputBelow() {
			return false
		}
	}

	return true
}

// Render renders every scene from bottom to top.
func (m *Manager) Render(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, current := range m.stack {
		if err := current.scene.Render(ctx); err != nil {
			return fmt.Errorf("navigation: render %q: %w", current.id, err)
		}
	}

	return nil
}

// Close exits and destroys the full stack from top to bottom.
func (m *Manager) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	old := m.stack
	// Detach the stack before lifecycle callbacks so the manager remains closed
	// even when cleanup returns one or more errors.
	m.stack = nil

	return closeEntries(ctx, old)
}

// Stack returns scene IDs from bottom to top.
func (m *Manager) Stack() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]string, len(m.stack))
	for index, current := range m.stack {
		result[index] = current.id
	}

	return result
}

// Focused returns the ID of the topmost scene receiving input. The value is a
// stable observation only; callers must not use it to coordinate navigation.
func (m *Manager) Focused() (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.stack) == 0 {
		return "", false
	}

	return m.stack[len(m.stack)-1].id, true
}

// createAndEnter completes both startup phases before returning an entry. Any
// partial scene is destroyed so callers never publish an unusable stack item.
func (m *Manager) createAndEnter(ctx context.Context, id string) (entry, error) {
	factory, exists := m.factories[id]
	if !exists {
		return entry{}, fmt.Errorf("navigation: scene %q is not registered", id)
	}

	scene, err := factory(ctx)
	if err != nil {
		return entry{}, fmt.Errorf("navigation: create %q: %w", id, err)
	}

	if scene == nil {
		return entry{}, fmt.Errorf("navigation: create %q: factory returned nil", id)
	}

	if err := scene.Create(ctx); err != nil {
		// Join preserves both the primary lifecycle error and a cleanup failure.
		return entry{}, errors.Join(fmt.Errorf("navigation: initialize %q: %w", id, err), scene.Destroy(ctx))
	}

	if err := scene.Enter(ctx); err != nil {
		return entry{}, errors.Join(fmt.Errorf("navigation: enter %q: %w", id, err), scene.Destroy(ctx))
	}

	return entry{id: id, scene: scene}, nil
}

// closeEntries unwinds the stack from top to bottom and attempts both cleanup
// phases for every scene, joining failures so one bad callback cannot leak the
// remaining scenes.
func closeEntries(ctx context.Context, entries []entry) error {
	var errs []error

	for index := len(entries) - 1; index >= 0; index-- {
		current := entries[index]
		if err := current.scene.Exit(ctx); err != nil {
			errs = append(errs, fmt.Errorf("navigation: exit %q: %w", current.id, err))
		}

		if err := current.scene.Destroy(ctx); err != nil {
			errs = append(errs, fmt.Errorf("navigation: destroy %q: %w", current.id, err))
		}
	}

	return errors.Join(errs...)
}
