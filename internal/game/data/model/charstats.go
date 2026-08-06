package models

// CharStats holds the starting stats for each of the classes
type CharStats struct {
	Class               string `csv:"class"`
	Strength            int    `csv:"str"`
	Dexterity           int    `csv:"dex"`
	Intelligence        int    `csv:"int"`
	Vitality            int    `csv:"vit"`
	Total               int    `csv:"tot"`
	Stamina             int    `csv:"stamina"`
	HPAdd               int    `csv:"hpadd"`
	PercentStrength     int    `csv:"PercentStr"`
	PercentDexterity    int    `csv:"PercentDex"`
	PercentIntelligence int    `csv:"PercentInt"`
	PercentVitality     int    `csv:"PercentVit"`
	ManaRegen           int    `csv:"ManaRegen"`
	ToHitFactor         int    `csv:"ToHitFactor"`
	WalkVelocity        int    `csv:"WalkVelocity"`
	RunVelocity         int    `csv:"RunVelocity"`
	RunDrain            int    `csv:"RunDrain"`
	Comment             string `csv:"Comment"`
	LifePerLevel        int    `csv:"LifePerLevel"`
	StaminaPerLevel     int    `csv:"StaminaPerLevel"`
	ManaPerLevel        int    `csv:"ManaPerLevel"`
	LifePerVitality     int    `csv:"LifePerVitality"`
	StaminaPerVitality  int    `csv:"StaminaPerVitality"`
	ManaPerMagic        int    `csv:"ManaPerMagic"`
	StatPerLevel        int    `csv:"StatPerLevel"`
	Walk                int    `csv:"#walk"`
	Run                 int    `csv:"#run"`
	Swing               int    `csv:"#swing"`
	Spell               int    `csv:"#spell"`
	GetHit              int    `csv:"#gethit"`
	Bow                 int    `csv:"#bow"`
	BlockFactor         int    `csv:"BlockFactor"`
	StartSkill          string `csv:"StartSkill"`
	Skill1              string `csv:"Skill 1"`
	Skill2              string `csv:"Skill 2"`
	Skill3              string `csv:"Skill 3"`
	Skill4              string `csv:"Skill 4"`
	Skill5              string `csv:"Skill 5"`
	Skill6              string `csv:"Skill 6"`
	Skill7              string `csv:"Skill 7"`
	Skill8              string `csv:"Skill 8"`
	Skill9              string `csv:"Skill 9"`
	Skill10             string `csv:"Skill 10"`
	StrAllSkills        string `csv:"StrAllSkills"`
	StrSkillTab1        string `csv:"StrSkillTab1"`
	StrSkillTab2        string `csv:"StrSkillTab2"`
	StrSkillTab3        string `csv:"StrSkillTab3"`
	StrClassOnly        string `csv:"StrClassOnly"`
	BaseWClass          string `csv:"baseWClass"`
	Item1               string `csv:"item1"`
	Item1Loc            string `csv:"item1loc"`
	Item1Count          string `csv:"item1count"`
	Item2               string `csv:"item2"`
	Item2Loc            string `csv:"item2loc"`
	Item2Count          string `csv:"item2count"`
	Item3               string `csv:"item3"`
	Item3Loc            string `csv:"item3loc"`
	Item3Count          string `csv:"item3count"`
	Item4               string `csv:"item4"`
	Item4Loc            string `csv:"item4loc"`
	Item4Count          string `csv:"item4count"`
	Item5               string `csv:"item5"`
	Item5Loc            string `csv:"item5loc"`
	Item5Count          string `csv:"item5count"`
	Item6               string `csv:"item6"`
	Item6Loc            string `csv:"item6loc"`
	Item6Count          string `csv:"item6count"`
	Item7               string `csv:"item7"`
	Item7Loc            string `csv:"item7loc"`
	Item7Count          string `csv:"item7count"`
	Item8               string `csv:"item8"`
	Item8Loc            string `csv:"item8loc"`
	Item8Count          string `csv:"item8count"`
	Item9               string `csv:"item9"`
	Item9Loc            string `csv:"item9loc"`
	Item9Count          string `csv:"item9count"`
	Item10              string `csv:"item10"`
	Item10Loc           string `csv:"item10loc"`
	Item10Count         string `csv:"item10count"`
}

