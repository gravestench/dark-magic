package clientapp

import (
	"testing"

	"github.com/gravestench/akara"
)

// assertIronGolem verifies invalidation, item consumption, provenance, and local damage.
func (scenario *spellLabScenario) assertIronGolem(t *testing.T) {
	t.Helper()

	item := scenario.prepareIronGolemItem(t)
	scenario.assertInvalidatedIronGolem(t, item)
	scenario.moveIronItem(t, "world", scenario.playerX, scenario.playerY)
	scenario.fixture.advanceOffline(t, 2)
	scenario.castGolem(t, 90, ironGolemItemID)
	iron := scenario.onlyGolem(t, "irongolem")
	scenario.assertIronGolemProvenance(t, iron, item)
}

// prepareIronGolemItem moves and validates the fixture weapon, then adds local damage.
func (scenario *spellLabScenario) prepareIronGolemItem(t *testing.T) ironGolemItem {
	t.Helper()

	scenario.moveIronItem(t, "world", scenario.playerX, scenario.playerY)
	scenario.fixture.advanceOffline(t, 2)

	item := scenario.findIronGolemItem(t)
	item.modifierEntity = scenario.fixture.app.entitySimulation.World().MustCreateEntity()
	modifiers, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.item.stat_modifier",
	)

	_, err := modifiers.Set(item.modifierEntity, map[string]any{
		"item":        item.entity,
		"source_id":   "acceptance-enhanced-damage",
		"source_kind": "affix",
		"stat":        "damagepercent",
		"operation":   "local_percent",
		"value":       int64(50),
		"order":       int64(10),
	})
	if err != nil {
		t.Fatal(err)
	}

	return item
}

// findIronGolemItem validates the identified metal weapon and records its base damage.
func (scenario *spellLabScenario) findIronGolemItem(t *testing.T) ironGolemItem {
	t.Helper()

	world := scenario.fixture.app.entitySimulation.World()
	items, _ := akara.GetDynamicStore(world, "d2legacy.item.identity")
	placements, _ := akara.GetDynamicStore(world, "d2legacy.item.placement")
	inactive, _ := akara.GetDynamicStore(world, "d2legacy.world.inactive")
	meleeStats, _ := akara.GetDynamicStore(world, "d2legacy.item.melee")

	for _, entity := range items.Entities() {
		item, _ := items.Get(entity)

		id, _ := item.Get("id")
		if id != ironGolemItemID {
			continue
		}

		scenario.assertIronItemAvailable(t, entity, item, placements, inactive)
		melee, _ := meleeStats.Get(entity)
		minimum, _ := melee.Get("physical_min")
		maximum, _ := melee.Get("physical_max")

		return ironGolemItem{
			entity:  entity,
			minimum: minimum.(int64),
			maximum: maximum.(int64),
		}
	}

	t.Fatal("Spell Lab Iron target item is missing before cast")

	return ironGolemItem{}
}

// assertIronItemAvailable verifies material, identification, placement, and room activity.
func (scenario *spellLabScenario) assertIronItemAvailable(
	t *testing.T,
	entity akara.Entity,
	item *akara.DynamicComponent,
	placements *akara.DynamicStore,
	inactive *akara.DynamicStore,
) {
	t.Helper()

	itemTypes, _ := item.Get("item_types")
	materialFlags, _ := item.Get("material_flags")
	identified, _ := item.Get("identified")
	placement, _ := placements.Get(entity)

	container, _ := placement.Get("container")
	if itemTypes == "" || materialFlags.(int64)&2 == 0 || identified != true || container != "world" {
		t.Fatalf(
			"Spell Lab Iron target types=%v material=%v identified=%v container=%v",
			itemTypes,
			materialFlags,
			identified,
			container,
		)
	}

	if _, unavailable := inactive.Get(entity); unavailable {
		t.Fatal("Spell Lab Iron target item became inactive in the player's current room")
	}
}

