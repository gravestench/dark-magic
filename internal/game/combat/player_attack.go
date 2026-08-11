package combat

import (
	"fmt"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const AttackApproach = "dm.combat.attack_approach"

// PathFinder is the narrow collision-aware route service needed by an attack
// approach. world.Map satisfies it; tests may omit it to exercise open ground.
type PathFinder interface {
	FindPath(gameworld.PathRequest) ([]gameworld.Point, error)
}

// RegisterPlayerBasicAttack converts the reviewed general Attack skill effect
// into the same melee request consumed for monsters. It never applies damage;
// the combat-phase transaction still revalidates target, range, hit, and life.
func RegisterPlayerBasicAttack(engine *gameecs.Engine, skillID int64, paths PathFinder) error {
	if engine == nil || skillID < 0 {
		return fmt.Errorf("combat: player basic attack requires an engine and skill ID")
	}
	requests, _, selectables, positions, locations, profiles, _, _, err := registerMeleeStores(engine)
	if err != nil {
		return err
	}
	casts, receipts, controls, approaches, velocities, colliders, animations, movementModes, err := registerPlayerAttackStores(engine)
	if err != nil {
		return err
	}
	if err := engine.Register(gameecs.Definition{
		ID: "combat.player_basic_attack", Phase: gameecs.PhasePreSimulate, After: []string{gameskill.CastLifecycleSystemID},
		All: []akara.ComponentType{casts}, None: []akara.ComponentType{receipts}, Read: []akara.ComponentType{casts, controls}, Write: []akara.ComponentType{receipts, approaches},
		Update: func(_ gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			return translatePlayerAttacks(entities, commands, skillID, casts, receipts, controls, approaches)
		},
	}); err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: "combat.player_attack_approach", Phase: gameecs.PhasePreSimulate, After: []string{"combat.player_basic_attack"},
		All:   []akara.ComponentType{approaches, positions, locations, profiles, velocities},
		Read:  []akara.ComponentType{approaches, positions, locations, profiles, selectables, colliders, movementModes},
		Write: []akara.ComponentType{approaches, velocities, animations, requests},
		Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			return updateAttackApproaches(context, entities, commands, paths, approaches, requests, selectables, positions, locations, profiles, velocities, colliders, animations, movementModes)
		},
	})
}

func registerPlayerAttackStores(engine *gameecs.Engine) (casts, receipts, controls, approaches, velocities, colliders, animations, movementModes *akara.DynamicStore, err error) {
	casts, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: gameskill.CastEventComponent, Version: 1, Fields: []akara.Field{
		{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64}, {Name: "player", Kind: akara.FieldString},
		{Name: "skill_id", Kind: akara.FieldInt64}, {Name: "skill_level", Kind: akara.FieldInt64}, {Name: "behavior", Kind: akara.FieldString},
		{Name: "target_x", Kind: akara.FieldFloat64}, {Name: "target_y", Kind: akara.FieldFloat64}, {Name: "target_id", Kind: akara.FieldString}, {Name: "reason", Kind: akara.FieldString},
	}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	receipts, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: BasicAttackReceipt, Version: 1, Fields: []akara.Field{{Name: "processed", Kind: akara.FieldBool}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	controls, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	approaches, err = akara.RegisterSchema(engine.World(), attackApproachSchema())
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	velocities, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.world.velocity", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	colliders, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.world.collider", Version: 1, Fields: []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	animations, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.player.animation", Version: 1, Fields: []akara.Field{{Name: "direction", Kind: akara.FieldInt64}, {Name: "mode", Kind: akara.FieldString}}})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, nil, err
	}
	movementModes, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.player.movement_mode", Version: 1, Fields: []akara.Field{{Name: "running", Kind: akara.FieldBool}}})
	return casts, receipts, controls, approaches, velocities, colliders, animations, movementModes, err
}

func translatePlayerAttacks(entities []akara.Entity, commands *akara.CommandBuffer, skillID int64, casts, receipts, controls, approaches *akara.DynamicStore) error {
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
		values := newAttackApproach(targetID.(string), tick.(int64))
		if pending, exists := approaches.Get(caster); exists {
			// A fresh click replaces the old pending target immediately. There is
			// only one left-button basic action per controlled player.
			for _, field := range []string{"target_id", "request_tick", "goal_x", "goal_y", "waypoint_x", "waypoint_y", "has_waypoint"} {
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
