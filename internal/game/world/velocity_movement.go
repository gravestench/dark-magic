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

func RegisterVelocityMovement(engine *gameecs.Engine, paths VelocityPathFinder, components VelocityComponents) error {
	if engine == nil {
		return fmt.Errorf("world: velocity movement requires an engine")
	}
	if components.Position == "" || components.Velocity == "" || components.Collider == "" {
		return fmt.Errorf("world: velocity movement requires position, velocity, and collider schemas")
	}
	marker, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "engine.world.velocity_mover", Version: 1})
	if err != nil {
		return err
	}
	position, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: components.Position, Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		return err
	}
	velocity, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: components.Velocity, Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		return err
	}
	collider, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: components.Collider, Version: 1, Fields: []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}}})
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{ID: VelocityMovementSystemID, Phase: gameecs.PhaseMovement,
		All: []akara.ComponentType{marker, position, velocity, collider}, Read: []akara.ComponentType{velocity, collider}, Write: []akara.ComponentType{position, velocity},
		Update: func(context gameecs.Context, entities []akara.Entity, _ *gameecs.StructuralCommands) error {
			seconds := context.Delta.Seconds()
			if seconds <= 0 {
				return nil
			}
			for _, entity := range entities {
				p, _ := position.Get(entity)
				v, _ := velocity.Get(entity)
				c, _ := collider.Get(entity)
				xv, _ := p.Get("x")
				yv, _ := p.Get("y")
				vxv, _ := v.Get("x")
				vyv, _ := v.Get("y")
				rv, _ := c.Get("radius")
				x, y := xv.(float64), yv.(float64)
				nx, ny := x+vxv.(float64)*seconds, y+vyv.(float64)*seconds
				if math.IsNaN(nx) || math.IsNaN(ny) || math.IsInf(nx, 0) || math.IsInf(ny, 0) {
					return fmt.Errorf("world: velocity movement produced a non-finite position")
				}
				// Collision is cell-based. Movement inside the current collision
				// cell cannot enter a new blocked footprint, so avoid allocating a
				// complete A* search for every small fixed-tick velocity step.
				startCellX, startCellY := CollisionCell(x), CollisionCell(y)
				goalCellX, goalCellY := CollisionCell(nx), CollisionCell(ny)
				if paths != nil && (startCellX != goalCellX || startCellY != goalCellY) {
					allowed := false
					dx, dy := absInt(goalCellX-startCellX), absInt(goalCellY-startCellY)
					if validator, ok := paths.(velocityStepValidator); ok && dx <= 1 && dy <= 1 {
						allowed = validator.WalkableStep(Point{X: x, Y: y}, Point{X: nx, Y: ny}, rv.(float64))
					} else {
						route, e := paths.FindPath(PathRequest{Start: Point{X: x, Y: y}, Goal: Point{X: nx, Y: ny}, Radius: rv.(float64)})
						allowed = e == nil && len(route) > 0
					}
					if !allowed {
						_ = v.Set("x", float64(0))
						_ = v.Set("y", float64(0))
						continue
					}
				}
				if err := p.Set("x", nx); err != nil {
					return err
				}
				if err := p.Set("y", ny); err != nil {
					return err
				}
			}
			return nil
		}})
}