// assertInvalidatedIronGolem proves effect-time revalidation preserves item and current summon.
func (scenario *spellLabScenario) assertInvalidatedIronGolem(
	t *testing.T,
	item ironGolemItem,
) {
	t.Helper()

	playerVitals, _ := scenario.vitals.Get(scenario.player)
	beforeMana, _ := playerVitals.Get("mana_raw")

	if err := scenario.fixture.app.commandIntents.Submit("player.assign_skills", map[string]any{
		"right": int64(90),
	}); err != nil {
		t.Fatal(err)
	}

	scenario.fixture.advanceOffline(t, 1)

	if err := scenario.submitSkill("right", ironGolemItemID, scenario.playerX+2, scenario.playerY); err != nil {
		t.Fatal(err)
	}

	scenario.fixture.advanceOffline(t, 1)
	scenario.moveIronItem(t, "inventory", 8, 0)
	scenario.fixture.advanceOffline(t, 17)

	playerVitals, _ = scenario.vitals.Get(scenario.player)

	afterMana, _ := playerVitals.Get("mana_raw")
	if afterMana.(int64) >= beforeMana.(int64) {
		t.Fatal("Spell Lab invalidated Iron Golem cast did not retain its paid mana cost")
	}

	scenario.onlyGolem(t, "firegolem")

	items, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.item.identity",
	)
	if _, found := items.Get(item.entity); !found {
		t.Fatal("Spell Lab invalidated Iron Golem cast consumed its target item")
	}

	if !scenario.hasInvalidatedIronEvent() {
		t.Fatal("Spell Lab emitted no item_not_on_ground result for invalidated Iron Golem cast")
	}
}

// hasInvalidatedIronEvent finds the expected item-not-on-ground summon result.
func (scenario *spellLabScenario) hasInvalidatedIronEvent() bool {
	events, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.skill.summon_event",
	)
	for _, entity := range events.Entities() {
		event, _ := events.Get(entity)
		outcome, _ := event.Get("outcome")

		reason, _ := event.Get("reason")
		if outcome == "invalidated" && reason == "item_not_on_ground" {
			return true
		}
	}

	return false
}

// moveIronItem moves the fixture weapon through the production inventory command.
func (scenario *spellLabScenario) moveIronItem(
	t *testing.T,
	container string,
	x float64,
	y float64,
) {
	t.Helper()

	err := scenario.fixture.app.commandIntents.Submit("item.move", map[string]any{
		"item_id": ironGolemItemID,
		"destination": map[string]any{
			"container": container,
			"x":         x,
			"y":         y,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// assertIronGolemProvenance verifies item consumption and locally enhanced damage transfer.
func (scenario *spellLabScenario) assertIronGolemProvenance(
	t *testing.T,
	iron akara.Entity,
	item ironGolemItem,
) {
	t.Helper()

	provenance, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.summon.item_provenance",
	)

	source, found := provenance.Get(iron)
	if !found {
		t.Fatal("Spell Lab Iron Golem has no consumed-item provenance")
	}

	consumedID, _ := source.Get("item_id")
	identified, _ := source.Get("identified")
	resolvedMinimum, _ := source.Get("resolved_weapon_minimum_raw")
	resolvedMaximum, _ := source.Get("resolved_weapon_maximum_raw")

	if consumedID != ironGolemItemID || identified != true {
		t.Fatalf("Spell Lab Iron Golem provenance = %v identified=%v", consumedID, identified)
	}

	if resolvedMinimum != item.minimum*150/100 || resolvedMaximum != item.maximum*150/100 {
		t.Fatalf(
			"Spell Lab Iron Golem local Enhanced Damage = %v-%v, want %d-%d",
			resolvedMinimum,
			resolvedMaximum,
			item.minimum*150/100,
			item.maximum*150/100,
		)
	}

	modifiers, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.item.stat_modifier",
	)
	if _, found := modifiers.Get(item.modifierEntity); found {
		t.Fatal("Spell Lab Iron Golem retained the consumed item's local modifier entity")
	}

	stats, _ := scenario.stats.Get(iron)
	minimum, _ := stats.Get("physical_min")

	maximum, _ := stats.Get("physical_max")
	if minimum.(int64) <= item.minimum || maximum.(int64) <= item.maximum {
		t.Fatalf(
			"Spell Lab Iron Golem base-item damage = %v-%v, consumed item %d-%d",
			minimum,
			maximum,
			item.minimum,
			item.maximum,
		)
	}
}
