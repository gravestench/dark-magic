package realm

import (
	"context"
	"errors"
	"strings"
	"time"
)

// departureReceipt makes character commit and worker removal independently
// retryable so interrupted departures cannot commit a projection twice.
type departureReceipt struct {
	Record        CharacterRecord `json:"record"`
	PlayerID      string          `json:"player_id"`
	WorkerRemoved bool            `json:"worker_removed"`
}

// GameDrainResult reports how many active characters reached durable departure
// while draining a game. Retries may therefore report only newly committed work.
type GameDrainResult struct {
	GameID              string `json:"game_id"`
	CommittedCharacters int    `json:"committed_characters"`
}

// LeaveGame commits only the worker's canonical projection. The authenticated
// account selects the membership; clients cannot provide player IDs, revisions,
// lease tokens, or replacement character state.
func (control *ControlPlane) LeaveGame(
	ctx context.Context,
	token string,
	gameID string,
) (record CharacterRecord, err error) {
	event := AuditEvent{
		Operation: AuditGameLeave,
		GameID:    strings.TrimSpace(gameID),
	}
	defer func() {
		event.CharacterID = record.Character.ID
		event.CharacterName = record.Character.Name
		control.recordAudit(ctx, event, err)
	}()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}

	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID

	// Serialize receipt creation with membership consumption so concurrent HTTP
	// retries cannot both pass the pre-commit lookup.
	control.departureFlowMu.Lock()
	defer control.departureFlowMu.Unlock()

	characterID, err := control.accounts.SelectedCharacter(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}

	if completed, found, lookupErr := control.departure(ctx, gameID, characterID); lookupErr != nil {
		return CharacterRecord{}, lookupErr
	} else if found {
		return completed.Record, control.completeDeparture(ctx, gameID, principal.accountID, completed)
	}

	if control.allocator == nil || control.admissions == nil {
		return CharacterRecord{}, ErrGameUnavailable
	}

	playerID, baseline, err := control.admissions.CharacterMembership(
		gameID,
		principal.accountID,
		characterID,
	)
	if err != nil {
		return CharacterRecord{}, err
	}

	event.CharacterID = baseline.Character.ID
	event.CharacterName = baseline.Character.Name

	record, commitErr := control.admissions.LeaveCanonicalMembership(ctx, gameID, playerID)
	if record.Character.ID == "" {
		return CharacterRecord{}, commitErr
	}

	membership, receiptErr := control.membershipStore.ByCharacter(ctx, gameID, characterID)
	if receiptErr != nil || membership.Departure == nil {
		return record, errors.Join(commitErr, receiptErr, ErrMembership)
	}

	receipt := cloneDepartureReceipt(*membership.Departure)

	return record, control.completeDeparture(ctx, gameID, principal.accountID, receipt)
}

// completeDeparture resumes cleanup from a durable receipt. Worker removal is
// persisted before roster mutation so retries never target a replacement player.
func (control *ControlPlane) completeDeparture(
	ctx context.Context,
	gameID string,
	accountID string,
	receipt departureReceipt,
) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if !receipt.WorkerRemoved {
		worker, found := control.allocator.Game(gameID)
		if !found {
			return ErrGameNotFound
		}

		if err := worker.RemoveCharacter(cleanupCtx, receipt.PlayerID); err != nil {
			return err
		}

		updated, err := control.membershipStore.MarkWorkerRemoved(
			cleanupCtx,
			gameID,
			receipt.PlayerID,
		)
		if err != nil {
			return err
		}

		receipt = updated
	}

	detail, rosterErr := control.games.RemovePlayer(cleanupCtx, gameID, receipt.Record.Character.ID)
	if errors.Is(rosterErr, ErrGameNotFound) || errors.Is(rosterErr, ErrCharacterNotFound) {
		return nil
	}

	if rosterErr != nil {
		return rosterErr
	}

	if detail.Entry.Players == 0 {
		return control.removeAllocatedGame(cleanupCtx, gameID, true, nil)
	}

	return nil
}

