package models

// Shrine represents the data structure for the shrines.txt file.
type Shrine struct {
	ShrineType string `csv:"Shrine Type"` // The type of shrine
	ShrineName string `csv:"Shrine name"` // The name of the shrine
	Effect     string `csv:"Effect"`      // Description of the shrine's effect
	Code       int    `csv:"Code"`        // Code function used to define the shrine's function
	Arg0       int    `csv:"Arg0"`        // Parameter 1 for the shrine's function
	Arg1       int    `csv:"Arg1"`        // Parameter 2 for the shrine's function
	// Duration of the shrine's effects in frames (1 second = 25 frames)
	DurationFrames int    `csv:"Duration in frames"`
	ResetTimeMin   int    `csv:"reset time in minutes"` // Time in minutes before the shrine is available to use again
	Rarity         int    `csv:"rarity"`                // Rarity of the shrine
	ViewName       string `csv:"view name"`             // View name of the shrine
	NiftyPhrase    string `csv:"niftyphrase"`           // Activation phrase displayed when the shrine is used
	EffectClass    int    `csv:"effectclass"`           // The shrine's archetype involved in calculating region stats
	LevelMin       int    `csv:"LevelMin"`              // Minimum area level where the shrine can spawn
}
