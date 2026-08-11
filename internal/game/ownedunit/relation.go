package ownedunit

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gravestench/akara"
)

const Component = "dm.owned_unit"

// Relation is checkpointed ECS state. Immediate owner and ultimate owner are
// both retained because a trap or minion kill is not identical to a direct
// player kill even when both credit the same player.
type Relation struct {
	Unit               akara.Entity `json:"unit"`
	Owner              akara.Entity `json:"owner"`
	OwnerID            string       `json:"owner_id"`
	UltimateOwnerID    string       `json:"ultimate_owner_id"`
	Category           string       `json:"category"`
	Group              int64        `json:"group"`
	Limit              int64        `json:"limit"`
	Replacement        Replacement  `json:"replacement"`
	CreatedTick        uint64       `json:"created_tick"`
	ExpiresTick        uint64       `json:"expires_tick"`
	DurableID          string       `json:"durable_id"`
	Durable            bool         `json:"durable"`
	Unsummon           bool         `json:"unsummon"`
	WarpWithOwner      bool         `json:"warp_with_owner"`
	RangeLimited       bool         `json:"range_limited"`
	Active             bool         `json:"active"`
	SurvivesOwnerDeath bool         `json:"survives_owner_death"`
}

// Decision explains deterministic limit handling to the spawning caller.
type Decision struct {
	Accepted    bool           `json:"accepted"`
	Inactivated []akara.Entity `json:"inactivated,omitempty"`
}

func Schema() akara.Schema {
	return akara.Schema{Name: Component, Version: 1, Fields: []akara.Field{
		{Name: "owner", Kind: akara.FieldEntity}, {Name: "owner_id", Kind: akara.FieldString},
		{Name: "ultimate_owner_id", Kind: akara.FieldString}, {Name: "category", Kind: akara.FieldString},
		{Name: "group", Kind: akara.FieldInt64}, {Name: "limit", Kind: akara.FieldInt64},
		{Name: "replacement", Kind: akara.FieldString}, {Name: "created_tick", Kind: akara.FieldUint64},
		{Name: "expires_tick", Kind: akara.FieldUint64}, {Name: "durable_id", Kind: akara.FieldString},
		{Name: "durable", Kind: akara.FieldBool}, {Name: "unsummon", Kind: akara.FieldBool},
		{Name: "warp_with_owner", Kind: akara.FieldBool}, {Name: "range_limited", Kind: akara.FieldBool},
		{Name: "active", Kind: akara.FieldBool}, {Name: "survives_owner_death", Kind: akara.FieldBool},
	}}
}

// Attach validates and commits one relationship plus any deterministic limit
// replacement. The unit itself already belongs to its monster/trap/hireling
// owner; this function adds only the cross-cutting relationship.
func Attach(world *akara.World, relation Relation, category Category) (Decision, error) {
	if world == nil || relation.Unit == 0 || relation.Owner == 0 || relation.Unit == relation.Owner || !entityExists(world, relation.Unit) || !entityExists(world, relation.Owner) {
		return Decision{}, fmt.Errorf("owned unit: distinct live unit and owner are required")
	}
	if err := category.validate(); err != nil {
		return Decision{}, err
	}
	relation.OwnerID = strings.TrimSpace(relation.OwnerID)
	relation.UltimateOwnerID = strings.TrimSpace(relation.UltimateOwnerID)
	if relation.OwnerID == "" || relation.UltimateOwnerID == "" {
		return Decision{}, fmt.Errorf("owned unit: immediate and ultimate owner identities are required")
	}
	store, err := akara.RegisterSchema(world, Schema())
	if err != nil {
		return Decision{}, err
	}
	if store.Has(relation.Unit) {
		return Decision{}, fmt.Errorf("owned unit: entity %d already has an owner", relation.Unit)
	}
	relation.Category, relation.Group, relation.Limit, relation.Replacement = category.ID, category.Group, category.BaseMax, category.Replacement
	relation.Durable, relation.Unsummon = category.Durable, category.Unsummon
	relation.WarpWithOwner, relation.RangeLimited = category.WarpWithOwner, category.RangeLimited
	relation.Active, relation.SurvivesOwnerDeath = true, category.SurvivesOwnerDeath
	victims, err := replacementVictims(store, relation, category)
	if err != nil {
		return Decision{}, err
	}
	if _, err := store.Set(relation.Unit, relationValues(relation)); err != nil {
		return Decision{}, err
	}
	for _, victim := range victims {
		component, _ := store.Get(victim)
		values, err := component.Snapshot()
		if err != nil {
			return Decision{}, err
		}
		values["active"] = false
		if _, err := store.Set(victim, values); err != nil {
			return Decision{}, err
		}
	}
	return Decision{Accepted: true, Inactivated: victims}, nil
}

