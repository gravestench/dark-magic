package models

// LevelMazeData represents the data from the "LvlMaze.txt" file.
type LevelMazeData struct {
	// Name is a reference field to describe the area level.
	Name string `csv:"Name"`

	// Level refers to the "Id" field from the Levels.txt file.
	Level int `csv:"Level"`

	// Rooms, RoomsN, RoomsH control the total number of rooms that a Level Maze will generate for Normal, Nightmare, and Hell Difficulty, respectively.
	Rooms, RoomsN, RoomsH int `csv:"Rooms,Rooms(N),Rooms(H)"`

	// SizeX and SizeY control the length and width sizes of each room (ds1 map files) that is added to the Level Maze. This is measured in tile sizes.
	SizeX, SizeY int `csv:"SizeX,SizeY"`

	// Merge affects the probability that a room gets an adjacent room linked next to it. This is a random chance out of 1000.
	Merge int `csv:"Merge"`
}
