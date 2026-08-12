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

func RegisterVelocityMovement(engine *gameecs.Engine, paths VelocityPathFinder) error {
	if engine == nil {
		return fmt.Errorf("world: velocity movement requires an engine")
	}
	marker, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "engine.world.velocity_mover", Version: 1})
	if err != nil {
		return err
	}
	position, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		return err
	}
	velocity, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.velocity", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		return err
	}
	collider, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.collider", Version: 1, Fields: []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}}})
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{ID: VelocityMovementSystemID, Phase: gameecs.PhaseMovement,
		All: []akara.ComponentType{marker, position, velocity, collider}, Read: []akara.ComponentType{velocity, collider}, Write: []akara.ComponentType{position, velocity},
		Update: func(context gameecs.Context, entities []akara.Entity, _ *akara.CommandBuffer) error {
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
				if paths != nil {
					route, e := paths.FindPath(PathRequest{Start: Point{X: x, Y: y}, Goal: Point{X: nx, Y: ny}, Radius: rv.(float64)})
					if e != nil || len(route) == 0 {
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
