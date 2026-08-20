package models

// ItemStatCost represents the functionalities for each possible stat on a unit.
type ItemStatCost struct {
	Name string `csv:"Stat"` // Defines the unique pointer for this stat, used in other files.
	// If 1, add the stat to a new monster if it has no state and an item mask; else ignore.
	SendOther int `csv:"Send Other"`
	// If 1, stat is treated as a signed integer; else it's an unsigned integer. Applies to stats with state bits.
	Signed int `csv:"Signed"`
	// Controls how many bits of data for the stat to send to the game client.
	SendBits int `csv:"Send Bits"`
	// Controls how many bits of data for the stat's parameter value to send to the client for a unit.
	SendParamBits  int `csv:"Send Param Bits"`
	UpdateAnimRate int `csv:"UpdateAnimRate"` // If 1, the stat will adjust the unit's speed when changed; else ignore.
	// If 1, this state will be stored in the Character Save file; else ignore.
	Saved int `csv:"Saved"`
	// If 1, the stat will be saved as a signed integer in the Character Save file; else it's unsigned.
	CSvSigned int `csv:"CSvSigned"`
	// Controls how many bits of data for the stat to save in the Character Save file.
	CSvBits int `csv:"CSvBits"`
	// Controls how many bits of data for the stat's parameter value to save in the Character Save file.
	CSvParam int `csv:"CSvParam"`
	// If 1, any changes to the stat will call the Callback function to update character's states, skills, or item events.
	FCallback int `csv:"fCallback"`
	// If 1, the stat has a minimum value that cannot be reduced further (See MinAccr); else ignore.
	FMin    int `csv:"fMin"`
	MinAccr int `csv:"MinAccr"` // The minimum value of a stat. Used if FMin is enabled.
	// Controls how the stat will modify an item's buy, sell, and repair costs.
	Encode int `csv:"Encode"`
	// Flat integer modification to the Unique item's buy, sell, and repair costs. Added after Multiply field.
	Add      int `csv:"Add"`
	Multiply int `csv:"Multiply"` // Multiplicative modifier for the item's buy, sell, and repair costs.
	ValShift int `csv:"ValShift"` // Shifts the stat's input value by a number of bits for calculations.
	// Controls how many bits are allocated for the overall size of the stat when saving/reading an item from a Character
	// Save in 1.09d or older.
	SaveBits109 int `csv:"1.09-Save Bits"`
	// Controls how many bits are allocated for the stat's value when saving/reading an item from a Character Save in 1.09d
	// or older.
	SaveAdd109 int `csv:"1.09-Save Add"`
	// Controls how many bits are allocated for the overall size of the stat when saving/reading an item from a Character
	// Save.
	SaveBits int `csv:"Save Bits"`
	// Controls how many bits are allocated for the stat's value when saving/reading an item from a Character Save.
	SaveAdd int `csv:"Save Add"`
	// Controls how many bits for the stat's parameter value to use when saving/reading an item from a Character Save.
	SaveParamBits int `csv:"Save Param Bits"`
	// If 1, this stat remains on the change list even if its value is 0; else ignore.
	KeepZero int `csv:"keepzero"`
	// Name operator for advanced stat modification when calculating the value of a stat.
	Op      string `csv:"op"`
	OpParam string `csv:"op param"` // Possible parameter value for the Op function.
	OpBase  string `csv:"op base"`  // Possible parameter value for the Op function.
	OpStat1 string `csv:"op stat1"` // Possible parameter value for the Op function.
	OpStat2 string `csv:"op stat2"` // Possible parameter value for the Op function.
	OpStat3 string `csv:"op stat3"` // Possible parameter value for the Op function.
	// If 1, the stat updates in relation to its maxstat field to ensure it never exceeds that value in certain skill
	// functions.
	Direct int `csv:"direct"`
	// Controls which stat is associated with this stat to be treated as the maximum version of this stat.
	MaxStat string `csv:"maxstat"`
	// If 1, this stat is exclusive to the item and will not add to the unit; else it always adds to the unit.
	DamageRelated int `csv:"damagerelated"`
	// Event that will activate the specified function defined by itemeventfunc1.
	ItemEvent1 string `csv:"itemevent1"`
	// Event that will activate the specified function defined by itemeventfunc2.
	ItemEvent2     string `csv:"itemevent2"`
	ItemEventFunc1 string `csv:"itemeventfunc1"` // Function to use after the related item event occurred.
	ItemEventFunc2 string `csv:"itemeventfunc2"` // Function to use after the related item event occurred.
	DescPriority   int    `csv:"descpriority"`   // Controls how this stat is sorted in item tooltips.
	DescFunc       int    `csv:"descfunc"`       // Controls how the stat is displayed in tooltips.
	// Possible parameter value for the DescFunc function. Controls how the value of the stat is displayed.
	DescVal string `csv:"descval"`
	// Possible parameter value for the DescFunc function. Uses a string to display the item stat in a tooltip when its
	// value is positive.
	DescStrPos string `csv:"descstrpos"`
	// Possible parameter value for the DescFunc function. Uses a string to display the item stat in a tooltip when its
	// value is negative.
	DescStrNeg string `csv:"descstrneg"`
	// Possible parameter value for the DescFunc function. Uses a string to append to an item stat's string in a tooltip.
	DescStr2 string `csv:"descstr2"`
	DGrp     int    `csv:"dgrp"`     // Assigns the stat to a group ID value.
	DGrpFunc int    `csv:"dgrpfunc"` // Controls how the shared group of stats is displayed in tooltips.
	// Possible parameter value for the DGrpFunc function. Controls how the value of the stat is displayed.
	DGrpVal int `csv:"dgrpval"`
	// Possible parameter value for the DGrpFunc function. Uses a string to display the item stat in a tooltip when its
	// value is positive.
	DGrpStrPos string `csv:"dgrpstrpos"`
	// Possible parameter value for the DGrpFunc function. Uses a string to display the item stat in a tooltip when its
	// value is negative.
	DGrpStrNeg string `csv:"dgrpstrneg"`
	// Possible parameter value for the DGrpFunc function. Uses a string to append to an item stat's string in a tooltip.
	DGrpStr2 string `csv:"dgrpstr2"`
	// Used as a bit shift value for handling the conversion of skill IDs and skill levels to bit values for the stat.
	Stuff      int `csv:"stuff"`
	AdvDisplay int `csv:"advdisplay"` // Controls how the stat appears in the Advanced Stats UI.
}
