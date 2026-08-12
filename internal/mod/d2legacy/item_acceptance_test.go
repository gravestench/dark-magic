package d2legacy

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestItemReconnectRecoveryEquipmentVendorAndServiceRestoreIdentically(t *testing.T) {
	ctx := context.Background()
	initial := itemAcceptanceBootstrap()
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	authority, err := StartWithConfig(ctx, content.D2Legacy(), fixtureRecords{}, engine, session, Config{Seed: 91, InitialData: initial})
	if err != nil {
		t.Fatal(err)
	}

	if err := session.Submit(simulation.Command{
		Tick: 1, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: itemAcceptancePlayerPayload(t),
	}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	replay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	checkpoint := replay.Checkpoints[len(replay.Checkpoints)-1]
	assertItemPlacement(t, engine, "sale", "held", "")

	submitItemAcceptanceCommands(t, session)
	itemAcceptanceSteps(t, session)
	originalReplay, err := session.Replay()
	if err != nil {
		t.Fatal(err)
	}
	original := originalReplay.Checkpoints[len(originalReplay.Checkpoints)-1]
	assertItemAcceptanceOutcome(t, engine)
	if err := authority.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	if err := session.Close(); err != nil {
		t.Fatal(err)
	}
	if err := engine.Close(); err != nil {
		t.Fatal(err)
	}

	restoredEngine, err := gameecs.RestoreSnapshot(*checkpoint.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	defer restoredEngine.Close()
	restoredSession, err := gamesession.New(restoredEngine, gamesession.Config{CheckpointInterval: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer restoredSession.Close()
	restored, err := StartWithConfig(ctx, content.D2Legacy(), fixtureRecords{}, restoredEngine, restoredSession, Config{Seed: 91, InitialData: initial, Restore: checkpoint.Participants})
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Stop(ctx)

	// The held sale item survived the disconnect/checkpoint boundary before any
	// transaction resumes against the reconstructed Lua runtime.
	assertItemPlacement(t, restoredEngine, "sale", "held", "")
	submitItemAcceptanceCommands(t, restoredSession)
	itemAcceptanceSteps(t, restoredSession)
	restoredReplay, err := restoredSession.Replay()
	if err != nil {
		t.Fatal(err)
	}
	continued := restoredReplay.Checkpoints[len(restoredReplay.Checkpoints)-1]
	if continued.Checksum != original.Checksum {
		t.Fatalf("restored item matrix checksum = %s, want %s", continued.Checksum, original.Checksum)
	}
	assertItemAcceptanceOutcome(t, restoredEngine)
}

func TestEquippedAttackRatingSourceIsAddedAndRemoved(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), fixtureRecords{}, engine, session,
		Config{Seed: 92, InitialData: itemAcceptanceBootstrap()})
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(t.Context())
	if err := session.Submit(simulation.Command{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.player.enter", Payload: itemAcceptancePlayerPayload(t)}); err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(itemCommand(t, 1, 1, "item.vendor_sell", map[string]any{
		"item_id": "sale", "vendor": "charsi", "category": "weap",
	})); err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(itemCommand(t, 1, 2, "item.move", map[string]any{
		"item_id": "weapon", "destination": map[string]any{"container": "held"},
	})); err != nil {
		t.Fatal(err)
	}
	if err := session.Submit(itemCommand(t, 2, 3, "item.move", map[string]any{
		"item_id": "weapon", "place_held": true,
		"destination": map[string]any{"container": "equipment", "slot": "rarm", "weapon_set": 0},
	})); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	assertPlayerAttackRating(t, engine, 1000, 1)
	if err := session.Submit(itemCommand(t, 4, 4, "item.move", map[string]any{
		"item_id": "weapon", "destination": map[string]any{"container": "inventory", "x": 2, "y": 0},
	})); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	assertPlayerAttackRating(t, engine, 100, 0)
}

func assertPlayerAttackRating(t *testing.T, engine *gameecs.Engine, wanted int64, sources int) {
	t.Helper()
	stats, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.combat_stats")
	value, _ := stats.Get(stats.Entities()[0])
	rating, _ := value.Get("attack_rating")
	if rating != wanted {
		t.Fatalf("attack rating = %v, want %d", rating, wanted)
	}
	store, _ := akara.GetDynamicStore(engine.World(), "d2legacy.stat.source")
	count := 0
	for _, entity := range store.Entities() {
		source, _ := store.Get(entity)
		stat, _ := source.Get("stat")
		if stat == "attack_rating" {
			count++
		}
	}
	if count != sources {
		t.Fatalf("attack-rating sources = %d, want %d", count, sources)
	}
}

func submitItemAcceptanceCommands(t *testing.T, session *gamesession.Session) {
	t.Helper()
	commands := []simulation.Command{
		itemCommand(t, 2, 1, "item.vendor_sell", map[string]any{
			"item_id": "sale", "vendor": "charsi", "category": "weap",
		}),
		itemCommand(t, 2, 2, "item.move", map[string]any{
			"item_id": "weapon", "destination": map[string]any{"container": "held"},
		}),
		itemCommand(t, 2, 3, "item.move", map[string]any{
			"item_id": "corpse-boots", "destination": map[string]any{"container": "inventory", "x": 0, "y": 0},
		}),
		itemCommand(t, 2, 4, "item.service_complete", map[string]any{"service": "socket"}),
		itemCommand(t, 3, 5, "item.move", map[string]any{
			"item_id": "weapon", "place_held": true,
			"destination": map[string]any{"container": "equipment", "slot": "rarm", "weapon_set": 0},
		}),
	}
	for _, command := range commands {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
}

func itemCommand(t *testing.T, tick, sequence uint64, kind string, payload map[string]any) simulation.Command {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return simulation.Command{Tick: tick, Player: "alice", Authority: simulation.AuthorityPlayer, Sequence: sequence, Kind: kind, Payload: data}
}

func itemAcceptanceSteps(t *testing.T, session *gamesession.Session) {
	t.Helper()
	for range 2 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
}

func assertItemAcceptanceOutcome(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	assertItemPlacement(t, engine, "sale", "vendor", "weap")
	assertItemPlacement(t, engine, "corpse-boots", "inventory", "")
	assertItemPlacement(t, engine, "weapon", "equipment", "rarm")
	items, _ := akara.GetDynamicStore(engine.World(), "d2legacy.item.identity")
	found := map[string]any{}
	for _, entity := range items.Entities() {
		item, _ := items.Get(entity)
		id, _ := item.Get("id")
		found[id.(string)] = item
	}
	if _, exists := found["service-material"]; exists {
		t.Fatal("completed service retained its consumed material")
	}
	target := found["service-target"].(*akara.DynamicComponent)
	applied, _ := target.Get("applied_services")
	if applied != "socket" {
		t.Fatalf("applied services = %v, want socket", applied)
	}
	layouts, _ := akara.GetDynamicStore(engine.World(), "d2legacy.items.layout")
	layout, _ := layouts.Get(layouts.Entities()[0])
	gold, _ := layout.Get("carried_gold")
	if gold != int64(1025) {
		t.Fatalf("carried gold = %v, want 1025 after sale and service", gold)
	}
	profiles, _ := akara.GetDynamicStore(engine.World(), "d2legacy.combat.melee_profile")
	profile, _ := profiles.Get(profiles.Entities()[0])
	rangeValue, _ := profile.Get("range")
	minimum, _ := profile.Get("physical_min")
	maximum, _ := profile.Get("physical_max")
	if rangeValue != float64(3) || minimum != int64(512) || maximum != int64(1024) {
		t.Fatalf("equipped melee profile = %v/%v-%v", rangeValue, minimum, maximum)
	}
}

func assertItemPlacement(t *testing.T, engine *gameecs.Engine, wanted, container, slot string) {
	t.Helper()
	items, _ := akara.GetDynamicStore(engine.World(), "d2legacy.item.identity")
	placements, _ := akara.GetDynamicStore(engine.World(), "d2legacy.item.placement")
	for _, entity := range items.Entities() {
		item, _ := items.Get(entity)
		id, _ := item.Get("id")
		if id != wanted {
			continue
		}
		placed, present := placements.Get(entity)
		if !present {
			t.Fatalf("item %q has no placement", wanted)
		}
		actualContainer, _ := placed.Get("container")
		actualSlot, _ := placed.Get("slot")
		if actualContainer != container || actualSlot != slot {
			t.Fatalf("item %q placement = %v/%v, want %s/%s", wanted, actualContainer, actualSlot, container, slot)
		}
		return
	}
	t.Fatalf("item %q not found", wanted)
}

func itemAcceptancePlayerPayload(t *testing.T) []byte {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"character_id": "hero", "player": "alice", "name": "Hero", "class": "Amazon",
		"level": 1, "experience": 0, "dexterity": 20, "defense": 0,
		"health": 50, "max_health": 50, "mana": 20, "max_mana": 20,
		"expansion": true, "hardcore": false, "cof": "", "palette": "units",
		"direction": 0, "mode": "NU", "x": 0, "y": 0,
		"world_width": 100, "world_height": 100, "act": 1, "level_id": 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func itemAcceptanceBootstrap() map[string]any {
	item := func(id, code, container string) map[string]any {
		return map[string]any{"id": id, "code": code, "width": 1.0, "height": 1.0, "container": container}
	}
	weapon := item("weapon", "ssd", "inventory")
	weapon["x"], weapon["y"] = 2.0, 0.0
	weapon["body_slots"] = "rarm,larm"
	weapon["melee_range"], weapon["physical_min"], weapon["physical_max"] = 3.0, 512.0, 1024.0
	weapon["melee_weapon_class"] = "1HS"
	weapon["attack_rating"] = 900.0
	sale := item("sale", "cap", "held")
	sale["base_cost"] = 100.0
	corpse := item("corpse-boots", "lbt", "corpse")
	target := item("service-target", "ssd", "quest_service")
	target["slot"] = "target"
	material := item("service-material", "r01", "quest_service")
	material["slot"] = "material"
	return map[string]any{
		"d2legacy.items": map[string]any{
			"owner": "alice", "inventory_width": 10.0, "inventory_height": 4.0,
			"stash_width": 6.0, "stash_height": 8.0, "cube_width": 3.0, "cube_height": 4.0,
			"belt_capacity": 4.0, "vendor_width": 10.0, "vendor_height": 10.0,
			"carried_gold": 1000.0, "stashed_gold": 0.0,
			"items": []any{sale, weapon, corpse, target, material},
			"trade_terms": map[string]any{"charsi": map[string]any{
				"buy_multiplier": 512.0, "sell_multiplier": 1024.0, "max_buy": 0.0,
			}},
			"service_rules": []any{map[string]any{
				"id": "socket", "target_slot": "target", "consume_slots": "material", "gold_cost": 25.0,
			}},
		},
	}
}
