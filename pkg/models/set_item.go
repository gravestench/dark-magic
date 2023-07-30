package models

// SetItemData represents the data fields for each Set item in a Set.
type SetItemData struct {
	Index  string `csv:"index"`  // Links to a string key for displaying the Set item name
	Set    string `csv:"set"`    // Defines the Set to link to this Set Item (must match the "index" field from Sets.txt)
	Item   string `csv:"item"`   // Defines the baseline item code to use for this Set item (must match the "code" field value from weapons.txt, armor.txt, or misc.txt)
	Rarity string `csv:"rarity"` // Modifies the chances that this Unique item will spawn compared to the other Set items. This value acts as a numerator and a denominator. Each "rarity" value gets summed together to give a total denominator, used for the random roll for the item.

	// The item level for the item, which controls what object or monster needs to be in order to drop this item
	Lvl    int `csv:"lvl"`
	LvlReq int `csv:"lvl req"`

	// Controls the color change of the item when equipped on a character or dropped on the ground. If empty, then the item will have the default item color. (Uses Color Codes from the reference file colors.txt)
	ChrTransform string `csv:"chrtransform"`
	InvTransform string `csv:"invtransform"`

	// An override for the "invfile" field from the weapon.txt, armor.txt, or misc.txt files. By default, the Set Item will use what was defined by the baseline item from the "item" field.
	InvFile string `csv:"invfile"`

	// An override for the "flippyfile" field from the weapon.txt, armor.txt, or misc.txt files. By default, the Set Item will use what was defined by the baseline item from the "item" field.
	FlippyFile string `csv:"flippyfile"`

	// An override for the "dropsound" field from the weapon.txt, armor.txt, or misc.txt files. By default, the Set Item will use what was defined by the baseline item from the "item" field.
	DropSound string `csv:"dropsound"`

	// An override for the "dropsfxframe" field from the weapon.txt, armor.txt, or misc.txt files. By default, the Set Item will use what was defined by the baseline item from the "item" field.
	DropSfxFrame string `csv:"dropsfxframe"`

	// An override for the "usesound" field from the weapon.txt, armor.txt, or misc.txt files. By default, the Set Item will use what was defined by the baseline item from the "item" field.
	UseSound string `csv:"usesound"`

	CostMult float64 `csv:"cost mult"` // Multiplicative modifier for the Set item's buy, sell, and repair costs
	CostAdd  int     `csv:"cost add"`  // Flat integer modification to the Set item's buy, sell, and repair costs

	AddFunc int `csv:"add func"` // Controls how the additional Set item properties (aprop#a & aprop#b) will function on the Set item based on other related set items are equipped

	// Controls the item properties that are add baseline to the Set Item (Uses the "code" field from Properties.txt)
	Prop1 string `csv:"prop1"`
	Prop2 string `csv:"prop2"`
	Prop3 string `csv:"prop3"`
	Prop4 string `csv:"prop4"`
	Prop5 string `csv:"prop5"`
	Prop6 string `csv:"prop6"`
	Prop7 string `csv:"prop7"`
	Prop8 string `csv:"prop8"`
	Prop9 string `csv:"prop9"`

	// The stat's "parameter" value associated with the related property (prop#). Usage depends on the property function (See the "func" field on Properties.txt)
	Par1 string `csv:"par1"`
	Par2 string `csv:"par2"`
	Par3 string `csv:"par3"`
	Par4 string `csv:"par4"`
	Par5 string `csv:"par5"`
	Par6 string `csv:"par6"`
	Par7 string `csv:"par7"`
	Par8 string `csv:"par8"`
	Par9 string `csv:"par9"`

	// The stat's "min" value to assign to the related property (prop#). Usage depends on the property function (See the "func" field on Properties.txt)
	Min1 int `csv:"min1"`
	Min2 int `csv:"min2"`
	Min3 int `csv:"min3"`
	Min4 int `csv:"min4"`
	Min5 int `csv:"min5"`
	Min6 int `csv:"min6"`
	Min7 int `csv:"min7"`
	Min8 int `csv:"min8"`
	Min9 int `csv:"min9"`

	// The stat's "max" value to assign to the related property (prop#). Usage depends on the property function (See the "func" field on Properties.txt)
	Max1 int `csv:"max1"`
	Max2 int `csv:"max2"`
	Max3 int `csv:"max3"`
	Max4 int `csv:"max4"`
	Max5 int `csv:"max5"`
	Max6 int `csv:"max6"`
	Max7 int `csv:"max7"`
	Max8 int `csv:"max8"`
	Max9 int `csv:"max9"`

	// Controls the item properties that are added to the Set Item when other pieces of the Set are also equipped (Uses the "code" field from Properties.txt)
	AProp1a string `csv:"aprop1a"`
	AProp2a string `csv:"aprop2a"`
	AProp3a string `csv:"aprop3a"`
	AProp4a string `csv:"aprop4a"`
	AProp5a string `csv:"aprop5a"`

	// The stat's "parameter" value associated with the related property (aprop#a). Usage depends on the property function (See the "func" field on Properties.txt)
	APar1a string `csv:"apar1a"`
	APar2a string `csv:"apar2a"`
	APar3a string `csv:"apar3a"`
	APar4a string `csv:"apar4a"`
	APar5a string `csv:"apar5a"`

	// The stat's "min" value to assign to the related property (aprop#a). Usage depends on the property function (See the "func" field on Properties.txt)
	AMin1a int `csv:"amin1a"`
	AMin2a int `csv:"amin2a"`
	AMin3a int `csv:"amin3a"`
	AMin4a int `csv:"amin4a"`
	AMin5a int `csv:"amin5a"`

	// The stat's "max" value to assign to the related property (aprop#a). Usage depends on the property function (See the "func" field on Properties.txt)
	AMax1a int `csv:"amax1a"`
	AMax2a int `csv:"amax2a"`
	AMax3a int `csv:"amax3a"`
	AMax4a int `csv:"amax4a"`
	AMax5a int `csv:"amax5a"`

	// Controls the item properties that are added to the Set Item when other pieces of the Set are also equipped. Each of these numbered fields are paired with the related "aprop#a" field as an additional item property. (Uses the "code" field from Properties.txt)
	AProp1b string `csv:"aprop1b"`
	AProp2b string `csv:"aprop2b"`
	AProp3b string `csv:"aprop3b"`
	AProp4b string `csv:"aprop4b"`
	AProp5b string `csv:"aprop5b"`

	// The stat's "parameter" value associated with the related property (aprop#b). Usage depends on the property function (See the "func" field on Properties.txt)
	APar1b string `csv:"apar1b"`
	APar2b string `csv:"apar2b"`
	APar3b string `csv:"apar3b"`
	APar4b string `csv:"apar4b"`
	APar5b string `csv:"apar5b"`

	// The stat's "min" value to assign to the related property (aprop#b). Usage depends on the property function (See the "func" field on Properties.txt)
	AMin1b int `csv:"amin1b"`
	AMin2b int `csv:"amin2b"`
	AMin3b int `csv:"amin3b"`
	AMin4b int `csv:"amin4b"`
	AMin5b int `csv:"amin5b"`

	// The stat's "max" value to assign to the related property (aprop#b). Usage depends on the property function (See the "func" field on Properties.txt)
	AMax1b int `csv:"amax1b"`
	AMax2b int `csv:"amax2b"`
	AMax3b int `csv:"amax3b"`
	AMax4b int `csv:"amax4b"`
	AMax5b int `csv:"amax5b"`

	// The amount of weight added to the diablo clone progress when this item is sold. When offline, selling this item will instead immediately spawn diablo clone.
	DiabloCloneWeight int `csv:"diablocloneweight"`
}
