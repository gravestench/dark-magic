package clientapp

import (
	"encoding/json"
	"fmt"

	"github.com/gravestench/akara"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// installPrivateProjection rebuilds the authenticated owner's skills, inventory, and interaction as
// one disposable graph. These facts are private to the admitted player and never join public mirrors.
func installPrivateProjection(
	world *akara.World,
	hero akara.Entity,
	owner string,
	learned []playeradapter.HUDLearnedSkill,
	private playeradapter.PrivateView,
) error {
	if err := clearPrivateProjection(world); err != nil {
		return err
	}

	if err := installLearnedSkills(world, hero, learned); err != nil {
		return err
	}

	layout, err := installPrivateItemLayout(world, owner, private.Items.Layout)
	if err != nil {
		return err
	}

	if err := installPrivateItems(world, layout, private.Items.Items); err != nil {
		return err
	}

	return installPrivateInteraction(world, owner, private.Interaction)
}

// clearPrivateProjection removes roots of every owner-private graph before replacement. Destroying
// whole graphs avoids retaining items or targets omitted from the newest complete private view.
func clearPrivateProjection(world *akara.World) error {
	components := []string{
		"d2legacy.player.learned_skill",
		"d2legacy.items.layout",
		"d2legacy.interaction.context",
		"d2legacy.interaction.target",
		"d2legacy.interaction.null_target",
	}

	for _, name := range components {
		store, found := akara.GetDynamicStore(world, name)
		if !found {
			return fmt.Errorf("remote presentation: component %q is unavailable", name)
		}

		for _, entity := range store.Entities() {
			world.DestroyEntity(entity)
		}
	}

	return nil
}

// installLearnedSkills attaches projected learned-skill facts to the authenticated hero for menus and
// input validation; creating these entities does not grant or execute skills locally.
func installLearnedSkills(
	world *akara.World,
	hero akara.Entity,
	learned []playeradapter.HUDLearnedSkill,
) error {
	store, _ := akara.GetDynamicStore(world, "d2legacy.player.learned_skill")

	for _, skill := range learned {
		entity, err := world.CreateEntity()
		if err != nil {
			return err
		}

		_, err = store.Set(entity, map[string]any{
			"owner":         hero,
			"skill_id":      skill.SkillID,
			"level":         skill.Level,
			"list_row":      skill.ListRow,
			"left_allowed":  skill.LeftAllowed,
			"right_allowed": skill.RightAllowed,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// installPrivateItemLayout creates the graph owner containing dimensions, weapon-set selection, and
// gold totals. Individual item entities reference it instead of duplicating owner identity.
func installPrivateItemLayout(
	world *akara.World,
	owner string,
	layout playeradapter.ItemLayoutView,
) (akara.Entity, error) {
	entity, err := world.CreateEntity()
	if err != nil {
		return 0, err
	}

	store, _ := akara.GetDynamicStore(world, "d2legacy.items.layout")

	_, err = store.Set(entity, map[string]any{
		"owner":             owner,
		"inventory_width":   layout.InventoryWidth,
		"inventory_height":  layout.InventoryHeight,
		"stash_width":       layout.StashWidth,
		"stash_height":      layout.StashHeight,
		"cube_width":        layout.CubeWidth,
		"cube_height":       layout.CubeHeight,
		"belt_capacity":     layout.BeltCapacity,
		"active_weapon_set": layout.ActiveWeaponSet,
		"vendor_width":      layout.VendorWidth,
		"vendor_height":     layout.VendorHeight,
		"carried_gold":      layout.CarriedGold,
		"stashed_gold":      layout.StashedGold,
	})
	if err != nil {
		return 0, err
	}

	return entity, nil
}

// installPrivateItems treats the private item list as complete, destroys the previous list, and
// rebuilds identity, placement, and presentation together for each item.
func installPrivateItems(
	world *akara.World,
	layout akara.Entity,
	items []playeradapter.ItemEntityView,
) error {
	identityStore, identityOK := akara.GetDynamicStore(world, "d2legacy.item.identity")
	placementStore, placementOK := akara.GetDynamicStore(world, "d2legacy.item.placement")
	presentationStore, presentationOK := akara.GetDynamicStore(world, "d2legacy.item.presentation")

	if !identityOK || !placementOK || !presentationOK {
		return fmt.Errorf("remote presentation: item stores are unavailable")
	}

	for _, entity := range identityStore.Entities() {
		world.DestroyEntity(entity)
	}

	for _, item := range items {
		entity, err := world.CreateEntity()
		if err != nil {
			return err
		}

		if _, err := identityStore.Set(entity, privateItemIdentity(layout, item)); err != nil {
			return err
		}

		if _, err := placementStore.Set(entity, privateItemPlacement(item)); err != nil {
			return err
		}

		if _, err := presentationStore.Set(entity, privateItemPresentation(item)); err != nil {
			return err
		}
	}

	return nil
}

// privateItemIdentity maps allowlisted stable facts and links the item to its private layout. It does
// not copy authority-only rolls or behavior into the presentation ECS.
func privateItemIdentity(
	layout akara.Entity,
	item playeradapter.ItemEntityView,
) map[string]any {
	return map[string]any{
		"owner":            layout,
		"id":               item.ID,
		"code":             item.Code,
		"width":            item.Width,
		"height":           item.Height,
		"body_slots":       item.BodySlots,
		"belt_eligible":    item.BeltEligible,
		"base_cost":        item.BaseCost,
		"applied_services": item.AppliedServices,
	}
}

// privateItemPlacement keeps mutually applicable inventory, equipment, belt, weapon-set, and page
// coordinates in the registered schema shape used by UI systems.
func privateItemPlacement(item playeradapter.ItemEntityView) map[string]any {
	return map[string]any{
		"container":  item.Container,
		"x":          item.X,
		"y":          item.Y,
		"slot":       item.Slot,
		"belt_slot":  item.BeltSlot,
		"weapon_set": item.WeaponSet,
		"page":       item.Page,
	}
}

// privateItemPresentation exposes only asset and composite identifiers required for rendering; item
// effects remain authoritative and absent from the connected replica.
func privateItemPresentation(item playeradapter.ItemEntityView) map[string]any {
	return map[string]any{
		"inventory_dc6":  item.InventoryDC6,
		"world_dc6":      item.WorldDC6,
		"world_animated": item.WorldAnimated,
		"composite":      item.Composite,
		"weapon_class":   item.WeaponClass,
	}
}

// installPrivateInteraction rebuilds owner context around either an active target or an explicit null
// target, giving Lua a stable relationship without inventing a zero entity reference.
func installPrivateInteraction(
	world *akara.World,
	owner string,
	interaction playeradapter.InteractionView,
) error {
	target, err := installPrivateInteractionTarget(world, interaction)
	if err != nil {
		return err
	}

	contextEntity, err := world.CreateEntity()
	if err != nil {
		return err
	}

	contextStore, _ := akara.GetDynamicStore(world, "d2legacy.interaction.context")
	_, err = contextStore.Set(contextEntity, map[string]any{
		"owner":  owner,
		"target": target,
	})

	return err
}

// installPrivateInteractionTarget always creates a schema-valid null target first, then replaces it
// with an allowlisted active target when authority reports one.
func installPrivateInteractionTarget(
	world *akara.World,
	interaction playeradapter.InteractionView,
) (akara.Entity, error) {
	nullTarget, err := world.CreateEntity()
	if err != nil {
		return 0, err
	}

	if store, found := akara.GetDynamicStore(world, "d2legacy.interaction.null_target"); found {
		if _, err := store.Set(nullTarget, map[string]any{}); err != nil {
			return 0, err
		}
	}

	if !interaction.Active || interaction.Target == nil {
		return nullTarget, nil
	}

	entity, err := world.CreateEntity()
	if err != nil {
		return 0, err
	}

	store, found := akara.GetDynamicStore(world, "d2legacy.interaction.target")
	if !found {
		return 0, fmt.Errorf("remote presentation: interaction target store is unavailable")
	}

	target := interaction.Target

	_, err = store.Set(entity, map[string]any{
		"id":         target.ID,
		"npc":        target.NPC,
		"vendor":     target.Vendor,
		"categories": target.Categories,
		"services":   target.Services,
		"x":          target.X,
		"y":          target.Y,
		"radius":     target.Radius,
	})
	if err != nil {
		return 0, err
	}

	return entity, nil
}

// privateProjectionFingerprint removes transport tick metadata before hashing complete private
// content. Unchanged inventories and skill graphs therefore retain ECS identity across corrections.
func privateProjectionFingerprint(
	learned []playeradapter.HUDLearnedSkill,
	private playeradapter.PrivateView,
) (string, error) {
	// Tick is transport metadata and changes even when presentation does not.
	private.Tick = 0

	payload, err := json.Marshal(struct {
		Learned []playeradapter.HUDLearnedSkill `json:"learned"`
		Private playeradapter.PrivateView       `json:"private"`
	}{
		Learned: learned,
		Private: private,
	})
	if err != nil {
		return "", fmt.Errorf("remote presentation: fingerprint private projection: %w", err)
	}

	return string(payload), nil
}
