package easing

import (
	"math"
)

var _ EaseFunctionProvider = &ExponentialOutEaseProvider{}
var _ EaseFunctionProvider = &ExponentialInEaseProvider{}
var _ EaseFunctionProvider = &ExponentialInOutEaseProvider{}

// ExponentialOutEaseProvider approaches the destination rapidly then settles.
type ExponentialOutEaseProvider struct{}

// New returns the established exponential-out approximation.
func (*ExponentialOutEaseProvider) New(_ []float64) func(float64) float64 {
	expo := func(v float64) float64 {
		return 1 - math.Pow(2, -10*v)
	}

	return expo
}

// ExponentialInEaseProvider starts almost stationary then accelerates sharply.
type ExponentialInEaseProvider struct{}

// New preserves the small historical offset used near the starting endpoint.
func (*ExponentialInEaseProvider) New(_ []float64) func(float64) float64 {
	expo := func(v float64) float64 {
		return math.Pow(2, 10*(v-1)) - 0.001
	}

	return expo
}

// ExponentialInOutEaseProvider exposes the repository's historical two-branch exponential formula.
type ExponentialInOutEaseProvider struct{}

// New returns the legacy two-branch exponential formula. Its endpoint behavior
// is intentionally left unchanged because animations may depend on exact values.
func (*ExponentialInOutEaseProvider) New(_ []float64) func(float64) float64 {
	expo := func(v float64) float64 {
		v *= 2
		if v < 1 {
			return 0.5 * math.Pow(2, 10*(v-1))
		} else {
			return 0.5 * math.Pow(2, -10*(v-1))
		}
	}

	return expo
}
