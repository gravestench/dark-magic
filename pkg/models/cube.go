package models

// OpID represents the operation ID for additional input requirement
type OpID int

const (
	OpNone OpID = iota
	OpDayOfMonthLessThan
	OpDayOfMonthGreaterThan
	OpDayOfWeekNotEqual
	OpPlayerStatGreaterThan
	OpPlayerStatLessThan
	OpPlayerStatNotEqual
	OpPlayerStatEqual
	OpPlayerBaseStatGreaterThan
	OpPlayerBaseStatLessThan
	OpPlayerBaseStatNotEqual
	OpPlayerBaseStatEqual
	OpPlayerNonBaseStatGreaterThan
	OpPlayerNonBaseStatLessThan
	OpPlayerNonBaseStatNotEqual
	OpPlayerNonBaseStatEqual
	OpInputItemStatGreaterThan
	OpInputItemStatLessThan
	OpInputItemStatNotEqual
	OpInputItemStatEqual
	OpInputItemBaseStatGreaterThan
	OpInputItemBaseStatLessThan
	OpInputItemBaseStatNotEqual
	OpInputItemBaseStatEqual
	OpInputItemNonBaseStatGreaterThan
	OpInputItemNonBaseStatLessThan
	OpInputItemNonBaseStatNotEqual
	OpInputItemNonBaseStatEqual
	OpItemModClassNotEqual
	OpQuestDiffCheck
)

// OpCode represents the operation code for class restriction
type OpCode string

const (
	OpAnyClass        OpCode = ""
	OpAmazonOnly      OpCode = "ama"
	OpBarbarianOnly   OpCode = "bar"
	OpPaladinOnly     OpCode = "pal"
	OpNecromancerOnly OpCode = "nec"
	OpSorceressOnly   OpCode = "sor"
	OpDruidOnly       OpCode = "dru"
	OpAssassinOnly    OpCode = "ass"
)

// CubeItemType represents the input and output item types
type CubeItemType string

const (
	ItemTypeQuantity    CubeItemType = "qty=#"
	ItemTypeLow         CubeItemType = "low"
	ItemTypeNormal      CubeItemType = "nor"
	ItemTypeHigh        CubeItemType = "hiq"
	ItemTypeMagic       CubeItemType = "mag"
	ItemTypeSet         CubeItemType = "set"
	ItemTypeRare        CubeItemType = "rar"
	ItemTypeUnique      CubeItemType = "uni"
	ItemTypeCrafted     CubeItemType = "crf"
	ItemTypeTempered    CubeItemType = "tmp"
	ItemTypeNoSockets   CubeItemType = "nos"
	ItemTypeSockets     CubeItemType = "sock=#"
	ItemTypeNotEthereal CubeItemType = "noe"
	ItemTypeEthereal    CubeItemType = "eth"
	ItemTypeUpgradable  CubeItemType = "upg"
	ItemTypeBasic       CubeItemType = "bas"
	ItemTypeExceptional CubeItemType = "exc"
	ItemTypeElite       CubeItemType = "eli"
	ItemTypeNotRuneWord CubeItemType = "nru"
)

// OutputCode represents the output code for creating output items
type OutputCode string

const (
	OutputCowPortal               OutputCode = "Cow Portal"
	OutputPandemoniumPortal       OutputCode = "Pandemonium Portal"
	OutputPandemoniumFinalePortal OutputCode = "Pandemonium Finale Portal"
	OutputRedPortal               OutputCode = "Red Portal"
	OutputUseType                 OutputCode = "usetype"
	OutputUseItem                 OutputCode = "useitem"
	OutputQuantity                OutputCode = "qty=#"
	OutputPrefix                  OutputCode = "pre=#"
	OutputSuffix                  OutputCode = "suf=#"
	OutputLow                     OutputCode = "low"
	OutputNormal                  OutputCode = "nor"
	OutputHigh                    OutputCode = "hiq"
	OutputMagic                   OutputCode = "mag"
	OutputSet                     OutputCode = "set"
	OutputRare                    OutputCode = "rar"
	OutputUnique                  OutputCode = "uni"
	OutputCrafted                 OutputCode = "crf"
	OutputTempered                OutputCode = "tmp"
	OutputEthereal                OutputCode = "eth"
	OutputSockets                 OutputCode = "sock"
	OutputModifiers               OutputCode = "mod"
	OutputDestroyGemsRunesJewels  OutputCode = "uns"
	OutputRemoveGemsRunesJewels   OutputCode = "rem"
	OutputRegenerateUnique        OutputCode = "reg"
	OutputExceptional             OutputCode = "exc"
	OutputElite                   OutputCode = "eli"
	OutputRepair                  OutputCode = "rep"
	OutputRecharge                OutputCode = "rch"
	OutputLevel                   OutputCode = "lvl=#"
)

