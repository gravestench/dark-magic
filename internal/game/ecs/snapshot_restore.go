package ecs

import (
	"fmt"

	"github.com/gravestench/akara"
)

// RestoreSnapshot creates an engine containing the exact runtime-defined world state and simulation tick captured by
// snapshot. Systems are intentionally absent because trusted session composition must register them separately.
func RestoreSnapshot(snapshot Snapshot) (_ *Engine, err error) {
	if snapshot.Version != SnapshotVersion {
		return nil, fmt.Errorf("%w: %d", ErrSnapshotVersion, snapshot.Version)
	}

	engine := New()

	defer func() {
		if err != nil {
			_ = engine.Close()
		}
	}()

	if err := restoreEntityIdentities(engine.world, snapshot.Entities); err != nil {
		return nil, err
	}

	if err := restoreDynamicComponents(engine.world, snapshot.Components); err != nil {
		return nil, err
	}

	engine.tick = snapshot.Tick

	return engine, nil
}

// restoreEntityIdentities recreates entities before components so entity-valued fields and the allocator's next ID
// retain the identities recorded in the snapshot.
func restoreEntityIdentities(world *akara.World, entities []uint64) error {
	for _, id := range entities {
		if err := world.CreateEntityWithID(akara.Entity(id)); err != nil {
			return fmt.Errorf("game ecs: restore entity %d: %w", id, err)
		}
	}

	return nil
}

// restoreDynamicComponents recreates each schema before populating its instances. Preserving snapshot order keeps
// subsequent dynamic-store enumeration and canonical checksums stable.
func restoreDynamicComponents(world *akara.World, components []ComponentSnapshot) error {
	for _, component := range components {
		schema := akara.Schema{Name: component.Name, Version: component.Version}
		for _, field := range component.Fields {
			schema.Fields = append(schema.Fields, akara.Field{Name: field.Name, Kind: field.Kind})
		}

		store, err := akara.RegisterSchema(world, schema)
		if err != nil {
			return fmt.Errorf("game ecs: restore schema %q: %w", component.Name, err)
		}

		if err := restoreComponentInstances(store, schema, component); err != nil {
			return err
		}
	}

	return nil
}

// restoreComponentInstances validates positional values against their schema before mutating the store. A mismatch is
// rejected rather than truncating or inventing values, which protects replay and checksum compatibility.
func restoreComponentInstances(
	store *akara.DynamicStore,
	schema akara.Schema,
	component ComponentSnapshot,
) error {
	for _, instance := range component.Instances {
		if len(instance.Values) != len(schema.Fields) {
			return fmt.Errorf(
				"game ecs: restore %q entity %d: %d values for %d fields",
				component.Name,
				instance.Entity,
				len(instance.Values),
				len(schema.Fields),
			)
		}

		values := make(map[string]any, len(schema.Fields))
		for index, field := range schema.Fields {
			value, err := restoreValue(field.Kind, instance.Values[index])
			if err != nil {
				return fmt.Errorf(
					"game ecs: restore %q.%s entity %d: %w",
					component.Name,
					field.Name,
					instance.Entity,
					err,
				)
			}

			values[field.Name] = value
		}

		if _, err := store.Set(akara.Entity(instance.Entity), values); err != nil {
			return fmt.Errorf("game ecs: restore %q entity %d: %w", component.Name, instance.Entity, err)
		}
	}

	return nil
}

// Restore replaces the engine's world with a snapshot while preserving its registered systems and fixed-step
// configuration. Subscriptions are rebuilt because each subscription belongs to exactly one Akara world.
func (engine *Engine) Restore(snapshot Snapshot) error {
	restored, err := RestoreSnapshot(snapshot)
	if err != nil {
		return err
	}

	// Match Close's lock order so restoration cannot deadlock with teardown or race a running system callback.
	engine.updateMu.Lock()
	defer engine.updateMu.Unlock()

	engine.mu.Lock()
	defer engine.mu.Unlock()

	candidate := make(map[string]*registeredSystem, len(engine.systems))
	for id, current := range engine.systems {
		definition := cloneDefinition(current.definition)

		definition, remapErr := remapDynamicDefinition(definition, restored.world)
		if remapErr != nil {
			_ = restored.Close()

			return fmt.Errorf("game ecs: restore system %q: %w", id, remapErr)
		}

		options := filterOptions(definition)

		subscription, subscribeErr := restored.world.Subscribe(options...)
		if subscribeErr != nil {
			_ = restored.Close()
			for _, registered := range candidate {
				_ = registered.subscription.Close()
			}

			return fmt.Errorf("game ecs: restore subscription %q: %w", id, subscribeErr)
		}

		candidate[id] = &registeredSystem{definition: definition, subscription: subscription}
	}

	order, err := compileOrder(candidate)
	if err != nil {
		_ = restored.Close()

		return fmt.Errorf("game ecs: restore system order: %w", err)
	}

	oldWorld := engine.world
	for _, registered := range engine.systems {
		_ = registered.subscription.Close()
	}

	// Publish every replacement field while holding mu so readers see either the old world or the complete restored
	// world, never a mixture of the two.
	engine.world = restored.world
	engine.systems = candidate
	engine.order = order
	engine.schedule.Store(&compiledSchedule{systems: order})
	engine.tick = snapshot.Tick

	// Transfer world ownership before closing the temporary engine; otherwise Close would destroy the adopted world.
	restored.world = akara.NewWorld()
	_ = restored.Close()

	return oldWorld.Close()
}

// remapDynamicDefinition replaces component handles from the old world with same-named stores in the restored world.
// Static component types cannot be reconstructed from snapshots and are rejected before the live engine is mutated.
func remapDynamicDefinition(definition Definition, world *akara.World) (Definition, error) {
	var err error

	if definition.All, err = remapDynamicComponents(definition.All, world); err != nil {
		return Definition{}, err
	}

	if definition.Any, err = remapDynamicComponents(definition.Any, world); err != nil {
		return Definition{}, err
	}

	if definition.None, err = remapDynamicComponents(definition.None, world); err != nil {
		return Definition{}, err
	}

	if definition.Read, err = remapDynamicComponents(definition.Read, world); err != nil {
		return Definition{}, err
	}

	if definition.Write, err = remapDynamicComponents(definition.Write, world); err != nil {
		return Definition{}, err
	}

	return definition, nil
}

// remapDynamicComponents resolves component stores by schema name, preserving filter order because definitions expose
// these slices to scheduling and diagnostic code.
func remapDynamicComponents(source []akara.ComponentType, world *akara.World) ([]akara.ComponentType, error) {
	result := make([]akara.ComponentType, 0, len(source))

	for _, component := range source {
		store, dynamic := component.(*akara.DynamicStore)
		if !dynamic {
			return nil, fmt.Errorf("snapshot rollback cannot restore non-dynamic component filter %T", component)
		}

		restored, found := akara.GetDynamicStore(world, store.Schema().Name)
		if !found {
			return nil, fmt.Errorf("restored component %q is missing", store.Schema().Name)
		}

		result = append(result, restored)
	}

	return result, nil
}
