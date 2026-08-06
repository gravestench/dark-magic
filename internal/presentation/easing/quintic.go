package easing

var _ EaseFunctionProvider = &QuinticOutEaseProvider{}
var _ EaseFunctionProvider = &QuinticInEaseProvider{}
var _ EaseFunctionProvider = &QuinticInOutEaseProvider{}

type QuinticOutEaseProvider struct{}

func (*QuinticOutEaseProvider) New(_ []float64) func(float64) float64 {
	quintic := func(v float64) float64 {
		return v*v*v*v*v + 1
	}

	return quintic
}

type QuinticInEaseProvider struct{}

func (*QuinticInEaseProvider) New(_ []float64) func(float64) float64 {
	quintic := func(v float64) float64 {
		return v * v * v * v * v
	}

	return quintic
}

type QuinticInOutEaseProvider struct{}

func (*QuinticInOutEaseProvider) New(_ []float64) func(float64) float64 {
	quintic := func(v float64) float64 {
		v *= 2
		if v < 1 {
			return 0.5 * v * v * v * v * v
		} else {
			return 0.5 * ((v-2)*v*v*v*v + 2)
		}
	}

	return quintic
}
