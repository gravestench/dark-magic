package models

// Difficultylevel represents the data structure for the difficultylevels.txt file.
type Difficultylevel struct {
	Name                                     string  `csv:"Name"`                                     // The difficulty mode name.
	ResistPenalty                            int     `csv:"ResistPenalty"`                            // Baseline starting point for player character's resistances for Expansion mode.
	ResistPenaltyNonExpansion                int     `csv:"ResistPenaltyNonExpansion"`                // Baseline starting point for player character's resistances for Non-Expansion mode.
	DeathExpPenalty                          float64 `csv:"DeathExpPenalty"`                          // Modifies the percentage of current level experience lost when a player character dies.
	MonsterSkillBonus                        int     `csv:"MonsterSkillBonus"`                        // Additional skill levels added to skills used by monsters (defined from monstats.txt).
	MonsterFreezeDivisor                     int     `csv:"MonsterFreezeDivisor"`                     // Divisor affecting all Freeze Length values on monsters.
	MonsterColdDivisor                       int     `csv:"MonsterColdDivisor"`                       // Divisor affecting all Cold Length values on monsters.
	AiCurseDivisor                           int     `csv:"AiCurseDivisor"`                           // Divisor affecting durations of Curses on monsters.
	LifeStealDivisor                         int     `csv:"LifeStealDivisor"`                         // Divisor affecting the amount of Life Steal that player characters gain.
	ManaStealDivisor                         int     `csv:"ManaStealDivisor"`                         // Divisor affecting the amount of Mana Steal that player characters gain.
	UniqueDamageBonus                        float64 `csv:"UniqueDamageBonus"`                        // Percentage modifier for a Unique monster's overall Damage and Attack Rating.
	ChampionDamageBonus                      float64 `csv:"ChampionDamageBonus"`                      // Percentage modifier for a Champion monster's overall Damage and Attack Rating.
	PlayerDamagePercentVSPlayer              float64 `csv:"PlayerDamagePercentVSPlayer"`              // Percentage modifier for the total damage a player deals to another player.
	PlayerDamagePercentVSMercenary           float64 `csv:"PlayerDamagePercentVSMercenary"`           // Percentage modifier for the total damage a player deals to another player's mercenary.
	PlayerDamagePercentVSPrimeEvil           float64 `csv:"PlayerDamagePercentVSPrimeEvil"`           // Percentage modifier for the total damage a player deals to a Prime Evil boss.
	PlayerHitReactBufferVSPlayer             int     `csv:"PlayerHitReactBufferVSPlayer"`             // Frame length for the amount of time a player cannot be placed into another hit react from a player.
	PlayerHitReactBufferVSMonster            int     `csv:"PlayerHitReactBufferVSMonster"`            // Frame length for the amount of time a player cannot be placed into another hit react from a monster.
	MercenaryDamagePercentVSPlayer           float64 `csv:"MercenaryDamagePercentVSPlayer"`           // Percentage modifier for the total damage a player's mercenary deals to another player.
	MercenaryDamagePercentVSMercenary        float64 `csv:"MercenaryDamagePercentVSMercenary"`        // Percentage modifier for the total damage a player's mercenary deals to another player's mercenary.
	MercenaryDamagePercentVSBoss             float64 `csv:"MercenaryDamagePercentVSBoss"`             // Percentage modifier for the total damage a player's mercenary deals to a boss monster.
	MercenaryMaxStunLength                   int     `csv:"MercenaryMaxStunLength"`                   // Frame length for the maximum stun length allowed on a player's mercenary.
	PrimeEvilDamagePercentVSPlayer           float64 `csv:"PrimeEvilDamagePercentVSPlayer"`           // Percentage modifier applied to the total damage a Prime Evil boss deals to a player.
	PrimeEvilDamagePercentVSMercenary        float64 `csv:"PrimeEvilDamagePercentVSMercenary"`        // Percentage modifier for the total damage a Prime Evil boss deals to a player's mercenary.
	PrimeEvilDamagePercentVSPet              float64 `csv:"PrimeEvilDamagePercentVSPet"`              // Percentage modifier for the total damage a Prime Evil boss deals to a player's pet.
	PetDamagePercentVSPlayer                 float64 `csv:"PetDamagePercentVSPlayer"`                 // Percentage modifier for the total damage a player's pet deals to another player.
	MonsterCEDamagePercent                   float64 `csv:"MonsterCEDamagePercent"`                   // Percentage modifier that affects how much damage is dealt to a player by a Monster's version of Corpse Explosion.
	MonsterFireEnchantExplosionDamagePercent float64 `csv:"MonsterFireEnchantExplosionDamagePercent"` // Percentage modifier that affects how much damage is dealt to a player by a Monster's Fire Enchant explosion.
	StaticFieldMin                           float64 `csv:"StaticFieldMin"`                           // Percentage modifier for capping the amount of current Life damage dealt to monsters by the Sorceress Static Field skill.
	GambleRare                               int     `csv:"GambleRare"`                               // Odds to obtain a Rare item from gambling.
	GambleSet                                int     `csv:"GambleSet"`                                // Odds to obtain a Set item from gambling.
	GambleUnique                             int     `csv:"GambleUnique"`                             // Odds to obtain a Unique item from gambling.
	GambleUber                               int     `csv:"GambleUber"`                               // Odds to make the gambled item be an Exceptional Quality item.
	GambleUltra                              int     `csv:"GambleUltra"`                              // Odds to make the gambled item be an Elite Quality item.
}
