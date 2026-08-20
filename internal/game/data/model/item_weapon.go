package models

// ItemWeapon represents the functionalities for weapon type items.
type ItemWeapon struct {
	Name string `csv:"name"` // Reference field to define the item.
	// Game version to create this item (0 = Classic mode | 100 = Expansion mode).
	Version int `csv:"version"`
	// If 1, only the item's base stats will be stored in the character save, not any modifiers or additional stats. If 0,
	// all of the items stats will be saved.
	CompactSave int `csv:"compactsave"`
	// Determines the chance that the item will randomly spawn (1/#). Higher value means rarer item. Depends on spawnable,
	// quest, and item level being less than or equal to the area level. Affected by the relative Act number.
	Rarity int `csv:"rarity"`
	// If 1, this item can be randomly spawned. If 0, this item will never randomly spawn.
	Spawnable int `csv:"spawnable"`
	// For armor: Walk/Run Speed reduction when wearing the armor. For weapon: Attack Speed reduction when wearing the
	// weapon.
	Speed      int `csv:"speed"`
	ReqStr     int `csv:"reqstr"`     // Amount of Strength attribute needed to use the item.
	ReqDex     int `csv:"reqdex"`     // Amount of Dexterity attribute needed to use the item.
	Durability int `csv:"durability"` // Base durability amount that the item will spawn with.
	// If 1, the item will not have durability. If 0, the item will have durability.
	NoDurability int `csv:"nodurability"`
	// Base item level. Used for determining when the item is allowed to drop.
	Level int `csv:"level"`
	// If 1, display the item level next to the item name. If 0, ignore this.
	ShowLevel int `csv:"ShowLevel"`
	LevelReq  int `csv:"levelreq"` // Player level requirement for being able to use the item.
	// Base gold cost of the item when sold by an NPC. Affected by item modifiers and rarity.
	Cost       int `csv:"cost"`
	GambleCost int `csv:"gamble cost"` // Gambling gold cost of the item on the Gambling UI.
	// Unique 3 letter/number code for the item. Used as an identifier to reference the item.
	Code    string `csv:"code"`
	NameStr string `csv:"namestr"` // String Key used for the base item name.
	// Magic level of the item, affecting magical item modifiers that can appear on the item (See automagic.txt).
	MagicLvl int `csv:"magic lvl"`
	// Automatically picks an item affix name from a designated "group" value from the automagic.txt file, used when the
	// item is Magical quality.
	AutoPrefix int `csv:"auto prefix"`
	// Unique 3 letter/number code to determine what in-game graphics to display on the player character when the item is
	// equipped.
	AlternateGfx string `csv:"alternategfx"`
	// Links to a "code" field to determine the normal version of the item.
	NormCode string `csv:"normcode"`
	// Links to a "code" field to determine the Exceptional version of the item.
	UberCode string `csv:"ubercode"`
	// Links to a "code" field to determine the Elite version of the item.
	UltraCode string `csv:"ultracode"`
	// Determines the layer of player animation when the item is equipped, using a code referenced from the Composit.txt
	// file.
	Component string `csv:"component"`
	// Defines the width of grid cells that the item occupies in the player inventory.
	InvWidth int `csv:"invwidth"`
	// Defines the height of grid cells that the item occupies in the player inventory.
	InvHeight int `csv:"invheight"`
	// If 1, the item will have its own inventory, allowing for the capability to socket gems, runes, or jewels. If 0, the
	// item cannot have sockets.
	HasInv int `csv:"hasinv"`
	// Controls the maximum number of sockets allowed on this item, limited by the item's size based on "invwidth" and
	// "invheight".
	GemSockets int `csv:"gemsockets"`
	// Determines which effect from a gem or rune will be applied when socketed into this item (See gems.txt).
	GemApplyType string `csv:"gemapplytype"`
	// Controls which DC6 file to use for displaying the item in the game world when it is dropped on the ground (uses the
	// file name as the input).
	FlippyFile string `csv:"flippyfile"`
	// Controls which DC6 file to use for displaying the item graphics in the inventory (uses the file name as the input).
	InvFile string `csv:"invfile"`
	// Controls which DC6 file to use for displaying the item graphics in the inventory when it is a Unique quality item
	// (uses the file name as the input).
	UniqueInvFile string `csv:"uniqueinvfile"`
	// Controls which DC6 file to use for displaying the item graphics in the inventory when it is a Set quality item (uses
	// the file name as the input).
	SetInvFile string `csv:"setinvfile"`
	// If 1, the item can be used with the right-click mouse button command (works with specific belt items or quest
	// items). If 0, ignore this.
	Useable int `csv:"useable"`
	// If 1, the item will use a quantity field and handle stacking functionality. If 0, the item cannot be stacked.
	Stackable int `csv:"stackable"`
	// Controls the minimum stack count or quantity allowed on the item. Depends on "stackable" being enabled.
	MinStack int `csv:"minstack"`
	// Controls the maximum stack count or quantity allowed on the item. Depends on "stackable" being enabled.
	MaxStack int `csv:"maxstack"`
	// Controls the stack count or quantity that the item can spawn with. Depends on "stackable" being enabled.
	SpawnStack int `csv:"spawnstack"`
	// If 1, the item will use the transmogrify function. If 0, ignore this. Depends on "useable" being enabled.
	Transmogrify int `csv:"Transmogrify"`
	// Links to a "code" field to determine which item is chosen to transmogrify this item to.
	TMogType string `csv:"TMogType"`
	// Controls the minimum quantity that the transmogrify item will have. Depends on what item was chosen in the
	// "TMogType" field, and that the transmogrify item has quantity.
	TMogMin int `csv:"TMogMin"`
	// Controls the minimum quantity that the transmogrify item will have. Depends on what item was chosen in the
	// "TMogType" field, and that the transmogrify item has quantity.
	TMogMax int `csv:"TMogMax"`
	// Points to an Item Type defined in the ItemTypes.txt file, which controls how the item functions.
	Type string `csv:"type"`
	// Points to a secondary Item Type defined in the ItemTypes.txt file, which controls how the item functions. Optional,
	// adds more functionalities and possibilities with the item.
	Type2 string `csv:"type2"`
	// Points to a sound defined in the sounds.txt file. Used when the item is dropped on the ground.
	DropSound string `csv:"dropsound"`
	// Defines which frame in the "flippyfile" animation to play the "dropsound" sound when the item is dropped on the
	// ground.
	DropSfxFrame int `csv:"dropsfxframe"`
	// Points to a sound defined in the sounds.txt file. Used when the item is moved in the inventory or used.
	UseSound string `csv:"usesound"`
	// If 1, the item can only spawn as a Unique quality type. If 0, the item can spawn as other quality types.
	Unique int `csv:"unique"`
	// If 1, the item will be drawn transparent on the player model (similar to ethereal models). If 0, the item will
	// appear solid on the player model.
	Transparent int `csv:"transparent"`
	// Controls what type of transparency to use, based on the "transparent" field being enabled.
	TransTbl int `csv:"transtbl"`
	// Controls the value of the light radius that this item can apply on the monster. This only affects monsters with this
	// item equipped, not other types of units. Ignored if the item's component on the monster is "lit", "med", or "hvy".
	LightRadius int `csv:"lightradius"`
	// Controls which belt type to use for belt items only. Determines what index entry in the belts.txt file to use.
	Belt int `csv:"belt"`
	// Controls what quest class is tied to the item, enabling certain item functionalities for a specific quest. Any value
	// greater than 0 also means the item is flagged as a quest item, affecting how it is displayed, traded, its rarity,
	// and how it cannot be sold to an NPC. If 0, the item will not be flagged as a quest item.
	Quest int `csv:"quest"`
	// If 1 and the "quest" field is enabled, then the game will check the current difficulty setting and tie that
	// difficulty setting to the quest item. This means the player can have more than 1 of the same quest item as long as
	// they are obtained per difficulty mode (Normal / Nightmare / Hell). If 0 and the "quest" field is enabled, then the
	// player can only have 1 count of the quest item in the inventory, regardless of difficulty.
	QuestDiffCheck int `csv:"questdiffcheck"`
	// Points to the "Id" field from the Missiles.txt file, determining what type of missile is used when using the
	// throwing weapons.
	MissileType string `csv:"missiletype"`
	// Controls the threshold value for durability to display the low durability warning UI. Used if the item has
	// durability.
	DurWarning int `csv:"durwarning"`
	// Controls the threshold value for quantity to display the low quantity warning UI. Used if the item has stacks.
	QntWarning int `csv:"qntwarning"`
	MinDam     int `csv:"mindam"` // The minimum physical damage provided by the item.
	MaxDam     int `csv:"maxdam"` // The maximum physical damage provided by the item.
	// Percentage multiplier that gets multiplied by the player's current Strength attribute value to modify the bonus
	// damage percent from the equipped item. Default value is 100 if this equals 1.
	StrBonus int `csv:"StrBonus"`
	// Percentage multiplier that gets multiplied by the player's current Dexterity attribute value to modify the bonus
	// damage percent from the equipped item. Default value is 100 if this equals 1.
	DexBonus int `csv:"DexBonus"`
	// Determines the starting index offset for reading the gems.txt file when determining what effects gems or runes will
	// have on the item based on the "gemapplytype" field.
	GemOffset int `csv:"gemoffset"`
	// Controls different flags that can affect the item, using an integer value to check against different bit fields
	// using the "&" operator.
	BitField1 int `csv:"bitfield1"`
	// Controls the color palette change of the item for the character model graphics.
	Transform int `csv:"Transform"`
	// Controls the color palette change of the item for the inventory graphics.
	InvTrans int `csv:"InvTrans"`
	// If 1 and the item is Unique rarity, then skip adding the item's base name in its title. If 0, ignore this.
	SkipName int `csv:"SkipName"`
	// Links to another item's "code" field, determining which item will replace this item when generated in the NPC's
	// store while the game is playing in Nightmare difficulty. If this field's code equals "xxx", then this item will not
	// change in this difficulty.
	NightmareUpgrade string `csv:"NightmareUpgrade"`
	// Links to another item's "code" field, determining which item will replace this item when generated in the NPC's
	// store while the game is playing in Hell difficulty. If this field's code equals "xxx", then this item will not
	// change in this difficulty.
	HellUpgrade string `csv:"HellUpgrade"`
	// If 1, the item's name can be personalized by Anya for the Act 5 Betrayal of Harrogath quest reward. If 0, the item
	// cannot be used for the personalized name reward.
	Nameable int `csv:"Nameable"`
	// If 1, this item will always appear on the NPC's store. If 0, the item will randomly appear on the NPC's store when
	// appropriate.
	PermStoreItem int `csv:"PermStoreItem"`
	// The amount of weight added to the diablo clone progress when this item is sold. When offline, selling this item will
	// instead immediately spawn diablo clone.
	DiabloCloneWeight int `csv:"diablocloneweight"`
	// If 1, the item will be treated as a one-handed and two-handed weapon by the Barbarian class. If 0, the Barbarian can
	// only use this weapon as either one-handed or two-handed, but not both. (Exclusive to weapons.txt)
	OneOrTwoHanded int `csv:"1or2handed"`
	// If 1, the item will be treated as a two-handed weapon. If 0, the item will be treated as a one-handed weapon.
	// (Exclusive to weapons.txt)
	TwoHanded int `csv:"2handed"`
	// Defines the two-handed weapon class, controlling what character animations are used when the weapon is equipped.
	// (Exclusive to weapons.txt)
	TwoHandWClass string `csv:"2handedwclass"`
	// The minimum physical damage provided by the weapon if the item is two-handed. (Relies on the "2handed" field being
	// enabled)
	TwoHandMinDam int `csv:"2handmindam"`
	// The maximum physical damage provided by the weapon if the item is two-handed. (Relies on the "2handed" field being
	// enabled)
	TwoHandMaxDam int `csv:"2handmaxdam"`
	// Defines the hit class of the weapon, used to determine what SFX to use when the weapon hits an enemy.
	HitClass string `csv:"hitclass"`
	// The minimum physical damage provided by the item if it is a throwing weapon.
	MinMisDam int `csv:"minmisdam"`
	// The maximum physical damage provided by the item if it is a throwing weapon.
	MaxMisDam int `csv:"maxmisdam"`
	// Adds extra range in grid spaces for melee attacks while the melee weapon is equipped. The baseline melee range is 1,
	// and this field adds to that range.
	RangeAdder int `csv:"rangeadder"`
	// Defines the one-handed weapon class, controlling what character animations are used when the weapon is equipped.
	WeaponClass string `csv:"wclass"`
}
