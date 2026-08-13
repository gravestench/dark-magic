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
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

var ErrAssignment = errors.New("client session: invalid server assignment")
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

type SelfHostedAssignment struct {
	GameID            string
	Endpoint          realm.GameEndpoint
	Runtime           simulation.RuntimeIdentity
	ProfileCredential string
}

// View returns one coherent defensive copy while a correction stream may be
// installing newer state. Presentation code should use this accessor instead
// of reading the exported initial-view fields concurrently with Watch.
func (session *Session) View() (playeradapter.HUD, playeradapter.WorldView) {
	session.mu.Lock()
	defer session.mu.Unlock()
	hud, world := session.HUD, session.World
	world.Entities = append([]playeradapter.WorldEntity(nil), session.World.Entities...)
	for index := range world.Entities {
		if health := world.Entities[index].Health; health != nil {
			value := *health
			world.Entities[index].Health = &value
		}
		if maximum := world.Entities[index].MaxHealth; maximum != nil {
			value := *maximum
			world.Entities[index].MaxHealth = &value
		}
	}
	return hud, world
}

func Connect(ctx context.Context, assignment realm.JoinAssignment, tlsConfig *tls.Config) (*Session, error) {
	if tlsConfig == nil || strings.TrimSpace(assignment.Ticket) == "" || strings.TrimSpace(assignment.GameID) == "" {
		return nil, ErrAssignment
	}
	transport, digest, err := dialVerified(ctx, assignment.GameID, assignment.Endpoint, assignment.Runtime, tlsConfig)
	if err != nil {
		return nil, err
	}
	return joinVerified(ctx, transport, assignment.GameID, assignment.Runtime, digest, assignment.Ticket, "")
}

// ConnectSelfHosted submits only the selected player-profile character to an
// explicitly configured self-host, receives a one-use ticket, then enters the
// ordinary authenticated session path.
func ConnectSelfHosted(ctx context.Context, assignment SelfHostedAssignment, tlsConfig *tls.Config, profile *d2save.Store) (*Session, error) {
	if profile == nil {
		return nil, ErrAssignment
	}
	character, selected := profile.Selected()
	if !selected {
		return nil, ErrAssignment
	}
	transport, digest, err := dialVerified(ctx, assignment.GameID, assignment.Endpoint, assignment.Runtime, tlsConfig)
	if err != nil {
		return nil, err
	}
	offer, err := d2save.EncodeCharacterOffer(character)
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	ticket, err := transport.AdmitProfile(ctx, assignment.ProfileCredential, offer)
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	return joinVerified(ctx, transport, assignment.GameID, assignment.Runtime, digest, ticket, character.ID)
}

func dialVerified(ctx context.Context, gameID string, endpoint realm.GameEndpoint, identity simulation.RuntimeIdentity, tlsConfig *tls.Config) (*sessionquic.Client, string, error) {
	if tlsConfig == nil || strings.TrimSpace(gameID) == "" {
		return nil, "", ErrAssignment
	}
	if _, _, err := net.SplitHostPort(endpoint.Address); err != nil {
		return nil, "", fmt.Errorf("%w: endpoint address: %v", ErrAssignment, err)
	}
	digest, err := identity.Digest()
	if err != nil {
		return nil, "", fmt.Errorf("%w: runtime: %v", ErrAssignment, err)
	}
	var expected []byte
	if strings.TrimSpace(endpoint.TLSFingerprint) != "" {
		expected, err = parseFingerprint(endpoint.TLSFingerprint)
		if err != nil {
			return nil, "", err
		}
	}
	verifiedTLS := tlsConfig.Clone()
	previousVerify := verifiedTLS.VerifyPeerCertificate
	verifiedTLS.VerifyPeerCertificate = func(rawCerts [][]byte, chains [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return ErrAssignment
		}
		actual := sha256.Sum256(rawCerts[0])
		if len(expected) > 0 && subtle.ConstantTimeCompare(actual[:], expected) != 1 {
			return fmt.Errorf("%w: TLS fingerprint differs", ErrAssignment)
		}
		if previousVerify != nil {
			return previousVerify(rawCerts, chains)
		}
		return nil
	}
	transport, err := sessionquic.Dial(ctx, endpoint.Address, verifiedTLS)
	return transport, digest, err
}

func joinVerified(ctx context.Context, transport *sessionquic.Client, gameID string, identity simulation.RuntimeIdentity, digest, ticket, characterID string) (*Session, error) {
	joined, err := transport.Join(ctx, gameserver.JoinRequest{Version: gameserver.SessionProtocolVersion, Credential: ticket, Identity: identity})
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	if joined.Admission.SessionID != gameID || joined.Admission.IdentityHash != digest || joined.Snapshot.Version != gameserver.SessionProtocolVersion {
		_ = transport.Close()
		return nil, ErrAssignment
	}
	view, err := decodeView(joined.Snapshot)
	if err != nil {
		_ = transport.Close()
		return nil, err
	}
	if characterID != "" && (joined.Admission.CharacterID != characterID || view.HUD.Player.CharacterID != characterID) {
		_ = transport.Close()
		return nil, ErrAssignment
	}
	return &Session{transport: transport, credential: joined.Credential, identity: identity, Admission: joined, HUD: view.HUD, World: view.World}, nil
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
	if session.closed {
		session.mu.Unlock()
		return nil
	}
	session.closed = true
	transport, credential := session.transport, session.credential
	session.mu.Unlock()
	var leaveErr error
	if ctx.Err() == nil {
		leaveContext, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
		leaveErr = transport.Leave(leaveContext, credential)
		cancel()
	}
	closeErr := transport.Close()
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
