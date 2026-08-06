package host

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// State is the actual lifecycle state of a managed component.
type State string

const (
	StateDisabled  State = "disabled"
	StateEnabling  State = "enabling"
	StateEnabled   State = "enabled"
	StateDisabling State = "disabling"
	StateFailed    State = "failed"
)

// ManagedDefinition describes a component that can be instantiated and
// enabled or disabled at runtime.
type ManagedDefinition struct {
	ID        string
	DependsOn []string
	New       func(context.Context) (Component, error)
}

// StateExporter and StateImporter form the optional versioned persistence seam
// used during transactional replacement. State must not retain ownership of
// resources belonging to the old component instance.
type StateExporter interface {
	ExportState(context.Context) (any, error)
}

type StateImporter interface {
	ImportState(context.Context, any) error
}

// Status reports desired and actual state without exposing the instance.
type Status struct {
	ID      string `json:"id"`
	Desired bool   `json:"desired"`
	State   State  `json:"state"`
	Err     error  `json:"-"`
}

// Event records an observed state transition. Events are diagnostic; callers
// must not use them as a lifecycle synchronization mechanism.
type Event struct {
	ID       string
	Previous State
	State    State
	Err      error
	Time     time.Time
}

type managedEntry struct {
	definition ManagedDefinition
	desired    bool
	state      State
	instance   Component
	err        error
}

// Manager reconciles runtime component state. All transitions are serialized.
type Manager struct {
	mu          sync.Mutex
	entries     map[string]*managedEntry
	order       []string
	subscribers map[uint64]chan Event
	nextSubID   uint64
}

// NewManager constructs an empty runtime manager.
func NewManager() *Manager {
	return &Manager{
		entries:     make(map[string]*managedEntry),
		subscribers: make(map[uint64]chan Event),
	}
}

// Register adds an available component definition without enabling it.
func (m *Manager) Register(def ManagedDefinition) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if def.ID == "" {
		return errors.New("host: managed component ID is required")
	}
	if def.New == nil {
		return fmt.Errorf("host: managed component %q has no factory", def.ID)
	}
	if _, exists := m.entries[def.ID]; exists {
		return fmt.Errorf("host: managed component %q is already registered", def.ID)
	}

	dependencies := append([]string(nil), def.DependsOn...)
	def.DependsOn = dependencies
	m.entries[def.ID] = &managedEntry{definition: def, state: StateDisabled}
	m.order = append(m.order, def.ID)
	if err := m.validateKnownCycles(); err != nil {
		delete(m.entries, def.ID)
		m.order = m.order[:len(m.order)-1]
		return err
	}
	return nil
}

// Enable sets desired state and enables dependencies before id. Components
// enabled solely for a failed request are rolled back.
func (m *Manager) Enable(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.entries[id]
	if !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}
	entry.desired = true

	var enabled []string
	if err := m.enable(ctx, id, make(map[string]bool), &enabled); err != nil {
		entry.err = err
		for i := len(enabled) - 1; i >= 0; i-- {
			rollback := m.entries[enabled[i]]
			if !rollback.desired {
				_ = m.disable(ctx, enabled[i])
			}
		}
		return err
	}
	return nil
}

// Disable disables id if it has no enabled dependents.
func (m *Manager) Disable(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.entries[id]
	if !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}
	if dependents := m.activeDependents(id); len(dependents) != 0 {
		return fmt.Errorf("host: cannot disable %q; active dependents: %v", id, dependents)
	}
	entry.desired = false
	return m.disable(ctx, id)
}

// DisableCascade disables active dependents before disabling id.
func (m *Manager) DisableCascade(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.entries[id]; !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}
	order := m.dependentOrder(id)
	var errs []error
	for _, current := range order {
		m.entries[current].desired = false
		if err := m.disable(ctx, current); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Restart replaces an enabled component after first safely disabling it. Active
// dependents make a non-cascading restart invalid.
func (m *Manager) Restart(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.entries[id]
	if !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}
	if entry.state != StateEnabled {
		return fmt.Errorf("host: cannot restart %q from state %q", id, entry.state)
	}
	if dependents := m.activeDependents(id); len(dependents) != 0 {
		return fmt.Errorf("host: cannot restart %q; active dependents: %v", id, dependents)
	}
	entry.desired = true
	if err := m.disable(ctx, id); err != nil {
		return err
	}
	var enabled []string
	return m.enable(ctx, id, make(map[string]bool), &enabled)
}

