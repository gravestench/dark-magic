package skill

import (
	"fmt"
	"math"
	"sort"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/targeting"
)

const (
	CastLifecycleSystemID = "skill.cast_lifecycle"
	CastStateComponent    = "d2legacy.skill.cast_state"
	CastEventComponent    = "d2legacy.skill.cast_event"

	TargetPoint = "point"
	TargetUnit  = "unit"
	// BehaviorPointEvent is the first generic family: it emits its authoritative
	// effect at a point/semantic target for a later trusted effect consumer.
	BehaviorPointEvent = "basic.point_event"
	// BehaviorStraightMissile emits the same semantic effect event; the missile
	// authority consumes it using a separately verified projectile definition.
	BehaviorStraightMissile = "basic.straight_missile"
	// BehaviorBasicMelee emits a target-unit effect for combat's shared melee
	// transaction. The skill layer schedules the action but never applies damage.
	BehaviorBasicMelee = "basic.melee"

	EventCastStarted     = "cast_started"
	EventSkillEffect     = "skill_effect"
	EventCastCompleted   = "cast_completed"
	EventCastInterrupted = "cast_interrupted"
	EventCastRejected    = "cast_rejected"
)

// Definition is a verified, normalized skill behavior contract. Raw Skills.txt
// rows do not enter simulation until their formulas and server functions have
// been deliberately translated into this representation.
type Definition struct {
	SkillID      int64
	Behavior     string
	TargetPolicy string
	ManaCost     int64
	// ManaCostRaw uses Diablo's 8.8 fixed-point unit (256 == one mana).
	// ManaCost remains a whole-number compatibility input for older fixtures.
	ManaCostRaw   int64
	EffectDelay   uint64
	CompleteDelay uint64
	Interruptible bool
}

// Registry is immutable after construction and rejects ambiguous definitions.
type Registry struct{ definitions map[int64]Definition }

func NewRegistry(definitions ...Definition) (Registry, error) {
	result := Registry{definitions: make(map[int64]Definition, len(definitions))}
	for _, definition := range definitions {
		if definition.ManaCostRaw == 0 && definition.ManaCost > 0 {
			definition.ManaCostRaw = definition.ManaCost * 256
		}
		if definition.SkillID < 0 || !supportedBehavior(definition.Behavior) || definition.ManaCost < 0 || definition.ManaCostRaw < 0 || definition.EffectDelay == 0 || definition.CompleteDelay < definition.EffectDelay || definition.TargetPolicy != TargetPoint && definition.TargetPolicy != TargetUnit {
			return Registry{}, fmt.Errorf("skill: invalid normalized definition for %d", definition.SkillID)
		}
		if _, found := result.definitions[definition.SkillID]; found {
			return Registry{}, fmt.Errorf("skill: duplicate normalized definition %d", definition.SkillID)
		}
		result.definitions[definition.SkillID] = definition
	}
	return result, nil
}

func supportedBehavior(behavior string) bool {
	return behavior == BehaviorPointEvent || behavior == BehaviorStraightMissile || behavior == BehaviorBasicMelee
}

func castStateSchema() akara.Schema {
	return akara.Schema{Name: CastStateComponent, Version: 1, Fields: []akara.Field{
		{Name: "player", Kind: akara.FieldString}, {Name: "side", Kind: akara.FieldString}, {Name: "skill_id", Kind: akara.FieldInt64}, {Name: "skill_level", Kind: akara.FieldInt64},
		{Name: "behavior", Kind: akara.FieldString}, {Name: "target_policy", Kind: akara.FieldString}, {Name: "target_x", Kind: akara.FieldFloat64}, {Name: "target_y", Kind: akara.FieldFloat64}, {Name: "target_id", Kind: akara.FieldString},
		{Name: "start_tick", Kind: akara.FieldInt64}, {Name: "effect_tick", Kind: akara.FieldInt64}, {Name: "complete_tick", Kind: akara.FieldInt64},
		{Name: "phase", Kind: akara.FieldString}, {Name: "interruptible", Kind: akara.FieldBool}, {Name: "interruption_requested", Kind: akara.FieldBool}, {Name: "mana_cost", Kind: akara.FieldInt64}, {Name: "mana_cost_raw", Kind: akara.FieldInt64},
	}}
}