type candidate struct {
	entity akara.Entity
	tick   uint64
}

func replacementVictims(store *akara.DynamicStore, relation Relation, category Category) ([]akara.Entity, error) {
	var sameCategory, sameGroup []candidate
	for _, entity := range store.Entities() {
		component, _ := store.Get(entity)
		active, _ := component.Get("active")
		owner, _ := component.Get("owner")
		if active != true || owner != relation.Owner {
			continue
		}
		categoryValue, _ := component.Get("category")
		group, _ := component.Get("group")
		created, _ := component.Get("created_tick")
		entry := candidate{entity: entity, tick: created.(uint64)}
		if categoryValue == category.ID {
			sameCategory = append(sameCategory, entry)
		}
		if category.Group > 0 && group == category.Group && categoryValue != category.ID {
			sameGroup = append(sameGroup, entry)
		}
	}
	needed := max(0, len(sameCategory)+1-int(category.BaseMax))
	if needed == 0 && len(sameGroup) == 0 {
		return nil, nil
	}
	if category.Replacement == Reject {
		return nil, fmt.Errorf("owned unit: owner %q reached category %q limit", relation.OwnerID, category.ID)
	}
	victims := append([]candidate(nil), sameGroup...)
	sortCandidates(sameCategory, category.Replacement)
	victims = append(victims, sameCategory[:needed]...)
	sort.Slice(victims, func(i, j int) bool { return victims[i].entity < victims[j].entity })
	result := make([]akara.Entity, len(victims))
	for index := range victims {
		result[index] = victims[index].entity
	}
	return result, nil
}

func sortCandidates(items []candidate, policy Replacement) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].tick == items[j].tick {
			return items[i].entity < items[j].entity
		}
		if policy == ReplaceNewest {
			return items[i].tick > items[j].tick
		}
		return items[i].tick < items[j].tick
	})
}

func relationValues(relation Relation) map[string]any {
	return map[string]any{"owner": relation.Owner, "owner_id": relation.OwnerID, "ultimate_owner_id": relation.UltimateOwnerID, "category": relation.Category, "group": relation.Group, "limit": relation.Limit, "replacement": string(relation.Replacement), "created_tick": relation.CreatedTick, "expires_tick": relation.ExpiresTick, "durable_id": relation.DurableID, "durable": relation.Durable, "unsummon": relation.Unsummon, "warp_with_owner": relation.WarpWithOwner, "range_limited": relation.RangeLimited, "active": relation.Active, "survives_owner_death": relation.SurvivesOwnerDeath}
}

func entityExists(world *akara.World, target akara.Entity) bool {
	for _, entity := range world.Entities() {
		if entity == target {
			return true
		}
	}
	return false
}

// Attribution resolves both identities without guessing from proximity or
// appearance. Combat may credit UltimateOwnerID while retaining SourceID.
type Attribution struct{ SourceID, ImmediateOwnerID, UltimateOwnerID string }

func ResolveAttribution(world *akara.World, unit akara.Entity, sourceID string) (Attribution, bool) {
	store, found := akara.GetDynamicStore(world, Component)
	if !found {
		return Attribution{}, false
	}
	component, found := store.Get(unit)
	if !found {
		return Attribution{}, false
	}
	owner, _ := component.Get("owner_id")
	ultimate, _ := component.Get("ultimate_owner_id")
	return Attribution{SourceID: strings.TrimSpace(sourceID), ImmediateOwnerID: owner.(string), UltimateOwnerID: ultimate.(string)}, true
}

// ActiveFor returns stable entity order for UI, limits, transitions, and tests.
func ActiveFor(world *akara.World, owner akara.Entity, category string) []akara.Entity {
	store, found := akara.GetDynamicStore(world, Component)
	if !found {
		return nil
	}
	var result []akara.Entity
	for _, entity := range store.Entities() {
		component, _ := store.Get(entity)
		active, _ := component.Get("active")
		ownerValue, _ := component.Get("owner")
		categoryValue, _ := component.Get("category")
		if active == true && ownerValue == owner && (category == "" || categoryValue == category) {
			result = append(result, entity)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}
