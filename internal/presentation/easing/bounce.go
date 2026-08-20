package easing

// These coefficients preserve the repository's historical four-branch bounce math. The formulas intentionally retain
// their original multiplication order, even though it differs from common normalized bounce implementations.
const (
	bounceSegmentScale  = 2.75
	bounceCurveScale    = 7.5625
	thirdSegmentLimit   = 2.5
	thirdSegmentOffset  = 0.9375
	thirdSegmentShift   = 2.25
	fourthSegmentShift  = 2.625
	fourthSegmentOffset = 0.984375
	secondSegmentShift  = 1.5
	secondSegmentOffset = 0.75
	secondSegmentLimit  = 2
)

var _ EaseFunctionProvider = &BounceOutEaseProvider{}
var _ EaseFunctionProvider = &BounceInEaseProvider{}
var _ EaseFunctionProvider = &BounceInOutEaseProvider{}

// BounceOutEaseProvider exposes the legacy piecewise curve registered under the bounce-out name.
type BounceOutEaseProvider struct{}

// New returns the fixed four-segment formula without normalizing its historical overshoot.
func (*BounceOutEaseProvider) New(_ []float64) func(float64) float64 {
	bounce := func(v float64) float64 {
		return bounceOutValue(v)
	}

	return bounce
}

// BounceInEaseProvider reflects the historical bounce-out values around both axes.
type BounceInEaseProvider struct{}

// New returns the fixed bounce-in curve by evaluating reflected progress.
func (*BounceInEaseProvider) New(_ []float64) func(float64) float64 {
	bounce := func(v float64) float64 {
		// Reflecting before evaluation preserves the exact bounce-out partition.
		v = 1 - v

		return 1 - bounceOutValue(v)
	}

	return bounce
}

// BounceInOutEaseProvider maps both halves through the same historical piecewise formula.
type BounceInOutEaseProvider struct{}

// New preserves the existing midpoint reflection while sharing coefficients between both halves.
func (*BounceInOutEaseProvider) New(_ []float64) func(float64) float64 {
	bounce := func(v float64) float64 {
		reverse := false

		// Normalize either half to the bounce-out domain, then reflect the result
		// for the first half after evaluating the common piecewise formula.
		if v < 0.5 {
			v = 1 - (v * 2)
			reverse = true
		} else {
			v = (v * 2) - 1
		}

		v = bounceOutValue(v)

		if reverse {
			return (1 - v) * 0.5
		} else {
			return v*0.5 + 0.5
		}
	}

	return bounce
}

// bounceOutValue centralizes the exact historical boundaries and arithmetic used by all three providers.
func bounceOutValue(v float64) float64 {
	if v < 1/bounceSegmentScale {
		return bounceCurveScale * v * v
	}

	if v < secondSegmentLimit/bounceSegmentScale {
		return bounceCurveScale*(v-secondSegmentShift/bounceSegmentScale)*v + secondSegmentOffset
	}

	if v < thirdSegmentLimit/bounceSegmentScale {
		return bounceCurveScale*(v-thirdSegmentShift/bounceSegmentScale)*v + thirdSegmentOffset
	}

	return bounceCurveScale*(v-fourthSegmentShift/bounceSegmentScale)*v + fourthSegmentOffset
}
