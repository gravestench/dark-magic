package main

import (
	"reflect"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// recordingKeyState exposes both configured key state and the order in which production movement consulted it.
type recordingKeyState struct {
	pressed map[int32]bool
	queries []int32
}

// isDown records polling order and returns configured key state so tests can enforce input timing implications.
func (state *recordingKeyState) isDown(key int32) bool {
	state.queries = append(state.queries, key)

	return state.pressed[key]
}

// TestHeroMovementPreservesPollingOrder verifies short-circuit queries and opposing-direction cancellation together.
func TestHeroMovementPreservesPollingOrder(t *testing.T) {
	keys := &recordingKeyState{
		pressed: map[int32]bool{
			rl.KeyA:     true,
			rl.KeyRight: true,
			rl.KeyW:     true,
			rl.KeyS:     true,
		},
	}

	dx, dy := heroMovement(10, keys.isDown)
	if dx != 0 || dy != 0 {
		t.Fatalf("movement = (%v, %v), want opposing directions to cancel", dx, dy)
	}

	expectedQueries := []int32{rl.KeyA, rl.KeyD, rl.KeyRight, rl.KeyW, rl.KeyS}
	if !reflect.DeepEqual(keys.queries, expectedQueries) {
		t.Fatalf("key queries = %v, want %v", keys.queries, expectedQueries)
	}
}

// TestHeroMovementCombinesAxes ensures a diagonal uses the full per-frame distance independently on both axes.
func TestHeroMovementCombinesAxes(t *testing.T) {
	keys := &recordingKeyState{
		pressed: map[int32]bool{
			rl.KeyLeft: true,
			rl.KeyDown: true,
		},
	}

	dx, dy := heroMovement(12.5, keys.isDown)
	if dx != -12.5 || dy != 12.5 {
		t.Fatalf("movement = (%v, %v), want (-12.5, 12.5)", dx, dy)
	}
}
