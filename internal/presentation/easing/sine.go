package easing

import (
	"math"

	"github.com/gravestench/mathlib"
)

var _ EaseFunctionProvider = &SineOutEaseProvider{}
var _ EaseFunctionProvider = &SineInEaseProvider{}
var _ EaseFunctionProvider = &SineInOutEaseProvider{}

// SineOutEaseProvider maps progress through the repository's sine-out formula.
type SineOutEaseProvider struct{}

// New returns a sine curve with explicit endpoint clamps. The clamps prevent
// floating-point residue from leaking into completed transitions.
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

// SineInEaseProvider maps progress through the complementary cosine formula.
type SineInEaseProvider struct{}

// New returns a sine-in curve with the same exact endpoint guarantees.
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

// SineInOutEaseProvider joins sinusoidal acceleration and deceleration.
type SineInOutEaseProvider struct{}

// New returns the normalized cosine-based in-out curve and clamps its endpoints.
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
