package realm

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GameHandoff combines the public directory view with the private admission
// assignment a client needs to connect to its allocated worker.
type GameHandoff struct {
	Game       GameDetail     `json:"game"`
	Assignment JoinAssignment `json:"assignment"`
}

// CreateGame reserves the directory entry, allocates and registers its worker,
// then admits the creator. Each failure unwinds every earlier durable phase.
func (control *ControlPlane) CreateGame(
	ctx context.Context,
	token string,
	request CreateGameRequest,
) (handoff GameHandoff, err error) {
	event := AuditEvent{
		Operation: AuditGameCreate,
		GameName:  strings.TrimSpace(request.Name),
	}
	defer func() {
		event.GameID = handoff.Game.Entry.GameID
		event.GameName = firstNonEmpty(handoff.Game.Entry.Name, event.GameName)
		control.recordAudit(ctx, event, err)
	}()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return GameHandoff{}, err
	}

	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID
	control.addSelectedCharacterToAudit(ctx, token, principal, &event)

	if control.allocator == nil || control.admissions == nil {
		return GameHandoff{}, ErrGameUnavailable
	}

	detail, err := control.games.Create(ctx, principal, request)
	if err != nil {
		return GameHandoff{}, err
	}

	allocationID := uuid.New().String()
	if _, err := control.allocations.Request(ctx, detail.Entry.GameID, allocationID); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		return GameHandoff{}, errors.Join(err, control.games.Remove(cleanupCtx, detail.Entry.GameID))
	}

	allocation, err := control.allocator.Allocate(ctx, GameSpec{
		GameID:         detail.Entry.GameID,
		AllocationID:   allocationID,
		Difficulty:     detail.Entry.Difficulty,
		Hardcore:       detail.Entry.Hardcore,
		MaximumPlayers: detail.Entry.MaximumPlayers,
	})
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		return GameHandoff{}, errors.Join(
			err,
			control.allocations.Fail(cleanupCtx, detail.Entry.GameID, err),
			control.games.Remove(cleanupCtx, detail.Entry.GameID),
		)
	}

	if allocation.AllocationID != allocationID || allocation.Worker == nil {
		err = ErrWorker

		return GameHandoff{}, errors.Join(
			err,
			control.removeAllocatedGame(ctx, detail.Entry.GameID, false, err),
		)
	}

	description, err := allocation.Worker.Describe(ctx)
	if err != nil {
		return GameHandoff{}, errors.Join(
			err,
			control.removeAllocatedGame(ctx, detail.Entry.GameID, false, err),
		)
	}

	if _, err := control.allocations.Ready(
		ctx,
		detail.Entry.GameID,
		allocation.Endpoint,
		description.Runtime,
	); err != nil {
		return GameHandoff{}, errors.Join(
			err,
			control.removeAllocatedGame(ctx, detail.Entry.GameID, false, err),
		)
	}

	if err := control.admissions.RegisterGame(
		detail.Entry.GameID,
		allocation.Tickets,
		allocation.Endpoint,
	); err != nil {
		return GameHandoff{}, errors.Join(
			err,
			control.removeAllocatedGame(ctx, detail.Entry.GameID, false, err),
		)
	}

	handoff, err = control.joinResolvedGame(ctx, token, principal, detail)
	if err != nil {
		return GameHandoff{}, errors.Join(
			err,
			control.removeAllocatedGame(ctx, detail.Entry.GameID, true, err),
		)
	}

	return handoff, nil
}

// ListGames returns the public directory only after reauthorizing the caller.
func (control *ControlPlane) ListGames(
	ctx context.Context,
	token string,
	filter GameFilter,
) ([]GameDirectoryEntry, error) {
	if _, err := control.authorize(ctx, token); err != nil {
		return nil, err
	}

	return control.games.List(ctx, filter)
}

// GameDetail resolves a public game reference without exposing private
// admission state to unauthenticated callers.
func (control *ControlPlane) GameDetail(
	ctx context.Context,
	token string,
	reference string,
) (GameDetail, error) {
	if _, err := control.authorize(ctx, token); err != nil {
		return GameDetail{}, err
	}

	return control.games.Detail(ctx, reference)
}

// ResolveGameJoin validates a public reference and password without reserving
// a roster slot, allowing clients to distinguish discovery from admission.
func (control *ControlPlane) ResolveGameJoin(
	ctx context.Context,
	token string,
	reference string,
	password string,
) (gameID string, err error) {
	event := AuditEvent{
		Operation:     AuditGameResolve,
		GameReference: strings.TrimSpace(reference),
	}
	defer func() {
		event.GameID = gameID
		control.recordAudit(ctx, event, err)
	}()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return "", err
	}

	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID

	return control.games.ResolveJoin(ctx, reference, password)
}

// JoinGame resolves the directory reference and admits the selected character
// through the same reservation path used by game creation.
func (control *ControlPlane) JoinGame(
	ctx context.Context,
	token string,
	reference string,
	password string,
) (handoff GameHandoff, err error) {
	event := AuditEvent{
		Operation:     AuditGameJoin,
		GameReference: strings.TrimSpace(reference),
	}
	defer func() {
		event.GameID = handoff.Game.Entry.GameID
		control.recordAudit(ctx, event, err)
	}()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return GameHandoff{}, err
	}

	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID
	control.addSelectedCharacterToAudit(ctx, token, principal, &event)

	if control.allocator == nil || control.admissions == nil {
		return GameHandoff{}, ErrGameUnavailable
	}

	gameID, err := control.games.ResolveJoin(ctx, reference, password)
	if err != nil {
		return GameHandoff{}, err
	}

	detail, err := control.games.admissionDetail(ctx, gameID)
	if err != nil {
		return GameHandoff{}, err
	}

	handoff, err = control.joinResolvedGame(ctx, token, principal, detail)

	return handoff, err
}

