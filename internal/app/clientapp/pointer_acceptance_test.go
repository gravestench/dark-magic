package clientapp

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/inputstate"
)

func TestPointerMovementAcceptanceUsesOneClickAndWaitsForSettledAuthority(t *testing.T) {
	fixture := &pointerMovementAcceptance{cursorX: 500, cursorY: 250}
	original := inputstate.Frame{Actions: map[string]inputstate.ActionState{"inventory": {Pressed: true}}}

	clicked := fixture.Frame(original, 10, 20, true)
	if clicked.CursorX != 500 || clicked.CursorY != 250 || !clicked.Actions["pointer_primary"].Pressed {
		t.Fatalf("injected pointer frame = %#v", clicked)
	}
	if !original.Actions["inventory"].Pressed || original.Actions["pointer_primary"].Pressed {
		t.Fatal("fixture mutated the native input snapshot")
	}

	fixture.Frame(inputstate.Frame{}, 12, 20, true)
	if !fixture.moved || !fixture.Busy() {
		t.Fatal("authoritative displacement did not begin acceptance")
	}
	for range pointerAcceptanceStableFrames {
		frame := fixture.Frame(inputstate.Frame{}, 12, 20, true)
		if frame.Actions["pointer_primary"].Pressed {
			t.Fatal("fixture injected more than one click")
		}
	}
	if fixture.Busy() {
		t.Fatal("settled authoritative movement did not complete acceptance")
	}
}

func TestPointerMovementAcceptanceWaitsForPlayerAdmission(t *testing.T) {
	fixture := &pointerMovementAcceptance{cursorX: 500, cursorY: 250}
	frame := fixture.Frame(inputstate.Frame{}, 0, 0, false)
	if fixture.clicked || frame.Actions["pointer_primary"].Pressed || !fixture.Busy() {
		t.Fatalf("pre-admission fixture = %#v, %#v", fixture, frame)
	}
}
