package realm

import (
	"context"
	"errors"
	"time"
)

const workerFailureThreshold = 3

// ReconcileGames renews healthy memberships and checkpoints, restores workers
// only after consecutive failures, and fails closed when restoration is unsafe.
func (control *ControlPlane) ReconcileGames(ctx context.Context) (int, error) {
	if control == nil || ctx == nil || control.allocator == nil || control.admissions == nil {
		return 0, ErrGameUnavailable
	}

	reconciled := 0
	var result error
	gameIDs, listErr := control.games.gameIDs(ctx)
	if listErr != nil {
		return 0, listErr
	}
	for _, gameID := range gameIDs {
		worker, found := control.allocator.Game(gameID)
		healthy := found
		var status WorkerStatus
		if found {
			var err error
			status, err = worker.Status(ctx)
			healthy = err == nil && status.Ready
		}
		if healthy {
			control.clearHealthFailure(gameID)
			result = errors.Join(result, control.allocations.Healthy(ctx, gameID))
			result = errors.Join(result, control.checkpointGame(ctx, gameID, worker))
			_, renewErr := control.admissions.RenewGameMemberships(ctx, gameID)
			result = errors.Join(result, renewErr)
			for _, playerID := range status.ExpiredPlayers {
				result = errors.Join(result, control.reconcileExpiredPlayer(ctx, gameID, playerID))
			}
			continue
		}
		if control.noteHealthFailure(gameID) < workerFailureThreshold {
			continue
		}

		control.departureFlowMu.Lock()
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		if _, directoryErr := control.games.admissionDetail(
			cleanupCtx,
			gameID,
		); errors.Is(directoryErr, ErrGameNotFound) {
			cancel()
			control.departureFlowMu.Unlock()
			control.clearHealthFailure(gameID)
			continue
		}

		restoreCtx, stopRestore := context.WithTimeout(context.WithoutCancel(ctx), 45*time.Second)
		restoreErr := control.restoreAllocatedGame(restoreCtx, gameID)
		stopRestore()
		if restoreErr == nil {
			cancel()
			control.departureFlowMu.Unlock()
			control.recordAudit(ctx, AuditEvent{Operation: AuditGameRestore, GameID: gameID}, nil)
			control.clearHealthFailure(gameID)
			reconciled++
			continue
		}

		err := control.admissions.AbandonGame(cleanupCtx, gameID)
		if releaseErr := control.allocator.Release(cleanupCtx, gameID); releaseErr != nil &&
			!errors.Is(releaseErr, ErrGameNotFound) {
			err = errors.Join(err, releaseErr)
		}
		err = errors.Join(err, control.games.Remove(cleanupCtx, gameID))
		err = errors.Join(
			err,
			control.allocations.Fail(cleanupCtx, gameID, errors.Join(ErrWorker, err)),
		)
		cancel()
		control.departureFlowMu.Unlock()

		control.recordAudit(ctx, AuditEvent{Operation: AuditGameReconcile, GameID: gameID}, err)
		control.clearHealthFailure(gameID)
		result = errors.Join(result, err)
		reconciled++
	}
	return reconciled, result
}

