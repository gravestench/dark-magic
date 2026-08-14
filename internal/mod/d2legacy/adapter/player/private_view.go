package player

import (
	"sort"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const PrivateViewVersion uint32 = 1

// PrivateView is owner-only state needed by connected presentation. It is an
// explicit allowlist rather than a serialized ECS snapshot, so another
// player's inventory or interaction context cannot cross the protocol boundary.
type PrivateView struct {
	Version     uint32          `json:"version"`
	Tick        uint64          `json:"tick"`
	Items       ItemView        `json:"items"`
	Interaction InteractionView `json:"interaction"`
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
	Active bool                   `json:"active"`
	Target *InteractionTargetView `json:"target,omitempty"`
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
}

func ProjectPrivateView(playerID string, checkpoint simulation.Checkpoint) (PrivateView, error) {
	view := PrivateView{Version: PrivateViewVersion, Tick: checkpoint.Tick, Items: ItemView{Items: []ItemEntityView{}}}
	if checkpoint.Snapshot == nil {
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
	contexts, found := findComponent(snapshot, "d2legacy.interaction.context")
	if found {
		if _, fields, present := findString(contexts, "owner", playerID); present {
			targetEntity := entityField(fields, "target")
			if targets, exists := findComponent(snapshot, "d2legacy.interaction.target"); exists {
				if target, active := findInstance(targets, targetEntity); active {
					view.Interaction.Active = true
					view.Interaction.Target = &InteractionTargetView{
						ID: stringField(target, "id"), NPC: stringField(target, "npc"), Vendor: stringField(target, "vendor"),
						Categories: stringField(target, "categories"), Services: stringField(target, "services"),
						X: floatField(target, "x"), Y: floatField(target, "y"), Radius: floatField(target, "radius"),
					}
				}
			}
		}
	}
	return view, nil
}

func itemLayout(fields map[string]gameecs.ValueSnapshot) ItemLayoutView {
	return ItemLayoutView{
		InventoryWidth: intField(fields, "inventory_width"), InventoryHeight: intField(fields, "inventory_height"),
		StashWidth: intField(fields, "stash_width"), StashHeight: intField(fields, "stash_height"),
		CubeWidth: intField(fields, "cube_width"), CubeHeight: intField(fields, "cube_height"),
		BeltCapacity: intField(fields, "belt_capacity"), ActiveWeaponSet: intField(fields, "active_weapon_set"),
		VendorWidth: intField(fields, "vendor_width"), VendorHeight: intField(fields, "vendor_height"),
		CarriedGold: intField(fields, "carried_gold"), StashedGold: intField(fields, "stashed_gold"),
	}
}

func projectItems(snapshot gameecs.Snapshot, layoutEntity uint64) []ItemEntityView {
	identities, found := findComponent(snapshot, "d2legacy.item.identity")
	if !found {
		return []ItemEntityView{}
	}
	placements, _ := findComponent(snapshot, "d2legacy.item.placement")
	presentations, _ := findComponent(snapshot, "d2legacy.item.presentation")
	items := []ItemEntityView{}
	for _, instance := range identities.Instances {
		identity, present := findInstance(identities, instance.Entity)
		if !present || entityField(identity, "owner") != layoutEntity {
			continue
		}
		placement, _ := findInstance(placements, instance.Entity)
		presentation, _ := findInstance(presentations, instance.Entity)
		items = append(items, ItemEntityView{
			ID: stringField(identity, "id"), Code: stringField(identity, "code"), Width: intField(identity, "width"), Height: intField(identity, "height"),
			BodySlots: stringField(identity, "body_slots"), BeltEligible: boolField(identity, "belt_eligible"), BaseCost: intField(identity, "base_cost"), AppliedServices: stringField(identity, "applied_services"),
			Container: stringField(placement, "container"), X: intField(placement, "x"), Y: intField(placement, "y"), Slot: stringField(placement, "slot"), BeltSlot: intField(placement, "belt_slot"), WeaponSet: intField(placement, "weapon_set"), Page: intField(placement, "page"),
			InventoryDC6: stringField(presentation, "inventory_dc6"), WorldDC6: stringField(presentation, "world_dc6"), WorldAnimated: boolField(presentation, "world_animated"), Composite: stringField(presentation, "composite"), WeaponClass: stringField(presentation, "weapon_class"),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}
