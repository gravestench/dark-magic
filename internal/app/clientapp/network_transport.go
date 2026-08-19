package clientapp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/clientsession"
	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/logging"
)

const maximumInFlightSubmissions = 8

// send drains staged input with bounded concurrent transport requests.
func (controller *networkController) send(
	ctx context.Context,
	client *clientsession.Session,
) {
	inFlight := make(chan struct{}, maximumInFlightSubmissions)

	var submissions sync.WaitGroup
	defer submissions.Wait()

	for {
		select {
		case intent := <-controller.submissions:
			if !reserveSubmission(ctx, inFlight) {
				return
			}

			submissions.Add(1)
			go func(intent gameserver.CommandIntent) {
				defer submissions.Done()
				defer func() { <-inFlight }()

				controller.sendIntent(ctx, client, intent)
			}(intent)
		case <-ctx.Done():
			return
		}
	}
}

// reserveSubmission waits for bounded capacity or connection cancellation.
func reserveSubmission(ctx context.Context, inFlight chan<- struct{}) bool {
	select {
	case inFlight <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

// sendIntent retries one staged command across recoverable transport failures.
func (controller *networkController) sendIntent(
	ctx context.Context,
	client *clientsession.Session,
	intent gameserver.CommandIntent,
) {
	logging.Trace(
		slog.Default(),
		"sending network command",
		"sequence", intent.Sequence,
		"kind", intent.Kind,
	)

	for {
		epoch := controller.currentConnectionEpoch()
		err := client.Submit(ctx, intent)

		if err == nil || ctx.Err() != nil {
			return
		}

		if isRemoteProtocolError(err) {
			client.DiscardInput(intent.Sequence)
			slog.Debug(
				"network command rejected",
				"sequence", intent.Sequence,
				"kind", intent.Kind,
				"error", err,
			)

			return
		}

		if err := controller.recover(ctx, client, epoch, err); err != nil {
			controller.fail(controller.currentGeneration(), err)

			return
		}
	}
}

// watch consumes correction streams and recovers interrupted transports.
func (controller *networkController) watch(
	ctx context.Context,
	client *clientsession.Session,
) {
	for ctx.Err() == nil {
		epoch := controller.currentConnectionEpoch()
		err := receiveCorrectionStream(ctx, client)

		if ctx.Err() != nil {
			return
		}

		if isRemoteProtocolError(err) {
			controller.fail(controller.currentGeneration(), err)

			return
		}

		if err := controller.recover(ctx, client, epoch, err); err != nil {
			if ctx.Err() == nil {
				controller.fail(controller.currentGeneration(), err)
			}

			return
		}
	}
}

// receiveCorrectionStream drains one watch stream until failure or closure.
func receiveCorrectionStream(
	ctx context.Context,
	client *clientsession.Session,
) error {
	deltas, failures, err := client.Watch(ctx)
	if err != nil {
		return err
	}

	var streamErr error

	for deltas != nil || failures != nil {
		select {
		case delta, open := <-deltas:
			if !open {
				deltas = nil

				continue
			}

			logging.Trace(
				slog.Default(),
				"received network correction",
				"tick", delta.Tick,
				"upserts", len(delta.Upserts),
				"removed", len(delta.Removed),
			)
		case err, open := <-failures:
			if !open {
				failures = nil

				continue
			}

			if err != nil && ctx.Err() == nil {
				streamErr = err
				deltas = nil
				failures = nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if streamErr != nil {
		return streamErr
	}

	return errors.New("network correction stream ended")
}

// isRemoteProtocolError reports a server-declared terminal request failure.
func isRemoteProtocolError(err error) bool {
	var remote *sessionquic.RemoteError

	return errors.As(err, &remote)
}

// recover reconnects the current transport, then asks Realm for replacement.
func (controller *networkController) recover(
	ctx context.Context,
	client *clientsession.Session,
	observedEpoch uint64,
	cause error,
) error {
	controller.reconnectMu.Lock()
	defer controller.reconnectMu.Unlock()

	mode, address, reconnect, err := controller.beginRecovery(client, observedEpoch)
	if err != nil || !reconnect {
		return err
	}

	slog.Debug(
		"network connection interrupted; reconnecting",
		"mode", mode,
		"address", address,
		"error", cause,
	)

	recoveryContext, cancel := context.WithTimeout(ctx, 8*time.Second)
	lastErr := controller.retryTransport(recoveryContext, client, mode, address, cause)

	cancel()

	if lastErr == nil {
		return nil
	}

	if mode == "realm" && controller.app != nil && controller.app.realm != nil && ctx.Err() == nil {
		lastErr = controller.recoverThroughRealm(ctx, client, lastErr)
		if lastErr == nil {
			return nil
		}
	}

	return fmt.Errorf("network reconnect lease expired: %w", lastErr)
}

// beginRecovery rejects stale recovery work and marks the client reconnecting.
func (controller *networkController) beginRecovery(
	client *clientsession.Session,
	observedEpoch uint64,
) (string, string, bool, error) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.phase == "closed" || controller.client != client {
		return "", "", false, context.Canceled
	}

	if controller.connectionEpoch != observedEpoch && controller.phase == "connected" {
		return "", "", false, nil
	}

	controller.phase = "reconnecting"

	return controller.mode, controller.address, true, nil
}

// retryTransport reconnects the existing endpoint with bounded backoff.
func (controller *networkController) retryTransport(
	ctx context.Context,
	client *clientsession.Session,
	mode string,
	address string,
	cause error,
) error {
	delay := time.Duration(0)
	lastErr := cause

	for attempt := 1; ctx.Err() == nil; attempt++ {
		if !waitForRetry(ctx, delay) {
			break
		}

		attemptContext, stop := context.WithTimeout(ctx, 2*time.Second)
		lastErr = client.Reconnect(attemptContext)

		stop()

		if lastErr == nil {
			if err := controller.completeTransportRecovery(client); err != nil {
				return err
			}

			slog.Debug(
				"network session reconnected",
				"attempt", attempt,
				"mode", mode,
				"address", address,
			)

			return nil
		}

		delay = min(1600*time.Millisecond, max(200*time.Millisecond, delay*2))
	}

	return lastErr
}

// waitForRetry waits for a backoff delay without hiding cancellation.
func waitForRetry(ctx context.Context, delay time.Duration) bool {
	if delay <= 0 {
		return ctx.Err() == nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// completeTransportRecovery commits one successful endpoint reconnect.
func (controller *networkController) completeTransportRecovery(
	client *clientsession.Session,
) error {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.client != client || controller.phase == "closed" {
		return context.Canceled
	}

	controller.connectionEpoch++
	controller.phase = "connected"
	controller.failure = ""

	return nil
}

// recoverThroughRealm requests and commits a replacement worker assignment.
func (controller *networkController) recoverThroughRealm(
	ctx context.Context,
	client *clientsession.Session,
	previousErr error,
) error {
	replacementContext, stop := context.WithTimeout(ctx, 50*time.Second)
	assignment, err := controller.reassignThroughRealm(replacementContext, client)

	stop()

	if err != nil {
		return errors.Join(previousErr, err)
	}

	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.client != client || controller.phase == "closed" {
		return context.Canceled
	}

	controller.address = assignment.Endpoint.Address
	controller.phase = "connected"
	controller.failure = ""
	controller.connectionEpoch++

	slog.Debug(
		"Realm session reconnected through replacement assignment",
		"game_id", assignment.GameID,
		"address", assignment.Endpoint.Address,
	)

	return nil
}

// reassignThroughRealm retries replacement allocation and authenticated reassignment.
func (controller *networkController) reassignThroughRealm(
	ctx context.Context,
	client *clientsession.Session,
) (realm.JoinAssignment, error) {
	delay := time.Duration(0)

	var lastErr error

	for attempt := 1; ctx.Err() == nil; attempt++ {
		if !waitForRetry(ctx, delay) {
			break
		}

		assignment, err := controller.tryRealmReassignment(ctx, client)
		if err == nil {
			return assignment, nil
		}

		lastErr = err
		delay = min(4*time.Second, max(500*time.Millisecond, delay*2))

		slog.Debug("Realm replacement assignment not ready", "attempt", attempt, "error", err)
	}

	return realm.JoinAssignment{}, errors.Join(lastErr, ctx.Err())
}

// tryRealmReassignment requests one assignment and applies its pinned transport.
func (controller *networkController) tryRealmReassignment(
	ctx context.Context,
	client *clientsession.Session,
) (realm.JoinAssignment, error) {
	attemptContext, stop := context.WithTimeout(ctx, 4*time.Second)
	defer stop()

	assignment, err := controller.app.realm.reconnectAssignment(attemptContext)
	if err != nil {
		return realm.JoinAssignment{}, err
	}

	var tlsConfig *tls.Config

	tlsConfig, err = networktrust.PinnedTLSFingerprint(assignment.Endpoint.TLSFingerprint)
	if err != nil {
		return realm.JoinAssignment{}, err
	}

	if err := client.Reassign(attemptContext, assignment, tlsConfig); err != nil {
		return realm.JoinAssignment{}, err
	}

	return assignment, nil
}
