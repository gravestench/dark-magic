package input

import "testing"

func TestKeyStateReturnsOnlyRequestedKey(t *testing.T) {
	service := &Service{keyStates: map[int32]InputState{
		1: StateDown,
		2: StateReleased,
	}}

	if got := service.KeyState(1); got != StateDown {
		t.Fatalf("KeyState(1) = %v, want %v", got, StateDown)
	}
	if got := service.KeyState(3); got != StateUp {
		t.Fatalf("KeyState(3) = %v, want zero-value StateUp", got)
	}
}
