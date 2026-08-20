package easing

import (
	"math"
)

var _ EaseFunctionProvider = &CircularOutEaseProvider{}
var _ EaseFunctionProvider = &CircularInEaseProvider{}
var _ EaseFunctionProvider = &CircularInOutEaseProvider{}

// CircularOutEaseProvider follows the upper quarter of a unit circle.
type CircularOutEaseProvider struct{}

// New returns the parameter-free circular-out formula.
func (*CircularOutEaseProvider) New(_ []float64) func(float64) float64 {
	circ := func(v float64) float64 {
		return math.Sqrt(1 - (v * v))
	}

	return circ
}

// CircularInEaseProvider mirrors circular-out into an accelerating curve.
type CircularInEaseProvider struct{}

// New returns the parameter-free circular-in formula.
func (*CircularInEaseProvider) New(_ []float64) func(float64) float64 {
	circ := func(v float64) float64 {
		return 1 - math.Sqrt(1-v*v)
	}

	return circ
}

// CircularInOutEaseProvider joins accelerating and decelerating circle arcs.
type CircularInOutEaseProvider struct{}

// New returns a curve whose two branches meet at normalized progress 0.5.
func (*CircularInOutEaseProvider) New(_ []float64) func(float64) float64 {
	circ := func(v float64) float64 {
		// Scaling once keeps the branch formulas expressed in their conventional
		// zero-to-one domains without changing the shared midpoint.
		v *= 2
		if v < 1 {
			return -0.5 * (math.Sqrt(1-v*v) - 1)
		}

		return 0.5 * (math.Sqrt(1-(v-2)*v) + 1)
	}

	return circ
}
