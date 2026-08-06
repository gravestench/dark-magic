package models

// MonsterStats represents the main functionalities and statistics for a monster in the game.
type MonsterStats struct {
	Id                      string `csv:"Id"`                                       // Controls the unique name ID to define the monster.
	BaseId                  string `csv:"BaseId"`                                   // Points to the "Id" of another monster to define the monster's base type.
	NextInClass             string `csv:"NextInClass"`                              // Points to the "Id" of another monster to signify the next monster in the group of this monster's type.
	TransLvl                string `csv:"TransLvl"`                                 // Defines the color transform level to use for this monster, affecting its color palette.
	NameStr                 string `csv:"NameStr"`                                  // String Key. Used to define the monster's name, shown in the Life bar UI when targeted.
	MonStatsEx              string `csv:"MonStatsEx"`                               // Controls a pointer to the "Id" of a monster to define which entry to use in the monstats2.txt file.
	MonProp                 string `csv:"MonProp"`                                  // Points to the "Id" field from the MonProp.txt file. Used to add special modifiers to the monster.
	MonType                 string `csv:"MonType"`                                  // Points to the "type" field from the MonType.txt file. Used to handle the monster's classification.
	AI                      string `csv:"AI"`                                       // Points to a type of AI script to use for the monster (See monai.txt).
	DescStr                 string `csv:"DescStr"`                                  // String Key. Used to add an additional description below the Life bar UI when the monster is targeted.
	Code                    string `csv:"Code"`                                     // Controls the token used for choosing the proper cells to display the monster's graphics.
	Enabled                 bool   `csv:"enabled"`                                  // Boolean Field. If true, then this monster is allowed to spawn in the game.
	RangedType              bool   `csv:"rangedtype"`                               // Boolean Field. If true, then the monster will be classified as a ranged type.
	PlaceSpawn              bool   `csv:"placespawn"`                               // Boolean Field. If true, then this monster will be treated as a spawner for other monsters.
	Spawn                   string `csv:"spawn"`                                    // Points to the "Id" of another monster to control what kind of monster is spawned from this monster.
	SpawnX                  string `csv:"spawnx"`                                   // Controls the X offset for where another monster is displaced when being spawned by this monster.
	SpawnY                  string `csv:"spawny"`                                   // Controls the Y offset for where another monster is displaced when being spawned by this monster.
	SpawnMode               string `csv:"spawnmode"`                                // Defines the animation mode that the spawned monsters will be initiated with.
	Minion1                 string `csv:"minion1"`                                  // Points to the "Id" of another monster to control what kind of monster is spawned with this monster.
	Minion2                 string `csv:"minion2"`                                  // Points to the "Id" of another monster to control what kind of monster is spawned with this monster.
	SetBoss                 bool   `csv:"SetBoss"`                                  // Boolean Field. If true, then set the monster AI to use the Boss AI type.
	BossXfer                bool   `csv:"BossXfer"`                                 // Boolean Field. If true, then the monster's AI will transfer its boss recognition to another monster.
	PartyMin                string `csv:"PartyMin"`                                 // The minimum number of minions that can spawn with this monster.
	PartyMax                string `csv:"PartyMax"`                                 // The maximum number of minions that can spawn with this monster.
	MinGrp                  string `csv:"MinGrp"`                                   // The minimum number of duplicates of this monster that can spawn together.
	MaxGrp                  string `csv:"MaxGrp"`                                   // The maximum number of duplicates of this monster that can spawn together.
	SparsePopulate          string `csv:"sparsePopulate"`                           // Controls the percent chance that this monster does not spawn.
	Velocity                string `csv:"Velocity"`                                 // Determines the movement velocity of the monster.
	Run                     string `csv:"Run"`                                      // Determines the run speed of the monster as opposed to walk speed.
	Rarity                  string `csv:"Rarity"`                                   // Modifies the chance that this monster will be chosen to spawn in the area level.
	Level                   string `csv:"Level"`                                    // Determines the monster's level.
	MonSound                string `csv:"MonSound"`                                 // Points to the "Id" field of a monster sound from the monsounds.txt file.
	UMonSound               string `csv:"UMonSound"`                                // Points to the "Id" field of a monster sound from the monsounds.txt file.
	Threat                  string `csv:"threat"`                                   // Controls the AI threat value of the monster, affecting targeting priorities of enemy AIs.
	Aidel                   string `csv:"aidel"`                                    // Controls the delay in frame length for how often the monster's AI will update its commands.
	Aidist                  string `csv:"aidist"`                                   // Controls the maximum distance (measured in tiles) between the monster and an enemy until the monster's AI becomes aggressive. If equals 0, then default to 35.
	AIP1                    string `csv:"aip1"`                                     // Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP2                    string `csv:"aip2"`                                     // Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP3                    string `csv:"aip3"`                                     // Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP4                    string `csv:"aip4"`                                     // Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP5                    string `csv:"aip5"`                                     // Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP6                    string `csv:"aip6"`                                     // Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP7                    string `csv:"aip7"`                                     // Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	AIP8                    string `csv:"aip8"`                                     // Numeric parameter used to control various functions of the monster's AI. Depends on the AI script being used.
	MissA1                  string `csv:"MissA1"`                                   // Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Attack 1 mode.
	MissA2                  string `csv:"MissA2"`                                   // Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Attack 2 mode.
	MissS1                  string `csv:"MissS1"`                                   // Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Skill 1 mode.
	MissS2                  string `csv:"MissS2"`                                   // Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Skill 2 mode.
	MissS3                  string `csv:"MissS3"`                                   // Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Skill 3 mode.
	MissS4                  string `csv:"MissS4"`                                   // Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Skill 4 mode.
	MissC                   string `csv:"MissC"`                                    // Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Cast mode.
	MissSQ                  string `csv:"MissSQ"`                                   // Points to the "Missile" field from Missiles.txt to determine which missile to use when the monster is in Sequence mode.
	Align                   string `csv:"Align"`                                    // Controls the monster's alignment, determining if it will be an enemy, ally, or neutral to the player.
	IsSpawn                 bool   `csv:"isSpawn"`                                  // Boolean Field. If true, then the monster is allowed to spawn in an area level.
	IsMelee                 bool   `csv:"isMelee"`                                  // Boolean Field. If true, then the monster is classified as a melee only type.
	Npc                     bool   `csv:"npc"`                                      // Boolean Field. If true, then the monster is classified as an NPC (Non-Playable Character).
	Interact                bool   `csv:"interact"`                                 // Boolean Field. If true, then the monster is interactable, allowing the player to click on the monster to perform an interact command.
	Inventory               bool   `csv:"inventory"`                                // Boolean Field. If true, then monster will have an inventory with randomly generated items.
	InTown                  bool   `csv:"inTown"`                                   // Boolean Field. If true, then the monster is allowed to be in town.
	LUndead                 bool   `csv:"lUndead"`                                  // Boolean Field. If true, then the monster is treated as a Low Undead type.
	HUndead                 bool   `csv:"hUndead"`                                  // Boolean Field. If true, then the monster is treated as a High Undead type.
	Demon                   bool   `csv:"demon"`                                    // Boolean Field. If true, then the monster is classified as a Demon type.
	Flying                  bool   `csv:"flying"`                                   // Boolean Field. If true, then the monster is flagged as a flying type.
	OpenDoors               bool   `csv:"opendoors"`                                // Boolean Field. If true, then the monster will use its AI to open doors if necessary.
	Boss                    bool   `csv:"boss"`                                     // Boolean Field. If true, then the monster is classified as a Boss type.
	PrimeEvil               bool   `csv:"primeevil"`                                // Boolean Field. If true, then the monster is classified as a Prime Evil type.
	Killable                bool   `csv:"killable"`                                 // Boolean Field. If true, then the monster can be killed, damaged, and be put in a Death or Dead mode.
	SwitchAI                bool   `csv:"switchai"`                                 // Boolean Field. If true, then monster's AI can be switched, such as by the Assassin's Mind Blast ability.
	NoAura                  bool   `csv:"noAura"`                                   // Boolean Field. If true, then the monster cannot be affected by friendly auras.
	NoMultishot             bool   `csv:"nomultishot"`                              // Boolean Field. If true, then the monster is not allowed to spawn with the Multi-Shot unique monster modifier.
	NeverCount              bool   `csv:"neverCount"`                               // Boolean Field. If true, then the monster is not counted on the list of active monsters in the area.
	PetIgnore               bool   `csv:"petIgnore"`                                // Boolean Field. If true, then pet AI scripts will ignore this monster.
	DeathDmg                bool   `csv:"deathDmg"`                                 // Boolean Field. If true, then the monster will explode on death or use a general death damage function.
	GenericSpawn            bool   `csv:"genericSpawn"`                             // Boolean Field. If true, the monster is flagged as a possible selection for the AI generic spawner function.
	Zoo                     bool   `csv:"zoo"`                                      // Boolean Field. If true, then the monster will be flagged as a zoo type monster.
	CannotDesecrate         bool   `csv:"CannotDesecrate"`                          // Boolean Field. If true, then the monster will not be able to be desecrated when inside a desecrated level.
	RightArmItemType        string `csv:"rightArmItemType"`                         // Determines what type of items the monster is allowed to hold in its right arm.
	LeftArmItemType         string `csv:"leftArmItemType"`                          // Determines what type of items the monster is allowed to hold in its left arm.
	CanNotUseTwoHandedItems bool   `csv:"canNotUseTwoHandedItems"`                  // Boolean Field. If true, then the monster can't use items marked as two-handed.
	SendSkills              byte   `csv:"SendSkills"`                               // Determines which of the monster's skill's level should be sent to the client.
	Skill1                  string `csv:"Skill1"`                                   // Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires "Sk#mode").
	Skill2                  string `csv:"Skill2"`                                   // Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires "Sk#mode").
	Skill3                  string `csv:"Skill3"`                                   // Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires "Sk#mode").
	Skill4                  string `csv:"Skill4"`                                   // Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires "Sk#mode").
	Skill5                  string `csv:"Skill5"`                                   // Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires "Sk#mode").
	Skill6                  string `csv:"Skill6"`                                   // Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires "Sk#mode").
	Skill7                  string `csv:"Skill7"`                                   // Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires "Sk#mode").
	Skill8                  string `csv:"Skill8"`                                   // Points to a skill from the "skill" field in the skills.txt file. Gives the monster the skill to use (requires "Sk#mode").
	Sk1Mode                 string `csv:"Sk1mode"`                                  // Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the monseq.txt file.
	Sk2Mode                 string `csv:"Sk2mode"`                                  // Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the monseq.txt file.
	Sk3Mode                 string `csv:"Sk3mode"`                                  // Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the monseq.txt file.
	Sk4Mode                 string `csv:"Sk4mode"`                                  // Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the monseq.txt file.
	Sk5Mode                 string `csv:"Sk5mode"`                                  // Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the monseq.txt file.
	Sk6Mode                 string `csv:"Sk6mode"`                                  // Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the monseq.txt file.
	Sk7Mode                 string `csv:"Sk7mode"`                                  // Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the monseq.txt file.
	Sk8Mode                 string `csv:"Sk8mode"`                                  // Determines the monster's animation mode when using the related skill. Can also point to a "sequence" defined in the monseq.txt file.
	Sk1Lvl                  string `csv:"Sk1lvl"`                                   // Controls the base skill level of the related skill on the monster.
	Sk2Lvl                  string `csv:"Sk2lvl"`                                   // Controls the base skill level of the related skill on the monster.
	Sk3Lvl                  string `csv:"Sk3lvl"`                                   // Controls the base skill level of the related skill on the monster.
	Sk4Lvl                  string `csv:"Sk4lvl"`                                   // Controls the base skill level of the related skill on the monster.
	Sk5Lvl                  string `csv:"Sk5lvl"`                                   // Controls the base skill level of the related skill on the monster.
	Sk6Lvl                  string `csv:"Sk6lvl"`                                   // Controls the base skill level of the related skill on the monster.
	Sk7Lvl                  string `csv:"Sk7lvl"`                                   // Controls the base skill level of the related skill on the monster.
	Sk8Lvl                  string `csv:"Sk8lvl"`                                   // Controls the base skill level of the related skill on the monster.
	Drain                   string `csv:"Drain"`                                    // Controls the monster's overall Life and Mana steal percentage.
	ColdEffect              string `csv:"coldeffect"`                               // Sets the percentage change in movement speed and attack rate when the monster is chilled by a cold effect.
	ResDm                   string `csv:"ResDm"`                                    // Sets the monster's Physical Damage Resistance stat.
	ResMa                   string `csv:"ResMa"`                                    // Sets the monster's Magic Resistance stat.
	ResFi                   string `csv:"ResFi"`                                    // Sets the monster's Fire Resistance stat.
	ResLi                   string `csv:"ResLi"`                                    // Sets the monster's Lightning Resistance stat.
	ResCo                   string `csv:"ResCo"`                                    // Sets the monster's Cold Resistance stat.
	ResPo                   string `csv:"ResPo"`                                    // Sets the monster's Poison Resistance stat.
	DamageRegen             string `csv:"DamageRegen"`                              // Controls the monster's Life regeneration per frame.
	SkillDamage             string `csv:"SkillDamage"`                              // Points to a skill from the "skill" field in the skills.txt file. Changes the monster's min physical damage, max physical damage, and Attack Rating.
	NoRatio                 bool   `csv:"noRatio"`                                  // Boolean Field. If true, then use this file's fields to determine the monster's baseline stats. If false, use the MonLvl.txt file to determine the monster's baseline stats.
	ShieldBlockOverride     string `csv:"ShieldBlockOverride"`                      // If equals 1, then the monster can block without a shield. If equals 2, then the monster cannot block at all, even with a shield equipped. If equals 0, then ignore this.
	ToBlock                 string `csv:"ToBlock"`                                  // The monster's percent chance to block an attack.
	Crit                    string `csv:"Crit"`                                     // The percent chance for the monster to score a critical hit when attacking an enemy, causing the attack to deal double damage.
	MinHP                   string `csv:"minHP"`                                    // The monster's minimum amount of Life when spawned.
	MaxHP                   string `csv:"maxHP"`                                    // The monster's maximum amount of Life when spawned.
	AC                      string `csv:"AC"`                                       // The monster's Defense value.
	Exp                     string `csv:"Exp"`                                      // The amount of Experience that is rewarded to the player when the monster is killed.
	A1MinD                  string `csv:"A1MinD"`                                   // The minimum damage dealt by the monster when using the Attack 1 (A1) animation mode.
	A1MaxD                  string `csv:"A1MaxD"`                                   // The maximum damage dealt by the monster when using the Attack 1 (A1) animation mode.
	A1TH                    string `csv:"A1TH"`                                     // The monster's Attack Rating when using the Attack 1 (A1) animation mode.
	A2MinD                  string `csv:"A2MinD"`                                   // The minimum damage dealt by the monster when using the Attack 2 (A2) animation mode.
	A2MaxD                  string `csv:"A2MaxD"`                                   // The maximum damage dealt by the monster when using the Attack 2 (A2) animation mode.
	A2TH                    string `csv:"A2TH"`                                     // The monster's Attack Rating when using the Attack 2 (A2) animation mode.
	S1MinD                  string `csv:"S1MinD"`                                   // The minimum damage dealt by the monster when using the Skill 1 (S1) animation mode.
	S1MaxD                  string `csv:"S1MaxD"`                                   // The maximum damage dealt by the monster when using the Skill 1 (S1) animation mode.
	S1TH                    string `csv:"S1TH"`                                     // The monster's Attack Rating when using the Skill 1 (S1) animation mode.
	El1Mode                 string `csv:"El1Mode"`                                  // Determines which animation mode will trigger an additional elemental damage type when used.
	El1Pct                  string `csv:"El1Pct"`                                   // Controls the random percent chance (out of 100) that the monster will append the element damage to the attack. This field is used when El#Mode is not null.
	El1MinD                 string `csv:"El1MinD"`                                  // The minimum element damage applied to the attack. This field is used when El#Mode is not null.
	El1MaxD                 string `csv:"El1MaxD"`                                  // The maximum element damage applied to the attack. This field is used when El#Mode is not null.
	El1Dur                  string `csv:"El1Dur"`                                   // Controls the duration of the related element mode in frame lengths (25 Frames = 1 Second). This is only applicable for the Cold, Poison, Stun, Burning, Freeze elements. This field is used when El#Mode is not null.
	El2Mode                 string `csv:"El2Mode"`                                  // Determines which animation mode will trigger an additional elemental damage type when used.
	El2Pct                  string `csv:"El2Pct"`                                   // Controls the random percent chance (out of 100) that the monster will append the element damage to the attack. This field is used when El#Mode is not null.
	El2MinD                 string `csv:"El2MinD"`                                  // The minimum element damage applied to the attack. This field is used when El#Mode is not null.
	El2MaxD                 string `csv:"El2MaxD"`                                  // The maximum element damage applied to the attack. This field is used when El#Mode is not null.
	El2Dur                  string `csv:"El2Dur"`                                   // Controls the duration of the related element mode in frame lengths (25 Frames = 1 Second). This is only applicable for the Cold, Poison, Stun, Burning, Freeze elements. This field is used when El#Mode is not null.
	El3Mode                 string `csv:"El3Mode"`                                  // Determines which animation mode will trigger an additional elemental damage type when used.
	El3Pct                  string `csv:"El3Pct"`                                   // Controls the random percent chance (out of 100) that the monster will append the element damage to the attack. This field is used when El#Mode is not null.
	El3MinD                 string `csv:"El3MinD"`                                  // The minimum element damage applied to the attack. This field is used when El#Mode is not null.
	El3MaxD                 string `csv:"El3MaxD"`                                  // The maximum element damage applied to the attack. This field is used when El#Mode is not null.
	El3Dur                  string `csv:"El3Dur"`                                   // Controls the duration of the related element mode in frame lengths (25 Frames = 1 Second). This is only applicable for the Cold, Poison, Stun, Burning, Freeze elements. This field is used when El#Mode is not null.
	TreasureClass1          string `csv:"TreasureClass1"`                           // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass2          string `csv:"TreasureClass2"`                           // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass3          string `csv:"TreasureClass3"`                           // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass4          string `csv:"TreasureClass4"`                           // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass1Nightmare string `csv:"TreasureClass1(N)"`                        // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass2Nightmare string `csv:"TreasureClass2(N)"`                        // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass3Nightmare string `csv:"TreasureClass3(N)"`                        // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass4Nightmare string `csv:"TreasureClass4(N)"`                        // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass1Hell      string `csv:"TreasureClass1(H)"`                        // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass2Hell      string `csv:"TreasureClass2(H)"`                        // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass3Hell      string `csv:"TreasureClass3(H)"`                        // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TreasureClass4Hell      string `csv:"TreasureClass4(H)"`                        // Defines which Treasure Class is used by the monster when it is killed. Points to the "Treasure Class" field from the TreasureClassEx.txt file. Used for normal monster types.
	TCQuestId               string `csv:"TCQuestId"`                                // Checks to see if the player has a quest flag progress. If not, then use the "TreasureClass4" field, based on the game's current difficulty.
	RespecFromAkara         bool   `csv:"Respecialization from Akara is Completed"` // Indicates if the respecialization from Akara is completed.
	TCQuestCP               string `csv:"TCQuestCP"`                                // Controls which Quest Checkpoint, or current progress within a quest (based on the "TCQuestId" value), is needed to use the "TreasureClass4" field, based on the game's current difficulty.
}

