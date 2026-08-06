package models

// InventoryData represents the data from the "inventory.txt" file.
type InventoryData struct {
	// Class is a reference field to define the type of inventory screen.
	Class string `csv:"class"`

	// InvLeft is the starting X coordinate pixel position of the inventory panel.
	InvLeft int `csv:"invLeft"`

	// InvRight is the ending X coordinate pixel position of the inventory panel (Includes the "InvLeft" value with the inventory width size).
	InvRight int `csv:"invRight"`

	// InvTop is the starting Y coordinate pixel position of the inventory panel.
	InvTop int `csv:"invTop"`

	// InvBottom is the ending Y coordinate pixel position of the inventory panel (Includes the "InvTop" value with the inventory height size).
	InvBottom int `csv:"invBottom"`

	// GridX is the column number size of the inventory grid, measured in the number of grid boxes to use.
	GridX int `csv:"gridX"`

	// GridY is the column row size of the inventory grid, measured in the number of grid boxes to use.
	GridY int `csv:"gridY"`

	// GridLeft is the starting X coordinate location of the inventory's left grid side.
	GridLeft int `csv:"gridLeft"`

	// GridRight is the ending X coordinate location of the inventory's right grid side (Includes the "GridLeft" value with the grid width size).
	GridRight int `csv:"gridRight"`

	// GridTop is the starting Y coordinate location of the inventory's top grid side.
	GridTop int `csv:"gridTop"`

	// GridBottom is the ending Y coordinate location of the inventory's bottom grid side (Includes the "GridTop" value with the grid height size).
	GridBottom int `csv:"gridBottom"`

	// GridBoxWidth is the width size of an inventory's box cell.
	GridBoxWidth int `csv:"gridBoxWidth"`

	// GridBoxHeight is the height size of an inventory's box cell.
	GridBoxHeight int `csv:"gridBoxHeight"`

	// RArmLeft is the starting X coordinate location of the Right Weapon Slot.
	RArmLeft int `csv:"rArmLeft"`

	// RArmRight is the ending X coordinate location of the Right Weapon Slot (Includes the "RArmLeft" value with the "RArmWidth" value).
	RArmRight int `csv:"rArmRight"`

	// RArmTop is the starting Y coordinate location of the Right Weapon Slot.
	RArmTop int `csv:"rArmTop"`

	// RArmBottom is the ending Y coordinate location of the Right Weapon Slot (Includes the "RArmTop" value with the "RArmHeight" value).
	RArmBottom int `csv:"rArmBottom"`

	// RArmWidth is the pixel width of the Right Weapon Slot.
	RArmWidth int `csv:"rArmWidth"`

	// RArmHeight is the pixel height of the Right Weapon Slot.
	RArmHeight int `csv:"rArmHeight"`

	// TorsoLeft is the starting X coordinate location of the Body Armor Slot.
	TorsoLeft int `csv:"torsoLeft"`

	// TorsoRight is the ending X coordinate location of the Body Armor Slot (Includes the "TorsoLeft" value with the "TorsoWidth" value).
	TorsoRight int `csv:"torsoRight"`

	// TorsoTop is the starting Y coordinate location of the Body Armor Slot.
	TorsoTop int `csv:"torsoTop"`

	// TorsoBottom is the ending Y coordinate location of the Body Armor Slot (Includes the "TorsoTop" value with the "TorsoHeight" value).
	TorsoBottom int `csv:"torsoBottom"`

	// TorsoWidth is the pixel width of the Body Armor Slot.
	TorsoWidth int `csv:"torsoWidth"`

	// TorsoHeight is the pixel height of the Body Armor Slot.
	TorsoHeight int `csv:"torsoHeight"`

	// LArmLeft is the starting X coordinate location of the Left Weapon Slot.
	LArmLeft int `csv:"lArmLeft"`

	// LArmRight is the ending X coordinate location of the Left Weapon Slot (Includes the "LArmLeft" value with the "LArmWidth" value).
	LArmRight int `csv:"lArmRight"`

	// LArmTop is the starting Y coordinate location of the Left Weapon Slot.
	LArmTop int `csv:"lArmTop"`

	// LArmBottom is the ending Y coordinate location of the Left Weapon Slot (Includes the "LArmTop" value with the "LArmHeight" value).
	LArmBottom int `csv:"lArmBottom"`

	// LArmWidth is the pixel width of the Left Weapon Slot.
	LArmWidth int `csv:"lArmWidth"`

	// LArmHeight is the pixel height of the Left Weapon Slot.
	LArmHeight int `csv:"lArmHeight"`

	// HeadLeft is the starting X coordinate location of the Helm Slot.
	HeadLeft int `csv:"headLeft"`

	// HeadRight is the ending X coordinate location of the Helm Slot (Includes the "HeadLeft" value with the "HeadWidth" value).
	HeadRight int `csv:"headRight"`

	// HeadTop is the starting Y coordinate location of the Helm Slot.
	HeadTop int `csv:"headTop"`

	// HeadBottom is the ending Y coordinate location of the Helm Slot (Includes the "HeadTop" value with the "HeadHeight" value).
	HeadBottom int `csv:"headBottom"`

	// HeadWidth is the pixel width of the Helm Slot.
	HeadWidth int `csv:"headWidth"`

	// HeadHeight is the pixel height of the Helm Slot.
	HeadHeight int `csv:"headHeight"`

	// NeckLeft is the starting X coordinate location of the Amulet Slot.
	NeckLeft int `csv:"neckLeft"`

	// NeckRight is the ending X coordinate location of the Amulet Slot (Includes the "NeckLeft" value with the "NeckWidth" value).
	NeckRight int `csv:"neckRight"`

	// NeckTop is the starting Y coordinate location of the Amulet Slot.
	NeckTop int `csv:"neckTop"`

	// NeckBottom is the ending Y coordinate location of the Amulet Slot (Includes the "NeckTop" value with the "NeckHeight" value).
	NeckBottom int `csv:"neckBottom"`

	// NeckWidth is the pixel width of the Amulet Slot.
	NeckWidth int `csv:"neckWidth"`

	// NeckHeight is the pixel height of the Amulet Slot.
	NeckHeight int `csv:"neckHeight"`

	// RHandLeft is the starting X coordinate location of the Right Ring Slot.
	RHandLeft int `csv:"rHandLeft"`

	// RHandRight is the ending X coordinate location of the Right Ring Slot (Includes the "RHandLeft" value with the "RHandWidth" value).
	RHandRight int `csv:"rHandRight"`

	// RHandTop is the starting Y coordinate location of the Right Ring Slot.
	RHandTop int `csv:"rHandTop"`

	// RHandBottom is the ending Y coordinate location of the Right Ring Slot (Includes the "RHandTop" value with the "RHandHeight" value).
	RHandBottom int `csv:"rHandBottom"`

	// RHandWidth is the pixel width of the Right Ring Slot.
	RHandWidth int `csv:"rHandWidth"`

	// RHandHeight is the pixel height of the Right Ring Slot.
	RHandHeight int `csv:"rHandHeight"`

	// LHandLeft is the starting X coordinate location of the Left Ring Slot.
	LHandLeft int `csv:"lHandLeft"`

	// LHandRight is the ending X coordinate location of the Left Ring Slot (Includes the "LHandLeft" value with the "LHandWidth" value).
	LHandRight int `csv:"lHandRight"`

	// LHandTop is the starting Y coordinate location of the Left Ring Slot.
	LHandTop int `csv:"lHandTop"`

	// LHandBottom is the ending Y coordinate location of the Left Ring Slot (Includes the "LHandTop" value with the "LHandHeight" value).
	LHandBottom int `csv:"lHandBottom"`

	// LHandWidth is the pixel width of the Left Ring Slot.
	LHandWidth int `csv:"lHandWidth"`

	// LHandHeight is the pixel height of the Left Ring Slot.
	LHandHeight int `csv:"lHandHeight"`

	// BeltLeft is the starting X coordinate location of the Belt Slot.
	BeltLeft int `csv:"beltLeft"`

	// BeltRight is the ending X coordinate location of the Belt Slot (Includes the "BeltLeft" value with the "BeltWidth" value).
	BeltRight int `csv:"beltRight"`

	// BeltTop is the starting Y coordinate location of the Belt Slot.
	BeltTop int `csv:"beltTop"`

	// BeltBottom is the ending Y coordinate location of the Belt Slot (Includes the "BeltTop" value with the "BeltHeight" value).
	BeltBottom int `csv:"beltBottom"`

	// BeltWidth is the pixel width of the Belt Slot.
	BeltWidth int `csv:"beltWidth"`

	// BeltHeight is the pixel height of the Belt Slot.
	BeltHeight int `csv:"beltHeight"`

	// FeetLeft is the starting X coordinate location of the Boots Slot.
	FeetLeft int `csv:"feetLeft"`

	// FeetRight is the ending X coordinate location of the Boots Slot (Includes the "FeetLeft" value with the "FeetWidth" value).
	FeetRight int `csv:"feetRight"`

	// FeetTop is the starting Y coordinate location of the Boots Slot.
	FeetTop int `csv:"feetTop"`

	// FeetBottom is the ending Y coordinate location of the Boots Slot (Includes the "FeetTop" value with the "FeetHeight" value).
	FeetBottom int `csv:"feetBottom"`

	// FeetWidth is the pixel width of the Boots Slot.
	FeetWidth int `csv:"feetWidth"`

	// FeetHeight is the pixel height of the Boots Slot.
	FeetHeight int `csv:"feetHeight"`

	// GlovesLeft is the starting X coordinate location of the Gloves Slot.
	GlovesLeft int `csv:"glovesLeft"`

	// GlovesRight is the ending X coordinate location of the Gloves Slot (Includes the "GlovesLeft" value with the "GlovesWidth" value).
	GlovesRight int `csv:"glovesRight"`

	// GlovesTop is the starting Y coordinate location of the Gloves Slot.
	GlovesTop int `csv:"glovesTop"`

	// GlovesBottom is the ending Y coordinate location of the Gloves Slot (Includes the "GlovesTop" value with the "GlovesHeight" value).
	GlovesBottom int `csv:"glovesBottom"`

	// GlovesWidth is the pixel width of the Gloves Slot.
	GlovesWidth int `csv:"glovesWidth"`

	// GlovesHeight is the pixel height of the Gloves Slot.
	GlovesHeight int `csv:"glovesHeight"`
}
