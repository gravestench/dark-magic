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
	// StateDisabled has no live component instance.
	StateDisabled State = "disabled"
	// StateEnabling covers factory construction and component startup.
	StateEnabling State = "enabling"
	// StateEnabled has one live component instance available to dependents.
	StateEnabled State = "enabled"
	// StateDisabling covers component shutdown before instance ownership is released.
	StateDisabling State = "disabling"
	// StateFailed retains the latest lifecycle error for status inspection.
	StateFailed State = "failed"
)

// ManagedDefinition describes a component that can be instantiated and reconciled at runtime.
type ManagedDefinition struct {
	ID        string
	DependsOn []string
	New       func(context.Context) (Component, error)
}

// StateExporter releases a replacement-safe snapshot that does not retain old-instance resources.
type StateExporter interface {
	ExportState(context.Context) (any, error)
}

// StateImporter restores optional replacement state before the new component becomes observable.
type StateImporter interface {
	ImportState(context.Context, any) error
}

// Status reports desired and actual state without exposing the live component instance.
type Status struct {
	ID      string `json:"id"`
	Desired bool   `json:"desired"`
	State   State  `json:"state"`
	Err     error  `json:"-"`
}

// Event records a diagnostic transition; callers must use Status rather than events for synchronization.
type Event struct {
	ID       string
	Previous State
	State    State
	Err      error
	Time     time.Time
}

// managedEntry keeps definition, desired state, actual state, and instance ownership under the manager mutex.
type managedEntry struct {
	definition ManagedDefinition
	desired    bool
	state      State
	instance   Component
	err        error
}

// Manager serializes runtime component reconciliation and emits non-blocking diagnostic transitions.
type Manager struct {
	mu          sync.Mutex
	entries     map[string]*managedEntry
	order       []string
	subscribers map[uint64]chan Event
	nextSubID   uint64
}

// NewManager constructs an empty manager whose registration order remains the deterministic reconciliation order.
func NewManager() *Manager {
	return &Manager{
		entries:     make(map[string]*managedEntry),
		subscribers: make(map[uint64]chan Event),
	}
}

// Register defensively owns dependency metadata and rejects newly completed cycles before publishing a definition.
func (manager *Manager) Register(definition ManagedDefinition) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if definition.ID == "" {
		return errors.New("host: managed component ID is required")
	}

	if definition.New == nil {
		return fmt.Errorf("host: managed component %q has no factory", definition.ID)
	}

	if _, exists := manager.entries[definition.ID]; exists {
		return fmt.Errorf("host: managed component %q is already registered", definition.ID)
	}

	// The manager owns graph metadata because later caller mutations would otherwise change live reconciliation.
	definition.DependsOn = append([]string(nil), definition.DependsOn...)
	manager.entries[definition.ID] = &managedEntry{
		definition: definition,
		state:      StateDisabled,
	}
	manager.order = append(manager.order, definition.ID)

	// Missing dependencies are legal during staged registration, but completing a cycle is not.
	if err := manager.validateKnownCycles(); err != nil {
		delete(manager.entries, definition.ID)
		manager.order = manager.order[:len(manager.order)-1]

		return err
	}

	return nil
}

// Status returns one mutex-consistent snapshot and never exposes mutable entry or instance ownership.
func (manager *Manager) Status(id string) (Status, bool) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, exists := manager.entries[id]
	if !exists {
		return Status{}, false
	}

	return statusForEntry(id, entry), true
}

// Statuses follows registration order so diagnostics remain stable across map iteration and process runs.
func (manager *Manager) Statuses() []Status {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	result := make([]Status, 0, len(manager.order))
	for _, id := range manager.order {
		result = append(result, statusForEntry(id, manager.entries[id]))
	}

	return result
}

// statusForEntry copies the externally visible fields while the caller holds the manager mutex.
func statusForEntry(id string, entry *managedEntry) Status {
	return Status{
		ID:      id,
		Desired: entry.desired,
		State:   entry.state,
		Err:     entry.err,
	}
}

// Subscribe returns a bounded diagnostic stream whose slow consumers cannot stall lifecycle transitions.
func (manager *Manager) Subscribe(buffer int) (<-chan Event, func()) {
	if buffer < 1 {
		buffer = 1
	}

	manager.mu.Lock()
	id := manager.nextSubID
	manager.nextSubID++
	events := make(chan Event, buffer)
	manager.subscribers[id] = events
	manager.mu.Unlock()

	var once sync.Once

	cancel := func() {
		once.Do(func() {
			// Sharing the publication mutex prevents a transition from sending to a closed channel.
			manager.mu.Lock()
			delete(manager.subscribers, id)
			close(events)
			manager.mu.Unlock()
		})
	}

	return events, cancel
}

// transition updates authoritative state before best-effort event publication under the manager mutex.
func (manager *Manager) transition(entry *managedEntry, state State, err error) {
	previous := entry.state
	entry.state = state
	entry.err = err

	event := Event{
		ID:       entry.definition.ID,
		Previous: previous,
		State:    state,
		Err:      err,
		Time:     time.Now(),
	}

	for _, subscriber := range manager.subscribers {
		select {
		case subscriber <- event:
		default:
			// Dropping diagnostics is preferable to blocking the authoritative lifecycle transition.
		}
	}
}