// MonsterStats2 represents additional functionalities and statistics for a monster in the game (continuation of MonsterStats1).
type MonsterStats2 struct {
	Id                  string `csv:"Id"`              // Controls the unique name ID to define the monster. This must match the same value in the monstats.txt file.
	Height              int    `csv:"Height"`          // Determines the height of the monster.
	Code1               string `csv:"Code1"`           // Code for the first purpose of Height.
	Code1Desc           string `csv:"Description1"`    // Description for the first purpose of Height.
	Code2               string `csv:"Code2"`           // Code for the second purpose of Height.
	Code2Desc           string `csv:"Description2"`    // Description for the second purpose of Height.
	OverlayHeight       int    `csv:"OverlayHeight"`   // Determines the height value of overlays for the monster.
	PixHeight           int    `csv:"pixHeight"`       // Determines the pixel height value for the damage bar when the monster is selected.
	SizeX               int    `csv:"SizeX"`           // Determines the tile grid size of the monster for handling placement when the monster spawns or uses movement skills.
	SizeY               int    `csv:"SizeY"`           // Determines the tile grid size of the monster for handling placement when the monster spawns or uses movement skills.
	SpawnColCode        int    `csv:"spawnCol"`        // Controls the method for spawning the monster based on the collisions in the environment.
	SpawnColDesc        string `csv:"Description3"`    // Description for the SpawnColCode.
	MeleeRng            int    `csv:"MeleeRng"`        // Controls the range of the monster's melee attack.
	BaseWeapon          string `csv:"BaseW"`           // Defines the monster's base weapon class.
	BaseWeaponDesc      string `csv:"Description4"`    // Description for the BaseWeapon.
	HitClass            int    `csv:"HitClass"`        // Defines the specific class of an attack when the monster successfully hits with an attack.
	HitClassDesc        string `csv:"Description5"`    // Description for the HitClass.
	HDv                 string `csv:"HDv"`             // Head visual.
	TRv                 string `csv:"TRv"`             // Torso visual.
	LGv                 string `csv:"LGv"`             // Legs visual.
	RAv                 string `csv:"RAv"`             // Right Arm visual.
	LAv                 string `csv:"LAv"`             // Left Arm visual.
	RHv                 string `csv:"RHv"`             // Right Hand visual.
	LHv                 string `csv:"LHv"`             // Left Hand visual.
	SHv                 string `csv:"SHv"`             // Shield visual.
	S1v                 string `csv:"S1v"`             // Special 1 visual.
	S2v                 string `csv:"S2v"`             // Special 2 visual.
	S3v                 string `csv:"S3v"`             // Special 3 visual.
	S4v                 string `csv:"S4v"`             // Special 4 visual.
	S5v                 string `csv:"S5v"`             // Special 5 visual.
	S6v                 string `csv:"S6v"`             // Special 6 visual.
	S7v                 string `csv:"S7v"`             // Special 7 visual.
	S8v                 string `csv:"S8v"`             // Special 8 visual.
	HD                  bool   `csv:"HD"`              // Head enabled.
	TR                  bool   `csv:"TR"`              // Torso enabled.
	LG                  bool   `csv:"LG"`              // Legs enabled.
	RA                  bool   `csv:"RA"`              // Right Arm enabled.
	LA                  bool   `csv:"LA"`              // Left Arm enabled.
	RH                  bool   `csv:"RH"`              // Right Hand enabled.
	LH                  bool   `csv:"LH"`              // Left Hand enabled.
	SH                  bool   `csv:"SH"`              // Shield enabled.
	S1                  bool   `csv:"S1"`              // Special 1 enabled.
	S2                  bool   `csv:"S2"`              // Special 2 enabled.
	S3                  bool   `csv:"S3"`              // Special 3 enabled.
	S4                  bool   `csv:"S4"`              // Special 4 enabled.
	S5                  bool   `csv:"S5"`              // Special 5 enabled.
	S6                  bool   `csv:"S6"`              // Special 6 enabled.
	S7                  bool   `csv:"S7"`              // Special 7 enabled.
	S8                  bool   `csv:"S8"`              // Special 8 enabled.
	TotalPieces         int    `csv:"TotalPieces"`     // Defines the total amount of component pieces that the monster uses.
	DeathMode           bool   `csv:"mDT"`             // If equals 1, then enable the Death Mode for the monster.
	NeutralMode         bool   `csv:"mNU"`             // If equals 1, then enable the Neutral Mode for the monster.
	WalkMode            bool   `csv:"mWL"`             // If equals 1, then enable the Walk Mode for the monster.
	GetHitMode          bool   `csv:"mGH"`             // If equals 1, then enable the Get Hit Mode for the monster.
	Attack1Mode         bool   `csv:"mA1"`             // If equals 1, then enable the Attack 1 Mode for the monster.
	Attack2Mode         bool   `csv:"mA2"`             // If equals 1, then enable the Attack 2 Mode for the monster.
	BlockMode           bool   `csv:"mBL"`             // If equals 1, then enable the Block Mode for the monster.
	CastMode            bool   `csv:"mSC"`             // If equals 1, then enable the Cast Mode for the monster.
	Skill1Mode          bool   `csv:"mS1"`             // If equals 1, then enable the Skill 1 Mode for the monster.
	Skill2Mode          bool   `csv:"mS2"`             // If equals 1, then enable the Skill 2 Mode for the monster.
	Skill3Mode          bool   `csv:"mS3"`             // If equals 1, then enable the Skill 3 Mode for the monster.
	Skill4Mode          bool   `csv:"mS4"`             // If equals 1, then enable the Skill 4 Mode for the monster.
	DeadMode            bool   `csv:"mDD"`             // If equals 1, then enable the Dead Mode for the monster.
	KnockbackMode       bool   `csv:"mKB"`             // If equals 1, then enable the Knockback Mode for the monster.
	SequenceMode        bool   `csv:"mSQ"`             // If equals 1, then enable the Sequence Mode for the monster.
	RunMode             bool   `csv:"mRN"`             // If equals 1, then enable the Run Mode for the monster.
	DeathDirections     int    `csv:"dDT"`             // Defines the number of directions that the monster can face during Death Mode.
	NeutralDirections   int    `csv:"dNU"`             // Defines the number of directions that the monster can face during Neutral Mode.
	WalkDirections      int    `csv:"dWL"`             // Defines the number of directions that the monster can face during Walk Mode.
	GetHitDirections    int    `csv:"dGH"`             // Defines the number of directions that the monster can face during Get Hit Mode.
	Attack1Directions   int    `csv:"dA1"`             // Defines the number of directions that the monster can face during Attack 1 Mode.
	Attack2Directions   int    `csv:"dA2"`             // Defines the number of directions that the monster can face during Attack 2 Mode.
	BlockDirections     int    `csv:"dBL"`             // Defines the number of directions that the monster can face during Block Mode.
	CastDirections      int    `csv:"dSC"`             // Defines the number of directions that the monster can face during Cast Mode.
	Skill1Directions    int    `csv:"dS1"`             // Defines the number of directions that the monster can face during Skill 1 Mode.
	Skill2Directions    int    `csv:"dS2"`             // Defines the number of directions that the monster can face during Skill 2 Mode.
	Skill3Directions    int    `csv:"dS3"`             // Defines the number of directions that the monster can face during Skill 3 Mode.
	Skill4Directions    int    `csv:"dS4"`             // Defines the number of directions that the monster can face during Skill 4 Mode.
	DeadDirections      int    `csv:"dDD"`             // Defines the number of directions that the monster can face during Dead Mode.
	KnockbackDirections int    `csv:"dKB"`             // Defines the number of directions that the monster can face during Knockback Mode.
	SequenceDirections  int    `csv:"dSQ"`             // Defines the number of directions that the monster can face during Sequence Mode.
	RunDirections       int    `csv:"dRN"`             // Defines the number of directions that the monster can face during Run Mode.
	Attack1Moving       bool   `csv:"A1mv"`            // If equals 1, then enable the Attack 1 Mode while the monster is moving with the Walk mode or Run mode.
	Attack2Moving       bool   `csv:"A2mv"`            // If equals 1, then enable the Attack 2 Mode while the monster is moving with the Walk mode or Run mode.
	CastMoving          bool   `csv:"SCmv"`            // If equals 1, then enable the Cast Mode while the monster is moving with the Walk mode or Run mode.
	Skill1Moving        bool   `csv:"S1mv"`            // If equals 1, then enable the Skill 1 Mode while the monster is moving with the Walk mode or Run mode.
	Skill2Moving        bool   `csv:"S2mv"`            // If equals 1, then enable the Skill 2 Mode while the monster is moving with the Walk mode or Run mode.
	Skill3Moving        bool   `csv:"S3mv"`            // If equals 1, then enable the Skill 3 Mode while the monster is moving with the Walk mode or Run mode.
	Skill4Moving        bool   `csv:"S4mv"`            // If equals 1, then enable the Skill 4 Mode while the monster is moving with the Walk mode or Run mode.
	NoGfxHitTest        bool   `csv:"noGfxHitTest"`    // If equals 1, then enable the mouse selection bounding box functionality around the monster.
	HtTop               int    `csv:"htTop"`           // Define the pixel top offset around the monster for the mouse selection bounding box functionality.
	HtLeft              int    `csv:"htLeft"`          // Define the pixel left offset around the monster for the mouse selection bounding box functionality.
	HtWidth             int    `csv:"htWidth"`         // Define the pixel right offset around the monster for the mouse selection bounding box functionality.
	HtHeight            int    `csv:"htHeight"`        // Define the pixel bottom offset around the monster for the mouse selection bounding box functionality.
	Restore             int    `csv:"restore"`         // Determines if the monster should be placed on the inactive list, to be saved when the level unloads.
	AutomapCel          int    `csv:"automapCel"`      // Controls what index of the Automap tiles to use to display this monster on the Automap.
	NoMap               bool   `csv:"noMap"`           // If equals 1, then the monster will not appear on the Automap.
	NoOvly              bool   `csv:"noOvly"`          // If equals 1, then no looping overlays will be drawn on the monster.
	IsSelectable        bool   `csv:"isSel"`           // If equals 1, then the monster is selectable and can be targeted.
	AlwaysSelectable    bool   `csv:"alSel"`           // If equals 1, then the player can always select the monster, regardless of being an ally or enemy.
	NeverSelectable     bool   `csv:"noSel"`           // If equals 1, then the player can never select the monster.
	ShiftSelectable     bool   `csv:"shiftSel"`        // If equals 1, then the player can target this monster when holding the Shift key and clicking to use a skill.
	CorpseSelectable    bool   `csv:"corpseSel"`       // If equals 1, then the monster's corpse can be with the mouse cursor.
	IsAttackable        bool   `csv:"isAtt"`           // If equals 1, then the monster can be attacked.
	CanRevive           bool   `csv:"revive"`          // If equals 1, then the monster is allowed to be revived by the Necromancer Revive skill.
	LimitCorpses        bool   `csv:"limitCorpses"`    // If equals 1, then the monster's corpse will be placed into a pool with all other corpses with this field checked.
	IsCritter           bool   `csv:"critter"`         // If equals 1, then the monster will be flagged as a critter.
	IsSmallType         bool   `csv:"small"`           // If equals 1, then the monster will be classified as a small type.
	IsLargeType         bool   `csv:"large"`           // If equals 1, then the monster will be classified as a large type.
	IsSoftBodied        bool   `csv:"soft"`            // If equals 1, then the monster's corpse is classified as soft-bodied.
	IsInert             bool   `csv:"inert"`           // If equals 1, then the monster will never attack its enemies.
	HasObjectCollision  bool   `csv:"objCol"`          // If equals 1 and the monster class is "barricadedoor", "barricadedoor2", or "evilhut", then the monster will place an invisible object with collision.
	HasDeadCollision    bool   `csv:"deadCol"`         // If equals 1, then the monster's corpse will have collision with other units.
	IsUnflatDead        bool   `csv:"unflatDead"`      // If equals 1, then ignore the corpse draw order for rendering the sprite on top of others, while the monster is dead.
	HasShadow           bool   `csv:"Shadow"`          // If equals 1, then the monster will project a shadow on the ground.
	NoUniqueShift       bool   `csv:"noUniqueShift"`   // If equals 1 and the monster is a Unique monster, then the monster will not have random color palette transform shifts.
	UseComponentDeath   bool   `csv:"compositeDeath"`  // If equals 1, then the monster's Death Mode and Dead mode will make use of its component system.
	LocalBloodColor     int    `csv:"localBlood"`      // Controls the color of the monster's blood based on the region locale.
	Bleed               int    `csv:"Bleed"`           // Controls if the monster will create blood missiles.
	LightRadiusSize     int    `csv:"Light"`           // Controls the monster's minimum Light Radius size.
	LightColorR         int    `csv:"light-r"`         // Controls the red color value of the monster's Light Radius.
	LightColorG         int    `csv:"light-g"`         // Controls the green color value of the monster's Light Radius.
	LightColorB         int    `csv:"light-b"`         // Controls the blue color value of the monster's Light Radius.
	Utrans              int    `csv:"Utrans"`          // Modifies the color palette transform for the monster in Normal difficulty.
	UtransN             int    `csv:"Utrans(N)"`       // Modifies the color palette transform for the monster in Nightmare difficulty.
	UtransH             int    `csv:"Utrans(H)"`       // Modifies the color palette transform for the monster in Hell difficulty.
	UtransDesc          string `csv:"Description6"`    // Description for the Utrans codes.
	InfernoLength       int    `csv:"InfernoLen"`      // The frame length to hold the channel cast time of the inferno skill.
	InfernoAnim         int    `csv:"InfernoAnim"`     // The exact frame in the channel animation to loop back and start at again.
	InfernoRollback     int    `csv:"InfernoRollback"` // The exact frame in the channel animation to determine when to roll back to the "InfernoAnim" frame.
	ResurrectMode       string `csv:"ResurrectMode"`   // Controls which monster mode to set on the monster when it is resurrected.
	ResurrectModeDesc   string `csv:"Description7"`    // Description for the ResurrectMode codes.
	ResurrectSkill      string `csv:"ResurrectSkill"`  // Controls what skill should the monster use when it is resurrected.
	SpawnUniqueMod      string `csv:"SpawnUniqueMod"`  // Controls what unique modifier the monster should always spawn with.
}
