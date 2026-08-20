package save

import "testing"

// TestSelectionAndDefensiveCharacterList protects unique identities, list copying, and explicit selection.
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

// TestUpdateSelectedPreservesRosterIdentityAndDefensiveOwnership blocks cross-character replacement and aliasing.
func TestUpdateSelectedPreservesRosterIdentityAndDefensiveOwnership(t *testing.T) {
	store := New(Character{ID: "hero", Name: "Before"}, Character{ID: "other", Name: "Other"})
	if err := store.Select("hero"); err != nil {
		t.Fatal(err)
	}

	updated := Character{ID: "hero", Name: "After", Stats: &Stats{Health: 12}}
	if err := store.UpdateSelected(updated); err != nil {
		t.Fatal(err)
	}

	updated.Stats.Health = 0

	selected, ok := store.Selected()
	if !ok || selected.Name != "After" || selected.Stats.Health != 12 || len(store.Characters()) != 2 {
		t.Fatalf("updated selected character = %#v roster=%#v", selected, store.Characters())
	}

	if err := store.UpdateSelected(Character{ID: "other"}); err == nil {
		t.Fatal("non-selected identity replaced the selected character")
	}
}

// TestCharacterAppearanceIsDefensivelyCopied protects nested component maps across every Store read boundary.
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

// TestCharacterStatsAreDefensivelyCopied protects pointer-backed sheet snapshots on both insertion and listing.
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

// TestCreateRequiresOpaqueIdentityAndRejectsDuplicateID keeps Store validation intentionally limited to identity.
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

// TestCreateSelectedReplacesExistingSelectionAtomically protects the creation flow from retaining an old selection.
func TestCreateSelectedReplacesExistingSelectionAtomically(t *testing.T) {
	store := New(Character{ID: "existing", Name: "Existing", Class: "Assassin"})
	if err := store.Select("existing"); err != nil {
		t.Fatal(err)
	}

	created := Character{ID: "new", Name: "New", Class: "Barbarian"}
	if err := store.CreateSelected(created); err != nil {
		t.Fatal(err)
	}

	selected, ok := store.Selected()
	if !ok || selected.ID != created.ID || selected.Class != created.Class {
		t.Fatalf("selected after create = %#v, %v", selected, ok)
	}
}

// TestDeleteClearsSelectedCharacter prevents a deleted identity from remaining active while retaining other entries.
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
