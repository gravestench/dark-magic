package models

// Difficultylevel represents the data structure for the difficultylevels.txt file.
type Difficultylevel struct {
	// The difficulty mode name.
	Name string `csv:"Name"`
	// Baseline starting point for player character's resistances for Expansion mode.
	ResistPenalty int `csv:"ResistPenalty"`
	// Baseline starting point for player character's resistances for Non-Expansion mode.
	ResistPenaltyNonExpansion int `csv:"ResistPenaltyNonExpansion"`
	// Modifies the percentage of current level experience lost when a player character dies.
	DeathExpPenalty float64 `csv:"DeathExpPenalty"`
	// Additional skill levels added to skills used by monsters (defined from monstats.txt).
	MonsterSkillBonus int `csv:"MonsterSkillBonus"`
	// Divisor affecting all Freeze Length values on monsters.
	MonsterFreezeDivisor int `csv:"MonsterFreezeDivisor"`
	// Divisor affecting all Cold Length values on monsters.
	MonsterColdDivisor int `csv:"MonsterColdDivisor"`
	// Divisor affecting durations of Curses on monsters.
	AiCurseDivisor int `csv:"AiCurseDivisor"`
	// Divisor affecting the amount of Life Steal that player characters gain.
	LifeStealDivisor int `csv:"LifeStealDivisor"`
	// Divisor affecting the amount of Mana Steal that player characters gain.
	ManaStealDivisor int `csv:"ManaStealDivisor"`
	// Percentage modifier for a Unique monster's overall Damage and Attack Rating.
	UniqueDamageBonus float64 `csv:"UniqueDamageBonus"`
	// Percentage modifier for a Champion monster's overall Damage and Attack Rating.
	ChampionDamageBonus float64 `csv:"ChampionDamageBonus"`
	// Percentage modifier for the total damage a player deals to another player.
	PlayerDamagePercentVSPlayer float64 `csv:"PlayerDamagePercentVSPlayer"`
	// Percentage modifier for the total damage a player deals to another player's mercenary.
	PlayerDamagePercentVSMercenary float64 `csv:"PlayerDamagePercentVSMercenary"`
	// Percentage modifier for the total damage a player deals to a Prime Evil boss.
	PlayerDamagePercentVSPrimeEvil float64 `csv:"PlayerDamagePercentVSPrimeEvil"`
	// Frame length for the amount of time a player cannot be placed into another hit react from a player.
	PlayerHitReactBufferVSPlayer int `csv:"PlayerHitReactBufferVSPlayer"`
	// Frame length for the amount of time a player cannot be placed into another hit react from a monster.
	PlayerHitReactBufferVSMonster int `csv:"PlayerHitReactBufferVSMonster"`
	// Percentage modifier for the total damage a player's mercenary deals to another player.
	MercenaryDamagePercentVSPlayer float64 `csv:"MercenaryDamagePercentVSPlayer"`
	// Percentage modifier for the total damage a player's mercenary deals to another player's mercenary.
	MercenaryDamagePercentVSMercenary float64 `csv:"MercenaryDamagePercentVSMercenary"`
	// Percentage modifier for the total damage a player's mercenary deals to a boss monster.
	MercenaryDamagePercentVSBoss float64 `csv:"MercenaryDamagePercentVSBoss"`
	// Frame length for the maximum stun length allowed on a player's mercenary.
	MercenaryMaxStunLength int `csv:"MercenaryMaxStunLength"`
	// Percentage modifier applied to the total damage a Prime Evil boss deals to a player.
	PrimeEvilDamagePercentVSPlayer float64 `csv:"PrimeEvilDamagePercentVSPlayer"`
	// Percentage modifier for the total damage a Prime Evil boss deals to a player's mercenary.
	PrimeEvilDamagePercentVSMercenary float64 `csv:"PrimeEvilDamagePercentVSMercenary"`
	// Percentage modifier for the total damage a Prime Evil boss deals to a player's pet.
	PrimeEvilDamagePercentVSPet float64 `csv:"PrimeEvilDamagePercentVSPet"`
	// Percentage modifier for the total damage a player's pet deals to another player.
	PetDamagePercentVSPlayer float64 `csv:"PetDamagePercentVSPlayer"`
	// Percentage modifier that affects how much damage is dealt to a player by a Monster's version of Corpse Explosion.
	MonsterCEDamagePercent float64 `csv:"MonsterCEDamagePercent"`
	// Percentage modifier that affects how much damage is dealt to a player by a Monster's Fire Enchant explosion.
	MonsterFireEnchantExplosionDamagePercent float64 `csv:"MonsterFireEnchantExplosionDamagePercent"`
	// Percentage modifier for capping the amount of current Life damage dealt to monsters by the Sorceress Static Field
	// skill.
	StaticFieldMin float64 `csv:"StaticFieldMin"`
	// Odds to obtain a Rare item from gambling.
	GambleRare int `csv:"GambleRare"`
	// Odds to obtain a Set item from gambling.
	GambleSet int `csv:"GambleSet"`
	// Odds to obtain a Unique item from gambling.
	GambleUnique int `csv:"GambleUnique"`
	// Odds to make the gambled item be an Exceptional Quality item.
	GambleUber int `csv:"GambleUber"`
	// Odds to make the gambled item be an Elite Quality item.
	GambleUltra int `csv:"GambleUltra"`
}
