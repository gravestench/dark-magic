package models

// LevelPreset represents the data structure for Level Preset information.
type LevelPreset struct {
	// Name is a reference field to define the Level Preset.
	Name string `csv:"Name"`

	// Def defines the unique numeric ID for the Level Preset. This is referenced in other files.
	Def int `csv:"Def"`

	// LevelId refers to the "Id" field from the Levels.txt file. If this value is not equal to 0,
	// then this Level Preset is used to build that entire area level. If this value is equal to 0,
	// then the Level Preset does not define the entire area level and is used as a part of constructing area levels.
	LevelId int `csv:"LevelId"`

	// Populate is a Boolean Field. If equals 1, then units are allowed to spawn in the Level Preset.
	// If equals 0, then units will never spawn in the Level Preset.
	Populate int `csv:"Populate"`

	// Logicals is a Boolean Field. If equals 1, then the Level Preset allows for wall transparency to function.
	// If equals 0, then walls will always appear solid.
	Logicals int `csv:"Logicals"`

	// Outdoors is a Boolean Field. If equals 1, then the Level Preset will be classified as an outdoor area,
	// which can mean that lighting will function differently. If equals 0, then the Level Preset will be classified as an indoor area.
	Outdoors int `csv:"Outdoors"`

	// Animate is a Boolean Field. If equals 1, then the game will animate the tiles in the Level Preset.
	// If equals 0, then ignore this.
	Animate int `csv:"Animate"`

	// KillEdge is a Boolean Field. If equals 1, then the game will remove tiles that border the size of the Level Preset.
	// If equals 0, then ignore this.
	KillEdge int `csv:"KillEdge"`

	// FillBlanks is a Boolean Field. If equals 1, then all blank tiles in the Level Preset will be filled with unwalkable tiles.
	// If equals 0, then ignore this.
	FillBlanks int `csv:"FillBlanks"`

	// SizeX specifies the Length tile size value of the Level Preset, which is used for determining how big to build area levels.
	// This value is equal to 0 for Level Presets that are static.
	SizeX int `csv:"SizeX"`

	// SizeY specifies the Width tile size value of the Level Preset, which is used for determining how big to build area levels.
	// This value is equal to 0 for Level Presets that are static.
	SizeY int `csv:"SizeY"`

	// AutoMap is a Boolean Field. If equals 1, then this Level Preset will be automatically completely revealed on the Automap.
	// If equals 0, then this Level Preset will be hidden on the Automap and will need to be explored.
	AutoMap int `csv:"AutoMap"`

	// Scan is a Boolean Field. If equals 1, then this Level Preset will allow the usage of warping with waypoints
	// (This requires that the Level Preset has a waypoint object). If equals 0, then ignore this.
	Scan int `csv:"Scan"`

	// Pops defines how many Pop tiles are defined in the Level Preset file.
	// These Pop tiles are mainly used for controlling the roof and wall popping when a player enters a building in an area.
	Pops int `csv:"Pops"`

	// PopPad determines the size of the Pop tile area, by using an offset value.
	// This offset value can increase or decrease the size of the Pop tile size if it has a positive or negative value.
	PopPad int `csv:"PopPad"`

	// Files determines the number of different versions to use for the Level Preset.
	// This value acts as a range, which the game will use for randomly choosing one of the "File#" fields to build the Level Preset.
	// This is how the Level Presets have variety when the area level is being built.
	Files int `csv:"Files"`

	// File1 to File6 specify the name of which ds1 file to use.
	// The ds1 files contain data for building Level Presets.
	// If this value equals 0, then this field will be ignored.
	// The number of these defined fields should match the value used in the "Files" field.
	File1 string `csv:"File1"`
	File2 string `csv:"File2"`
	File3 string `csv:"File3"`
	File4 string `csv:"File4"`
	File5 string `csv:"File5"`
	File6 string `csv:"File6"`

	// Dt1Mask functions as a bit field mask with a size of a 32-bit value.
	// This explains to the ds1 file which of the 32 dt1 tile files to use from a Level ItemSuperType when assembling the Level Preset.
	// Each "File#" field from LevelType.txt is assigned a bit value, up to the 32 possible bit values.
	// (For example: File1 = 1, File2=2, File3 = 4, File4=8, File5=16....File32 = 2147483648).
	// To build the "Dt1Mask", you would select which "File#" fields to use from LevelTypes.txt
	// and add their associated bit values together for a total value. This total value is the bitmask value.
	Dt1Mask int `csv:"Dt1Mask"`
}
