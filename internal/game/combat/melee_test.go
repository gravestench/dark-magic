package combat

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
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

type meleeStores struct {
	requests, events, selectables, positions, locations, monsterStats, playerVitals *akara.DynamicStore
}

func meleeTestStores(t *testing.T, engine *gameecs.Engine) meleeStores {
	t.Helper()
	requests, _ := akara.GetDynamicStore(engine.World(), BasicAttackRequest)
	events, _ := akara.GetDynamicStore(engine.World(), CombatEvent)
	selectables, _ := akara.GetDynamicStore(engine.World(), targeting.Component)
	positions, _ := akara.GetDynamicStore(engine.World(), "dm.world.position")
	locations, _ := akara.GetDynamicStore(engine.World(), "dm.world.location")
	monsterStats, _ := akara.GetDynamicStore(engine.World(), "dm.monster.stats")
	playerVitals, _ := akara.GetDynamicStore(engine.World(), "dm.player.vitals")
	return meleeStores{requests, events, selectables, positions, locations, monsterStats, playerVitals}
}

func setMeleeAttacker(t *testing.T, stores meleeStores, entity akara.Entity, id, target string, x, y float64, damage Amount) {
	t.Helper()
	mustSet := func(store *akara.DynamicStore, values map[string]any) {
		if _, err := store.Set(entity, values); err != nil {
			t.Fatal(err)
		}
	}
	mustSet(stores.requests, map[string]any{"target_id": target, "request_tick": int64(1), "range": 2.0})
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
