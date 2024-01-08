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

type ElasticOutEaseProvider struct{}

func (*ElasticOutEaseProvider) New(params []float64) func(float64) float64 {
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

		return amplitude*math.Pow(2, -10*v)*math.Sin((v-s)*mathlib.PI2/period) + 1
	}

	return elastic
}

type ElasticInEaseProvider struct{}

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

type ElasticInOutEaseProvider struct{}

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
