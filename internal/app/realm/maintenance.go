package realm

import (
	"context"
	"errors"
	"time"
)

type MaintenanceResult struct {
	PrunedSessions  int
	PrunedPresence  int
	ReconciledGames int
	Err             error
}

// RunMaintenance owns periodic ephemeral cleanup, lease renewal, worker health,
// and reconnect-expiry reconciliation below the composition root. Observation
// is optional and must not alter lifecycle decisions.
func RunMaintenance(ctx context.Context, control *ControlPlane, reconcileWorkers bool, interval time.Duration, observe func(MaintenanceResult)) {
	if ctx == nil || control == nil {
		return
	}
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result := MaintenanceResult{}
			result.PrunedSessions, result.Err = control.PruneExpiredSessions(ctx)
			var presenceErr error
			result.PrunedPresence, presenceErr = control.PruneInactivePresence(ctx)
			result.Err = errors.Join(result.Err, presenceErr)
			if reconcileWorkers {
				var reconcileErr error
				result.ReconciledGames, reconcileErr = control.ReconcileGames(ctx)
				result.Err = errors.Join(result.Err, reconcileErr)
			}
			if observe != nil {
				observe(result)
			}
		}
	}
}
