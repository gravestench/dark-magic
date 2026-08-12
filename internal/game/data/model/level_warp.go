package models

// LevelWarp is the lossless typed schema for one LvlWarp.txt row. These fields
// retain authored numbers and tokens only; d2legacy decides how they affect
// transitions, interaction, highlighting, and arrival behavior.
type LevelWarp struct {
	Name       string `csv:"Name"`
	Id         int    `csv:"Id"`
	SelectX    int    `csv:"SelectX"`
	SelectY    int    `csv:"SelectY"`
	SelectDX   int    `csv:"SelectDX"`
	SelectDY   int    `csv:"SelectDY"`
	ExitWalkX  int    `csv:"ExitWalkX"`
	ExitWalkY  int    `csv:"ExitWalkY"`
	OffsetX    int    `csv:"OffsetX"`
	OffsetY    int    `csv:"OffsetY"`
	LitVersion int    `csv:"LitVersion"`
	Tiles      int    `csv:"Tiles"`
	NoInteract int    `csv:"NoInteract"`
	Direction  string `csv:"Direction"`
	UniqueId   int    `csv:"UniqueId"`
}