func castEventSchema() akara.Schema {
	return akara.Schema{Name: CastEventComponent, Version: 1, Fields: []akara.Field{
		{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64}, {Name: "player", Kind: akara.FieldString},
		{Name: "skill_id", Kind: akara.FieldInt64}, {Name: "skill_level", Kind: akara.FieldInt64}, {Name: "behavior", Kind: akara.FieldString},
		{Name: "target_x", Kind: akara.FieldFloat64}, {Name: "target_y", Kind: akara.FieldFloat64}, {Name: "target_id", Kind: akara.FieldString},
		{Name: "reason", Kind: akara.FieldString},
	}}
}

// RegisterCastLifecycle installs headless start/effect/complete scheduling.
func RegisterCastLifecycle(engine *gameecs.Engine, registry Registry) error {
	if engine == nil || registry.definitions == nil {
		return fmt.Errorf("skill: cast lifecycle requires an engine and registry")
	}
	requests, states, events, vitals, selectables, err := registerCastStores(engine)
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: CastLifecycleSystemID, Phase: gameecs.PhasePreSimulate,
		Any: []akara.ComponentType{requests, states}, Read: []akara.ComponentType{requests, states, vitals, selectables}, Write: []akara.ComponentType{requests, states, events, vitals},
		Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			for _, owner := range entities {
				if state, present := states.Get(owner); present {
					if err := advanceCast(context, owner, state, commands, engine.World(), states, events); err != nil {
						return err
					}
					continue
				}
				request, present := requests.Get(owner)
				if !present {
					continue
				}
				if err := beginCast(context, owner, request, registry, commands, engine.World(), requests, states, events, vitals, selectables); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

func registerCastStores(engine *gameecs.Engine) (requests, states, events, vitals, selectables *akara.DynamicStore, err error) {
	schemas := []akara.Schema{
		castRequestSchema(), castStateSchema(), castEventSchema(),
		{Name: "d2legacy.player.vitals", Version: 1, Fields: []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "mana", Kind: akara.FieldInt64}, {Name: "max_mana", Kind: akara.FieldInt64}, {Name: "mana_raw", Kind: akara.FieldInt64}, {Name: "max_mana_raw", Kind: akara.FieldInt64}}},
		targeting.Schema(),
	}
	stores := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		stores[index], err = akara.RegisterSchema(engine.World(), schema)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}
	return stores[0], stores[1], stores[2], stores[3], stores[4], nil
}

func beginCast(context gameecs.Context, owner akara.Entity, request *akara.DynamicComponent, registry Registry, commands *akara.CommandBuffer, world *akara.World, requests, states, events, vitals, selectables *akara.DynamicStore) error {
	skillValue, _ := request.Get("skill_id")
	requestTick, _ := request.Get("request_tick")
	if requestTick.(int64) > int64(context.Tick) {
		rejectCast(context, owner, request, "request tick is in the future", commands, world, requests, events)
		return nil
	}
	definition, found := registry.definitions[skillValue.(int64)]
	if !found {
		rejectCast(context, owner, request, fmt.Sprintf("unsupported skill %d", skillValue.(int64)), commands, world, requests, events)
		return nil
	}
	targetX, _ := request.Get("target_x")
	targetY, _ := request.Get("target_y")
	targetID, _ := request.Get("target_id")
	if err := validateTarget(definition, targetX.(float64), targetY.(float64), targetID.(string), selectables); err != nil {
		rejectCast(context, owner, request, err.Error(), commands, world, requests, events)
		return nil
	}
	vital, present := vitals.Get(owner)
	if !present {
		return fmt.Errorf("skill: caster lacks vitals")
	}
	manaValue, _ := vital.Get("mana_raw")
	// Some small tests and tools only populate the whole-mana convenience field.
	// Promote it at the authority boundary so they exercise the same cast path.
	if manaValue.(int64) == 0 {
		wholeMana, _ := vital.Get("mana")
		if wholeMana.(int64) > 0 {
			manaValue = wholeMana.(int64) * 256
			if err := vital.Set("mana_raw", manaValue); err != nil {
				return err
			}
		}
	}
	if manaValue.(int64) < definition.ManaCostRaw {
		rejectCast(context, owner, request, "insufficient mana", commands, world, requests, events)
		return nil
	}
	remainingMana := manaValue.(int64) - definition.ManaCostRaw
	if err := vital.Set("mana_raw", remainingMana); err != nil {
		return err
	}
	if err := vital.Set("mana", remainingMana/256); err != nil {
		return err
	}
	player, _ := request.Get("player")
	side, _ := request.Get("side")
	level, _ := request.Get("skill_level")
	stateValues := map[string]any{
		"player": player.(string), "side": side.(string), "skill_id": definition.SkillID, "skill_level": level.(int64), "behavior": definition.Behavior, "target_policy": definition.TargetPolicy,
		"target_x": targetX.(float64), "target_y": targetY.(float64), "target_id": targetID.(string), "start_tick": int64(context.Tick), "effect_tick": int64(context.Tick + definition.EffectDelay),
		"complete_tick": int64(context.Tick + definition.CompleteDelay), "phase": "started", "interruptible": definition.Interruptible, "interruption_requested": false, "mana_cost": definition.ManaCost, "mana_cost_raw": definition.ManaCostRaw,
	}
	commands.AddDynamic(states, owner, stateValues)
	commands.Remove(requests, owner)
	commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: castEventValues(EventCastStarted, context.Tick, stateValues)})
	return nil
}

