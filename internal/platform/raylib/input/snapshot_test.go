package input

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/inputstate"
)

// TestMergeActionStatesMakesAnySkipSourceVisible verifies the logical skip action retains every phase reported by its
// keyboard, mouse, or gamepad sources instead of allowing the last source to overwrite earlier input.
func TestMergeActionStatesMakesAnySkipSourceVisible(t *testing.T) {
	got := mergeActionStates(
		inputstate.ActionState{Down: true},
		inputstate.ActionState{Pressed: true},
		inputstate.ActionState{Released: true},
	)
	if !got.Down || !got.Pressed || !got.Released {
		t.Fatalf("merged action = %#v", got)
	}
}
