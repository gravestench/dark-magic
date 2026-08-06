package easing

const (
	defaultSteps = 1
)

var _ EaseFunctionProvider = &SteppedEaseProvider{}

type SteppedEaseProvider struct{}

func (*SteppedEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureSteppedParams(params)
	steps := params[0]

	linear := func(v float64) float64 {
		if v <= 0 {
			return 0
		} else if v >= 1 {
			return 1
		}

		return ((steps * v) + 1) * (1 / steps)
	}

	return linear
}

func ensureSteppedParams(params []float64) []float64 {
	if params == nil {
		params = []float64{defaultSteps}
	}

	switch len(params) {
	case 0:
		params = []float64{defaultSteps}
	}

	return params
}
