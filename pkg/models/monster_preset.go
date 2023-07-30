package models

// MonsterPreset represents the monsters preloaded in a preset based on the Act number.
type MonsterPreset struct {
	Act   int    `csv:"Act"`   // Defines the Act number used for each Monster Preset (values between 1 to 5).
	Place string `csv:"Place"` // Defines the Monster Preset which is used for preloading, such as during level transitions.
}
