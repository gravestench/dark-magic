package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

const (
	movementAcknowledgementTimeout = 10 * time.Second
	checkpointPollInterval         = 25 * time.Millisecond
)

// connectWorker pins the authenticated handoff fingerprint before opening the QUIC session.
func connectWorker(ctx context.Context, handoff realm.GameHandoff) (*clientsession.Session, error) {
	workerTLS, err := pinnedWorkerTLS(handoff)
	if err != nil {
		return nil, fmt.Errorf("pin worker identity: %w", err)
	}

	connected, err := clientsession.Connect(ctx, handoff.Assignment, workerTLS)
	if err != nil {
		return nil, fmt.Errorf("join worker over QUIC: %w", err)
	}

	return connected, nil
}

// pinnedWorkerTLS converts only the Realm-authenticated fingerprint into transport trust.
func pinnedWorkerTLS(handoff realm.GameHandoff) (*tls.Config, error) {
	return networktrust.PinnedTLSFingerprint(handoff.Assignment.Endpoint.TLSFingerprint)
}

// admittedPlayerID rejects a connected worker unless its admission and first HUD agree with the selected character.
func admittedPlayerID(connected *clientsession.Session, characterID string) (string, error) {
	hud, _ := connected.View()
	if connected.Admission.Admission.CharacterID != characterID || hud.Player.PlayerID == "" {
		return "", errors.New("worker admitted the wrong character identity")
	}

	return hud.Player.PlayerID, nil
}

// exerciseWorkerSession proves movement, live reconnect, and optional post-restart reassignment before Realm commit.
func exerciseWorkerSession(
	ctx context.Context,
	client *realm.RealmClient,
	connected *clientsession.Session,
	gameID string,
) error {
	watchCtx, stopWatch := context.WithCancel(ctx)
	// Watch may fail before the session starts, so register cancellation before taking that early-return path.
	defer stopWatch()

	deltas, watchErrors, err := connected.Watch(watchCtx)
	if err != nil {
		return fmt.Errorf("watch worker: %w", err)
	}

	payload, err := json.Marshal(movement.MovePayload{X: 1})
	if err != nil {
		return err
	}

	initialTick, err := submitMovement(ctx, connected, payload)
	if err != nil {
		return err
	}

	if err := awaitPlayed(ctx, connected, deltas, watchErrors, initialTick); err != nil {
		return err
	}

	if err := reconnectLiveWorker(ctx, connected); err != nil {
		return err
	}

	if err := checkpointBarrier(ctx); err != nil {
		return err
	}

	if os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_REALM_RESTART") == "1" {
		return reassignAfterRealmRestart(ctx, client, connected, gameID)
	}

	return nil
}

// submitMovement captures the pre-command tick and submits one sequenced input for acknowledgement testing.
func submitMovement(ctx context.Context, connected *clientsession.Session, payload []byte) (uint64, error) {
	_, initialWorld := connected.View()

	intent := gameserver.CommandIntent{
		Sequence:   1,
		TargetTick: connected.NextInputTick(time.Now()),
		Kind:       movement.MoveCommand,
		Payload:    payload,
	}
	if err := connected.Submit(ctx, intent); err != nil {
		return 0, fmt.Errorf("submit movement: %w", err)
	}

	return initialWorld.Tick, nil
}

// awaitPlayed waits for both world advancement and an empty pending queue so a correction alone cannot pass the test.
func awaitPlayed(
	ctx context.Context,
	session *clientsession.Session,
	deltas <-chan playeradapter.WorldDelta,
	watchErrors <-chan error,
	initialTick uint64,
) error {
	timer := time.NewTimer(movementAcknowledgementTimeout)
	defer timer.Stop()

	for {
		_, world := session.View()
		if world.Tick > initialTick && len(session.PendingInputs()) == 0 {
			return nil
		}

		select {
		case _, open := <-deltas:
			if !open {
				return errors.New("worker correction stream closed before movement acknowledgement")
			}
		case err := <-watchErrors:
			return fmt.Errorf("worker correction stream: %w", err)
		case <-timer.C:
			return errors.New("worker did not acknowledge movement before timeout")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// reconnectLiveWorker proves that credential rotation can refresh the already-admitted session in place.
func reconnectLiveWorker(ctx context.Context, connected *clientsession.Session) error {
	if err := connected.Reconnect(ctx); err != nil {
		return fmt.Errorf("reconnect live QUIC session: %w", err)
	}

	if _, err := connected.Refresh(ctx); err != nil {
		return fmt.Errorf("refresh reconnected session: %w", err)
	}

	return nil
}

// reassignAfterRealmRestart obtains a fresh authenticated handoff and moves the existing session to its replacement.
func reassignAfterRealmRestart(
	ctx context.Context,
	client *realm.RealmClient,
	connected *clientsession.Session,
	gameID string,
) error {
	reassignment, err := client.ReconnectGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("obtain post-restart assignment: %w", err)
	}

	replacementTLS, err := pinnedWorkerTLS(reassignment)
	if err != nil {
		return fmt.Errorf("pin replacement worker identity: %w", err)
	}

	if err := connected.Reassign(ctx, reassignment.Assignment, replacementTLS); err != nil {
		return fmt.Errorf("reassign post-restart QUIC session: %w", err)
	}

	if _, err := connected.Refresh(ctx); err != nil {
		return fmt.Errorf("refresh post-restart session: %w", err)
	}

	return nil
}

// checkpointBarrier coordinates an external Realm restart without imposing filesystem polling on normal runs.
func checkpointBarrier(ctx context.Context) error {
	ready := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_READY_FILE"))

	proceed := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_CONTINUE_FILE"))
	if ready == "" && proceed == "" {
		return nil
	}

	if ready == "" || proceed == "" {
		return errors.New("ready and continue files must be configured together")
	}

	// Restrictive permissions keep the synchronization files from becoming a shared-user signaling surface.
	if err := os.WriteFile(ready, []byte("ready\n"), 0o600); err != nil {
		return fmt.Errorf("publish checkpoint barrier: %w", err)
	}

	for {
		if _, err := os.Stat(proceed); err == nil {
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(checkpointPollInterval):
		}
	}
}
