package world

// This file contains a reusable fixed-tick velocity integrator. It deliberately
// knows nothing about monsters, AI, Diablo modes, or combat: a mod opts an
// entity in by attaching engine.world.velocity_mover alongside position,
// velocity, and collider components.

import (
	"fmt"
	"math"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

const VelocityMovementSystemID = "engine.world.velocity_movement"

type VelocityPathFinder interface {
	FindPath(PathRequest) ([]Point, error)
}

type velocityStepValidator interface {
	WalkableStep(Point, Point, float64) bool
}

// VelocityComponents names the mod-owned spatial schemas consumed by the
// generic integrator. The engine does not prescribe a mod namespace.
type VelocityComponents struct {
	Position, Velocity, Collider string
}

type velocityStores struct {
	position *akara.DynamicStore
	velocity *akara.DynamicStore
	collider *akara.DynamicStore
}

// RegisterVelocityMovement registers the opt-in marker and fixed-tick movement system against mod-owned spatial
// schemas. Component names remain configurable so the engine does not impose a namespace on game content.
func RegisterVelocityMovement(engine *gameecs.Engine, paths VelocityPathFinder, components VelocityComponents) error {
	if engine == nil {
		return fmt.Errorf("world: velocity movement requires an engine")
	}

	if components.Position == "" || components.Velocity == "" || components.Collider == "" {
		return fmt.Errorf("world: velocity movement requires position, velocity, and collider schemas")
	}

	marker, err := akara.RegisterSchema(engine.World(), akara.Schema{
		Name:    "engine.world.velocity_mover",
		Version: 1,
	})
	if err != nil {
		return err
	}

	position, err := registerVelocityVector(engine, components.Position)
	if err != nil {
		return err
	}

	velocity, err := registerVelocityVector(engine, components.Velocity)
	if err != nil {
		return err
	}

	collider, err := registerVelocityCollider(engine, components.Collider)
	if err != nil {
		return err
	}

	stores := velocityStores{position: position, velocity: velocity, collider: collider}

	return engine.Register(gameecs.Definition{
		ID:     VelocityMovementSystemID,
		Phase:  gameecs.PhaseMovement,
		All:    []akara.ComponentType{marker, position, velocity, collider},
		Read:   []akara.ComponentType{velocity, collider},
		Write:  []akara.ComponentType{position, velocity},
		Update: velocityMovementUpdate(paths, stores),
	})
}

// registerVelocityVector declares the shared two-float shape used by position and velocity. Registering fields in x/y
// order preserves schema identity and deterministic snapshot serialization.
func registerVelocityVector(engine *gameecs.Engine, name string) (*akara.DynamicStore, error) {
	return akara.RegisterSchema(engine.World(), akara.Schema{
		Name:    name,
		Version: 1,
		Fields: []akara.Field{
			{Name: "x", Kind: akara.FieldFloat64},
			{Name: "y", Kind: akara.FieldFloat64},
		},
	})
}

// registerVelocityCollider declares the radius consumed by collision checks. Keeping it separate from vector schemas
// makes the access contract explicit to both Akara and future maintainers.
func registerVelocityCollider(engine *gameecs.Engine, name string) (*akara.DynamicStore, error) {
	return akara.RegisterSchema(engine.World(), akara.Schema{
		Name:    name,
		Version: 1,
		Fields:  []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}},
	})
}

// velocityMovementUpdate adapts the ECS batch callback to one-entity movement. A non-positive delta performs no reads
// or writes, preserving fixed-tick behavior when the engine intentionally executes a zero-duration update.
func velocityMovementUpdate(paths VelocityPathFinder, stores velocityStores) gameecs.UpdateFunc {
	// The closure captures immutable schema stores once, avoiding name-based lookup during every fixed-tick batch.
	return func(context gameecs.Context, entities []akara.Entity, _ *gameecs.StructuralCommands) error {
		seconds := context.Delta.Seconds()
		if seconds <= 0 {
			return nil
		}

		for _, entity := range entities {
			if err := moveVelocityEntity(paths, stores, entity, seconds); err != nil {
				return err
			}
		}

		return nil
	}
}

// moveVelocityEntity integrates one fixed-tick destination, validates collision only when crossing a collision-cell
// boundary, and writes x before y to preserve the component mutation/error order used by existing simulation.
func moveVelocityEntity(paths VelocityPathFinder, stores velocityStores, entity akara.Entity, seconds float64) error {
	position, _ := stores.position.Get(entity)
	velocity, _ := stores.velocity.Get(entity)
	collider, _ := stores.collider.Get(entity)
	xValue, _ := position.Get("x")
	yValue, _ := position.Get("y")
	velocityX, _ := velocity.Get("x")
	velocityY, _ := velocity.Get("y")
	radiusValue, _ := collider.Get("radius")

	x, y := xValue.(float64), yValue.(float64)
	nextX := x + velocityX.(float64)*seconds

	nextY := y + velocityY.(float64)*seconds
	if math.IsNaN(nextX) || math.IsNaN(nextY) || math.IsInf(nextX, 0) || math.IsInf(nextY, 0) {
		return fmt.Errorf("world: velocity movement produced a non-finite position")
	}

	if !velocityDestinationAllowed(
		paths,
		Point{X: x, Y: y},
		Point{X: nextX, Y: nextY},
		radiusValue,
	) {
		// A rejected move consumes velocity so the same blocked request does not repeat on every fixed tick.
		_ = velocity.Set("x", float64(0))
		_ = velocity.Set("y", float64(0))

		return nil
	}

	if err := position.Set("x", nextX); err != nil {
		return err
	}

	return position.Set("y", nextY)
}

// velocityDestinationAllowed skips collision work inside one rounded collision cell. Adjacent transitions use the
// allocation-free validator when available; longer jumps retain full pathfinder semantics.
func velocityDestinationAllowed(paths VelocityPathFinder, start, goal Point, radiusValue any) bool {
	if paths == nil {
		return true
	}

	startCell := navCell{CollisionCell(start.X), CollisionCell(start.Y)}

	goalCell := navCell{CollisionCell(goal.X), CollisionCell(goal.Y)}
	if startCell == goalCell {
		return true
	}

	// Preserve the same-cell fast path before asserting the collider field; no collision consumer needs it in that case.
	radius := radiusValue.(float64)

	dx, dy := absInt(goalCell.x-startCell.x), absInt(goalCell.y-startCell.y)
	if validator, ok := paths.(velocityStepValidator); ok && dx <= 1 && dy <= 1 {
		return validator.WalkableStep(start, goal, radius)
	}

	route, err := paths.FindPath(PathRequest{Start: start, Goal: goal, Radius: radius})

	return err == nil && len(route) > 0
}
