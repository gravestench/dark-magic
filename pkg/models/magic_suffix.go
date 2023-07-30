package models

// MagicSuffix represents an item affix that is applied as a suffix for an item.
type MagicSuffix struct {
	Name           string  `csv:"Name"`           // Defines the item affix name
	Version        int     `csv:"version"`        // Defines which game version to use this item affix (<100 = Classic mode | 100 = Expansion mode)
	Spawnable      bool    `csv:"spawnable"`      // If equals 1, then this item affix is used as part of the game’s randomizer for assigning item modifiers when an item spawns. If equals 0, then this item affix is never used.
	Rare           bool    `csv:"rare"`           // If equals 1, then this item affix can be used when randomly assigning item modifiers when a rare item spawns. If equals 0, then this item affix is not used for rare items.
	Level          int     `csv:"level"`          // The minimum item level required for this item affix to spawn on the item. If the item level is below this value, then the item affix will not spawn on the item.
	MaxLevel       int     `csv:"maxlevel"`       // The maximum item level required for this item affix to spawn on the item. If the item level is above this value, then the item affix will not spawn on the item.
	LevelReq       int     `csv:"levelreq"`       // The minimum character level required to equip an item that has this item affix.
	ClassSpecific  string  `csv:"classspecific"`  // Controls if this item affix should only be used for class-specific items. This relies on the class specified in the “Class” field from ItemTypes.txt, for the specific item.
	Class          string  `csv:"class"`          // Controls which character class is required for the class-specific level requirement “classlevelreq” field.
	ClassLevelReq  *int    `csv:"classlevelreq"`  // The minimum character level required for a specific class to equip an item that has this item affix. If equals null, then the class will default to using the “levelreq” field.
	Frequency      float64 `csv:"frequency"`      // Controls the probability that the affix appears on the item (a higher value means that the item affix will appear on the item more often).
	Group          *int    `csv:"group"`          // Assigns an item affix to a specific group number. Items cannot spawn with more than 1 item affix with the same group number. This is used to guarantee that certain item affixes do not overlap on the same item. If this field is null, then the group number will default to group 0.
	Mod1Code       string  `csv:"mod1code"`       // Controls the item properties for the item affix. (Uses the “code” field from Properties.txt)
	Mod1Param      *string `csv:"mod1param"`      // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod1Min        *string `csv:"mod1min"`        // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod1Max        *string `csv:"mod1max"`        // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	TransformColor *string `csv:"transformcolor"` // Controls the color change of the item after spawning with this item affix. If empty, then the item affix will not change the item’s color. (Uses Color Codes from the reference file colors.txt)
	IType1         string  `csv:"itype1"`         // Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	IType2         string  `csv:"itype2"`         // (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	IType3         string  `csv:"itype3"`         // (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	IType4         string  `csv:"itype4"`         // (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	IType5         string  `csv:"itype5"`         // (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	IType6         string  `csv:"itype6"`         // (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	IType7         string  `csv:"itype7"`         // (Continued...) Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	EType1         string  `csv:"etype1"`         // Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	EType2         string  `csv:"etype2"`         // (Continued...) Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	EType3         string  `csv:"etype3"`         // (Continued...) Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	EType4         string  `csv:"etype4"`         // (Continued...) Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	EType5         string  `csv:"etype5"`         // (Continued...) Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	Multiply       int     `csv:"multiply"`       // Multiplicative modifier for the item’s buy and sell costs, based on the item affix (Calculated in 1024ths for buy cost and 4096ths for sell cost)
	Add            int     `csv:"add"`            // Flat integer modification to the item’s buy and sell costs, based on the item affix
}
