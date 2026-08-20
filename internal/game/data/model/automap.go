package models

// Overview:
// AutoMap.txt controls how the Automap in-game displays the discovered parts of the area level and stores this progress
// in character save files.
// The Automap is composed of many different image files depicted as small icons to convey what part of the area level
// is being displayed.
// This file assigns image files to their related map cells to properly build the Automap as the player explores the
// area.
// Not all tiles will have image files assigned, and in these cases, those parts of the Automap will remain blank.

type AutoMapEntry struct {
	LevelName LevelName `csv:"LevelName" lua:"LevelName"` // Act number and name of the level type
	// Orientation of the tile on the Automap (string codes)
	TileName TileName `csv:"TileName" lua:"TileName"`
	// Group numeric ID for the range of cells with the same style
	Style int `csv:"Style" lua:"Style"`
	// Start index value for valid "Cel#" field on the Automap
	StartSequence int `csv:"StartSequence" lua:"StartSequence"`
	// End index value for valid "Cel#" field on the Automap
	EndSequence int `csv:"EndSequence" lua:"EndSequence"`
	// Unique image frames from MaxiMap.dc6 for Automap display
	Cel1 int `csv:"Cel1" lua:"Cel1"`
	// Unique image frames from MaxiMap.dc6 for Automap display
	Cel2 int `csv:"Cel2" lua:"Cel2"`
	// Unique image frames from MaxiMap.dc6 for Automap display
	Cel3 int `csv:"Cel3" lua:"Cel3"`
	// Unique image frames from MaxiMap.dc6 for Automap display
	Cel4 int `csv:"Cel4" lua:"Cel4"`
}

// LevelName represents the level types for the Automap.
type LevelName string

// TileName represents the tile orientations on the Automap.
type TileName string
