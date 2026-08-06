package models

type Hireling struct {
	Hireling                    string  `csv:"Hireling"`                 // Reference field to define the Hireling name
	Version                     int     `csv:"Version"`                  // Defines which game version to use this hireling (0 = Classic mode | 100 = Expansion mode)
	ID                          int     `csv:"Id"`                       // The unique identification number to define each hireling type
	Class                       string  `csv:"Class"`                    // Refers to the “hcIdx” field in MonStats.txt, defines the base type of unit to use for the hireling
	Act                         int     `csv:"Act"`                      // The Act that the hireling belongs to (values 1 to 5 equal Act 1 to Act 5, respectively)
	Difficulty                  int     `csv:"Difficulty"`               // The difficulty mode associated with the hireling (1 = Normal | 2 = Nightmare | 3 = Hell)
	Level                       int     `csv:"Level"`                    // The starting level of the unit
	Seller                      string  `csv:"Seller"`                   // Refers to the “hcIdx” field in MonStats.txt, defines the unit NPC that sells this hireling
	NameFirst                   string  `csv:"NameFirst"`                // String key for generating random hireling names (sequential range from “NameFirst” to “NameLast”, max name length is 48 characters)
	NameLast                    string  `csv:"NameLast"`                 // String key for generating random hireling names (sequential range from “NameFirst” to “NameLast”, max name length is 48 characters)
	Gold                        int     `csv:"Gold"`                     // Initial cost of the hireling, used in hire price calculation
	ExpPerLevel                 float64 `csv:"Exp/Lvl"`                  // Modifier for calculating the amount of Experience needed for the hireling’s next level
	HP                          int     `csv:"HP"`                       // Starting amount of Life at base Level
	HPPerLevel                  int     `csv:"HP/Lvl"`                   // Amount of Life gained per Level
	Defense                     int     `csv:"Defense"`                  // Starting amount of Defense at base Level
	DefensePerLevel             int     `csv:"Def/Lvl"`                  // Amount of Defense gained per Level
	Strength                    int     `csv:"Str"`                      // Starting amount of Strength at base Level
	StrengthPerLevel            int     `csv:"Str/Lvl"`                  // Amount of Strength gained per Level (Calculated in 8ths)
	Dexterity                   int     `csv:"Dex"`                      // Starting amount of Dexterity at base Level
	DexterityPerLevel           int     `csv:"Dex/Lvl"`                  // Amount of Dexterity gained per Level (Calculated in 8ths)
	AttackRating                int     `csv:"AR"`                       // Starting amount of Attack Rating at base Level
	AttackRatingPerLevel        int     `csv:"AR/Lvl"`                   // Amount of Attack Rating gained per Level
	MinPhysicalDamage           int     `csv:"Dmg-Min"`                  // Starting amount of minimum Physical Damage for attacks
	MaxPhysicalDamage           int     `csv:"Dmg-Max"`                  // Starting amount of maximum Physical Damage for attacks
	PhysicalDamagePerLevel      int     `csv:"Dmg/Lvl"`                  // Amount of Physical Damage gained per level (Calculated in 8ths)
	FireResistance              int     `csv:"ResistFire"`               // Starting amount of Fire Resistance at base Level
	FireResistancePerLevel      int     `csv:"ResistFire/Lvl"`           // Amount of Fire Resistance gained per Level (Calculated in 4ths)
	ColdResistance              int     `csv:"ResistCold"`               // Starting amount of Cold Resistance at base Level
	ColdResistancePerLevel      int     `csv:"ResistCold/Lvl"`           // Amount of Cold Resistance gained per Level (Calculated in 4ths)
	LightningResistance         int     `csv:"ResistLightning"`          // Starting amount of Lightning Resistance at base Level
	LightningResistancePerLevel int     `csv:"ResistLightning/Lvl"`      // Amount of Lightning Resistance gained per Level (Calculated in 4ths)
	PoisonResistance            int     `csv:"ResistPoison"`             // Starting amount of Poison Resistance at base Level
	PoisonResistancePerLevel    int     `csv:"ResistPoison/Lvl"`         // Amount of Poison Resistance gained per Level (Calculated in 4ths)
	HirelingDescription         string  `csv:"HireDesc"`                 // String key used to display the special description of the hireling in the hire UI window
	DefaultChance               int     `csv:"DefaultChance"`            // Chance for the hireling to attack with their weapon instead of using a Skill (denominator value for random roll)
	Skill1                      string  `csv:"Skill1"`                   // Points to a skill from the “skill” field in skills.txt file, gives the hireling the Skill to use (requires “Mode1”, “Chance1”, “ChancePerLvl1”)
	Skill2                      string  `csv:"Skill2"`                   // Points to a skill from the “skill” field in skills.txt file, gives the hireling the Skill to use (requires “Mode2”, “Chance2”, “ChancePerLvl2”)
	Skill3                      string  `csv:"Skill3"`                   // Points to a skill from the “skill” field in skills.txt file, gives the hireling the Skill to use (requires “Mode3”, “Chance3”, “ChancePerLvl3”)
	Skill4                      string  `csv:"Skill4"`                   // Points to a skill from the “skill” field in skills.txt file, gives the hireling the Skill to use (requires “Mode4”, “Chance4”, “ChancePerLvl4”)
	Skill5                      string  `csv:"Skill5"`                   // Points to a skill from the “skill” field in skills.txt file, gives the hireling the Skill to use (requires “Mode5”, “Chance5”, “ChancePerLvl5”)
	Skill6                      string  `csv:"Skill6"`                   // Points to a skill from the “skill” field in skills.txt file, gives the hireling the Skill to use (requires “Mode6”, “Chance6”, “ChancePerLvl6”)
	Mode1                       int     `csv:"Mode1"`                    // Uses a monster mode to determine the hireling’s behavior when using the related Skill (numeric ID of the monster mode, not the Token)
	Mode2                       int     `csv:"Mode2"`                    // Uses a monster mode to determine the hireling’s behavior when using the related Skill (numeric ID of the monster mode, not the Token)
	Mode3                       int     `csv:"Mode3"`                    // Uses a monster mode to determine the hireling’s behavior when using the related Skill (numeric ID of the monster mode, not the Token)
	Mode4                       int     `csv:"Mode4"`                    // Uses a monster mode to determine the hireling’s behavior when using the related Skill (numeric ID of the monster mode, not the Token)
	Mode5                       int     `csv:"Mode5"`                    // Uses a monster mode to determine the hireling’s behavior when using the related Skill (numeric ID of the monster mode, not the Token)
	Mode6                       int     `csv:"Mode6"`                    // Uses a monster mode to determine the hireling’s behavior when using the related Skill (numeric ID of the monster mode, not the Token)
	Chance1                     int     `csv:"Chance1"`                  // Base chance for the hireling to use Skill1, denominator value for random roll
	Chance2                     int     `csv:"Chance2"`                  // Base chance for the hireling to use Skill2, denominator value for random roll
	Chance3                     int     `csv:"Chance3"`                  // Base chance for the hireling to use Skill3, denominator value for random roll
	Chance4                     int     `csv:"Chance4"`                  // Base chance for the hireling to use Skill4, denominator value for random roll
	Chance5                     int     `csv:"Chance5"`                  // Base chance for the hireling to use Skill5, denominator value for random roll
	Chance6                     int     `csv:"Chance6"`                  // Base chance for the hireling to use Skill6, denominator value for random roll
	ChancePerLevel1             int     `csv:"ChancePerLvl1"`            // Chance for the hireling to use Skill1, affected by the difference in current Level and “Level” field
	ChancePerLevel2             int     `csv:"ChancePerLvl2"`            // Chance for the hireling to use Skill2, affected by the difference in current Level and “Level” field
	ChancePerLevel3             int     `csv:"ChancePerLvl3"`            // Chance for the hireling to use Skill3, affected by the difference in current Level and “Level” field
	ChancePerLevel4             int     `csv:"ChancePerLvl4"`            // Chance for the hireling to use Skill4, affected by the difference in current Level and “Level” field
	ChancePerLevel5             int     `csv:"ChancePerLvl5"`            // Chance for the hireling to use Skill5, affected by the difference in current Level and “Level” field
	ChancePerLevel6             int     `csv:"ChancePerLvl6"`            // Chance for the hireling to use Skill6, affected by the difference in current Level and “Level” field
	Skill1Level                 int     `csv:"Level1"`                   // Starting Level for Skill1
	Skill2Level                 int     `csv:"Level2"`                   // Starting Level for Skill2
	Skill3Level                 int     `csv:"Level3"`                   // Starting Level for Skill3
	Skill4Level                 int     `csv:"Level4"`                   // Starting Level for Skill4
	Skill5Level                 int     `csv:"Level5"`                   // Starting Level for Skill5
	Skill6Level                 int     `csv:"Level6"`                   // Starting Level for Skill6
	Skill1LevelPerLevel         int     `csv:"LvlPerLvl1"`               // Modifier to increase Skill1 level for every Level gained
	Skill2LevelPerLevel         int     `csv:"LvlPerLvl2"`               // Modifier to increase Skill2 level for every Level gained
	Skill3LevelPerLevel         int     `csv:"LvlPerLvl3"`               // Modifier to increase Skill3 level for every Level gained
	Skill4LevelPerLevel         int     `csv:"LvlPerLvl4"`               // Modifier to increase Skill4 level for every Level gained
	Skill5LevelPerLevel         int     `csv:"LvlPerLvl5"`               // Modifier to increase Skill5 level for every Level gained
	Skill6LevelPerLevel         int     `csv:"LvlPerLvl6"`               // Modifier to increase Skill6 level for every Level gained
	HiringMaxLevelDifference    int     `csv:"HiringMaxLevelDifference"` // Used to generate a range for the hireling's starting Level in the hiring UI window
	ResurrectCostMultiplier     int     `csv:"resurrectcostmultiplier"`  // Modifier used to calculate the hireling’s current resurrect cost
	ResurrectCostDivisor        int     `csv:"resurrectcostdivisor"`     // Modifier used to calculate the hireling’s current resurrect cost
	ResurrectCostMax            int     `csv:"resurrectcostmax"`         // Maximum Gold cost to resurrect this hireling
	EquivalentCharClass         string  `csv:"equivalentcharclass"`      // Determines what class this hireling is treated like under the hood for calculating skill level bonuses and gear restrictions
}

type HirelingDescription struct {
	ID             int  `csv:"id"`             // The id of the hireling monster class as defined in monstats.txt
	AlternateVoice bool `csv:"alternateVoice"` // Boolean field. If equals 1, then the hireling will use the alternate (feminine) voice type for voice lines. If equals 0, then it will use the masculine voice type.
}
