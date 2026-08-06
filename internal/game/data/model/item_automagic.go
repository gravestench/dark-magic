package models

// AutoMagicData represents item affixes (groups of item modifiers) that are automatically applied to items, regardless of their quality type.
type AutoMagicData struct {
	Name           string `csv:"Name"`           // Defines the item affix name.
	Version        int    `csv:"version"`        // Defines which game version to use this item affix (<100 = Classic mode | 100 = Expansion mode).
	Spawnable      bool   `csv:"spawnable"`      // Boolean field. If equals 1, then this item affix is used as part of the game's randomizer for assigning item modifiers when an item spawns. If equals 0, then this item affix is never used.
	Rare           bool   `csv:"rare"`           // Boolean field. If equals 1, then this item affix can be used when randomly assigning item modifiers when a rare item spawns. If equals 0, then this item affix is not used for rare items.
	Level          int    `csv:"level"`          // The minimum item level required for this item affix to spawn on the item. If the item level is below this value, then the item affix will not spawn on the item.
	MaxLevel       int    `csv:"maxlevel"`       // The maximum item level required for this item affix to spawn on the item. If the item level is above this value, then the item affix will not spawn on the item.
	LevelReq       int    `csv:"levelreq"`       // The minimum character level required to equip an item that has this item affix.
	ClassSpecific  string `csv:"classspecific"`  // Controls if this item affix should only be used for class-specific items.
	Class          string `csv:"class"`          // Controls which character class is required for the class-specific level requirement "classlevelreq" field.
	ClassLevelReq  int    `csv:"classlevelreq"`  // The minimum character level required for a specific class in order to equip an item that has this item affix.
	Frequency      int    `csv:"frequency"`      // Controls the probability that the affix appears on the item.
	Group          int    `csv:"group"`          // Assigns an item affix to a specific group number.
	ModCode1       string `csv:"mod1code"`       // Controls the item properties for the item affix (Uses the "code" field from Properties.txt).
	ModCode2       string `csv:"mod2code"`       // Controls the item properties for the item affix (Uses the "code" field from Properties.txt).
	ModCode3       string `csv:"mod3code"`       // Controls the item properties for the item affix (Uses the "code" field from Properties.txt).
	ModParam1      int    `csv:"mod1param"`      // The "parameter" value associated with the related property (mod#code). Usage depends on the property function (See the "func" field on Properties.txt).
	ModParam2      int    `csv:"mod2param"`      // The "parameter" value associated with the related property (mod#code). Usage depends on the property function (See the "func" field on Properties.txt).
	ModParam3      int    `csv:"mod3param"`      // The "parameter" value associated with the related property (mod#code). Usage depends on the property function (See the "func" field on Properties.txt).
	ModMin1        int    `csv:"mod1min"`        // The "min" value to assign to the listed related (mod#code). Usage depends on the property function (See the "func" field on Properties.txt).
	ModMin2        int    `csv:"mod2min"`        // The "min" value to assign to the listed related (mod#code). Usage depends on the property function (See the "func" field on Properties.txt).
	ModMin3        int    `csv:"mod3min"`        // The "min" value to assign to the listed related (mod#code). Usage depends on the property function (See the "func" field on Properties.txt).
	ModMaxe1       int    `csv:"mod1max"`        // The "max" value to assign to the listed related (mod#code). Usage depends on the property function (See the "func" field on Properties.txt).
	ModMaxe2       int    `csv:"mod2max"`        // The "max" value to assign to the listed related (mod#code). Usage depends on the property function (See the "func" field on Properties.txt).
	ModMaxe3       int    `csv:"mod3max"`        // The "max" value to assign to the listed related (mod#code). Usage depends on the property function (See the "func" field on Properties.txt).
	TransformColor string `csv:"transformcolor"` // Controls the color change of the item after spawning with this item affix.
	ITypeCode1     string `csv:"itype1"`         // Controls what Item Types are allowed to spawn with this item affix.
	ITypeCode2     string `csv:"itype2"`         // Controls what Item Types are allowed to spawn with this item affix.
	ITypeCode3     string `csv:"itype3"`         // Controls what Item Types are allowed to spawn with this item affix.
	ITypeCode4     string `csv:"itype4"`         // Controls what Item Types are allowed to spawn with this item affix.
	ITypeCode5     string `csv:"itype5"`         // Controls what Item Types are allowed to spawn with this item affix.
	ITypeCode6     string `csv:"itype6"`         // Controls what Item Types are allowed to spawn with this item affix.
	ITypeCode7     string `csv:"itype7"`         // Controls what Item Types are allowed to spawn with this item affix.
	ETypeCode1     string `csv:"etype1,"`        // Controls what Item Types are forbidden to spawn with this item affix.
	ETypeCode2     string `csv:"etype2,"`        // Controls what Item Types are forbidden to spawn with this item affix.
	ETypeCode3     string `csv:"etype3,"`        // Controls what Item Types are forbidden to spawn with this item affix.
	ETypeCode4     string `csv:"etype4,"`        // Controls what Item Types are forbidden to spawn with this item affix.
	ETypeCode5     string `csv:"etype5,"`        // Controls what Item Types are forbidden to spawn with this item affix.
	Multiply       int    `csv:"multiply"`       // Multiplicative modifier for the item's buy and sell costs, based on the item affix (Calculated in 1024ths for buy cost and 4096ths for sell cost).
	Add            int    `csv:"add"`            // Flat integer modification to the item's buy and sell costs, based on the item affix.
}

type AutomagicColor string // represents the color change of the item after spawning with this item affix.

const (
	AutomagicColorNone         AutomagicColor = ""     // represents no color change.
	AutomagicColorWhite        AutomagicColor = "whit" // represents white color.
	AutomagicColorLightGrey    AutomagicColor = "lgry" // represents light grey color.
	AutomagicColorDarkGrey     AutomagicColor = "dgry" // represents dark grey color.
	AutomagicColorBlack        AutomagicColor = "blac" // represents black color.
	AutomagicColorLightBlue    AutomagicColor = "lblu" // represents light blue color.
	AutomagicColorDarkBlue     AutomagicColor = "dblu" // represents dark blue color.
	AutomagicColorCrystalBlue  AutomagicColor = "cblu" // represents crystal blue color.
	AutomagicColorLightRed     AutomagicColor = "lred" // represents light red color.
	AutomagicColorDarkRed      AutomagicColor = "dred" // represents dark red color.
	AutomagicColorCrystalRed   AutomagicColor = "cred" // represents crystal red color.
	AutomagicColorLightGreen   AutomagicColor = "lgrn" // represents light green color.
	AutomagicColorDarkGreen    AutomagicColor = "dgrn" // represents dark green color.
	AutomagicColorCrystalGreen AutomagicColor = "cgrn" // represents crystal green color.
	AutomagicColorLightYellow  AutomagicColor = "lyel" // represents light yellow color.
	AutomagicColorDarkYellow   AutomagicColor = "dyel" // represents dark yellow color.
	AutomagicColorLightGold    AutomagicColor = "lgld" // represents light gold color.
	AutomagicColorDarkGold     AutomagicColor = "dgld" // represents dark gold color.
	AutomagicColorLightPurple  AutomagicColor = "lpur" // represents light purple color.
	AutomagicColorDarkPurple   AutomagicColor = "dpur" // represents dark purple color.
	AutomagicColorOrange       AutomagicColor = "oran" // represents orange color.
	AutomagicColorBrightWhite  AutomagicColor = "bwht" // represents bright white color.
)
