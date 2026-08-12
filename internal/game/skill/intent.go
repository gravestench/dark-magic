// Package skill owns authoritative cast preparation and execution state.
package skill

import (
	"fmt"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

const (
	IntentConsumerSystemID = "skill.consume_player_intent"
	CastRequestComponent   = "d2legacy.skill.cast_request"
)

func castRequestSchema() akara.Schema {
	return akara.Schema{Name: CastRequestComponent, Version: 1, Fields: []akara.Field{
		{Name: "player", Kind: akara.FieldString}, {Name: "side", Kind: akara.FieldString},
		{Name: "skill_id", Kind: akara.FieldInt64}, {Name: "skill_level", Kind: akara.FieldInt64},
		{Name: "target_x", Kind: akara.FieldFloat64}, {Name: "target_y", Kind: akara.FieldFloat64},
		{Name: "target_id", Kind: akara.FieldString}, {Name: "request_tick", Kind: akara.FieldInt64},
	}}
}

// RegisterIntentConsumer turns the mutable input mailbox into one immutable
// cast snapshot. Later assignment or learned-level changes cannot rewrite it.
func RegisterIntentConsumer(engine *gameecs.Engine) error {
	if engine == nil {
		return fmt.Errorf("skill: intent consumer requires an engine")
	}
	intents, controls, learned, requests, err := registerIntentStores(engine)
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: IntentConsumerSystemID, Phase: gameecs.PhaseIntent,
		All:   []akara.ComponentType{intents, controls},
		Read:  []akara.ComponentType{intents, controls, learned, requests},
		Write: []akara.ComponentType{intents, requests},
		Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			for _, owner := range entities {
				intent, _ := intents.Get(owner)
				sideValue, _ := intent.Get("side")
				side := sideValue.(string)
				if side == "" {
					continue
				}
				if side != "left" && side != "right" {
					return fmt.Errorf("skill: invalid admitted side %q", side)
				}
				if requests.Has(owner) {
					continue
				}
				skillValue, _ := intent.Get("skill_id")
				skillID := skillValue.(int64)
				level, allowed, err := learnedLevel(learned, owner, skillID, side)
				if err != nil {
					return err
				}
				if !allowed {
					return fmt.Errorf("skill: assigned %s skill %d is not learned or allowed", side, skillID)
				}
				control, _ := controls.Get(owner)
				player, _ := control.Get("player")
				targetX, _ := intent.Get("target_x")
				targetY, _ := intent.Get("target_y")
				targetID, _ := intent.Get("target_id")
				commands.AddDynamic(requests, owner, map[string]any{
					"player": player.(string), "side": side, "skill_id": skillID, "skill_level": level,
					"target_x": targetX.(float64), "target_y": targetY.(float64), "target_id": targetID.(string),
					"request_tick": int64(context.Tick),
				})
				if err := clearIntent(intent, targetX.(float64), targetY.(float64)); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

func registerIntentStores(engine *gameecs.Engine) (intents, controls, learned, requests *akara.DynamicStore, err error) {
	schemas := []akara.Schema{
		{Name: "d2legacy.player.skill_intent", Version: 1, Fields: []akara.Field{{Name: "side", Kind: akara.FieldString}, {Name: "skill_id", Kind: akara.FieldInt64}, {Name: "target_x", Kind: akara.FieldFloat64}, {Name: "target_y", Kind: akara.FieldFloat64}, {Name: "target_id", Kind: akara.FieldString}}},
		{Name: "d2legacy.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}},
		{Name: "d2legacy.player.learned_skill", Version: 1, Fields: []akara.Field{{Name: "owner", Kind: akara.FieldEntity}, {Name: "skill_id", Kind: akara.FieldInt64}, {Name: "level", Kind: akara.FieldInt64}, {Name: "list_row", Kind: akara.FieldInt64}, {Name: "left_allowed", Kind: akara.FieldBool}, {Name: "right_allowed", Kind: akara.FieldBool}}},
		castRequestSchema(),
	}
	stores := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		stores[index], err = akara.RegisterSchema(engine.World(), schema)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}
	return stores[0], stores[1], stores[2], stores[3], nil
}

func learnedLevel(learned *akara.DynamicStore, owner akara.Entity, skillID int64, side string) (int64, bool, error) {
	for _, entity := range learned.Entities() {
		component, _ := learned.Get(entity)
		candidateOwner, _ := component.Get("owner")
		candidateID, _ := component.Get("skill_id")
		if candidateOwner != owner || candidateID != skillID {
			continue
		}
		allowedValue, err := component.Get(side + "_allowed")
		if err != nil {
			return 0, false, err
		}
		levelValue, err := component.Get("level")
		if err != nil {
			return 0, false, err
		}
		level := levelValue.(int64)
		return level, allowedValue.(bool) && level > 0, nil
	}
	return 0, false, nil
}

func clearIntent(intent *akara.DynamicComponent, x, y float64) error {
	values := map[string]any{"side": "", "skill_id": int64(0), "target_x": x, "target_y": y, "target_id": ""}
	for field, value := range values {
		if err := intent.Set(field, value); err != nil {
			return err
		}
	}
	return nil
}
