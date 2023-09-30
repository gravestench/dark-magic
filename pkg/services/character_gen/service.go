// service.go

package character_gen

import (
	"fmt"
)

// CharacterService defines the service for generating characters.
type CharacterService struct {
	CharacterGenerator CharacterGenerator
}

// NewCharacterService creates a new CharacterService instance.
func NewCharacterService(generator CharacterGenerator) *CharacterService {
	return &CharacterService{
		CharacterGenerator: generator,
	}
}

// GenerateCharacter generates a character using the CharacterGenerator.
func (cs *CharacterService) GenerateCharacter() (*Character, error) {
	return cs.CharacterGenerator.Generate()
}

// PrintCharacterAttributes prints the attributes of a character.
func PrintCharacterAttributes(character *Character) {
	fmt.Printf("Name: %s, Strength: %d, Dexterity: %d, Vitality: %d, Energy: %d\n", character.Name, character.Attributes.Strength, character.Attributes.Dexterity, character.Attributes.Vitality, character.Attributes.Energy)
}
