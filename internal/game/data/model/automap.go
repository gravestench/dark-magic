package models

// Overview:
// AutoMap.txt controls how the Automap in-game displays the discovered parts of the area level and stores this progress in character save files.
// The Automap is composed of many different image files depicted as small icons to convey what part of the area level is being displayed.
// This file assigns image files to their related map cells to properly build the Automap as the player explores the area.
// Not all tiles will have image files assigned, and in these cases, those parts of the Automap will remain blank.

type AutoMapEntry struct {
	LevelName     LevelName `csv:"LevelName" lua:"LevelName"`         // Act number and name of the level type
	TileName      TileName  `csv:"TileName" lua:"TileName"`           // Orientation of the tile on the Automap (string codes)
	Style         int       `csv:"Style" lua:"Style"`                 // Group numeric ID for the range of cells with the same style
	StartSequence int       `csv:"StartSequence" lua:"StartSequence"` // Start index value for valid "Cel#" field on the Automap
	EndSequence   int       `csv:"EndSequence" lua:"EndSequence"`     // End index value for valid "Cel#" field on the Automap
	Cel1          int       `csv:"Cel1" lua:"Cel1"`                   // Unique image frames from MaxiMap.dc6 for Automap display
	Cel2          int       `csv:"Cel2" lua:"Cel2"`                   // Unique image frames from MaxiMap.dc6 for Automap display
	Cel3          int       `csv:"Cel3" lua:"Cel3"`                   // Unique image frames from MaxiMap.dc6 for Automap display
	Cel4          int       `csv:"Cel4" lua:"Cel4"`                   // Unique image frames from MaxiMap.dc6 for Automap display
}

// LevelName represents the level types for the Automap.
type LevelName string

// TileName represents the tile orientations on the Automap.
type TileName string

// Level type constants for the Automap.
const (
	LevelTypeNone        LevelName = "Level ItemSuperType None"
	LevelType1Town       LevelName = "Level ItemSuperType 1 Town"
	LevelType1Wilderness LevelName = "Level ItemSuperType 1 Wilderness"
	LevelType1Cave       LevelName = "Level ItemSuperType 1 Cave"
	LevelType1Crypt      LevelName = "Level ItemSuperType 1 Crypt"
	LevelType1Monastery  LevelName = "Level ItemSuperType 1 Monestary"
	LevelType1Courtyard  LevelName = "Level ItemSuperType 1 Courtyard"
	LevelType1Barracks   LevelName = "Level ItemSuperType 1 Barracks"
	LevelType1Jail       LevelName = "Level ItemSuperType 1 Jail"
	LevelType1Cathedral  LevelName = "Level ItemSuperType 1 Cathedral"
	LevelType1Catacombs  LevelName = "Level ItemSuperType 1 Catacombs"
	LevelType1Tristram   LevelName = "Level ItemSuperType 1 Tristram"
	LevelType2Town       LevelName = "Level ItemSuperType 2 Town"
	LevelType2Sewer      LevelName = "Level ItemSuperType 2 Sewer"
	LevelType2Harem      LevelName = "Level ItemSuperType 2 Desert"
	LevelType2Basement   LevelName = "Level ItemSuperType 2 Sewer"
	LevelType2Desert     LevelName = "Level ItemSuperType 2 Desert"
	LevelType2Tomb       LevelName = "Level ItemSuperType 2 Tomb"
	LevelType2Lair       LevelName = "Level ItemSuperType 2 Lair"
	LevelType2Arcane     LevelName = "Level ItemSuperType 2 Arcane"
	LevelType3Town       LevelName = "Level ItemSuperType 3 Town"
	LevelType3Jungle     LevelName = "Level ItemSuperType 3 Jungle"
	LevelType3Kurast     LevelName = "Level ItemSuperType 3 Kurast"
	LevelType3Spider     LevelName = "Level ItemSuperType 3 Spider"
	LevelType3Dungeon    LevelName = "Level ItemSuperType 3 Dungeon"
	LevelType3Sewer      LevelName = "Level ItemSuperType 3 Sewer"
	LevelType4Town       LevelName = "Level ItemSuperType 4 Town"
	LevelType4Mesa       LevelName = "Level ItemSuperType 4 Mesa"
	LevelType4Lava       LevelName = "Level ItemSuperType 4 Hell"
	LevelType5Town       LevelName = "Level ItemSuperType 5 Town"
	LevelType5Siege      LevelName = "Level ItemSuperType 6 Siege"
	LevelType5Barricade  LevelName = "Level ItemSuperType 6 Barricade"
	LevelType5Temple     LevelName = "Level ItemSuperType 6 Temple"
	LevelType5IceCaves   LevelName = "Level ItemSuperType 5 Ice Caves"
	LevelType5Baal       LevelName = "Level ItemSuperType 5 Baal"
	LevelType5Lava       LevelName = "Level ItemSuperType 5 Lava"
)

// Tile name constants for the Automap.
const (
	TileBaseFloor                  TileName = "fl"
	TileBaseLeftWall               TileName = "wl"
	TileBaseRightWall              TileName = "wr"
	TileBaseUpperTopCornerRight    TileName = "wtlr"
	TileBaseUpperTopCornerLeft     TileName = "wtll"
	TileBaseUpperTopCorner         TileName = "wtr"
	TileBaseLowerBottomCornerLeft  TileName = "wbl"
	TileBaseLowerBottomCornerRight TileName = "wbr"
	TileBaseLeftDoor               TileName = "wld"
	TileBaseRightDoor              TileName = "wrd"
	TileBaseLeftExit               TileName = "wle"
	TileBaseRightExit              TileName = "wre"
	TileBaseColumn                 TileName = "co"
	TileBaseShadow                 TileName = "sh"
	TileBaseTree                   TileName = "tr"
	TileBaseRoof                   TileName = "rf"
	TileBaseLeftWallDown           TileName = "ld"
	TileBaseRightWallDown          TileName = "rd"
	TileBaseFullWallDown           TileName = "fd"
	TileBaseFrontWallDown          TileName = "fi"
)
