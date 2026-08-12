package monster

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gravestench/akara"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameloot "github.com/gravestench/dark-magic/internal/game/loot"
	gameownedunit "github.com/gravestench/dark-magic/internal/game/ownedunit"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/game/targeting"
)

func TestDeathTransactionCommitsConsequencesAndReplays(t *testing.T) {
	policy := DeathPolicy{WorldSeed: 73, Loot: gameloot.Catalog{
		"fallen": {Name: "fallen", Picks: 1, Entries: []gameloot.Entry{{Code: "gld", Weight: 1}}},
	}}
	engine := deathFixture(t, policy)
	before, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Update(time.Millisecond); err != nil {
		t.Fatal(err)
	}
	assertDeathConsequences(t, engine)
	want, err := engine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	wantChecksum, _ := want.Checksum()

	replayed, err := gameecs.RestoreSnapshot(before)
	if err != nil {
		t.Fatal(err)
	}
	defer replayed.Close()
	if err := RegisterDeath(replayed, policy); err != nil {
		t.Fatal(err)
	}
	if err := replayed.Update(time.Millisecond); err != nil {
		t.Fatal(err)
	}
	got, err := replayed.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	gotChecksum, _ := got.Checksum()
	if gotChecksum != wantChecksum {
		t.Fatalf("replayed checksum = %s, want %s", gotChecksum, wantChecksum)
	}
}

func deathFixture(t *testing.T, policy DeathPolicy) *gameecs.Engine {
	t.Helper()
	stats, graphics, level := ordinaryFixture()
	stats.TreasureClass1 = "fallen"
	definition, err := JoinDefinition(stats, graphics, &level, Normal)
	if err != nil {
		t.Fatal(err)
	}
	spawn, err := NewSpawn("blood-moor:fallen:1", definition, 42, 12, 13, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(spawn)
	engine := gameecs.New()
	t.Cleanup(func() { _ = engine.Close() })
	if err := materialize(engine, simulation.Command{Tick: 1, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := RegisterDeath(engine, policy); err != nil {
		t.Fatal(err)
	}
	stores, err := registerDeathStores(engine)
	if err != nil {
		t.Fatal(err)
	}
	player, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stores.selectable.Set(player, map[string]any{"id": "player:hero", "kind": targeting.KindPlayer, "label": "Hero", "owner": "hero", "radius": 1.0, "priority": int64(10)}); err != nil {
		t.Fatal(err)
	}
	if _, err := stores.progress.Set(player, map[string]any{"level": int64(1), "experience": int64(5)}); err != nil {
		t.Fatal(err)
	}
	minion, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stores.selectable.Set(minion, map[string]any{"id": "monster:skeleton", "kind": targeting.KindHostile, "label": "Skeleton", "owner": "hero", "radius": 1.0, "priority": int64(20)}); err != nil {
		t.Fatal(err)
	}
	if _, err := gameownedunit.Attach(engine.World(), gameownedunit.Relation{Unit: minion, Owner: player, OwnerID: "player:hero", UltimateOwnerID: "player:hero", CreatedTick: 1}, gameownedunit.Category{ID: "skeleton", BaseMax: 1, Replacement: gameownedunit.Reject}); err != nil {
		t.Fatal(err)
	}
	event, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stores.combatEvents.Set(event, map[string]any{"kind": gamecombat.EventUnitDied, "tick": int64(1), "attacker_id": "monster:skeleton", "target_id": "monster:blood-moor:fallen:1", "hit": true, "physical": int64(10), "remaining_health": int64(0)}); err != nil {
		t.Fatal(err)
	}
	duplicate, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stores.combatEvents.Set(duplicate, map[string]any{"kind": gamecombat.EventUnitDied, "tick": int64(1), "attacker_id": "monster:skeleton", "target_id": "monster:blood-moor:fallen:1", "hit": true, "physical": int64(10), "remaining_health": int64(0)}); err != nil {
		t.Fatal(err)
	}
	return engine
}

func assertDeathConsequences(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	deaths, _ := akara.GetDynamicStore(engine.World(), DeathTransaction)
	if deaths.Len() != 1 {
		t.Fatalf("death transactions = %d", deaths.Len())
	}
	death, _ := deaths.Get(deaths.Entities()[0])
	drops, _ := death.Get("drops")
	if drops != `[{"code":"gld","path":["fallen"]}]` {
		t.Fatalf("drops = %v", drops)
	}
	active, _ := death.Get("active")
	if active != false {
		t.Fatalf("active = %v", active)
	}
	killer, _ := death.Get("killer_id")
	credited, _ := death.Get("credited_id")
	if killer != "monster:skeleton" || credited != "player:hero" {
		t.Fatalf("killer=%v credited=%v", killer, credited)
	}
	events, _ := akara.GetDynamicStore(engine.World(), DeathEvent)
	if events.Len() != 4 {
		t.Fatalf("semantic events = %d", events.Len())
	}
	progress, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.progress")
	playerProgress, _ := progress.Get(progress.Entities()[0])
	experience, _ := playerProgress.Get("experience")
	if experience != int64(5+ordinaryExperience(t)) {
		t.Fatalf("experience = %v", experience)
	}
	selectables, _ := akara.GetDynamicStore(engine.World(), targeting.Component)
	if selectables.Len() != 2 {
		t.Fatalf("selectables = %d", selectables.Len())
	}
	colliders, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.collider")
	if colliders.Len() != 0 {
		t.Fatalf("colliders = %d", colliders.Len())
	}
	appearance, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.appearance")
	component, _ := appearance.Get(deaths.Entities()[0])
	mode, _ := component.Get("mode")
	if mode != "DT" {
		t.Fatalf("mode = %v", mode)
	}
}

func ordinaryExperience(t *testing.T) int64 {
	t.Helper()
	stats, graphics, level := ordinaryFixture()
	definition, err := JoinDefinition(stats, graphics, &level, Normal)
	if err != nil {
		t.Fatal(err)
	}
	return definition.Experience
}
