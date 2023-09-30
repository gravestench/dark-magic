package character_gen

import (
	"errors"
	"math/rand"
	"time"
)

// Attributes represent character attributes.
type Attributes struct {
	Strength  int
	Dexterity int
	Vitality  int
	Energy    int
}

// Character represents a character with attributes.
type Character struct {
	Name       string
	Attributes *Attributes
}

// CharacterGenerator defines the interface for generating characters.
type CharacterGenerator interface {
	Generate() (*Character, error) // Return an error if something goes wrong
}

// Diablo2CharacterGenerator generates characters with Diablo 2 attributes.
type Diablo2CharacterGenerator struct {
	Names   []string
	randSrc *rand.Rand
}

func NewDiablo2CharacterGenerator(names []string) *Diablo2CharacterGenerator {
	randSrc := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &Diablo2CharacterGenerator{
		Names:   names,
		randSrc: randSrc,
	}
}

func (cg *Diablo2CharacterGenerator) generateName() string {
	index := cg.randSrc.Intn(len(cg.Names))
	return cg.Names[index]
}

func (cg *Diablo2CharacterGenerator) generateAttribute() *Attributes {
	return &Attributes{
		Strength:  cg.randSrc.Intn(10) + 1,
		Dexterity: cg.randSrc.Intn(10) + 1,
		Vitality:  cg.randSrc.Intn(10) + 1,
		Energy:    cg.randSrc.Intn(10) + 1,
	}
}

func (cg *Diablo2CharacterGenerator) Generate() (*Character, error) {
	name := cg.generateName()
	attributes := cg.generateAttribute()

	if name == "" {
		// Return an error if name generation fails
		return nil, errors.New("failed to generate a name")
	}

	return &Character{
		Name:       name,
		Attributes: attributes,
	}, nil
}
