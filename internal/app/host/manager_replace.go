package host

import (
	"context"
	"errors"
	"fmt"
)

// Replace transactionally starts a restored replacement before stopping and publishing the existing instance.
func (manager *Manager) Replace(ctx context.Context, definition ManagedDefinition) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	entry, previousDefinition, err := manager.prepareReplacement(definition)
	if err != nil {
		return err
	}

	if replacementOnlyUpdatesDefinition(entry) {
		// Without an owned instance, publishing metadata is enough and avoids creating an unwanted component.
		entry.err = nil
		if entry.state == StateFailed {
			manager.transition(entry, StateDisabled, nil)
		}

		return nil
	}

	if entry.state != StateEnabled || entry.instance == nil {
		entry.definition = previousDefinition

		return fmt.Errorf("host: cannot replace %q from state %q", definition.ID, entry.state)
	}

	var implicitlyEnabled []string
	if err := manager.enableReplacementDependencies(ctx, definition, &implicitlyEnabled); err != nil {
		manager.abortReplacement(ctx, entry, previousDefinition, implicitlyEnabled)

		return err
	}

	replacement, err := constructReplacement(ctx, definition)
	if err != nil {
		manager.abortReplacement(ctx, entry, previousDefinition, implicitlyEnabled)

		return fmt.Errorf("host: replace %q: %w", definition.ID, err)
	}

	if err := transferReplacementState(ctx, definition.ID, entry.instance, replacement); err != nil {
		manager.abortReplacement(ctx, entry, previousDefinition, implicitlyEnabled)

		return err
	}

	if err := replacement.Start(ctx); err != nil {
		manager.abortStartedReplacement(ctx, entry, previousDefinition, implicitlyEnabled, replacement)

		return fmt.Errorf("host: start replacement %q: %w", definition.ID, err)
	}

	// Retire the old instance only after the restored candidate is ready, preventing replacement downtime.
	if err := entry.instance.Stop(ctx); err != nil {
		manager.abortStartedReplacement(ctx, entry, previousDefinition, implicitlyEnabled, replacement)

		return fmt.Errorf("host: stop replaced component %q: %w", definition.ID, err)
	}

	entry.instance = replacement
	entry.err = nil
	// Enabled-to-enabled is a meaningful diagnostic event: subscribers can observe a successful replacement.
	manager.transition(entry, StateEnabled, nil)

	return nil
}

// prepareReplacement temporarily publishes copied metadata so cycle validation sees the proposed graph.
func (manager *Manager) prepareReplacement(
	definition ManagedDefinition,
) (*managedEntry, ManagedDefinition, error) {
	entry, exists := manager.entries[definition.ID]
	if !exists {
		return nil, ManagedDefinition{}, fmt.Errorf(
			"host: managed component %q is not registered",
			definition.ID,
		)
	}

	if definition.New == nil {
		return nil, ManagedDefinition{}, fmt.Errorf(
			"host: managed component %q has no factory",
			definition.ID,
		)
	}

	// Copy before temporary publication so caller-owned slices cannot later mutate the validated graph.
	definition.DependsOn = append([]string(nil), definition.DependsOn...)
	previousDefinition := entry.definition

	entry.definition = definition
	if err := manager.validateKnownCycles(); err != nil {
		entry.definition = previousDefinition

		return nil, ManagedDefinition{}, err
	}

	return entry, previousDefinition, nil
}

// replacementOnlyUpdatesDefinition identifies inactive entries that need no instance transaction.
func replacementOnlyUpdatesDefinition(entry *managedEntry) bool {
	return entry.state == StateDisabled || entry.state == StateFailed && entry.instance == nil
}

// enableReplacementDependencies establishes the proposed graph while tracking only newly introduced instances.
func (manager *Manager) enableReplacementDependencies(
	ctx context.Context,
	definition ManagedDefinition,
	implicitlyEnabled *[]string,
) error {
	for _, dependency := range definition.DependsOn {
		if err := manager.enable(ctx, dependency, make(map[string]bool), implicitlyEnabled); err != nil {
			return fmt.Errorf(
				"host: replace %q dependency %q: %w",
				definition.ID,
				dependency,
				err,
			)
		}
	}

	return nil
}

// constructReplacement normalizes a nil factory result into the same lifecycle failure as ordinary enable.
func constructReplacement(ctx context.Context, definition ManagedDefinition) (Component, error) {
	replacement, err := definition.New(ctx)
	if err == nil && replacement == nil {
		err = errors.New("factory returned a nil component")
	}

	return replacement, err
}

// transferReplacementState exports and imports before startup so dependents never observe partially restored state.
func transferReplacementState(
	ctx context.Context,
	id string,
	existing Component,
	replacement Component,
) error {
	var state any

	if exporter, ok := existing.(StateExporter); ok {
		exported, err := exporter.ExportState(ctx)
		if err != nil {
			return fmt.Errorf("host: export state for %q: %w", id, err)
		}

		state = exported
	}

	// A nil snapshot means there is intentionally nothing to restore, even if the candidate supports import.
	if importer, ok := replacement.(StateImporter); ok && state != nil {
		if err := importer.ImportState(ctx, state); err != nil {
			return fmt.Errorf("host: import state for %q: %w", id, err)
		}
	}

	return nil
}

// abortReplacement restores definition metadata and removes dependencies introduced solely by this transaction.
func (manager *Manager) abortReplacement(
	ctx context.Context,
	entry *managedEntry,
	previousDefinition ManagedDefinition,
	implicitlyEnabled []string,
) {
	entry.definition = previousDefinition

	manager.rollbackEnabled(ctx, implicitlyEnabled)
}

// abortStartedReplacement stops the unpublished new instance before restoring graph and dependency state.
func (manager *Manager) abortStartedReplacement(
	ctx context.Context,
	entry *managedEntry,
	previousDefinition ManagedDefinition,
	implicitlyEnabled []string,
	replacement Component,
) {
	entry.definition = previousDefinition
	_ = replacement.Stop(ctx)
	manager.rollbackEnabled(ctx, implicitlyEnabled)
}
