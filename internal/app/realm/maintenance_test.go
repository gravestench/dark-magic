package realm

import (
	"context"
	"testing"
	"time"
)

// TestRunMaintenancePrunesExpiredSessionsAndStopsWithContext verifies run maintenance prunes expired sessions and
// stops with context. The scenario keeps the maintenance contract visible to maintainers.
func TestRunMaintenancePrunesExpiredSessionsAndStopsWithContext(t *testing.T) {
	control, err := NewControlPlane(ControlPlaneConfig{SessionLifetime: time.Minute})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Unix(100, 0)

	control.accounts.(*Accounts).now = func() time.Time { return now }
	if _, err := control.CreateAccount(t.Context(), "Alice", "long enough password"); err != nil {
		t.Fatal(err)
	}

	if _, err := control.Authenticate(t.Context(), "Alice", "long enough password"); err != nil {
		t.Fatal(err)
	}

	now = now.Add(time.Minute)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	results := make(chan MaintenanceResult, 1)

	go func() {
		RunMaintenance(ctx, control, false, time.Millisecond, func(result MaintenanceResult) {
			select {
			case results <- result:
			default:
			}
		})
		close(done)
	}()

	select {
	case result := <-results:
		if result.Err != nil || result.PrunedSessions != 1 || result.ReconciledGames != 0 {
			t.Fatalf("maintenance result = %#v", result)
		}
	case <-time.After(time.Second):
		t.Fatal("maintenance did not run")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("maintenance did not stop")
	}
}
