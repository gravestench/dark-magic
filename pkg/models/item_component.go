package models

// Component represents the layers of player animation when the item is equipped.
type Component int

const (
	ComponentHead Component = iota
	ComponentTorso
	ComponentLegs
	ComponentRightArm
	ComponentLeftArm
	ComponentRightHand
	ComponentLeftHand
	ComponentShield
	ComponentSpecial1
	ComponentSpecial2
	ComponentSpecial3
	ComponentSpecial4
	ComponentSpecial5
	ComponentSpecial6
	ComponentSpecial7
	ComponentSpecial8
	ComponentNone
)
