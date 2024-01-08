package easing

var _ EaseFunctionProvider = &QuarticOutEaseProvider{}
var _ EaseFunctionProvider = &QuarticInEaseProvider{}
var _ EaseFunctionProvider = &QuarticInOutEaseProvider{}

type QuarticOutEaseProvider struct{}

func (*QuarticOutEaseProvider) New(_ []float64) func(float64) float64 {
	quartic := func(v float64) float64 {
		return 1 - v*v*v*v
	}

	return quartic
}

type QuarticInEaseProvider struct{}

func (*QuarticInEaseProvider) New(_ []float64) func(float64) float64 {
	quartic := func(v float64) float64 {
		return v * v * v * v
	}

	return quartic
}

type QuarticInOutEaseProvider struct{}

func (*QuarticInOutEaseProvider) New(_ []float64) func(float64) float64 {
	quartic := func(v float64) float64 {
		v *= 2
		if v < 1 {
			return 0.5 * v * v * v * v
		} else {
			return -0.5 * ((v-2)*v*v*v - 2)
		}
	}

	return quartic
}
