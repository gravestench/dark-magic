package combat

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
	"github.com/gravestench/dark-magic/internal/game/targeting"
)

func TestBasicMeleeAppliesDamageAndEmitsDeath(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	if err := RegisterBasicMelee(engine, BasicMeleePolicy{HitChance: 100}); err != nil {
		t.Fatal(err)
	}
	stores := meleeTestStores(t, engine)
	attacker := engine.World().MustCreateEntity()
	target := engine.World().MustCreateEntity()
	setMeleeAttacker(t, stores, attacker, "monster:fallen", "player:hero", 2, 2, MustWhole(5))
	setMeleeTarget(t, stores, target, "player:hero", 3, 2, 5)
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	vitals, _ := stores.playerVitals.Get(target)
	health, _ := vitals.Get("health")
	if health != int64(0) {
		t.Fatalf("health = %v, want 0", health)
	}
	if stores.requests.Has(attacker) {
		t.Fatal("consumed attack request remains")
	}
	wantKinds := []string{EventAttackAttempted, EventHitResolved, EventDamageApplied, EventUnitDied}
	if stores.events.Len() != len(wantKinds) {
		t.Fatalf("events = %d, want %d", stores.events.Len(), len(wantKinds))
	}
	for index, entity := range stores.events.Entities() {
		event, _ := stores.events.Get(entity)
		kind, _ := event.Get("kind")
		if kind != wantKinds[index] {
			t.Fatalf("event %d = %v, want %s", index, kind, wantKinds[index])
		}
	}
}

