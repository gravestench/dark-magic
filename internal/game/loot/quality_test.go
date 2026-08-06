package loot

import (
	"reflect"
	"strings"
	"testing"
)

func testQualityRatio() QualityRatio {
	return QualityRatio{
		Unique:   QualityRule{Base: 400, Divisor: 2, Minimum: 6400},
		Set:      QualityRule{Base: 300, Divisor: 2, Minimum: 5600},
		Rare:     QualityRule{Base: 200, Divisor: 2, Minimum: 4800},
		Magic:    QualityRule{Base: 100, Divisor: 2, Minimum: 3200},
		Superior: QualityRule{Base: 20, Divisor: 2},
		Normal:   QualityRule{Base: 2, Divisor: 1},
	}
}

func TestCalculateQualityChances(t *testing.T) {
	ratio := testQualityRatio()
	plain, err := CalculateQualityChances(ratio, QualityContext{MonsterLevel: 20, ItemLevel: 10})
	if err != nil {
		t.Fatal(err)
	}
	if plain.Unique != 395*128 || plain.Magic != 95*128 || plain.Superior != 15*128 || plain.Normal != 128 {
		t.Fatalf("plain chances = %#v", plain)
	}
	boosted, err := CalculateQualityChances(ratio, QualityContext{
		MonsterLevel: 20, ItemLevel: 10, MagicFind: 100,
		Modifiers: QualityModifiers{Unique: 512, Set: 256, Rare: 128, Magic: 64},
	})
	if err != nil {
		t.Fatal(err)
	}
	if boosted.Unique >= plain.Unique || boosted.Set >= plain.Set || boosted.Rare >= plain.Rare || boosted.Magic >= plain.Magic {
		t.Fatalf("boosted chances should have smaller denominators: plain=%#v boosted=%#v", plain, boosted)
	}
	if boosted.Superior != plain.Superior || boosted.Normal != plain.Normal {
		t.Fatalf("magic find changed non-magical fallback checks: plain=%#v boosted=%#v", plain, boosted)
	}
}

func TestQualityMinimumIsAppliedBeforeTreasureModifier(t *testing.T) {
	ratio := testQualityRatio()
	ratio.Unique = QualityRule{Base: 2, Minimum: 6400}
	chances, err := CalculateQualityChances(ratio, QualityContext{Modifiers: QualityModifiers{Unique: 512}})
	if err != nil {
		t.Fatal(err)
	}
	if chances.Unique != 3200 {
		t.Fatalf("unique denominator = %d, want 3200", chances.Unique)
	}
}

func TestRollQualityIsDeterministic(t *testing.T) {
	ratio := testQualityRatio()
	context := QualityContext{MonsterLevel: 30, ItemLevel: 15, MagicFind: 75}
	quality, chances, err := RollQuality(ratio, context, 123)
	if err != nil {
		t.Fatal(err)
	}
	qualityAgain, chancesAgain, err := RollQuality(ratio, context, 123)
	if err != nil {
		t.Fatal(err)
	}
	if quality != qualityAgain || !reflect.DeepEqual(chances, chancesAgain) {
		t.Fatalf("same seed differs: %q/%#v != %q/%#v", quality, chances, qualityAgain, chancesAgain)
	}
}

func TestCalculateQualityRejectsInvalidRules(t *testing.T) {
	ratio := testQualityRatio()
	ratio.Unique.Base = 0
	_, err := CalculateQualityChances(ratio, QualityContext{})
	if err == nil || !strings.Contains(err.Error(), "unique quality") {
		t.Fatalf("error = %v", err)
	}
}
