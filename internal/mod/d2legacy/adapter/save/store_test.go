package save

import "testing"

func TestSelectionAndDefensiveCharacterList(t *testing.T) {
	store := New()
	if err := store.Create(Character{ID: "hero", Name: "Hero", Class: "Amazon"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Create(Character{ID: "hero", Name: "Other", Class: "Amazon"}); err == nil {
		t.Fatal("expected duplicate error")
	}
	list := store.Characters()
	list[0].Name = "mutated"
	if store.Characters()[0].Name != "Hero" {
		t.Fatal("character list was not defensive")
	}
	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}
	selected, ok := store.Selected()
	if !ok || selected.ID != "hero" {
		t.Fatalf("selected = %#v, %v", selected, ok)
	}
}

func TestCharacterAppearanceIsDefensivelyCopied(t *testing.T) {
	t.Parallel()

	appearance := &Appearance{
		COF:        "data/global/chars/AM/COF/AMTNHTH.cof",
		Palette:    "data/global/Palette/ACT1/pal.dat",
		Direction:  2,
		Components: map[string]string{"HD": "data/global/chars/AM/HD/head.dcc"},
	}
	store := New(Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1, Appearance: appearance})
	appearance.Components["HD"] = "mutated"
	listed := store.Characters()
	listed[0].Appearance.Components["HD"] = "also-mutated"
	if got := store.Characters()[0].Appearance.Components["HD"]; got != "data/global/chars/AM/HD/head.dcc" {
		t.Fatalf("stored appearance component = %q", got)
	}
	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}
	selected, _ := store.Selected()
	selected.Appearance.Components["HD"] = "selected-mutated"
	if got := store.Characters()[0].Appearance.Components["HD"]; got != "data/global/chars/AM/HD/head.dcc" {
		t.Fatalf("stored appearance component after Selected mutation = %q", got)
	}
}

func TestCharacterStatsAreDefensivelyCopied(t *testing.T) {
	t.Parallel()

	stats := &Stats{Strength: 25, Health: 70, MaxHealth: 70}
	store := New(Character{ID: "hero", Name: "Hero", Class: "Amazon", Level: 1, Stats: stats})
	stats.Strength = 999
	listed := store.Characters()
	listed[0].Stats.Health = 0
	got := store.Characters()[0].Stats
	if got.Strength != 25 || got.Health != 70 {
		t.Fatalf("stored stats = %#v", got)
	}
}

func TestCreateRequiresOpaqueIdentityAndRejectsDuplicateID(t *testing.T) {
	store := New()
	if err := store.Create(Character{}); err == nil {
		t.Fatal("expected missing ID error")
	}
	if err := store.Create(Character{ID: "opaque", Name: "mod-owned value", Class: "anything"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Create(Character{ID: "opaque", Name: "different", Class: "different"}); err == nil {
		t.Fatal("expected duplicate ID error")
	}
}

func TestDeleteClearsSelectedCharacter(t *testing.T) {
	store := New(
		Character{ID: "amazon-hero", Name: "Hero", Class: "Amazon", Level: 1},
		Character{ID: "druid-wolf", Name: "Wolf", Class: "Druid", Level: 2},
	)
	if err := store.Select("amazon-hero"); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete("amazon-hero"); err != nil {
		t.Fatal(err)
	}
	if _, selected := store.Selected(); selected {
		t.Fatal("deleted character remained selected")
	}
	characters := store.Characters()
	if len(characters) != 1 || characters[0].ID != "druid-wolf" {
		t.Fatalf("characters after delete = %#v", characters)
	}
	if err := store.Delete("missing"); err == nil {
		t.Fatal("expected unknown-character error")
	}
}
