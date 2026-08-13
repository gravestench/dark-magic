package player

import (
	"encoding/json"
	"math"
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
		intsComponent("d2legacy.player.vitals", []string{"health", "max_health", "mana", "max_mana"}, 1, 25, 30, 12, 15),
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
	if view.Player.CharacterID != "character:a" || view.Player.Name != "Alice" || view.Vitals.Health != 25 || view.Position.Y != 20.25 {
		t.Fatalf("HUD = %#v", view)
	}
	if strings.Contains(string(payload), "secret-item") || strings.Contains(string(payload), "Bob") {
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
