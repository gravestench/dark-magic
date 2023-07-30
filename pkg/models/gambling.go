package models

// GamblingOptions represents the Item Types that will appear as possible items to purchase in the Gambling UI.
type GamblingOptions struct {
	// Overview: This file controls what Item Types will appear as possible items to purchase in the Gambling UI.

	// Data Fields:
	Name string // This is a reference field to describe the Item.
	Code string // This is a pointer to the “code” field from weapons.txt/armor.txt/misc.txt.
}
