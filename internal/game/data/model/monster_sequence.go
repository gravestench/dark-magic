package models

// MonsterSequence represents the sequence table used for specifying events during certain animation frames, such as when using skills.
// A sequence is divided into multiple lines in this file to define each frame in the animation and to determine which event to use during a specific frame.
type MonsterSequence struct {
	Sequence string `csv:"sequence"` // Establishes the Monster Sequence index. An entire monster sequence can be composed of multiple sequence lines, which means that each line needs to have matching "sequence" fields and must be in contiguous order.
	Mode     string `csv:"mode"`     // Defines which monster mode animation to use for the sequence
	Frame    int    `csv:"frame"`    // The in-game frame number for the animation. For the first line in the sequence, this value will establish where the starting frame for the animation. These values should be in contiguous order for the sequence.
	Dir      int    `csv:"dir"`      // Defines the numeric animation direction that the frame uses. Most animations have between 8 to 64 maximum directions.
	Event    int    `csv:"event"`    // Defines what type of event will be used when the frame triggers
	// Event Codes Description:
	// 0 - No event, do nothing
	// 1 - Do Melee attack
	// 2 - Do Missile attack
	// 3 - Play a sound
	// 4 - Use a Skill
}
