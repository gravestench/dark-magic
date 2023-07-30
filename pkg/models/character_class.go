package models

// CharacterClass represents the character classes.
type CharacterClass int

const (
	CharacterClassNone        CharacterClass = iota // None
	CharacterClassAmazon                            // Amazon
	CharacterClassBarbarian                         // Barbarian
	CharacterClassPaladin                           // Paladin
	CharacterClassNecromancer                       // Necromancer
	CharacterClassSorceress                         // Sorceress
	CharacterClassDruid                             // Druid
	CharacterClassAssassin                          // Assassin
)
