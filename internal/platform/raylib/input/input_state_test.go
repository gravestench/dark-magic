package input

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TestKeyStateReturnsOnlyRequestedKey verifies direct lookup neither leaks another key's state nor needs a special
// missing-key branch because StateUp is deliberately the zero value.
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

// TestRetainLogicalCursorOutsideViewport verifies letterbox motion retains the last usable logical coordinate while
// motion inside the game surface replaces it.
func TestRetainLogicalCursorOutsideViewport(t *testing.T) {
	x, y := retainLogicalCursor(320, 240, 0, 0, false)
	if x != 320 || y != 240 {
		t.Fatalf("outside cursor = %d,%d, want last valid 320,240", x, y)
	}

	x, y = retainLogicalCursor(320, 240, 17, 29, true)
	if x != 17 || y != 29 {
		t.Fatalf("inside cursor = %d,%d, want 17,29", x, y)
	}
}

// TestHeldKeyRepeatAppliesOnlyToNavigationAndEditing protects text-navigation ergonomics without turning activation,
// cancellation, or function keys into repeated actions.
func TestHeldKeyRepeatAppliesOnlyToNavigationAndEditing(t *testing.T) {
	for _, key := range []int32{rl.KeyLeft, rl.KeyDown, rl.KeyPageUp, rl.KeyHome, rl.KeyBackspace, rl.KeyDelete} {
		if !repeatsWhenHeld(key) {
			t.Fatalf("key %d should repeat", key)
		}
	}

	for _, key := range []int32{rl.KeyEnter, rl.KeyEscape, rl.KeySpace, rl.KeyF1} {
		if repeatsWhenHeld(key) {
			t.Fatalf("key %d should remain one-shot", key)
		}
	}
}
