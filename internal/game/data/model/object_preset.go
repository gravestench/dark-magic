package models

// ObjectPreset represents an object preset, which controls which objects are preloaded in a preset, based on the Act
// number.
type ObjectPreset struct {
	// Assigns a unique numeric ID to the Object Preset so that it can be properly referenced.
	Index int `csv:"Index"`
	// Defines the Act number used for each Object Preset. Uses values between 1 to 5.
	Act int `csv:"Act"`
	// Uses the "Class" field from objects.txt, which assigns an Object to this Object Preset.
	ObjectClass string `csv:"ObjectClass"`
}
