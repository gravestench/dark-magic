// Package host provides the explicit lifecycle boundary for Dark Magic's
// long-lived native and scripted components.
package host

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Component is a long-lived application component. Start must not return until
// the component is ready for its dependents. Stop must be safe after a
// successful Start and should honor cancellation and deadlines from ctx.
type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Definition describes one component and its explicit dependencies.
type Definition struct {
	ID        string
	DependsOn []string
	Component Component
}

// Host starts registered components in dependency order and stops them in the
// reverse order. Lifecycle transitions are serialized.
type Host struct {
	mu          sync.Mutex
	definitions map[string]Definition
	order       []string
	started     []string
}

// New constructs an empty host whose registration order breaks ties between otherwise independent components.
func New() *Host {
	return &Host{definitions: make(map[string]Definition)}
}

// Register adds a component definition. Definitions may only be registered
// while the host is stopped.
func (h *Host) Register(def Definition) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.started) != 0 {
		return errors.New("host: cannot register while components are running")
	}

	if def.ID == "" {
		return errors.New("host: component ID is required")
	}

	if def.Component == nil {
		return fmt.Errorf("host: component %q is nil", def.ID)
	}

	if _, exists := h.definitions[def.ID]; exists {
		return fmt.Errorf("host: component %q is already registered", def.ID)
	}

	// Callers often reuse option slices; lifecycle ordering must not change if they mutate one later.
	def.DependsOn = append([]string(nil), def.DependsOn...)
	h.definitions[def.ID] = def
	h.order = append(h.order, def.ID)

	return nil
}

// Start validates the dependency graph and starts every component. If startup
// fails, already-started components are stopped in reverse order before Start
// returns. A host whose startup failed may be started again.
func (h *Host) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.started) != 0 {
		return nil
	}

	order, err := h.dependencyStartOrder()
	if err != nil {
		return err
	}

	for _, id := range order {
		if err := ctx.Err(); err != nil {
			return h.rollback(ctx, fmt.Errorf("host: start %q: %w", id, err))
		}

		if err := h.definitions[id].Component.Start(ctx); err != nil {
			return h.rollback(ctx, fmt.Errorf("host: start %q: %w", id, err))
		}

		h.started = append(h.started, id)
	}

	return nil
}

// Stop stops all successfully started components in reverse order. It attempts
// every stop even if one fails and joins all resulting errors. Repeated calls
// are safe.
func (h *Host) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	err := h.stopStarted(ctx)
	h.started = nil

	return err
}

// Started returns the IDs of successfully started components in start order.
func (h *Host) Started() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	result := make([]string, len(h.started))
	copy(result, h.started)

	return result
}

// rollback attempts reverse cleanup and joins cleanup failure with the original startup cause.
func (h *Host) rollback(ctx context.Context, cause error) error {
	stopErr := h.stopStarted(ctx)
	h.started = nil

	if stopErr == nil {
		return cause
	}

	return errors.Join(cause, fmt.Errorf("host: rollback: %w", stopErr))
}

// stopStarted attempts every reverse-order stop so one component cannot strand earlier dependencies.
func (h *Host) stopStarted(ctx context.Context) error {
	var errs []error

	for index := len(h.started) - 1; index >= 0; index-- {
		id := h.started[index]
		if err := h.definitions[id].Component.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("host: stop %q: %w", id, err))
		}
	}

	return errors.Join(errs...)
}