func advanceCast(context gameecs.Context, owner akara.Entity, state *akara.DynamicComponent, commands *akara.CommandBuffer, world *akara.World, states, events *akara.DynamicStore) error {
	interruptible, _ := state.Get("interruptible")
	interrupted, _ := state.Get("interruption_requested")
	if interruptible.(bool) && interrupted.(bool) {
		commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: castEventFromState(EventCastInterrupted, context.Tick, state)})
		commands.Remove(states, owner)
		return nil
	}
	phase, _ := state.Get("phase")
	effectTick, _ := state.Get("effect_tick")
	completeTick, _ := state.Get("complete_tick")
	if phase == "started" && int64(context.Tick) >= effectTick.(int64) {
		commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: castEventFromState(EventSkillEffect, context.Tick, state)})
		if err := state.Set("phase", "effect"); err != nil {
			return err
		}
		phase = "effect"
	}
	if phase == "effect" && int64(context.Tick) >= completeTick.(int64) {
		commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: castEventFromState(EventCastCompleted, context.Tick, state)})
		commands.Remove(states, owner)
	}
	return nil
}

func validateTarget(definition Definition, x, y float64, targetID string, selectables *akara.DynamicStore) error {
	if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
		return fmt.Errorf("skill: target point must be finite")
	}
	if definition.TargetPolicy == TargetPoint {
		return nil
	}
	// Diablo II's Shift modifier permits a basic melee swing at a point even
	// without a unit underneath it. This still produces the authored animation
	// and timing, but the later melee transaction has no victim to damage.
	if definition.Behavior == BehaviorBasicMelee && targetID == "" {
		return nil
	}
	if targetID == "" {
		return fmt.Errorf("skill: unit target is required")
	}
	ids := make([]string, 0)
	for _, entity := range selectables.Entities() {
		component, _ := selectables.Get(entity)
		id, _ := component.Get("id")
		ids = append(ids, id.(string))
	}
	sort.Strings(ids)
	for _, id := range ids {
		if id == targetID {
			return nil
		}
	}
	return fmt.Errorf("skill: unit target %q does not exist", targetID)
}

func rejectCast(context gameecs.Context, owner akara.Entity, request *akara.DynamicComponent, reason string, commands *akara.CommandBuffer, world *akara.World, requests, events *akara.DynamicStore) {
	values := map[string]any{"behavior": "", "reason": reason}
	for _, field := range []string{"player", "skill_id", "skill_level", "target_x", "target_y", "target_id"} {
		values[field], _ = request.Get(field)
	}
	commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: castEventValues(EventCastRejected, context.Tick, values)})
	commands.Remove(requests, owner)
}

func castEventValues(kind string, tick uint64, state map[string]any) map[string]any {
	reason, _ := state["reason"]
	if reason == nil {
		reason = ""
	}
	return map[string]any{"kind": kind, "tick": int64(tick), "player": state["player"], "skill_id": state["skill_id"], "skill_level": state["skill_level"], "behavior": state["behavior"], "target_x": state["target_x"], "target_y": state["target_y"], "target_id": state["target_id"], "reason": reason}
}

func castEventFromState(kind string, tick uint64, state *akara.DynamicComponent) map[string]any {
	values := map[string]any{}
	for _, field := range []string{"player", "skill_id", "skill_level", "behavior", "target_x", "target_y", "target_id"} {
		values[field], _ = state.Get(field)
	}
	return castEventValues(kind, tick, values)
}
