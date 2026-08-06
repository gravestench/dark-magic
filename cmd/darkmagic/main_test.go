package main

import "testing"

func TestDevelopmentCharacters(t *testing.T) {
	t.Parallel()

	if characters := developmentCharacters(0); characters != nil {
		t.Fatalf("developmentCharacters(0) = %#v, want nil", characters)
	}
	characters := developmentCharacters(10)
	if len(characters) != 10 {
		t.Fatalf("developmentCharacters(10) length = %d", len(characters))
	}
	wantClasses := []string{"Amazon", "Sorceress", "Necromancer", "Paladin", "Barbarian", "Assassin", "Druid", "Amazon"}
	for index, wantClass := range wantClasses {
		character := characters[index]
		if character.Class != wantClass || character.Level != index+1 || !character.Expansion {
			t.Errorf("character %d = %#v", index, character)
		}
	}
	if characters[0].ID != "fixture-01" || characters[0].Name != "Hero01" {
		t.Fatalf("first character = %#v", characters[0])
	}
	if !characters[2].Hardcore || characters[1].Hardcore || characters[3].Hardcore {
		t.Fatalf("unexpected hardcore sequence: %#v", characters[:4])
	}
}
