package missile

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameskill "github.com/gravestench/dark-magic/internal/game/skill"
)

func TestStraightMissileCastMovesHitsAndReplays(t *testing.T) {
	skills, missiles := testRegistries(t)
	engine := gameecs.New()
	defer engine.Close()
	registerTestSystems(t, engine, skills, missiles)
	stores, err := registerStores(engine)
	if err != nil {
		t.Fatal(err)
	}
	requests := store(t, engine, gameskill.CastRequestComponent)
	vitals := store(t, engine, "d2.player.vitals")
	monsterStats := registerMonsterStats(t, engine)

	caster := engine.World().MustCreateEntity()
	set(t, stores.controls, caster, map[string]any{"player": "alpha"})
	set(t, stores.positions, caster, map[string]any{"x": 0.0, "y": 0.0})
	set(t, stores.locations, caster, map[string]any{"act": int64(1), "level_id": int64(2)})
	set(t, vitals, caster, map[string]any{"health": int64(10), "max_health": int64(10), "mana": int64(10), "max_mana": int64(10)})
	set(t, requests, caster, map[string]any{"player": "alpha", "side": "right", "skill_id": int64(42), "skill_level": int64(3), "target_x": 4.0, "target_y": 0.0, "target_id": "monster", "request_tick": int64(1)})

	target := engine.World().MustCreateEntity()
	set(t, stores.selectables, target, map[string]any{"id": "monster", "kind": "hostile", "label": "test monster", "owner": "", "radius": 0.25, "priority": int64(0)})
	set(t, stores.positions, target, map[string]any{"x": 2.0, "y": 0.0})
	set(t, stores.locations, target, map[string]any{"act": int64(1), "level_id": int64(2)})
	set(t, stores.colliders, target, map[string]any{"radius": 0.25})
	set(t, monsterStats, target, map[string]any{"level": int64(1), "health": gamecombat.MustWhole(10).Raw(), "max_health": gamecombat.MustWhole(10).Raw(), "defense": int64(0), "attack_rating": int64(0), "physical_min": int64(0), "physical_max": int64(0), "experience": int64(0)})

	step(t, engine) // cast starts
	step(t, engine) // effect spawns and advances the missile from 0 to 1
	if stores.missiles.Len() != 1 {
		t.Fatalf("missiles=%d", stores.missiles.Len())
	}
	instance, _ := stores.missiles.Get(stores.missiles.Entities()[0])
	dcc, _ := instance.Get("dcc")
	travel, _ := instance.Get("travel_sound")
	if dcc != "data/global/missiles/test.dcc" || travel != "test_travel" {
		t.Fatalf("presentation recipe dcc=%q travel=%q", dcc, travel)
	}
	snapshot, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := gameecs.RestoreSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	registerTestSystems(t, restored, skills, missiles)

	step(t, engine)
	step(t, restored)
	assertHitResult(t, engine, target)
	left, _ := engine.Snapshot()
	right, _ := restored.Snapshot()
	leftSum, _ := left.Checksum()
	rightSum, _ := right.Checksum()
	if leftSum != rightSum {
		t.Fatalf("replay checksum %s != %s", leftSum, rightSum)
	}
}

func TestStraightMissileExpiresWithoutContact(t *testing.T) {
	skills, _ := testRegistries(t)
	missiles, _ := NewRegistry(Definition{SkillID: 42, SpeedPerTick: 1, MaxRange: 20, LifetimeTicks: 2, CollisionRadius: 0.1, PhysicalDamage: gamecombat.MustWhole(3)})
	engine := gameecs.New()
	defer engine.Close()
	registerTestSystems(t, engine, skills, missiles)
	stores, _ := registerStores(engine)
	requests := store(t, engine, gameskill.CastRequestComponent)
	vitals := store(t, engine, "d2.player.vitals")
	caster := engine.World().MustCreateEntity()
	set(t, stores.controls, caster, map[string]any{"player": "alpha"})
	set(t, stores.positions, caster, map[string]any{"x": 0.0, "y": 0.0})
	set(t, stores.locations, caster, map[string]any{"act": int64(1), "level_id": int64(2)})
	set(t, vitals, caster, map[string]any{"health": int64(10), "max_health": int64(10), "mana": int64(10), "max_mana": int64(10)})
	set(t, requests, caster, map[string]any{"player": "alpha", "side": "right", "skill_id": int64(42), "skill_level": int64(1), "target_x": 10.0, "target_y": 0.0, "target_id": "", "request_tick": int64(1)})
	for stores.missiles.Len() == 0 {
		step(t, engine)
	}
	for stores.missiles.Len() > 0 {
		step(t, engine)
	}
	if kinds := eventKinds(stores.events); kinds[len(kinds)-1] != EventExpired {
		t.Fatalf("events=%v", kinds)
	}
}

