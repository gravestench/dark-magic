package host

import (
	"context"
	"errors"
	"fmt"
)

// Enable marks one component explicitly desired and rolls back only dependencies introduced by a failed request.
func (manager *Manager) Enable(ctx context.Context, id string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, exists := manager.entries[id]
	if !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}

	// Desired remains true after failure so status preserves the operator's unmet request.
	entry.desired = true

	var enabled []string
	if err := manager.enable(ctx, id, make(map[string]bool), &enabled); err != nil {
		entry.err = err

		manager.rollbackEnabled(ctx, enabled)

		return err
	}

	return nil
}

// Disable refuses to invalidate an active dependent and clears desired state only after that safety check.
func (manager *Manager) Disable(ctx context.Context, id string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, exists := manager.entries[id]
	if !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}

	if dependents := manager.activeDependents(id); len(dependents) != 0 {
		return fmt.Errorf("host: cannot disable %q; active dependents: %v", id, dependents)
	}

	entry.desired = false

	return manager.disable(ctx, id)
}

// DisableCascade stops transitive active dependents before their dependencies and attempts every planned stop.
func (manager *Manager) DisableCascade(ctx context.Context, id string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	if _, exists := manager.entries[id]; !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}

	// Post-order places the furthest dependent first, keeping each dependency live until it is unused.
	order := manager.dependentOrder(id)

	var errs []error

	for _, current := range order {
		manager.entries[current].desired = false
		if err := manager.disable(ctx, current); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Restart replaces an enabled leaf instance while rejecting dependents that would observe the interruption.
func (manager *Manager) Restart(ctx context.Context, id string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, exists := manager.entries[id]
	if !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}

	if entry.state != StateEnabled {
		return fmt.Errorf("host: cannot restart %q from state %q", id, entry.state)
	}

	if dependents := manager.activeDependents(id); len(dependents) != 0 {
		return fmt.Errorf("host: cannot restart %q; active dependents: %v", id, dependents)
	}

	entry.desired = true

	if err := manager.disable(ctx, id); err != nil {
		return err
	}

	var enabled []string

	return manager.enable(ctx, id, make(map[string]bool), &enabled)
}

// ApplyDesired enables in registration/dependency order and disables in reverse registration order.
func (manager *Manager) ApplyDesired(ctx context.Context, desired map[string]bool) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	// Validate the complete request before changing any desired bits, preserving all-or-nothing input handling.
	for id := range desired {
		if _, exists := manager.entries[id]; !exists {
			return fmt.Errorf("host: managed component %q is not registered", id)
		}
	}

	for id, entry := range manager.entries {
		entry.desired = desired[id]
	}

	var errs []error

	// Recursion supplies dependency order while registration order makes independent choices deterministic.
	for _, id := range manager.order {
		if !manager.entries[id].desired {
			continue
		}

		var enabled []string
		if err := manager.enable(ctx, id, make(map[string]bool), &enabled); err != nil {
			errs = append(errs, err)
		}
	}

	// Reverse registration order is a candidate plan; active dependents remain the actual safety condition.
	for index := len(manager.order) - 1; index >= 0; index-- {
		id := manager.order[index]
		if manager.entries[id].desired || len(manager.activeDependents(id)) != 0 {
			continue
		}

		if err := manager.disable(ctx, id); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// enable recursively establishes dependencies, then publishes an instance only after successful startup.
func (manager *Manager) enable(
	ctx context.Context,
	id string,
	visiting map[string]bool,
	enabled *[]string,
) error {
	entry, exists := manager.entries[id]
	if !exists {
		return fmt.Errorf("host: managed component %q is not registered", id)
	}

	if entry.state == StateEnabled {
		return nil
	}

	if visiting[id] {
		return fmt.Errorf("host: dependency cycle while enabling %q", id)
	}

	// visiting covers only the active recursion path, so shared dependencies are legal while cycles are not.
	visiting[id] = true
	defer delete(visiting, id)

	for _, dependency := range entry.definition.DependsOn {
		if err := manager.enable(ctx, dependency, visiting, enabled); err != nil {
			return fmt.Errorf("host: enable %q dependency %q: %w", id, dependency, err)
		}
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("host: enable %q: %w", id, err)
	}

	manager.transition(entry, StateEnabling, nil)

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
		manager.transition(entry, StateFailed, wrapped)

		return wrapped
	}

	// Publish ownership only after Start succeeds; dependents must never observe a half-started instance.
	entry.instance = instance
	manager.transition(entry, StateEnabled, nil)

	*enabled = append(*enabled, id)

	return nil
}

// disable retains a failed instance when Stop fails so a later cleanup attempt can still reach it.
func (manager *Manager) disable(ctx context.Context, id string) error {
	entry := manager.entries[id]
	if entry.state == StateDisabled {
		return nil
	}

	if entry.instance == nil {
		// Factory failures own no resource, but disabling must still clear their failed status.
		manager.transition(entry, StateDisabled, nil)

		return nil
	}

	manager.transition(entry, StateDisabling, nil)

	if err := entry.instance.Stop(ctx); err != nil {
		wrapped := fmt.Errorf("host: disable %q: %w", id, err)
		manager.transition(entry, StateFailed, wrapped)

		return wrapped
	}

	entry.instance = nil
	manager.transition(entry, StateDisabled, nil)

	return nil
}

// rollbackEnabled unwinds only components created by the current operation and not explicitly desired elsewhere.
func (manager *Manager) rollbackEnabled(ctx context.Context, enabled []string) {
	for index := len(enabled) - 1; index >= 0; index-- {
		entry := manager.entries[enabled[index]]
		if !entry.desired {
			// Cleanup is best effort because the initiating failure is the most useful error to preserve.
			_ = manager.disable(ctx, enabled[index])
		}
	}
}

// activeDependents returns enabled or enabling direct dependents in deterministic registration order.
func (manager *Manager) activeDependents(id string) []string {
	var result []string

	for _, candidateID := range manager.order {
		candidate := manager.entries[candidateID]
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

// dependentOrder performs a post-order walk so every active dependent precedes the requested dependency.
func (manager *Manager) dependentOrder(id string) []string {
	seen := make(map[string]bool)

	var result []string

	var visit func(string)

	visit = func(current string) {
		if seen[current] {
			return
		}

		seen[current] = true
		for _, dependent := range manager.activeDependents(current) {
			visit(dependent)
		}

		result = append(result, current)
	}
	visit(id)

	return result
}

// validateKnownCycles rejects cycles among registered definitions while permitting not-yet-registered dependencies.
func (manager *Manager) validateKnownCycles() error {
	states := make(map[string]uint8, len(manager.entries))

	var visit func(string) error

	visit = func(id string) error {
		switch states[id] {
		case 1:
			return fmt.Errorf("host: managed dependency cycle includes %q", id)
		case 2:
			return nil
		}

		states[id] = 1
		for _, dependency := range manager.entries[id].definition.DependsOn {
			if _, exists := manager.entries[dependency]; !exists {
				continue
			}

			if err := visit(dependency); err != nil {
				return err
			}
		}

		states[id] = 2

		return nil
	}

	for id := range manager.entries {
		if err := visit(id); err != nil {
			return err
		}
	}

	return nil
}
