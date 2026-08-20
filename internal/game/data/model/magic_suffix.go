package models

// MagicSuffix represents an item affix that is applied as a suffix for an item.
type MagicSuffix struct {
	Name string `csv:"Name"` // Defines the item affix name
	// Defines which game version to use this item affix (<100 = Classic mode | 100 = Expansion mode)
	Version int `csv:"version"`
	// If equals 1, then this item affix is used as part of the game’s randomizer for assigning item modifiers when an
	// item spawns. If equals 0, then this item affix is never used.
	Spawnable bool `csv:"spawnable"`
	// If equals 1, then this item affix can be used when randomly assigning item modifiers when a rare item spawns. If
	// equals 0, then this item affix is not used for rare items.
	Rare bool `csv:"rare"`
	// The minimum item level required for this item affix to spawn on the item. If the item level is below this value,
	// then the item affix will not spawn on the item.
	Level int `csv:"level"`
	// The maximum item level required for this item affix to spawn on the item. If the item level is above this value,
	// then the item affix will not spawn on the item.
	MaxLevel int `csv:"maxlevel"`
	// The minimum character level required to equip an item that has this item affix.
	LevelReq int `csv:"levelreq"`
	// Controls if this item affix should only be used for class-specific items. This relies on the class specified in the
	// “Class” field from ItemTypes.txt, for the specific item.
	ClassSpecific string `csv:"classspecific"`
	// Controls which character class is required for the class-specific level requirement “classlevelreq” field.
	Class string `csv:"class"`
	// The minimum character level required for a specific class to equip an item that has this item affix. If equals null,
	// then the class will default to using the “levelreq” field.
	ClassLevelReq *int `csv:"classlevelreq"`
	// Controls the probability that the affix appears on the item (a higher value means that the item affix will appear on
	// the item more often).
	Frequency float64 `csv:"frequency"`
	// Assigns an item affix to a specific group number. Items cannot spawn with more than 1 item affix with the same group
	// number. This is used to guarantee that certain item affixes do not overlap on the same item. If this field is null,
	// then the group number will default to group 0.
	Group *int `csv:"group"`
	// Controls the item properties for the item affix. (Uses the “code” field from Properties.txt)
	Mod1Code string `csv:"mod1code"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	Mod1Param *string `csv:"mod1param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	Mod1Min *string `csv:"mod1min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	Mod1Max *string `csv:"mod1max"`
	// Controls the color change of the item after spawning with this item affix. If empty, then the item affix will not
	// change the item’s color. (Uses Color Codes from the reference file colors.txt)
	TransformColor *string `csv:"transformcolor"`
	// Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	IType1 string `csv:"itype1"`
	// (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	IType2 string `csv:"itype2"`
	// (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	IType3 string `csv:"itype3"`
	// (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	IType4 string `csv:"itype4"`
	// (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	IType5 string `csv:"itype5"`
	// (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	IType6 string `csv:"itype6"`
	// (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	IType7 string `csv:"itype7"`
	// Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	EType1 string `csv:"etype1"`
	// (Continued...) Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	EType2 string `csv:"etype2"`
	// (Continued...) Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	EType3 string `csv:"etype3"`
	// (Continued...) Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	EType4 string `csv:"etype4"`
	// (Continued...) Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from
	// ItemTypes.txt
	EType5 string `csv:"etype5"`
	// Multiplicative modifier for the item’s buy and sell costs, based on the item affix (Calculated in 1024ths for buy
	// cost and 4096ths for sell cost)
	Multiply int `csv:"multiply"`
	// Flat integer modification to the item’s buy and sell costs, based on the item affix
	Add int `csv:"add"`
}