func testRegistries(t *testing.T) (gameskill.Registry, Registry) {
	t.Helper()
	skills, err := gameskill.NewRegistry(gameskill.Definition{SkillID: 42, Behavior: gameskill.BehaviorStraightMissile, TargetPolicy: gameskill.TargetPoint, EffectDelay: 1, CompleteDelay: 2})
	if err != nil {
		t.Fatal(err)
	}
	missiles, err := NewRegistry(Definition{SkillID: 42, SpeedPerTick: 1, MaxRange: 2, LifetimeTicks: 10, CollisionRadius: 0.1, DamageChannel: gamecombat.Fire, MinimumDamage: gamecombat.MustWhole(3), MaximumDamage: gamecombat.MustWhole(3), Presentation: Presentation{MissileID: "test", DCC: "data/global/missiles/test.dcc", Palette: "data/global/palette/units/pal.dat", TravelSound: "test_travel", HitSound: "test_hit", Directions: 8, FramesPerSecond: 25, Loop: true}})
	if err != nil {
		t.Fatal(err)
	}
	return skills, missiles
}

func registerTestSystems(t *testing.T, engine *gameecs.Engine, skills gameskill.Registry, missiles Registry) {
	t.Helper()
	if err := gameskill.RegisterCastLifecycle(engine, skills); err != nil {
		t.Fatal(err)
	}
	if err := Register(engine, missiles); err != nil {
		t.Fatal(err)
	}
}
func store(t *testing.T, engine *gameecs.Engine, name string) *akara.DynamicStore {
	t.Helper()
	result, found := akara.GetDynamicStore(engine.World(), name)
	if !found {
		t.Fatalf("store %q is not registered", name)
	}
	return result
}
func registerMonsterStats(t *testing.T, engine *gameecs.Engine) *akara.DynamicStore {
	t.Helper()
	store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.monster.stats", Version: 1, Fields: []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "defense", Kind: akara.FieldInt64}, {Name: "attack_rating", Kind: akara.FieldInt64}, {Name: "physical_min", Kind: akara.FieldInt64}, {Name: "physical_max", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}}})
	if err != nil {
		t.Fatal(err)
	}
	return store
}
func set(t *testing.T, store *akara.DynamicStore, entity akara.Entity, values map[string]any) {
	t.Helper()
	if _, err := store.Set(entity, values); err != nil {
		t.Fatal(err)
	}
}
func step(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	if err := engine.Update(time.Second / 25); err != nil {
		t.Fatal(err)
	}
}
func assertHitResult(t *testing.T, engine *gameecs.Engine, target akara.Entity) {
	t.Helper()
	stores, _ := registerStores(engine)
	monsters := registerMonsterStats(t, engine)
	if stores.missiles.Len() != 0 {
		t.Fatalf("missiles=%d", stores.missiles.Len())
	}
	stats, _ := monsters.Get(target)
	health, _ := stats.Get("health")
	if health != gamecombat.MustWhole(7).Raw() {
		t.Fatalf("health=%v", health)
	}
	kinds := eventKinds(stores.events)
	if len(kinds) != 2 || kinds[0] != EventSpawned || kinds[1] != EventHit {
		t.Fatalf("events=%v", kinds)
	}
	combatEvents := store(t, engine, gamecombat.CombatEvent)
	if combatEvents.Len() != 2 {
		t.Fatalf("combat events=%d", combatEvents.Len())
	}
	for _, entity := range combatEvents.Entities() {
		event, _ := combatEvents.Get(entity)
		kind, _ := event.Get("kind")
		if kind == gamecombat.EventDamageApplied {
			channel, _ := event.Get("damage_channel")
			physical, _ := event.Get("physical")
			if channel != "fire" || physical != int64(0) {
				t.Fatalf("fire missile event channel=%v physical=%v", channel, physical)
			}
		}
	}
}
func eventKinds(events *akara.DynamicStore) []string {
	result := make([]string, 0, events.Len())
	for _, entity := range events.Entities() {
		event, _ := events.Get(entity)
		kind, _ := event.Get("kind")
		result = append(result, kind.(string))
	}
	return result
}
