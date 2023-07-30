package models

// LevelGroup represents the data structure for controlling how the game groups levels together.
// This has currently no gameplay purpose and is mainly used for condensing level names in desecrated (terror) zone messaging.
type LevelGroup struct {
	Name      string `csv:"Name"`      // Defines the unique name pointer for the level group, which is used in other files
	ID        int    `csv:"Id"`        // Defines the unique numeric ID for the level group, which is used in other files
	GroupName string `csv:"GroupName"` // String Field. Used for displaying the name of the level group, such as when all levels in a group have been desecrated.
}
