package input

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/inputcore"
)

func TestMergeActionStatesMakesAnySkipSourceVisible(t *testing.T) {
	got := mergeActionStates(
		inputcore.ActionState{Down: true},
		inputcore.ActionState{Pressed: true},
		inputcore.ActionState{Released: true},
	)
	if !got.Down || !got.Pressed || !got.Released {
		t.Fatalf("merged action = %#v", got)
	}
}