// DrainGame is the trusted operator lifecycle path. BeginDrain first closes
// discovery and all new admission atomically. Every active membership then
// follows the ordinary canonical projection, durable departure receipt, worker
// removal, and roster cleanup sequence. A partial failure leaves the durable
// game draining and is safe to retry.
func (control *ControlPlane) DrainGame(
	ctx context.Context,
	gameID string,
) (result GameDrainResult, err error) {
	gameID = strings.TrimSpace(gameID)
	result.GameID = gameID

	event := AuditEvent{Operation: AuditGameDrain, GameID: gameID}
	defer func() { control.recordAudit(ctx, event, err) }()

	if control == nil ||
		ctx == nil ||
		control.allocator == nil ||
		control.admissions == nil ||
		gameID == "" {
		return result, ErrGameUnavailable
	}

	if err := control.games.BeginDrain(ctx, gameID); err != nil {
		return result, err
	}

	players, err := control.membershipStore.DrainPlayerIDs(ctx, gameID)
	if err != nil {
		return result, err
	}

	if len(players) == 0 {
		err = control.removeAllocatedGame(ctx, gameID, true, nil)
		return result, err
	}

	for _, playerID := range players {
		membership, lookupErr := control.membershipStore.ByPlayer(ctx, gameID, playerID)
		if lookupErr != nil {
			return result, lookupErr
		}

		if err := control.reconcileExpiredPlayer(ctx, gameID, playerID); err != nil {
			return result, err
		}

		if membership.State == MembershipActive {
			result.CommittedCharacters++
		}
	}

	return result, nil
}

// reconcileExpiredPlayer handles a trusted worker notification emitted only
// after transport reconnect grace has elapsed. Canonical projection and the
// lease commit still happen before entity, roster, or allocation removal.
func (control *ControlPlane) reconcileExpiredPlayer(
	ctx context.Context,
	gameID string,
	playerID string,
) (err error) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return ErrWorker
	}

	control.departureFlowMu.Lock()
	defer control.departureFlowMu.Unlock()

	accountID, baseline, membershipErr := control.admissions.PlayerMembership(gameID, playerID)

	receiptAccountID, receipt, completed, receiptErr := control.departureByPlayer(ctx, gameID, playerID)
	if receiptErr != nil {
		return receiptErr
	}

	if membershipErr != nil && !completed {
		return membershipErr
	}

	if completed {
		accountID = receiptAccountID
	}

	if !completed {
		record, commitErr := control.admissions.LeaveCanonicalMembership(ctx, gameID, playerID)
		if record.Character.ID == "" {
			return commitErr
		}

		membership, durableErr := control.membershipStore.ByPlayer(ctx, gameID, playerID)
		if durableErr != nil || membership.Departure == nil {
			return errors.Join(commitErr, durableErr, ErrMembership)
		}

		receipt = cloneDepartureReceipt(*membership.Departure)
	}

	event := AuditEvent{
		Operation: AuditGameLeave,
		GameID:    gameID,
		AccountID: accountID,
		CharacterID: firstNonEmpty(
			receipt.Record.Character.ID,
			baseline.Character.ID,
		),
		CharacterName: firstNonEmpty(
			receipt.Record.Character.Name,
			baseline.Character.Name,
		),
	}
	defer func() { control.recordAudit(ctx, event, err) }()

	return control.completeDeparture(ctx, gameID, accountID, receipt)
}

// departure finds a completed receipt by character so an authenticated retry
// can resume cleanup without recomputing the canonical projection.
func (control *ControlPlane) departure(
	ctx context.Context,
	gameID string,
	characterID string,
) (departureReceipt, bool, error) {
	record, err := control.membershipStore.ByCharacter(ctx, gameID, characterID)
	if errors.Is(err, ErrMembership) {
		return departureReceipt{}, false, nil
	}

	if err != nil {
		return departureReceipt{}, false, err
	}

	if record.State != MembershipDeparted || record.Departure == nil {
		return departureReceipt{}, false, nil
	}

	return cloneDepartureReceipt(*record.Departure), true, nil
}

// departureByPlayer finds a completed receipt by worker player identity so
// reconnect-expiry cleanup remains retryable after in-memory admission is gone.
func (control *ControlPlane) departureByPlayer(
	ctx context.Context,
	gameID string,
	playerID string,
) (string, departureReceipt, bool, error) {
	record, err := control.membershipStore.ByPlayer(ctx, gameID, playerID)
	if errors.Is(err, ErrMembership) {
		return "", departureReceipt{}, false, nil
	}

	if err != nil {
		return "", departureReceipt{}, false, err
	}

	if record.State != MembershipDeparted || record.Departure == nil {
		return "", departureReceipt{}, false, nil
	}

	return record.AccountID, cloneDepartureReceipt(*record.Departure), true, nil
}
