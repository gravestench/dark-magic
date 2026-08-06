package input

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/inputstate"
)

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
