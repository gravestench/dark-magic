package models

// MonsterStats represents the main functionalities and statistics for a monster in the game.
type MonsterStats struct {
	Id string `csv:"Id"` // Controls the unique name ID to define the monster.
	// Points to the "Id" of another monster to define the monster's base type.
	BaseId string `csv:"BaseId"`
	// Points to the "Id" of another monster to signify the next monster in the group of this monster's type.
	NextInClass string `csv:"NextInClass"`
	// Defines the color transform level to use for this monster, affecting its color palette.
	TransLvl string `csv:"TransLvl"`
	// String Key. Used to define the monster's name, shown in the Life bar UI when targeted.
	NameStr string `csv:"NameStr"`
	// Controls a pointer to the "Id" of a monster to define which entry to use in the monstats2.txt file.
	MonStatsEx string `csv:"MonStatsEx"`
	// Points to the "Id" field from the MonProp.txt file. Used to add special modifiers to the monster.
	MonProp string `csv:"MonProp"`
	// Points to the "type" field from the MonType.txt file. Used to handle the monster's classification.
	MonType string `csv:"MonType"`
	// Points to a type of AI script to use for the monster (See monai.txt).
	AI string `csv:"AI"`
	// String Key. Used to add an additional description below the Life bar UI when the monster is targeted.
	DescStr string `csv:"DescStr"`
	// Controls the token used for choosing the proper cells to display the monster's graphics.
	Code string `csv:"Code"`
	// Boolean Field. If true, then this monster is allowed to spawn in the game.
	Enabled bool `csv:"enabled"`
	// Boolean Field. If true, then the monster will be classified as a ranged type.
	RangedType bool `csv:"rangedtype"`
	// Boolean Field. If true, then this monster will be treated as a spawner for other monsters.
	PlaceSpawn bool `csv:"placespawn"`
	// Points to the "Id" of another monster to control what kind of monster is spawned from this monster.
	Spawn string `csv:"spawn"`
	// Controls the X offset for where another monster is displaced when being spawned by this monster.
	SpawnX string `csv:"spawnx"`
	// Controls the Y offset for where another monster is displaced when being spawned by this monster.
	SpawnY string `csv:"spawny"`
	// Defines the animation mode that the spawned monsters will be initiated with.
	SpawnMode string `csv:"spawnmode"`
	// Points to the "Id" of another monster to control what kind of monster is spawned with this monster.
	Minion1 string `csv:"minion1"`
	// Points to the "Id" of another monster to control what kind of monster is spawned with this monster.
	Minion2 string `csv:"minion2"`
	// Boolean Field. If true, then set the monster AI to use the Boss AI type.
	SetBoss bool `csv:"SetBoss"`
	// Boolean Field. If true, then the monster's AI will transfer its boss recognition to another monster.
	BossXfer bool `csv:"BossXfer"`
	// The minimum number of minions that can spawn with this monster.
	PartyMin string `csv:"PartyMin"`
	// The maximum number of minions that can spawn with this monster.
	PartyMax string `csv:"PartyMax"`
	// The minimum number of duplicates of this monster that can spawn together.
	MinGrp string `csv:"MinGrp"`
	// The maximum number of duplicates of this monster that can spawn together.
	MaxGrp string `csv:"MaxGrp"`
	// Controls the percent chance that this monster does not spawn.
	SparsePopulate string `csv:"sparsePopulate"`
	Velocity       int    `csv:"Velocity"` // Determines the movement velocity of the monster.
	// Determines the run speed of the monster as opposed to walk speed.
	Run string `csv:"Run"`
	// Modifies the chance that this monster will be chosen to spawn in the area level.
	Rarity string `csv:"Rarity"`
	Level  int    `csv:"Level"` // Determines the monster's level.
	// Points to the "Id" field of a monster sound from the monsounds.txt file.
	MonSound string `csv:"MonSound"`
	// Points to the "Id" field of a monster sound from the monsounds.txt file.
	UMonSound string `csv:"UMonSound"`
	// Controls the AI threat value of the monster, affecting targeting priorities of enemy AIs.
	Threat string `csv:"threat"`
	// Controls the delay in frame length for how often the monster's AI will update its commands.
	Aidel int `csv:"aidel"`
	// Controls the maximum distance (measured in tiles) between the monster and an enemy until the monster's AI becomes
	// aggressive. If equals 0, then default to 35.
	Aidist int `csv:"aidist"`
	// Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP1 string `csv:"aip1"`
	// Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP2 string `csv:"aip2"`
	// Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP3 string `csv:"aip3"`
	// Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP4 string `csv:"aip4"`
	// Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP5 string `csv:"aip5"`
	// Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP6 string `csv:"aip6"`
	// Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP7 string `csv:"aip7"`
	// Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP8 string `csv:"aip8"`
	// Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Attack 1
	// mode.
	MissA1 string `csv:"MissA1"`
	// Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Attack 2
	// mode.
	MissA2 string `csv:"MissA2"`
	// Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Skill 1
	// mode.
	MissS1 string `csv:"MissS1"`
	// Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Skill 2
	// mode.
	MissS2 string `csv:"MissS2"`
	// Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Skill 3
	// mode.
	MissS3 string `csv:"MissS3"`
	// Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Skill 4
	// mode.
	MissS4 string `csv:"MissS4"`
	// Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Cast mode.
	MissC string `csv:"MissC"`
	// Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Sequence
	// mode.
	MissSQ string `csv:"MissSQ"`
	// Controls the monster's alignment, determining if it will be an enemy, ally, or neutral to the player.
	Align string `csv:"Align"`
	// Boolean Field. If true, then the monster is allowed to spawn in an area level.
	IsSpawn bool `csv:"isSpawn"`
	// Boolean Field. If true, then the monster is classified as a melee only type.
	IsMelee bool `csv:"isMelee"`
	// Boolean Field. If true, then the monster is classified as an NPC (Non-Playable Character).
	Npc bool `csv:"npc"`
	// Boolean Field. If true, then the monster is interactable, allowing the player to click on the monster to perform an
	// interact command.
	Interact bool `csv:"interact"`
	// Boolean Field. If true, then monster will have an inventory with randomly generated items.
	Inventory bool `csv:"inventory"`
	// Boolean Field. If true, then the monster is allowed to be in town.
	InTown bool `csv:"inTown"`
	// Boolean Field. If true, then the monster is treated as a Low Undead type.
	LUndead bool `csv:"lUndead"`
	// Boolean Field. If true, then the monster is treated as a High Undead type.
	HUndead bool `csv:"hUndead"`
	// Boolean Field. If true, then the monster is classified as a Demon type.
	Demon bool `csv:"demon"`
	// Boolean Field. If true, then the monster is flagged as a flying type.
	Flying bool `csv:"flying"`
	// Boolean Field. If true, then the monster will use its AI to open doors if necessary.
	OpenDoors bool `csv:"opendoors"`
	// Boolean Field. If true, then the monster is classified as a Boss type.
	Boss bool `csv:"boss"`
	// Boolean Field. If true, then the monster is classified as a Prime Evil type.
	PrimeEvil bool `csv:"primeevil"`
	// Boolean Field. If true, then the monster can be killed, damaged, and be put in a Death or Dead mode.
	Killable bool `csv:"killable"`
	// Boolean Field. If true, then monster's AI can be switched, such as by the Assassin's Mind Blast ability.
	SwitchAI bool `csv:"switchai"`
	// Boolean Field. If true, then the monster cannot be affected by friendly auras.
	NoAura bool `csv:"noAura"`
	// Boolean Field. If true, then the monster is not allowed to spawn with the Multi-Shot unique monster modifier.
	NoMultishot bool `csv:"nomultishot"`
	// Boolean Field. If true, then the monster is not counted on the list of active monsters in the area.
	NeverCount bool `csv:"neverCount"`
	// Boolean Field. If true, then pet AI scripts will ignore this monster.
	PetIgnore bool `csv:"petIgnore"`
	// Boolean Field. If true, then the monster will explode on death or use a general death damage function.
	DeathDmg bool `csv:"deathDmg"`
	// Boolean Field. If true, the monster is flagged as a possible selection for the AI generic spawner function.
	GenericSpawn bool `csv:"genericSpawn"`
	// Boolean Field. If true, then the monster will be flagged as a zoo type monster.
	Zoo bool `csv:"zoo"`
	// Boolean Field. If true, then the monster will not be able to be desecrated when inside a desecrated level.
	CannotDesecrate bool `csv:"CannotDesecrate"`
	// Determines what type of items the monster is allowed to hold in its right arm.
	RightArmItemType string `csv:"rightArmItemType"`
	// Determines what type of items the monster is allowed to hold in its left arm.
	LeftArmItemType string `csv:"leftArmItemType"`
	// Boolean Field. If true, then the monster can't use items marked as two-handed.
	CanNotUseTwoHandedItems bool `csv:"canNotUseTwoHandedItems"`
	// Determines which of the monster's skill's level should be sent to the client.
	SendSkills byte `csv:"SendSkills"`
	// Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires
	// "Sk#mode").
	Skill1 string `csv:"Skill1"`
	// Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires
	// "Sk#mode").
	Skill2 string `csv:"Skill2"`
	// Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires
	// "Sk#mode").
	Skill3 string `csv:"Skill3"`
	// Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires
	// "Sk#mode").
	Skill4 string `csv:"Skill4"`
	// Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires
	// "Sk#mode").
	Skill5 string `csv:"Skill5"`
	// Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires
	// "Sk#mode").
	Skill6 string `csv:"Skill6"`
	// Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires
	// "Sk#mode").
	Skill7 string `csv:"Skill7"`
	// Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires
	// "Sk#mode").
	Skill8 string `csv:"Skill8"`
	// Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the
	// monseq.txt file.
	Sk1Mode string `csv:"Sk1mode"`
	// Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the
	// monseq.txt file.
	Sk2Mode string `csv:"Sk2mode"`
	// Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the
	// monseq.txt file.
	Sk3Mode string `csv:"Sk3mode"`
	// Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the
	// monseq.txt file.
	Sk4Mode string `csv:"Sk4mode"`
	// Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the
	// monseq.txt file.
	Sk5Mode string `csv:"Sk5mode"`
	// Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the
	// monseq.txt file.
	Sk6Mode string `csv:"Sk6mode"`
	// Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the
	// monseq.txt file.
	Sk7Mode string `csv:"Sk7mode"`
	// Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the
	// monseq.txt file.
	Sk8Mode string `csv:"Sk8mode"`
	// Controls the base skill level of the related skill on the monster.
	Sk1Lvl string `csv:"Sk1lvl"`
	// Controls the base skill level of the related skill on the monster.
	Sk2Lvl string `csv:"Sk2lvl"`
	// Controls the base skill level of the related skill on the monster.
	Sk3Lvl string `csv:"Sk3lvl"`
	// Controls the base skill level of the related skill on the monster.
	Sk4Lvl string `csv:"Sk4lvl"`
	// Controls the base skill level of the related skill on the monster.
	Sk5Lvl string `csv:"Sk5lvl"`
	// Controls the base skill level of the related skill on the monster.
	Sk6Lvl string `csv:"Sk6lvl"`
	// Controls the base skill level of the related skill on the monster.
	Sk7Lvl string `csv:"Sk7lvl"`
	// Controls the base skill level of the related skill on the monster.
	Sk8Lvl string `csv:"Sk8lvl"`
	// Controls the monster's overall Life and Mana steal percentage.
	Drain string `csv:"Drain"`
	// Sets the percentage change in movement speed and attack rate when the monster is chilled by a cold effect.
	ColdEffect  string `csv:"coldeffect"`
	ResDm       string `csv:"ResDm"`       // Sets the monster's Physical Damage Resistance stat.
	ResMa       string `csv:"ResMa"`       // Sets the monster's Magic Resistance stat.
	ResFi       string `csv:"ResFi"`       // Sets the monster's Fire Resistance stat.
	ResLi       string `csv:"ResLi"`       // Sets the monster's Lightning Resistance stat.
	ResCo       string `csv:"ResCo"`       // Sets the monster's Cold Resistance stat.
	ResPo       string `csv:"ResPo"`       // Sets the monster's Poison Resistance stat.
	DamageRegen string `csv:"DamageRegen"` // Controls the monster's Life regeneration per frame.
	// Points to a skill from the "skill" field in the skills.txt file. Changes the monster's min physical damage, max
	// physical damage, and Attack Rating.
	SkillDamage string `csv:"SkillDamage"`
	// Boolean Field. If true, then use this file's fields to determine the monster's baseline stats. If false, use the
	// MonLvl.txt file to determine the monster's baseline stats.
	NoRatio bool `csv:"noRatio"`
	// If equals 1, then the monster can block without a shield. If equals 2, then the monster cannot block at all, even
	// with a shield equipped. If equals 0, then ignore this.
	ShieldBlockOverride string `csv:"ShieldBlockOverride"`
	ToBlock             string `csv:"ToBlock"` // The monster's percent chance to block an attack.
	// The percent chance for the monster to score a critical hit when attacking an enemy, causing the attack to deal
	// double damage.
	Crit    string `csv:"Crit"`
	MinHP   int    `csv:"minHP"` // Normal-difficulty life ratio or direct value.
	MinHPN  int    `csv:"minHP(N)"`
	MinHPH  int    `csv:"minHP(H)"`
	MaxHP   int    `csv:"maxHP"`
	MaxHPN  int    `csv:"maxHP(N)"`
	MaxHPH  int    `csv:"maxHP(H)"`
	AC      int    `csv:"AC"`
	ACN     int    `csv:"AC(N)"`
	ACH     int    `csv:"AC(H)"`
	Exp     int    `csv:"Exp"`
	ExpN    int    `csv:"Exp(N)"`
	ExpH    int    `csv:"Exp(H)"`
	A1MinD  int    `csv:"A1MinD"`
	A1MinDN int    `csv:"A1MinD(N)"`
	A1MinDH int    `csv:"A1MinD(H)"`
	A1MaxD  int    `csv:"A1MaxD"`
	A1MaxDN int    `csv:"A1MaxD(N)"`
	A1MaxDH int    `csv:"A1MaxD(H)"`
	A1TH    int    `csv:"A1TH"`
	A1THN   int    `csv:"A1TH(N)"`
	A1THH   int    `csv:"A1TH(H)"`
	// The minimum damage dealt by the monster when using the Attack 2 (A2) animation mode.
	A2MinD string `csv:"A2MinD"`
	// The maximum damage dealt by the monster when using the Attack 2 (A2) animation mode.
	A2MaxD string `csv:"A2MaxD"`
	// The monster's Attack Rating when using the Attack 2 (A2) animation mode.
	A2TH string `csv:"A2TH"`
	// The minimum damage dealt by the monster when using the Skill 1 (S1) animation mode.
	S1MinD string `csv:"S1MinD"`
	// The maximum damage dealt by the monster when using the Skill 1 (S1) animation mode.
	S1MaxD string `csv:"S1MaxD"`
	// The monster's Attack Rating when using the Skill 1 (S1) animation mode.
	S1TH string `csv:"S1TH"`
	// Determines which animation mode will trigger an additional elemental damage type when used.
	El1Mode string `csv:"El1Mode"`
	// Controls the random percent chance (out of 100) that the monster will append the element damage to the attack. This
	// field is used when El#Mode is not null.
	El1Pct string `csv:"El1Pct"`
	// The minimum element damage applied to the attack. This field is used when El#Mode is not null.
	El1MinD string `csv:"El1MinD"`
	// The maximum element damage applied to the attack. This field is used when El#Mode is not null.
	El1MaxD string `csv:"El1MaxD"`
	// Controls the duration of the related element mode in frame lengths (25 Frames = 1 Second). This is only applicable
	// for the Cold, Poison, Stun, Burning, Freeze elements. This field is used when El#Mode is not null.
	El1Dur string `csv:"El1Dur"`
	// Determines which animation mode will trigger an additional elemental damage type when used.
	El2Mode string `csv:"El2Mode"`
	// Controls the random percent chance (out of 100) that the monster will append the element damage to the attack. This
	// field is used when El#Mode is not null.
	El2Pct string `csv:"El2Pct"`
	// The minimum element damage applied to the attack. This field is used when El#Mode is not null.
	El2MinD string `csv:"El2MinD"`
	// The maximum element damage applied to the attack. This field is used when El#Mode is not null.
	El2MaxD string `csv:"El2MaxD"`
	// Controls the duration of the related element mode in frame lengths (25 Frames = 1 Second). This is only applicable
	// for the Cold, Poison, Stun, Burning, Freeze elements. This field is used when El#Mode is not null.
	El2Dur string `csv:"El2Dur"`
	// Determines which animation mode will trigger an additional elemental damage type when used.
	El3Mode string `csv:"El3Mode"`
	// Controls the random percent chance (out of 100) that the monster will append the element damage to the attack. This
	// field is used when El#Mode is not null.
	El3Pct string `csv:"El3Pct"`
	// The minimum element damage applied to the attack. This field is used when El#Mode is not null.
	El3MinD string `csv:"El3MinD"`
	// The maximum element damage applied to the attack. This field is used when El#Mode is not null.
	El3MaxD string `csv:"El3MaxD"`
	// Controls the duration of the related element mode in frame lengths (25 Frames = 1 Second). This is only applicable
	// for the Cold, Poison, Stun, Burning, Freeze elements. This field is used when El#Mode is not null.
	El3Dur string `csv:"El3Dur"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass1 string `csv:"TreasureClass1"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass2 string `csv:"TreasureClass2"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass3 string `csv:"TreasureClass3"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass4 string `csv:"TreasureClass4"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass1Nightmare string `csv:"TreasureClass1(N)"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass2Nightmare string `csv:"TreasureClass2(N)"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass3Nightmare string `csv:"TreasureClass3(N)"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass4Nightmare string `csv:"TreasureClass4(N)"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass1Hell string `csv:"TreasureClass1(H)"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass2Hell string `csv:"TreasureClass2(H)"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass3Hell string `csv:"TreasureClass3(H)"`
	// Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the
	// TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass4Hell string `csv:"TreasureClass4(H)"`
	// Checks to see if the player has a quest flag progress. If not, then use the "TreasureClass4" field, based on the
	// game's current difficulty.
	TCQuestId string `csv:"TCQuestId"`
	// Indicates if the respecialization from Akara is completed.
	RespecFromAkara bool `csv:"Respecialization from Akara is Completed"`
	// Controls which Quest Checkpoint, or current progress within a quest (based on the "TCQuestId" value), is needed to
	// use the "TreasureClass4" field, based on the game's current difficulty.
	TCQuestCP string `csv:"TCQuestCP"`
}

// MonsterStats2 represents additional functionalities and statistics for a monster in the game (continuation of
// MonsterStats1).
type MonsterStats2 struct {
	// Controls the unique name ID to define the monster. This must match the same value in the monstats.txt file.
	Id            string `csv:"Id"`
	Height        int    `csv:"Height"`        // Determines the height of the monster.
	Code1         string `csv:"Code1"`         // Code for the first purpose of Height.
	Code1Desc     string `csv:"Description1"`  // Description for the first purpose of Height.
	Code2         string `csv:"Code2"`         // Code for the second purpose of Height.
	Code2Desc     string `csv:"Description2"`  // Description for the second purpose of Height.
	OverlayHeight int    `csv:"OverlayHeight"` // Determines the height value of overlays for the monster.
	// Determines the pixel height value for the damage bar when the monster is selected.
	PixHeight int `csv:"pixHeight"`
	// Determines the tile grid size of the monster for handling placement when the monster spawns or uses movement skills.
	SizeX int `csv:"SizeX"`
	// Determines the tile grid size of the monster for handling placement when the monster spawns or uses movement skills.
	SizeY int `csv:"SizeY"`
	// Controls the method for spawning the monster based on the collisions in the environment.
	SpawnColCode   int    `csv:"spawnCol"`
	SpawnColDesc   string `csv:"Description3"` // Description for the SpawnColCode.
	MeleeRng       int    `csv:"MeleeRng"`     // Controls the range of the monster's melee attack.
	BaseWeapon     string `csv:"BaseW"`        // Defines the monster's base weapon class.
	BaseWeaponDesc string `csv:"Description4"` // Description for the BaseWeapon.
	// Defines the specific class of an attack when the monster successfully hits with an attack.
	HitClass     int    `csv:"HitClass"`
	HitClassDesc string `csv:"Description5"` // Description for the HitClass.
	HDv          string `csv:"HDv"`          // Head visual.
	TRv          string `csv:"TRv"`          // Torso visual.
	LGv          string `csv:"LGv"`          // Legs visual.
	RAv          string `csv:"RAv"`          // Right Arm visual.
	LAv          string `csv:"LAv"`          // Left Arm visual.
	RHv          string `csv:"RHv"`          // Right Hand visual.
	LHv          string `csv:"LHv"`          // Left Hand visual.
	SHv          string `csv:"SHv"`          // Shield visual.
	S1v          string `csv:"S1v"`          // Special 1 visual.
	S2v          string `csv:"S2v"`          // Special 2 visual.
	S3v          string `csv:"S3v"`          // Special 3 visual.
	S4v          string `csv:"S4v"`          // Special 4 visual.
	S5v          string `csv:"S5v"`          // Special 5 visual.
	S6v          string `csv:"S6v"`          // Special 6 visual.
	S7v          string `csv:"S7v"`          // Special 7 visual.
	S8v          string `csv:"S8v"`          // Special 8 visual.
	HD           bool   `csv:"HD"`           // Head enabled.
	TR           bool   `csv:"TR"`           // Torso enabled.
	LG           bool   `csv:"LG"`           // Legs enabled.
	RA           bool   `csv:"RA"`           // Right Arm enabled.
	LA           bool   `csv:"LA"`           // Left Arm enabled.
	RH           bool   `csv:"RH"`           // Right Hand enabled.
	LH           bool   `csv:"LH"`           // Left Hand enabled.
	SH           bool   `csv:"SH"`           // Shield enabled.
	S1           bool   `csv:"S1"`           // Special 1 enabled.
	S2           bool   `csv:"S2"`           // Special 2 enabled.
	S3           bool   `csv:"S3"`           // Special 3 enabled.
	S4           bool   `csv:"S4"`           // Special 4 enabled.
	S5           bool   `csv:"S5"`           // Special 5 enabled.
	S6           bool   `csv:"S6"`           // Special 6 enabled.
	S7           bool   `csv:"S7"`           // Special 7 enabled.
	S8           bool   `csv:"S8"`           // Special 8 enabled.
	// Defines the total amount of component pieces that the monster uses.
	TotalPieces   int  `csv:"TotalPieces"`
	DeathMode     bool `csv:"mDT"` // If equals 1, then enable the Death Mode for the monster.
	NeutralMode   bool `csv:"mNU"` // If equals 1, then enable the Neutral Mode for the monster.
	WalkMode      bool `csv:"mWL"` // If equals 1, then enable the Walk Mode for the monster.
	GetHitMode    bool `csv:"mGH"` // If equals 1, then enable the Get Hit Mode for the monster.
	Attack1Mode   bool `csv:"mA1"` // If equals 1, then enable the Attack 1 Mode for the monster.
	Attack2Mode   bool `csv:"mA2"` // If equals 1, then enable the Attack 2 Mode for the monster.
	BlockMode     bool `csv:"mBL"` // If equals 1, then enable the Block Mode for the monster.
	CastMode      bool `csv:"mSC"` // If equals 1, then enable the Cast Mode for the monster.
	Skill1Mode    bool `csv:"mS1"` // If equals 1, then enable the Skill 1 Mode for the monster.
	Skill2Mode    bool `csv:"mS2"` // If equals 1, then enable the Skill 2 Mode for the monster.
	Skill3Mode    bool `csv:"mS3"` // If equals 1, then enable the Skill 3 Mode for the monster.
	Skill4Mode    bool `csv:"mS4"` // If equals 1, then enable the Skill 4 Mode for the monster.
	DeadMode      bool `csv:"mDD"` // If equals 1, then enable the Dead Mode for the monster.
	KnockbackMode bool `csv:"mKB"` // If equals 1, then enable the Knockback Mode for the monster.
	SequenceMode  bool `csv:"mSQ"` // If equals 1, then enable the Sequence Mode for the monster.
	RunMode       bool `csv:"mRN"` // If equals 1, then enable the Run Mode for the monster.
	// Defines the number of directions that the monster can face during Death Mode.
	DeathDirections int `csv:"dDT"`
	// Defines the number of directions that the monster can face during Neutral Mode.
	NeutralDirections int `csv:"dNU"`
	// Defines the number of directions that the monster can face during Walk Mode.
	WalkDirections int `csv:"dWL"`
	// Defines the number of directions that the monster can face during Get Hit Mode.
	GetHitDirections int `csv:"dGH"`
	// Defines the number of directions that the monster can face during Attack 1 Mode.
	Attack1Directions int `csv:"dA1"`
	// Defines the number of directions that the monster can face during Attack 2 Mode.
	Attack2Directions int `csv:"dA2"`
	// Defines the number of directions that the monster can face during Block Mode.
	BlockDirections int `csv:"dBL"`
	// Defines the number of directions that the monster can face during Cast Mode.
	CastDirections int `csv:"dSC"`
	// Defines the number of directions that the monster can face during Skill 1 Mode.
	Skill1Directions int `csv:"dS1"`
	// Defines the number of directions that the monster can face during Skill 2 Mode.
	Skill2Directions int `csv:"dS2"`
	// Defines the number of directions that the monster can face during Skill 3 Mode.
	Skill3Directions int `csv:"dS3"`
	// Defines the number of directions that the monster can face during Skill 4 Mode.
	Skill4Directions int `csv:"dS4"`
	// Defines the number of directions that the monster can face during Dead Mode.
	DeadDirections int `csv:"dDD"`
	// Defines the number of directions that the monster can face during Knockback Mode.
	KnockbackDirections int `csv:"dKB"`
	// Defines the number of directions that the monster can face during Sequence Mode.
	SequenceDirections int `csv:"dSQ"`
	// Defines the number of directions that the monster can face during Run Mode.
	RunDirections int `csv:"dRN"`
	// If equals 1, then enable the Attack 1 Mode while the monster is moving with the Walk mode or Run mode.
	Attack1Moving bool `csv:"A1mv"`
	// If equals 1, then enable the Attack 2 Mode while the monster is moving with the Walk mode or Run mode.
	Attack2Moving bool `csv:"A2mv"`
	// If equals 1, then enable the Cast Mode while the monster is moving with the Walk mode or Run mode.
	CastMoving bool `csv:"SCmv"`
	// If equals 1, then enable the Skill 1 Mode while the monster is moving with the Walk mode or Run mode.
	Skill1Moving bool `csv:"S1mv"`
	// If equals 1, then enable the Skill 2 Mode while the monster is moving with the Walk mode or Run mode.
	Skill2Moving bool `csv:"S2mv"`
	// If equals 1, then enable the Skill 3 Mode while the monster is moving with the Walk mode or Run mode.
	Skill3Moving bool `csv:"S3mv"`
	// If equals 1, then enable the Skill 4 Mode while the monster is moving with the Walk mode or Run mode.
	Skill4Moving bool `csv:"S4mv"`
	// If equals 1, then enable the mouse selection bounding box functionality around the monster.
	NoGfxHitTest bool `csv:"noGfxHitTest"`
	// Define the pixel top offset around the monster for the mouse selection bounding box functionality.
	HtTop int `csv:"htTop"`
	// Define the pixel left offset around the monster for the mouse selection bounding box functionality.
	HtLeft int `csv:"htLeft"`
	// Define the pixel right offset around the monster for the mouse selection bounding box functionality.
	HtWidth int `csv:"htWidth"`
	// Define the pixel bottom offset around the monster for the mouse selection bounding box functionality.
	HtHeight int `csv:"htHeight"`
	// Determines if the monster should be placed on the inactive list, to be saved when the level unloads.
	Restore int `csv:"restore"`
	// Controls what index of the Automap tiles to use to display this monster on the Automap.
	AutomapCel int  `csv:"automapCel"`
	NoMap      bool `csv:"noMap"` // If equals 1, then the monster will not appear on the Automap.
	// If equals 1, then no looping overlays will be drawn on the monster.
	NoOvly       bool `csv:"noOvly"`
	IsSelectable bool `csv:"isSel"` // If equals 1, then the monster is selectable and can be targeted.
	// If equals 1, then the player can always select the monster, regardless of being an ally or enemy.
	AlwaysSelectable bool `csv:"alSel"`
	NeverSelectable  bool `csv:"noSel"` // If equals 1, then the player can never select the monster.
	// If equals 1, then the player can target this monster when holding the Shift key and clicking to use a skill.
	ShiftSelectable bool `csv:"shiftSel"`
	// If equals 1, then the monster's corpse can be with the mouse cursor.
	CorpseSelectable bool `csv:"corpseSel"`
	IsAttackable     bool `csv:"isAtt"` // If equals 1, then the monster can be attacked.
	// If equals 1, then the monster is allowed to be revived by the Necromancer Revive skill.
	CanRevive bool `csv:"revive"`
	// If equals 1, then the monster's corpse will be placed into a pool with all other corpses with this field checked.
	LimitCorpses bool `csv:"limitCorpses"`
	IsCritter    bool `csv:"critter"` // If equals 1, then the monster will be flagged as a critter.
	IsSmallType  bool `csv:"small"`   // If equals 1, then the monster will be classified as a small type.
	IsLargeType  bool `csv:"large"`   // If equals 1, then the monster will be classified as a large type.
	// If equals 1, then the monster's corpse is classified as soft-bodied.
	IsSoftBodied bool `csv:"soft"`
	IsInert      bool `csv:"inert"` // If equals 1, then the monster will never attack its enemies.
	// If equals 1 and the monster class is "barricadedoor", "barricadedoor2", or "evilhut", then the monster will place an
	// invisible object with collision.
	HasObjectCollision bool `csv:"objCol"`
	// If equals 1, then the monster's corpse will have collision with other units.
	HasDeadCollision bool `csv:"deadCol"`
	// If equals 1, then ignore the corpse draw order for rendering the sprite on top of others, while the monster is dead.
	IsUnflatDead bool `csv:"unflatDead"`
	// If equals 1, then the monster will project a shadow on the ground.
	HasShadow bool `csv:"Shadow"`
	// If equals 1 and the monster is a Unique monster, then the monster will not have random color palette transform
	// shifts.
	NoUniqueShift bool `csv:"noUniqueShift"`
	// If equals 1, then the monster's Death Mode and Dead mode will make use of its component system.
	UseComponentDeath bool `csv:"compositeDeath"`
	// Controls the color of the monster's blood based on the region locale.
	LocalBloodColor int `csv:"localBlood"`
	Bleed           int `csv:"Bleed"`   // Controls if the monster will create blood missiles.
	LightRadiusSize int `csv:"Light"`   // Controls the monster's minimum Light Radius size.
	LightColorR     int `csv:"light-r"` // Controls the red color value of the monster's Light Radius.
	LightColorG     int `csv:"light-g"` // Controls the green color value of the monster's Light Radius.
	LightColorB     int `csv:"light-b"` // Controls the blue color value of the monster's Light Radius.
	// Modifies the color palette transform for the monster in Normal difficulty.
	Utrans int `csv:"Utrans"`
	// Modifies the color palette transform for the monster in Nightmare difficulty.
	UtransN int `csv:"Utrans(N)"`
	// Modifies the color palette transform for the monster in Hell difficulty.
	UtransH    int    `csv:"Utrans(H)"`
	UtransDesc string `csv:"Description6"` // Description for the Utrans codes.
	// The frame length to hold the channel cast time of the inferno skill.
	InfernoLength int `csv:"InfernoLen"`
	// The exact frame in the channel animation to loop back and start at again.
	InfernoAnim int `csv:"InfernoAnim"`
	// The exact frame in the channel animation to determine when to roll back to the "InfernoAnim" frame.
	InfernoRollback int `csv:"InfernoRollback"`
	// Controls which monster mode to set on the monster when it is resurrected.
	ResurrectMode     string `csv:"ResurrectMode"`
	ResurrectModeDesc string `csv:"Description7"` // Description for the ResurrectMode codes.
	// Controls what skill should the monster use when it is resurrected.
	ResurrectSkill string `csv:"ResurrectSkill"`
	// Controls what unique modifier the monster should always spawn with.
	SpawnUniqueMod string `csv:"SpawnUniqueMod"`
}
