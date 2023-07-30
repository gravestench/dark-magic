package models

// ItemStatCost represents the functionalities for each possible stat on a unit.
type ItemStatCost struct {
	Stat           string `csv:"Stat"`            // Defines the unique pointer for this stat, used in other files.
	SendOther      int    `csv:"Send Other"`      // If 1, add the stat to a new monster if it has no state and an item mask; else ignore.
	Signed         int    `csv:"Signed"`          // If 1, stat is treated as a signed integer; else it's an unsigned integer. Applies to stats with state bits.
	SendBits       int    `csv:"Send Bits"`       // Controls how many bits of data for the stat to send to the game client.
	SendParamBits  int    `csv:"Send Param Bits"` // Controls how many bits of data for the stat's parameter value to send to the client for a unit.
	UpdateAnimRate int    `csv:"UpdateAnimRate"`  // If 1, the stat will adjust the unit's speed when changed; else ignore.
	Saved          int    `csv:"Saved"`           // If 1, this state will be stored in the Character Save file; else ignore.
	CSvSigned      int    `csv:"CSvSigned"`       // If 1, the stat will be saved as a signed integer in the Character Save file; else it's unsigned.
	CSvBits        int    `csv:"CSvBits"`         // Controls how many bits of data for the stat to save in the Character Save file.
	CSvParam       int    `csv:"CSvParam"`        // Controls how many bits of data for the stat's parameter value to save in the Character Save file.
	FCallback      int    `csv:"fCallback"`       // If 1, any changes to the stat will call the Callback function to update character's states, skills, or item events.
	FMin           int    `csv:"fMin"`            // If 1, the stat has a minimum value that cannot be reduced further (See MinAccr); else ignore.
	MinAccr        int    `csv:"MinAccr"`         // The minimum value of a stat. Used if FMin is enabled.
	Encode         int    `csv:"Encode"`          // Controls how the stat will modify an item's buy, sell, and repair costs.
	Add            int    `csv:"Add"`             // Flat integer modification to the Unique item's buy, sell, and repair costs. Added after Multiply field.
	Multiply       int    `csv:"Multiply"`        // Multiplicative modifier for the item's buy, sell, and repair costs.
	ValShift       int    `csv:"ValShift"`        // Shifts the stat's input value by a number of bits for calculations.
	SaveBits109    int    `csv:"1.09-Save Bits"`  // Controls how many bits are allocated for the overall size of the stat when saving/reading an item from a Character Save in 1.09d or older.
	SaveAdd109     int    `csv:"1.09-Save Add"`   // Controls how many bits are allocated for the stat's value when saving/reading an item from a Character Save in 1.09d or older.
	SaveBits       int    `csv:"Save Bits"`       // Controls how many bits are allocated for the overall size of the stat when saving/reading an item from a Character Save.
	SaveAdd        int    `csv:"Save Add"`        // Controls how many bits are allocated for the stat's value when saving/reading an item from a Character Save.
	SaveParamBits  int    `csv:"Save Param Bits"` // Controls how many bits for the stat's parameter value to use when saving/reading an item from a Character Save.
	KeepZero       int    `csv:"keepzero"`        // If 1, this stat remains on the change list even if its value is 0; else ignore.
	Op             string `csv:"op"`              // Stat operator for advanced stat modification when calculating the value of a stat.
	OpParam        string `csv:"op param"`        // Possible parameter value for the Op function.
	OpBase         string `csv:"op base"`         // Possible parameter value for the Op function.
	OpStat1        string `csv:"op stat1"`        // Possible parameter value for the Op function.
	OpStat2        string `csv:"op stat2"`        // Possible parameter value for the Op function.
	OpStat3        string `csv:"op stat3"`        // Possible parameter value for the Op function.
	Direct         int    `csv:"direct"`          // If 1, the stat updates in relation to its maxstat field to ensure it never exceeds that value in certain skill functions.
	MaxStat        string `csv:"maxstat"`         // Controls which stat is associated with this stat to be treated as the maximum version of this stat.
	DamageRelated  int    `csv:"damagerelated"`   // If 1, this stat is exclusive to the item and will not add to the unit; else it always adds to the unit.
	ItemEvent1     string `csv:"itemevent1"`      // Event that will activate the specified function defined by itemeventfunc1.
	ItemEvent2     string `csv:"itemevent2"`      // Event that will activate the specified function defined by itemeventfunc2.
	ItemEventFunc1 string `csv:"itemeventfunc1"`  // Function to use after the related item event occurred.
	ItemEventFunc2 string `csv:"itemeventfunc2"`  // Function to use after the related item event occurred.
	DescPriority   string `csv:"descpriority"`    // Controls how this stat is sorted in item tooltips.
	DescFunc       string `csv:"descfunc"`        // Controls how the stat is displayed in tooltips.
	DescVal        string `csv:"descval"`         // Possible parameter value for the DescFunc function. Controls how the value of the stat is displayed.
	DescStrPos     string `csv:"descstrpos"`      // Possible parameter value for the DescFunc function. Uses a string to display the item stat in a tooltip when its value is positive.
	DescStrNeg     string `csv:"descstrneg"`      // Possible parameter value for the DescFunc function. Uses a string to display the item stat in a tooltip when its value is negative.
	DescStr2       string `csv:"descstr2"`        // Possible parameter value for the DescFunc function. Uses a string to append to an item stat's string in a tooltip.
	DGrp           int    `csv:"dgrp"`            // Assigns the stat to a group ID value.
	DGrpFunc       int    `csv:"dgrpfunc"`        // Controls how the shared group of stats is displayed in tooltips.
	DGrpVal        int    `csv:"dgrpval"`         // Possible parameter value for the DGrpFunc function. Controls how the value of the stat is displayed.
	DGrpStrPos     string `csv:"dgrpstrpos"`      // Possible parameter value for the DGrpFunc function. Uses a string to display the item stat in a tooltip when its value is positive.
	DGrpStrNeg     string `csv:"dgrpstrneg"`      // Possible parameter value for the DGrpFunc function. Uses a string to display the item stat in a tooltip when its value is negative.
	DGrpStr2       string `csv:"dgrpstr2"`        // Possible parameter value for the DGrpFunc function. Uses a string to append to an item stat's string in a tooltip.
	Stuff          int    `csv:"stuff"`           // Used as a bit shift value for handling the conversion of skill IDs and skill levels to bit values for the stat.
	AdvDisplay     int    `csv:"advdisplay"`      // Controls how the stat appears in the Advanced Stats UI.
}
