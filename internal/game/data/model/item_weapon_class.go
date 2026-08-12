package models

// WeaponClassID represents the base weapon class
type WeaponClassID string

// WeaponClass represents different weapon classes.
type WeaponClass struct {
	Name string        `csv:"Weapon Class"` // The name of the weapon class.
	Code WeaponClassID `csv:"Code"`         // The unique 3-letter/number code for the weapon class.
}
