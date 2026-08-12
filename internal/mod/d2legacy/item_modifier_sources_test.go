package d2legacy

import (
	"testing"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestEquippedItemModifierSourcesPreserveAffixAndSocketProvenance(t *testing.T) {
	initial := itemAcceptanceBootstrap()
	items := initial["d2legacy.items"].(map[string]any)["items"].([]any)
	for _, raw := range items {
		item := raw.(map[string]any)
		switch item["id"] {
		case "weapon":
			item["stat_modifiers"] = []any{
				map[string]any{"source_id": "precision", "source_kind": "affix", "stat": "attack_rating", "operation": "add", "value": 75.0, "order": 10.0},
				map[string]any{"source_id": "socket-rune", "source_kind": "socket", "stat": "attack_rating", "operation": "add", "value": 25.0, "order": 20.0},
			}
		case "armor":
			item["stat_modifiers"] = []any{
				map[string]any{"source_id": "sturdy", "source_kind": "affix", "stat": "defense", "operation": "add", "value": 7.0, "order": 10.0},
				map[string]any{"source_id": "socket-jewel", "source_kind": "socket", "stat": "defense", "operation": "add", "value": 3.0, "order": 20.0},
			}
		}
	}

	engine := gameecs.New()
	defer engine.Close()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	authority, err := StartWithConfig(t.Context(), content.D2Legacy(), fixtureRecords{}, engine, session,
		Config{Seed: 94, InitialData: initial})
	if err != nil {
		t.Fatal(err)
	}
	defer authority.Stop(t.Context())

	commands := []simulation.Command{
		{Tick: 1, Player: "system", Authority: simulation.AuthoritySystem, Sequence: 1,
			Kind: "system.player.enter", Payload: itemAcceptancePlayerPayload(t)},
		itemCommand(t, 1, 1, "item.vendor_sell", map[string]any{
			"item_id": "sale", "vendor": "charsi", "category": "weap",
		}),
		itemCommand(t, 1, 2, "item.move", map[string]any{
			"item_id": "weapon", "destination": map[string]any{"container": "held"},
		}),
		itemCommand(t, 2, 3, "item.move", map[string]any{
			"item_id": "weapon", "place_held": true,
			"destination": map[string]any{"container": "equipment", "slot": "rarm", "weapon_set": 0},
		}),
		itemCommand(t, 3, 4, "item.move", map[string]any{
			"item_id": "armor", "destination": map[string]any{"container": "held"},
		}),
		itemCommand(t, 4, 5, "item.move", map[string]any{
			"item_id": "armor", "place_held": true,
			"destination": map[string]any{"container": "equipment", "slot": "head"},
		}),
	}
	for _, command := range commands {
		if err := session.Submit(command); err != nil {
			t.Fatal(err)
		}
	}
	for range 5 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}

	assertPlayerCombatStat(t, engine, "attack_rating", 1100, 3)
	assertPlayerCombatStat(t, engine, "defense", 50, 3)
	assertStatSourceValue(t, engine, "attack_rating", "equipment:attack:weapon", 900)
	assertStatSourceValue(t, engine, "attack_rating", "equipment:modifier:attack_rating:weapon:affix:10:precision", 75)
	assertStatSourceValue(t, engine, "attack_rating", "equipment:modifier:attack_rating:weapon:socket:20:socket-rune", 25)
	assertStatSourceValue(t, engine, "defense", "equipment:defense:armor", 40)
	assertStatSourceValue(t, engine, "defense", "equipment:modifier:defense:armor:affix:10:sturdy", 7)
	assertStatSourceValue(t, engine, "defense", "equipment:modifier:defense:armor:socket:20:socket-jewel", 3)

	if err := session.Submit(itemCommand(t, 6, 6, "item.move", map[string]any{
		"item_id": "weapon", "destination": map[string]any{"container": "inventory", "x": 2, "y": 0},
	})); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	assertPlayerCombatStat(t, engine, "attack_rating", 100, 0)
	assertPlayerCombatStat(t, engine, "defense", 50, 3)

	if err := session.Submit(itemCommand(t, 8, 7, "item.move", map[string]any{
		"item_id": "armor", "destination": map[string]any{"container": "inventory", "x": 4, "y": 0},
	})); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := session.Step(); err != nil {
			t.Fatal(err)
		}
	}
	assertPlayerCombatStat(t, engine, "defense", 0, 0)
}

func assertStatSourceValue(t *testing.T, engine *gameecs.Engine, statName, sourceID string, wanted int64) {
	t.Helper()
	store, _ := akara.GetDynamicStore(engine.World(), "d2legacy.stat.source")
	for _, entity := range store.Entities() {
		source, _ := store.Get(entity)
		stat, _ := source.Get("stat")
		id, _ := source.Get("source_id")
		if stat != statName || id != sourceID {
			continue
		}
		value, _ := source.Get("value")
		if value != wanted {
			t.Fatalf("%s source %q = %v, want %d", statName, sourceID, value, wanted)
		}
		return
	}
	t.Fatalf("%s source %q not found", statName, sourceID)
}
