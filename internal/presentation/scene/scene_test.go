package scene

import (
	"bytes"
	"testing"
)

// TestMovementCameraBoundsAndPersistence protects the invariant that save/load retains the clamped tracked view.
func TestMovementCameraBoundsAndPersistence(t *testing.T) {
	state := New(42, 100, 80)
	state.MoveHero(75, -100)

	if state.Hero != (Point{X: 100, Y: 0}) || state.Camera != state.Hero {
		t.Fatalf("unexpected tracked position: %+v", state)
	}

	var saved bytes.Buffer
	if err := state.Save(&saved); err != nil {
		t.Fatal(err)
	}

	restored, err := Load(&saved)
	if err != nil {
		t.Fatal(err)
	}

	if *restored != *state {
		t.Fatalf("restored scene differs: got %+v want %+v", restored, state)
	}
}
