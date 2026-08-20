package clientapp

import (
	"path/filepath"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// TestLoadPlayerProfileStartsEmptyAndRestoresPersistedSelection covers first-run
// creation and durable selection recovery.
func TestLoadPlayerProfileStartsEmptyAndRestoresPersistedSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")

	store, writablePath, err := loadPlayerProfile(path, nil)
	if err != nil {
		t.Fatal(err)
	}

	if writablePath != path || len(store.Characters()) != 0 {
		t.Fatalf("new profile store/path = %#v %q", store.Characters(), writablePath)
	}

	if err := store.Create(d2save.Character{ID: "hero", Name: "Hero"}); err != nil {
		t.Fatal(err)
	}

	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}

	if err := d2save.WriteProfileFile(path, store.Profile()); err != nil {
		t.Fatal(err)
	}

	restored, _, err := loadPlayerProfile(path, nil)
	if err != nil {
		t.Fatal(err)
	}

	selected, ok := restored.Selected()
	if !ok || selected.Name != "Hero" {
		t.Fatalf("restored selection = %#v", selected)
	}
}

// TestDevelopmentFixturesCannotOverwritePlayerProfile ensures disposable lab characters disable persistence ownership.
func TestDevelopmentFixturesCannotOverwritePlayerProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")

	store, writablePath, err := loadPlayerProfile(path, []d2save.Character{{ID: "fixture"}})
	if err != nil {
		t.Fatal(err)
	}

	if writablePath != "" || store.Characters()[0].ID != "fixture" {
		t.Fatalf("fixture store/path = %#v %q", store.Characters(), writablePath)
	}
}

// TestSelfHostedCanonicalCharacterRoundTripsThroughPlayerProfile preserves local-only fields around authority updates.
func TestSelfHostedCanonicalCharacterRoundTripsThroughPlayerProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	baseline := d2save.Character{
		ID: "hero", Name: "Before", Class: "Amazon", Level: 1, Expansion: true,
		Appearance: &d2save.Appearance{COF: "hero.cof", Components: map[string]string{"HD": "head.dcc"}},
		Stats:      &d2save.Stats{Strength: 25, Health: 10, MaxHealth: 20},
	}

	store := d2save.New(baseline)
	if err := store.Select(baseline.ID); err != nil {
		t.Fatal(err)
	}

	hud := playeradapter.HUD{
		Version:  playeradapter.HUDVersion,
		Player:   playeradapter.HUDIdentity{CharacterID: baseline.ID, Name: "After", Class: "Amazon"},
		Vitals:   playeradapter.HUDVitals{Health: 18, MaxHealth: 30, Mana: 9, MaxMana: 12},
		Progress: playeradapter.HUDProgress{Level: 4, Experience: 200},
		Combat:   playeradapter.HUDCombat{Defense: 7},
	}
	if err := updateSelectedCharacter(store, hud); err != nil {
		t.Fatal(err)
	}

	if err := d2save.WriteProfileFile(path, store.Profile()); err != nil {
		t.Fatal(err)
	}

	restored, _, err := loadPlayerProfile(path, nil)
	if err != nil {
		t.Fatal(err)
	}

	character, selected := restored.Selected()
	if !selected || character.Name != "After" || character.Level != 4 ||
		character.Stats.Health != 18 || character.Stats.Defense != 7 {
		t.Fatalf("restored canonical character = %#v selected=%t", character, selected)
	}

	if !character.Expansion || character.Stats.Strength != 25 ||
		character.Appearance.COF != "hero.cof" || character.Appearance.Components["HD"] != "head.dcc" {
		t.Fatalf("player-profile-only fields drifted: %#v", character)
	}
}

// TestSelfHostedProjectionCannotReplaceSelectedIdentity treats the selected save identity as an admission invariant.
func TestSelfHostedProjectionCannotReplaceSelectedIdentity(t *testing.T) {
	store := d2save.New(d2save.Character{ID: "hero"})
	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}

	err := updateSelectedCharacter(store, playeradapter.HUD{
		Version: playeradapter.HUDVersion,
		Player:  playeradapter.HUDIdentity{CharacterID: "attacker"},
	})
	if err == nil {
		t.Fatal("network projection replaced selected profile identity")
	}

	selected, _ := store.Selected()
	if selected.ID != "hero" {
		t.Fatalf("selected identity changed to %q", selected.ID)
	}
}

