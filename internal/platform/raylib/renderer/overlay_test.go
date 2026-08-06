package raylibRenderer

import "testing"

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
