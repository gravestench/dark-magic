package clientapp

import (
	"math"
	"testing"

	"github.com/gravestench/akara"
)

// combatLabScenario retains only authoritative stores and selected entities, ensuring combat checks
// cannot accidentally pass against presentation mirrors.
type combatLabScenario struct {
	fixture   *realD2LegacyFixture
	positions *akara.DynamicStore
	player    akara.Entity
	target    akara.Entity
	playerX   float64
	playerY   float64
}

// TestCombatLabFixtureEntersBloodMoor proves the production fixture admits a player into wilderness
// and routes a basic attack through real assignment, execution, and damage policy.
func TestCombatLabFixtureEntersBloodMoor(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene:        "combat_lab",
		fixtureCharacters: 1,
		fixtureWorldLevel: 2,
	})
	fixture.advanceOffline(t, 10)

	scenario := newCombatLabScenario(t, fixture)
	scenario.attackNearbyMonster(t)
}

// newCombatLabScenario establishes authoritative player and hostile identities before combat, so a
// later failure is attributable to attack behavior rather than fixture admission.
func newCombatLabScenario(t *testing.T, fixture *realD2LegacyFixture) *combatLabScenario {
	t.Helper()

	identities := requireRealStore(
		t,
		fixture,
		"d2legacy.player.identity",
		"Combat Lab admitted no player identity store",
	)
	if identities.Len() != 1 {
		t.Fatalf("Combat Lab admitted players = %d, want 1", identities.Len())
	}

	player := identities.Entities()[0]
	assertCombatLabPlayerLocation(t, fixture, player)
	monsters := requireCombatLabMonsters(t, fixture)
	positions := requireRealStore(
		t,
		fixture,
		"d2legacy.world.position",
		"Combat Lab has no authoritative positions",
	)
	playerX, playerY := dynamicPosition(positions, player)

	target, found := nearbyMonster(monsters, positions, playerX, playerY, 14)
	if !found {
		t.Fatal("Combat Lab placed no hostile within its visible encounter radius")
	}

	return &combatLabScenario{
		fixture:   fixture,
		positions: positions,
		player:    player,
		target:    target,
		playerX:   playerX,
		playerY:   playerY,
	}
}

// assertCombatLabPlayerLocation requires the controlled fixture player to occupy Blood Moor in
// authority state; presentation-only level changes cannot satisfy the scenario.
func assertCombatLabPlayerLocation(
	t *testing.T,
	fixture *realD2LegacyFixture,
	player akara.Entity,
) {
	t.Helper()

	locations := requireRealStore(
		t,
		fixture,
		"d2legacy.world.location",
		"Combat Lab has no authoritative world locations",
	)

	location, found := locations.Get(player)
	if !found {
		t.Fatal("Combat Lab player has no authoritative location")
	}

	level, _ := location.Get("level_id")
	if level != int64(2) {
		t.Fatalf("Combat Lab player level = %v, want Blood Moor level 2", level)
	}
}

// requireCombatLabMonsters requires the real population system to create hostiles, preventing the
// test from manufacturing a target that bypasses spawn policy.
func requireCombatLabMonsters(
	t *testing.T,
	fixture *realD2LegacyFixture,
) *akara.DynamicStore {
	t.Helper()

	monsters := requireRealStore(
		t,
		fixture,
		"d2legacy.monster.identity",
		"Combat Lab admitted no production Blood Moor hostiles",
	)
	if monsters.Len() == 0 {
		t.Fatal("Combat Lab admitted no production Blood Moor hostiles")
	}

	return monsters
}

// nearbyMonster selects a naturally generated hostile close enough for a deterministic basic attack,
// preserving production population while avoiding route-length sensitivity.
func nearbyMonster(
	monsters *akara.DynamicStore,
	positions *akara.DynamicStore,
	playerX float64,
	playerY float64,
	radius float64,
) (akara.Entity, bool) {
	for _, monster := range monsters.Entities() {
		x, y, found := dynamicPositionIfPresent(positions, monster)
		if found && math.Hypot(x-playerX, y-playerY) <= radius {
			return monster, true
		}
	}

	return 0, false
}