// ReconnectGame reissues an assignment only for an existing durable membership
// owned by the authenticated account and currently selected character.
func (control *ControlPlane) ReconnectGame(
	ctx context.Context,
	token string,
	gameID string,
) (handoff GameHandoff, err error) {
	event := AuditEvent{
		Operation: AuditGameReconnect,
		GameID:    strings.TrimSpace(gameID),
	}
	defer func() { control.recordAudit(ctx, event, err) }()

	if control == nil || control.admissions == nil {
		return GameHandoff{}, ErrGameUnavailable
	}

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return GameHandoff{}, err
	}

	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID

	detail, err := control.games.admissionDetail(ctx, strings.TrimSpace(gameID))
	if err != nil {
		return GameHandoff{}, err
	}

	characterID, err := control.accounts.SelectedCharacter(ctx, token)
	if err != nil {
		return GameHandoff{}, err
	}

	assignment, err := control.admissions.ReconnectAssignment(
		ctx,
		detail.Entry.GameID,
		principal.accountID,
		characterID,
	)
	if err != nil {
		return GameHandoff{}, err
	}

	event.CharacterID = characterID

	return GameHandoff{Game: detail, Assignment: assignment}, nil
}

// addSelectedCharacterToAudit enriches lifecycle events when a selection is
// available. Missing selection is intentionally non-fatal to the outer request.
func (control *ControlPlane) addSelectedCharacterToAudit(
	ctx context.Context,
	token string,
	principal AuthenticatedPrincipal,
	event *AuditEvent,
) {
	characterID, selectedErr := control.accounts.SelectedCharacter(ctx, token)
	if selectedErr != nil {
		return
	}

	record, recordErr := control.characters.Get(ctx, principal.accountID, characterID)
	if recordErr != nil {
		return
	}

	event.CharacterID = record.Character.ID
	event.CharacterName = record.Character.Name
}

// joinResolvedGame validates character compatibility before reserving a roster
// slot. Worker admission commits before lobby presence is removed.
func (control *ControlPlane) joinResolvedGame(
	ctx context.Context,
	token string,
	principal AuthenticatedPrincipal,
	detail GameDetail,
) (GameHandoff, error) {
	characterID, err := control.accounts.SelectedCharacter(ctx, token)
	if err != nil {
		return GameHandoff{}, err
	}

	record, err := control.characters.Get(ctx, principal.accountID, characterID)
	if err != nil {
		return GameHandoff{}, err
	}

	if record.Character.Expansion != detail.Entry.Expansion ||
		record.Character.Hardcore != detail.Entry.Hardcore {
		return GameHandoff{}, ErrGameDirectoryInput
	}

	if detail.Entry.CharacterDifference > 0 && len(detail.Players) > 0 {
		difference := record.Character.Level - detail.Players[0].Level
		if difference < 0 {
			difference = -difference
		}

		if difference > detail.Entry.CharacterDifference {
			return GameHandoff{}, ErrGameLevelRange
		}
	}

	playerID := uuid.New().String()

	reservation, err := control.games.ReservePlayer(ctx, detail.Entry.GameID, GamePlayer{
		CharacterID: record.Character.ID,
		Name:        record.Character.Name,
		Class:       record.Character.Class,
		Level:       record.Character.Level,
	})
	if err != nil {
		return GameHandoff{}, err
	}

	assignment, err := control.admissions.Join(ctx, JoinRequest{
		AccountID:   principal.accountID,
		CharacterID: record.Character.ID,
		PlayerID:    playerID,
		GameID:      detail.Entry.GameID,
		Destination: control.entryDestination,
	})
	if err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		return GameHandoff{}, errors.Join(err, control.games.CancelPlayer(cleanupCtx, reservation))
	}

	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	updated, err := control.games.CommitPlayer(cleanupCtx, reservation)
	if err != nil {
		return GameHandoff{}, errors.Join(
			err,
			control.admissions.CancelMembership(cleanupCtx, detail.Entry.GameID, playerID),
			control.games.CancelPlayer(cleanupCtx, reservation),
		)
	}

	// Successful admission moves the session out of the public channel at once.
	// Presence cleanup is subordinate to the committed admission; maintenance
	// will remove it later if this best-effort leave fails.
	_ = control.channels.Leave(cleanupCtx, principal)

	return GameHandoff{Game: updated, Assignment: assignment}, nil
}

// removeAllocatedGame unwinds allocation, directory, checkpoint, and admission
// state in their established order. A clean retirement completes the durable
// allocation; a failed retirement records the joined cause instead.
func (control *ControlPlane) removeAllocatedGame(
	ctx context.Context,
	gameID string,
	registered bool,
	cause error,
) error {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	var result error
	if registered {
		result = errors.Join(result, control.admissions.UnregisterGame(gameID))
	}

	result = errors.Join(result, control.allocator.Release(cleanupCtx, gameID))
	result = errors.Join(result, control.games.Remove(cleanupCtx, gameID))

	if cause == nil && result == nil {
		result = errors.Join(result, control.allocations.Complete(cleanupCtx, gameID))
		if result == nil {
			result = errors.Join(result, control.checkpoints.Remove(cleanupCtx, gameID))
			control.checkpointMu.Lock()
			delete(control.lastCheckpoint, gameID)
			control.checkpointMu.Unlock()
		}
	} else {
		result = errors.Join(result, control.allocations.Fail(cleanupCtx, gameID, errors.Join(cause, result)))
	}

	return result
}