// CubeRecipe holds the recipes for the Horadric Cube
type CubeRecipe struct {
	Description       string     `csv:"description"`    // This is a reference field to define the cube recipe
	Enabled           bool       `csv:"enabled"`        // Boolean field. If equals 1, then the recipe can be used in-game. If equals 0, then the recipe cannot be used in-game.
	FirstLadderSeason int        `csv:"ladder"`         // Integer field. The first ladder season this cube recipe can be made on (inclusive). If blank or 0 then it is available in non-ladder.
	MinDifficulty     int        `csv:"min diff"`       // The minimum game difficulty to use the recipe (0 = All Game Difficulties | 1 = Nightmare and Hell Difficulty only | 2 = Hell Difficulty only)
	Version           int        `csv:"version"`        // Defines which game version to use this recipe (0 = Classic mode | 100 = Expansion mode)
	Op                OpCode     `csv:"op"`             // Uses a function as an additional input requirement for the recipe
	Param             string     `csv:"param"`          // Parameters for the "op" function
	Value             string     `csv:"value"`          // Value for the "op" function
	Class             string     `csv:"class"`          // Defines the recipe to be only usable by a defined class
	NumInputs         int        `csv:"numinputs"`      // Controls the number of items that need to be inside the cube for the recipe
	Input1            string     `csv:"input 1"`        // Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters by adding a comma “,” to the input and using a code.
	Input2            string     `csv:"input 2"`        // Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters by adding a comma “,” to the input and using a code.
	Input3            string     `csv:"input 3"`        // Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters by adding a comma “,” to the input and using a code.
	Input4            string     `csv:"input 4"`        // Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters by adding a comma “,” to the input and using a code.
	Input5            string     `csv:"input 5"`        // Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters by adding a comma “,” to the input and using a code.
	Input6            string     `csv:"input 6"`        // Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters by adding a comma “,” to the input and using a code.
	Input7            string     `csv:"input 7"`        // Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters by adding a comma “,” to the input and using a code.
	Output            OutputCode `csv:"output"`         // Controls the first output item. Uses the item’s unique code. Users can also add output parameters by adding a comma “,” to the output and using a code.
	OutputLevel       int        `csv:"lvl"`            // Forces the output item level to be a specific level. If this field is used, then ignore the “plvl” and “ilvl” fields.
	OutputPlayerLvl   float64    `csv:"plvl"`           // This is a numeric ratio that gets multiplied with the current player’s level, to add to the output item’s level requirement
	OutputInputLvl    float64    `csv:"ilvl"`           // This is a numeric ratio that gets multiplied with “input 1” item’s level, to add to the output item’s level requirement
	OutputMod1        string     `csv:"mod 1"`          // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputModChance1  float64    `csv:"mod 1 chance"`   // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputModParam1   string     `csv:"mod 1 param"`    // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMin1     float64    `csv:"mod 1 min"`      // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMax1     float64    `csv:"mod 1 max"`      // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputMod2        string     `csv:"mod 2"`          // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputModChance2  float64    `csv:"mod 2 chance"`   // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputModParam2   string     `csv:"mod 2 param"`    // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMin2     float64    `csv:"mod 2 min"`      // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMax2     float64    `csv:"mod 2 max"`      // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputMod3        string     `csv:"mod 3"`          // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputModChance3  float64    `csv:"mod 3 chance"`   // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputModParam3   string     `csv:"mod 3 param"`    // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMin3     float64    `csv:"mod 3 min"`      // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMax3     float64    `csv:"mod 3 max"`      // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputMod4        string     `csv:"mod 4"`          // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputModChance4  float64    `csv:"mod 4 chance"`   // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputModParam4   string     `csv:"mod 4 param"`    // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMin4     float64    `csv:"mod 4 min"`      // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMax4     float64    `csv:"mod 4 max"`      // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputMod5        string     `csv:"mod 5"`          // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputModChance5  float64    `csv:"mod 5 chance"`   // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputModParam5   string     `csv:"mod 5 param"`    // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMin5     float64    `csv:"mod 5 min"`      // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputModMax5     float64    `csv:"mod 5 max"`      // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputB           OutputCode `csv:"output b"`       // Controls the second output item. Uses the item’s unique code. Users can also add output parameters by adding a comma “,” to the output and using a code. (See “output” for more details)
	OutputBLevel      int        `csv:"b lvl"`          // Forces the output item level to be a specific level. If this field is used, then ignore the “plvl” and “ilvl” fields.
	OutputBPLvl       float64    `csv:"b plvl"`         // This is a numeric ratio that gets multiplied with the current player’s level, to add to the output item’s level requirement
	OutputBILvl       float64    `csv:"b ilvl"`         // This is a numeric ratio that gets multiplied with “input 2” item’s level, to add to the output item’s level requirement
	OutputBMod1       string     `csv:"b mod 1"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBModChance1 float64    `csv:"b mod 1 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputBModParam1  string     `csv:"b mod 1 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMin1    float64    `csv:"b mod 1 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMax1    float64    `csv:"b mod 1 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBMod2       string     `csv:"b mod 2"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBModChance2 float64    `csv:"b mod 2 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputBModParam2  string     `csv:"b mod 2 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMin2    float64    `csv:"b mod 2 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMax2    float64    `csv:"b mod 2 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBMod3       string     `csv:"b mod 3"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBModChance3 float64    `csv:"b mod 3 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputBModParam3  string     `csv:"b mod 3 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMin3    float64    `csv:"b mod 3 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMax3    float64    `csv:"b mod 3 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBMod4       string     `csv:"b mod 4"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBModChance4 float64    `csv:"b mod 4 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputBModParam4  string     `csv:"b mod 4 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMin4    float64    `csv:"b mod 4 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMax4    float64    `csv:"b mod 4 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBMod5       string     `csv:"b mod 5"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBModChance5 float64    `csv:"b mod 5 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputBModParam5  string     `csv:"b mod 5 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMin5    float64    `csv:"b mod 5 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputBModMax5    float64    `csv:"b mod 5 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputC           OutputCode `csv:"output c"`       // Controls the third output item. Uses the item’s unique code. Users can also add output parameters by adding a comma “,” to the output and using a code. (See “output” for more details)
	OutputCLevel      int        `csv:"c lvl"`          // Forces the output item level to be a specific level. If this field is used, then ignore the “plvl” and “ilvl” fields.
	OutputCPLvl       float64    `csv:"c plvl"`         // This is a numeric ratio that gets multiplied with the current player’s level, to add to the output item’s level requirement
	OutputCILvl       float64    `csv:"c ilvl"`         // This is a numeric ratio that gets multiplied with “input 3” item’s level, to add to the output item’s level requirement
	OutputCMod1       string     `csv:"c mod 1"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCModChance1 float64    `csv:"c mod 1 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputCModParam1  string     `csv:"c mod 1 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMin1    float64    `csv:"c mod 1 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMax1    float64    `csv:"c mod 1 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCMod2       string     `csv:"c mod 2"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCModChance2 float64    `csv:"c mod 2 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputCModParam2  string     `csv:"c mod 2 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMin2    float64    `csv:"c mod 2 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMax2    float64    `csv:"c mod 2 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCMod3       string     `csv:"c mod 3"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCModChance3 float64    `csv:"c mod 3 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputCModParam3  string     `csv:"c mod 3 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMin3    float64    `csv:"c mod 3 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMax3    float64    `csv:"c mod 3 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCMod4       string     `csv:"c mod 4"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCModChance4 float64    `csv:"c mod 4 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputCModParam4  string     `csv:"c mod 4 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMin4    float64    `csv:"c mod 4 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMax4    float64    `csv:"c mod 4 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCMod5       string     `csv:"c mod 5"`        // Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCModChance5 float64    `csv:"c mod 5 chance"` // The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be assigned.
	OutputCModParam5  string     `csv:"c mod 5 param"`  // The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMin5    float64    `csv:"c mod 5 min"`    // The “min” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	OutputCModMax5    float64    `csv:"c mod 5 max"`    // The “max” value to assign to the listed property (mod). Usage depends on the property function (See the “func” field on Properties.txt)
	// *eol: End of line marker
}
