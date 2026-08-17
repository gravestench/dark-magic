package player

import (
	"encoding/json"
	"math"
	"sort"
	"strings"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestProjectHUDSelectsAuthenticatedPlayerAndAllowlistedFields(t *testing.T) {
	snapshot := gameecs.Snapshot{Version: gameecs.SnapshotVersion, Tick: 9, Components: []gameecs.ComponentSnapshot{
		stringsComponent("d2legacy.player.identity", []string{"character_id", "player", "name", "class"},
			[]any{uint64(1), "character:a", "alice", "Alice", "Amazon"},
			[]any{uint64(2), "character:b", "bob", "Bob", "Barbarian"}),
		intsComponent("d2legacy.player.vitals", []string{"health", "max_health", "mana", "max_mana", "stamina", "max_stamina", "stamina_raw", "max_stamina_raw"}, 1, 25, 30, 12, 15, 42, 84, 42*256+7, 84*256),
		intsComponent("d2legacy.player.movement_stats", []string{"run_drain", "velocitypercent", "item_fastermovevelocity", "staminarecoverybonus", "item_staminadrainpct", "armor_run_drain"}, 1, 20, -5, 30, 10, 25, 5),
		intsComponent("d2legacy.player.progress", []string{"level", "experience", "unspent_skill_points"}, 1, 7, 1234, 2),
		intsComponent("d2legacy.player.combat_stats", []string{"attack_rating", "defense"}, 1, 44, 21),
		floatsComponent("d2legacy.world.position", []string{"x", "y"}, 1, 10.5, 20.25),
		intsComponent("d2legacy.world.location", []string{"act", "level_id"}, 1, 1, 40),
		stringsComponent("d2legacy.player.belt", []string{"slot_1"}, []any{uint64(1), "secret-item"}),
	}}
	payload, err := ProjectHUD("alice", simulation.Checkpoint{Tick: 9, Snapshot: &snapshot})
	if err != nil {
		t.Fatal(err)
	}
	var view HUD
	if err := json.Unmarshal(payload, &view); err != nil {
		t.Fatal(err)
	}
	if view.Player.CharacterID != "character:a" || view.Player.Name != "Alice" || view.Vitals.Health != 25 || view.Vitals.StaminaRaw != 42*256+7 || view.Movement.ItemFasterMoveVelocity != 30 || view.Movement.RunDrain != 20 || view.Position.Y != 20.25 {
		t.Fatalf("HUD = %#v", view)
	}
	if view.Belt.Slots[0] != "secret-item" {
		t.Fatalf("owner-private belt was not projected: %#v", view.Belt)
	}
	if strings.Contains(string(payload), "Bob") {
		t.Fatalf("HUD leaked another/private field: %s", payload)
	}
}

func TestProjectHUDRejectsUnknownPlayer(t *testing.T) {
	snapshot := gameecs.Snapshot{Version: gameecs.SnapshotVersion, Components: []gameecs.ComponentSnapshot{
		stringsComponent("d2legacy.player.identity", []string{"player"}, []any{uint64(1), "alice"}),
	}}
	if _, err := ProjectHUD("mallory", simulation.Checkpoint{Snapshot: &snapshot}); err != ErrHUDPlayer {
		t.Fatalf("error = %v", err)
	}
}

func TestProjectPrivateViewSelectsOnlyAuthenticatedOwner(t *testing.T) {
	aliceItems := entityAndStringsComponent("d2legacy.item.identity", 11, 10, map[string]string{"id": "alice-item", "code": "ssd"})
	bobItems := entityAndStringsComponent("d2legacy.item.identity", 21, 20, map[string]string{"id": "bob-secret", "code": "rin"})
	aliceItems.Instances = append(aliceItems.Instances, bobItems.Instances...)
	snapshot := gameecs.Snapshot{Version: gameecs.SnapshotVersion, Tick: 4, Components: []gameecs.ComponentSnapshot{
		stringsComponent("d2legacy.items.layout", []string{"owner"}, []any{uint64(10), "alice"}, []any{uint64(20), "bob"}),
		aliceItems,
	}}
	view, err := ProjectPrivateView("alice", simulation.Checkpoint{Tick: 4, Snapshot: &snapshot})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(view)
	if len(view.Items.Items) != 1 || view.Items.Items[0].ID != "alice-item" || strings.Contains(string(encoded), "bob-secret") {
		t.Fatalf("private projection = %s", encoded)
	}
}

func entityAndStringsComponent(name string, entity, owner uint64, strings map[string]string) gameecs.ComponentSnapshot {
	component := gameecs.ComponentSnapshot{Name: name, Version: 1, Fields: []gameecs.FieldSnapshot{{Name: "owner", Kind: akara.FieldEntity}}}
	instance := gameecs.InstanceSnapshot{Entity: entity, Values: []gameecs.ValueSnapshot{{Entity: &owner}}}
	keys := make([]string, 0, len(strings))
	for key := range strings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := strings[key]
		component.Fields = append(component.Fields, gameecs.FieldSnapshot{Name: key, Kind: akara.FieldString})
		instance.Values = append(instance.Values, gameecs.ValueSnapshot{String: &value})
	}
	component.Instances = []gameecs.InstanceSnapshot{instance}
	return component
}

func stringsComponent(name string, names []string, rows ...[]any) gameecs.ComponentSnapshot {
	component := gameecs.ComponentSnapshot{Name: name, Version: 1}
	for _, name := range names {
		component.Fields = append(component.Fields, gameecs.FieldSnapshot{Name: name, Kind: akara.FieldString})
	}
	for _, row := range rows {
		instance := gameecs.InstanceSnapshot{Entity: row[0].(uint64)}
		for _, value := range row[1:] {
			copied := value.(string)
			instance.Values = append(instance.Values, gameecs.ValueSnapshot{String: &copied})
		}
		component.Instances = append(component.Instances, instance)
	}
	return component
}

func intsComponent(name string, names []string, entity uint64, values ...int64) gameecs.ComponentSnapshot {
	component := gameecs.ComponentSnapshot{Name: name, Version: 1}
	instance := gameecs.InstanceSnapshot{Entity: entity}
	for index, name := range names {
		component.Fields = append(component.Fields, gameecs.FieldSnapshot{Name: name, Kind: akara.FieldInt64})
		value := values[index]
		instance.Values = append(instance.Values, gameecs.ValueSnapshot{Int: &value})
	}
	component.Instances = []gameecs.InstanceSnapshot{instance}
	return component
}

func floatsComponent(name string, names []string, entity uint64, values ...float64) gameecs.ComponentSnapshot {
	component := gameecs.ComponentSnapshot{Name: name, Version: 1}
	instance := gameecs.InstanceSnapshot{Entity: entity}
	for index, name := range names {
		component.Fields = append(component.Fields, gameecs.FieldSnapshot{Name: name, Kind: akara.FieldFloat64})
		value := math.Float64bits(values[index])
		instance.Values = append(instance.Values, gameecs.ValueSnapshot{Float: &value})
	}
	component.Instances = []gameecs.InstanceSnapshot{instance}
	return component
}
