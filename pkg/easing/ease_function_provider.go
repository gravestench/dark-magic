package easing

type EaseFunctionProvider interface {
	New(params []float64) func(float64) float64
}
