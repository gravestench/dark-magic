package monster

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/gravestench/akara"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameloot "github.com/gravestench/dark-magic/internal/game/loot"
	"github.com/gravestench/dark-magic/internal/game/targeting"
)

const (
	DeathSystemID     = "monster.death_transaction"
	DeathTransaction  = "dm.monster.death"
	DeathEvent        = "dm.monster.death_event"
	DeathEventKilled  = "monster_killed"
	DeathEventLoot    = "monster_loot"
	DeathEventQuest   = "monster_quest_kill"
	DeathEventPresent = "monster_death_presented"
)

// DeathPolicy contains immutable session inputs. The initial policy deliberately
// awards XP only to a player who dealt the lethal hit. Party sharing, summon
// ownership, corpse skills, and corpse expiry require their own verified rules.
type DeathPolicy struct {
	WorldSeed uint64
	Loot      gameloot.Catalog
}

func deathSchema() akara.Schema {
	return akara.Schema{Name: DeathTransaction, Version: 1, Fields: []akara.Field{
		{Name: "tick", Kind: akara.FieldInt64}, {Name: "killer_id", Kind: akara.FieldString},
		{Name: "xp", Kind: akara.FieldInt64}, {Name: "loot_seed", Kind: akara.FieldString},
		{Name: "treasure_class", Kind: akara.FieldString}, {Name: "drops", Kind: akara.FieldString},
		{Name: "active", Kind: akara.FieldBool}, {Name: "corpse_usable", Kind: akara.FieldBool},
	}}
}

func deathEventSchema() akara.Schema {
	return akara.Schema{Name: DeathEvent, Version: 1, Fields: []akara.Field{
		{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64},
		{Name: "monster_id", Kind: akara.FieldString}, {Name: "killer_id", Kind: akara.FieldString},
		{Name: "xp", Kind: akara.FieldInt64}, {Name: "loot_seed", Kind: akara.FieldString},
		{Name: "treasure_class", Kind: akara.FieldString}, {Name: "drops", Kind: akara.FieldString},
	}}
}

// RegisterDeath installs the single effects-phase commit point for monster
// death. Health reaching zero is only an input; this system owns consequences.
func RegisterDeath(engine *gameecs.Engine, policy DeathPolicy) error {
	if engine == nil {
		return fmt.Errorf("monster: death transaction requires an engine")
	}
	stores, err := registerDeathStores(engine)
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: DeathSystemID, Phase: gameecs.PhaseEffects,
		All: []akara.ComponentType{stores.combatEvents}, Read: []akara.ComponentType{stores.combatEvents, stores.identity, stores.stats, stores.selectable, stores.progress},
		Write: []akara.ComponentType{stores.combatEvents, stores.death, stores.deathEvents, stores.appearance, stores.ai, stores.velocity, stores.collider, stores.selectable, stores.progress},
		Update: func(context gameecs.Context, events []akara.Entity, commands *akara.CommandBuffer) error {
			return commitDeaths(context, events, commands, engine.World(), stores, policy)
		},
	})
}

type deathStores struct {
	combatEvents, identity, stats, appearance, ai, velocity, collider, selectable *akara.DynamicStore
	progress, death, deathEvents                                                  *akara.DynamicStore
}

func registerDeathStores(engine *gameecs.Engine) (deathStores, error) {
	schemas := []akara.Schema{
		{Name: gamecombat.CombatEvent, Version: 1, Fields: []akara.Field{{Name: "kind", Kind: akara.FieldString}, {Name: "tick", Kind: akara.FieldInt64}, {Name: "attacker_id", Kind: akara.FieldString}, {Name: "target_id", Kind: akara.FieldString}, {Name: "hit", Kind: akara.FieldBool}, {Name: "physical", Kind: akara.FieldInt64}, {Name: "remaining_health", Kind: akara.FieldInt64}}},
		{Name: "dm.monster.identity", Version: 2, Fields: []akara.Field{{Name: "spawn_id", Kind: akara.FieldString}, {Name: "definition_id", Kind: akara.FieldString}, {Name: "base_id", Kind: akara.FieldString}, {Name: "graphics_id", Kind: akara.FieldString}, {Name: "seed", Kind: akara.FieldString}, {Name: "treasure_class", Kind: akara.FieldString}}},
		{Name: "dm.monster.stats", Version: 1, Fields: []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "defense", Kind: akara.FieldInt64}, {Name: "attack_rating", Kind: akara.FieldInt64}, {Name: "physical_min", Kind: akara.FieldInt64}, {Name: "physical_max", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}}},
		{Name: "dm.monster.appearance", Version: 1, Fields: []akara.Field{{Name: "token", Kind: akara.FieldString}, {Name: "mode", Kind: akara.FieldString}, {Name: "weapon_class", Kind: akara.FieldString}, {Name: "name_key", Kind: akara.FieldString}}},
		aiSchema(),
		{Name: "dm.world.velocity", Version: 1, Fields: []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}},
		{Name: "dm.world.collider", Version: 1, Fields: []akara.Field{{Name: "radius", Kind: akara.FieldFloat64}}},
		targeting.Schema(),
		{Name: "dm.player.progress", Version: 1, Fields: []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}}},
		deathSchema(), deathEventSchema(),
	}
	registered := make([]*akara.DynamicStore, len(schemas))
	for index, schema := range schemas {
		store, err := akara.RegisterSchema(engine.World(), schema)
		if err != nil {
			return deathStores{}, err
		}
		registered[index] = store
	}
	return deathStores{registered[0], registered[1], registered[2], registered[3], registered[4], registered[5], registered[6], registered[7], registered[8], registered[9], registered[10]}, nil
}

