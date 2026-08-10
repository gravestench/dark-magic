package clientapp

import (
	"fmt"

	"github.com/gravestench/dark-magic/internal/persistence"
)

// DevelopmentCharacters makes predictable pretend heroes for UI development.
// Nothing is written to disk; closing the program makes them disappear.
func DevelopmentCharacters(count int) []persistence.Character {
	if count <= 0 {
		return nil
	}
	classes := []string{"Amazon", "Sorceress", "Necromancer", "Paladin", "Barbarian", "Assassin", "Druid"}
	characters := make([]persistence.Character, 0, count)
	for index := 0; index < count; index++ {
		characters = append(characters, developmentCharacter(index, classes[index%len(classes)]))
	}
	return characters
}

func developmentCharacter(index int, class string) persistence.Character {
	return persistence.Character{
		ID:        fmt.Sprintf("fixture-%02d", index+1),
		Name:      fmt.Sprintf("Hero%02d", index+1),
		Class:     class,
		Level:     index + 1,
		Expansion: true,
		Hardcore:  index%3 == 2,
		Stats:     developmentStats(),
	}
}

func developmentStats() *persistence.Stats {
	return &persistence.Stats{
		Experience: 1200, NextLevelExperience: 2250,
		Strength: 25, Dexterity: 20, Vitality: 25, Energy: 15,
		Defense: 42, Health: 70, MaxHealth: 70, Mana: 30, MaxMana: 30,
		Stamina: 84, MaxStamina: 84,
	}
}
