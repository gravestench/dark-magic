package clientsession

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

// reconnectAttempt freezes the identity and credential facts that a recovery response must preserve.
// reconnectMu keeps these values stable until commitReconnectAttempt finishes.
type reconnectAttempt struct {
	originalTransport *sessionquic.Client
	credential        gameserver.SessionCredential
	identity          simulation.RuntimeIdentity
	nonce             string
	identityHash      string
	characterID       string
	owner             playeradapter.HUDIdentity
	gameID            string
	endpoint          realm.GameEndpoint
}

// Reconnect resumes the same authority through credential-safe direct recovery or a pinned redial.
func (session *Session) Reconnect(ctx context.Context) error {
	return session.reconnect(ctx)
}

// Reassign consumes a fresh Realm ticket while preserving the durable game, runtime, character, and
// player identities. Unacknowledged exact input is resubmitted after the atomic transport swap.
func (session *Session) Reassign(
	ctx context.Context,
	assignment realm.JoinAssignment,
	tlsConfig *tls.Config,
) error {
	return session.reassign(ctx, assignment, tlsConfig)
}

// reconnect recovers the same authority directly when possible, otherwise dialing the same pinned
// endpoint again. It never changes durable game, runtime, character, or player identity.
func (session *Session) reconnect(ctx context.Context) error {
	session.reconnectMu.Lock()
	defer session.reconnectMu.Unlock()

	attempt, err := session.prepareReconnectAttempt()
	if err != nil {
		return err
	}

	transport, joined, err := session.executeReconnectAttempt(ctx, attempt)
	if err != nil {
		return err
	}

	view, err := decodeView(joined.Snapshot)
	if err != nil {
		closeReplacementTransport(transport, attempt.originalTransport)

		return err
	}

	if err := validateReconnectIdentity(attempt, joined, view); err != nil {
		closeReplacementTransport(transport, attempt.originalTransport)

		return err
	}

	return session.commitReconnectAttempt(attempt, transport, joined, view)
}

// prepareReconnectAttempt creates one reusable nonce and snapshots all protected recovery invariants.
func (session *Session) prepareReconnectAttempt() (reconnectAttempt, error) {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return reconnectAttempt{}, errors.New("client session: closed")
	}

	if session.reconnectNonce == "" {
		nonce, err := reconnectNonce()
		if err != nil {
			return reconnectAttempt{}, err
		}

		session.reconnectNonce = nonce
	}

	return reconnectAttempt{
		originalTransport: session.transport,
		credential:        session.credential,
		identity:          session.identity,
		nonce:             session.reconnectNonce,
		identityHash:      session.Admission.Admission.IdentityHash,
		characterID:       session.Admission.Admission.CharacterID,
		owner:             session.reliableHUD.Player,
		gameID:            session.gameID,
		endpoint:          session.endpoint,
	}, nil
}

// reconnectNonce returns opaque replay protection retained across retries of one interrupted session.
func reconnectNonce() (string, error) {
	var value [16]byte

	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(value[:]), nil
}

// executeReconnectAttempt gives the existing transport a short recovery window before redialing the
// same endpoint with the stored TLS policy.
func (session *Session) executeReconnectAttempt(
	ctx context.Context,
	attempt reconnectAttempt,
) (*sessionquic.Client, gameserver.JoinResponse, error) {
	request := gameserver.ReconnectRequest{
		Credential: attempt.credential,
		Identity:   attempt.identity,
		Nonce:      attempt.nonce,
	}

	directContext, cancelDirect := context.WithTimeout(ctx, 300*time.Millisecond)
	joined, err := attempt.originalTransport.Reconnect(directContext, request)

	cancelDirect()

	if err == nil {
		return attempt.originalTransport, joined, nil
	}

	session.mu.Lock()
	tlsConfig := session.tlsConfig
	session.mu.Unlock()

	if tlsConfig == nil {
		return nil, gameserver.JoinResponse{}, err
	}

	replacement, _, dialErr := dialVerified(
		ctx,
		attempt.gameID,
		attempt.endpoint,
		attempt.identity,
		tlsConfig,
	)
	if dialErr != nil {
		return nil, gameserver.JoinResponse{}, dialErr
	}

	joined, err = replacement.Reconnect(ctx, request)
	if err != nil {
		_ = replacement.Close()

		return nil, gameserver.JoinResponse{}, err
	}

	return replacement, joined, nil
}

// validateReconnectIdentity rejects a response that resumes different authority-owned identity.
func validateReconnectIdentity(
	attempt reconnectAttempt,
	joined gameserver.JoinResponse,
	view playeradapter.ClientView,
) error {
	if joined.Admission.SessionID != attempt.gameID ||
		joined.Admission.IdentityHash != attempt.identityHash ||
		joined.Admission.CharacterID != attempt.characterID {
		return ErrAssignment
	}

	return validateOwnerIdentity(attempt.owner, view.HUD.Player)
}

// commitReconnectAttempt rotates transport and credential only if the interrupted session is still
// current. A competing close or recovery makes the response stale.
func (session *Session) commitReconnectAttempt(
	attempt reconnectAttempt,
	transport *sessionquic.Client,
	joined gameserver.JoinResponse,
	view playeradapter.ClientView,
) error {
	session.mu.Lock()

	if session.closed || session.credential != attempt.credential {
		currentTransport := session.transport
		session.mu.Unlock()

		if transport != currentTransport {
			_ = transport.Close()
		}

		return ErrStaleCorrection
	}

	oldTransport := session.transport
	session.credential = joined.Credential
	session.Admission = joined
	session.HUD = view.HUD
	session.World = view.World
	session.Private = view.Private
	session.Party = view.Party
	session.Events = view.Events
	session.reliableHUD = view.HUD
	session.reliableWorld = view.World
	session.viewRevision++
	session.eventEpoch++
	session.transport = transport
	session.reconnectNonce = ""
	session.observeSnapshotLocked(joined.Snapshot, time.Now())
	session.discardAcknowledgedLocked(joined.Snapshot.AcknowledgedInput)
	session.publishPresentationLocked()
	session.mu.Unlock()

	closeReplacementTransport(oldTransport, transport)

	return nil
}

