package combat

import (
	"math"
	"testing"
)

func TestEightBitFractionVectors(t *testing.T) {
	vectors := []struct {
		raw        int64
		towardZero int64
		floor      int64
		ceil       int64
		nearest    int64
	}{
		{raw: 0, towardZero: 0, floor: 0, ceil: 0, nearest: 0},
		{raw: 1, towardZero: 0, floor: 0, ceil: 1, nearest: 0},
		{raw: 127, towardZero: 0, floor: 0, ceil: 1, nearest: 0},
		{raw: 128, towardZero: 0, floor: 0, ceil: 1, nearest: 1},
		{raw: 255, towardZero: 0, floor: 0, ceil: 1, nearest: 1},
		{raw: 256, towardZero: 1, floor: 1, ceil: 1, nearest: 1},
		{raw: -1, towardZero: 0, floor: -1, ceil: 0, nearest: 0},
		{raw: -128, towardZero: 0, floor: -1, ceil: 0, nearest: -1},
		{raw: -255, towardZero: 0, floor: -1, ceil: 0, nearest: -1},
		{raw: -256, towardZero: -1, floor: -1, ceil: -1, nearest: -1},
	}
	for _, vector := range vectors {
		amount := FromRaw(vector.raw)
		assertWhole(t, amount, RoundTowardZero, vector.towardZero)
		assertWhole(t, amount, RoundFloor, vector.floor)
		assertWhole(t, amount, RoundCeil, vector.ceil)
		assertWhole(t, amount, RoundNearestAway, vector.nearest)
	}
}

func TestScaleUsesWideIntermediateAndExplicitRounding(t *testing.T) {
	large := FromRaw(math.MaxInt64 - 1)
	result, err := large.Scale(2, 2, RoundTowardZero)
	if err != nil {
		t.Fatal(err)
	}
	if result != large {
		t.Fatalf("wide intermediate result = %d, want %d", result, large)
	}
	half := FromRaw(One.Raw())
	positive, err := half.Scale(1, 3, RoundNearestAway)
	if err != nil || positive.Raw() != 85 {
		t.Fatalf("one third = %d, err=%v", positive.Raw(), err)
	}
	negative, err := FromRaw(-256).Scale(1, 3, RoundFloor)
	if err != nil || negative.Raw() != -86 {
		t.Fatalf("negative floor third = %d, err=%v", negative.Raw(), err)
	}
}

func TestAmountOverflowIsRejected(t *testing.T) {
	if _, err := FromWhole(math.MaxInt64); err == nil {
		t.Fatal("expected whole conversion overflow")
	}
	if _, err := FromRaw(math.MaxInt64).Add(1); err == nil {
		t.Fatal("expected addition overflow")
	}
	if _, err := FromRaw(math.MinInt64).Sub(1); err == nil {
		t.Fatal("expected subtraction overflow")
	}
	if _, err := FromRaw(math.MaxInt64).Scale(2, 1, RoundTowardZero); err == nil {
		t.Fatal("expected scale overflow")
	}
	maximum, err := FromRaw(-1).Sub(FromRaw(math.MinInt64))
	if err != nil || maximum.Raw() != math.MaxInt64 {
		t.Fatalf("-1 - MinInt64 = %d, err=%v", maximum.Raw(), err)
	}
	zero, err := FromRaw(math.MinInt64).Sub(FromRaw(math.MinInt64))
	if err != nil || zero != 0 {
		t.Fatalf("MinInt64 - MinInt64 = %d, err=%v", zero, err)
	}
}

func TestScaleSignIncludesNegativeDenominator(t *testing.T) {
	vectors := []struct {
		left, numerator, denominator int64
		want                         int64
	}{
		{left: 256, numerator: 1, denominator: -3, want: -86},
		{left: -256, numerator: 1, denominator: -3, want: 85},
		{left: -256, numerator: -1, denominator: -3, want: -86},
	}
	for _, vector := range vectors {
		got, err := FromRaw(vector.left).Scale(vector.numerator, vector.denominator, RoundFloor)
		if err != nil || got.Raw() != vector.want {
			t.Fatalf("%d * %d / %d = %d, err=%v", vector.left, vector.numerator, vector.denominator, got.Raw(), err)
		}
	}
}

func TestScaleHotPathDoesNotAllocate(t *testing.T) {
	allocations := testing.AllocsPerRun(1000, func() {
		if _, err := FromRaw(123456).Scale(128, 100, RoundTowardZero); err != nil {
			t.Fatal(err)
		}
	})
	if allocations != 0 {
		t.Fatalf("Scale allocations = %v, want 0", allocations)
	}
}

func TestMultiplyDivideSupportsWholeAuthoritativeValues(t *testing.T) {
	got, err := MultiplyDivide(9, 186, 100, RoundTowardZero)
	if err != nil || got != 16 {
		t.Fatalf("9 * 186 / 100 = %d, err=%v", got, err)
	}
}

func assertWhole(t *testing.T, amount Amount, rounding Rounding, want int64) {
	t.Helper()
	got, err := amount.Whole(rounding)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("raw %d rounding %d = %d, want %d", amount.Raw(), rounding, got, want)
	}
}
