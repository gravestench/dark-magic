package models

// ItemArmor represents an item in the game.
type ItemArmor struct {
	// Name is the reference field to define the item.
	Name string `csv:"name"`

	// Version defines which game version to create this item (0 = Classic mode | 100 = Expansion mode).
	Version int `csv:"version"`

	// CompactSave is a boolean field. If equals 1, then only the item's base stats will be stored in the character save,
	// but not any modifiers or additional stats. If equals 0, then all of the item's stats will be saved.
	CompactSave int `csv:"compactsave"`

	// Rarity determines the chance that the item will randomly spawn (1/#).
	// The higher the value, the rarer the item will be.
	// This field depends on the "spawnable" field being enabled, the "quest" field being disabled,
	// and the item level being less than or equal to the area level.
	// This value is also affected by the relative Act number that the item is dropping in,
	// where the higher the Act number, the more common the item will drop.
	Rarity int `csv:"rarity"`

	// Spawnable is a boolean field. If equals 1, then this item can be randomly spawned.
	// If equals 0, then this item will never randomly spawn.
	Spawnable int `csv:"spawnable"`

	// MinAC is the minimum amount of Defense that an armor item type can have.
	MinAC int `csv:"minac"`

	// MaxAC is the maximum amount of Defense that an armor item type can have.
	MaxAC int `csv:"maxac"`

	// Absorbs specifies the amount of damage that an item can absorb.
	Absorbs int `csv:"absorbs"`

	// Speed affects the Walk/Run Speed reduction when wearing the armor.
	Speed int `csv:"speed"`

	// ReqStr defines the amount of the Strength attribute needed to use the item.
	ReqStr int `csv:"reqstr"`

	// Block controls the block percent chance that the item provides (out of 100, but caps at 75).
	Block int `csv:"block"`

	// Durability defines the base durability amount that the item will spawn with.
	Durability int `csv:"durability"`

	// NoDurability is a boolean field. If equals 1, then the item will not have durability.
	// If equals 0, then the item will have durability.
	NoDurability int `csv:"nodurability"`

	// Level controls the base item level. This is used for determining when the item is allowed to drop,
	// such as making sure that the item level is not greater than the monster's level or the area level.
	Level int `csv:"level"`

	// LevelReq controls the player level requirement for being able to use the item.
	LevelReq int `csv:"levelreq"`

	// Cost defines the base gold cost of the item when being sold by an NPC.
	// This can be affected by item modifiers and the rarity of the item.
	Cost int `csv:"cost"`

	// GambleCost defines the gambling gold cost of the item on the Gambling UI.
	GambleCost int `csv:"gamble cost"`

	// Code defines a unique 3 letter/number code for the item. This is used as an identifier to reference the item.
	Code string `csv:"code"`

	// NameStr is the String Key that is used for the base item name.
	NameStr string `csv:"namestr"`

	// MagicLvl defines the magic level of the item, which can affect how magical item modifiers that can appear on the item.
	MagicLvl int `csv:"magic lvl"`

	// AutoPrefix is a boolean field. If equals 1, then the item automatically picks an item affix name
	// from a designated "group" value from the automagic.txt file, instead of using random prefixes.
	// This is only used when the item is of Magical quality.
	AutoPrefix int `csv:"auto prefix"`

	// AlternateGfx uses a unique 3 letter/number code similar to the defined "Code" fields
	// to determine what in-game graphics to display on the player character when the item is equipped.
	AlternateGfx string `csv:"alternategfx"`

	// OpenBetaGfx is a field for graphics used during the open beta version of the game.
	OpenBetaGfx string `csv:"OpenBetaGfx"`

	// NormCode links to a "Code" field to determine the normal version of the item.
	NormCode string `csv:"normcode"`

	// UberCode links to a "Code" field to determine the Exceptional version of the item.
	UberCode string `csv:"ubercode"`

	// UltraCode links to a "Code" field to determine the Elite version of the item.
	UltraCode string `csv:"ultracode"`

	// SpellOffset controls the offset index for reading spell data related to the item from the spells.txt file.
	SpellOffset int `csv:"spelloffset"`

	// Component determines the layer of player animation when the item is equipped.
	// This uses a code referenced from the Composit.txt file.
	Component int `csv:"component"`

	// InvWidth and InvHeight define the width and height of grid cells that the item occupies in the player inventory.
	InvWidth  int `csv:"invwidth"`
	InvHeight int `csv:"invheight"`

	// HasInv is a boolean field. If equals 1, then the item will have its own inventory allowing for the capability to socket gems, runes, or jewels.
	// If equals 0, then the item will not have its own inventory.
	HasInv int `csv:"hasinv"`

	// GemSockets determines the number of sockets that the item can have, allowing for socketing gems or runes.
	GemSockets int `csv:"gemsockets"`

	// GemApplyType defines the type of gems that can be used with the item.
	GemApplyType int `csv:"gemapplytype"`

	// FlippyFile is the .dc6 file that contains the graphics of the item on the ground when it is dropped.
	FlippyFile string `csv:"flippyfile"`

	// InvFile is the .dc6 file that contains the graphics of the item in the player's inventory.
	InvFile string `csv:"invfile"`

	// UniqueInvFile is the .dc6 file that contains the graphics of a unique item in the player's inventory.
	UniqueInvFile string `csv:"uniqueinvfile"`

	// SetInvFile is the .dc6 file that contains the graphics of a set item in the player's inventory.
	SetInvFile string `csv:"setinvfile"`

	// RArm, LArm, Torso, and Legs fields are used to specify which inventory slot the item should appear on the player model.
	RArm  int `csv:"rArm"`
	LArm  int `csv:"lArm"`
	Torso int `csv:"Torso"`
	Legs  int `csv:"Legs"`

	// RSPad and LSPad fields are used for specifying padding for 3D rendering of the item's graphic.
	RSPad int `csv:"rSPad"`
	LSPad int `csv:"lSPad"`

	// Useable is a boolean field. If equals 1, then the item can be used by right-clicking it.
	// If equals 0, then the item cannot be used.
	Useable int `csv:"useable"`

	// Throwable is a boolean field. If equals 1, then the item can be thrown.
	// If equals 0, then the item cannot be thrown.
	Throwable int `csv:"throwable"`

	// Stackable is a boolean field. If equals 1, then the item can be stacked in the inventory.
	// If equals 0, then the item cannot be stacked.
	Stackable int `csv:"stackable"`

	// MinStack and MaxStack define the minimum and maximum number of items that can be stacked together.
	MinStack int `csv:"minstack"`
	MaxStack int `csv:"maxstack"`

	// Type and Type2 fields are used to categorize the item and determine what class of items it belongs to.
	Type  string `csv:"type"`
	Type2 string `csv:"type2"`

	// DropSound is the sound that is played when the item is dropped on the ground.
	DropSound string `csv:"dropsound"`

	// DropSfxFrame is the frame number in the flippy file where the item's sound effect is triggered.
	DropSfxFrame int `csv:"dropsfxframe"`

	// UseSound specifies the sound that is played when the item is used.
	UseSound string `csv:"usesound"`

	// Unique is a boolean field. If equals 1, then the item is unique and can only be found once in the game.
	// If equals 0, then the item is not unique and can be found multiple times.
	Unique int `csv:"unique"`

	// Transparent is a boolean field. If equals 1, then the item has transparency.
	// If equals 0, then the item does not have transparency.
	Transparent int `csv:"transparent"`

	// TransTbl defines the transform table used for certain item types to determine how they appear in-game.
	TransTbl int `csv:"transtbl"`

	// Quivered is a boolean field. If equals 1, then the item can be equipped in the quiver slot.
	// If equals 0, then the item cannot be equipped in the quiver slot.
	Quivered int `csv:"quivered"`

	// LightRadius defines the light radius of the item when equipped.
	LightRadius int `csv:"lightradius"`

	// Belt is a boolean field. If equals 1, then the item can be equipped in the belt slot.
	// If equals 0, then the item cannot be equipped in the belt slot.
	Belt int `csv:"belt"`

	// Quest is a boolean field. If equals 1, then the item is a quest item.
	// If equals 0, then the item is not a quest item.
	Quest int `csv:"quest"`

	// MissileType is used to determine the type of missile projectile the item uses when thrown or shot.
	MissileType string `csv:"missiletype"`

	// DurWarning is a boolean field. If equals 1, then the game will warn the player when the item's durability is low.
	// If equals 0, then there will be no warning for low durability.
	DurWarning int `csv:"durwarning"`

	// QntWarning is a boolean field. If equals 1, then the game will warn the player when the item's quantity is low.
	// If equals 0, then there will be no warning for low quantity.
	QntWarning int `csv:"qntwarning"`

	// MinDam and MaxDam define the minimum and maximum damage values for the item.
	MinDam int `csv:"mindam"`
	MaxDam int `csv:"maxdam"`

	// StrBonus and DexBonus specify the strength and dexterity bonuses provided by the item when equipped.
	StrBonus int `csv:"StrBonus"`
	DexBonus int `csv:"DexBonus"`

	// GemOffset determines the index of the gem information in the gems.txt file.
	GemOffset int `csv:"gemoffset"`

	// Bitfield1 is a field that contains various bit flags for different item properties.
	Bitfield1 int `csv:"bitfield1"`

	// The following fields contain minimum and maximum values for magical modifiers provided by different NPCs in the game.
	// Each NPC has its own min/max magical modifiers for different item properties.

	CharsiMin      int `csv:"CharsiMin"`
	CharsiMax      int `csv:"CharsiMax"`
	CharsiMagicMin int `csv:"CharsiMagicMin"`
	CharsiMagicMax int `csv:"CharsiMagicMax"`
	CharsiMagicLvl int `csv:"CharsiMagicLvl"`

	GheedMin      int `csv:"GheedMin"`
	GheedMax      int `csv:"GheedMax"`
	GheedMagicMin int `csv:"GheedMagicMin"`
	GheedMagicMax int `csv:"GheedMagicMax"`
	GheedMagicLvl int `csv:"GheedMagicLvl"`

	AkaraMin      int `csv:"AkaraMin"`
	AkaraMax      int `csv:"AkaraMax"`
	AkaraMagicMin int `csv:"AkaraMagicMin"`
	AkaraMagicMax int `csv:"AkaraMagicMax"`
	AkaraMagicLvl int `csv:"AkaraMagicLvl"`

	FaraMin      int `csv:"FaraMin"`
	FaraMax      int `csv:"FaraMax"`
	FaraMagicMin int `csv:"FaraMagicMin"`
	FaraMagicMax int `csv:"FaraMagicMax"`
	FaraMagicLvl int `csv:"FaraMagicLvl"`

	LysanderMin      int `csv:"LysanderMin"`
	LysanderMax      int `csv:"LysanderMax"`
	LysanderMagicMin int `csv:"LysanderMagicMin"`
	LysanderMagicMax int `csv:"LysanderMagicMax"`
	LysanderMagicLvl int `csv:"LysanderMagicLvl"`

	DrognanMin      int `csv:"DrognanMin"`
	DrognanMax      int `csv:"DrognanMax"`
	DrognanMagicMin int `csv:"DrognanMagicMin"`
	DrognanMagicMax int `csv:"DrognanMagicMax"`
	DrognanMagicLvl int `csv:"DrognanMagicLvl"`

	HraltiMin      int `csv:"HraltiMin"`
	HraltiMax      int `csv:"HraltiMax"`
	HraltiMagicMin int `csv:"HraltiMagicMin"`
	HraltiMagicMax int `csv:"HraltiMagicMax"`
	HratliMagicLvl int `csv:"HratliMagicLvl"`

	AlkorMin      int `csv:"AlkorMin"`
	AlkorMax      int `csv:"AlkorMax"`
	AlkorMagicMin int `csv:"AlkorMagicMin"`
	AlkorMagicMax int `csv:"AlkorMagicMax"`
	AlkorMagicLvl int `csv:"AlkorMagicLvl"`

	OrmusMin      int `csv:"OrmusMin"`
	OrmusMax      int `csv:"OrmusMax"`
	OrmusMagicMin int `csv:"OrmusMagicMin"`
	OrmusMagicMax int `csv:"OrmusMagicMax"`
	OrmusMagicLvl int `csv:"OrmusMagicLvl"`

	ElzixMin      int `csv:"ElzixMin"`
	ElzixMax      int `csv:"ElzixMax"`
	ElzixMagicMin int `csv:"ElzixMagicMin"`
	ElzixMagicMax int `csv:"ElzixMagicMax"`
	ElzixMagicLvl int `csv:"ElzixMagicLvl"`

	AshearaMin      int `csv:"AshearaMin"`
	AshearaMax      int `csv:"AshearaMax"`
	AshearaMagicMin int `csv:"AshearaMagicMin"`
	AshearaMagicMax int `csv:"AshearaMagicMax"`
	AshearaMagicLvl int `csv:"AshearaMagicLvl"`

	CainMin      int `csv:"CainMin"`
	CainMax      int `csv:"CainMax"`
	CainMagicMin int `csv:"CainMagicMin"`
	CainMagicMax int `csv:"CainMagicMax"`
	CainMagicLvl int `csv:"CainMagicLvl"`

	HalbuMin      int `csv:"HalbuMin"`
	HalbuMax      int `csv:"HalbuMax"`
	HalbuMagicMin int `csv:"HalbuMagicMin"`
	HalbuMagicMax int `csv:"HalbuMagicMax"`
	HalbuMagicLvl int `csv:"HalbuMagicLvl"`

	JamellaMin      int `csv:"JamellaMin"`
	JamellaMax      int `csv:"JamellaMax"`
	JamellaMagicMin int `csv:"JamellaMagicMin"`
	JamellaMagicMax int `csv:"JamellaMagicMax"`
	JamellaMagicLvl int `csv:"JamellaMagicLvl"`

	LarzukMin      int `csv:"LarzukMin"`
	LarzukMax      int `csv:"LarzukMax"`
	LarzukMagicMin int `csv:"LarzukMagicMin"`
	LarzukMagicMax int `csv:"LarzukMagicMax"`
	LarzukMagicLvl int `csv:"LarzukMagicLvl"`

	MalahMin      int `csv:"MalahMin"`
	MalahMax      int `csv:"MalahMax"`
	MalahMagicMin int `csv:"MalahMagicMin"`
	MalahMagicMax int `csv:"MalahMagicMax"`
	MalahMagicLvl int `csv:"MalahMagicLvl"`

	DrehyaMin      int `csv:"DrehyaMin"`
	DrehyaMax      int `csv:"DrehyaMax"`
	DrehyaMagicMin int `csv:"DrehyaMagicMin"`
	DrehyaMagicMax int `csv:"DrehyaMagicMax"`
	DrehyaMagicLvl int `csv:"DrehyaMagicLvl"`

	// SourceArt and GameArt specify the art that is used for the item in the source file and the game file, respectively.
	SourceArt string `csv:"Source Art"`
	GameArt   string `csv:"Game Art"`

	// Transform is a boolean field. If equals 1, then the item can transform or "morph" the player character
	// when equipped. If equals 0, then the item does not have transformation capabilities.
	Transform int `csv:"Transform"`

	// InvTrans is a boolean field. If equals 1, then the item can be moved in the inventory while it is still in the process of being identified.
	// If equals 0, then the item cannot be moved until the identification process is completed.
	InvTrans int `csv:"InvTrans"`

	// SkipName is a boolean field. If equals 1, then the item name will not be displayed when the item is dropped.
	// If equals 0, then the item name will be displayed when the item is dropped.
	SkipName int `csv:"SkipName"`

	// NightmareUpgrade is a boolean field. If equals 1, then the item can be upgraded to an exceptional version in Nightmare difficulty.
	// If equals 0, then the item cannot be upgraded in Nightmare difficulty.
	NightmareUpgrade string `csv:"NightmareUpgrade"`

	// HellUpgrade is a boolean field. If equals 1, then the item can be upgraded to an elite version in Hell difficulty.
	// If equals 0, then the item cannot be upgraded in Hell difficulty.
	HellUpgrade string `csv:"HellUpgrade"`

	// Nameable is a boolean field. If equals 1, then the item can be named.
	// If equals 0, then the item cannot be named.
	Nameable int `csv:"nameable"`
}
