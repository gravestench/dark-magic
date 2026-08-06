package models

// ArmorComponent represents the character's graphics and animations for different components when wearing armor.
type ArmorComponent int

const (
	ArmorComponentLight ArmorComponent = iota
	ArmorComponentMedium
	ArmorComponentHeavy
)
