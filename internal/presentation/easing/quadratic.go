package easing

var _ EaseFunctionProvider = &QuadraticOutEaseProvider{}
var _ EaseFunctionProvider = &QuadraticInEaseProvider{}
var _ EaseFunctionProvider = &QuadraticInOutEaseProvider{}

type QuadraticOutEaseProvider struct{}

func (*QuadraticOutEaseProvider) New(_ []float64) func(float64) float64 {
	quadratic := func(v float64) float64 {
		return v - (2 * v)
	}

	return quadratic
}

type QuadraticInEaseProvider struct{}

func (*QuadraticInEaseProvider) New(_ []float64) func(float64) float64 {
	quadratic := func(v float64) float64 {
		return v * v
	}

	return quadratic
}

type QuadraticInOutEaseProvider struct{}

func (*QuadraticInOutEaseProvider) New(_ []float64) func(float64) float64 {
	quadratic := func(v float64) float64 {
		v *= 2
		if v < 1 {
			return 0.5 * v * v
		} else {
			return -0.5 * (v*(v-2) - 1)
		}
	}

	return quadratic
}