// TestSinglePlayerCanonicalCharacterUpdatesSelectedProfile persists authority facts without dropping baseline stats.
func TestSinglePlayerCanonicalCharacterUpdatesSelectedProfile(t *testing.T) {
	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })

	installOfflineCharacterCheckpoint(t, engine)

	store := d2save.New(d2save.Character{
		ID: "hero", Name: "Before", Class: "Amazon", Level: 1,
		Stats: &d2save.Stats{Strength: 20, Health: 10, MaxHealth: 10},
	})
	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}

	if err := persistOfflineCharacter(store, session, "local-player"); err != nil {
		t.Fatal(err)
	}

	character, selected := store.Selected()
	if !selected || character.Name != "Canonical" || character.Level != 3 ||
		character.Stats.Health != 14 || character.Stats.Strength != 20 {
		t.Fatalf("single-player persisted character = %#v selected=%t", character, selected)
	}
}

// offlineCheckpointComponent pairs a projection schema with the canonical values installed for one fixture entity.
type offlineCheckpointComponent struct {
	schema akara.Schema
	values map[string]any
}

// installOfflineCharacterCheckpoint materializes the minimum authority components used by character projection.
func installOfflineCharacterCheckpoint(t *testing.T, engine *gameecs.Engine) {
	t.Helper()

	entity := engine.World().MustCreateEntity()
	for _, component := range offlineCheckpointComponents() {
		store, err := akara.RegisterSchema(engine.World(), component.schema)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := store.Set(entity, component.values); err != nil {
			t.Fatal(err)
		}
	}
}

// offlineCheckpointComponents describes projection inputs by domain so failures identify the missing authority fact.
func offlineCheckpointComponents() []offlineCheckpointComponent {
	return []offlineCheckpointComponent{
		{
			schema: checkpointSchema("d2legacy.player.identity",
				akara.Field{Name: "character_id", Kind: akara.FieldString},
				akara.Field{Name: "player", Kind: akara.FieldString},
				akara.Field{Name: "name", Kind: akara.FieldString},
				akara.Field{Name: "class", Kind: akara.FieldString},
			),
			values: map[string]any{
				"character_id": "hero", "player": "local-player", "name": "Canonical", "class": "Amazon",
			},
		},
		{
			schema: checkpointSchema("d2legacy.player.vitals",
				akara.Field{Name: "health", Kind: akara.FieldInt64},
				akara.Field{Name: "max_health", Kind: akara.FieldInt64},
				akara.Field{Name: "mana", Kind: akara.FieldInt64},
				akara.Field{Name: "max_mana", Kind: akara.FieldInt64},
			),
			values: map[string]any{
				"health": int64(14), "max_health": int64(20), "mana": int64(8), "max_mana": int64(12),
			},
		},
		{
			schema: checkpointSchema("d2legacy.player.progress",
				akara.Field{Name: "level", Kind: akara.FieldInt64},
				akara.Field{Name: "experience", Kind: akara.FieldInt64},
				akara.Field{Name: "unspent_skill_points", Kind: akara.FieldInt64},
			),
			values: map[string]any{
				"level": int64(3), "experience": int64(100), "unspent_skill_points": int64(0),
			},
		},
		{
			schema: checkpointSchema("d2legacy.player.combat_stats",
				akara.Field{Name: "attack_rating", Kind: akara.FieldInt64},
				akara.Field{Name: "defense", Kind: akara.FieldInt64},
			),
			values: map[string]any{"attack_rating": int64(10), "defense": int64(4)},
		},
		{
			schema: checkpointSchema("d2legacy.world.position",
				akara.Field{Name: "x", Kind: akara.FieldFloat64},
				akara.Field{Name: "y", Kind: akara.FieldFloat64},
			),
			values: map[string]any{"x": 10.0, "y": 20.0},
		},
		{
			schema: checkpointSchema("d2legacy.world.location",
				akara.Field{Name: "act", Kind: akara.FieldInt64},
				akara.Field{Name: "level_id", Kind: akara.FieldInt64},
			),
			values: map[string]any{"act": int64(1), "level_id": int64(40)},
		},
	}
}

// checkpointSchema applies the version shared by all synthetic projection components.
func checkpointSchema(name string, fields ...akara.Field) akara.Schema {
	return akara.Schema{Name: name, Version: 1, Fields: fields}
}
