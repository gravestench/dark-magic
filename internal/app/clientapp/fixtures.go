package clientapp

import (
	"fmt"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// DevelopmentCharacters makes predictable pretend heroes for UI development.
// Nothing is written to disk; closing the program makes them disappear.
func DevelopmentCharacters(count int) []d2save.Character {
	if count <= 0 {
		return nil
	}
	classes := []string{"Amazon", "Sorceress", "Necromancer", "Paladin", "Barbarian", "Assassin", "Druid"}
	characters := make([]d2save.Character, 0, count)
	for index := 0; index < count; index++ {
		characters = append(characters, developmentCharacter(index, classes[index%len(classes)]))
	}
	return characters
}

// developmentCharacter derives stable identities so captures remain comparable across runs.
func developmentCharacter(index int, class string) d2save.Character {
	return d2save.Character{
		ID:        fmt.Sprintf("fixture-%02d", index+1),
		Name:      fmt.Sprintf("Hero%02d", index+1),
		Class:     class,
		Level:     index + 1,
		Expansion: true,
		Hardcore:  index%3 == 2,
		Stats:     developmentStats(),
	}
}

// developmentStats supplies plausible nonzero resources without depending on a persisted save or game-data lookup.
func developmentStats() *d2save.Stats {
	return &d2save.Stats{
		Experience: 1200, NextLevelExperience: 2250,
		Strength: 25, Dexterity: 20, Vitality: 25, Energy: 15,
		Defense: 42, Health: 70, MaxHealth: 70, Mana: 30, MaxMana: 30,
		Stamina: 84, MaxStamina: 84,
	}
}
