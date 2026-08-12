package combat

import (
	"fmt"

	"github.com/gravestench/akara"
	gameaction "github.com/gravestench/dark-magic/internal/game/action"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const AttackApproach = gameaction.AttackApproachComponent

// PathFinder is the narrow collision-aware route service needed by an attack
// approach. world.Map satisfies it; tests may omit it to exercise open ground.
type PathFinder interface {
	FindPath(gameworld.PathRequest) ([]gameworld.Point, error)
}

// RegisterPlayerBasicAttack converts the reviewed general Attack skill effect
// into the same melee request consumed for monsters. It never applies damage;
// the combat-phase transaction still revalidates target, range, hit, and life.
func RegisterPlayerBasicAttack(engine *gameecs.Engine, skillID int64, paths PathFinder, timings AttackTimingResolver) error {
	if engine == nil || skillID < 0 {
		return fmt.Errorf("combat: player basic attack requires an engine and skill ID")
	}
	requests, _, selectables, positions, locations, profiles, _, _, err := registerMeleeStores(engine)
	if err != nil {
		return err
	}
	casts, receipts, controls, approaches, attackAnimations, velocities, colliders, animations, movementModes, appearances, err := registerPlayerAttackStores(engine)
	if err != nil {
		return err
	}
	if err := engine.Register(gameecs.Definition{
		ID: "combat.player_basic_attack", Phase: gameecs.PhasePreSimulate, After: []string{gameskill.CastLifecycleSystemID},
		All: []akara.ComponentType{casts}, None: []akara.ComponentType{receipts}, Read: []akara.ComponentType{casts, controls, approaches, attackAnimations, positions, appearances}, Write: []akara.ComponentType{receipts, approaches, attackAnimations, velocities, animations},
		Update: func(_ gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			return translatePlayerAttacks(entities, commands, skillID, timings, casts, receipts, controls, approaches, attackAnimations, positions, velocities, animations, appearances)
		},
	}); err != nil {
		return err
	}
	if err := engine.Register(gameecs.Definition{
		ID: "combat.player_attack_approach", Phase: gameecs.PhasePreSimulate, After: []string{"combat.player_basic_attack"},
		All:   []akara.ComponentType{approaches, positions, locations, profiles, velocities},
		Read:  []akara.ComponentType{approaches, positions, locations, profiles, selectables, colliders, movementModes, appearances},
		Write: []akara.ComponentType{approaches, attackAnimations, velocities, animations},
		Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			return updateAttackApproaches(context, entities, commands, paths, timings, approaches, attackAnimations, selectables, positions, locations, profiles, velocities, colliders, animations, movementModes, appearances)
		},
	}); err != nil {
		return err
	}
	return registerAttackAnimationSystem(engine, requests, attackAnimations, animations)
}