func commitDeaths(context gameecs.Context, eventEntities []akara.Entity, commands *akara.CommandBuffer, world *akara.World, stores deathStores, policy DeathPolicy) error {
	processed := make(map[akara.Entity]bool)
	for _, eventEntity := range eventEntities {
		event, _ := stores.combatEvents.Get(eventEntity)
		kind, _ := event.Get("kind")
		if kind != gamecombat.EventUnitDied {
			continue
		}
		attackerValue, _ := event.Get("attacker_id")
		targetValue, _ := event.Get("target_id")
		monster, found := findSelectable(stores, targetValue.(string), true)
		if !found || stores.death.Has(monster) || processed[monster] {
			commands.Destroy(world, eventEntity)
			continue
		}
		processed[monster] = true
		identity, _ := stores.identity.Get(monster)
		stats, _ := stores.stats.Get(monster)
		spawnValue, _ := identity.Get("spawn_id")
		treasureValue, _ := identity.Get("treasure_class")
		xpValue, _ := stats.Get("experience")
		seed, drops, err := deathLoot(policy, spawnValue.(string), treasureValue.(string))
		if err != nil {
			return err
		}
		dropsJSON, err := json.Marshal(drops)
		if err != nil {
			return err
		}
		killerID := attackerValue.(string)
		xp := xpValue.(int64)
		awardXP(commands, stores, killerID, xp)
		commands.AddDynamic(stores.death, monster, map[string]any{"tick": int64(context.Tick), "killer_id": killerID, "xp": xp, "loot_seed": fmt.Sprint(seed), "treasure_class": treasureValue.(string), "drops": string(dropsJSON), "active": false, "corpse_usable": true})
		if appearance, ok := stores.appearance.Get(monster); ok {
			commands.AddDynamic(stores.appearance, monster, copyWith(appearance, "mode", "DT"))
		}
		commands.AddDynamic(stores.velocity, monster, map[string]any{"x": 0.0, "y": 0.0})
		commands.Remove(stores.ai, monster)
		commands.Remove(stores.collider, monster)
		commands.Remove(stores.selectable, monster)
		for _, eventKind := range []string{DeathEventKilled, DeathEventLoot, DeathEventQuest, DeathEventPresent} {
			commands.CreateDynamic(world, map[*akara.DynamicStore]map[string]any{stores.deathEvents: {"kind": eventKind, "tick": int64(context.Tick), "monster_id": spawnValue.(string), "killer_id": killerID, "xp": xp, "loot_seed": fmt.Sprint(seed), "treasure_class": treasureValue.(string), "drops": string(dropsJSON)}})
		}
		commands.Destroy(world, eventEntity)
	}
	return nil
}

func deathLoot(policy DeathPolicy, spawnID, class string) (uint64, []gameloot.Drop, error) {
	entityID := stableID(spawnID)
	seed, err := gameloot.EventSeed(policy.WorldSeed, gameloot.Event{Kind: gameloot.EventMonster, EntityID: entityID, Sequence: 1})
	if err != nil {
		return 0, nil, err
	}
	if strings.TrimSpace(class) == "" || policy.Loot == nil {
		return seed, nil, nil
	}
	drops, err := gameloot.RollEvent(policy.Loot, class, policy.WorldSeed, gameloot.Event{Kind: gameloot.EventMonster, EntityID: entityID, Sequence: 1})
	return seed, drops, err
}

func stableID(value string) uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(value))
	result := hash.Sum64()
	if result == 0 {
		return 1
	}
	return result
}

func findSelectable(stores deathStores, id string, monster bool) (akara.Entity, bool) {
	for _, entity := range stores.selectable.Entities() {
		component, _ := stores.selectable.Get(entity)
		value, _ := component.Get("id")
		if value == id && (!monster || stores.identity.Has(entity)) {
			return entity, true
		}
	}
	return 0, false
}

func awardXP(commands *akara.CommandBuffer, stores deathStores, killerID string, xp int64) {
	killer, found := findSelectable(stores, killerID, false)
	if !found || !stores.progress.Has(killer) || xp <= 0 {
		return
	}
	progress, _ := stores.progress.Get(killer)
	level, _ := progress.Get("level")
	current, _ := progress.Get("experience")
	commands.AddDynamic(stores.progress, killer, map[string]any{"level": level, "experience": current.(int64) + xp})
}

func copyWith(component *akara.DynamicComponent, field string, value any) map[string]any {
	result := make(map[string]any)
	for _, name := range []string{"token", "mode", "weapon_class", "name_key"} {
		result[name], _ = component.Get(name)
	}
	result[field] = value
	return result
}
