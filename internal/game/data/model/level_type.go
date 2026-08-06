package models

// LevelType represents the data from LvlTypes.txt.
// This file controls which files containing tile graphics are used for creating maps.
// It looks at dt1 files, which contain tile images of the environments found in the game.
// Each line in this file defines a Level ItemSuperType and what files it uses.
//
// The order of each Level ItemSuperType defined in this file conveys its ID value,
// which is referenced by the following files: Levels.txt, LvlPrest.txt
// The order of these Level Types should not be changed.
type LevelType struct {
	// Name is a reference field to define the Level ItemSuperType.
	Name string `csv:"Name"`

	// File1 (to File32) specifies the name of which dt1 file to use. The dt1 files contain the images for each area tile found in each Act. If this value equals 0, then this field will be ignored.
	File1  string `csv:"File 1"`
	File2  string `csv:"File 2"`
	File3  string `csv:"File 3"`
	File4  string `csv:"File 4"`
	File5  string `csv:"File 5"`
	File6  string `csv:"File 6"`
	File7  string `csv:"File 7"`
	File8  string `csv:"File 8"`
	File9  string `csv:"File 9"`
	File10 string `csv:"File 10"`
	File11 string `csv:"File 11"`
	File12 string `csv:"File 12"`
	File13 string `csv:"File 13"`
	File14 string `csv:"File 14"`
	File15 string `csv:"File 15"`
	File16 string `csv:"File 16"`
	File17 string `csv:"File 17"`
	File18 string `csv:"File 18"`
	File19 string `csv:"File 19"`
	File20 string `csv:"File 20"`
	File21 string `csv:"File 21"`
	File22 string `csv:"File 22"`
	File23 string `csv:"File 23"`
	File24 string `csv:"File 24"`
	File25 string `csv:"File 25"`
	File26 string `csv:"File 26"`
	File27 string `csv:"File 27"`
	File28 string `csv:"File 28"`
	File29 string `csv:"File 29"`
	File30 string `csv:"File 30"`
	File31 string `csv:"File 31"`
	File32 string `csv:"File 32"`

	// Act defines which Act is related to the Level ItemSuperType. When loading an Act, the game will only use the Level Types associated with that Act number. Uses a decimal number to convey each Act number (Ex: A value of 3 means Act 3).
	Act uint `csv:"Act"`
}
