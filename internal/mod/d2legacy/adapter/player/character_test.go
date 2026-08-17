package player

import (
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestProjectCharacterMergesCanonicalFieldsIntoDurableBaseline(t *testing.T) {
	snapshot := gameecs.Snapshot{Version: gameecs.SnapshotVersion, Tick: 9, Components: []gameecs.ComponentSnapshot{
		stringsComponent("d2legacy.player.identity", []string{"character_id", "player", "name", "class"},
			[]any{uint64(1), "character", "player", "Canonical Hero", "Amazon"}),
		intsComponent("d2legacy.player.vitals", []string{"health", "max_health", "mana", "max_mana", "stamina", "max_stamina", "stamina_raw", "max_stamina_raw"}, 1, 25, 30, 12, 15, 42, 84, 42*256, 84*256),
		intsComponent("d2legacy.player.progress", []string{"level", "experience", "unspent_skill_points"}, 1, 7, 1234, 2),
		intsComponent("d2legacy.player.combat_stats", []string{"attack_rating", "defense"}, 1, 44, 21),
		floatsComponent("d2legacy.world.position", []string{"x", "y"}, 1, 10, 20),
		intsComponent("d2legacy.world.location", []string{"act", "level_id"}, 1, 1, 40),
	}}
	baseline := d2save.Character{ID: "character", Name: "Old", Class: "Amazon", Level: 1,
		Expansion: true, Hardcore: true, Appearance: &d2save.Appearance{COF: "hero.cof", Components: map[string]string{"HD": "head.dcc"}},
		Stats: &d2save.Stats{Strength: 30, Vitality: 25, Health: 10}}
	result, err := ProjectCharacter("player", baseline, simulation.Checkpoint{Tick: 9, Snapshot: &snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Canonical Hero" || result.Level != 7 || result.Stats.Experience != 1234 || result.Stats.Health != 25 || result.Stats.Stamina != 42 || result.Stats.MaxStamina != 84 || result.Stats.Defense != 21 {
		t.Fatalf("projected character = %#v", result)
	}
	if !result.Expansion || !result.Hardcore || result.Stats.Strength != 30 || result.Stats.Vitality != 25 || result.Appearance.COF != "hero.cof" {
		t.Fatalf("baseline fields were not preserved: %#v", result)
	}
	result.Appearance.Components["HD"] = "changed"
	if baseline.Appearance.Components["HD"] != "head.dcc" {
		t.Fatal("projection aliases durable baseline")
	}
}

func TestProjectCharacterRejectsMismatchedCharacter(t *testing.T) {
	snapshot := gameecs.Snapshot{Version: gameecs.SnapshotVersion, Components: []gameecs.ComponentSnapshot{
		stringsComponent("d2legacy.player.identity", []string{"character_id", "player", "name", "class"},
			[]any{uint64(1), "other", "player", "Other", "Amazon"}),
		intsComponent("d2legacy.player.vitals", []string{"health", "max_health", "mana", "max_mana"}, 1, 1, 1, 1, 1),
		intsComponent("d2legacy.player.progress", []string{"level", "experience", "unspent_skill_points"}, 1, 1, 0, 0),
		intsComponent("d2legacy.player.combat_stats", []string{"attack_rating", "defense"}, 1, 1, 1),
		floatsComponent("d2legacy.world.position", []string{"x", "y"}, 1, 1, 1),
		intsComponent("d2legacy.world.location", []string{"act", "level_id"}, 1, 1, 1),
	}}
	if _, err := ProjectCharacter("player", d2save.Character{ID: "expected"}, simulation.Checkpoint{Snapshot: &snapshot}); err == nil {
		t.Fatal("mismatched checkpoint character was accepted")
	}
}
