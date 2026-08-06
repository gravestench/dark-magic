package models

type WarpCode int

// LevelWarp represents the data from LvlWarp.txt.
// This file controls how the player is moved to different area levels,
// such as entrances and exits between different areas. This player transportation
// of between levels is defined as a Level Warp. Level Warps function as special
// tiles that are added to the area for controlling the location for where to transport the player.
//
// This file is used by the Levels.txt file.
// LevelWarp represents the data from the "LvlWarp.txt" file.
type LevelWarp struct {
	// Name is the reference field to define the Level Warp.
	Name string `csv:"Name"`

	// Id defines the numeric ID for the type of Level Warp.
	Id string `csv:"Id"`

	// SelectX defines the horizontal offset of the starting left corner position of the Level Warp area.
	SelectX int `csv:"SelectX"`

	// SelectY defines the vertical offset of the starting left corner position of the Level Warp area.
	SelectY int `csv:"SelectY"`

	// SelectDX defines the horizontal offset from the starting position of the Level Warp area.
	SelectDX int `csv:"SelectDX"`

	// SelectDY defines the vertical offset from the starting position of the Level Warp area.
	SelectDY int `csv:"SelectDY"`

	// ExitWalkX defines the horizontal position of the destination location where the player will walk to after exiting to this Level Warp.
	ExitWalkX int `csv:"ExitWalkX"`

	// ExitWalkY defines the vertical position of the destination location where the player will walk to after exiting to this Level Warp.
	ExitWalkY int `csv:"ExitWalkY"`

	// OffsetX defines the horizontal position of the sub-tile for the Level Warp, where the player will appear when exiting to this area level.
	OffsetX int `csv:"OffsetX"`

	// OffsetY defines the vertical position of the sub-tile for the Level Warp, where the player will appear when exiting to this area level.
	OffsetY int `csv:"OffsetY"`

	// LitVersion is a boolean field. If it equals 1, then Level Warp tiles will change their appearance when highlighted. If it equals 0, then the Level Warp tiles will not change appearance when highlighted.
	LitVersion int `csv:"LitVersion"`

	// Tiles defines an index offset to determine which tile to use in the tile set for the highlighted version of the Level Warp.
	Tiles int `csv:"Tiles"`

	// NoInteract is a boolean field. If it equals 1, then the Level Warp cannot be directly interacted by the player. If it equals 0, then the player can interact with the Level Warp.
	NoInteract int `csv:"NoInteract"`

	// Direction defines the orientation of the Level Warp.
	Direction string `csv:"Direction"`

	// UniqueId defines the unique numeric ID for the Level Warp. Each Level Warp should have a unique ID so that the game can handle loading that specific Level Warp’s related files.
	UniqueId int `csv:"UniqueId"`
}
