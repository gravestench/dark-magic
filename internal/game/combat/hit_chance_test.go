package combat

import "testing"

func TestLegacyHitChancePreservesRecoveredIntegerOrder(t *testing.T) {
	tests := []struct {
		name  string
		input HitChanceInput
		want  int
	}{
		{name: "equal ratings and levels", input: HitChanceInput{AttackerLevel: 10, DefenderLevel: 10, AttackRating: 100, Defense: 100}, want: 50},
		{name: "rating division truncates first", input: HitChanceInput{AttackerLevel: 7, DefenderLevel: 11, AttackRating: 101, Defense: 100}, want: 38},
		{name: "minimum clamp", input: HitChanceInput{AttackerLevel: 1, DefenderLevel: 99, AttackRating: 1, Defense: 10000}, want: 5},
		{name: "maximum clamp", input: HitChanceInput{AttackerLevel: 99, DefenderLevel: 1, AttackRating: 10000, Defense: 1}, want: 95},
		{name: "zero ratings use recovered full factor", input: HitChanceInput{AttackerLevel: 10, DefenderLevel: 10}, want: 95},
		{name: "negative defense adds to attack rating", input: HitChanceInput{AttackerLevel: 10, DefenderLevel: 10, AttackRating: 50, Defense: -50}, want: 95},
		{name: "negative attack rating adds to defense", input: HitChanceInput{AttackerLevel: 10, DefenderLevel: 10, AttackRating: -50, Defense: 50}, want: 5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := LegacyHitChance(test.input)
			if err != nil || got != test.want {
				t.Fatalf("chance = %d, %v; want %d", got, err, test.want)
			}
		})
	}
}

func TestLegacyHitChanceRejectsNonPositiveLevels(t *testing.T) {
	for _, input := range []HitChanceInput{{DefenderLevel: 1}, {AttackerLevel: 1}} {
		if _, err := LegacyHitChance(input); err == nil {
			t.Fatalf("accepted invalid levels: %#v", input)
		}
	}
}

func TestLegacyHitRollUsesStrictModuloComparison(t *testing.T) {
	for _, test := range []struct {
		chance int
		roll   uint64
		want   bool
	}{{95, 94, true}, {95, 95, false}, {5, 4, true}, {5, 5, false}, {50, 149, true}, {50, 150, false}} {
		got, err := LegacyHitRoll(test.chance, test.roll)
		if err != nil || got != test.want {
			t.Fatalf("roll(%d,%d) = %t, %v; want %t", test.chance, test.roll, got, err, test.want)
		}
	}
}
