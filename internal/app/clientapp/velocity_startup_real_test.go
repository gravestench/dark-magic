package clientapp

import "testing"

// The location schema is Lua-owned and only exists after authority startup.
// Offline assembly must reach that startup before installing native movement.
func TestOfflineVelocityStartupWaitsForAuthoritativeSchemas(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{})
	fixture.advanceOffline(t, 1)
	checkpoint, err := fixture.app.offlineSession.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if checkpoint.Tick == 0 {
		t.Fatal("offline authority did not complete its first tick")
	}
}
