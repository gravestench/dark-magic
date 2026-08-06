package models

// DifficultyLevels represents the global parameters for game rules between each difficulty mode
type DifficultyLevels struct {
	// Overview: This file controls global parameters for game rules and how they work between each difficulty mode.
	// Users cannot add new difficulty modes from this file.

	// Data Fields:
	Name                                     string // This is a reference field to define the difficulty mode.
	ResistPenalty                            int    // Defines the baseline starting point for a player character’s resistances for Expansion mode.
	ResistPenaltyNonExpansion                int    // Defines the baseline starting point for a player character’s resistances for Non-Expansion mode.
	DeathExpPenalty                          int    // Modifies the percentage of current level experience lost when a player character dies.
	MonsterSkillBonus                        int    // Adds additional skill levels to skills used by monsters (defined from monstats.txt).
	MonsterFreezeDivisor                     int    // Divisor that affects all Freeze Length values on monsters. The attempted Freeze Length value is divided by this divisor to determine the actual Freeze Length.
	MonsterColdDivisor                       int    // Divisor that affects all Cold Length values on monsters. The attempted Cold Length value is divided by this divisor to determine the actual Cold Length.
	AiCurseDivisor                           int    // Divisor that affects all durations of Curses on monsters. The attempted Curse duration is divided by this divisor to determine the actual Curse duration.
	LifeStealDivisor                         int    // Divisor that affects the amount of Life Steal that player characters gain. The attempted Life Steal value is divided by this divisor to determine the actual Life Steal.
	ManaStealDivisor                         int    // Divisor that affects the amount of Mana Steal that player characters gain. The attempted Mana Steal value is divided by this divisor to determine the actual Mana Steal.
	UniqueDamageBonus                        int    // Percentage modifier for a Unique monster’s overall Damage and Attack Rating. This is applied after calculating the monster’s other modifications.
	ChampionDamageBonus                      int    // Percentage modifier for a Champion monster’s overall Damage and Attack Rating. This is applied after calculating the monster’s other modifications.
	PlayerDamagePercentVSPlayer              int    // Percentage modifier for the total damage a player deals to another player.
	PlayerDamagePercentVSMercenary           int    // Percentage modifier for the total damage a player deals to another player’s mercenary.
	PlayerDamagePercentVSPrimeEvil           int    // Percentage modifier for the total damage a player deals to a Prime Evil boss.
	PlayerHitReactBufferVSPlayer             int    // The frame length for the amount of time a player cannot be placed into another hit react from a player (25 frames = 1 second).
	PlayerHitReactBufferVSMonster            int    // The frame length for the amount of time a player cannot be placed into another hit react from a monster (25 frames = 1 second).
	MercenaryDamagePercentVSPlayer           int    // Percentage modifier for the total damage a player’s mercenary deals to another player.
	MercenaryDamagePercentVSMercenary        int    // Percentage modifier for the total damage a player’s mercenary deals to another player’s mercenary.
	MercenaryDamagePercentVSBoss             int    // Percentage modifier for the total damage a player’s mercenary deals to a boss monster.
	MercenaryMaxStunLength                   int    // The frame length for the maximum stun length allowed on a player’s mercenary (25 Frames = 1 second).
	PrimeEvilDamagePercentVSPlayer           int    // Percentage modifier applied to the total damage a Prime Evil boss deals to a player.
	PrimeEvilDamagePercentVSMercenary        int    // Percentage modifier for the total damage a Prime Evil boss deals to a player’s mercenary.
	PrimeEvilDamagePercentVSPet              int    // Percentage modifier for the total damage a Prime Evil boss deals to a player’s pet.
	PetDamagePercentVSPlayer                 int    // Percentage modifier for the total damage a player’s pet deals to another player.
	MonsterCEDamagePercent                   int    // Percentage modifier that affects how much damage is dealt to a player by a Monster’s version of Corpse Explosion. For example, when certain monsters die and explode on death.
	MonsterFireEnchantExplosionDamagePercent int    // Percentage modifier that affects how much damage is dealt to a player by a Monster’s Fire Enchant explosion. The Fire Enchant death explosion uses the same Corpse Explosion functionality and this value is applied after the “MonsterCEDamagePercent” value.
	StaticFieldMin                           int    // Percentage modifier for capping the amount of current Life damage dealt to monsters by the Sorceress Static Field skill. This field only affects games in Expansion mode.
	GambleRare                               int    // The odds to obtain a Rare item from gambling. The game rolls a random number between 0 to 100000. If that rolled number is less than this value, then the gambled item will be a Rare item.
	GambleSet                                int    // The odds to obtain a Set item from gambling. The game rolls a random number between 0 to 100000. If that rolled number is less than this value, then the gambled item will be a Set item.
	GambleUnique                             int    // The odds to obtain a Unique item from gambling. The game rolls a random number between 0 to 100000. If that rolled number is less than this value, then the gambled item will be a Unique item.
	GambleUber                               int    // The odds to make the gambled item be an Exceptional Quality item. The game rolls a random number between 0 to 10000. This rolled number is then compared to the following formula: ([Item Level] - [Base Item Level]) * [“GambleUber”] + 1. If the rolled number is less than this value, then the item becomes an Exceptional Quality item, and the game will roll for upgrading it to Elite Quality (See “GambleUltra”).
	GambleUltra                              int    // The odds to make the gambled item be an Elite Quality item. The game rolls a random number between 0 to 10000. This rolled number is then compared to the following formula: ([Item Level] - [Base Item Level]) * [“GambleUltra”] + 1. If the rolled number is less than this value, then the item is upgraded to an Elite Quality item. This only happens if the item successfully rolled for Exceptional Quality.
}
