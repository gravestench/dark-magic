package player

import (
	"sort"
	"strings"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const PrivateViewVersion uint32 = 3

// PrivateView is owner-only state needed by connected presentation. It is an
// explicit allowlist rather than a serialized ECS snapshot, so another
// player's inventory or interaction context cannot cross the protocol boundary.
type PrivateView struct {
	Version     uint32          `json:"version"`
	Tick        uint64          `json:"tick"`
	Items       ItemView        `json:"items"`
	Interaction InteractionView `json:"interaction"`
	Quest       *QuestView      `json:"quest,omitempty"`
}

type QuestView struct {
	QuestID             int64 `json:"quest_id"`
	Stage               int64 `json:"stage"`
	Completed           bool  `json:"completed"`
	Rewarded            bool  `json:"rewarded"`
	Remaining           int64 `json:"remaining"`
	UnsettledRooms      int64 `json:"unsettled_rooms"`
	PopulationAvailable bool  `json:"population_available"`
}

type ItemView struct {
	Layout ItemLayoutView   `json:"layout"`
	Items  []ItemEntityView `json:"items"`
}

type ItemLayoutView struct {
	InventoryWidth  int64 `json:"inventory_width"`
	InventoryHeight int64 `json:"inventory_height"`
	StashWidth      int64 `json:"stash_width"`
	StashHeight     int64 `json:"stash_height"`
	CubeWidth       int64 `json:"cube_width"`
	CubeHeight      int64 `json:"cube_height"`
	BeltCapacity    int64 `json:"belt_capacity"`
	ActiveWeaponSet int64 `json:"active_weapon_set"`
	VendorWidth     int64 `json:"vendor_width"`
	VendorHeight    int64 `json:"vendor_height"`
	CarriedGold     int64 `json:"carried_gold"`
	StashedGold     int64 `json:"stashed_gold"`
}

type ItemEntityView struct {
	Vendor          string `json:"vendor,omitempty"`
	BuyPrice        int64  `json:"buy_price,omitempty"`
	ID              string `json:"id"`
	Code            string `json:"code"`
	BodySlots       string `json:"body_slots"`
	AppliedServices string `json:"applied_services"`
	Width           int64  `json:"width"`
	Height          int64  `json:"height"`
	BaseCost        int64  `json:"base_cost"`
	BeltEligible    bool   `json:"belt_eligible"`
	Container       string `json:"container"`
	Slot            string `json:"slot"`
	X               int64  `json:"x"`
	Y               int64  `json:"y"`
	BeltSlot        int64  `json:"belt_slot"`
	WeaponSet       int64  `json:"weapon_set"`
	Page            int64  `json:"page"`
	InventoryDC6    string `json:"inventory_dc6"`
	WorldDC6        string `json:"world_dc6"`
	Composite       string `json:"composite"`
	WeaponClass     string `json:"weapon_class"`
	WorldAnimated   bool   `json:"world_animated"`
}

type InteractionView struct {
	TradeAvailable   bool                   `json:"trade_available,omitempty"`
	Active           bool                   `json:"active"`
	Target           *InteractionTargetView `json:"target,omitempty"`
	DialoguePage     string                 `json:"dialogue_page,omitempty"`
	DialogueLine     string                 `json:"dialogue_line,omitempty"`
	DialogueRevision int64                  `json:"dialogue_revision,omitempty"`
}

type InteractionTargetView struct {
	ID         string  `json:"id"`
	NPC        string  `json:"npc"`
	Vendor     string  `json:"vendor"`
	Categories string  `json:"categories"`
	Services   string  `json:"services"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Radius     float64 `json:"radius"`
	Act        int64   `json:"act,omitempty"`
	LevelID    int64   `json:"level_id,omitempty"`
}

// ProjectPrivateView selects inventory and interaction state owned by the
// authenticated player. Missing optional domains produce empty views rather
// than exposing another player's nearest matching component.
func ProjectPrivateView(playerID string, checkpoint simulation.Checkpoint) (PrivateView, error) {
	view := PrivateView{
		Version: PrivateViewVersion,
		Tick:    checkpoint.Tick,
		Items:   ItemView{Items: []ItemEntityView{}},
	}
	if strings.TrimSpace(playerID) == "" || checkpoint.Snapshot == nil {
		return view, ErrHUDPlayer
	}

	snapshot := *checkpoint.Snapshot

	layouts, found := findComponent(snapshot, "d2legacy.items.layout")
	if found {
		if layoutEntity, fields, present := findString(layouts, "owner", playerID); present {
			view.Items.Layout = itemLayout(fields)
			view.Items.Items = projectItems(snapshot, layoutEntity)
		}
	}

	view.Interaction = projectInteraction(snapshot, playerID)
	if identities, found := findComponent(snapshot, "d2legacy.player.identity"); found {
		if entity, _, present := findString(identities, "player", playerID); present {
			if quests, found := findComponent(snapshot, "d2legacy.quest.progress"); found {
				if fields, present := findInstance(quests, entity); present {
					view.Quest = &QuestView{QuestID: intField(fields, "quest_id"), Stage: intField(fields, "stage"),
						Completed: boolField(fields, "completed"), Rewarded: boolField(fields, "rewarded"),
						Remaining: intField(fields, "remaining"), UnsettledRooms: intField(fields, "unsettled_rooms"),
						PopulationAvailable: boolField(fields, "population_available")}
				}
			}
		}
	}

	return view, nil
}

// projectInteraction follows the owner's context relationship to its active
// target. Both links must exist before the view is marked active, keeping the
// Active/Target wire invariant intact.
func projectInteraction(snapshot gameecs.Snapshot, playerID string) InteractionView {
	contexts, found := findComponent(snapshot, "d2legacy.interaction.context")
	if !found {
		return InteractionView{}
	}

	_, fields, present := findString(contexts, "owner", playerID)
	if !present {
		return InteractionView{}
	}

	targets, exists := findComponent(snapshot, "d2legacy.interaction.target")
	if !exists {
		return InteractionView{}
	}

	target, active := findInstance(targets, entityField(fields, "target"))
	if !active {
		return InteractionView{}
	}

	view := InteractionView{
		Active:           true,
		DialoguePage:     stringField(fields, "dialogue_page"),
		DialogueLine:     stringField(fields, "dialogue_line"),
		DialogueRevision: intField(fields, "dialogue_revision"),
		TradeAvailable:   boolField(fields, "trade_available"),
		Target: &InteractionTargetView{
			ID:         stringField(target, "id"),
			NPC:        stringField(target, "npc"),
			Vendor:     stringField(target, "vendor"),
			Categories: stringField(target, "categories"),
			Services:   stringField(target, "services"),
			X:          floatField(target, "x"),
			Y:          floatField(target, "y"),
			Radius:     floatField(target, "radius"),
		},
	}
	if locations, found := findComponent(snapshot, "d2legacy.world.location"); found {
		if location, present := findInstance(locations, entityField(fields, "target")); present {
			view.Target.Act = intField(location, "act")
			view.Target.LevelID = intField(location, "level_id")
		}
	}

	return view
}

// itemLayout copies only the bounded presentation dimensions and balances; raw
// inventory authority and item-placement mechanics remain in ECS.
func itemLayout(fields map[string]gameecs.ValueSnapshot) ItemLayoutView {
	return ItemLayoutView{
		InventoryWidth:  intField(fields, "inventory_width"),
		InventoryHeight: intField(fields, "inventory_height"),
		StashWidth:      intField(fields, "stash_width"),
		StashHeight:     intField(fields, "stash_height"),
		CubeWidth:       intField(fields, "cube_width"),
		CubeHeight:      intField(fields, "cube_height"),
		BeltCapacity:    intField(fields, "belt_capacity"),
		ActiveWeaponSet: intField(fields, "active_weapon_set"),
		VendorWidth:     intField(fields, "vendor_width"),
		VendorHeight:    intField(fields, "vendor_height"),
		CarriedGold:     intField(fields, "carried_gold"),
		StashedGold:     intField(fields, "stashed_gold"),
	}
}

// projectItems joins owner identity with optional placement and presentation
// components. Sorting by stable item ID removes ECS iteration order from JSON.
func projectItems(snapshot gameecs.Snapshot, layoutEntity uint64) []ItemEntityView {
	identities, found := findComponent(snapshot, "d2legacy.item.identity")
	if !found {
		return []ItemEntityView{}
	}

	placements, _ := findComponent(snapshot, "d2legacy.item.placement")
	offers, _ := findComponent(snapshot, "d2legacy.vendor.offer")
	presentations, _ := findComponent(snapshot, "d2legacy.item.presentation")
	inactive, _ := findComponent(snapshot, "d2legacy.world.inactive")
	items := []ItemEntityView{}

	for _, instance := range identities.Instances {
		identity, present := findInstance(identities, instance.Entity)
		if !present || entityField(identity, "owner") != layoutEntity {
			continue
		}

		if _, hidden := findInstance(inactive, instance.Entity); hidden {
			continue
		}

		placement, _ := findInstance(placements, instance.Entity)
		presentation, _ := findInstance(presentations, instance.Entity)
		item := projectItem(identity, placement, presentation)
		offer, _ := findInstance(offers, instance.Entity)
		item.Vendor = stringField(offer, "vendor")
		item.BuyPrice = intField(offer, "price")
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })

	return items
}

// projectItem keeps the mechanical join separate from owner filtering, making
// the exact private-field allowlist visible in one place.
func projectItem(
	identity map[string]gameecs.ValueSnapshot,
	placement map[string]gameecs.ValueSnapshot,
	presentation map[string]gameecs.ValueSnapshot,
) ItemEntityView {
	return ItemEntityView{
		ID:              stringField(identity, "id"),
		Code:            stringField(identity, "code"),
		BodySlots:       stringField(identity, "body_slots"),
		AppliedServices: stringField(identity, "applied_services"),
		Width:           intField(identity, "width"),
		Height:          intField(identity, "height"),
		BaseCost:        intField(identity, "base_cost"),
		BeltEligible:    boolField(identity, "belt_eligible"),
		Container:       stringField(placement, "container"),
		Slot:            stringField(placement, "slot"),
		X:               intField(placement, "x"),
		Y:               intField(placement, "y"),
		BeltSlot:        intField(placement, "belt_slot"),
		WeaponSet:       intField(placement, "weapon_set"),
		Page:            intField(placement, "page"),
		InventoryDC6:    stringField(presentation, "inventory_dc6"),
		WorldDC6:        stringField(presentation, "world_dc6"),
		Composite:       stringField(presentation, "composite"),
		WeaponClass:     stringField(presentation, "weapon_class"),
		WorldAnimated:   boolField(presentation, "world_animated"),
	}
}
