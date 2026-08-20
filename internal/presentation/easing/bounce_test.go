package easing

import (
	"math"
	"testing"
)

// TestBounceCurvesPreserveHistoricalSamples locks the nonstandard overshoot that authored animations may already use.
func TestBounceCurvesPreserveHistoricalSamples(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		curve  func(float64) float64
		values []float64
	}{
		{
			name:   "out",
			curve:  (&BounceOutEaseProvider{}).New(nil),
			values: []float64{0, 0.47265625, 0.578125, 0.55078125, 1.328125},
		},
		{
			name:   "in",
			curve:  (&BounceInEaseProvider{}).New(nil),
			values: []float64{-0.328125, 0.44921875, 0.421875, 0.52734375, 1},
		},
		{
			name:   "in-out",
			curve:  (&BounceInOutEaseProvider{}).New(nil),
			values: []float64{-0.1640625, 0.2109375, 0.5, 0.7890625, 1.1640625},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			for index, want := range test.values {
				progress := float64(index) / float64(len(test.values)-1)
				if got := test.curve(progress); math.Abs(got-want) > 1e-12 {
					t.Fatalf("curve(%g) = %.16g, want %.16g", progress, got, want)
				}
			}
		})
	}
}
