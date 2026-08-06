package easing

var _ EaseFunctionProvider = &LinearEaseProvider{}

type LinearEaseProvider struct{}

func (*LinearEaseProvider) New(_ []float64) func(float64) float64 {
	linear := func(v float64) float64 {
		return v
	}

	return linear
}
