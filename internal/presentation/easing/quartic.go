package easing

var _ EaseFunctionProvider = &QuarticOutEaseProvider{}
var _ EaseFunctionProvider = &QuarticInEaseProvider{}
var _ EaseFunctionProvider = &QuarticInOutEaseProvider{}

// QuarticOutEaseProvider exposes the repository's historical fourth-degree out formula.
type QuarticOutEaseProvider struct{}

// New retains the existing decreasing polynomial rather than substituting the conventional shifted formula.
func (*QuarticOutEaseProvider) New(_ []float64) func(float64) float64 {
	quartic := func(v float64) float64 {
		return 1 - v*v*v*v
	}

	return quartic
}

// QuarticInEaseProvider accelerates progress with a fourth-degree polynomial.
type QuarticInEaseProvider struct{}

// New returns the parameter-free quartic-in curve.
func (*QuarticInEaseProvider) New(_ []float64) func(float64) float64 {
	quartic := func(v float64) float64 {
		return v * v * v * v
	}

	return quartic
}

// QuarticInOutEaseProvider joins quartic acceleration and deceleration.
type QuarticInOutEaseProvider struct{}

// New returns a normalized two-branch quartic curve.
func (*QuarticInOutEaseProvider) New(_ []float64) func(float64) float64 {
	quartic := func(v float64) float64 {
		v *= 2
		if v < 1 {
			return 0.5 * v * v * v * v
		} else {
			return -0.5 * ((v-2)*v*v*v - 2)
		}
	}

	return quartic
}