// Replace transactionally swaps a definition. For an enabled component it
// constructs, restores, and starts the replacement before stopping and
// publishing it. A failure leaves the existing definition and instance active.
func (m *Manager) Replace(ctx context.Context, def ManagedDefinition) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.entries[def.ID]
	if !exists {
		return fmt.Errorf("host: managed component %q is not registered", def.ID)
	}
	if def.New == nil {
		return fmt.Errorf("host: managed component %q has no factory", def.ID)
	}
	def.DependsOn = append([]string(nil), def.DependsOn...)
	previousDefinition := entry.definition
	entry.definition = def
	if err := m.validateKnownCycles(); err != nil {
		entry.definition = previousDefinition
		return err
	}
	if entry.state == StateDisabled || entry.state == StateFailed && entry.instance == nil {
		entry.err = nil
		if entry.state == StateFailed {
			m.transition(entry, StateDisabled, nil)
		}
		return nil
	}
	if entry.state != StateEnabled || entry.instance == nil {
		entry.definition = previousDefinition
		return fmt.Errorf("host: cannot replace %q from state %q", def.ID, entry.state)
	}

	var implicitlyEnabled []string
	for _, dependency := range def.DependsOn {
		if err := m.enable(ctx, dependency, make(map[string]bool), &implicitlyEnabled); err != nil {
			entry.definition = previousDefinition
			m.rollbackEnabled(ctx, implicitlyEnabled)
			return fmt.Errorf("host: replace %q dependency %q: %w", def.ID, dependency, err)
		}
	}
	replacement, err := def.New(ctx)
	if err == nil && replacement == nil {
		err = errors.New("factory returned a nil component")
	}
	if err != nil {
		entry.definition = previousDefinition
		m.rollbackEnabled(ctx, implicitlyEnabled)
		return fmt.Errorf("host: replace %q: %w", def.ID, err)
	}

	var state any
	if exporter, ok := entry.instance.(StateExporter); ok {
		state, err = exporter.ExportState(ctx)
		if err != nil {
			entry.definition = previousDefinition
			m.rollbackEnabled(ctx, implicitlyEnabled)
			return fmt.Errorf("host: export state for %q: %w", def.ID, err)
		}
	}
	if importer, ok := replacement.(StateImporter); ok && state != nil {
		if err := importer.ImportState(ctx, state); err != nil {
			entry.definition = previousDefinition
			m.rollbackEnabled(ctx, implicitlyEnabled)
			return fmt.Errorf("host: import state for %q: %w", def.ID, err)
		}
	}
	if err := replacement.Start(ctx); err != nil {
		entry.definition = previousDefinition
		_ = replacement.Stop(ctx)
		m.rollbackEnabled(ctx, implicitlyEnabled)
		return fmt.Errorf("host: start replacement %q: %w", def.ID, err)
	}
	if err := entry.instance.Stop(ctx); err != nil {
		entry.definition = previousDefinition
		_ = replacement.Stop(ctx)
		m.rollbackEnabled(ctx, implicitlyEnabled)
		return fmt.Errorf("host: stop replaced component %q: %w", def.ID, err)
	}

	entry.instance = replacement
	entry.err = nil
	m.transition(entry, StateEnabled, nil)
	return nil
}

// ApplyDesired reconciles all registered components to desired. Enabling occurs
// in dependency order; disabling occurs in reverse dependency order.
func (m *Manager) ApplyDesired(ctx context.Context, desired map[string]bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id := range desired {
		if _, exists := m.entries[id]; !exists {
			return fmt.Errorf("host: managed component %q is not registered", id)
		}
	}
	for id, entry := range m.entries {
		entry.desired = desired[id]
	}

	var errs []error
	for _, id := range m.order {
		if !m.entries[id].desired {
			continue
		}
		var enabled []string
		if err := m.enable(ctx, id, make(map[string]bool), &enabled); err != nil {
			errs = append(errs, err)
		}
	}
	for i := len(m.order) - 1; i >= 0; i-- {
		id := m.order[i]
		if m.entries[id].desired || len(m.activeDependents(id)) != 0 {
			continue
		}
		if err := m.disable(ctx, id); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Status returns a snapshot for id.
func (m *Manager) Status(id string) (Status, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.entries[id]
	if !exists {
		return Status{}, false
	}
	return Status{ID: id, Desired: entry.desired, State: entry.state, Err: entry.err}, true
}

// Statuses returns deterministic snapshots for every registered definition.
func (m *Manager) Statuses() []Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]Status, 0, len(m.order))
	for _, id := range m.order {
		entry := m.entries[id]
		result = append(result, Status{ID: id, Desired: entry.desired, State: entry.state, Err: entry.err})
	}
	return result
}

