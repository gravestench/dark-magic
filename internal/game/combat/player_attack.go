package combat

import (
	"fmt"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
)

// RegisterPlayerBasicAttack converts the reviewed general Attack skill effect
// into the same melee request consumed for monsters. It never applies damage;
// the combat-phase transaction still revalidates target, range, hit, and life.
func RegisterPlayerBasicAttack(engine *gameecs.Engine, skillID int64) error {
	if engine == nil || skillID < 0 {
		return fmt.Errorf("combat: player basic attack requires an engine and skill ID")
	}
	requests, _, _, _, _, _, _, _, err := registerMeleeStores(engine)
	if err != nil {
		return err
	}
	casts, receipts, controls, err := registerPlayerAttackStores(engine)
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: "combat.player_basic_attack", Phase: gameecs.PhasePreSimulate, After: []string{gameskill.CastLifecycleSystemID},
		All: []akara.ComponentType{casts}, None: []akara.ComponentType{receipts}, Read: []akara.ComponentType{casts, controls}, Write: []akara.ComponentType{receipts, requests},
		Update: func(_ gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			return translatePlayerAttacks(entities, commands, skillID, casts, receipts, controls, requests)
		},
	})
}

func registerPlayerAttackStores(engine *gameecs.Engine) (casts, receipts, controls *akara.DynamicStore, err error) {
	casts, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: gameskill.CastEventComponent, Version: 1, Fields: []akara.Field{
		{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64}, {Name: "player", Kind: akara.FieldString},
		{Name: "skill_id", Kind: akara.FieldInt64}, {Name: "skill_level", Kind: akara.FieldInt64}, {Name: "behavior", Kind: akara.FieldString},
		{Name: "target_x", Kind: akara.FieldFloat64}, {Name: "target_y", Kind: akara.FieldFloat64}, {Name: "target_id", Kind: akara.FieldString}, {Name: "reason", Kind: akara.FieldString},
	}})
	if err != nil {
		return nil, nil, nil, err
	}
	receipts, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: BasicAttackReceipt, Version: 1, Fields: []akara.Field{{Name: "processed", Kind: akara.FieldBool}}})
	if err != nil {
		return nil, nil, nil, err
	}
	controls, err = akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	return casts, receipts, controls, err
}

func translatePlayerAttacks(entities []akara.Entity, commands *akara.CommandBuffer, skillID int64, casts, receipts, controls, requests *akara.DynamicStore) error {
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
		commands.AddDynamic(requests, caster, map[string]any{"target_id": targetID.(string), "request_tick": tick.(int64)})
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
