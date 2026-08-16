package save

import (
	"errors"
	"strings"
	"unicode"
)

var ErrCharacterPolicy = errors.New("d2legacy: invalid character")

// CharacterRequest is the complete player choice accepted at character
// creation. Every other field is supplied by d2legacy policy, never by an
// untrusted client.
type CharacterRequest struct {
	ID        string
	Name      string
	Class     string
	Expansion bool
	Hardcore  bool
}

type classDefaults struct {
	name                                  string
	strength, dexterity, vitality, energy int
	health, mana, stamina                 int
}

var characterClasses = map[string]classDefaults{
	"amazon":      {"Amazon", 20, 25, 20, 15, 50, 15, 84},
	"sorceress":   {"Sorceress", 10, 25, 10, 35, 40, 35, 74},
	"necromancer": {"Necromancer", 15, 25, 15, 25, 45, 25, 79},
	"paladin":     {"Paladin", 25, 20, 25, 15, 55, 15, 89},
	"barbarian":   {"Barbarian", 30, 20, 25, 10, 55, 10, 92},
	"assassin":    {"Assassin", 20, 20, 20, 25, 50, 25, 95},
	"druid":       {"Druid", 15, 20, 25, 20, 55, 20, 84},
}

// NewCharacter is the single production constructor used by local-profile and
// realm character creation. It prevents those two flows from drifting and
// deliberately ignores client-supplied progression, appearance, and stats.
func NewCharacter(request CharacterRequest) (Character, error) {
	id := strings.TrimSpace(request.ID)
	name := strings.TrimSpace(request.Name)
	defaults, found := characterClasses[strings.ToLower(strings.TrimSpace(request.Class))]
	if id == "" || !found || !validCharacterName(name) {
		return Character{}, ErrCharacterPolicy
	}
	return Character{
		ID: id, Name: name, Class: defaults.name, Level: 1,
		Expansion: request.Expansion, Hardcore: request.Hardcore,
		Stats: &Stats{
			Strength: defaults.strength, Dexterity: defaults.dexterity,
			Vitality: defaults.vitality, Energy: defaults.energy,
			Health: defaults.health, MaxHealth: defaults.health,
			Mana: defaults.mana, MaxMana: defaults.mana,
			Stamina: defaults.stamina, MaxStamina: defaults.stamina,
		},
	}, nil
}

func validCharacterName(name string) bool {
	if len(name) < 2 || len(name) > 15 {
		return false
	}
	for index, value := range name {
		if unicode.IsLetter(value) && value <= unicode.MaxASCII {
			continue
		}
		if (value == '-' || value == '\'') && index > 0 && index < len(name)-1 {
			previous := name[index-1]
			if previous != '-' && previous != '\'' {
				continue
			}
		}
		return false
	}
	return true
}
