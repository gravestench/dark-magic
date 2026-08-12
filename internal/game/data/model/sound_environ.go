package models

// SoundEnvironment represents the settings for music and ambient sounds
// played in the game's area levels.
type SoundEnvironment struct {
	Handle             string `csv:"Handle"`                // Reference field to define the name of the Sound Environment
	Song               string `csv:"Song"`                  // Background music while the player is in an area level (Points to "Sound" value in sounds.txt)
	DayAmbience        string `csv:"Day Ambience"`          // Ambient sound during daytime in the game (Points to "Sound" value in sounds.txt)
	HDDayAmbience      string `csv:"HD Day Ambience"`       // Ambient sound during daytime in the game while playing in new graphics mode (Points to "Sound" value in sounds.txt)
	NightAmbience      string `csv:"Night Ambience"`        // Ambient sound during nighttime in the game (Points to "Sound" value in sounds.txt)
	HDNightAmbience    string `csv:"HD Night Ambience"`     // Ambient sound during nighttime in the game while playing in new graphics mode (Points to "Sound" value in sounds.txt)
	DayEvent           string `csv:"Day Event"`             // Random background sound during daytime in the game (Points to "Sound" value in sounds.txt)
	HDDayEvent         string `csv:"HD Day Event"`          // Random background sound during daytime in the game while playing in new graphics mode (Points to "Sound" value in sounds.txt)
	NightEvent         string `csv:"Night Event"`           // Random background sound during nighttime in the game (Points to "Sound" value in sounds.txt)
	HDNightEvent       string `csv:"HD Night Event"`        // Random background sound during nighttime in the game while playing in new graphics mode (Points to "Sound" value in sounds.txt)
	EventDelay         int    `csv:"Event Delay"`           // Baseline number of frames to wait before playing "Day Event" or "Night Event" sound in SD mode
	HDEventDelay       int    `csv:"HD Event Delay"`        // Baseline number of frames to wait before playing "Day Event" or "Night Event" sound in new graphics mode
	Indoors            int    `csv:"Indoors"`               // Boolean field for obstructed sound if current sound is "event_thunder_1"
	Material1          int    `csv:"Material 1"`            // Material of Sound Environment affecting footstep sounds (Code descriptions in data)
	Material2          int    `csv:"Material 2"`            // Material of Sound Environment affecting footstep sounds (Code descriptions in data)
	HDMaterial1        int    `csv:"HD Material 1"`         // Material of Sound Environment affecting footstep sounds in new graphics mode (Code descriptions in data)
	HDMaterial2        int    `csv:"HD Material 2"`         // Material of Sound Environment affecting footstep sounds in new graphics mode (Code descriptions in data)
	SFXEAXEnviron      int    `csv:"SFX EAX Environ"`       // Raw authored environment code; d2legacy owns its meaning.
	SFXEAXRoomVol      int    `csv:"SFX EAX Room Vol"`      // Room effect level at mid frequencies for special effects sounds
	SFXEAXRoomHF       int    `csv:"SFX EAX Room HF"`       // Relative room effect level at high frequencies for special effects sounds
	SFXEAXDecayTime    int    `csv:"SFX EAX Decay Time"`    // Reverberation decay time at mid frequencies for special effects sounds
	SFXEAXDecayHF      int    `csv:"SFX EAX Decay HF"`      // High-frequency to mid-frequency decay time ratio for special effects sounds
	SFXEAXReflect      int    `csv:"SFX EAX Reflect"`       // Early reflections level relative to room effect for special effects sounds
	SFXEAXReflectDelay int    `csv:"SFX EAX Reflect Delay"` // Initial reflection delay time for special effects sounds
	SFXEAXReverb       int    `csv:"SFX EAX Reverb"`        // Late reverberation level relative to room effect for special effects sounds
	SFXEAXRevDelay     int    `csv:"SFX EAX Rev Delay"`     // Late reverberation delay time relative to initial reflection for special effects sounds
	VOXEAXEnviron      int    `csv:"VOX EAX Environ"`       // Environment preset for default sound reverberation settings for Voice sounds (Code descriptions in data)
	VOXEAXRoomVol      int    `csv:"VOX EAX Room Vol"`      // Room effect level at mid frequencies for Voice sounds
	VOXEAXRoomHF       int    `csv:"VOX EAX Room HF"`       // Relative room effect level at high frequencies for Voice sounds
	VOXEAXDecayTime    int    `csv:"VOX EAX Decay Time"`    // Reverberation decay time at mid frequencies for Voice sounds
	VOXEAXDecayHF      int    `csv:"VOX EAX Decay HF"`      // High-frequency to mid-frequency decay time ratio for Voice sounds
	VOXEAXReflect      int    `csv:"VOX EAX Reflect"`       // Early reflections level relative to room effect for Voice sounds
	VOXEAXReflectDelay int    `csv:"VOX EAX Reflect Delay"` // Initial reflection delay time for Voice sounds
	VOXEAXReverb       int    `csv:"VOX EAX Reverb"`        // Late reverberation level relative to room effect for Voice sounds
	VOXEAXRevDelay     int    `csv:"VOX EAX Rev Delay"`     // Late reverberation delay time relative to initial reflection for Voice sounds
	InheritEnvironment int    `csv:"InheritEnvironment"`    // Boolean field for inheriting values from the existing environment and overwriting other values
}
