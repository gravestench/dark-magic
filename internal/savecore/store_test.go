package savecore

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

func TestCreateValidatesCharacterIdentity(t *testing.T) {
	store := New()
	for _, character := range []Character{
		{ID: "short", Name: "A", Class: "Amazon"},
		{ID: "punctuation", Name: "-Hero", Class: "Amazon"},
		{ID: "unknown", Name: "Hero", Class: "Monk"},
	} {
		if err := store.Create(character); err == nil {
			t.Fatalf("expected validation error for %#v", character)
		}
	}
	if err := store.Create(Character{ID: "hero", Name: "D'Artagnan", Class: "amazon"}); err != nil {
		t.Fatal(err)
	}
	if got := store.Characters()[0].Class; got != "Amazon" {
		t.Fatalf("canonical class = %q", got)
	}
	if err := store.Create(Character{ID: "other", Name: "d'artagnan", Class: "Druid"}); err == nil {
		t.Fatal("expected case-insensitive duplicate-name error")
	}
}

func TestCreateNamedOwnsStorageIdentity(t *testing.T) {
	store := New()
	character, err := store.CreateNamed("D'Artagnan", "druid")
	if err != nil {
		t.Fatal(err)
	}
	if character.ID != "druid-d-artagnan" || character.Class != "Druid" || character.Level != 1 {
		t.Fatalf("created character = %#v", character)
	}
}

func TestCreateNamedStoresCreationOptions(t *testing.T) {
	store := New()
	character, err := store.CreateNamedWithOptions("Iron-Wolf", "Paladin", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if character.Expansion || !character.Hardcore {
		t.Fatalf("creation options = expansion %v, hardcore %v", character.Expansion, character.Hardcore)
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
