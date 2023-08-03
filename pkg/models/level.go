package models

// LevelData represents the data fields in the "Levels.txt" file.
type LevelData struct {
	// Name defines the unique name pointer for the area level, which is used in other files
	Name string `csv:"Name"`

	// Id defines the unique numeric ID for the area level, which is used in other files
	Id int `csv:"Id"`

	// Pal defines which palette file to use for the area level.
	// This uses index values from 0 to 4 to convey Act 1 to Act 5.
	Pal int `csv:"Pal"`

	// Act defines the Act number that the area level is a part of.
	// This uses index values from 0 to 4 to convey Act 1 to Act 5.
	Act int `csv:"Act"`

	// QuestFlag controls what quest record that the player needs to have completed before being allowed to enter this area level, while playing in Classic Mode.
	// Each quest can have multiple quest records, and this field is looking for a specific quest record from a quest.
	QuestFlag int `csv:"QuestFlag"`

	// QuestFlagEx controls what quest record that the player needs to have completed before being allowed to enter this area level, while playing in Expansion Mode.
	// Each quest can have multiple quest records, and this field is looking for a specific quest record from a quest.
	QuestFlagEx int `csv:"QuestFlagEx"`

	// Layer defines a unique numeric ID that is used to identify which Automap data belongs to which area level when saving and loading data from the character save.
	Layer int `csv:"Layer"`

	// SizeX & SizeX(N) & SizeX(H) specify the Length tile size values of an entire area level,
	// which are used for determining how to build the level, for Normal, Nightmare, and Hell Difficulty, respectively.
	SizeX, SizeXN, SizeXH int `csv:"SizeX,SizeX(N),SizeX(H)"`

	// SizeY & SizeY(N) & SizeY(H) specify the Width tile size values of an entire area level,
	// which are used for determining how to build the level, for Normal, Nightmare, and Hell Difficulty, respectively.
	SizeY, SizeYN, SizeYH int `csv:"SizeY,SizeY(N),SizeY(H)"`

	// OffsetX & OffsetY specify the location offset coordinates (measured in tile size) for the origin point of the area level in the world.
	OffsetX, OffsetY int `csv:"OffsetX,OffsetY"`

	// Depend assigns another level to be this area level’s depended level,
	// which controls this area level’s position and how it starts building its tiles.
	// Uses the level "Id" field. If this equals 0, then ignore this.
	Depend int `csv:"Depend"`

	// Teleport controls the functionality of the Sorceress Teleport skill and the Assassin Dragon Flight skill on the area level.
	Teleport int `csv:"Teleport"`

	// Rain is a Boolean Field. If equals 1, then allow rain to play its effects on the area level.
	// If the level is part of Act 5, then it will snow on the area level, instead of rain.
	// If equals 0, then it will never rain on the area level.
	Rain int `csv:"Rain"`

	// Mud is a Boolean Field. If equals 1, then random bubbles will animate on the tiles that are flagged as water tiles.
	// If equals 0, then ignore this.
	Mud int `csv:"Mud"`

	// NoPer is a Boolean Field. If equals 1, then allow the use of the display option of Perspective Mode while the player is in the level.
	// If equals 0, then disable the option of Perspective Mode and force the player to use Orthographic Mode while the player is in the level.
	NoPer int `csv:"NoPer"`

	// LOSDraw is a Boolean field. If equals 1, then the level will check the player’s line of sight before drawing monsters.
	// If equals 0, then ignore this.
	LOSDraw int `csv:"LOSDraw"`

	// FloorFilter is a Boolean field. If equals 1 and if the floor’s layer in the area level equals 1,
	// then draw the floor tiles with a linear texture sampler. If equals 0, then draw the floor tiles with a nearest texture sampler.
	FloorFilter int `csv:"FloorFilter"`

	// BlankScreen is a Boolean field. If equals 1, then draw the area level screen.
	// If equals 0, then do not draw the area level screen, meaning that the level will be a blank screen.
	BlankScreen int `csv:"BlankScreen"`

	// DrawEdges is a Boolean field. If equals 1, then draw the areas in levels that are not covered by floor tiles.
	// If equals 0, then ignore this.
	DrawEdges int `csv:"DrawEdges"`

	// DrlgType determines the type of Dynamic Random Level Generation used for building and handling different elements of the area level.
	// Uses a numeric code to handle which type of DRLG is used.
	DrlgType int `csv:"DrlgType"`

	// LevelType defines the Level Type used for this area level.
	// Uses the Level Type’s ID, which is determined by what order it is defined in the LvlType.txt file.
	LevelType int `csv:"LevelType"`

	// SubType controls the group of tile substitutions for the area level (see LvlSub.txt).
	// There are defined subtypes to choose from.
	SubType int `csv:"SubType"`

	// SubTheme controls which theme number to use in a Level Substitution (See LvlSub.txt).
	// The allowed values are 0 to 4, which convey which "Prob#", "Trials#", and "Max#" field to use from the LvlSub.txt file.
	// If this equals -1, then there is no sub-theme for the area level.
	SubTheme int `csv:"SubTheme"`

	// SubWaypoint controls the level substitutions for adding waypoints in the area level (see LvlSub.txt).
	// This uses a defined subtype to choose from (See "SubType"). This will depend on the room having a waypoint tile.
	SubWaypoint int `csv:"SubWaypoint"`

	// SubShrine controls the level substitutions for adding shrines in the area level (see LvlSub.txt).
	// This uses a defined subtype to choose from (See "SubType"). This will depend on the room allowing for a shrine to spawn.
	SubShrine int `csv:"SubShrine"`

	// Vis0 (to Vis7) defines the visibility of other area levels involved with this area level,
	// allowing for travel functionalities between levels.
	// This uses the "Id" field of another defined area level to link with this area level.
	// If this equals 0, then no area level is specified.
	Vis0, Vis1, Vis2, Vis3, Vis4, Vis5, Vis6, Vis7 int `csv:"Vis0,Vis1,Vis2,Vis3,Vis4,Vis5,Vis6,Vis7"`

	// Warp0 (to Warp7) uses the "Id" field from LevelWarp.txt, which defines which Level Warp to use when exiting the area level.
	// This is connected with the definition of the related "Vis#" field.
	// If this equals -1, then no Level Warp is specified, which should also mean that the related "Vis#" field is not defined.
	Warp0, Warp1, Warp2, Warp3, Warp4, Warp5, Warp6, Warp7 int `csv:"Warp0,Warp1,Warp2,Warp3,Warp4,Warp5,Warp6,Warp7"`

	// Intensity controls the intensity value of the area level’s ambient colors.
	// This affects the brightness of the room’s RGB colors.
	// Uses a value between 0 and 128.
	// If all these related fields equal 0, then the game ignores setting the area level’s ambient colors.
	Intensity int `csv:"Intensity"`

	// Red controls the red value of the area level’s ambient colors.
	// Uses a value between 0 and 255.
	Red int `csv:"Red"`

	// Green controls the green value of the area level’s ambient colors.
	// Uses a value between 0 and 255.
	Green int `csv:"Green"`

	// Blue controls the blue value of the area level’s ambient colors.
	// Uses a value between 0 and 255.
	Blue int `csv:"Blue"`

	// Portal is a Boolean Field. If equals 1, then this area level will be flagged as a portal level,
	// which is saved in the player’s information and can be used for keeping track of the player’s portal functionalities.
	// If equals 0, then ignore this.
	Portal int `csv:"Portal"`

	// Position is a Boolean Field. If equals 1, then enable special casing for positioning the player on the area level.
	// This can mean that the player could spawn in a different location on the area level,
	// depending on the level room’s position type.
	// An example can be when the player spawns in a town when loading the game, or using a waypoint, or using a town portal.
	// If equals 0, then ignore this.
	Position int `csv:"Position"`

	// SaveMonsters is a Boolean Field. If equals 1, then the game will save the monsters in the area level,
	// such as when all players leave the area level.
	// If equals 0, then monsters will not be saved and will be removed.
	// This is usually disabled for areas where monsters do not spawn.
	SaveMonsters int `csv:"SaveMonsters"`

	// Quest controls what quest record is attached to monsters that spawn in this area level.
	// This is used for specific quests handling lists of monsters in the area level.
	Quest int `csv:"Quest"`

	// WarpDist defines the minimum pixel distance from a Level Warp that a monster is allowed to spawn near.
	// Tile distance values are converted to game pixel distance values by multiplying the tile distance value by 160 / 32, where 160 is the width of pixels of a tile.
	WarpDist int `csv:"WarpDist"`

	// MonLvl & MonLvl(N) & MonLvl(H) control the overall monster level for the area level for Normal, Nightmare, and Hell Difficulty, respectively.
	// This is for Classic mode only. This can affect the highest item level allowed to drop in this area level.
	MonLvl, MonLvlN, MonLvlH int `csv:"MonLvl,MonLvl(N),MonLvl(H)"`

	// MonLvlEx & MonLvlEx(N) & MonLvlEx(H) control the overall monster level for the area level for Normal, Nightmare, and Hell Difficulty, respectively.
	// This is for Expansion mode only. This can affect the highest item level allowed to drop in this area level.
	MonLvlEx, MonLvlExN, MonLvlExH int `csv:"MonLvlEx,MonLvlEx(N),MonLvlEx(H)"`

	// MonDen & MonDen(N) & MonDen(H) control the monster density on the area level for Normal, Nightmare, and Hell Difficulty, respectively.
	// This is a random value out of 100000, which will determine whether to spawn or not spawn a monster pack in the room of the area level.
	// If this value equals 0, then no random monsters will populate on the area level.
	MonDen, MonDenN, MonDenH int `csv:"MonDen,MonDen(N),MonDen(H)"`

	// MonUMin & MonUMin(N) & MonUMin(H) define the minimum number of Unique Monsters that can spawn in the area level for Normal, Nightmare, and Hell Difficulty, respectively.
	// This field depends on the related "MonDen" field being defined.
	MonUMin, MonUMinN, MonUMinH int `csv:"MonUMin,MonUMin(N),MonUMin(H)"`

	// MonUMax & MonUMax(N) & MonUMax(H) define the maximum number of Unique Monsters that can spawn in the area level for Normal, Nightmare, and Hell Difficulty, respectively.
	// This field depends on the related "MonDen" field being defined.
	// Each room in the area level will attempt to spawn a Unique Monster with a 5/100 random chance,
	// and this field’s value will cap the number of successful attempts for the entire area level.
	MonUMax, MonUMaxN, MonUMaxH int `csv:"MonUMax,MonUMax(N),MonUMax(H)"`

	// MonWndr is a Boolean Field. If equals 1, then allow Wandering Monsters to spawn on this area level (see wanderingmon.txt).
	// This field depends on the related "MonDen" field being defined. If equals 0, then ignore this.
	MonWndr int `csv:"MonWndr"`

	// MonSpcWalk defines a distance value, used to handle monster pathing AI when the level has certain pathing blockers,
	// such as jail bars or rivers.
	// In these cases, monsters will walk randomly until a player is located within this distance value,
	// or when the monsters find a possible path to target the player.
	// If this equals 0, then ignore this field.
	MonSpcWalk int `csv:"MonSpcWalk"`

	// NumMon controls the number of different monsters randomly spawn in the area level.
	// The maximum value is 13. This controls the number of random selections from the 25 related "mon#" and "umon#" fields or "nmmon#" fields, depending on the game difficulty.
	NumMon int `csv:"NumMon"`

	// mon1 (to mon25) define which monsters can spawn on the area level for Normal Difficulty.
	// Uses the monster "Id" field from the monstats.txt file.
	Mon1, Mon2, Mon3, Mon4, Mon5, Mon6, Mon7, Mon8, Mon9, Mon10, Mon11, Mon12, Mon13, Mon14, Mon15, Mon16, Mon17, Mon18, Mon19, Mon20, Mon21, Mon22, Mon23, Mon24, Mon25 int `csv:"mon1,mon2,mon3,mon4,mon5,mon6,mon7,mon8,mon9,mon10,mon11,mon12,mon13,mon14,mon15,mon16,mon17,mon18,mon19,mon20,mon21,mon22,mon23,mon24,mon25"`

	// rangedspawn is a Boolean Field. If equals 1, then for the first monster, try to pick a ranged type.
	// If equals 0, then ignore this.
	RangedSpawn int `csv:"rangedspawn"`

	// nmon1 (to nmon25) define which monsters can spawn on the area level for Nightmare Difficulty and Hell Difficulty.
	// Uses the monster "Id" field from the monstats.txt file.
	Nmon1, Nmon2, Nmon3, Nmon4, Nmon5, Nmon6, Nmon7, Nmon8, Nmon9, Nmon10, Nmon11, Nmon12, Nmon13, Nmon14, Nmon15, Nmon16, Nmon17, Nmon18, Nmon19, Nmon20, Nmon21, Nmon22, Nmon23, Nmon24, Nmon25 int `csv:"nmon1,nmon2,nmon3,nmon4,nmon5,nmon6,nmon7,nmon8,nmon9,nmon10,nmon11,nmon12,nmon13,nmon14,nmon15,nmon16,nmon17,nmon18,nmon19,nmon20,nmon21,nmon22,nmon23,nmon24,nmon25"`

	// umon1 (to umon25) define which monsters can spawn as Unique monsters on this area level for Normal Difficulty.
	// Uses the monster "Id" field from the monstats.txt file.
	Umon1, Umon2, Umon3, Umon4, Umon5, Umon6, Umon7, Umon8, Umon9, Umon10, Umon11, Umon12, Umon13, Umon14, Umon15, Umon16, Umon17, Umon18, Umon19, Umon20, Umon21, Umon22, Umon23, Umon24, Umon25 int `csv:"umon1,umon2,umon3,umon4,umon5,umon6,umon7,umon8,umon9,umon10,umon11,umon12,umon13,umon14,umon15,umon16,umon17,umon18,umon19,umon20,umon21,umon22,umon23,umon24,umon25"`

	// cmon1 (to cmon4) define which Critter monsters can spawn on the area level.
	// Uses the monster "Id" field from the monstats.txt file.
	// Critter monsters are determined by the "critter" field from the monstats2.txt file.
	Cmon1, Cmon2, Cmon3, Cmon4 int `csv:"cmon1,cmon2,cmon3,cmon4"`

	// cpct1 (to cpct4) control the percent chance (out of 100) to spawn a Critter monster on the area level.
	Cpct1, Cpct2, Cpct3, Cpct4 int `csv:"cpct1,cpct2,cpct3,cpct4"`

	// camt1 (to camt4) control the amount of Critter monsters to spawn on the area level after they succeeded their random spawn chance from the related "cpct#" field.
	Camt1, Camt2, Camt3, Camt4 int `csv:"camt1,camt2,camt3,camt4"`

	// Themes control the type of theme when building a level room.
	// This value is a summation of possible values to build a bit mask for determining which themes to use when building a level room.
	Themes int `csv:"Themes"`

	// SoundEnv uses the "Index" field from SoundEnviron.txt, which controls what music is played while the player is in the area level.
	SoundEnv int `csv:"SoundEnv"`

	// Waypoint defines the unique numeric ID for the Waypoint in the area level.
	// If this value is greater than or equal to 255, then ignore this field.
	Waypoint int `csv:"Waypoint"`

	// LevelName is a String Field. Used for displaying the name of the area level on the Automap or when hovering the mouse cursor over the waypoint button.
	// The value is the string identifier defined in the String.tbl file.
	LevelName string `csv:"LevelName"`

	// ObjectID defines a numeric ID to specify which object group belongs to the area level (See LevelWarp.txt).
	// If this value is greater than or equal to 255, then ignore this field.
	ObjGrp0, ObjGrp1, ObjGrp2, ObjGrp3, ObjGrp4, ObjGrp5, ObjGrp6, ObjGrp7 int `csv:"ObjGrp0,ObjGrp1,ObjGrp2,ObjGrp3,ObjGrp4,ObjGrp5,ObjGrp6,ObjGrp7"`
	ObjPrb0, ObjPrb1, ObjPrb2, ObjPrb3, ObjPrb4, ObjPrb5, ObjPrb6, ObjPrb7 int `csv:"ObjPrb0,ObjPrb1,ObjPrb2,ObjPrb3,ObjPrb4,ObjPrb5,ObjPrb6,ObjPrb7"`

	// Defines what group this level belongs to. Used for condensing level names
	// in desecrated (terror) zones messaging. See LevelGroups.txt.
	LevelGroup string
}