// this might be the d2 resurrected version...
// CharStartingAttributes holds the starting stats for each of the classes
//type CharStartingAttributes struct {
//	Class            string `csv:"class"`            // Character class
//	Str              int    `csv:"str"`              // Base strength
//	Dex              int    `csv:"dex"`              // Base dexterity
//	Intelligence     int    `csv:"int"`              // Base intelligence
//	Vitality         int    `csv:"vit"`              // Base vitality
//	Stamina          int    `csv:"stamina"`          // Base stamina
//	HP               int    `csv:"hp"`               // Base health points (HP)
//	Mana             int    `csv:"mana"`             // Base mana points
//	MaxHP            int    `csv:"maxhp"`            // Maximum health points (HP)
//	MaxMana          int    `csv:"maxmana"`          // Maximum mana points
//	Stat1            int    `csv:"stat1"`            // Unknown stat
//	Stat2            int    `csv:"stat2"`            // Unknown stat
//	Stat3            int    `csv:"stat3"`            // Unknown stat
//	Stat4            int    `csv:"stat4"`            // Unknown stat
//	Stat5            int    `csv:"stat5"`            // Unknown stat
//	Stat6            int    `csv:"stat6"`            // Unknown stat
//	MinHP            int    `csv:"minhp"`            // Minimum health points (HP)
//	MaxHPFactor      int    `csv:"maxhpfactor"`      // Health points factor for maximum health calculation
//	MinMana          int    `csv:"minmana"`          // Minimum mana points
//	MaxManaFactor    int    `csv:"maxmanafactor"`    // Mana points factor for maximum mana calculation
//	ManaRegen        int    `csv:"manaregen"`        // Mana regeneration rate
//	ToHitFactor      int    `csv:"tohitfactor"`      // To-hit factor
//	ToHitFactorN     int    `csv:"tohitfactorn"`     // To-hit factor (nightmare difficulty)
//	BlockFactor      int    `csv:"blockfactor"`      // Block factor
//	DamageFactor     int    `csv:"damagefactor"`     // Damage factor
//	DamageFactorN    int    `csv:"damagefactorn"`    // Damage factor (nightmare difficulty)
//	MinDam           int    `csv:"mindam"`           // Minimum damage
//	MaxDam           int    `csv:"maxdam"`           // Maximum damage
//	MinLDam          int    `csv:"minldam"`          // Minimum lightning damage
//	MaxLDam          int    `csv:"maxldam"`          // Maximum lightning damage
//	DamageBonus      int    `csv:"damagebonus"`      // Damage bonus
//	ToBlock          int    `csv:"toblock"`          // To-block chance
//	MinToBlock       int    `csv:"mintoblock"`       // Minimum to-block chance
//	ToBlockFactor    int    `csv:"toblockfactor"`    // To-block factor
//	MinToBlockFactor int    `csv:"mintoblockfactor"` // Minimum to-block factor
//	AC               int    `csv:"ac"`               // Armor class (AC)
//	MinAC            int    `csv:"minac"`            // Minimum armor class (AC)
//	ACFactor         int    `csv:"acfactor"`         // Armor class factor
//	MinACFactor      int    `csv:"minacfactor"`      // Minimum armor class factor
//	Parry            int    `csv:"parry"`            // Parry chance
//	ParryValue       int    `csv:"parryvalue"`       // Parry value
//	ArmorBytime      int    `csv:"armorbytime"`      // Unknown armor stat
//	ArmorPerLevel    int    `csv:"armorperlevel"`    // Armor increase per level
//	ArmorPerLevelN   int    `csv:"armorperleveln"`   // Armor increase per level (nightmare difficulty)
//	ArmorMax         int    `csv:"armormax"`         // Maximum armor
//	ArmorClass       int    `csv:"armorclass"`       // Armor class type
//	Walk             int    `csv:"walk"`             // Walking speed
//	Run              int    `csv:"run"`              // Running speed
//	Runspeed         int    `csv:"runspeed"`         // Run speed
//	Runrecover       int    `csv:"runrecover"`       // Run recover
//	AttackVel        int    `csv:"attackvel"`        // Attack velocity
//	Recover          int    `csv:"recover"`          // Recover
//	AttackRate       int    `csv:"attackrate"`       // Attack rate
//	MinDamage        int    `csv:"mindamage"`        // Minimum damage
//	MaxDamage        int    `csv:"maxdamage"`        // Maximum damage
//	MinLevDam        int    `csv:"minlevdam"`        // Minimum level damage
//	MaxLevDam        int    `csv:"maxlevdam"`        // Maximum level damage
//	NormAttackVel    int    `csv:"normattackvel"`    // Normal attack velocity
//}
