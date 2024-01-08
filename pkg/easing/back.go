package easing

const (
	defaultOvershoot          = 1.70158
	magicOvershootInOutScalar = 1.525
)

var _ EaseFunctionProvider = &BackOutEaseProvider{}
var _ EaseFunctionProvider = &BackInEaseProvider{}
var _ EaseFunctionProvider = &BackInOutEaseProvider{}

type BackOutEaseProvider struct{}

func (*BackOutEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureBackParams(params)
	overshoot := params[0]
	back := func(v float64) float64 {
		return v*v*((overshoot+1)*v+overshoot) + 1
	}

	return back
}

type BackInEaseProvider struct{}

func (*BackInEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureBackParams(params)
	overshoot := params[0]
	back := func(v float64) float64 {
		return v * v * ((overshoot+1)*v - overshoot)
	}

	return back
}

type BackInOutEaseProvider struct{}

func (*BackInOutEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureBackParams(params)
	overshoot := params[0]
	back := func(v float64) float64 {
		v *= 2
		s := overshoot * magicOvershootInOutScalar
		if v < 1 {
			return 0.5 * (v * v * ((s+1)*v - s))
		}

		return 0.5 * ((v-2)*v*((s+1)*v+s) + 2)
	}

	return back
}

func ensureBackParams(params []float64) []float64 {
	if params == nil {
		params = []float64{defaultOvershoot}
	}

	switch len(params) {
	case 0:
		params = []float64{defaultOvershoot}
	}

	return params
}
