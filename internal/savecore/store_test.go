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