func TestPlayerAttackSkillUsesSharedMeleeTransaction(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	registry, err := gameskill.NewRegistry(gameskill.Definition{
		SkillID: 0, Behavior: gameskill.BehaviorBasicMelee, TargetPolicy: gameskill.TargetUnit,
		EffectDelay: 1, CompleteDelay: 2, Interruptible: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gameskill.RegisterCastLifecycle(engine, registry); err != nil {
		t.Fatal(err)
	}
	if err := RegisterPlayerBasicAttack(engine, 0, nil); err != nil {
		t.Fatal(err)
	}
	if err := RegisterBasicMelee(engine, BasicMeleePolicy{HitChance: 100}); err != nil {
		t.Fatal(err)
	}
	stores := meleeTestStores(t, engine)
	player := engine.World().MustCreateEntity()
	target := engine.World().MustCreateEntity()
	mustSetEntity(t, stores.selectables, player, map[string]any{"id": "player:hero", "kind": targeting.KindPlayer, "label": "Hero", "owner": "hero", "radius": 1.0, "priority": int64(10)})
	mustSetEntity(t, stores.positions, player, map[string]any{"x": 2.0, "y": 2.0})
	mustSetEntity(t, stores.locations, player, map[string]any{"act": int64(1), "level_id": int64(2)})
	mustSetEntity(t, stores.profiles, player, map[string]any{"range": 2.0, "physical_min": MustWhole(2).Raw(), "physical_max": MustWhole(2).Raw()})
	mustSetEntity(t, stores.playerVitals, player, map[string]any{"health": int64(10), "max_health": int64(10), "mana": int64(0), "max_mana": int64(0), "mana_raw": int64(0), "max_mana_raw": int64(0)})
	velocities, _ := akara.GetDynamicStore(engine.World(), "dm.world.velocity")
	mustSetEntity(t, velocities, player, map[string]any{"x": 0.0, "y": 0.0})
	controls, _ := akara.RegisterSchema(engine.World(), akara.Schema{Name: "dm.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	mustSetEntity(t, controls, player, map[string]any{"player": "hero"})
	mustSetEntity(t, stores.selectables, target, map[string]any{"id": "monster:fallen", "kind": targeting.KindHostile, "label": "Fallen", "owner": "", "radius": 0.5, "priority": int64(20)})
	mustSetEntity(t, stores.positions, target, map[string]any{"x": 8.0, "y": 2.0})
	mustSetEntity(t, stores.locations, target, map[string]any{"act": int64(1), "level_id": int64(2)})
	mustSetEntity(t, stores.monsterStats, target, map[string]any{"level": int64(1), "health": MustWhole(5).Raw(), "max_health": MustWhole(5).Raw(), "defense": int64(0), "attack_rating": int64(0), "physical_min": int64(0), "physical_max": int64(0), "experience": int64(0)})
	requests, _, _, _, _, _ := gameskillStores(t, engine)
	mustSetEntity(t, requests, player, map[string]any{"player": "hero", "side": "left", "skill_id": int64(0), "skill_level": int64(1), "target_x": 8.0, "target_y": 2.0, "target_id": "monster:fallen", "request_tick": int64(1)})

	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	// The skill effect records a pending action at the pre-simulation barrier.
	// The following fixed tick revalidates its target and admits the hit.
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	velocity, _ := velocities.Get(player)
	velocityX, _ := velocity.Get("x")
	if velocityX.(float64) <= 0 {
		t.Fatalf("approach velocity x = %v, want movement toward target", velocityX)
	}
	targetStats, _ := stores.monsterStats.Get(target)
	before, _ := targetStats.Get("health")
	if before != MustWhole(5).Raw() {
		t.Fatalf("out-of-range target health = %v, want unchanged", before)
	}
	playerPosition, _ := stores.positions.Get(player)
	if err := playerPosition.Set("x", 7.0); err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
	stats, _ := stores.monsterStats.Get(target)
	health, _ := stats.Get("health")
	if health != MustWhole(3).Raw() {
		t.Fatalf("monster health = %v, want %d", health, MustWhole(3).Raw())
	}
}

func gameskillStores(t *testing.T, engine *gameecs.Engine) (*akara.DynamicStore, *akara.DynamicStore, *akara.DynamicStore, *akara.DynamicStore, *akara.DynamicStore, *akara.DynamicStore) {
	t.Helper()
	requests, _ := akara.GetDynamicStore(engine.World(), gameskill.CastRequestComponent)
	states, _ := akara.GetDynamicStore(engine.World(), gameskill.CastStateComponent)
	events, _ := akara.GetDynamicStore(engine.World(), gameskill.CastEventComponent)
	vitals, _ := akara.GetDynamicStore(engine.World(), "dm.player.vitals")
	selectables, _ := akara.GetDynamicStore(engine.World(), targeting.Component)
	controls, _ := akara.GetDynamicStore(engine.World(), "dm.world.player_control")
	return requests, states, events, vitals, selectables, controls
}

func mustSetEntity(t *testing.T, store *akara.DynamicStore, entity akara.Entity, values map[string]any) {
	t.Helper()
	if _, err := store.Set(entity, values); err != nil {
		t.Fatal(err)
	}
}

func TestBasicMeleeMissAndIllegalRangeDoNotMutateHealth(t *testing.T) {
	for _, test := range []struct {
		name      string
		hitChance int
		targetX   float64
	}{
		{name: "synthetic miss", hitChance: 0, targetX: 3},
		{name: "out of range", hitChance: 100, targetX: 20},
	} {
		t.Run(test.name, func(t *testing.T) {
			engine := gameecs.New()
			defer engine.Close()
			if err := RegisterBasicMelee(engine, BasicMeleePolicy{HitChance: test.hitChance}); err != nil {
				t.Fatal(err)
			}
			stores := meleeTestStores(t, engine)
			attacker := engine.World().MustCreateEntity()
			target := engine.World().MustCreateEntity()
			setMeleeAttacker(t, stores, attacker, "monster:fallen", "player:hero", 2, 2, MustWhole(3))
			setMeleeTarget(t, stores, target, "player:hero", test.targetX, 2, 10)
			if err := engine.Update(time.Second / 25); err != nil {
				t.Fatal(err)
			}
			vitals, _ := stores.playerVitals.Get(target)
			health, _ := vitals.Get("health")
			if health != int64(10) {
				t.Fatalf("health = %v, want 10", health)
			}
			if stores.events.Len() != 2 {
				t.Fatalf("events = %d, want attempt and miss", stores.events.Len())
			}
		})
	}
}

func TestRollDamageKeepsWholeAuthoredEndpoints(t *testing.T) {
	for roll := uint64(0); roll < 20; roll++ {
		damage, err := rollDamage(MustWhole(2), MustWhole(4), roll)
		if err != nil {
			t.Fatal(err)
		}
		whole, _ := damage.Whole(RoundTowardZero)
		if damage != MustWhole(whole) || whole < 2 || whole > 4 {
			t.Fatalf("roll %d produced %v", roll, damage)
		}
	}
}

func TestPhysicalDamageReportsOnlyTheLethalTransition(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	if err := RegisterBasicMelee(engine, BasicMeleePolicy{HitChance: 100}); err != nil {
		t.Fatal(err)
	}
	stores := meleeTestStores(t, engine)
	target := engine.World().MustCreateEntity()
	setMeleeTarget(t, stores, target, "player:hero", 1, 1, 1)
	_, firstDied, err := applyPhysical(target, MustWhole(1), stores.monsterStats, stores.playerVitals)
	if err != nil {
		t.Fatal(err)
	}
	_, secondDied, err := applyPhysical(target, MustWhole(1), stores.monsterStats, stores.playerVitals)
	if err != nil {
		t.Fatal(err)
	}
	if !firstDied || secondDied {
		t.Fatalf("death transitions = first:%t second:%t", firstDied, secondDied)
	}
}

type meleeStores struct {
	requests, events, selectables, positions, locations, profiles, monsterStats, playerVitals *akara.DynamicStore
}

func meleeTestStores(t *testing.T, engine *gameecs.Engine) meleeStores {
	t.Helper()
	requests, _ := akara.GetDynamicStore(engine.World(), BasicAttackRequest)
	events, _ := akara.GetDynamicStore(engine.World(), CombatEvent)
	selectables, _ := akara.GetDynamicStore(engine.World(), targeting.Component)
	positions, _ := akara.GetDynamicStore(engine.World(), "dm.world.position")
	locations, _ := akara.GetDynamicStore(engine.World(), "dm.world.location")
	profiles, _ := akara.GetDynamicStore(engine.World(), MeleeProfile)
	monsterStats, _ := akara.GetDynamicStore(engine.World(), "dm.monster.stats")
	playerVitals, _ := akara.GetDynamicStore(engine.World(), "dm.player.vitals")
	return meleeStores{requests, events, selectables, positions, locations, profiles, monsterStats, playerVitals}
}

func setMeleeAttacker(t *testing.T, stores meleeStores, entity akara.Entity, id, target string, x, y float64, damage Amount) {
	t.Helper()
	mustSet := func(store *akara.DynamicStore, values map[string]any) {
		if _, err := store.Set(entity, values); err != nil {
			t.Fatal(err)
		}
	}
	mustSet(stores.requests, map[string]any{"target_id": target, "request_tick": int64(1)})
	mustSet(stores.profiles, map[string]any{"range": 2.0, "physical_min": damage.Raw(), "physical_max": damage.Raw()})
	mustSet(stores.selectables, map[string]any{"id": id, "kind": targeting.KindHostile, "label": "Fallen", "owner": "", "radius": 1.0, "priority": int64(20)})
	mustSet(stores.positions, map[string]any{"x": x, "y": y})
	mustSet(stores.locations, map[string]any{"act": int64(1), "level_id": int64(2)})
	mustSet(stores.monsterStats, map[string]any{"level": int64(1), "health": MustWhole(10).Raw(), "max_health": MustWhole(10).Raw(), "defense": int64(0), "attack_rating": int64(1), "physical_min": damage.Raw(), "physical_max": damage.Raw(), "experience": int64(0)})
}

func setMeleeTarget(t *testing.T, stores meleeStores, entity akara.Entity, id string, x, y float64, health int64) {
	t.Helper()
	mustSet := func(store *akara.DynamicStore, values map[string]any) {
		if _, err := store.Set(entity, values); err != nil {
			t.Fatal(err)
		}
	}
	mustSet(stores.selectables, map[string]any{"id": id, "kind": targeting.KindPlayer, "label": "Hero", "owner": "hero", "radius": 1.0, "priority": int64(10)})
	mustSet(stores.positions, map[string]any{"x": x, "y": y})
	mustSet(stores.locations, map[string]any{"act": int64(1), "level_id": int64(2)})
	mustSet(stores.playerVitals, map[string]any{"health": health, "max_health": health, "mana": int64(0), "max_mana": int64(0)})
}
