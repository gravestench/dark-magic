package models

// SoundEnvironment represents the settings for music and ambient sounds
// played in the game's area levels.
type SoundEnvironment struct {
	Handle             string            `csv:"Handle"`                // Reference field to define the name of the Sound Environment
	Song               string            `csv:"Song"`                  // Background music while the player is in an area level (Points to "Sound" value in sounds.txt)
	DayAmbience        string            `csv:"Day Ambience"`          // Ambient sound during daytime in the game (Points to "Sound" value in sounds.txt)
	HDDayAmbience      string            `csv:"HD Day Ambience"`       // Ambient sound during daytime in the game while playing in new graphics mode (Points to "Sound" value in sounds.txt)
	NightAmbience      string            `csv:"Night Ambience"`        // Ambient sound during nighttime in the game (Points to "Sound" value in sounds.txt)
	HDNightAmbience    string            `csv:"HD Night Ambience"`     // Ambient sound during nighttime in the game while playing in new graphics mode (Points to "Sound" value in sounds.txt)
	DayEvent           string            `csv:"Day Event"`             // Random background sound during daytime in the game (Points to "Sound" value in sounds.txt)
	HDDayEvent         string            `csv:"HD Day Event"`          // Random background sound during daytime in the game while playing in new graphics mode (Points to "Sound" value in sounds.txt)
	NightEvent         string            `csv:"Night Event"`           // Random background sound during nighttime in the game (Points to "Sound" value in sounds.txt)
	HDNightEvent       string            `csv:"HD Night Event"`        // Random background sound during nighttime in the game while playing in new graphics mode (Points to "Sound" value in sounds.txt)
	EventDelay         int               `csv:"Event Delay"`           // Baseline number of frames to wait before playing "Day Event" or "Night Event" sound in SD mode
	HDEventDelay       int               `csv:"HD Event Delay"`        // Baseline number of frames to wait before playing "Day Event" or "Night Event" sound in new graphics mode
	Indoors            int               `csv:"Indoors"`               // Boolean field for obstructed sound if current sound is "event_thunder_1"
	Material1          int               `csv:"Material 1"`            // Material of Sound Environment affecting footstep sounds (Code descriptions in data)
	Material2          int               `csv:"Material 2"`            // Material of Sound Environment affecting footstep sounds (Code descriptions in data)
	HDMaterial1        int               `csv:"HD Material 1"`         // Material of Sound Environment affecting footstep sounds in new graphics mode (Code descriptions in data)
	HDMaterial2        int               `csv:"HD Material 2"`         // Material of Sound Environment affecting footstep sounds in new graphics mode (Code descriptions in data)
	SFXEAXEnviron      EnvironmentPreset `csv:"SFX EAX Environ"`       // Environment preset for default sound reverberation settings for special effects sounds (Code descriptions in data)
	SFXEAXRoomVol      int               `csv:"SFX EAX Room Vol"`      // Room effect level at mid frequencies for special effects sounds
	SFXEAXRoomHF       int               `csv:"SFX EAX Room HF"`       // Relative room effect level at high frequencies for special effects sounds
	SFXEAXDecayTime    int               `csv:"SFX EAX Decay Time"`    // Reverberation decay time at mid frequencies for special effects sounds
	SFXEAXDecayHF      int               `csv:"SFX EAX Decay HF"`      // High-frequency to mid-frequency decay time ratio for special effects sounds
	SFXEAXReflect      int               `csv:"SFX EAX Reflect"`       // Early reflections level relative to room effect for special effects sounds
	SFXEAXReflectDelay int               `csv:"SFX EAX Reflect Delay"` // Initial reflection delay time for special effects sounds
	SFXEAXReverb       int               `csv:"SFX EAX Reverb"`        // Late reverberation level relative to room effect for special effects sounds
	SFXEAXRevDelay     int               `csv:"SFX EAX Rev Delay"`     // Late reverberation delay time relative to initial reflection for special effects sounds
	VOXEAXEnviron      int               `csv:"VOX EAX Environ"`       // Environment preset for default sound reverberation settings for Voice sounds (Code descriptions in data)
	VOXEAXRoomVol      int               `csv:"VOX EAX Room Vol"`      // Room effect level at mid frequencies for Voice sounds
	VOXEAXRoomHF       int               `csv:"VOX EAX Room HF"`       // Relative room effect level at high frequencies for Voice sounds
	VOXEAXDecayTime    int               `csv:"VOX EAX Decay Time"`    // Reverberation decay time at mid frequencies for Voice sounds
	VOXEAXDecayHF      int               `csv:"VOX EAX Decay HF"`      // High-frequency to mid-frequency decay time ratio for Voice sounds
	VOXEAXReflect      int               `csv:"VOX EAX Reflect"`       // Early reflections level relative to room effect for Voice sounds
	VOXEAXReflectDelay int               `csv:"VOX EAX Reflect Delay"` // Initial reflection delay time for Voice sounds
	VOXEAXReverb       int               `csv:"VOX EAX Reverb"`        // Late reverberation level relative to room effect for Voice sounds
	VOXEAXRevDelay     int               `csv:"VOX EAX Rev Delay"`     // Late reverberation delay time relative to initial reflection for Voice sounds
	InheritEnvironment int               `csv:"InheritEnvironment"`    // Boolean field for inheriting values from the existing environment and overwriting other values
}

// EnvironmentPreset represents the environment presets for default sound reverberation settings.
type EnvironmentPreset int

// List of environment presets with their corresponding codes and descriptions.
const (
	Generic         EnvironmentPreset = iota // 0 - Generic
	PaddedCell                               // 1 - Padded Cell
	Room                                     // 2 - Room
	Bathroom                                 // 3 - Bathroom
	Livingroom                               // 4 - Livingroom
	StoneRoom                                // 5 - Stone Room
	Auditorium                               // 6 - Auditorium
	ConcertHall                              // 7 - Concert Hall
	Cave                                     // 8 - Cave
	Arena                                    // 9 - Arena
	Hanger                                   // 10 - Hanger
	CarpetedHallway                          // 11 - Carpeted Hallway
	Hallway                                  // 12 - Hallway
	StoneCorridor                            // 13 - Stone Corridor
	Alley                                    // 14 - Alley
	Forest                                   // 15 - Forest
	City                                     // 16 - City
	Mountains                                // 17 - Mountains
	Quarry                                   // 18 - Quarry
	Plain                                    // 19 - Plain
	ParkingLot                               // 20 - Parking Lot
	SewerPipe                                // 21 - Sewer Pipe
	Underwater                               // 22 - Underwater
	Drugged                                  // 23 - Drugged
	Dizzy                                    // 24 - Dizzy
	Psychotic                                // 25 - Psychotic
	ProgrammerTest                           // 26 - Programmer Test (A long distant echo)
)

func (ep EnvironmentPreset) String() string {
	descriptions := []string{
		"Generic",
		"Padded Cell",
		"Room",
		"Bathroom",
		"Livingroom",
		"Stone Room",
		"Auditorium",
		"Concert Hall",
		"Cave",
		"Arena",
		"Hanger",
		"Carpeted Hallway",
		"Hallway",
		"Stone Corridor",
		"Alley",
		"Forest",
		"City",
		"Mountains",
		"Quarry",
		"Plain",
		"Parking Lot",
		"Sewer Pipe",
		"Underwater",
		"Drugged",
		"Dizzy",
		"Psychotic",
		"Programmer Test (A long distant echo)",
	}

	if ep >= 0 && int(ep) < len(descriptions) {
		return descriptions[ep]
	}

	return "Unknown"
}
