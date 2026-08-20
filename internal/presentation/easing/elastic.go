package easing

import (
	"math"

	"github.com/gravestench/mathlib"
)

const (
	defaultAmplitude = 0.1
	defaultPeriod    = 0.1
)

var _ EaseFunctionProvider = &ElasticOutEaseProvider{}
var _ EaseFunctionProvider = &ElasticInEaseProvider{}
var _ EaseFunctionProvider = &ElasticInOutEaseProvider{}

// ElasticOutEaseProvider overshoots and oscillates while approaching the destination.
type ElasticOutEaseProvider struct{}

// New builds an elastic-out curve from amplitude and period parameters.
func (*ElasticOutEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureElasticParams(params)
	amplitude, period := params[0], params[1]
	elastic := func(v float64) float64 {
		// Exact endpoints prevent the oscillation formula from leaving tiny
		// floating-point remnants at the beginning or end of a transition.
		if math.Abs(0-v) < math.SmallestNonzeroFloat64 {
			return 0
		} else if math.Abs(1-v) < math.SmallestNonzeroFloat64 {
			return 1
		}

		s := period / 4

		// The phase shift uses asin(1/amplitude), so the legacy clamp keeps that calculation in the real-number domain.
		if amplitude < 1 {
			amplitude = 1
		} else {
			s = period * math.Asin(1/amplitude) / mathlib.PI2
		}

		return amplitude*math.Pow(2, -10*v)*math.Sin((v-s)*mathlib.PI2/period) + 1
	}

	return elastic
}

// ElasticInEaseProvider oscillates before accelerating toward the destination.
type ElasticInEaseProvider struct{}

// New builds an elastic-in curve using the same parameter semantics as out.
func (*ElasticInEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureElasticParams(params)
	amplitude, period := params[0], params[1]
	elastic := func(v float64) float64 {
		if math.Abs(0-v) < math.SmallestNonzeroFloat64 {
			return 0
		} else if math.Abs(1-v) < math.SmallestNonzeroFloat64 {
			return 1
		}

		s := period / 4

		if amplitude < 1 {
			amplitude = 1
		} else {
			s = period * math.Asin(1/amplitude) / mathlib.PI2
		}

		return -(amplitude * math.Pow(2, 10*(v-1)) * math.Sin((v-s)*mathlib.PI2/period))
	}

	return elastic
}

// ElasticInOutEaseProvider applies the legacy two-branch elastic formula around normalized midpoint.
type ElasticInOutEaseProvider struct{}

// New builds a two-half elastic curve while preserving the legacy phase math.
func (*ElasticInOutEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureElasticParams(params)
	amp, period := params[0], params[1]
	elastic := func(v float64) float64 {
		if math.Abs(0-v) < math.SmallestNonzeroFloat64 {
			return 0
		} else if math.Abs(1-v) < math.SmallestNonzeroFloat64 {
			return 1
		}

		s := period / 4

		if amp < 1 {
			amp = 1
		} else {
			s = period * math.Asin(1/amp) / mathlib.PI2
		}

		v *= 2
		if v < 1 {
			return -0.5 * (amp * math.Pow(2, 10*(v-1)) * math.Sin((v-s)*mathlib.PI2/period))
		}

		return amp*math.Pow(2, -10*(v-1))*math.Sin((v-s)*mathlib.PI2/period)*0.5 + 1
	}

	return elastic
}

// ensureElasticParams supplies missing amplitude and period values without
// rejecting extra values that older authored content may include.
func ensureElasticParams(params []float64) []float64 {
	if params == nil {
		params = []float64{defaultAmplitude, defaultPeriod}
	}

	switch len(params) {
	case 0:
		params = []float64{defaultAmplitude, defaultPeriod}
	case 1:
		params = append(params, defaultPeriod)
	}

	return params
}
