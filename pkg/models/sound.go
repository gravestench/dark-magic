package models

type SoundEntry struct {
	Sound          string `csv:"Sound"`          // Defines the unique name ID for the sound, which is how other files can reference the sound
	Redirect       string `csv:"Redirect"`       // Points the sound to the index of another sound in the data file for new graphics mode
	Channel        string `csv:"Channel"`        // Declares which channel the sound is initialized in
	FileName       string `csv:"FileName"`       // File path and name of the sound file to play
	IsLocal        int    `csv:"IsLocal"`        // Boolean field for localized sound
	IsMusic        int    `csv:"IsMusic"`        // Boolean field for music sound
	IsAmbientScene int    `csv:"IsAmbientScene"` // Boolean field for ambient scene sound
	IsAmbientEvent int    `csv:"IsAmbientEvent"` // Boolean field for ambient event sound
	IsUI           int    `csv:"IsUI"`           // Boolean field for UI sound
	VolumeMin      int    `csv:"Volume Min"`     // Minimum volume of the sound (0 to 255)
	VolumeMax      int    `csv:"Volume Max"`     // Maximum volume of the sound (0 to 255)
	PitchMin       int    `csv:"Pitch Min"`      // Minimum pitch percentage of the sound
	PitchMax       int    `csv:"Pitch Max"`      // Maximum pitch percentage of the sound
	GroupSize      int    `csv:"Group Size"`     // Sound Group size to group related sounds
	GroupWeight    int    `csv:"Group Weight"`   // Weighted random chance for group sound selection
	Loop           int    `csv:"Loop"`           // Boolean field for sound loop
	FadeIn         int    `csv:"Fade In"`        // Gradual increase in volume when the sound starts playing
	FadeOut        int    `csv:"Fade Out"`       // Gradual decrease in volume when the sound stops playing
	DeferInst      int    `csv:"Defer Inst"`     // Boolean field to stop duplicate sound instance
	StopInst       int    `csv:"Stop Inst"`      // Boolean field to stop previous sound instance
	Duration       int    `csv:"Duration"`       // Length of time to play the sound (0 to ignore)
	Compound       int    `csv:"Compound"`       // Game tick time limit for sound compounding
	Falloff        int    `csv:"Falloff"`        // Defines the range of falloff for hearing the sound, based on distance
	LFEMix         int    `csv:"LFEMix"`         // Percentage of the sound's Low-Frequency Effects channel
	SpatialSpread  int    `csv:"3dSpread"`       // 3D spread angle of the sound (0 to 255)
	Priority       int    `csv:"Priority"`       // Priority of the sound compared to other sounds
	Stream         int    `csv:"Stream"`         // Boolean field for file streaming of the sound
	Is2D           int    `csv:"Is2D"`           // Boolean field for 2D or 3D sound
	Tracking       int    `csv:"Tracking"`       // Boolean field for sound tracking a unit
	Solo           int    `csv:"Solo"`           // Boolean field for reducing volume of other sounds while playing
	MusicVol       int    `csv:"Music Vol"`      // Boolean field for sound's volume affected by music volume
	Block1         int    `csv:"Block 1"`        // Offset time value in sound for sound environment or looping
	Block2         int    `csv:"Block 2"`        // Offset time value in sound for sound environment
	Block3         int    `csv:"Block 3"`        // Offset time value in sound for sound environment
	HDOptOut       int    `csv:"HDOptOut"`       // Boolean field for not playing sound in new graphics mode
	Delay          int    `csv:"Delay"`          // Delay to the starting tick of the sound when it starts playing
}
