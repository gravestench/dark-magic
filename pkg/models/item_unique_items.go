package models

// ItemUnique represents a unique item in the game.
type ItemUnique struct {
	Index             string `csv:"index"`             // Points to a string key value to use as the Unique item's name.
	Version           int    `csv:"version"`           // Defines which game version to create this item (<100 = Classic mode | 100 = Expansion mode).
	Enabled           bool   `csv:"enabled"`           // If equals 1, then this item can be rolled as a choice when randomly dropping a unique. If equals 0, then this item cannot be dropped randomly, but can still be dropped explicitly from a treasure class.
	FirstLadderSeason int    `csv:"firstLadderSeason"` // The first ladder season the unique item can be dropped or created on (inclusive). If blank or 0, then it is available in non-ladder.
	LastLadderSeason  int    `csv:"lastLadderSeason"`  // The last ladder season the unique item is ladder-only (inclusive). Must be used in conjunction with firstLadderSeason.
	Rarity            int    `csv:"rarity"`            // Modifies the chances that this Unique item will spawn compared to other Unique items. Each "rarity" value gets summed together to give a total denominator used for the random roll for the item.
	NoLimit           bool   `csv:"nolimit"`           // Requires the "quest" field from the misc.txt file to be enabled. If equals 1, then this item can be created and will automatically be identified. If equals 0, then ignore this.
	Lvl               int    `csv:"lvl"`               // The item level for the item, which controls what object or monster needs to be in order to drop this item.
	LvlReq            int    `csv:"lvl req"`           // The minimum character level required to equip the item.
	Code              string `csv:"code"`              // Defines the baseline item code to use for this Unique item (must match the "code" field value from weapons.txt, armor.txt, or misc.txt).
	Carry1            bool   `csv:"carry1"`            // If equals 1, then players can only carry one of these items in their inventory. If equals 0, then ignore this.
	CostMult          int    `csv:"cost mult"`         // Multiplicative modifier for the Unique item's buy, sell, and repair costs.
	CostAdd           int    `csv:"cost add"`          // Flat integer modification to the Unique item's buy, sell, and repair costs. Added after the "cost mult" has modified the costs.
	ChrTransform      string `csv:"chrtransform"`      // Controls the color change of the item when equipped on a character or dropped on the ground. If empty, then the item will have the default item color.
	InvTransform      string `csv:"invtransform"`      // Controls the color change of the item in the inventory UI. If empty, then the item will have the default item color.
	InvFile           string `csv:"invfile"`           // An override for the "invfile" field from the weapon.txt, armor.txt, or misc.txt files.
	FlippyFile        string `csv:"flippyfile"`        // An override for the "flippyfile" field from the weapon.txt, armor.txt, or misc.txt files.
	DropSound         string `csv:"dropsound"`         // An override for the "dropsound" field from the weapon.txt, armor.txt, or misc.txt files.
	DropSFXFrame      int    `csv:"dropsfxframe"`      // An override for the "dropsfxframe" field from the weapon.txt, armor.txt, or misc.txt files.
	UseSound          string `csv:"usesound"`          // An override for the "usesound" field from the weapon.txt, armor.txt, or misc.txt files.
	Prop1             string `csv:"prop1"`             // Controls the first item property for the Unique item.
	Par1              string `csv:"par1"`              // The stat's "parameter" value associated with the first property (prop1).
	Min1              int    `csv:"min1"`              // The stat's "min" value to assign to the first property (prop1).
	Max1              int    `csv:"max1"`              // The stat's "max" value to assign to the first property (prop1).
	Prop2             string `csv:"prop2"`             // Controls the second item property for the Unique item.
	Par2              string `csv:"par2"`              // The stat's "parameter" value associated with the second property (prop2).
	Min2              int    `csv:"min2"`              // The stat's "min" value to assign to the second property (prop2).
	Max2              int    `csv:"max2"`              // The stat's "max" value to assign to the second property (prop2).
	Prop3             string `csv:"prop3"`             // Controls the third item property for the Unique item.
	Par3              string `csv:"par3"`              // The stat's "parameter" value associated with the third property (prop3).
	Min3              int    `csv:"min3"`              // The stat's "min" value to assign to the third property (prop3).
	Max3              int    `csv:"max3"`              // The stat's "max" value to assign to the third property (prop3).
	Prop4             string `csv:"prop4"`             // Controls the fourth item property for the Unique item.
	Par4              string `csv:"par4"`              // The stat's "parameter" value associated with the fourth property (prop4).
	Min4              int    `csv:"min4"`              // The stat's "min" value to assign to the fourth property (prop4).
	Max4              int    `csv:"max4"`              // The stat's "max" value to assign to the fourth property (prop4).
	Prop5             string `csv:"prop5"`             // Controls the fifth item property for the Unique item.
	Par5              string `csv:"par5"`              // The stat's "parameter" value associated with the fifth property (prop5).
	Min5              int    `csv:"min5"`              // The stat's "min" value to assign to the fifth property (prop5).
	Max5              int    `csv:"max5"`              // The stat's "max" value to assign to the fifth property (prop5).
	Prop6             string `csv:"prop6"`             // Controls the sixth item property for the Unique item.
	Par6              string `csv:"par6"`              // The stat's "parameter" value associated with the sixth property (prop6).
	Min6              int    `csv:"min6"`              // The stat's "min" value to assign to the sixth property (prop6).
	Max6              int    `csv:"max6"`              // The stat's "max" value to assign to the sixth property (prop6).
	Prop7             string `csv:"prop7"`             // Controls the seventh item property for the Unique item.
	Par7              string `csv:"par7"`              // The stat's "parameter" value associated with the seventh property (prop7).
	Min7              int    `csv:"min7"`              // The stat's "min" value to assign to the seventh property (prop7).
	Max7              int    `csv:"max7"`              // The stat's "max" value to assign to the seventh property (prop7).
	Prop8             string `csv:"prop8"`             // Controls the eighth item property for the Unique item.
	Par8              string `csv:"par8"`              // The stat's "parameter" value associated with the eighth property (prop8).
	Min8              int    `csv:"min8"`              // The stat's "min" value to assign to the eighth property (prop8).
	Max8              int    `csv:"max8"`              // The stat's "max" value to assign to the eighth property (prop8).
	Prop9             string `csv:"prop9"`             // Controls the ninth item property for the Unique item.
	Par9              string `csv:"par9"`              // The stat's "parameter" value associated with the ninth property (prop9).
	Min9              int    `csv:"min9"`              // The stat's "min" value to assign to the ninth property (prop9).
	Max9              int    `csv:"max9"`              // The stat's "max" value to assign to the ninth property (prop9).
	Prop10            string `csv:"prop10"`            // Controls the tenth item property for the Unique item.
	Par10             string `csv:"par10"`             // The stat's "parameter" value associated with the tenth property (prop10).
	Min10             int    `csv:"min10"`             // The stat's "min" value to assign to the tenth property (prop10).
	Max10             int    `csv:"max10"`             // The stat's "max" value to assign to the tenth property (prop10).
	Prop11            string `csv:"prop11"`            // Controls the eleventh item property for the Unique item.
	Par11             string `csv:"par11"`             // The stat's "parameter" value associated with the eleventh property (prop11).
	Min11             int    `csv:"min11"`             // The stat's "min" value to assign to the eleventh property (prop11).
	Max11             int    `csv:"max11"`             // The stat's "max" value to assign to the eleventh property (prop11).
	Prop12            string `csv:"prop12"`            // Controls the twelfth item property for the Unique item.
	Par12             string `csv:"par12"`             // The stat's "parameter" value associated with the twelfth property (prop12).
	Min12             int    `csv:"min12"`             // The stat's "min" value to assign to the twelfth property (prop12).
	Max12             int    `csv:"max12"`             // The stat's "max" value to assign to the twelfth property (prop12).
	DiabloCloneWeight int    `csv:"diablocloneweight"` // The amount of weight added to the Diablo Clone progress when this item is sold.
}
