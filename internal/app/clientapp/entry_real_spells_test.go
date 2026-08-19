package clientapp

import (
	"testing"

	"github.com/gravestench/akara"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
)

var spellLabSkillIDs = []int64{
	0, 36, 40, 45, 47, 48, 50, 52, 54, 55, 60, 66, 70, 72, 75, 80, 85, 90,
	94, 95, 98, 99, 100, 103, 104, 105, 108, 109, 110, 115, 120, 124, 125,
	251, 256, 257, 261, 262, 266, 271, 272, 276, 277,
}

// spellLabScenario owns the player, target, and stores shared by production skill assertions.
type spellLabScenario struct {
	fixture     *realD2LegacyFixture
	player      akara.Entity
	target      akara.Entity
	playerX     float64
	playerY     float64
	targetID    string
	monsters    *akara.DynamicStore
	positions   *akara.DynamicStore
	stats       *akara.DynamicStore
	selectables *akara.DynamicStore
	vitals      *akara.DynamicStore
}

// TestSpellLabCastsProductionSkillFamilies exercises production skill families.
func TestSpellLabCastsProductionSkillFamilies(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene:         "spell_lab",
		applySceneDefaults: true,
	})
	fixture.advanceOffline(t, 10)

	scenario := newSpellLabScenario(t, fixture)
	scenario.assertLearnedSkills(t)
	scenario.prepareTarget(t)
	scenario.castFireBolt(t)
	scenario.assertGolemFamilies(t)
}

// newSpellLabScenario verifies that Spell Lab admitted exactly one authoritative player.
func newSpellLabScenario(t *testing.T, fixture *realD2LegacyFixture) *spellLabScenario {
	t.Helper()

	identities := requireRealStore(
		t,
		fixture,
		"d2legacy.player.identity",
		"Spell Lab has no authoritative player identities",
	)
	if identities.Len() != 1 {
		t.Fatalf("Spell Lab admitted players = %d, want 1", identities.Len())
	}

	return &spellLabScenario{
		fixture: fixture,
		player:  identities.Entities()[0],
	}
}

// assertLearnedSkills verifies exact fixture learning and initial left/right assignments.
func (scenario *spellLabScenario) assertLearnedSkills(t *testing.T) {
	t.Helper()

	learned := requireRealStore(
		t,
		scenario.fixture,
		"d2legacy.player.learned_skill",
		"Spell Lab has no authoritative learned skills",
	)
	if learned.Len() != len(spellLabSkillIDs) {
		t.Fatalf(
			"Spell Lab learned skills = %d, want %d exact-ID behaviors",
			learned.Len(),
			len(spellLabSkillIDs),
		)
	}

	learnedIDs := make(map[int64]bool, len(spellLabSkillIDs))

	for _, entity := range learned.Entities() {
		value, _ := learned.Get(entity)

		owner, _ := value.Get("owner")
		if owner != scenario.player {
			continue
		}

		id, _ := value.Get("skill_id")
		level, _ := value.Get("level")
		learnedIDs[id.(int64)] = level == int64(20)
	}

	for _, id := range spellLabSkillIDs {
		if !learnedIDs[id] {
			t.Fatalf("Spell Lab skill %d is missing or not level 20: %#v", id, learnedIDs)
		}
	}

	assignments, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.player.skill_assignment",
	)
	assignment, _ := assignments.Get(scenario.player)
	left, _ := assignment.Get("left")

	right, _ := assignment.Get("right")
	if left != int64(36) || right != int64(66) {
		t.Fatalf(
			"Spell Lab assignments = left %v right %v, want Fire Bolt/Amplify Damage",
			left,
			right,
		)
	}
}

// prepareTarget places one real hostile in Fire Bolt's deterministic line of travel.
func (scenario *spellLabScenario) prepareTarget(t *testing.T) {
	t.Helper()

	scenario.monsters = scenario.requireMonsters(t)
	scenario.positions, _ = akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.world.position",
	)
	scenario.stats = requireRealStore(
		t,
		scenario.fixture,
		"d2legacy.monster.stats",
		"Spell Lab has no authoritative monster stats",
	)
	scenario.selectables = requireRealStore(
		t,
		scenario.fixture,
		"d2legacy.world.selectable",
		"Spell Lab has no authoritative selectable targets",
	)
	scenario.vitals, _ = akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.player.vitals",
	)

	scenario.target = scenario.monsters.Entities()[0]
	scenario.playerX, scenario.playerY = dynamicPosition(scenario.positions, scenario.player)
	targetPosition, _ := scenario.positions.Get(scenario.target)

	// The running lab keeps natural monster placement. Only this acceptance target
	// is held on the projectile line so locomotion cannot obscure cast/contact behavior.
	if err := targetPosition.Set("x", scenario.playerX+6); err != nil {
		t.Fatal(err)
	}

	if err := targetPosition.Set("y", scenario.playerY); err != nil {
		t.Fatal(err)
	}

	selectable, _ := scenario.selectables.Get(scenario.target)
	targetID, _ := selectable.Get("id")
	scenario.targetID = targetID.(string)
}

