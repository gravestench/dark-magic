package models

// GambleRecord represents the data for each item type that appears as a possible option for the Gambling UI.
type GambleRecord struct {
	Name string `csv:"name"` // Reference field to describe the Item
	Code string `csv:"code"` // Pointer to the "code" field from weapons.txt/armor.txt/misc.txt
}
