package combat

import (
	"math"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const playerWalkSpeed = 10.0

func attackApproachSchema() akara.Schema {
	return akara.Schema{Name: AttackApproach, Version: 1, Fields: []akara.Field{
		{Name: "skill_id", Kind: akara.FieldInt64}, {Name: "target_id", Kind: akara.FieldString}, {Name: "request_tick", Kind: akara.FieldInt64},
		{Name: "goal_x", Kind: akara.FieldFloat64}, {Name: "goal_y", Kind: akara.FieldFloat64},
		{Name: "waypoint_x", Kind: akara.FieldFloat64}, {Name: "waypoint_y", Kind: akara.FieldFloat64},
		{Name: "has_waypoint", Kind: akara.FieldBool},
	}}
}

func newAttackApproach(skillID int64, targetID string, tick int64) map[string]any {
	return map[string]any{"skill_id": skillID, "target_id": targetID, "request_tick": tick, "goal_x": 0.0, "goal_y": 0.0, "waypoint_x": 0.0, "waypoint_y": 0.0, "has_waypoint": false}
}

// updateAttackApproaches owns the pending click-to-attack action. The target is
// resolved again every tick: disappearing, dying, or changing zone cancels the
// action instead of letting stale presentation state cause damage.
func updateAttackApproaches(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer, paths PathFinder, timings AttackTimingResolver, approaches, attackAnimations, selectables, positions, locations, profiles, velocities, colliders, animations, movementModes, appearances *akara.DynamicStore) error {
	for _, attacker := range entities {
		approach, _ := approaches.Get(attacker)
		targetID, _ := approach.Get("target_id")
		target, found := sameZoneTarget(attacker, targetID.(string), selectables, positions, locations)
		if !found {
			if err := stopAttackApproach(attacker, commands, approaches, velocities, animations); err != nil {
				return err
			}
			continue
		}
		position, _ := positions.Get(attacker)
		profile, _ := profiles.Get(attacker)
		ax, _ := position.Get("x")
		ay, _ := position.Get("y")
		rangeValue, _ := profile.Get("range")
		// Weapon range reaches the edge of a unit, not only its mathematical
		// center. Large monsters therefore become hittable when the attacker
		// reaches their occupied footprint. Pathfinding must use the same number
		// or it can stop successfully and then have combat reject the swing.
		attackRange := rangeValue.(float64) + target.radius
		if math.Hypot(target.x-ax.(float64), target.y-ay.(float64)) <= attackRange {
			if err := stopAttackApproach(attacker, commands, approaches, velocities, animations); err != nil {
				return err
			}
			timing, err := resolveAttackTiming(attacker, timings, appearances)
			if err != nil {
				return err
			}
			skillID, _ := approach.Get("skill_id")
			commands.AddDynamic(attackAnimations, attacker, newAttackAnimation(skillID.(int64), target.id, context.Tick, timing))
			if animation, present := animations.Get(attacker); present {
				if err := animation.Set("mode", "A1"); err != nil {
					return err
				}
				if err := animation.Set("direction", logicalDirection(target.x-ax.(float64), target.y-ay.(float64))); err != nil {
					return err
				}
			}
			continue
		}
		waypoint, ok, err := approachWaypoint(approach, ax.(float64), ay.(float64), target, attackRange, colliderRadius(colliders, attacker), paths)
		if err != nil {
			return err
		}
		if !ok {
			if err := stopAttackApproach(attacker, commands, approaches, velocities, animations); err != nil {
				return err
			}
			continue
		}
		dx, dy := waypoint.X-ax.(float64), waypoint.Y-ay.(float64)
		length := math.Hypot(dx, dy)
		running := false
		if mode, present := movementModes.Get(attacker); present {
			value, _ := mode.Get("running")
			running = value.(bool)
		}
		speed := playerWalkSpeed
		animationMode := "WL"
		if running {
			speed, animationMode = 15, "RN"
		}
		velocity, _ := velocities.Get(attacker)
		if err := setApproachVelocity(velocity, dx/length*speed, dy/length*speed); err != nil {
			return err
		}
		if animation, present := animations.Get(attacker); present {
			if err := animation.Set("mode", animationMode); err != nil {
				return err
			}
			if err := animation.Set("direction", logicalDirection(dx, dy)); err != nil {
				return err
			}
		}
	}
	return nil
}

type approachTarget struct {
	id     string
	x, y   float64
	radius float64
}

func sameZoneTarget(attacker akara.Entity, targetID string, selectables, positions, locations *akara.DynamicStore) (approachTarget, bool) {
	attackerLocation, ok := locations.Get(attacker)
	if !ok {
		return approachTarget{}, false
	}
	aAct, _ := attackerLocation.Get("act")
	aLevel, _ := attackerLocation.Get("level_id")
	for _, entity := range selectables.Entities() {
		selectable, _ := selectables.Get(entity)
		id, _ := selectable.Get("id")
		position, pok := positions.Get(entity)
		location, lok := locations.Get(entity)
		if entity == attacker || id != targetID || !pok || !lok {
			continue
		}
		act, _ := location.Get("act")
		level, _ := location.Get("level_id")
		if act != aAct || level != aLevel {
			return approachTarget{}, false
		}
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		radius, _ := selectable.Get("radius")
		return approachTarget{id: targetID, x: x.(float64), y: y.(float64), radius: radius.(float64)}, true
	}
	return approachTarget{}, false
}

func approachWaypoint(approach *akara.DynamicComponent, x, y float64, target approachTarget, attackRange, radius float64, paths PathFinder) (gameworld.Point, bool, error) {
	has, _ := approach.Get("has_waypoint")
	gx, _ := approach.Get("goal_x")
	gy, _ := approach.Get("goal_y")
	wx, _ := approach.Get("waypoint_x")
	wy, _ := approach.Get("waypoint_y")
	if has.(bool) && gx == target.x && gy == target.y && math.Hypot(wx.(float64)-x, wy.(float64)-y) > .1 {
		return gameworld.Point{X: wx.(float64), Y: wy.(float64)}, true, nil
	}
	waypoint := gameworld.Point{X: target.x, Y: target.y}
	if paths != nil {
		path, err := paths.FindPath(gameworld.PathRequest{Start: gameworld.Point{X: x, Y: y}, Goal: waypoint, Radius: radius, StopRadius: attackRange})
		if err != nil || len(path) < 2 {
			return gameworld.Point{}, false, nil
		}
		waypoint = path[1]
	}
	values := []struct {
		field string
		value any
	}{{"goal_x", target.x}, {"goal_y", target.y}, {"waypoint_x", waypoint.X}, {"waypoint_y", waypoint.Y}, {"has_waypoint", true}}
	for _, item := range values {
		if err := approach.Set(item.field, item.value); err != nil {
			return gameworld.Point{}, false, err
		}
	}
	return waypoint, true, nil
}

func colliderRadius(colliders *akara.DynamicStore, entity akara.Entity) float64 {
	if collider, ok := colliders.Get(entity); ok {
		radius, _ := collider.Get("radius")
		return radius.(float64)
	}
	return 0
}

func stopAttackApproach(entity akara.Entity, commands *akara.CommandBuffer, approaches, velocities, animations *akara.DynamicStore) error {
	if velocity, ok := velocities.Get(entity); ok {
		if err := setApproachVelocity(velocity, 0, 0); err != nil {
			return err
		}
	}
	if animation, ok := animations.Get(entity); ok {
		if err := animation.Set("mode", "NU"); err != nil {
			return err
		}
	}
	commands.Remove(approaches, entity)
	return nil
}

func setApproachVelocity(velocity *akara.DynamicComponent, x, y float64) error {
	if err := velocity.Set("x", x); err != nil {
		return err
	}
	return velocity.Set("y", y)
}

func logicalDirection(x, y float64) int64 {
	sx, sy := 0, 0
	if x < 0 {
		sx = -1
	} else if x > 0 {
		sx = 1
	}
	if y < 0 {
		sy = -1
	} else if y > 0 {
		sy = 1
	}
	directions := map[[2]int]int64{{0, 1}: 0, {-1, 0}: 1, {0, -1}: 2, {1, 0}: 3, {1, 1}: 4, {-1, 1}: 5, {-1, -1}: 6, {1, -1}: 7}
	return directions[[2]int{sx, sy}]
}
