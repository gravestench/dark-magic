package save

import "testing"

// TestNewCharacterOwnsCanonicalCreationDefaults protects normalization and server-owned starting progression.
func TestNewCharacterOwnsCanonicalCreationDefaults(t *testing.T) {
	character, err := NewCharacter(CharacterRequest{
		ID:        "id",
		Name:      "  Vale'era  ",
		Class:     "ASSASSIN",
		Expansion: true,
		Hardcore:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if character.Name != "Vale'era" || character.Class != "Assassin" || character.Level != 1 ||
		!character.Expansion || !character.Hardcore {
		t.Fatalf("character = %#v", character)
	}

	if character.Stats == nil || character.Stats.Strength != 20 || character.Stats.Dexterity != 20 ||
		character.Stats.Vitality != 20 || character.Stats.Energy != 25 || character.Stats.MaxHealth != 50 ||
		character.Stats.MaxMana != 25 || character.Stats.MaxStamina != 95 {
		t.Fatalf("stats = %#v", character.Stats)
	}
}

// TestNewCharacterRejectsInvalidClientChoices protects opaque identity, name grammar, and the class allowlist.
func TestNewCharacterRejectsInvalidClientChoices(t *testing.T) {
	for _, request := range []CharacterRequest{
		{Name: "Hero", Class: "Amazon"},
		{ID: "id", Name: "x", Class: "Amazon"},
		{ID: "id", Name: "Bad--Name", Class: "Amazon"},
		{ID: "id", Name: "Hero", Class: "Wizard"},
	} {
		if _, err := NewCharacter(request); err == nil {
			t.Fatalf("NewCharacter(%#v) succeeded", request)
		}
	}
}
