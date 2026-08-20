package easing

var _ EaseFunctionProvider = &LinearEaseProvider{}

// LinearEaseProvider returns progress unchanged for constant-speed motion.
type LinearEaseProvider struct{}

// New returns the identity curve; linear easing has no tunable parameters.
func (*LinearEaseProvider) New(_ []float64) func(float64) float64 {
	// linear keeps a named closure so every provider exposes the same factory
	// shape, even when its formula is only the identity operation.
	linear := func(v float64) float64 {
		return v
	}

	return linear
}
