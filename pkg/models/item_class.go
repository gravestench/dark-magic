package models

// ItemClass represents the classes that can use specific items
type ItemClass string

const (
	ItemClassNone        ItemClass = ""
	ItemClassAmazon      ItemClass = "ama"
	ItemClassBarbarian   ItemClass = "bar"
	ItemClassPaladin     ItemClass = "pal"
	ItemClassNecromancer ItemClass = "nec"
	ItemClassSorceress   ItemClass = "sor"
	ItemClassDruid       ItemClass = "dru"
	ItemClassAssassin    ItemClass = "ass"
)
