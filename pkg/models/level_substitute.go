package models

// LevelSubstitutionData represents the data from LvlSub.txt.
// This file controls how tiles can be substituted in for other tiles.
// The game divides the level into clusters and iterates through these clusters
// to randomly substitute tiles, creating more visual diversity.
type LevelSubstitutionData struct {
	// Name is a reference field to describe the Level Substitution.
	Name string `csv:"Name"`

	// Type refers to the "SubType" field from the Levels.txt file. This defines a group that multiple substitutions can share.
	Type int `csv:"ItemSuperType"`

	// File specifies the name of the ds1 file to use. The ds1 files contain data for building Level Presets.
	File string `csv:"File"`

	// CheckAll is a Boolean Field. If equals 1, then substitute each tile in the room. If equals 0, then substitute random tiles in the room.
	CheckAll int `csv:"CheckAll"`

	// BordType controls how often substituting tiles can work for border tiles.
	BordType int `csv:"BordType"`

	// GridSize controls the tile size of a cluster for substituting tiles. This evenly affects both the X and Y size values of a room.
	GridSize int `csv:"GridSize"`

	// Dt1Mask functions as a bit field mask with a size of a 32 bit value. This explains to the ds1 file which of the 32 dt1 tile files to use from a Level Type when assembling selecting a tile for substitution.
	Dt1Mask int `csv:"Dt1Mask"`

	// Prob0 (to Prob4) affects the probability that the tile substitution is used. This is a random chance out of 100. Which "Prob#" field that is checked depends on the "SubTheme" value from the Levels.txt file.
	Prob0, Prob1, Prob2, Prob3, Prob4 int `csv:"Prob0,Prob1,Prob2,Prob3,Prob4"`

	// Trials0 (to Trials4) controls the number of times to randomly substitute tiles in a cluster. If this value equals -1, then the game will try to do as many tile substitutions that can be allowed based on the cluster and tile size. This field depends on the "CheckAll" field being equal to 0.
	Trials0, Trials1, Trials2, Trials3, Trials4 int `csv:"Trials0,Trials1,Trials2,Trials3,Trials4"`
}
