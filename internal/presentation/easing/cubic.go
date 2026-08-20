package easing

var _ EaseFunctionProvider = &CubicOutEaseProvider{}
var _ EaseFunctionProvider = &CubicInEaseProvider{}
var _ EaseFunctionProvider = &CubicInOutEaseProvider{}

// CubicOutEaseProvider decelerates progress with a third-degree polynomial.
type CubicOutEaseProvider struct{}

// New returns the parameter-free cubic-out curve.
func (*CubicOutEaseProvider) New(_ []float64) func(float64) float64 {
	cubic := func(v float64) float64 {
		v--
		return v*v*v + 1
	}

	return cubic
}

// CubicInEaseProvider accelerates progress with a third-degree polynomial.
type CubicInEaseProvider struct{}

// New returns the parameter-free cubic-in curve.
func (*CubicInEaseProvider) New(_ []float64) func(float64) float64 {
	cubic := func(v float64) float64 {
		return v * v * v
	}

	return cubic
}

// CubicInOutEaseProvider joins cubic acceleration and deceleration.
type CubicInOutEaseProvider struct{}

// New returns a normalized two-branch cubic curve.
func (*CubicInOutEaseProvider) New(_ []float64) func(float64) float64 {
	cubic := func(v float64) float64 {
		v *= 2
		if v < 1 {
			return 0.5 * v * v * v
		}

		v -= 2

		return 0.5 * (v*v*v + 2)
	}

	return cubic
}