func registerPlayerAttackStores(engine *gameecs.Engine) (casts, receipts, controls, approaches, attackAnimations, velocities, colliders, animations, movementModes, appearances *akara.DynamicStore, err error) {
	casts, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: gameskill.CastEventComponent, Version: 1, Fields: []akara.Field{
		{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64}, {Name: "player", Kind: akara.FieldString},
		{Name: "skill_id", Kind: akara.FieldInt64}, {Name: "skill_level", Kind: akara.FieldInt64}, {Name: "behavior", Kind: akara.FieldString},
		{Name: "target_x", Kind: akara.FieldFloat64}, {Name: "target_y", Kind: akara.FieldFloat64}, {Name: "target_id", Kind: akara.FieldString}, {Name: "reason", Kind: akara.FieldString},
	}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	receipts, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: BasicAttackReceipt, Version: 1, Fields: []akara.Field{{Name: "processed", Kind: akara.FieldBool}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	controls, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	approaches, err = akara.RegisterSchema(engine.World(), attackApproachSchema())
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	attackAnimations, err = akara.RegisterSchema(engine.World(), attackAnimationSchema())
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	velocities, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.velocity", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	colliders, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.world.collider", Version: 1, Fields: []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	animations, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.player.animation", Version: 1, Fields: []akara.Field{{Name: "direction", Kind: akara.FieldInt64}, {Name: "mode", Kind: akara.FieldString}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	movementModes, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.player.movement_mode", Version: 1, Fields: []akara.Field{{Name: "running", Kind: akara.FieldBool}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	appearances, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2legacy.player.appearance", Version: 1, Fields: []akara.Field{{Name: "cof", Kind: akara.FieldString}, {Name: "token", Kind: akara.FieldString}, {Name: "palette", Kind: akara.FieldString}, {Name: "weapon_class", Kind: akara.FieldString}}})
	return casts, receipts, controls, approaches, attackAnimations, velocities, colliders, animations, movementModes, appearances, err
}

func translatePlayerAttacks(entities []akara.Entity, commands *akara.CommandBuffer, skillID int64, timings AttackTimingResolver, casts, receipts, controls, approaches, attackAnimations, positions, velocities, animations, appearances *akara.DynamicStore) error {
	for _, eventEntity := range entities {
		event, _ := casts.Get(eventEntity)
		kind, _ := event.Get("kind")
		behavior, _ := event.Get("behavior")
		candidateSkill, _ := event.Get("skill_id")
		if kind != gameskill.EventSkillEffect || behavior != gameskill.BehaviorBasicMelee || candidateSkill != skillID {
			continue
		}
		player, _ := event.Get("player")
		targetID, _ := event.Get("target_id")
		caster, found := controlledPlayer(controls, player.(string))
		if !found {
			return fmt.Errorf("combat: basic-attack caster %q does not exist", player.(string))
		}
		tick, _ := event.Get("tick")
		commands.AddDynamic(receipts, eventEntity, map[string]any{"processed": true})
		requestedTarget := targetID.(string)
		if pending, exists := approaches.Get(caster); exists {
			currentTarget, _ := pending.Get("target_id")
			currentSkill, _ := pending.Get("skill_id")
			if currentSkill == skillID && currentTarget == requestedTarget {
				// Repeated clicks on the same unit reaffirm the action. They must
				// not clear its cached waypoint or postpone reaching attack range.
				continue
			}
		}
		if active, exists := attackAnimations.Get(caster); exists {
			currentTarget, _ := active.Get("target_id")
			currentSkill, _ := active.Get("skill_id")
			if currentSkill == skillID && currentTarget == requestedTarget {
				// The authored swing and impact clock already own this target.
				// Spam clicking cannot restart the animation before its hit frame.
				continue
			}
			commands.Remove(attackAnimations, caster)
		}
		if requestedTarget == "" {
			// A targetless basic attack is the authoritative form of Shift-melee:
			// swing in place toward the pointer and never create a chase route.
			timing, err := resolveAttackTiming(caster, timings, appearances)
			if err != nil {
				return err
			}
			commands.AddDynamic(attackAnimations, caster, newAttackAnimation(skillID, "", uint64(tick.(int64)), timing))
			if velocity, present := velocities.Get(caster); present {
				if err := setApproachVelocity(velocity, 0, 0); err != nil {
					return err
				}
			}
			if animation, present := animations.Get(caster); present {
				if err := animation.Set("mode", "A1"); err != nil {
					return err
				}
				if position, found := positions.Get(caster); found {
					x, _ := position.Get("x")
					y, _ := position.Get("y")
					targetX, _ := event.Get("target_x")
					targetY, _ := event.Get("target_y")
					if err := animation.Set("direction", logicalDirection(targetX.(float64)-x.(float64), targetY.(float64)-y.(float64))); err != nil {
						return err
					}
				}
			}
			continue
		}
		values := newAttackApproach(skillID, requestedTarget, tick.(int64))
		if pending, exists := approaches.Get(caster); exists {
			// A fresh click replaces the old pending target immediately. There is
			// only one left-button basic action per controlled player.
			for _, field := range []string{"skill_id", "target_id", "request_tick", "goal_x", "goal_y", "waypoint_x", "waypoint_y", "has_waypoint"} {
				if err := pending.Set(field, values[field]); err != nil {
					return err
				}
			}
		} else {
			commands.AddDynamic(approaches, caster, values)
		}
	}
	return nil
}

func controlledPlayer(controls *akara.DynamicStore, player string) (akara.Entity, bool) {
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		candidate, _ := control.Get("player")
		if candidate == player {
			return entity, true
		}
	}
	return 0, false
}
