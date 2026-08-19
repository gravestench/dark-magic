package clientapp

import (
	"testing"

	"github.com/gravestench/akara"
)

const ironGolemItemID = "fixture-short-sword"

// ironGolemItem records the ground item facts that must survive invalidation and feed success.
type ironGolemItem struct {
	entity         akara.Entity
	modifierEntity akara.Entity
	minimum        int64
	maximum        int64
}

// assertGolemFamilies exercises replacement and each production golem's defining behavior.
func (scenario *spellLabScenario) assertGolemFamilies(t *testing.T) {
	t.Helper()

	scenario.assertClayGolem(t)
	scenario.assertBloodGolem(t)
	scenario.assertFireGolem(t)
	scenario.assertIronGolem(t)
}

// castGolem assigns, starts, validates, and completes one production summon cast.
func (scenario *spellLabScenario) castGolem(t *testing.T, skillID int64, targetID string) {
	t.Helper()

	if err := scenario.fixture.app.commandIntents.Submit("player.assign_skills", map[string]any{
		"right": skillID,
	}); err != nil {
		t.Fatal(err)
	}

	scenario.fixture.advanceOffline(t, 1)

	if err := scenario.submitSkill("right", targetID, scenario.playerX+2, scenario.playerY); err != nil {
		t.Fatal(err)
	}

	scenario.fixture.advanceOffline(t, 1)

	if skillID == 90 {
		scenario.assertIronGolemCastActive(t, targetID)
	}

	scenario.fixture.advanceOffline(t, 17)
}

// assertIronGolemCastActive provides preflight facts when the item-targeted cast fails early.
func (scenario *spellLabScenario) assertIronGolemCastActive(t *testing.T, targetID string) {
	t.Helper()

	casts, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.skill.cast",
	)
	if _, active := casts.Get(scenario.player); active {
		return
	}

	playerVitals, _ := scenario.vitals.Get(scenario.player)
	manaRaw, _ := playerVitals.Get("mana_raw")
	assignments, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.player.skill_assignment",
	)
	assignment, _ := assignments.Get(scenario.player)
	rightSkill, _ := assignment.Get("right")
	itemFacts := scenario.ironTargetFacts(targetID)

	t.Fatalf(
		"Spell Lab Iron Golem failed preflight; mana_raw=%v right_skill=%v target=%q facts=%#v",
		manaRaw,
		rightSkill,
		targetID,
		itemFacts,
	)
}

// ironTargetFacts snapshots the target item's availability for cast-preflight diagnostics.
func (scenario *spellLabScenario) ironTargetFacts(targetID string) map[string]any {
	world := scenario.fixture.app.entitySimulation.World()
	items, _ := akara.GetDynamicStore(world, "d2legacy.item.identity")
	placements, _ := akara.GetDynamicStore(world, "d2legacy.item.placement")
	locations, _ := akara.GetDynamicStore(world, "d2legacy.world.location")
	inactive, _ := akara.GetDynamicStore(world, "d2legacy.world.inactive")

	result := map[string]any{}

	for _, entity := range items.Entities() {
		item, _ := items.Get(entity)

		id, _ := item.Get("id")
		if id != targetID {
			continue
		}

		result["item_types"], _ = item.Get("item_types")

		result["identified"], _ = item.Get("identified")
		if component, found := placements.Get(entity); found {
			result["placement"], _ = component.Snapshot()
		}

		if component, found := scenario.positions.Get(entity); found {
			result["position"], _ = component.Snapshot()
		}

		if component, found := locations.Get(entity); found {
			result["location"], _ = component.Snapshot()
		}

		_, result["inactive"] = inactive.Get(entity)

		break
	}

	if playerLocation, found := locations.Get(scenario.player); found {
		result["player_location"], _ = playerLocation.Snapshot()
	}

	return result
}