// restoreAllocatedGame replaces a failed live worker from the checkpoint tied
// to its exact allocation generation and runtime identity.
func (control *ControlPlane) restoreAllocatedGame(
	ctx context.Context,
	gameID string,
) (err error) {
	restorer, supported := control.allocator.(GameRestorer)
	if !supported {
		return ErrWorker
	}
	allocation, err := control.allocations.Get(ctx, gameID)
	if err != nil || allocation.State != AllocationReady {
		return errors.Join(err, ErrAllocationRecord)
	}
	checkpoint, err := control.checkpoints.Latest(ctx, gameID)
	if err != nil || checkpoint.AllocationID != allocation.AllocationID {
		return errors.Join(err, ErrGameCheckpoint)
	}
	players, err := control.membershipStore.ActivePlayerIDs(ctx, gameID)
	if err != nil || len(players) == 0 {
		return errors.Join(err, ErrMembership)
	}
	recovery, err := NewGameRecovery(checkpoint.Checkpoint, players)
	if err != nil {
		return err
	}
	if releaseErr := control.allocator.Release(ctx, gameID); releaseErr != nil &&
		!errors.Is(releaseErr, ErrGameNotFound) {
		return releaseErr
	}

	replacement, err := restorer.Restore(ctx, GameSpec{
		GameID:       gameID,
		AllocationID: allocation.AllocationID,
	}, recovery)
	if err != nil {
		return err
	}
	installed := true
	defer func() {
		if err != nil && installed {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			err = errors.Join(err, control.allocator.Release(cleanupCtx, gameID))
			cancel()
		}
	}()

	description, err := replacement.Worker.Describe(ctx)
	if err != nil {
		return err
	}
	wantHash, wantErr := allocation.Runtime.Digest()
	gotHash, gotErr := description.Runtime.Digest()
	if wantErr != nil ||
		gotErr != nil ||
		wantHash != gotHash ||
		gotHash != checkpoint.IdentityHash ||
		description.GameID != gameID ||
		replacement.AllocationID != allocation.AllocationID ||
		replacement.GameID != gameID {
		return ErrWorker
	}
	if _, err = control.allocations.RestoreReady(
		ctx,
		gameID,
		allocation.AllocationID,
		replacement.Endpoint,
		description.Runtime,
	); err != nil {
		return err
	}
	if err = control.admissions.ReplaceGame(
		gameID,
		replacement.Tickets,
		replacement.Endpoint,
	); err != nil {
		return err
	}

	control.checkpointMu.Lock()
	delete(control.lastCheckpoint, gameID)
	control.checkpointMu.Unlock()
	_ = control.checkpointGame(ctx, gameID, replacement.Worker)
	installed = false
	return nil
}

// checkpointGame captures the worker's canonical simulation state only after
// the durable allocation has been confirmed ready. The allocation generation
// and complete runtime identity are pinned into the record so a stale or
// incompatible worker cannot overwrite a replacement authority's checkpoint.
func (control *ControlPlane) checkpointGame(
	ctx context.Context,
	gameID string,
	worker WorkerClient,
) error {
	if control == nil || control.allocations == nil || control.checkpoints == nil || worker == nil {
		return ErrGameUnavailable
	}
	now := time.Now().UTC()
	control.checkpointMu.Lock()
	last := control.lastCheckpoint[gameID]
	control.checkpointMu.Unlock()
	if !last.IsZero() && now.Sub(last) < control.checkpointInterval {
		return nil
	}

	allocation, err := control.allocations.Get(ctx, gameID)
	if err != nil || allocation.State != AllocationReady {
		return errors.Join(err, ErrAllocationRecord)
	}
	identityHash, err := allocation.Runtime.Digest()
	if err != nil {
		return ErrAllocationRecord
	}
	checkpoint, err := worker.Checkpoint(ctx)
	if err != nil {
		return err
	}
	record, err := NewGameCheckpoint(
		gameID,
		allocation.AllocationID,
		identityHash,
		checkpoint,
	)
	if err != nil {
		return err
	}
	_, err = control.checkpoints.Save(ctx, record)
	if err == nil {
		control.checkpointMu.Lock()
		control.lastCheckpoint[gameID] = now
		control.checkpointMu.Unlock()
	}
	return err
}

// RecoverInterruptedGames runs once before the realm accepts traffic. An
// allocator that supports both fencing and restore must positively stop the
// exact surviving allocation generation before checkpoint replacement. Other
// allocators, incomplete handoffs, and failed fences retain the conservative
// fail-closed cleanup path.
func (control *ControlPlane) RecoverInterruptedGames(ctx context.Context) (int, error) {
	if control == nil ||
		ctx == nil ||
		control.allocations == nil ||
		control.games == nil ||
		control.characters == nil {
		return 0, ErrGameUnavailable
	}
	records, err := control.allocations.Active(ctx)
	if err != nil {
		return 0, err
	}

	recovered := 0
	var result error
	for _, record := range records {
		recoveryCtx, stopRecovery := context.WithTimeout(context.WithoutCancel(ctx), 45*time.Second)
		recoveryErr := control.restoreInterruptedGame(recoveryCtx, record)
		stopRecovery()
		if recoveryErr == nil {
			control.recordAudit(
				ctx,
				AuditEvent{Operation: AuditGameRestore, GameID: record.GameID},
				nil,
			)
			recovered++
			continue
		}

		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		var cleanupErr error
		if control.allocator != nil {
			if _, found := control.allocator.Game(record.GameID); found {
				cleanupErr = errors.Join(
					cleanupErr,
					control.allocator.Release(cleanupCtx, record.GameID),
				)
			}
		}
		_, leaseErr := control.characters.ReleaseGame(cleanupCtx, record.GameID)
		cleanupErr = errors.Join(cleanupErr, leaseErr)
		cleanupErr = errors.Join(
			cleanupErr,
			control.membershipStore.AbandonGame(cleanupCtx, record.GameID),
		)
		if directoryErr := control.games.Remove(cleanupCtx, record.GameID); directoryErr != nil &&
			!errors.Is(directoryErr, ErrGameNotFound) {
			cleanupErr = errors.Join(cleanupErr, directoryErr)
		}
		cause := errors.Join(
			errors.New("realm: allocation interrupted by Realm restart"),
			recoveryErr,
			cleanupErr,
		)
		cleanupErr = errors.Join(
			cleanupErr,
			control.allocations.Fail(cleanupCtx, record.GameID, cause),
		)
		cancel()

		control.recordAudit(
			ctx,
			AuditEvent{Operation: AuditGameReconcile, GameID: record.GameID},
			cleanupErr,
		)
		result = errors.Join(result, cleanupErr)
		recovered++
	}
	return recovered, result
}

