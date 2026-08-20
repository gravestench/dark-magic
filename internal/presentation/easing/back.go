package easing

const (
	defaultOvershoot          = 1.70158
	magicOvershootInOutScalar = 1.525
)

var _ EaseFunctionProvider = &BackOutEaseProvider{}
var _ EaseFunctionProvider = &BackInEaseProvider{}
var _ EaseFunctionProvider = &BackInOutEaseProvider{}

// BackOutEaseProvider overshoots the destination before settling at it.
type BackOutEaseProvider struct{}

// New builds a back-out curve using the caller's overshoot or the standard
// default, preserving the authored animation's amount of rebound.
func (*BackOutEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureBackParams(params)
	overshoot := params[0]
	back := func(v float64) float64 {
		return v*v*((overshoot+1)*v+overshoot) + 1
	}

	return back
}

// BackInEaseProvider initially moves away from the destination before advancing.
type BackInEaseProvider struct{}

// New builds a back-in curve with the same parameter convention as back-out.
func (*BackInEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureBackParams(params)
	overshoot := params[0]
	back := func(v float64) float64 {
		return v * v * ((overshoot+1)*v - overshoot)
	}

	return back
}

// BackInOutEaseProvider applies symmetric anticipation and overshoot.
type BackInOutEaseProvider struct{}

// New builds the two-half back curve; its scalar matches the conventional
// in-out shape and is therefore part of visual compatibility.
func (*BackInOutEaseProvider) New(params []float64) func(float64) float64 {
	params = ensureBackParams(params)
	overshoot := params[0]
	back := func(v float64) float64 {
		// Doubling maps each half of normalized progress onto the full base curve.
		v *= 2

		s := overshoot * magicOvershootInOutScalar
		if v < 1 {
			return 0.5 * (v * v * ((s+1)*v - s))
		}

		return 0.5 * ((v-2)*v*((s+1)*v+s) + 2)
	}

	return back
}

// ensureBackParams supplies the historical overshoot when the parameter list
// is nil or empty; additional values remain ignored for data compatibility.
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
