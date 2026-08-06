package easing

import (
	"math"

	"github.com/gravestench/mathlib"
)

var _ EaseFunctionProvider = &SineOutEaseProvider{}
var _ EaseFunctionProvider = &SineInEaseProvider{}
var _ EaseFunctionProvider = &SineInOutEaseProvider{}

type SineOutEaseProvider struct{}

func (*SineOutEaseProvider) New(_ []float64) func(float64) float64 {
	sine := func(v float64) float64 {
		if v <= mathlib.Epsilon {
			return 0
		} else if math.Abs(1-v) <= mathlib.Epsilon {
			return 1
		}

		return math.Sin(v * mathlib.TAU)
	}

	return sine
}

type SineInEaseProvider struct{}

func (*SineInEaseProvider) New(_ []float64) func(float64) float64 {
	sine := func(v float64) float64 {
		if v <= mathlib.Epsilon {
			return 0
		} else if math.Abs(1-v) <= mathlib.Epsilon {
			return 1
		}

		return 1 - math.Cos(v*mathlib.TAU)
	}

	return sine
}

type SineInOutEaseProvider struct{}

func (*SineInOutEaseProvider) New(_ []float64) func(float64) float64 {
	sine := func(v float64) float64 {
		if v <= mathlib.Epsilon {
			return 0
		} else if math.Abs(1-v) <= mathlib.Epsilon {
			return 1
		}

		return 0.5 * (1 - math.Cos(mathlib.PI*v))
	}

	return sine
}