// requireMonsters returns real Blood Moor hostiles with generation diagnostics on failure.
func (scenario *spellLabScenario) requireMonsters(t *testing.T) *akara.DynamicStore {
	t.Helper()

	monsters, found := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.monster.identity",
	)
	if found && monsters.Len() > 0 {
		return monsters
	}

	app := scenario.fixture.app
	spawn := app.gameWorldSpawns[2]
	room, roomFound := entryworld.RoomIDAt(app.gameWorldZones[2], spawn[0], spawn[1])
	plan, installed := app.authoritativeState.Read("d2legacy.population.plan")
	t.Fatalf(
		"Spell Lab admitted no nearby Blood Moor hostiles; spawn=(%.1f,%.1f) room=%q "+
			"found=%v plan_registered=%v plan=%s",
		spawn[0],
		spawn[1],
		room,
		roomFound,
		installed,
		plan.Data,
	)

	return nil
}

// castFireBolt exercises mana payment, SC timing, projectile damage, and recovery.
func (scenario *spellLabScenario) castFireBolt(t *testing.T) {
	t.Helper()

	targetX, targetY := dynamicPosition(scenario.positions, scenario.target)
	targetStats, _ := scenario.stats.Get(scenario.target)
	beforeHealth, _ := targetStats.Get("health")

	if err := scenario.submitSkill("left", scenario.targetID, targetX, targetY); err != nil {
		t.Fatal(err)
	}

	scenario.fixture.advanceOffline(t, 1)
	maximumMana := scenario.assertFireBoltManaSpent(t)
	scenario.fixture.advanceOffline(t, 1)
	scenario.assertFireBoltCastTiming(t)
	scenario.fixture.advanceOffline(t, 18)
	scenario.assertFireBoltCompleted(t, maximumMana, beforeHealth.(int64))
}

// assertFireBoltManaSpent verifies that payment occurs at cast start.
func (scenario *spellLabScenario) assertFireBoltManaSpent(t *testing.T) int64 {
	t.Helper()

	playerVitals, _ := scenario.vitals.Get(scenario.player)
	immediateMana, _ := playerVitals.Get("mana_raw")

	maximumMana, _ := playerVitals.Get("max_mana_raw")
	if immediateMana.(int64) >= maximumMana.(int64) {
		t.Fatalf(
			"Spell Lab Fire Bolt did not spend mana at cast start: raw=%v max=%v",
			immediateMana,
			maximumMana,
		)
	}

	return maximumMana.(int64)
}

// assertFireBoltCastTiming verifies the production SC action and effect offsets.
func (scenario *spellLabScenario) assertFireBoltCastTiming(t *testing.T) {
	t.Helper()

	animations, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.player.animation",
	)
	animation, _ := animations.Get(scenario.player)
	mode, _ := animation.Get("mode")
	startTick, _ := animation.Get("start_tick")
	casts, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.skill.cast",
	)

	cast, active := casts.Get(scenario.player)
	if mode != "SC" || !active {
		t.Fatalf("Spell Lab Fire Bolt action = mode %v active %v, want SC/true", mode, active)
	}

	effectTick, _ := cast.Get("effect_tick")

	completeTick, _ := cast.Get("complete_tick")
	if effectTick.(int64)-startTick.(int64) != 7 || completeTick.(int64)-startTick.(int64) != 14 {
		t.Fatalf(
			"Spell Lab Fire Bolt SC timing = start %v effect %v complete %v, want +7/+14",
			startTick,
			effectTick,
			completeTick,
		)
	}
}

// assertFireBoltCompleted verifies animation reset, mana recovery, and target damage.
func (scenario *spellLabScenario) assertFireBoltCompleted(
	t *testing.T,
	maximumMana int64,
	beforeHealth int64,
) {
	t.Helper()

	animations, _ := akara.GetDynamicStore(
		scenario.fixture.app.entitySimulation.World(),
		"d2legacy.player.animation",
	)
	animation, _ := animations.Get(scenario.player)

	mode, _ := animation.Get("mode")
	if mode != "NU" {
		t.Fatalf("Spell Lab Fire Bolt animation after completion = %v, want NU", mode)
	}

	playerVitals, _ := scenario.vitals.Get(scenario.player)
	mana, _ := playerVitals.Get("mana")

	manaRaw, _ := playerVitals.Get("mana_raw")
	if mana != int64(4096) || manaRaw != maximumMana {
		t.Fatalf(
			"Spell Lab Fire Bolt mana recovery = %v raw=%v, want full %v after cast completion",
			mana,
			manaRaw,
			maximumMana,
		)
	}

	after, alive := scenario.stats.Get(scenario.target)
	if !alive {
		return
	}

	afterHealth, _ := after.Get("health")
	if afterHealth.(int64) >= beforeHealth {
		t.Fatalf("Spell Lab Fire Bolt left target health unchanged at %v", afterHealth)
	}
}

// submitSkill sends a production skill-use intent through the shared command controller.
func (scenario *spellLabScenario) submitSkill(
	side string,
	targetID string,
	targetX float64,
	targetY float64,
) error {
	return scenario.fixture.app.commandIntents.Submit("player.use_skill", map[string]any{
		"side":      side,
		"target_x":  targetX,
		"target_y":  targetY,
		"target_id": targetID,
	})
}
