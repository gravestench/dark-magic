// Package clientsession owns the transport-facing client side of one remote authoritative game
// session without depending on presentation or input implementations.
package clientsession

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// Connect verifies a Realm assignment, pins transport identity, consumes its one-use ticket, and
// returns a session only after runtime and admitted character identity agree with authority.
func Connect(
	ctx context.Context,
	assignment realm.JoinAssignment,
	tlsConfig *tls.Config,
) (*Session, error) {
	if tlsConfig == nil || strings.TrimSpace(assignment.Ticket) == "" ||
		strings.TrimSpace(assignment.GameID) == "" {
		return nil, ErrAssignment
	}

	transport, digest, err := dialVerified(
		ctx,
		assignment.GameID,
		assignment.Endpoint,
		assignment.Runtime,
		tlsConfig,
	)
	if err != nil {
		return nil, err
	}

	session, err := joinVerified(
		ctx,
		transport,
		assignment.GameID,
		assignment.Runtime,
		digest,
		assignment.Ticket,
		"",
	)
	if err != nil {
		return nil, err
	}

	session.setReconnectTarget(assignment.GameID, assignment.Endpoint, tlsConfig)

	return session, nil
}

// ConnectSelfHosted admits only the selected durable profile to an explicitly configured host. The
// profile credential obtains a one-use gameplay ticket before entering the ordinary verified join path.
func ConnectSelfHosted(
	ctx context.Context,
	assignment SelfHostedAssignment,
	tlsConfig *tls.Config,
	profile *d2save.Store,
) (*Session, error) {
	character, err := selectedProfileCharacter(profile)
	if err != nil {
		return nil, err
	}

	transport, digest, err := dialVerified(
		ctx,
		assignment.GameID,
		assignment.Endpoint,
		assignment.Runtime,
		tlsConfig,
	)
	if err != nil {
		return nil, err
	}

	ticket, err := admitSelectedProfile(ctx, transport, assignment.ProfileCredential, character)
	if err != nil {
		_ = transport.Close()

		return nil, err
	}

	session, err := joinVerified(
		ctx,
		transport,
		assignment.GameID,
		assignment.Runtime,
		digest,
		ticket,
		character.ID,
	)
	if err != nil {
		return nil, err
	}

	session.setReconnectTarget(assignment.GameID, assignment.Endpoint, tlsConfig)

	return session, nil
}

// selectedProfileCharacter requires explicit durable selection before self-hosted admission.
func selectedProfileCharacter(profile *d2save.Store) (d2save.Character, error) {
	if profile == nil {
		return d2save.Character{}, ErrAssignment
	}

	character, selected := profile.Selected()
	if !selected {
		return d2save.Character{}, ErrAssignment
	}

	return character, nil
}

// admitSelectedProfile sends the canonical save offer and returns the authority-issued gameplay ticket.
func admitSelectedProfile(
	ctx context.Context,
	transport *sessionquic.Client,
	credential string,
	character d2save.Character,
) (string, error) {
	offer, err := d2save.EncodeCharacterOffer(character)
	if err != nil {
		return "", err
	}

	return transport.AdmitProfile(ctx, credential, offer)
}

// setReconnectTarget retains endpoint trust material privately for recovery of the same logical game.
func (session *Session) setReconnectTarget(
	gameID string,
	endpoint realm.GameEndpoint,
	tlsConfig *tls.Config,
) {
	session.gameID = gameID
	session.endpoint = endpoint

	if tlsConfig != nil {
		session.tlsConfig = tlsConfig.Clone()
	}
}

// dialVerified validates endpoint and runtime identity, then adds assignment fingerprint pinning to the
// caller's TLS policy before any gameplay credential crosses the transport.
func dialVerified(
	ctx context.Context,
	gameID string,
	endpoint realm.GameEndpoint,
	identity simulation.RuntimeIdentity,
	tlsConfig *tls.Config,
) (*sessionquic.Client, string, error) {
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

	verifiedTLS, err := assignmentTLS(tlsConfig, endpoint.TLSFingerprint)
	if err != nil {
		return nil, "", err
	}

	transport, err := sessionquic.Dial(ctx, endpoint.Address, verifiedTLS)

	return transport, digest, err
}

