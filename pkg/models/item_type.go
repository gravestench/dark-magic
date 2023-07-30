package models

// ItemType represents the general statistics for each type of item, used for the item type fields in other files.
type ItemType struct {
	// This is a reference field to define the Item Type name.
	ItemType string `csv:"ItemType"`

	// Defines the unique pointer for this Item Type, which is used by the following files: weapons.txt, armor.txt, misc.txt, cubemain.txt, skills.txt, treasureclassex.txt.
	Code string `csv:"Code"`

	// Points to the index of another Item Type to reference as a parent. This is used to create a hierarchy for Item Types where the parents will have more universal settings shared across the related children.
	Equiv1 string `csv:"Equiv1"`

	// Points to the index of another Item Type to reference as a parent. This is used to create a hierarchy for Item Types where the parents will have more universal settings shared across the related children.
	Equiv2 string `csv:"Equiv2"`

	// Boolean Field. If equals 1, then the item can be repaired by an NPC in the shop UI. If equals 0, then the item cannot be repaired.
	Repair bool `csv:"Repair"`

	// Boolean Field. If equals 1, then the item can be equipped by a character (also will require the "BodyLoc1" & "BodyLoc2" fields as parameters). If equals 0, then the item can only be carried in the inventory, stash, or Horadric Cube.
	Body bool `csv:"Body"`

	// These are required parameters if the "Body" field is enabled. These fields specify the inventory slots where the item can be equipped.
	BodyLoc1 string `csv:"BodyLoc1"`
	BodyLoc2 string `csv:"BodyLoc2"`

	// Points to the index of another Item Type as the required equipped Item Type to be used as ammo.
	Shoots string `csv:"Shoots"`

	// Points to the index of another Item Type as the required equipped Item Type to be used as this ammo's weapon.
	Quiver string `csv:"Quiver"`

	// Boolean Field. If equals 1, then it determines that this item is a throwing weapon. If equals 0, then ignore this.
	Throwable bool `csv:"Throwable"`

	// Boolean Field. If equals 1, then the item (considered ammo in this case) will be automatically transferred from the inventory to the required "BodyLoc" when another item runs out of that specific ammo. If equals 0, then ignore this.
	Reload bool `csv:"Reload"`

	// Boolean Field. If equals 1, then the item in the inventory will replace a matching equipped item if that equipped item was destroyed. If equals 0, then ignore this.
	ReEquip bool `csv:"ReEquip"`

	// Boolean Field. If equals 1, then if the player picks up a matching Item Type, then they will try to automatically stack together. If equals 0, then ignore this.
	AutoStack bool `csv:"AutoStack"`

	// Boolean Field. If equals 1, then this item will always have the Magic quality (unless it is a Quest item). If equals 0, then ignore this.
	Magic bool `csv:"Magic"`

	// Boolean Field. If equals 1, then this item can spawn as a Rare quality. If equals 0, then ignore this.
	Rare bool `csv:"Rare"`

	// Boolean Field. If equals 1, then this item will always have the Normal quality. If equals 0, then ignore this.
	Normal bool `csv:"Normal"`

	// Boolean Field. If equals 1, then this item can be placed in the character's belt slots. If equals 0, then ignore this.
	Beltable bool `csv:"Beltable"`

	// Determines the maximum possible number of sockets that can be spawned on the item when the item level is greater than or equal to 1 and less than or equal to the "MaxSocketsLevelThreshold1" value. The number of sockets is capped by the "gemsockets" value from the weapons.txt/armor.txt/misc.txt file.
	MaxSockets1 int `csv:"MaxSockets1"`

	// Defines the item level threshold between using the "MaxSockets1" and "MaxSockets2" field.
	MaxSocketsLevelThreshold1 int `csv:"MaxSocketsLevelThreshold1"`

	// Determines the maximum possible number of sockets that can be spawned on the item when the item level is greater than the "MaxSocketsLevelThreshold1" value and less than or equal to the "MaxSocketsLevelThreshold2". The number of sockets is capped by the "gemsockets" value from the weapons.txt/armor.txt/misc.txt file.
	MaxSockets2 int `csv:"MaxSockets2"`

	// Defines the item level threshold between using the "MaxSockets2" and "MaxSockets3" field.
	MaxSocketsLevelThreshold2 int `csv:"MaxSocketsLevelThreshold2"`

	// Determines the maximum possible number of sockets that can be spawned on the item when the item level is greater than the "MaxSocketsLevelThreshold2" value. The number of sockets is capped by the "gemsockets" value from the weapons.txt/armor.txt/misc.txt file.
	MaxSockets3 int `csv:"MaxSockets3"`

	// Boolean Field. If equals 1, then allow this Item Type to be used in default treasure classes. If equals 0, then ignore this.
	TreasureClass bool `csv:"TreasureClass"`

	// Determines the chance for the item to spawn with stats, when created as a random Weapon/Armor/Misc item. Used in the following formula: IF RANDOM(0, ("Rarity" - [Current Act Level])) > 0, THEN spawn stats.
	Rarity int `csv:"Rarity"`

	// Determines if the Item Type should have class specific item skill modifiers.
	StaffMods string `csv:"StaffMods"`

	// Determines if this item should be usable only by a specific class.
	Class string `csv:"Class"`

	// Tracks the number of inventory graphics used for this item type. This number must match the number of "InvGfx" fields used.
	VarInvGfx int `csv:"VarInvGfx"`

	// Defines a DC6 file to use for the item's inventory graphics. The number of "InvGfx" fields used should match the value used in "VarInvGfx".
	InvGfx1 string `csv:"InvGfx1"`
	InvGfx2 string `csv:"InvGfx2"`
	InvGfx3 string `csv:"InvGfx3"`
	InvGfx4 string `csv:"InvGfx4"`
	InvGfx5 string `csv:"InvGfx5"`
	InvGfx6 string `csv:"InvGfx6"`
}
