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

type Session struct {
	mu         sync.Mutex
	transport  *sessionquic.Client
	credential gameserver.SessionCredential
	identity   simulation.RuntimeIdentity
	closed     bool
	Admission  gameserver.JoinResponse
	HUD        playeradapter.HUD
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
	hud, err := decodeHUD(joined.Snapshot)
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return &Session{transport: transport, credential: joined.Credential, identity: assignment.Runtime, Admission: joined, HUD: hud}, nil
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
	hud, err := decodeHUD(joined.Snapshot)
	if err != nil {
		return err
	}
	session.credential, session.Admission, session.HUD = joined.Credential, joined, hud
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

func decodeHUD(snapshot gameserver.Snapshot) (playeradapter.HUD, error) {
	var hud playeradapter.HUD
	if err := json.Unmarshal(snapshot.Payload, &hud); err != nil || hud.Version != playeradapter.HUDVersion || hud.Tick != snapshot.Tick {
		return playeradapter.HUD{}, fmt.Errorf("%w: invalid PlayerHUD/v1", ErrAssignment)
	}
	return hud, nil
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
