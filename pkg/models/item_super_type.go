package models

// ItemSuperType represents the item types.
type ItemSuperType int

const (
	TypeArmor ItemSuperType = iota
	TypeWeapon
	TypeMisc
)