// onlyGolem requires exactly one owned summon with the requested production definition.
func (scenario *spellLabScenario) onlyGolem(t *testing.T, definition string) akara.Entity {
	t.Helper()

	owned, found := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.owned_unit",
	)
	if !found || owned.Len() != 1 {
		count := 0
		if found {
			count = owned.Len()
		}

		t.Fatalf("Spell Lab owned golems = %d, want one", count)
	}

	entity := owned.Entities()[0]

	monster, found := scenario.monsters.Get(entity)
	if !found {
		t.Fatalf("Spell Lab golem %d has no monster identity", entity)
	}

	actual, _ := monster.Get("definition_id")
	if actual != definition {
		outcome, reason := scenario.lastGolemOutcome()
		t.Fatalf(
			"Spell Lab golem definition = %v, want %s; last outcome=%v reason=%v",
			actual,
			definition,
			outcome,
			reason,
		)
	}

	return entity
}

// lastGolemOutcome returns the latest summon result for replacement diagnostics.
func (scenario *spellLabScenario) lastGolemOutcome() (any, any) {
	events, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.skill.summon_event",
	)

	var (
		outcome any
		reason  any
	)

	for _, entity := range events.Entities() {
		event, _ := events.Get(entity)

		kind, _ := event.Get("kind")
		if kind == "golem_summon" {
			outcome, _ = event.Get("outcome")
			reason, _ = event.Get("reason")
		}
	}

	return outcome, reason
}

// createMeleeEvent injects a resolved production event to exercise reactive effects.
func (scenario *spellLabScenario) createMeleeEvent(
	t *testing.T,
	attackerID string,
	defenderID string,
	damageRaw int64,
) {
	t.Helper()

	entity := scenario.fixture.app.entitySimulation.World().MustCreateEntity()
	events, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.combat.melee_event",
	)

	_, err := events.Set(entity, map[string]any{
		"kind":                 "hit_resolved",
		"tick":                 int64(0),
		"attacker_id":          attackerID,
		"target_id":            defenderID,
		"hit":                  true,
		"damage_raw":           damageRaw,
		"remaining_health_raw": int64(1),
		"hand":                 "rarm",
		"attack_rating":        int64(1),
		"defense":              int64(0),
		"hit_chance":           int64(95),
		"outcome":              "hit",
	})
	if err != nil {
		t.Fatal(err)
	}
}

// assertClayGolem verifies the melee-triggered state and movement/attack penalties.
func (scenario *spellLabScenario) assertClayGolem(t *testing.T) {
	t.Helper()

	scenario.castGolem(t, 75, "")
	clay := scenario.onlyGolem(t, "claygolem")
	selectable, _ := scenario.selectables.Get(clay)
	clayID, _ := selectable.Get("id")
	scenario.createMeleeEvent(t, scenario.targetID, clayID.(string), 256)
	scenario.fixture.advanceOffline(t, 2)

	states, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.state.instance",
	)
	claySlow := false

	for _, entity := range states.Entities() {
		state, _ := states.Get(entity)
		target, _ := state.Get("target")
		stateID, _ := state.Get("state_id")
		claySlow = claySlow || target == scenario.target && stateID == "slowed"
	}

	velocitySlow, attackSlow := scenario.clayStatPenalties()
	if !claySlow || !velocitySlow || !attackSlow {
		t.Fatalf(
			"Spell Lab Clay Golem melee reaction state=%v velocity=%v attack=%v",
			claySlow,
			velocitySlow,
			attackSlow,
		)
	}
}

// clayStatPenalties finds the target's negative velocity and attack-rate sources.
func (scenario *spellLabScenario) clayStatPenalties() (bool, bool) {
	sources, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.stat.source",
	)
	velocitySlow := false
	attackSlow := false

	for _, entity := range sources.Entities() {
		source, _ := sources.Get(entity)
		target, _ := source.Get("target")
		stat, _ := source.Get("stat")

		value, _ := source.Get("value")
		if target != scenario.target || value.(int64) >= 0 {
			continue
		}

		velocitySlow = velocitySlow || stat == "velocitypercent"
		attackSlow = attackSlow || stat == "attackrate"
	}

	return velocitySlow, attackSlow
}

