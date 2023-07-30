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
	File1, File2, File3, File4, File5, File6, File7, File8, File9, File10, File11, File12, File13, File14, File15, File16, File17, File18, File19, File20, File21, File22, File23, File24, File25, File26, File27, File28, File29, File30, File31, File32 string `csv:"File 1,File 2,File 3,File 4,File 5,File 6,File 7,File 8,File 9,File 10,File 11,File 12,File 13,File 14,File 15,File 16,File 17,File 18,File 19,File 20,File 21,File 22,File 23,File 24,File 25,File 26,File 27,File 28,File 29,File 30,File 31,File 32"`

	// Act defines which Act is related to the Level ItemSuperType. When loading an Act, the game will only use the Level Types associated with that Act number. Uses a decimal number to convey each Act number (Ex: A value of 3 means Act 3).
	Act int `csv:"Act"`
}
