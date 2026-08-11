package combat

import (
	"fmt"
	"hash/fnv"
	"math"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/targeting"
)

const (
	BasicMeleeSystemID = "combat.basic_melee"
	BasicAttackRequest = "dm.monster.basic_attack_request"
	CombatEvent        = "dm.combat.event"

	EventAttackAttempted = "attack_attempted"
	EventHitResolved     = "hit_resolved"
	EventDamageApplied   = "damage_applied"
	EventUnitDied        = "unit_died"
)

// BasicMeleePolicy is intentionally Dark Magic policy, not a claim about the
// still-unverified legacy chance-to-hit formula. The probability is explicit
// so the transaction can prove both hit and miss paths with synthetic vectors.
type BasicMeleePolicy struct {
	HitChance int
}

func (policy BasicMeleePolicy) validate() error {
	if policy.HitChance < 0 || policy.HitChance > 100 {
		return fmt.Errorf("combat: hit chance must be in [0,100]")
	}
	return nil
}

func eventSchema() akara.Schema {
	return akara.Schema{Name: CombatEvent, Version: 2, Fields: []akara.Field{
		{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64},
		{Name: "attacker_id", Kind: akara.FieldString}, {Name: "target_id", Kind: akara.FieldString},
		{Name: "hit", Kind: akara.FieldBool}, {Name: "physical", Kind: akara.FieldInt64},
		{Name: "damage_channel", Kind: akara.FieldString}, {Name: "damage", Kind: akara.FieldInt64},
		{Name: "remaining_health", Kind: akara.FieldInt64},
	}}
}

// RegisterBasicMelee consumes semantic attack requests during the combat phase.
// AI chooses intent; this system alone validates and mutates health.
func RegisterBasicMelee(engine *gameecs.Engine, policy BasicMeleePolicy) error {
	if engine == nil {
		return fmt.Errorf("combat: melee requires an engine")
	}
	if err := policy.validate(); err != nil {
		return err
	}
	requests, events, selectables, positions, locations, monsterStats, playerVitals, err := registerMeleeStores(engine)
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: BasicMeleeSystemID, Phase: gameecs.PhaseCombat,
		All:   []akara.ComponentType{requests, selectables, positions, locations, monsterStats},
		Read:  []akara.ComponentType{requests, selectables, positions, locations},
		Write: []akara.ComponentType{requests, events, monsterStats, playerVitals},
		Update: func(context gameecs.Context, entities []akara.Entity, commands *akara.CommandBuffer) error {
			return resolveMelee(context, entities, commands, engine.World(), policy, requests, events, selectables, positions, locations, monsterStats, playerVitals)
		},
	})
}

func registerMeleeStores(engine *gameecs.Engine) (requests, events, selectables, positions, locations, monsterStats, playerVitals *akara.DynamicStore, err error) {
	schemas := []akara.Schema{
		{Name: BasicAttackRequest, Version: 1, Fields: []akara.Field{{Name: "target_id", Kind: akara.FieldString}, {Name: "request_tick", Kind: akara.FieldInt64}, {Name: "range", Kind: akara.FieldFloat64}}},
		eventSchema(), targeting.Schema(),
		{Name: "dm.world.position", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.world.location", Version: 1, Fields: []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}}},
		{Name: "dm.monster.stats", Version: 1, Fields: []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "defense", Kind: akara.FieldInt64}, {Name: "attack_rating", Kind: akara.FieldInt64}, {Name: "physical_min", Kind: akara.FieldInt64}, {Name: "physical_max", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}}},
		{Name: "dm.player.vitals", Version: 1, Fields: []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "mana", Kind: akara.FieldInt64}, {Name: "max_mana", Kind: akara.FieldInt64}}},
	}
	stores := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		stores[index], err = akara.RegisterSchema(engine.World(), schema)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
	}
	return stores[0], stores[1], stores[2], stores[3], stores[4], stores[5], stores[6], nil
}

