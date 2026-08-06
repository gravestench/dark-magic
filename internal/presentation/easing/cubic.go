package easing

var _ EaseFunctionProvider = &CubicOutEaseProvider{}
var _ EaseFunctionProvider = &CubicInEaseProvider{}
var _ EaseFunctionProvider = &CubicInOutEaseProvider{}

type CubicOutEaseProvider struct{}

func (*CubicOutEaseProvider) New(_ []float64) func(float64) float64 {
	cubic := func(v float64) float64 {
		return v*v*v + 1
	}

	return cubic
}

type CubicInEaseProvider struct{}

func (*CubicInEaseProvider) New(_ []float64) func(float64) float64 {
	cubic := func(v float64) float64 {
		return v * v * v
	}

	return cubic
}

type CubicInOutEaseProvider struct{}

func (*CubicInOutEaseProvider) New(_ []float64) func(float64) float64 {
	cubic := func(v float64) float64 {
		v *= 2
		if v < 1 {
			return 0.5 * v * v * v
		} else {
			return -0.5 * ((v-2)*v*v + 2)
		}
	}

	return cubic
}
