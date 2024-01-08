package easing

import (
	"math"
)

var _ EaseFunctionProvider = &ExponentialOutEaseProvider{}
var _ EaseFunctionProvider = &ExponentialInEaseProvider{}
var _ EaseFunctionProvider = &ExponentialInOutEaseProvider{}

type ExponentialOutEaseProvider struct{}

func (*ExponentialOutEaseProvider) New(_ []float64) func(float64) float64 {
	expo := func(v float64) float64 {
		return 1 - math.Pow(2, -10*v)
	}

	return expo
}

type ExponentialInEaseProvider struct{}

func (*ExponentialInEaseProvider) New(_ []float64) func(float64) float64 {
	expo := func(v float64) float64 {
		return math.Pow(2, 10*(v-1)) - 0.001
	}

	return expo
}

type ExponentialInOutEaseProvider struct{}

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
