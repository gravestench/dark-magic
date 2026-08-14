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
	if !selected || character.Name != "After" || character.Level != 4 || character.Stats.Health != 18 || character.Stats.Defense != 7 {
		t.Fatalf("restored canonical character = %#v selected=%t", character, selected)
	}
	if !character.Expansion || character.Stats.Strength != 25 || character.Appearance.COF != "hero.cof" || character.Appearance.Components["HD"] != "head.dcc" {
		t.Fatalf("player-profile-only fields drifted: %#v", character)
	}
}

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
	if !selected || character.Name != "Canonical" || character.Level != 3 || character.Stats.Health != 14 || character.Stats.Strength != 20 {
		t.Fatalf("single-player persisted character = %#v selected=%t", character, selected)
	}
}

func installOfflineCharacterCheckpoint(t *testing.T, engine *gameecs.Engine) {
	t.Helper()
	register := func(name string, fields []akara.Field) *akara.DynamicStore {
		store, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: name, Version: 1, Fields: fields})
		if err != nil {
			t.Fatal(err)
		}
		return store
	}
	entity := engine.World().MustCreateEntity()
	stores := map[*akara.DynamicStore]map[string]any{
		register("d2legacy.player.identity", []akara.Field{{Name: "character_id", Kind: akara.FieldString}, {Name: "player", Kind: akara.FieldString}, {Name: "name", Kind: akara.FieldString}, {Name: "class", Kind: akara.FieldString}}): {"character_id": "hero", "player": "local-player", "name": "Canonical", "class": "Amazon"},
		register("d2legacy.player.vitals", []akara.Field{{Name: "health", Kind: akara.FieldInt64}, {Name: "max_health", Kind: akara.FieldInt64}, {Name: "mana", Kind: akara.FieldInt64}, {Name: "max_mana", Kind: akara.FieldInt64}}):      {"health": int64(14), "max_health": int64(20), "mana": int64(8), "max_mana": int64(12)},
		register("d2legacy.player.progress", []akara.Field{{Name: "level", Kind: akara.FieldInt64}, {Name: "experience", Kind: akara.FieldInt64}, {Name: "unspent_skill_points", Kind: akara.FieldInt64}}):                                 {"level": int64(3), "experience": int64(100), "unspent_skill_points": int64(0)},
		register("d2legacy.player.combat_stats", []akara.Field{{Name: "attack_rating", Kind: akara.FieldInt64}, {Name: "defense", Kind: akara.FieldInt64}}):                                                                                {"attack_rating": int64(10), "defense": int64(4)},
		register("d2legacy.world.position", []akara.Field{{Name: "x", Kind: akara.FieldFloat64}, {Name: "y", Kind: akara.FieldFloat64}}):                                                                                                   {"x": 10.0, "y": 20.0},
		register("d2legacy.world.location", []akara.Field{{Name: "act", Kind: akara.FieldInt64}, {Name: "level_id", Kind: akara.FieldInt64}}):                                                                                              {"act": int64(1), "level_id": int64(40)},
	}
	for store, values := range stores {
		if _, err := store.Set(entity, values); err != nil {
			t.Fatal(err)
		}
	}
}