func resolveMelee(context gameecs.Context, attackers []akara.Entity, commands *akara.CommandBuffer, world *akara.World, policy BasicMeleePolicy, requests, events, selectables, positions, locations, monsterStats, playerVitals *akara.DynamicStore) error {
	for _, attacker := range attackers {
		request, _ := requests.Get(attacker)
		targetID, _ := request.Get("target_id")
		requestTick, _ := request.Get("request_tick")
		if requestTick.(int64) > int64(context.Tick) {
			return fmt.Errorf("combat: attack request is from future tick %d", requestTick.(int64))
		}
		attackRange, _ := request.Get("range")
		attackerID, legalTarget, err := legalMeleeTarget(attacker, targetID.(string), attackRange.(float64), selectables, positions, locations)
		commands.Remove(requests, attacker)
		if err != nil {
			return err
		}
		commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: eventValues(EventAttackAttempted, context.Tick, attackerID, targetID.(string), false, 0, 0)})
		if !legalTarget.found {
			commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: eventValues(EventHitResolved, context.Tick, attackerID, targetID.(string), false, 0, 0)})
			continue
		}
		hit := int(stableRoll("hit", context.Tick, attackerID, targetID.(string))%100) < policy.HitChance
		commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: eventValues(EventHitResolved, context.Tick, attackerID, targetID.(string), hit, 0, 0)})
		if !hit {
			continue
		}
		attackerComponent, _ := monsterStats.Get(attacker)
		minimum, _ := attackerComponent.Get("physical_min")
		maximum, _ := attackerComponent.Get("physical_max")
		damage, err := rollDamage(FromRaw(minimum.(int64)), FromRaw(maximum.(int64)), stableRoll("damage", context.Tick, attackerID, targetID.(string)))
		if err != nil {
			return err
		}
		remaining, died, err := applyPhysical(legalTarget.entity, damage, monsterStats, playerVitals)
		if err != nil {
			return err
		}
		commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: eventValues(EventDamageApplied, context.Tick, attackerID, targetID.(string), true, damage.Raw(), remaining.Raw())})
		if died {
			commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{events: eventValues(EventUnitDied, context.Tick, attackerID, targetID.(string), true, damage.Raw(), 0)})
		}
	}
	return nil
}

type targetResult struct {
	entity akara.Entity
	found  bool
}

func legalMeleeTarget(attacker akara.Entity, targetID string, attackRange float64, selectables, positions, locations *akara.DynamicStore) (string, targetResult, error) {
	if attackRange <= 0 || math.IsNaN(attackRange) || math.IsInf(attackRange, 0) {
		return "", targetResult{}, fmt.Errorf("combat: attack range must be finite and positive")
	}
	attackerSelectable, ok := selectables.Get(attacker)
	if !ok {
		return "", targetResult{}, fmt.Errorf("combat: attacker lacks selectable identity")
	}
	attackerIDValue, _ := attackerSelectable.Get("id")
	attackerPosition, _ := positions.Get(attacker)
	attackerLocation, _ := locations.Get(attacker)
	ax, _ := attackerPosition.Get("x")
	ay, _ := attackerPosition.Get("y")
	aAct, _ := attackerLocation.Get("act")
	aLevel, _ := attackerLocation.Get("level_id")
	for _, entity := range selectables.Entities() {
		selectable, _ := selectables.Get(entity)
		id, _ := selectable.Get("id")
		if id != targetID || entity == attacker {
			continue
		}
		position, pok := positions.Get(entity)
		location, lok := locations.Get(entity)
		if !pok || !lok {
			continue
		}
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		act, _ := location.Get("act")
		level, _ := location.Get("level_id")
		if act == aAct && level == aLevel && math.Hypot(x.(float64)-ax.(float64), y.(float64)-ay.(float64)) <= attackRange {
			return attackerIDValue.(string), targetResult{entity: entity, found: true}, nil
		}
	}
	return attackerIDValue.(string), targetResult{}, nil
}

func rollDamage(minimum, maximum Amount, roll uint64) (Amount, error) {
	minimumWhole, minimumErr := minimum.Whole(RoundTowardZero)
	maximumWhole, maximumErr := maximum.Whole(RoundTowardZero)
	if minimumErr != nil || maximumErr != nil || minimum < 0 || maximum < minimum || MustWhole(minimumWhole) != minimum || MustWhole(maximumWhole) != maximum {
		return 0, fmt.Errorf("combat: invalid physical damage range")
	}
	span := uint64(maximumWhole-minimumWhole) + 1
	return FromWhole(minimumWhole + int64(roll%span))
}

