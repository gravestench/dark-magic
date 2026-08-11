package combat

import (
	"fmt"
	"math"
	"math/bits"
)

const (
	// FractionBits records the source-derived 8-bit fractional vocabulary.
	FractionBits = 8
	// One is one whole point represented in raw simulation units.
	One Amount = 1 << FractionBits
)

// Amount is a signed fixed-point combat quantity. Its stored integer is raw:
// 256 means one whole point, 128 means one half, and -256 means negative one.
type Amount int64

// Rounding names how discarded fractional data is handled. Requiring a mode
// keeps truncation from becoming a hidden, accidental combat rule.
type Rounding uint8

const (
	RoundTowardZero Rounding = iota
	RoundFloor
	RoundCeil
	RoundNearestAway
)

func (rounding Rounding) validate() error {
	switch rounding {
	case RoundTowardZero, RoundFloor, RoundCeil, RoundNearestAway:
		return nil
	default:
		return fmt.Errorf("combat: unsupported rounding mode %d", rounding)
	}
}

// FromRaw labels an already-scaled value without changing it.
func FromRaw(raw int64) Amount { return Amount(raw) }

// FromWhole converts whole points to fixed-point units and rejects overflow.
func FromWhole(whole int64) (Amount, error) {
	if whole > math.MaxInt64/int64(One) || whole < math.MinInt64/int64(One) {
		return 0, fmt.Errorf("combat: whole amount %d overflows fixed point", whole)
	}
	return Amount(whole * int64(One)), nil
}

// MustWhole is convenient for compile-time-sized fixtures and constants. It
// panics on overflow and should not be used to admit untrusted runtime data.
func MustWhole(whole int64) Amount {
	amount, err := FromWhole(whole)
	if err != nil {
		panic(err)
	}
	return amount
}

// Raw returns the exact simulation representation.
func (amount Amount) Raw() int64 { return int64(amount) }

// Whole converts to display-sized whole points using an explicit rounding rule.
func (amount Amount) Whole(rounding Rounding) (int64, error) {
	return divideRounded(int64(amount), int64(One), rounding)
}

// Add rejects signed overflow rather than allowing a wrapped value to enter a
// replay checkpoint.
func (amount Amount) Add(other Amount) (Amount, error) {
	left, right := int64(amount), int64(other)
	if (right > 0 && left > math.MaxInt64-right) || (right < 0 && left < math.MinInt64-right) {
		return 0, fmt.Errorf("combat: amount addition overflows")
	}
	return Amount(left + right), nil
}

// Sub rejects signed overflow.
func (amount Amount) Sub(other Amount) (Amount, error) {
	if other == Amount(math.MinInt64) {
		if amount >= 0 {
			return 0, fmt.Errorf("combat: amount subtraction overflows")
		}
		// A negative left side can subtract MinInt64 without first negating it.
		return Amount(int64(amount) - math.MinInt64), nil
	}
	return amount.Add(-other)
}

// Scale multiplies by numerator/denominator with a 128-bit intermediate. Combat
// state remains int64, but a valid final result is not rejected merely because
// the temporary product is wider than int64. This path allocates no heap objects.
func (amount Amount) Scale(numerator, denominator int64, rounding Rounding) (Amount, error) {
	result, err := multiplyDivideRounded(int64(amount), numerator, denominator, rounding)
	if err != nil {
		return 0, err
	}
	return Amount(result), nil
}

func divideRounded(numerator, denominator int64, rounding Rounding) (int64, error) {
	return multiplyDivideRounded(numerator, 1, denominator, rounding)
}

// multiplyDivideRounded works with unsigned magnitudes, then restores the sign.
// bits.Mul64 gives the full two-word product and bits.Div64 divides that product
// without losing the fractional remainder needed for explicit rounding.
func multiplyDivideRounded(left, right, denominator int64, rounding Rounding) (int64, error) {
	if denominator == 0 {
		return 0, fmt.Errorf("combat: division by zero")
	}
	if err := rounding.validate(); err != nil {
		return 0, err
	}
	if left == 0 || right == 0 {
		return 0, nil
	}
	negativeFactors := 0
	if left < 0 {
		negativeFactors++
	}
	if right < 0 {
		negativeFactors++
	}
	if denominator < 0 {
		negativeFactors++
	}
	negative := negativeFactors%2 != 0
	high, low := bits.Mul64(magnitude(left), magnitude(right))
	divisor := magnitude(denominator)
	// Div64 requires high < divisor. If it is not, the unsigned quotient needs
	// more than 64 bits and cannot possibly fit our signed Amount.
	if high >= divisor {
		return 0, fmt.Errorf("combat: scaled amount overflows")
	}
	quotient, remainder := bits.Div64(high, low, divisor)
	adjust := false
	switch rounding {
	case RoundFloor:
		adjust = negative && remainder != 0
	case RoundCeil:
		adjust = !negative && remainder != 0
	case RoundNearestAway:
		// remainder*2 >= divisor, written without overflowing uint64.
		adjust = remainder != 0 && remainder >= divisor-remainder
	}
	if adjust {
		if quotient == math.MaxUint64 {
			return 0, fmt.Errorf("combat: rounded amount overflows")
		}
		quotient++
	}
	if !negative {
		if quotient > math.MaxInt64 {
			return 0, fmt.Errorf("combat: scaled amount overflows")
		}
		return int64(quotient), nil
	}
	const minMagnitude = uint64(math.MaxInt64) + 1
	if quotient > minMagnitude {
		return 0, fmt.Errorf("combat: scaled amount overflows")
	}
	if quotient == minMagnitude {
		return math.MinInt64, nil
	}
	return -int64(quotient), nil
}

func magnitude(value int64) uint64 {
	if value >= 0 {
		return uint64(value)
	}
	// Writing this as -(value+1)+1 avoids negating MinInt64 directly.
	return uint64(-(value + 1)) + 1
}