// closeReplacementTransport closes first only when it is not the retained second transport.
func closeReplacementTransport(first, second *sessionquic.Client) {
	if first != nil && first != second {
		_ = first.Close()
	}
}

// reassign consumes a Realm ticket for a replacement authority while retaining the Session object
// observed by input and presentation loops. Pending exact inputs are replayed after the atomic swap.
func (session *Session) reassign(
	ctx context.Context,
	assignment realm.JoinAssignment,
	tlsConfig *tls.Config,
) error {
	if session == nil || ctx == nil || tlsConfig == nil || strings.TrimSpace(assignment.Ticket) == "" {
		return ErrAssignment
	}

	session.reconnectMu.Lock()

	expected, err := session.reassignmentExpectation(assignment)
	if err != nil {
		session.reconnectMu.Unlock()

		return err
	}

	replacement, err := Connect(ctx, assignment, tlsConfig)
	if err != nil {
		session.reconnectMu.Unlock()

		return err
	}

	newTransport, err := detachValidatedReplacement(replacement, expected)
	if err != nil {
		session.reconnectMu.Unlock()

		_ = replacement.Close(context.Background())

		return err
	}

	oldTransport, pending, err := session.commitReassignment(
		assignment,
		tlsConfig,
		expected.identity,
		replacement,
		newTransport,
	)
	session.reconnectMu.Unlock()

	if err != nil {
		_ = newTransport.Close()

		return err
	}

	_ = oldTransport.Close()

	return session.resubmitPending(ctx, pending)
}

// reassignmentExpectation snapshots identities that a replacement worker is not allowed to change.
func (session *Session) reassignmentExpectation(
	assignment realm.JoinAssignment,
) (reconnectExpectation, error) {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed || assignment.GameID != session.gameID {
		return reconnectExpectation{}, ErrAssignment
	}

	expected := reconnectExpectation{
		identityHash: session.Admission.Admission.IdentityHash,
		characterID:  session.Admission.Admission.CharacterID,
		owner:        session.reliableHUD.Player,
		identity:     session.identity,
	}

	digest, err := assignment.Runtime.Digest()
	if err != nil || digest != expected.identityHash {
		return reconnectExpectation{}, ErrAssignment
	}

	return expected, nil
}

// reconnectExpectation contains durable identity shared by reconnect and replacement authorities.
type reconnectExpectation struct {
	identityHash string
	characterID  string
	owner        playeradapter.HUDIdentity
	identity     simulation.RuntimeIdentity
}

// detachValidatedReplacement verifies the freshly joined worker before transferring its transport.
func detachValidatedReplacement(
	replacement *Session,
	expected reconnectExpectation,
) (*sessionquic.Client, error) {
	replacement.mu.Lock()
	defer replacement.mu.Unlock()

	if replacement.Admission.Admission.IdentityHash != expected.identityHash ||
		replacement.Admission.Admission.CharacterID != expected.characterID ||
		validateOwnerIdentity(expected.owner, replacement.reliableHUD.Player) != nil {
		return nil, ErrAssignment
	}

	transport := replacement.transport
	replacement.transport = nil
	replacement.closed = true

	return transport, nil
}

// commitReassignment installs the replacement projections and returns sequence-ordered input to replay.
func (session *Session) commitReassignment(
	assignment realm.JoinAssignment,
	tlsConfig *tls.Config,
	identity simulation.RuntimeIdentity,
	replacement *Session,
	newTransport *sessionquic.Client,
) (*sessionquic.Client, []gameserver.CommandIntent, error) {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed || session.gameID != assignment.GameID {
		return nil, nil, ErrStaleCorrection
	}

	oldTransport := session.transport
	session.transport = newTransport
	session.credential = replacement.credential
	session.identity = identity
	session.endpoint = assignment.Endpoint
	session.tlsConfig = tlsConfig.Clone()
	session.Admission = replacement.Admission
	session.HUD = replacement.HUD
	session.World = replacement.World
	session.Private = replacement.Private
	session.Party = replacement.Party
	session.Events = replacement.Events
	session.reliableHUD = replacement.reliableHUD
	session.reliableWorld = replacement.reliableWorld
	session.clock = replacement.clock
	session.reconnectNonce = ""
	session.viewRevision++
	session.eventEpoch++
	session.discardAcknowledgedLocked(replacement.Admission.Snapshot.AcknowledgedInput)

	pending := make([]gameserver.CommandIntent, 0, len(session.pending))
	for _, intent := range session.pending {
		pending = append(pending, intent)
	}

	sort.Slice(pending, func(i, j int) bool { return pending[i].Sequence < pending[j].Sequence })
	session.publishPresentationLocked()

	return oldTransport, pending, nil
}

// resubmitPending replays exact unacknowledged sequences; restored authority suppresses duplicates it owns.
func (session *Session) resubmitPending(ctx context.Context, pending []gameserver.CommandIntent) error {
	for _, intent := range pending {
		if err := session.Submit(ctx, intent); err != nil {
			return err
		}
	}

	return nil
}
