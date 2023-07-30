package models

// MagicPrefix represents the data structure for controlling what item affixes (groups of item modifiers) are applied as the prefix for an item.
// These item affixes will appear at the start of an item’s name.
// This file is loaded together with other similar files in the following order: magicsuffix.txt, magicprefix.txt, automagic.txt.
// These combined files form the Item Mods structure.
type MagicPrefix struct {
	Name           string `csv:"Name"`           // Defines the item affix name
	Version        int    `csv:"version"`        // Defines which game version to use this item affix (<100 = Classic mode | 100 = Expansion mode)
	Spawnable      bool   `csv:"spawnable"`      // Boolean Field. If equals 1, then this item affix is used as part of the game’s randomizer for assigning item modifiers when an item spawns. If equals 0, then this item affix is never used.
	Rare           bool   `csv:"rare"`           // Boolean Field. If equals 1, then this item affix can be used when randomly assigning item modifiers when a rare item spawns. If equals 0, then this item affix is not used for rare items.
	Level          int    `csv:"level"`          // The minimum item level required for this item affix to spawn on the item. If the item level is below this value, then the item affix will not spawn on the item.
	MaxLevel       int    `csv:"maxlevel"`       // The maximum item level required for this item affix to spawn on the item. If the item level is above this value, then the item affix will not spawn on the item.
	LevelReq       int    `csv:"levelreq"`       // The minimum character level required to equip an item that has this item affix
	ClassSpecific  string `csv:"classspecific"`  // Controls if this item affix should only be used for class specific items. This relies on the class specified in the “Class” field from ItemTypes.txt, for the specific item.
	Class          string `csv:"class"`          // Controls which character class is required for the class specific level requirement “classlevelreq” field
	ClassLevelReq  int    `csv:"classlevelreq"`  // The minimum character level required for a specific class in order to equip an item that has this item affix. This relies on the class specified in the “class” field. If equals null, then the class will default to using the “levelreq” field.
	Frequency      int    `csv:"frequency"`      // Controls the probability that the affix appears on the item (a higher value means that the item affix will appear on the item more often). This value gets summed together with other “frequency” values from all possible item affixes that can spawn on the item, and then is used as a denominator value for the randomizer. Whichever item affix is randomly selected will be the one to appear on the item. The formula is calculated as the following: [Item Affix Selected] = [“frequency”] / [Total Frequency]. If the item has a magic level (from the “magic lvl” field in weapons.txt/armor.txt/misc.txt) then the magic level value is multiplied with this value. If equals 0, then this item affix will never appear on an item.
	Group          int    `csv:"group"`          // Assigns an item affix to a specific group number. Items cannot spawn with more than 1 item affix with the same group number. This is used to guarantee that certain item affixes do not overlap on the same item. If this field is null, then the group number will default to group 0.
	Mod1Code       string `csv:"mod1code"`       // Controls the item properties for the item affix (Uses the “code” field from Properties.txt)
	Mod1Param      int    `csv:"mod1param"`      // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod1Min        int    `csv:"mod1min"`        // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod1Max        int    `csv:"mod1max"`        // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod2Code       string `csv:"mod2code"`       // Controls the item properties for the item affix (Uses the “code” field from Properties.txt)
	Mod2Param      int    `csv:"mod2param"`      // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod2Min        int    `csv:"mod2min"`        // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod2Max        int    `csv:"mod2max"`        // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod3Code       string `csv:"mod3code"`       // Controls the item properties for the item affix (Uses the “code” field from Properties.txt)
	Mod3Param      int    `csv:"mod3param"`      // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod3Min        int    `csv:"mod3min"`        // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	Mod3Max        int    `csv:"mod3max"`        // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	TransformColor string `csv:"transformcolor"` // Controls the color change of the item after spawning with this item affix. If empty, then the item affix will not change the item’s color. (Uses Color Codes from the reference file colors.txt)
	ItemType1      string `csv:"itype1"`         // Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ItemType2      string `csv:"itype2"`         // Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ItemType3      string `csv:"itype3"`         // Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ItemType4      string `csv:"itype4"`         // Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ItemType5      string `csv:"itype5"`         // Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ItemType6      string `csv:"itype6"`         // Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ItemType7      string `csv:"itype7"`         // Controls what Item Types are allowed to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ExcludeType1   string `csv:"etype1"`         // Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ExcludeType2   string `csv:"etype2"`         // Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ExcludeType3   string `csv:"etype3"`         // Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ExcludeType4   string `csv:"etype4"`         // Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ExcludeType5   string `csv:"etype5"`         // Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ExcludeType6   string `csv:"etype6"`         // Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	ExcludeType7   string `csv:"etype7"`         // Controls what Item Types are excluded to spawn with this item affix. Uses the “code” field from ItemTypes.txt
	Multiply       int    `csv:"multiply"`       // Multiplicative modifier for the item’s buy and sell costs, based on the item affix (Calculated in 1024ths for buy cost and 4096ths for sell cost)
	Add            int    `csv:"add"`            // Flat integer modification to the item’s buy and sell costs, based on the item affix
}