// assertBloodGolem verifies its reactive life sharing in both transfer directions.
func (scenario *spellLabScenario) assertBloodGolem(t *testing.T) {
	t.Helper()

	scenario.castGolem(t, 85, "")
	blood := scenario.onlyGolem(t, "bloodgolem")

	reactions, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.combat.reactive_effect",
	)
	if _, found := reactions.Get(blood); !found {
		t.Fatal("Spell Lab Blood Golem has no record-derived reactive effect")
	}

	bloodStats, _ := scenario.stats.Get(blood)

	bloodMaximum, _ := bloodStats.Get("max_health")
	if err := bloodStats.Set("health", bloodMaximum.(int64)-20*256); err != nil {
		t.Fatal(err)
	}

	playerVitals, _ := scenario.vitals.Get(scenario.player)

	playerMaximum, _ := playerVitals.Get("max_health")
	if err := playerVitals.Set("health", playerMaximum.(int64)-10); err != nil {
		t.Fatal(err)
	}

	selectable, _ := scenario.selectables.Get(blood)
	bloodID, _ := selectable.Get("id")
	scenario.createMeleeEvent(t, bloodID.(string), scenario.targetID, 10*256)
	scenario.fixture.advanceOffline(t, 1)

	bloodHealth, _ := bloodStats.Get("health")

	playerHealth, _ := playerVitals.Get("health")
	if bloodHealth.(int64) <= bloodMaximum.(int64)-20*256 ||
		playerHealth.(int64) <= playerMaximum.(int64)-10 {
		t.Fatalf(
			"Spell Lab Blood Golem did not split stolen life: golem=%v owner=%v",
			bloodHealth,
			playerHealth,
		)
	}

	beforeOwnerTransfer := bloodHealth.(int64)

	if err := playerVitals.Set("health", playerHealth.(int64)+4); err != nil {
		t.Fatal(err)
	}

	scenario.fixture.advanceOffline(t, 1)

	bloodHealth, _ = bloodStats.Get("health")
	if bloodHealth.(int64)-beforeOwnerTransfer != 256 {
		t.Fatalf(
			"Spell Lab Blood Golem owner-healing transfer = %d raw, want 256",
			bloodHealth.(int64)-beforeOwnerTransfer,
		)
	}
}

// assertFireGolem verifies its granted aura facts and decoded periodic damage schedule.
func (scenario *spellLabScenario) assertFireGolem(t *testing.T) {
	t.Helper()

	scenario.castGolem(t, 94, "")
	fire := scenario.onlyGolem(t, "firegolem")
	scenario.assertFireGolemGrant(t, fire)
	scenario.assertFireGolemPulse(t, fire)
}

// assertFireGolemGrant verifies the record-derived Holy Fire grant and period.
func (scenario *spellLabScenario) assertFireGolemGrant(t *testing.T, fire akara.Entity) {
	t.Helper()

	grants, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.monster.granted_skill",
	)

	grant, found := grants.Get(fire)
	if !found {
		t.Fatal("Spell Lab Fire Golem has no granted Holy Fire fact")
	}

	name, _ := grant.Get("skill")

	level, _ := grant.Get("level")
	if name != "holy fire" || level != int64(27) {
		t.Fatalf("Spell Lab Fire Golem grant = %v level %v, want holy fire/27", name, level)
	}

	periodicDamage, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.combat.periodic_damage",
	)

	periodic, found := periodicDamage.Get(fire)
	if !found {
		t.Fatal("Spell Lab Fire Golem has no decoded periodic-damage schedule")
	}

	period, _ := periodic.Get("period_ticks")

	channel, _ := periodic.Get("channel")
	if period != int64(50) || channel != "fire" {
		t.Fatalf("Spell Lab Fire Golem periodic damage = period %v channel %v", period, channel)
	}
}

// assertFireGolemPulse verifies that Holy Fire emits an authoritative damage event.
func (scenario *spellLabScenario) assertFireGolemPulse(t *testing.T, fire akara.Entity) {
	t.Helper()

	selectable, _ := scenario.selectables.Get(fire)
	fireID, _ := selectable.Get("id")
	targetStats, _ := scenario.stats.Get(scenario.target)

	targetMaximum, _ := targetStats.Get("max_health")
	if err := targetStats.Set("health", targetMaximum); err != nil {
		t.Fatal(err)
	}

	scenario.fixture.advanceOffline(t, 51)

	events, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.combat.event",
	)
	for _, entity := range events.Entities() {
		event, _ := events.Get(entity)
		sourceKind, _ := event.Get("source_kind")

		attackerID, _ := event.Get("attacker_id")
		if sourceKind == "periodic_damage" && attackerID == fireID {
			return
		}
	}

	t.Fatal("Spell Lab Fire Golem emitted no Holy Fire damage event")
}
