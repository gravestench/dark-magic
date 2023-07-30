package models

// ItemHighQualityModifiers represents a group of item modifiers for High Quality (Superior) item types.
type ItemHighQualityModifiers struct {
	Mod1Code  string `csv:"mod1code"`  // First item property code (references Properties.txt).
	Mod1Param int    `csv:"mod1param"` // First item property parameter value.
	Mod1Min   int    `csv:"mod1min"`   // First item property min value.
	Mod1Max   int    `csv:"mod1max"`   // First item property max value.

	Mod2Code  string `csv:"mod2code"`  // Second item property code (references Properties.txt).
	Mod2Param int    `csv:"mod2param"` // Second item property parameter value.
	Mod2Min   int    `csv:"mod2min"`   // Second item property min value.
	Mod2Max   int    `csv:"mod2max"`   // Second item property max value.

	Armor   bool `csv:"armor"`   // If true, the modifier can be applied to torso armor and helmet item types.
	Weapon  bool `csv:"weapon"`  // If true, the modifier can be applied to melee weapon item types (except scepters, wands, and staves).
	Shield  bool `csv:"shield"`  // If true, the modifier can be applied to shield item types.
	Scepter bool `csv:"scepter"` // If true, the modifier can be applied to scepter item types.
	Wand    bool `csv:"wand"`    // If true, the modifier can be applied to wand item types.
	Staff   bool `csv:"staff"`   // If true, the modifier can be applied to staff item types.
	Bow     bool `csv:"bow"`     // If true, the modifier can be applied to bow or crossbow item types.
	Boots   bool `csv:"boots"`   // If true, the modifier can be applied to boots item types.
	Gloves  bool `csv:"gloves"`  // If true, the modifier can be applied to gloves item types.
	Belt    bool `csv:"belt"`    // If true, the modifier can be applied to belt item types.
}
