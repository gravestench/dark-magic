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

// Factory creates a fresh scene instance.
type Factory func(context.Context) (Scene, error)

type entry struct {
	id    string
	scene Scene
}

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
	start := 0
	for index := len(m.stack) - 1; index >= 0; index-- {
		if blocker, ok := m.stack[index].scene.(UpdateBlocker); ok && blocker.BlocksUpdateBelow() {
			start = index
			break
		}
	}
	for index := start; index < len(m.stack); index++ {
		var err error
		if updater, ok := m.stack[index].scene.(FocusedUpdater); ok {
			err = updater.UpdateFocused(ctx, elapsed, index == len(m.stack)-1)
		} else {
			err = m.stack[index].scene.Update(ctx, elapsed)
		}
		if err != nil {
			return fmt.Errorf("navigation: update %q: %w", m.stack[index].id, err)
		}
	}
	return nil
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
		return entry{}, errors.Join(fmt.Errorf("navigation: initialize %q: %w", id, err), scene.Destroy(ctx))
	}
	if err := scene.Enter(ctx); err != nil {
		return entry{}, errors.Join(fmt.Errorf("navigation: enter %q: %w", id, err), scene.Destroy(ctx))
	}
	return entry{id: id, scene: scene}, nil
}

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
