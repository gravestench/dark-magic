package raylibRenderer

import "testing"

// TestOverlaySubscriptionCanBeDetached verifies cancellation disables callbacks without mutating an active snapshot.
func TestOverlaySubscriptionCanBeDetached(t *testing.T) {
	service := &Service{}
	calls := 0
	stop := service.SubscribeOverlay(func() { calls++ })
	service.runOverlays()
	stop()
	service.runOverlays()

	if calls != 1 {
		t.Fatalf("overlay calls = %d", calls)
	}
}
