package models

// ObjectPreset represents an object preset, which controls which objects are preloaded in a preset, based on the Act number.
type ObjectPreset struct {
	Index       int    `csv:"Index"`       // Assigns a unique numeric ID to the Object Preset so that it can be properly referenced.
	Act         int    `csv:"Act"`         // Defines the Act number used for each Object Preset. Uses values between 1 to 5.
	ObjectClass string `csv:"ObjectClass"` // Uses the "Class" field from objects.txt, which assigns an Object to this Object Preset.
}
