// Package clientsession owns the transport-facing client side of one remote
// authoritative game session without depending on presentation or input.
package clientsession

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

var ErrAssignment = errors.New("client session: invalid realm assignment")
var ErrStaleCorrection = errors.New("client session: stale or conflicting correction")

type Session struct {
	mu         sync.Mutex
	transport  *sessionquic.Client
	credential gameserver.SessionCredential
	identity   simulation.RuntimeIdentity
	closed     bool
	Admission  gameserver.JoinResponse
	HUD        playeradapter.HUD
	World      playeradapter.WorldView
}

func Connect(ctx context.Context, assignment realm.JoinAssignment, tlsConfig *tls.Config) (*Session, error) {
	if tlsConfig == nil || strings.TrimSpace(assignment.Ticket) == "" || strings.TrimSpace(assignment.GameID) == "" {
		return nil, ErrAssignment
	}
	if _, _, err := net.SplitHostPort(assignment.Endpoint.Address); err != nil {
		return nil, fmt.Errorf("%w: endpoint address: %v", ErrAssignment, err)
	}
	digest, err := assignment.Runtime.Digest()
	if err != nil {
		return nil, fmt.Errorf("%w: runtime: %v", ErrAssignment, err)
	}
	expected, err := parseFingerprint(assignment.Endpoint.TLSFingerprint)
	if err != nil {
		return nil, err
	}
	verifiedTLS := tlsConfig.Clone()
	previousVerify := verifiedTLS.VerifyPeerCertificate
	verifiedTLS.VerifyPeerCertificate = func(rawCerts [][]byte, chains [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return ErrAssignment
		}
		actual := sha256.Sum256(rawCerts[0])
		if subtle.ConstantTimeCompare(actual[:], expected) != 1 {
			return fmt.Errorf("%w: TLS fingerprint differs", ErrAssignment)
		}
		if previousVerify != nil {
			return previousVerify(rawCerts, chains)
		}
		return nil
	}
	transport, err := sessionquic.Dial(ctx, assignment.Endpoint.Address, verifiedTLS)
	if err != nil {
		return nil, err
	}
	joined, err := transport.Join(ctx, gameserver.JoinRequest{Version: gameserver.SessionProtocolVersion, Credential: assignment.Ticket, Identity: assignment.Runtime})
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	if joined.Admission.SessionID != assignment.GameID || joined.Admission.IdentityHash != digest || joined.Snapshot.Version != gameserver.SessionProtocolVersion {
		_ = transport.Close()
		return nil, ErrAssignment
	}
	view, err := decodeView(joined.Snapshot)
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return &Session{transport: transport, credential: joined.Credential, identity: assignment.Runtime, Admission: joined, HUD: view.HUD, World: view.World}, nil
}

func (session *Session) Submit(ctx context.Context, intent gameserver.CommandIntent) error {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return errors.New("client session: closed")
	}
	return session.transport.Submit(ctx, session.credential, intent)
}

func (session *Session) Reconnect(ctx context.Context) error {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return errors.New("client session: closed")
	}
	joined, err := session.transport.Reconnect(ctx, gameserver.ReconnectRequest{Credential: session.credential, Identity: session.identity})
	if err != nil {
		return err
	}
	view, err := decodeView(joined.Snapshot)
	if err != nil {
		return err
	}
	session.credential, session.Admission, session.HUD, session.World = joined.Credential, joined, view.HUD, view.World
	return nil
}

// Refresh fetches one reliable canonical correction and returns the public
// world delta from the previously installed view.
func (session *Session) Refresh(ctx context.Context) (playeradapter.WorldDelta, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return playeradapter.WorldDelta{}, errors.New("client session: closed")
	}
	snapshot, err := session.transport.Refresh(ctx, session.credential)
	if err != nil {
		return playeradapter.WorldDelta{}, err
	}
	return session.applyCorrection(snapshot)
}

// Watch streams reliable canonical corrections until cancellation. Only one
// correction is buffered, so a slow consumer propagates backpressure.
func (session *Session) Watch(ctx context.Context) (<-chan playeradapter.WorldDelta, <-chan error, error) {
	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()
		return nil, nil, errors.New("client session: closed")
	}
	snapshots, transportErrors, err := session.transport.Watch(ctx, session.credential)
	session.mu.Unlock()
	if err != nil {
		return nil, nil, err
	}
	deltas := make(chan playeradapter.WorldDelta, 1)
	errorsOut := make(chan error, 1)
	go func() {
		defer close(deltas)
		defer close(errorsOut)
		for snapshots != nil || transportErrors != nil {
			select {
			case snapshot, open := <-snapshots:
				if !open {
					snapshots = nil
					continue
				}
				session.mu.Lock()
				delta, applyErr := session.applyCorrection(snapshot)
				session.mu.Unlock()
				if applyErr != nil {
					errorsOut <- applyErr
					return
				}
				select {
				case deltas <- delta:
				case <-ctx.Done():
					return
				}
			case streamErr, open := <-transportErrors:
				if !open {
					transportErrors = nil
					continue
				}
				if streamErr != nil {
					errorsOut <- streamErr
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return deltas, errorsOut, nil
}

// applyCorrection requires session.mu.
func (session *Session) applyCorrection(snapshot gameserver.Snapshot) (playeradapter.WorldDelta, error) {
	view, err := decodeView(snapshot)
	if err != nil {
		return playeradapter.WorldDelta{}, err
	}
	if err := validateCorrection(session.Admission.Snapshot, snapshot); err != nil {
		return playeradapter.WorldDelta{}, err
	}
	delta := playeradapter.DiffWorldView(session.World, view.World)
	session.Admission.Snapshot, session.HUD, session.World = snapshot, view.HUD, view.World
	return delta, nil
}

func validateCorrection(previous, next gameserver.Snapshot) error {
	if next.Tick < previous.Tick || (next.Tick == previous.Tick && next.Checksum != previous.Checksum) {
		return ErrStaleCorrection
	}
	return nil
}

func (session *Session) Close(ctx context.Context) error {
	if session == nil || session.transport == nil {
		return nil
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return nil
	}
	session.closed = true
	leaveErr := session.transport.Leave(ctx, session.credential)
	closeErr := session.transport.Close()
	return errors.Join(leaveErr, closeErr)
}

func decodeView(snapshot gameserver.Snapshot) (playeradapter.ClientView, error) {
	var view playeradapter.ClientView
	if err := json.Unmarshal(snapshot.Payload, &view); err != nil || view.Version != playeradapter.ClientViewVersion || view.Tick != snapshot.Tick || view.HUD.Version != playeradapter.HUDVersion || view.HUD.Tick != snapshot.Tick || view.World.Version != playeradapter.WorldViewVersion || view.World.Tick != snapshot.Tick {
		return playeradapter.ClientView{}, fmt.Errorf("%w: invalid ClientView/v1", ErrAssignment)
	}
	return view, nil
}

func parseFingerprint(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "sha256:") {
		return nil, ErrAssignment
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	if err != nil || len(decoded) != sha256.Size {
		return nil, ErrAssignment
	}
	return decoded, nil
}
