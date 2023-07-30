package models

// ItemRatioFunction represents the type of item ratio used for determining item quality.
type ItemRatioFunction string

// Constants for Item Ratio Functions.
const (
	ItemRatioUnique    ItemRatioFunction = "Unique"
	ItemRatioSet       ItemRatioFunction = "Set"
	ItemRatioRare      ItemRatioFunction = "Rare"
	ItemRatioMagic     ItemRatioFunction = "Magic"
	ItemRatioHiQuality ItemRatioFunction = "HiQuality"
	ItemRatioNormal    ItemRatioFunction = "Normal"
	ItemRatioLow       ItemRatioFunction = "Low"
)

// ItemRatio represents the quality of items when being spawned and the probability calculations.
type ItemRatio struct {
	Function      ItemRatioFunction `csv:"Function"`         // Reference field to define the item ratio name.
	Version       int               `csv:"Version"`          // Game version to use this item ratio (0 = Classic mode | 100 = Expansion mode).
	Uber          int               `csv:"Uber"`             // If 1, the item ratio applies to Exceptional or Elite Quality items; else it applies to Normal Quality items.
	ClassSpecific int               `csv:"Class Specific"`   // If 1, the item ratio applies to class-based items.
	Unique        int               `csv:"Unique"`           // Base value for calculating the Unique Quality chance. Higher value means rarer chance. (Calculated first)
	UniqueDivisor int               `csv:"UniqueDivisor"`    // Modifier for changing the Unique Quality chance based on the difference between the Monster Level and the Item's base level.
	UniqueMin     int               `csv:"UniqueMin"`        // The minimum value of the probability denominator for Unique Quality. (Calculated in 128ths)
	Set           int               `csv:"Set"`              // Base value for calculating the Set Quality chance. Higher value means rarer chance. (Calculated after Unique)
	SetDivisor    int               `csv:"SetDivisor"`       // Modifier for changing the Set Quality chance based on the difference between the Monster Level and the Item's base level.
	SetMin        int               `csv:"SetMin"`           // The minimum value of the probability denominator for Set Quality. (Calculated in 128ths)
	Rare          int               `csv:"Rare"`             // Base value for calculating the Rare Quality chance. Higher value means rarer chance. (Calculated after Set)
	RareDivisor   int               `csv:"RareDivisor"`      // Modifier for changing the Rare Quality chance based on the difference between the Monster Level and the Item's base level.
	RareMin       int               `csv:"RareMin"`          // The minimum value of the probability denominator for Rare Quality. (Calculated in 128ths)
	Magic         int               `csv:"Magic"`            // Base value for calculating the Magic Quality chance. Higher value means rarer chance. (Calculated after Rare)
	MagicDivisor  int               `csv:"MagicDivisor"`     // Modifier for changing the Magic Quality chance based on the difference between the Monster Level and the Item's base level.
	MagicMin      int               `csv:"MagicMin"`         // The minimum value of the probability denominator for Magic Quality. (Calculated in 128ths)
	HiQuality     int               `csv:"HiQuality"`        // Base value for calculating the High Quality (Superior) chance. Higher value means rarer chance. (Calculated after Magic)
	HiQualityDiv  int               `csv:"HiQualityDivisor"` // Modifier for changing the High Quality (Superior) chance based on the difference between the Monster Level and the Item's base level.
	Normal        int               `csv:"Normal"`           // Base value for calculating the Normal Quality chance. Higher value means rarer chance. (Calculated after Normal, defaults to Low Quality)
	NormalDivisor int               `csv:"NormalDivisor"`    // Modifier for changing the Normal Quality chance based on the difference between the Monster Level and the Item's base level.
}

// Overview:
// This file determines the quality of items when being spawned. After the game determines what Item ItemSuperType should spawn,
// it then uses this file to calculate the quality of that item. These Item Quality checks are used for most item drops
// in the game such as monster drops and chest drops. The following files related to these calculations: ItemTypes.txt,
// weapons.txt, armor.txt, misc.txt, Uniqueitems.txt, SetItems.txt, monstats.txt, TreasureClassEx.txt

// Calculations:
// Item Quality is determined by first calculating the roll chance for a specific Quality. Then the game will attempt a
// random roll for that Item Quality. If that roll fails, then the game will calculate for the next lower Item Quality.
// The order of Item Quality calculations is the following: Unique > Set > Rare > Magic > High Quality (Superior) > Normal > Low Quality

// The first part of the calculations is to obtain the base probability for rolling the item quality with the following formula:
// Probability = ( Quality - ( mlvl - ilvl ) / Divisor ) * 128
// This probability value is a ratio divisor, meaning that there is a 1 in probability chance of the game choosing that Item Quality,
// so the lower the probability value, then the better the chance the item will successfully roll that Item Quality. The multiplication
// with 128 is for decreasing rounding errors. The Quality value is the “Unique”/“Set”/“Rare”/“Magic”/“HiQuality”/“Normal” Data Field.
// The mlvl value is the Monster Level, which depends on the current area level and game difficulty (this can also be the level for the chest drop).
// The ilvl value is the Base Item Level (See the “level” field in the weapons.txt/armor.txt/misc.txt or the “lvl” field in UniqueItems.txt/SetItems.txt).
// The Divisor value is the “UniqueDivisor”/“SetDivisor”/“RareDivisor”/“MagicDivisor”/“HiQualityDivisor”/“NormalDivisor” Data Field.

// The game will then obtain the character’s magic find bonus from items. If the current calculation is for Unique/Set/ Rare Quality items
// and the magic find item bonus exceeds 110%, then the character’s magic find will be modified with diminishing returns:
// MagicFind = 100 + MF * dim / (MF + dim)
// The MF value is the character’s magic find bonus percentage value plus the baseline default 100 value (See “item_magicbonus” in ItemStatCost.txt).
// The dim value is a special modifier for adding diminishing returns to the magic find bonus, which differs based on the Item Quality being calculated (Unique = 250, Set = 500, Rare = 600)

// After calculating the proper magic find value, the probability value is modified with the following formula:
// Probability = Probability * 100 / MagicFind
// This will reduce the Probability value, giving the Item Quality a higher chance to be successfully rolled.

// After calculating the baseline probability with the magic find bonus, the game will then compare this value with the minimum value for the Item Quality
// to cap it from reducing any further (See “UniqueMin”/“SetMin”/“RareMin”/“MagicMin” Data Fields). High Quality (Superior) and Normal Quality do not have a minimum value.

// The game will then modify the probability with the value from the related Treasure Class:
// Probability = Probability - Probability * TreasureClass / 1024)
// The TreasureClass value is a modifier for this Item Quality based on the Treasure Class being used (See the “Unique”/“Set”/“Rare”/“Magic” field from the TreasureClassEx.txt file)

// Finally, after calculating the overall value of the probability for the Item Quality, the game will then find a random number between 0 and the probability value.
// If that random value is between 0 and 128, then the item has successfully rolled that specific Item Quality. Otherwise, the calculations will move on to checking for the next lower Item Quality.
