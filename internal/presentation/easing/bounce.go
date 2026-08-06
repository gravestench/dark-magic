package easing

// please, god, forgive my sins...
const (
	magic1  = 2.75
	magic2  = 7.5625
	magic3  = 2.5
	magic4  = 0.9375
	magic5  = 2.25
	magic6  = 2.625
	magic7  = 0.984375
	magic8  = 1.5
	magic9  = 0.75
	magic10 = 2
)

var _ EaseFunctionProvider = &BounceOutEaseProvider{}
var _ EaseFunctionProvider = &BounceInEaseProvider{}
var _ EaseFunctionProvider = &BounceInOutEaseProvider{}

type BounceOutEaseProvider struct{}

func (*BounceOutEaseProvider) New(_ []float64) func(float64) float64 {
	bounce := func(v float64) float64 {
		if v < 1/magic1 {
			return magic2 * v * v
		} else if v < magic10/magic1 {
			return magic2*(v-magic8/magic1)*v + magic9
		} else if v < magic3/magic1 {
			return magic2*(v-magic5/magic1)*v + magic4
		}

		return magic2*(v-magic6/magic1)*v + magic7
	}

	return bounce
}

type BounceInEaseProvider struct{}

func (*BounceInEaseProvider) New(_ []float64) func(float64) float64 {
	bounce := func(v float64) float64 {
		v = 1 - v

		if v < 1/magic1 {
			return 1 - (magic2 * v * v)
		} else if v < 2/magic1 {
			return 1 - (magic2*(v-magic8/magic1)*v + magic9)
		} else if v < magic3/magic1 {
			return 1 - (magic2*(v-magic5/magic1)*v + magic4)
		}

		return 1 - (magic2*(v-magic6/magic1)*v + magic7)
	}

	return bounce
}

type BounceInOutEaseProvider struct{}

func (*BounceInOutEaseProvider) New(_ []float64) func(float64) float64 {
	bounce := func(v float64) float64 {
		var reverse = false

		if v < 0.5 {
			v = 1 - (v * 2)
			reverse = true
		} else {
			v = (v * 2) - 1
		}

		if v < 1/magic1 {
			v = magic2 * v * v
		} else if v < magic10/magic1 {
			v = magic2*(v-magic8/magic1)*v + magic9
		} else if v < magic3/magic1 {
			v = magic2*(v-magic5/magic1)*v + magic4
		} else {
			v = magic2*(v-magic6/magic1)*v + magic7
		}

		if reverse {
			return (1 - v) * 0.5
		} else {
			return v*0.5 + 0.5
		}
	}

	return bounce
}
