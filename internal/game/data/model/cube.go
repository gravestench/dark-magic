package models

// OpID represents the operation ID for additional input requirement
type OpID int

// OpCode represents the operation code for class restriction
type OpCode string

// CubeItemType represents the input and output item types
type CubeItemType string

// OutputCode represents the output code for creating output items
type OutputCode string

// CubeRecipe holds the recipes for the Horadric Cube
type CubeRecipe struct {
	Description string `csv:"description"` // This is a reference field to define the cube recipe
	// Boolean field. If equals 1, then the recipe can be used in-game. If equals 0, then the recipe cannot be used
	// in-game.
	Enabled bool `csv:"enabled"`
	// Integer field. The first ladder season this cube recipe can be made on (inclusive). If blank or 0 then it is
	// available in non-ladder.
	FirstLadderSeason int `csv:"ladder"`
	// The minimum game difficulty to use the recipe (0 = All Game Difficulties | 1 = Nightmare and Hell Difficulty only |
	// 2 = Hell Difficulty only)
	MinDifficulty int `csv:"min diff"`
	// Defines which game version to use this recipe (0 = Classic mode | 100 = Expansion mode)
	Version int `csv:"version"`
	// Uses a function as an additional input requirement for the recipe
	Op    OpCode `csv:"op"`
	Param string `csv:"param"` // Parameters for the "op" function
	Value string `csv:"value"` // Value for the "op" function
	Class string `csv:"class"` // Defines the recipe to be only usable by a defined class
	// Controls the number of items that need to be inside the cube for the recipe
	NumInputs int `csv:"numinputs"`
	// Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters
	// by adding a comma “,” to the input and using a code.
	Input1 string `csv:"input 1"`
	// Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters
	// by adding a comma “,” to the input and using a code.
	Input2 string `csv:"input 2"`
	// Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters
	// by adding a comma “,” to the input and using a code.
	Input3 string `csv:"input 3"`
	// Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters
	// by adding a comma “,” to the input and using a code.
	Input4 string `csv:"input 4"`
	// Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters
	// by adding a comma “,” to the input and using a code.
	Input5 string `csv:"input 5"`
	// Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters
	// by adding a comma “,” to the input and using a code.
	Input6 string `csv:"input 6"`
	// Controls what items are required for the recipe. Uses the item’s unique code. Users can also add input parameters
	// by adding a comma “,” to the input and using a code.
	Input7 string `csv:"input 7"`
	// Controls the first output item. Uses the item’s unique code. Users can also add output parameters by adding a
	// comma “,” to the output and using a code.
	Output OutputCode `csv:"output"`
	// Forces the output item level to be a specific level. If this field is used, then ignore the “plvl” and
	// “ilvl” fields.
	OutputLevel int `csv:"lvl"`
	// This is a numeric ratio that gets multiplied with the current player’s level, to add to the output item’s level
	// requirement
	OutputPlayerLvl float64 `csv:"plvl"`
	// This is a numeric ratio that gets multiplied with “input 1” item’s level, to add to the output item’s level
	// requirement
	OutputInputLvl float64 `csv:"ilvl"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputMod1 string `csv:"mod 1"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputModChance1 float64 `csv:"mod 1 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModParam1 string `csv:"mod 1 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMin1 float64 `csv:"mod 1 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMax1 float64 `csv:"mod 1 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputMod2 string `csv:"mod 2"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputModChance2 float64 `csv:"mod 2 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModParam2 string `csv:"mod 2 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMin2 float64 `csv:"mod 2 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMax2 float64 `csv:"mod 2 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputMod3 string `csv:"mod 3"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputModChance3 float64 `csv:"mod 3 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModParam3 string `csv:"mod 3 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMin3 float64 `csv:"mod 3 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMax3 float64 `csv:"mod 3 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputMod4 string `csv:"mod 4"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputModChance4 float64 `csv:"mod 4 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModParam4 string `csv:"mod 4 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMin4 float64 `csv:"mod 4 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMax4 float64 `csv:"mod 4 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputMod5 string `csv:"mod 5"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputModChance5 float64 `csv:"mod 5 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModParam5 string `csv:"mod 5 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMin5 float64 `csv:"mod 5 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputModMax5 float64 `csv:"mod 5 max"`
	// Controls the second output item. Uses the item’s unique code. Users can also add output parameters by adding a
	// comma “,” to the output and using a code. (See “output” for more details)
	OutputB OutputCode `csv:"output b"`
	// Forces the output item level to be a specific level. If this field is used, then ignore the “plvl” and
	// “ilvl” fields.
	OutputBLevel int `csv:"b lvl"`
	// This is a numeric ratio that gets multiplied with the current player’s level, to add to the output item’s level
	// requirement
	OutputBPLvl float64 `csv:"b plvl"`
	// This is a numeric ratio that gets multiplied with “input 2” item’s level, to add to the output item’s level
	// requirement
	OutputBILvl float64 `csv:"b ilvl"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBMod1 string `csv:"b mod 1"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputBModChance1 float64 `csv:"b mod 1 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModParam1 string `csv:"b mod 1 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMin1 float64 `csv:"b mod 1 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMax1 float64 `csv:"b mod 1 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBMod2 string `csv:"b mod 2"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputBModChance2 float64 `csv:"b mod 2 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModParam2 string `csv:"b mod 2 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMin2 float64 `csv:"b mod 2 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMax2 float64 `csv:"b mod 2 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBMod3 string `csv:"b mod 3"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputBModChance3 float64 `csv:"b mod 3 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModParam3 string `csv:"b mod 3 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMin3 float64 `csv:"b mod 3 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMax3 float64 `csv:"b mod 3 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBMod4 string `csv:"b mod 4"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputBModChance4 float64 `csv:"b mod 4 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModParam4 string `csv:"b mod 4 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMin4 float64 `csv:"b mod 4 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMax4 float64 `csv:"b mod 4 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputBMod5 string `csv:"b mod 5"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputBModChance5 float64 `csv:"b mod 5 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModParam5 string `csv:"b mod 5 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMin5 float64 `csv:"b mod 5 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputBModMax5 float64 `csv:"b mod 5 max"`
	// Controls the third output item. Uses the item’s unique code. Users can also add output parameters by adding a
	// comma “,” to the output and using a code. (See “output” for more details)
	OutputC OutputCode `csv:"output c"`
	// Forces the output item level to be a specific level. If this field is used, then ignore the “plvl” and
	// “ilvl” fields.
	OutputCLevel int `csv:"c lvl"`
	// This is a numeric ratio that gets multiplied with the current player’s level, to add to the output item’s level
	// requirement
	OutputCPLvl float64 `csv:"c plvl"`
	// This is a numeric ratio that gets multiplied with “input 3” item’s level, to add to the output item’s level
	// requirement
	OutputCILvl float64 `csv:"c ilvl"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCMod1 string `csv:"c mod 1"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputCModChance1 float64 `csv:"c mod 1 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModParam1 string `csv:"c mod 1 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMin1 float64 `csv:"c mod 1 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMax1 float64 `csv:"c mod 1 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCMod2 string `csv:"c mod 2"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputCModChance2 float64 `csv:"c mod 2 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModParam2 string `csv:"c mod 2 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMin2 float64 `csv:"c mod 2 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMax2 float64 `csv:"c mod 2 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCMod3 string `csv:"c mod 3"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputCModChance3 float64 `csv:"c mod 3 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModParam3 string `csv:"c mod 3 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMin3 float64 `csv:"c mod 3 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMax3 float64 `csv:"c mod 3 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCMod4 string `csv:"c mod 4"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputCModChance4 float64 `csv:"c mod 4 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModParam4 string `csv:"c mod 4 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMin4 float64 `csv:"c mod 4 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMax4 float64 `csv:"c mod 4 max"`
	// Controls the output item properties (Uses the “code” field from Properties.txt)
	OutputCMod5 string `csv:"c mod 5"`
	// The percent chance that the property will be assigned. If this equals 0, then the ItemProperty will always be
	// assigned.
	OutputCModChance5 float64 `csv:"c mod 5 chance"`
	// The “parameter” value associated with the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModParam5 string `csv:"c mod 5 param"`
	// The “min” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMin5 float64 `csv:"c mod 5 min"`
	// The “max” value to assign to the listed property (mod). Usage depends on the property function (See the
	// “func” field on Properties.txt)
	OutputCModMax5 float64 `csv:"c mod 5 max"`
	// *eol: End of line marker
}
