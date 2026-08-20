package easing

// EaseFunctionProvider binds a named easing family to a factory. Parameters are
// family-specific, and the returned function expects normalized progress.
type EaseFunctionProvider interface {
	New(params []float64) func(float64) float64
}
