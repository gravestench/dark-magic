package combat

import "fmt"

const (
	minimumHitChance = 5
	maximumHitChance = 95
)

// HitChanceInput is the already-derived rating snapshot consumed by the
// legacy 1.10f hit roll. Equipment, skills, class attributes, monster tables,
// and special target-defense modifiers belong upstream; this function owns
// only the integer arithmetic shared by player and monster attacks.
type HitChanceInput struct {
	AttackerLevel int64
	DefenderLevel int64
	AttackRating  int64
	Defense       int64
}

// LegacyHitChance reproduces SUNITDMG_IsHitSuccessful's final integer steps.
// Each division intentionally truncates before the next multiplication. Do not
// simplify this into one fraction or floating-point expression: that changes
// boundary results from the recovered 1.10f executable behavior.
func LegacyHitChance(input HitChanceInput) (int, error) {
	if input.AttackerLevel <= 0 || input.DefenderLevel <= 0 {
		return 0, fmt.Errorf("combat: hit chance requires positive unit levels")
	}
	attackRating := input.AttackRating
	defense := input.Defense
	if defense < 0 {
		attackRating -= defense
		defense = 0
	}
	if attackRating < 0 {
		defense -= attackRating
		attackRating = 0
	}
	if defense < 0 {
		defense = 0
	}

	ratingFactor := int64(100)
	if divisor := attackRating + defense; divisor != 0 {
		ratingFactor = 100 * attackRating / divisor
	}
	chance := 2 * input.AttackerLevel * ratingFactor / (input.AttackerLevel + input.DefenderLevel)
	if chance < minimumHitChance {
		return minimumHitChance, nil
	}
	if chance > maximumHitChance {
		return maximumHitChance, nil
	}
	return int(chance), nil
}

// LegacyHitRoll uses the recovered comparison: random % 100 must be strictly
// less than chance. A 95% chance therefore admits rolls 0..94 and rejects 95.
func LegacyHitRoll(chance int, random uint64) (bool, error) {
	if chance < minimumHitChance || chance > maximumHitChance {
		return false, fmt.Errorf("combat: hit roll chance must be in [5,95]")
	}
	return int(random%100) < chance, nil
}
