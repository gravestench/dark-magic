package character_gen

import "fmt"

func main() {
	characterNames := []string{"Amazon", "Barbarian", "Necromancer", "Sorceress", "Paladin"}
	characterGenerator := NewDiablo2CharacterGenerator(characterNames)

	for i := 0; i < 5; i++ {
		character, _ := characterGenerator.Generate()
		fmt.Printf("Character %d: Name: %s, Strength: %d, Dexterity: %d, Vitality: %d, Energy: %d\n", i+1, character.Name, character.Attributes.Strength, character.Attributes.Dexterity, character.Attributes.Vitality, character.Attributes.Energy)
	}
}
