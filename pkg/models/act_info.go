package models

// ActInfoData represents the data from the "actinfo.txt" file.
type ActInfoData struct {
	// Act defines the ID for the Act.
	Act int `csv:"act"`

	// Town uses an area level ("Name field" from levels.txt) to define the Act's town area.
	Town string `csv:"town"`

	// Start uses an area level ("Name field" from levels.txt) to define where the player starts in the Act.
	Start string `csv:"start"`

	// MaxNPCItemLevel controls the maximum item level for items sold by the NPC in the Act.
	MaxNPCItemLevel int `csv:"maxnpcitemlevel"`

	// ClassLevelRangeStart uses an area level ("Name field" from levels.txt) with its MonLvl values as a global Act minimum monster level.
	// For example, this is used to determine chest levels in an Act.
	ClassLevelRangeStart string `csv:"classlevelrangestart"`

	// ClassLevelRangeEnd uses an area level ("Name field" from levels.txt) with its MonLvl values as a global Act maximum monster level.
	// For example, this is used to determine chest levels in an Act.
	ClassLevelRangeEnd string `csv:"classlevelrangeend"`

	// WanderingNPCStart uses an index to determine which wandering monster class to use when populating areas.
	// See wanderingmmon.txt for a list of possible monsters to spawn.
	WanderingNPCStart int `csv:"wanderingnpcstart"`

	// WanderingNPCRange is a modifier that gets added to the "WanderingNPCStart" value to randomly select an index.
	WanderingNPCRange int `csv:"wanderingnpcrange"`

	// CommonActCOF specifies which ".D2" file to use for the common Act COF file.
	// This is used to establish the seed when initializing the Act.
	CommonActCOF string `csv:"commonactcof"`

	// Waypoint1 to Waypoint9 uses an area level ("Name field" from levels.txt) as the designated waypoint selection in the Waypoint UI.
	Waypoint1 string `csv:"waypoint1"`
	Waypoint2 string `csv:"waypoint2"`
	Waypoint3 string `csv:"waypoint3"`
	Waypoint4 string `csv:"waypoint4"`
	Waypoint5 string `csv:"waypoint5"`
	Waypoint6 string `csv:"waypoint6"`
	Waypoint7 string `csv:"waypoint7"`
	Waypoint8 string `csv:"waypoint8"`
	Waypoint9 string `csv:"waypoint9"`

	// WanderingMonsterPopulateChance is the percent chance (from 0 to 100) to spawn a wandering monster.
	// See wanderingmmon.txt for a list of possible monsters to spawn.
	WanderingMonsterPopulateChance int `csv:"wanderingMonsterPopulateChance"`

	// WanderingMonsterRegionTotal is the maximum number of wandering monsters allowed at once.
	WanderingMonsterRegionTotal int `csv:"wanderingMonsterRegionTotal"`

	// WanderingPopulateRandomChance is a secondary percent chance (from 0 to #) to determine whether to attempt populating with monsters.
	// Only fails if the random chance selects 0.
	WanderingPopulateRandomChance int `csv:"wanderingPopulateRandomChance"`
}
