package models

// MonsterLevelStats represents how monster statistics increase per level based on the current game type and difficulty.
type MonsterLevelStats struct {
	Level int `csv:"Level"` // An integer value to determine how to scale the monster's statistics when at a specific level.

	// Non-Ladder Battle.net / Singleplayer / Open Battle.net / TCP
	Defense       float64 `csv:"AC"`    // Percentage multiplier for increasing the Monster's Defense (multiplies with the "AC" field in monstats.txt).
	DefenseN      float64 `csv:"AC(N)"` // Percentage multiplier for increasing the Monster's Defense in Nightmare difficulty.
	DefenseH      float64 `csv:"AC(H)"` // Percentage multiplier for increasing the Monster's Defense in Hell difficulty.
	AttackRating  float64 `csv:"TH"`    // Percentage multiplier for increasing the Monster's Attack Rating (multiplies with the "A1TH" and "A2TH" fields in monstats.txt).
	AttackRatingN float64 `csv:"TH(N)"` // Percentage multiplier for increasing the Monster's Attack Rating in Nightmare difficulty.
	AttackRatingH float64 `csv:"TH(H)"` // Percentage multiplier for increasing the Monster's Attack Rating in Hell difficulty.
	Life          float64 `csv:"HP"`    // Percentage multiplier for increasing the Monster's Life (multiplies with the "MinHP" and "MaxHP" fields in monstats.txt).
	LifeN         float64 `csv:"HP(N)"` // Percentage multiplier for increasing the Monster's Life in Nightmare difficulty.
	LifeH         float64 `csv:"HP(H)"` // Percentage multiplier for increasing the Monster's Life in Hell difficulty.
	Damage        float64 `csv:"DM"`    // Percentage multiplier for increasing the Monster's Damage (multiplies with various damage fields in monstats.txt).
	DamageN       float64 `csv:"DM(N)"` // Percentage multiplier for increasing the Monster's Damage in Nightmare difficulty.
	DamageH       float64 `csv:"DM(H)"` // Percentage multiplier for increasing the Monster's Damage in Hell difficulty.
	Experience    float64 `csv:"XP"`    // Percentage multiplier for increasing the Experience provided to the player when killing the Monster (multiplies with the "Exp" fields in monstats.txt).
	ExperienceN   float64 `csv:"XP(N)"` // Percentage multiplier for increasing the Experience in Nightmare difficulty.
	ExperienceH   float64 `csv:"XP(H)"` // Percentage multiplier for increasing the Experience in Hell difficulty.

	// Ladder Battle.net game type
	LadderDefense     float64 `csv:"L-AC"`    // Percentage multiplier for increasing the Monster's Defense in Ladder Battle.net (multiplies with the "AC" field in monstats.txt).
	LadderDefenseN    float64 `csv:"L-AC(N)"` // Percentage multiplier for increasing the Monster's Defense in Ladder Battle.net - Nightmare difficulty.
	LadderDefenseH    float64 `csv:"L-AC(H)"` // Percentage multiplier for increasing the Monster's Defense in Ladder Battle.net - Hell difficulty.
	LadderLife        float64 `csv:"L-HP"`    // Percentage multiplier for increasing the Monster's Life in Ladder Battle.net (multiplies with the "MinHP" and "MaxHP" fields in monstats.txt).
	LadderLifeN       float64 `csv:"L-HP(N)"` // Percentage multiplier for increasing the Monster's Life in Ladder Battle.net - Nightmare difficulty.
	LadderLifeH       float64 `csv:"L-HP(H)"` // Percentage multiplier for increasing the Monster's Life in Ladder Battle.net - Hell difficulty.
	LadderDamage      float64 `csv:"L-DM"`    // Percentage multiplier for increasing the Monster's Damage in Ladder Battle.net (multiplies with various damage fields in monstats.txt).
	LadderDamageN     float64 `csv:"L-DM(N)"` // Percentage multiplier for increasing the Monster's Damage in Ladder Battle.net - Nightmare difficulty.
	LadderDamageH     float64 `csv:"L-DM(H)"` // Percentage multiplier for increasing the Monster's Damage in Ladder Battle.net - Hell difficulty.
	LadderExperience  float64 `csv:"L-XP"`    // Percentage multiplier for increasing the Experience provided to the player when killing the Monster in Ladder Battle.net (multiplies with the "Exp" fields in monstats.txt).
	LadderExperienceN float64 `csv:"L-XP(N)"` // Percentage multiplier for increasing the Experience in Ladder Battle.net - Nightmare difficulty.
	LadderExperienceH float64 `csv:"L-XP(H)"` // Percentage multiplier for increasing the Experience in Ladder Battle.net - Hell difficulty.
}
