package monster

import (
	"fmt"
	"math"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const MovementSystemID = "monster.velocity_movement"

// RegisterMovement applies AI-owned velocity during the movement phase. The
// first policy treats authored velocity as subtiles per second and validates
// each fixed-step destination against the same map used for path selection.
// Dynamic unit separation and special movement modes remain later policies.
func RegisterMovement(engine *gameecs.Engine, paths PathFinder) error {
	if engine == nil {
		return fmt.Errorf("monster: movement requires an engine")
	}
	ai, _, positions, velocities, _, _, err := registerAIStores(engine)
	if err != nil {
		return err
	}
	colliders, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.collider", Version: 1, Fields: []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}}})
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: MovementSystemID, Phase: gameecs.PhaseMovement,
		All:  []akara.ComponentType{ai, positions, velocities, colliders},
		Read: []akara.ComponentType{velocities, colliders}, Write: []akara.ComponentType{positions, velocities},
		Update: func(context gameecs.Context, entities []akara.Entity, _ *akara.CommandBuffer) error {
			return moveMonsters(context, entities, paths, positions, velocities, colliders)
		},
	})
}

func moveMonsters(context gameecs.Context, entities []akara.Entity, paths PathFinder, positions, velocities, colliders *akara.DynamicStore) error {
	seconds := context.Delta.Seconds()
	if seconds <= 0 {
		return nil
	}
	for _, entity := range entities {
		position, _ := positions.Get(entity)
		velocity, _ := velocities.Get(entity)
		collider, _ := colliders.Get(entity)
		xValue, _ := position.Get("x")
		yValue, _ := position.Get("y")
		vxValue, _ := velocity.Get("x")
		vyValue, _ := velocity.Get("y")
		radiusValue, _ := collider.Get("radius")
		x, y := xValue.(float64), yValue.(float64)
		nextX, nextY := x+vxValue.(float64)*seconds, y+vyValue.(float64)*seconds
		if !finiteMovement(nextX, nextY) {
			return fmt.Errorf("monster: movement produced a non-finite position")
		}
		if paths != nil {
			path, err := paths.FindPath(gameworld.PathRequest{Start: gameworld.Point{X: x, Y: y}, Goal: gameworld.Point{X: nextX, Y: nextY}, Radius: radiusValue.(float64)})
			if err != nil || len(path) == 0 {
				if err := setVelocity(velocity, 0, 0); err != nil {
					return err
				}
				continue
			}
		}
		if err := position.Set("x", nextX); err != nil {
			return err
		}
		if err := position.Set("y", nextY); err != nil {
			return err
		}
	}
	return nil
}

func finiteMovement(x, y float64) bool {
	return !math.IsNaN(x) && !math.IsNaN(y) && !math.IsInf(x, 0) && !math.IsInf(y, 0)
}
