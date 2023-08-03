package models

import (
	lua "github.com/yuin/gopher-lua"
)

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

// ExportToLua exports the SoundEntry object to a Lua table.
func (s SoundEntry) ExportToLua(state *lua.LState) *lua.LTable {
	table := state.NewTable()

	// Set the fields of the table using the struct's values.
	table.RawSetString("Sound", lua.LString(s.Sound))
	table.RawSetString("Redirect", lua.LString(s.Redirect))
	table.RawSetString("Channel", lua.LString(s.Channel))
	table.RawSetString("FileName", lua.LString(s.FileName))
	table.RawSetString("IsLocal", lua.LNumber(s.IsLocal))
	table.RawSetString("IsMusic", lua.LNumber(s.IsMusic))
	table.RawSetString("IsAmbientScene", lua.LNumber(s.IsAmbientScene))
	table.RawSetString("IsAmbientEvent", lua.LNumber(s.IsAmbientEvent))
	table.RawSetString("IsUI", lua.LNumber(s.IsUI))
	table.RawSetString("VolumeMin", lua.LNumber(s.VolumeMin))
	table.RawSetString("VolumeMax", lua.LNumber(s.VolumeMax))
	table.RawSetString("PitchMin", lua.LNumber(s.PitchMin))
	table.RawSetString("PitchMax", lua.LNumber(s.PitchMax))
	table.RawSetString("GroupSize", lua.LNumber(s.GroupSize))
	table.RawSetString("GroupWeight", lua.LNumber(s.GroupWeight))
	table.RawSetString("Loop", lua.LNumber(s.Loop))
	table.RawSetString("FadeIn", lua.LNumber(s.FadeIn))
	table.RawSetString("FadeOut", lua.LNumber(s.FadeOut))
	table.RawSetString("DeferInst", lua.LNumber(s.DeferInst))
	table.RawSetString("StopInst", lua.LNumber(s.StopInst))
	table.RawSetString("Duration", lua.LNumber(s.Duration))
	table.RawSetString("Compound", lua.LNumber(s.Compound))
	table.RawSetString("Falloff", lua.LNumber(s.Falloff))
	table.RawSetString("LFEMix", lua.LNumber(s.LFEMix))
	table.RawSetString("SpatialSpread", lua.LNumber(s.SpatialSpread))
	table.RawSetString("Priority", lua.LNumber(s.Priority))
	table.RawSetString("Stream", lua.LNumber(s.Stream))
	table.RawSetString("Is2D", lua.LNumber(s.Is2D))
	table.RawSetString("Tracking", lua.LNumber(s.Tracking))
	table.RawSetString("Solo", lua.LNumber(s.Solo))
	table.RawSetString("MusicVol", lua.LNumber(s.MusicVol))
	table.RawSetString("Block1", lua.LNumber(s.Block1))
	table.RawSetString("Block2", lua.LNumber(s.Block2))
	table.RawSetString("Block3", lua.LNumber(s.Block3))
	table.RawSetString("HDOptOut", lua.LNumber(s.HDOptOut))
	table.RawSetString("Delay", lua.LNumber(s.Delay))

	return table
}
