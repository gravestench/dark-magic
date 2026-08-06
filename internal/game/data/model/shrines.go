package models

// Shrine represents the data structure for the shrines.txt file.
type Shrine struct {
	ShrineType     string `csv:"Shrine Type"`           // The type of shrine
	ShrineName     string `csv:"Shrine name"`           // The name of the shrine
	Effect         string `csv:"Effect"`                // Description of the shrine's effect
	Code           int    `csv:"Code"`                  // Code function used to define the shrine's function
	Arg0           int    `csv:"Arg0"`                  // Parameter 1 for the shrine's function
	Arg1           int    `csv:"Arg1"`                  // Parameter 2 for the shrine's function
	DurationFrames int    `csv:"Duration in frames"`    // Duration of the shrine's effects in frames (1 second = 25 frames)
	ResetTimeMin   int    `csv:"reset time in minutes"` // Time in minutes before the shrine is available to use again
	Rarity         int    `csv:"rarity"`                // Rarity of the shrine
	ViewName       string `csv:"view name"`             // View name of the shrine
	NiftyPhrase    string `csv:"niftyphrase"`           // Activation phrase displayed when the shrine is used
	EffectClass    int    `csv:"effectclass"`           // The shrine's archetype involved in calculating region stats
	LevelMin       int    `csv:"LevelMin"`              // Minimum area level where the shrine can spawn
}

// this might be the D2 resurrected revision...
//// Shrine represents the data structure for the shrines.txt file.
//type Shrine struct {
//	Name         string `csv:"Name"`         // Reference field to define the Shrine index.
//	Code         int    `csv:"Code"`         // Code function used to define the Shrine's function.
//	Arg0         int    `csv:"Arg0"`         // Parameter 1 for the Shrine's function.
//	Arg1         int    `csv:"Arg1"`         // Parameter 2 for the Shrine's function.
//	Duration     int    `csv:"Duration"`     // Duration of the effects of the Shrine (Calculated in Frames, where 25 Frames = 1 Second).
//	ResetTime    int    `csv:"reset time"`   // Controls the amount of time before the Shrine is available to use again. Each value of 1 equals 1200 Frames or 48 seconds. A value of 0 means that the Shrine is a one-time use.
//	StringName   string `csv:"StringName"`   // String to display as the Shrine's name.
//	StringPhrase string `csv:"StringPhrase"` // String to display as the Shrine's activation phrase when it is used.
//	EffectClass  string `csv:"effectclass"`  // Shrine's archetype involved in calculating region stats.
//	LevelMin     int    `csv:"LevelMin"`     // Earliest area level where the Shrine can spawn. Area levels are determined from levels.txt.
//}

// Note: The "Code ID" and "Parameter Description" data mentioned in the overview have been omitted as they are descriptive and not part of the struct.
// The "Arg0 & Arg1" comment is included as they represent parameter 1 and parameter 2 for the Shrine's function.
