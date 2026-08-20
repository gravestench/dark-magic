package raylibRenderer

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// TestUpdateUsesLatestCallbackSnapshot verifies subscriptions publish a complete callback list before frame traversal.
func TestUpdateUsesLatestCallbackSnapshot(t *testing.T) {
	service := &Service{
		rootNode: &node{
			enabled: true,
			local:   rl.MatrixIdentity(),
			world:   rl.MatrixIdentity(),
		},
	}
	firstCalls := 0
	secondCalls := 0

	service.OnFrame(func() { firstCalls++ })

	service.update()
	service.OnFrame(func() { secondCalls++ })
	service.update()

	if firstCalls != 2 || secondCalls != 1 {
		t.Fatalf("callback calls = (%d, %d), want (2, 1)", firstCalls, secondCalls)
	}
}

// TestCallbackCanRegisterCallbackDuringUpdate protects immutable snapshots from mutation during owner-thread iteration.
func TestCallbackCanRegisterCallbackDuringUpdate(t *testing.T) {
	service := &Service{
		rootNode: &node{
			enabled: true,
			local:   rl.MatrixIdentity(),
			world:   rl.MatrixIdentity(),
		},
	}
	nestedCalls := 0

	service.OnFrame(func() {
		service.OnFrame(func() { nestedCalls++ })
	})

	service.update()

	if nestedCalls != 0 {
		t.Fatalf("new callback ran in registration frame")
	}

	service.update()

	if nestedCalls != 1 {
		t.Fatalf("new callback calls = %d, want 1", nestedCalls)
	}
}