// assignmentTLS composes optional Realm fingerprint pinning with the caller's existing verifier.
func assignmentTLS(base *tls.Config, fingerprint string) (*tls.Config, error) {
	var expected []byte

	var err error
	if strings.TrimSpace(fingerprint) != "" {
		expected, err = parseFingerprint(fingerprint)
		if err != nil {
			return nil, err
		}
	}

	verified := base.Clone()
	previousVerify := verified.VerifyPeerCertificate
	verified.VerifyPeerCertificate = pinnedCertificateVerifier(expected, previousVerify)

	return verified, nil
}

// pinnedCertificateVerifier checks the leaf digest in constant time before preserving the base verifier.
func pinnedCertificateVerifier(
	expected []byte,
	previous func([][]byte, [][]*x509.Certificate) error,
) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, chains [][]*x509.Certificate) error {
		if len(rawCerts) == 0 {
			return ErrAssignment
		}

		actual := sha256.Sum256(rawCerts[0])
		if len(expected) > 0 && subtle.ConstantTimeCompare(actual[:], expected) != 1 {
			return fmt.Errorf("%w: TLS fingerprint differs", ErrAssignment)
		}

		if previous != nil {
			return previous(rawCerts, chains)
		}

		return nil
	}
}

// joinVerified consumes the ticket and validates the complete initial response before publishing Session.
func joinVerified(
	ctx context.Context,
	transport *sessionquic.Client,
	gameID string,
	identity simulation.RuntimeIdentity,
	digest string,
	ticket string,
	characterID string,
) (*Session, error) {
	joined, err := transport.Join(ctx, gameserver.JoinRequest{
		Version:    gameserver.SessionProtocolVersion,
		Credential: ticket,
		Identity:   identity,
	})
	if err != nil {
		_ = transport.Close()

		return nil, err
	}

	view, err := validateInitialJoin(joined, gameID, digest, characterID)
	if err != nil {
		_ = transport.Close()

		return nil, err
	}

	session := newJoinedSession(transport, identity, joined, view)

	return session, nil
}

// validateInitialJoin binds protocol, game, runtime, admitted character, and projected character together.
func validateInitialJoin(
	joined gameserver.JoinResponse,
	gameID string,
	digest string,
	characterID string,
) (playeradapter.ClientView, error) {
	if joined.Admission.SessionID != gameID || joined.Admission.IdentityHash != digest ||
		joined.Snapshot.Version != gameserver.SessionProtocolVersion {
		return playeradapter.ClientView{}, ErrAssignment
	}

	view, err := decodeView(joined.Snapshot)
	if err != nil {
		return playeradapter.ClientView{}, err
	}

	admittedCharacter := strings.TrimSpace(joined.Admission.CharacterID)
	if admittedCharacter == "" || view.HUD.Player.CharacterID != admittedCharacter ||
		characterID != "" && admittedCharacter != characterID {
		return playeradapter.ClientView{}, ErrAssignment
	}

	return view, nil
}

// newJoinedSession installs the first canonical projection and publishes revision one atomically.
func newJoinedSession(
	transport *sessionquic.Client,
	identity simulation.RuntimeIdentity,
	joined gameserver.JoinResponse,
	view playeradapter.ClientView,
) *Session {
	session := &Session{
		transport:     transport,
		credential:    joined.Credential,
		identity:      identity,
		Admission:     joined,
		HUD:           view.HUD,
		World:         view.World,
		Private:       view.Private,
		Party:         view.Party,
		Events:        view.Events,
		reliableHUD:   view.HUD,
		reliableWorld: view.World,
		viewRevision:  1,
		eventEpoch:    1,
		pending:       make(map[uint64]gameserver.CommandIntent),
	}

	session.observeSnapshotLocked(joined.Snapshot, time.Now())
	session.publishPresentationLocked()

	return session
}
