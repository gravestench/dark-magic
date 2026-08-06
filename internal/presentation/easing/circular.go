package easing

import (
	"math"
)

var _ EaseFunctionProvider = &CircularOutEaseProvider{}
var _ EaseFunctionProvider = &CircularInEaseProvider{}
var _ EaseFunctionProvider = &CircularInOutEaseProvider{}

type CircularOutEaseProvider struct{}

func (*CircularOutEaseProvider) New(_ []float64) func(float64) float64 {
	circ := func(v float64) float64 {
		return math.Sqrt(1 - (v * v))
	}

	return circ
}

type CircularInEaseProvider struct{}

func (*CircularInEaseProvider) New(_ []float64) func(float64) float64 {
	circ := func(v float64) float64 {
		return 1 - math.Sqrt(1-v*v)
	}

	return circ
}

type CircularInOutEaseProvider struct{}

func (*CircularInOutEaseProvider) New(_ []float64) func(float64) float64 {
	circ := func(v float64) float64 {
		v *= 2
		if v < 1 {
			return -0.5 * (math.Sqrt(1-v*v) - 1)
		}

		return 0.5 * (math.Sqrt(1-(v-2)*v) + 1)
	}

	return circ
}
