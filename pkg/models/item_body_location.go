package models

// ItemBodyLocation represents the possible body locations for equippable items
type ItemBodyLocation string

const (
	ItemBodyLocationNone   ItemBodyLocation = ""
	ItemBodyLocationHead   ItemBodyLocation = "head"
	ItemBodyLocationNeck   ItemBodyLocation = "neck"
	ItemBodyLocationTorso  ItemBodyLocation = "tors"
	ItemBodyLocationRArm   ItemBodyLocation = "rarm"
	ItemBodyLocationLArm   ItemBodyLocation = "larm"
	ItemBodyLocationRRing  ItemBodyLocation = "rrin"
	ItemBodyLocationLRing  ItemBodyLocation = "lrin"
	ItemBodyLocationBelt   ItemBodyLocation = "belt"
	ItemBodyLocationFeet   ItemBodyLocation = "feet"
	ItemBodyLocationGloves ItemBodyLocation = "glov"
)