// restoreInterruptedGame fences a potentially surviving allocation before
// installing a checkpoint replacement, preventing split-brain authority.
func (control *ControlPlane) restoreInterruptedGame(
	ctx context.Context,
	record AllocationRecord,
) (err error) {
	fencer, canFence := control.allocator.(GameFencer)
	restorer, canRestore := control.allocator.(GameRestorer)
	if !canFence || !canRestore || record.State != AllocationReady {
		return ErrWorker
	}
	checkpoint, err := control.checkpoints.Latest(ctx, record.GameID)
	if err != nil || checkpoint.AllocationID != record.AllocationID {
		return errors.Join(err, ErrGameCheckpoint)
	}
	players, err := control.membershipStore.ActivePlayerIDs(ctx, record.GameID)
	if err != nil || len(players) == 0 {
		return errors.Join(err, ErrMembership)
	}
	recovery, err := NewGameRecovery(checkpoint.Checkpoint, players)
	if err != nil {
		return err
	}
	spec := GameSpec{GameID: record.GameID, AllocationID: record.AllocationID}
	if err := fencer.Fence(ctx, spec); err != nil {
		return err
	}
	replacement, err := restorer.Restore(ctx, spec, recovery)
	if err != nil {
		return err
	}

	installed := true
	defer func() {
		if err != nil && installed {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			err = errors.Join(err, control.allocator.Release(cleanupCtx, record.GameID))
			cancel()
		}
	}()

	description, err := replacement.Worker.Describe(ctx)
	if err != nil {
		return err
	}
	wantHash, wantErr := record.Runtime.Digest()
	gotHash, gotErr := description.Runtime.Digest()
	if wantErr != nil ||
		gotErr != nil ||
		wantHash != gotHash ||
		gotHash != checkpoint.IdentityHash ||
		description.GameID != record.GameID ||
		replacement.GameID != record.GameID ||
		replacement.AllocationID != record.AllocationID {
		return ErrWorker
	}
	if _, err = control.allocations.RestoreReady(
		ctx,
		record.GameID,
		record.AllocationID,
		replacement.Endpoint,
		description.Runtime,
	); err != nil {
		return err
	}
	if _, err = control.admissions.ResumeGame(
		ctx,
		record.GameID,
		replacement.Tickets,
		replacement.Endpoint,
	); err != nil {
		return err
	}

	control.checkpointMu.Lock()
	delete(control.lastCheckpoint, record.GameID)
	control.checkpointMu.Unlock()
	_ = control.checkpointGame(ctx, record.GameID, replacement.Worker)
	installed = false
	return nil
}

// noteHealthFailure increments a game's consecutive failure count under the
// lifecycle lock so concurrent maintenance passes share one threshold.
func (control *ControlPlane) noteHealthFailure(gameID string) int {
	control.lifecycleMu.Lock()
	defer control.lifecycleMu.Unlock()
	control.healthFailures[gameID]++
	return control.healthFailures[gameID]
}

// clearHealthFailure resets a game's consecutive failure count after either a
// healthy observation or completed reconciliation.
func (control *ControlPlane) clearHealthFailure(gameID string) {
	control.lifecycleMu.Lock()
	defer control.lifecycleMu.Unlock()
	delete(control.healthFailures, gameID)
}
