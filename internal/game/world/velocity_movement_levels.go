package world

import (
	"fmt"
	"reflect"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

// VelocityLevelLocation names the authoritative integer level field on a mover.
type VelocityLevelLocation struct{ Component, Field string }

type velocityLevelSelection struct {
	location VelocityLevelLocation
	paths    map[int64]VelocityPathFinder
}

// RegisterVelocityMovementByLevel registers the same opt-in integrator with collision selected from each
// mover's current level. The lookup is copied; its pathfinders must remain immutable during simulation.
// Missing location or geometry fails the tick, including moves that stay inside one collision cell.
func RegisterVelocityMovementByLevel(engine *gameecs.Engine, levels map[int64]VelocityPathFinder, components VelocityComponents, location VelocityLevelLocation) error {
	if location.Component == "" || location.Field == "" {
		return fmt.Errorf("world: velocity movement requires a location schema and level field")
	}
	selection := &velocityLevelSelection{location: location, paths: make(map[int64]VelocityPathFinder, len(levels))}
	for level, paths := range levels {
		if !nilVelocityPaths(paths) {
			selection.paths[level] = paths
		}
	}
	return registerVelocityMovement(engine, nil, components, selection)
}

// pathsFor reads the current store rather than retaining component values across ticks or world replacement.
func (selection *velocityLevelSelection) pathsFor(store *akara.DynamicStore, entity akara.Entity) (VelocityPathFinder, error) {
	location, found := store.Get(entity)
	if !found {
		return nil, fmt.Errorf("world: velocity movement entity %d missing location component %q", entity, selection.location.Component)
	}
	value, err := location.Get(selection.location.Field)
	if err != nil {
		return nil, fmt.Errorf("world: velocity movement entity %d location field %q: %w", entity, selection.location.Field, err)
	}
	level, ok := value.(int64)
	if !ok || level <= 0 {
		return nil, fmt.Errorf("world: velocity movement entity %d invalid level %v", entity, value)
	}
	paths := selection.paths[level]
	if paths == nil {
		return nil, fmt.Errorf("world: velocity movement entity %d missing collision map for level %d", entity, level)
	}
	return paths, nil
}

// Interfaces can contain a typed nil pointer; reject those just as an absent map before the same-cell bypass.
func nilVelocityPaths(paths VelocityPathFinder) bool {
	if paths == nil {
		return true
	}
	value := reflect.ValueOf(paths)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
