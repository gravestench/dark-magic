package models

// GemData represents the data fields for gems and runes in the game.
type GemData struct {
	Name          string `csv:"name"`           // This is a reference field to define the gem/rune name
	Letter        string `csv:"letter"`         // Defines the string that appears in the item tooltip when a rune is socketed into an item
	Transform     string `csv:"transform"`      // Controls the color change of the item after being socketed by the gem/rune (Uses Color Codes from colors.txt)
	TransformCode int    `csv:"transform_code"` // Defines the code for the color transformation (0 to 20)
	Code          string `csv:"code"`           // Defines the unique item code used to create the gem/rune

	// Weapon Modifiers when socketed into gemapplytype == 0
	WeaponMod1Code  string `csv:"weapon_mod1_code"`  // Controls the item properties that the gem/rune provides when socketed into an item with a "gemapplytype" value that equals 0 (Uses the "code" field from Properties.txt)
	WeaponMod1Param string `csv:"weapon_mod1_param"` // The stat's "parameter" value associated with the listed property (WeaponMod1Code). Usage depends on the property function (See the "func" field on Properties.txt)
	WeaponMod1Min   int    `csv:"weapon_mod1_min"`   // The stat's "min" value associated with the listed property (WeaponMod1Code). Usage depends on the property function (See the "func" field on Properties.txt)
	WeaponMod1Max   int    `csv:"weapon_mod1_max"`   // The stat's "max" value to assign to the listed property (WeaponMod1Code). Usage depends on the property function (See the "func" field on Properties.txt)

	WeaponMod2Code  string `csv:"weapon_mod2_code"` // Repeat for WeaponMod2 (2nd weapon modifier)
	WeaponMod2Param string `csv:"weapon_mod2_param"`
	WeaponMod2Min   int    `csv:"weapon_mod2_min"`
	WeaponMod2Max   int    `csv:"weapon_mod2_max"`

	WeaponMod3Code  string `csv:"weapon_mod3_code"` // Repeat for WeaponMod3 (3rd weapon modifier)
	WeaponMod3Param string `csv:"weapon_mod3_param"`
	WeaponMod3Min   int    `csv:"weapon_mod3_min"`
	WeaponMod3Max   int    `csv:"weapon_mod3_max"`

	// Helmet Modifiers when socketed into gemapplytype == 1
	HelmMod1Code  string `csv:"helm_mod1_code"`  // Controls the item properties that the gem/rune provides when socketed into an item with a "gemapplytype" value that equals 1 (Uses the "code" field from Properties.txt)
	HelmMod1Param string `csv:"helm_mod1_param"` // The stat's "parameter" value associated with the listed property (HelmMod1Code). Usage depends on the property function (See the "func" field on Properties.txt)
	HelmMod1Min   int    `csv:"helm_mod1_min"`   // The stat's "min" value associated with the listed property (HelmMod1Code). Usage depends on the property function (See the "func" field on Properties.txt)
	HelmMod1Max   int    `csv:"helm_mod1_max"`   // The stat's "max" value to assign to the listed property (HelmMod1Code). Usage depends on the property function (See the "func" field on Properties.txt)

	HelmMod2Code  string `csv:"helm_mod2_code"` // Repeat for HelmMod2 (2nd helmet modifier)
	HelmMod2Param string `csv:"helm_mod2_param"`
	HelmMod2Min   int    `csv:"helm_mod2_min"`
	HelmMod2Max   int    `csv:"helm_mod2_max"`

	HelmMod3Code  string `csv:"helm_mod3_code"` // Repeat for HelmMod3 (3rd helmet modifier)
	HelmMod3Param string `csv:"helm_mod3_param"`
	HelmMod3Min   int    `csv:"helm_mod3_min"`
	HelmMod3Max   int    `csv:"helm_mod3_max"`

	// Shield Modifiers when socketed into gemapplytype == 2
	ShieldMod1Code  string `csv:"shield_mod1_code"`  // Controls the item properties that the gem/rune provides when socketed into an item with a "gemapplytype" value that equals 2 (Uses the "code" field from Properties.txt)
	ShieldMod1Param string `csv:"shield_mod1_param"` // The stat's "parameter" value associated with the listed property (ShieldMod1Code). Usage depends on the property function (See the "func" field on Properties.txt)
	ShieldMod1Min   int    `csv:"shield_mod1_min"`   // The stat's "min" value associated with the listed property (ShieldMod1Code). Usage depends on the property function (See the "func" field on Properties.txt)
	ShieldMod1Max   int    `csv:"shield_mod1_max"`   // The stat's "max" value to assign to the listed property (ShieldMod1Code). Usage depends on the property function (See the "func" field on Properties.txt)

	ShieldMod2Code  string `csv:"shield_mod2_code"` // Repeat for ShieldMod2 (2nd shield modifier)
	ShieldMod2Param string `csv:"shield_mod2_param"`
	ShieldMod2Min   int    `csv:"shield_mod2_min"`
	ShieldMod2Max   int    `csv:"shield_mod2_max"`

	ShieldMod3Code  string `csv:"shield_mod3_code"` // Repeat for ShieldMod3 (3rd shield modifier)
	ShieldMod3Param string `csv:"shield_mod3_param"`
	ShieldMod3Min   int    `csv:"shield_mod3_min"`
	ShieldMod3Max   int    `csv:"shield_mod3_max"`
}