func applyPhysical(entity akara.Entity, damage Amount, monsterStats, playerVitals *akara.DynamicStore) (Amount, bool, error) {
	if stats, ok := monsterStats.Get(entity); ok {
		health, _ := stats.Get("health")
		current := Amount(health.(int64))
		remaining := max(current-damage, 0)
		if err := stats.Set("health", remaining.Raw()); err != nil {
			return 0, false, err
		}
		return remaining, current > 0 && remaining == 0, nil
	}
	if vitals, ok := playerVitals.Get(entity); ok {
		health, _ := vitals.Get("health")
		current, err := FromWhole(health.(int64))
		if err != nil {
			return 0, false, err
		}
		remaining := max(current-damage, 0)
		whole, err := remaining.Whole(RoundTowardZero)
		if err != nil || Amount(whole*int64(One)) != remaining {
			return 0, false, fmt.Errorf("combat: player vitals cannot represent fractional health")
		}
		if err := vitals.Set("health", whole); err != nil {
			return 0, false, err
		}
		return remaining, current > 0 && remaining == 0, nil
	}
	return 0, false, fmt.Errorf("combat: target has no health authority")
}

// ApplyPhysical applies already-resolved physical damage through combat's
// existing health authority. Callers must perform their own hit validation.
func ApplyPhysical(engine *gameecs.Engine, entity akara.Entity, damage Amount) (Amount, bool, error) {
	if engine == nil || damage < 0 {
		return 0, false, fmt.Errorf("combat: engine and non-negative damage are required")
	}
	_, _, _, _, _, monsters, players, err := registerMeleeStores(engine)
	if err != nil {
		return 0, false, err
	}
	return applyPhysical(entity, damage, monsters, players)
}

// ApplyConfirmedPhysical records contact resolved by an authoritative caller,
// then applies damage through combat's health owner and emits standard events.
func ApplyConfirmedPhysical(engine *gameecs.Engine, commands *akara.CommandBuffer, tick uint64, attackerID, targetID string, target akara.Entity, damage Amount) (Amount, bool, error) {
	return ApplyConfirmedDamage(engine, commands, tick, attackerID, targetID, target, Physical, damage)
}

// ApplyConfirmedDamage preserves the semantic damage channel while applying
// the already-resolved amount to the target's shared life authority. Resistance
// and absorb belong before this boundary once their stat-source pipeline lands.
func ApplyConfirmedDamage(engine *gameecs.Engine, commands *akara.CommandBuffer, tick uint64, attackerID, targetID string, target akara.Entity, channel Channel, damage Amount) (Amount, bool, error) {
	if commands == nil || attackerID == "" || targetID == "" {
		return 0, false, fmt.Errorf("combat: confirmed impact requires commands and identities")
	}
	if !channel.valid() || damage < 0 {
		return 0, false, fmt.Errorf("combat: confirmed impact requires a valid channel and non-negative damage")
	}
	_, events, _, _, _, _, _, err := registerMeleeStores(engine)
	if err != nil {
		return 0, false, err
	}
	remaining, died, err := ApplyPhysical(engine, target, damage)
	if err != nil {
		return 0, false, err
	}
	commands.CreateDynamic(engine.World(), map[*akara.DynamicStore]map[string]any{events: channelEventValues(EventHitResolved, tick, attackerID, targetID, true, channel, 0, remaining.Raw())})
	commands.CreateDynamic(engine.World(), map[*akara.DynamicStore]map[string]any{events: channelEventValues(EventDamageApplied, tick, attackerID, targetID, true, channel, damage.Raw(), remaining.Raw())})
	if died {
		commands.CreateDynamic(engine.World(), map[*akara.DynamicStore]map[string]any{events: channelEventValues(EventUnitDied, tick, attackerID, targetID, true, channel, damage.Raw(), 0)})
	}
	return remaining, died, nil
}

func stableRoll(domain string, tick uint64, attacker, target string) uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(domain))
	_, _ = hash.Write([]byte(attacker))
	_, _ = hash.Write([]byte(target))
	for shift := 0; shift < 64; shift += 8 {
		_, _ = hash.Write([]byte{byte(tick >> shift)})
	}
	return hash.Sum64()
}

func eventValues(kind string, tick uint64, attacker, target string, hit bool, physical, remaining int64) map[string]any {
	return channelEventValues(kind, tick, attacker, target, hit, Physical, physical, remaining)
}

func channelEventValues(kind string, tick uint64, attacker, target string, hit bool, channel Channel, damage, remaining int64) map[string]any {
	physical := int64(0)
	if channel == Physical {
		physical = damage
	}
	return map[string]any{"kind": kind, "tick": int64(tick), "attacker_id": attacker, "target_id": target, "hit": hit, "physical": physical, "damage_channel": channel.String(), "damage": damage, "remaining_health": remaining}
}
