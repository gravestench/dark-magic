package easing

import "testing"

func TestCubicCurvesAreNormalizedAndMonotonic(t *testing.T) {
	tests := map[string]func(float64) float64{
		"in":     (&CubicInEaseProvider{}).New(nil),
		"out":    (&CubicOutEaseProvider{}).New(nil),
		"in-out": (&CubicInOutEaseProvider{}).New(nil),
	}
	for name, curve := range tests {
		t.Run(name, func(t *testing.T) {
			if curve(0) != 0 || curve(0.5) < 0 || curve(0.5) > 1 || curve(1) != 1 {
				t.Fatalf("endpoints/midpoint = %v, %v, %v", curve(0), curve(0.5), curve(1))
			}
			previous := curve(0)
			for step := 1; step <= 100; step++ {
				current := curve(float64(step) / 100)
				if current < previous || current < 0 || current > 1 {
					t.Fatalf("step %d: previous=%v current=%v", step, previous, current)
				}
				previous = current
			}
		})
	}
}

func TestCubicInOutMidpoint(t *testing.T) {
	if got := (&CubicInOutEaseProvider{}).New(nil)(0.5); got != 0.5 {
		t.Fatalf("midpoint = %v", got)
	}
}