// Subscribe returns a buffered stream of diagnostic state changes and a cancel
// function. Slow subscribers may miss events so they cannot accidentally stall
// lifecycle transitions; Status is authoritative.
func (m *Manager) Subscribe(buffer int) (<-chan Event, func()) {
	if buffer < 1 {
		buffer = 1
	}
	m.mu.Lock()
	id := m.nextSubID
	m.nextSubID++
	ch := make(chan Event, buffer)
	m.subscribers[id] = ch
	m.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			m.mu.Lock()
			delete(m.subscribers, id)
			close(ch)
			m.mu.Unlock()
		})
	}
	return ch, cancel
}

func (m *Manager) enable(ctx context.Context, id string, visiting map[string]bool, enabled *[]string) error {
	entry, exists := m.entries[id]
	if !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}
	if entry.state == StateEnabled {
		return nil
	}
	if visiting[id] {
		return fmt.Errorf("host: dependency cycle while enabling %q", id)
	}
	visiting[id] = true
	defer delete(visiting, id)

	for _, dependency := range entry.definition.DependsOn {
		if err := m.enable(ctx, dependency, visiting, enabled); err != nil {
			return fmt.Errorf("host: enable %q dependency %q: %w", id, dependency, err)
		}
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("host: enable %q: %w", id, err)
	}

	m.transition(entry, StateEnabling, nil)
	instance, err := entry.definition.New(ctx)
	if err == nil && instance == nil {
		err = errors.New("factory returned a nil component")
	}
	if err == nil {
		err = instance.Start(ctx)
	}
	if err != nil {
		entry.instance = nil
		wrapped := fmt.Errorf("host: enable %q: %w", id, err)
		m.transition(entry, StateFailed, wrapped)
		return wrapped
	}

	entry.instance = instance
	m.transition(entry, StateEnabled, nil)
	*enabled = append(*enabled, id)
	return nil
}

func (m *Manager) disable(ctx context.Context, id string) error {
	entry := m.entries[id]
	if entry.state == StateDisabled {
		return nil
	}
	if entry.instance == nil {
		m.transition(entry, StateDisabled, nil)
		return nil
	}

	m.transition(entry, StateDisabling, nil)
	if err := entry.instance.Stop(ctx); err != nil {
		wrapped := fmt.Errorf("host: disable %q: %w", id, err)
		m.transition(entry, StateFailed, wrapped)
		return wrapped
	}
	entry.instance = nil
	m.transition(entry, StateDisabled, nil)
	return nil
}

func (m *Manager) transition(entry *managedEntry, state State, err error) {
	previous := entry.state
	entry.state = state
	entry.err = err
	event := Event{ID: entry.definition.ID, Previous: previous, State: state, Err: err, Time: time.Now()}
	for _, subscriber := range m.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
}

func (m *Manager) rollbackEnabled(ctx context.Context, enabled []string) {
	for i := len(enabled) - 1; i >= 0; i-- {
		entry := m.entries[enabled[i]]
		if !entry.desired {
			_ = m.disable(ctx, enabled[i])
		}
	}
}

func (m *Manager) activeDependents(id string) []string {
	var result []string
	for _, candidateID := range m.order {
		candidate := m.entries[candidateID]
		if candidate.state != StateEnabled && candidate.state != StateEnabling {
			continue
		}
		for _, dependency := range candidate.definition.DependsOn {
			if dependency == id {
				result = append(result, candidateID)
				break
			}
		}
	}
	return result
}

func (m *Manager) dependentOrder(id string) []string {
	seen := make(map[string]bool)
	var result []string
	var visit func(string)
	visit = func(current string) {
		if seen[current] {
			return
		}
		seen[current] = true
		for _, dependent := range m.activeDependents(current) {
			visit(dependent)
		}
		result = append(result, current)
	}
	visit(id)
	return result
}

func (m *Manager) validateKnownCycles() error {
	states := make(map[string]uint8, len(m.entries))
	var visit func(string) error
	visit = func(id string) error {
		switch states[id] {
		case 1:
			return fmt.Errorf("host: managed dependency cycle includes %q", id)
		case 2:
			return nil
		}
		states[id] = 1
		for _, dependency := range m.entries[id].definition.DependsOn {
			if _, exists := m.entries[dependency]; !exists {
				continue
			}
			if err := visit(dependency); err != nil {
				return err
			}
		}
		states[id] = 2
		return nil
	}
	for id := range m.entries {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}