// dynamicPosition treats missing coordinates as a fixture programming error because authoritative
// combatants must always have world positions.
func dynamicPosition(positions *akara.DynamicStore, entity akara.Entity) (float64, float64) {
	x, y, _ := dynamicPositionIfPresent(positions, entity)

	return x, y
}

// dynamicPositionIfPresent joins both coordinate fields atomically for callers that need to filter
// optional population entities.
func dynamicPositionIfPresent(
	positions *akara.DynamicStore,
	entity akara.Entity,
) (float64, float64, bool) {
	position, found := positions.Get(entity)
	if !found {
		return 0, 0, false
	}

	x, _ := position.Get("x")
	y, _ := position.Get("y")

	return x.(float64), y.(float64), true
}

// attackNearbyMonster submits through the shared command path and advances fixed ticks until the
// production attack resolves, rather than invoking a damage helper directly.
func (scenario *combatLabScenario) attackNearbyMonster(t *testing.T) {
	t.Helper()

	selectables := requireRealStore(
		t,
		scenario.fixture,
		"d2legacy.world.selectable",
		"Combat Lab has no authoritative selectable targets",
	)
	selected, _ := selectables.Get(scenario.target)
	targetID, _ := selected.Get("id")

	// This acceptance test does not load the Lua movement integrator. Moving the
	// already-visible target into range isolates the native attack pipeline.
	targetPosition, _ := scenario.positions.Get(scenario.target)
	if err := targetPosition.Set("x", scenario.playerX+2.4); err != nil {
		t.Fatal(err)
	}

	if err := targetPosition.Set("y", scenario.playerY); err != nil {
		t.Fatal(err)
	}

	targetX, targetY := dynamicPosition(scenario.positions, scenario.target)

	stats := requireRealStore(
		t,
		scenario.fixture,
		"d2legacy.monster.stats",
		"Combat Lab has no authoritative monster stats",
	)
	before, _ := stats.Get(scenario.target)
	beforeHealth, _ := before.Get("health")

	if err := scenario.fixture.app.commandIntents.Submit("player.use_skill", map[string]any{
		"side":      "left",
		"target_x":  targetX,
		"target_y":  targetY,
		"target_id": targetID,
	}); err != nil {
		t.Fatal(err)
	}

	// Host-sized slices preserve the session's deliberate catch-up cap.
	scenario.fixture.advanceOffline(t, 20)
	scenario.assertMonsterDamaged(t, stats, beforeHealth.(int64))
}

// assertMonsterDamaged accepts entity removal as a valid kill outcome and otherwise requires lower
// authoritative health, covering both lethal and non-lethal production rolls.
func (scenario *combatLabScenario) assertMonsterDamaged(
	t *testing.T,
	stats *akara.DynamicStore,
	beforeHealth int64,
) {
	t.Helper()

	after, alive := stats.Get(scenario.target)
	if !alive {
		return
	}

	afterHealth, _ := after.Get("health")
	if afterHealth.(int64) < beforeHealth {
		return
	}

	playerX, playerY := dynamicPosition(scenario.positions, scenario.player)
	targetX, targetY := dynamicPosition(scenario.positions, scenario.target)
	approaches, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.combat.attack_approach",
	)
	animations, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.combat.attack_animation",
	)
	events, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.combat.event",
	)

	t.Fatalf(
		"Combat Lab basic attack left health at %v; player=(%.1f,%.1f) target=(%.1f,%.1f) "+
			"approaches=%d animations=%d events=%d",
		afterHealth,
		playerX,
		playerY,
		targetX,
		targetY,
		approaches.Len(),
		animations.Len(),
		events.Len(),
	)
}
