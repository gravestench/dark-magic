// Package clientsession owns the transport-facing client side of one remote
// authoritative game session without depending on presentation or input.
package clientsession

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/networkclock"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

var ErrAssignment = errors.New("client session: invalid server assignment")
var ErrStaleCorrection = errors.New("client session: stale or conflicting correction")

type Session struct {
	mu             sync.Mutex
	reconnectMu    sync.RWMutex
	transport      *sessionquic.Client
	credential     gameserver.SessionCredential
	identity       simulation.RuntimeIdentity
	closed         bool
	Admission      gameserver.JoinResponse
	HUD            playeradapter.HUD
	World          playeradapter.WorldView
	Private        playeradapter.PrivateView
	Party          playeradapter.PartyView
	Events         playeradapter.EventView
	reliableHUD    playeradapter.HUD
	reliableWorld  playeradapter.WorldView
	viewRevision   uint64
	eventEpoch     uint64
	presentation   atomic.Pointer[PresentationSnapshot]
	pending        map[uint64]gameserver.CommandIntent
	clock          *networkclock.Clock
	gameID         string
	endpoint       realm.GameEndpoint
	tlsConfig      *tls.Config
	reconnectNonce string
}

// PresentationSnapshot is immutable after publication. The network worker
// builds and atomically swaps it; the render thread may retain and read it
// without copying the world projection every frame.
type PresentationSnapshot struct {
	HUD        playeradapter.HUD
	World      playeradapter.WorldView
	Private    playeradapter.PrivateView
	Party      playeradapter.PartyView
	Events     playeradapter.EventView
	Revision   uint64
	EventEpoch uint64
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
	return clonePresentationView(session.HUD, session.World)
}

// PresentationView returns a coherent projected view plus a monotonic content
// revision. Reliable lifecycle metadata may advance the revision at the same
// simulation tick as an already received transform datagram.
func (session *Session) PresentationView() (playeradapter.HUD, playeradapter.WorldView, uint64) {
	session.mu.Lock()
	defer session.mu.Unlock()
	hud, world := clonePresentationView(session.HUD, session.World)
	return hud, world, session.viewRevision
}

func (session *Session) PresentationSnapshot() *PresentationSnapshot {
	if session == nil {
		return nil
	}
	if snapshot := session.presentation.Load(); snapshot != nil {
		return snapshot
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	session.publishPresentationLocked()
	return session.presentation.Load()
}

func (session *Session) publishPresentationLocked() {
	session.presentation.Store(&PresentationSnapshot{
		HUD: session.HUD, World: session.World, Private: session.Private,
		Party: clonePartyView(session.Party), Events: cloneEventView(session.Events),
		Revision: session.viewRevision, EventEpoch: session.eventEpoch,
	})
}

func clonePartyView(view playeradapter.PartyView) playeradapter.PartyView {
	view.Roster = append([]playeradapter.PartyRosterEntry(nil), view.Roster...)
	return view
}

func cloneEventView(view playeradapter.EventView) playeradapter.EventView {
	view.Events = append([]playeradapter.SemanticEvent(nil), view.Events...)
	for index := range view.Events {
		if view.Events[index].Cast != nil {
			value := *view.Events[index].Cast
			view.Events[index].Cast = &value
		}
		if view.Events[index].State != nil {
			value := *view.Events[index].State
			view.Events[index].State = &value
		}
	}
	return view
}

func clonePresentationView(hud playeradapter.HUD, world playeradapter.WorldView) (playeradapter.HUD, playeradapter.WorldView) {
	hud.Belt.Slots = append([]string(nil), hud.Belt.Slots...)
	hud.Skills.Learned = append([]playeradapter.HUDLearnedSkill(nil), hud.Skills.Learned...)
	world.Entities = append([]playeradapter.WorldEntity(nil), world.Entities...)
	world.Missiles = append([]playeradapter.WorldMissile(nil), world.Missiles...)
	world.States = append([]playeradapter.WorldState(nil), world.States...)
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

// PrivateView returns the authenticated player's owner-only projection. It is
// separate from View so public world consumers cannot accidentally grow a
// dependency on inventory or interaction state.
func (session *Session) PrivateView() playeradapter.PrivateView {
	session.mu.Lock()
	defer session.mu.Unlock()
	view := session.Private
	view.Items.Items = append([]playeradapter.ItemEntityView(nil), view.Items.Items...)
	if view.Interaction.Target != nil {
		target := *view.Interaction.Target
		view.Interaction.Target = &target
	}
	return view
}

// NextInputTick schedules the next fixed input sample inside the authority's
// advertised two-tick admission lead. Presentation may extrapolate farther
// during a host stall, but input must remain admissible; stale samples arrive
// late and use the authority's rollback window instead.
func (session *Session) NextInputTick(now time.Time) uint64 {
	session.mu.Lock()
	defer session.mu.Unlock()
	timeline := session.timelineLocked(now)
	return timeline.LatestServerTick + 2
}

// NetworkTimeline exposes the two client times without exposing mutable clock
// state. Prediction is for the owning player; interpolation is for remote
// presentation and intentionally trails prediction.
func (session *Session) NetworkTimeline(now time.Time) networkclock.Timeline {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.timelineLocked(now)
}

func (session *Session) StepDuration() time.Duration {
	session.mu.Lock()
	defer session.mu.Unlock()
	return time.Duration(session.Admission.Snapshot.StepNanos)
}

func Connect(ctx context.Context, assignment realm.JoinAssignment, tlsConfig *tls.Config) (*Session, error) {
	if tlsConfig == nil || strings.TrimSpace(assignment.Ticket) == "" || strings.TrimSpace(assignment.GameID) == "" {
		return nil, ErrAssignment
	}
	transport, digest, err := dialVerified(ctx, assignment.GameID, assignment.Endpoint, assignment.Runtime, tlsConfig)
	if err != nil {
		return nil, err
	}
	session, err := joinVerified(ctx, transport, assignment.GameID, assignment.Runtime, digest, assignment.Ticket, "")
	if err == nil {
		session.setReconnectTarget(assignment.GameID, assignment.Endpoint, tlsConfig)
	}
	return session, err
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
	session, err := joinVerified(ctx, transport, assignment.GameID, assignment.Runtime, digest, ticket, character.ID)
	if err == nil {
		session.setReconnectTarget(assignment.GameID, assignment.Endpoint, tlsConfig)
	}
	return session, err
}

func (session *Session) setReconnectTarget(gameID string, endpoint realm.GameEndpoint, tlsConfig *tls.Config) {
	session.gameID, session.endpoint = gameID, endpoint
	if tlsConfig != nil {
		session.tlsConfig = tlsConfig.Clone()
	}
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
	if strings.TrimSpace(joined.Admission.CharacterID) == "" || view.HUD.Player.CharacterID != joined.Admission.CharacterID ||
		characterID != "" && joined.Admission.CharacterID != characterID {
		_ = transport.Close()
		return nil, ErrAssignment
	}
	now := time.Now()
	session := &Session{transport: transport, credential: joined.Credential, identity: identity, Admission: joined, HUD: view.HUD, World: view.World, Private: view.Private, Party: view.Party, Events: view.Events,
		reliableHUD: view.HUD, reliableWorld: view.World, viewRevision: 1, eventEpoch: 1, pending: make(map[uint64]gameserver.CommandIntent)}
	session.observeSnapshotLocked(joined.Snapshot, now)
	session.publishPresentationLocked()
	return session, nil
}

func (session *Session) Submit(ctx context.Context, intent gameserver.CommandIntent) error {
	// Independent reliable QUIC streams may be in flight together, but a
	// credential rotation must wait until every request using the old credential
	// has completed. Never retain session.mu across network I/O: presentation
	// reads pending input and clock state every render frame.
	session.reconnectMu.RLock()
	defer session.reconnectMu.RUnlock()
	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()
		return errors.New("client session: closed")
	}
	if intent.ObservedServerTick == 0 {
		intent.ObservedServerTick = session.timelineLocked(time.Now()).Prediction.Tick
	}
	if intent.TargetTick == 0 {
		intent.TargetTick = intent.ObservedServerTick + 2
	}
	transport, credential := session.transport, session.credential
	session.mu.Unlock()
	if err := transport.Submit(ctx, credential, intent); err != nil {
		return err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return nil
	}
	if session.pending == nil {
		session.pending = make(map[uint64]gameserver.CommandIntent)
	}
	session.pending[intent.Sequence] = intent
	return nil
}

// StageInput makes a locally produced input visible to prediction before the
// asynchronous transport writer completes. The same sequence may be staged
// again only when every authoritative field is identical.
func (session *Session) StageInput(intent gameserver.CommandIntent) error {
	if intent.Sequence == 0 || intent.Kind == "" {
		return errors.New("client session: invalid staged input")
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return errors.New("client session: closed")
	}
	if existing, found := session.pending[intent.Sequence]; found {
		if existing.ObservedServerTick != intent.ObservedServerTick || existing.TargetTick != intent.TargetTick ||
			existing.Kind != intent.Kind || !bytes.Equal(existing.Payload, intent.Payload) {
			return errors.New("client session: conflicting staged input sequence")
		}
		return nil
	}
	intent.Payload = append(json.RawMessage(nil), intent.Payload...)
	session.pending[intent.Sequence] = intent
	return nil
}

// DiscardInput removes a locally staged command rejected before admission.
func (session *Session) DiscardInput(sequence uint64) {
	session.mu.Lock()
	delete(session.pending, sequence)
	session.mu.Unlock()
}

func (session *Session) timelineLocked(now time.Time) networkclock.Timeline {
	if session.clock == nil {
		session.observeSnapshotLocked(session.Admission.Snapshot, now)
	}
	return session.clock.Timeline(now)
}

func (session *Session) observeSnapshotLocked(snapshot gameserver.Snapshot, receivedAt time.Time) {
	if session.clock == nil {
		session.clock = networkclock.New(networkclock.Config{UpdateInterval: sessionquic.TransformInterval})
	}
	stats := sessionquic.NetworkStats{}
	if session.transport != nil {
		stats = session.transport.NetworkStats()
	}
	session.clock.Observe(networkclock.Sample{
		Tick: snapshot.Tick, Step: time.Duration(snapshot.StepNanos), ReceivedAt: receivedAt,
		SmoothedRTT: stats.SmoothedRTT, RTTVariation: stats.RTTVariation,
	})
}

// PendingInputs returns defensive, sequence-ordered input history retained for
// acknowledgement and future local prediction replay.
func (session *Session) PendingInputs() []gameserver.CommandIntent {
	session.mu.Lock()
	defer session.mu.Unlock()
	result := make([]gameserver.CommandIntent, 0, len(session.pending))
	sequences := make([]uint64, 0, len(session.pending))
	for sequence := range session.pending {
		sequences = append(sequences, sequence)
	}
	sort.Slice(sequences, func(i, j int) bool { return sequences[i] < sequences[j] })
	for _, sequence := range sequences {
		intent := session.pending[sequence]
		intent.Payload = append(json.RawMessage(nil), intent.Payload...)
		result = append(result, intent)
	}
	return result
}

func (session *Session) Reconnect(ctx context.Context) error {
	session.reconnectMu.Lock()
	defer session.reconnectMu.Unlock()
	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()
		return errors.New("client session: closed")
	}
	if session.reconnectNonce == "" {
		var nonce [16]byte
		if _, err := rand.Read(nonce[:]); err != nil {
			session.mu.Unlock()
			return err
		}
		session.reconnectNonce = hex.EncodeToString(nonce[:])
	}
	transport, credential, identity, nonce := session.transport, session.credential, session.identity, session.reconnectNonce
	originalTransport := transport
	identityHash := session.Admission.Admission.IdentityHash
	characterID, owner := session.Admission.Admission.CharacterID, session.reliableHUD.Player
	gameID, endpoint, tlsConfig := session.gameID, session.endpoint, session.tlsConfig
	session.mu.Unlock()
	request := gameserver.ReconnectRequest{Credential: credential, Identity: identity, Nonce: nonce}
	directContext, cancelDirect := context.WithTimeout(ctx, 300*time.Millisecond)
	joined, err := transport.Reconnect(directContext, request)
	cancelDirect()
	if err != nil && tlsConfig != nil {
		var replacement *sessionquic.Client
		replacement, _, err = dialVerified(ctx, gameID, endpoint, identity, tlsConfig)
		if err == nil {
			joined, err = replacement.Reconnect(ctx, request)
		}
		if err != nil {
			if replacement != nil {
				_ = replacement.Close()
			}
			return err
		}
		transport = replacement
	}
	if err != nil {
		return err
	}
	view, err := decodeView(joined.Snapshot)
	if err != nil {
		if transport != originalTransport {
			_ = transport.Close()
		}
		return err
	}
	if joined.Admission.SessionID != gameID || joined.Admission.IdentityHash != identityHash ||
		joined.Admission.CharacterID != characterID || validateOwnerIdentity(owner, view.HUD.Player) != nil {
		if transport != originalTransport {
			_ = transport.Close()
		}
		return ErrAssignment
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed || session.credential != credential {
		if transport != session.transport {
			_ = transport.Close()
		}
		return ErrStaleCorrection
	}
	oldTransport := session.transport
	session.credential, session.Admission, session.HUD, session.World, session.Private, session.Party, session.Events = joined.Credential, joined, view.HUD, view.World, view.Private, view.Party, view.Events
	session.reliableHUD, session.reliableWorld = view.HUD, view.World
	session.viewRevision++
	session.eventEpoch++
	session.transport = transport
	session.reconnectNonce = ""
	session.observeSnapshotLocked(joined.Snapshot, time.Now())
	session.discardAcknowledgedLocked(joined.Snapshot.AcknowledgedInput)
	session.publishPresentationLocked()
	if oldTransport != transport {
		_ = oldTransport.Close()
	}
	return nil
}

// Reassign consumes a fresh Realm ticket for a replacement authority while
// retaining the client-side session object used by input and presentation
// loops. The game, runtime, character, and authenticated player identities may
// not change. Unacknowledged exact inputs are resubmitted after the atomic
// transport swap; the recovered authority suppresses inputs it already owns.
func (session *Session) Reassign(ctx context.Context, assignment realm.JoinAssignment, tlsConfig *tls.Config) error {
	if session == nil || ctx == nil || tlsConfig == nil || strings.TrimSpace(assignment.Ticket) == "" {
		return ErrAssignment
	}
	session.reconnectMu.Lock()
	session.mu.Lock()
	if session.closed || assignment.GameID != session.gameID {
		session.mu.Unlock()
		session.reconnectMu.Unlock()
		return ErrAssignment
	}
	expectedHash := session.Admission.Admission.IdentityHash
	expectedCharacter := session.Admission.Admission.CharacterID
	expectedOwner := session.reliableHUD.Player
	expectedIdentity := session.identity
	session.mu.Unlock()
	digest, err := assignment.Runtime.Digest()
	if err != nil || digest != expectedHash {
		session.reconnectMu.Unlock()
		return ErrAssignment
	}
	replacement, err := Connect(ctx, assignment, tlsConfig)
	if err != nil {
		session.reconnectMu.Unlock()
		return err
	}
	replacement.mu.Lock()
	viewOwner := replacement.reliableHUD.Player
	if replacement.Admission.Admission.IdentityHash != expectedHash ||
		replacement.Admission.Admission.CharacterID != expectedCharacter ||
		validateOwnerIdentity(expectedOwner, viewOwner) != nil {
		replacement.mu.Unlock()
		session.reconnectMu.Unlock()
		_ = replacement.Close(context.Background())
		return ErrAssignment
	}
	newTransport := replacement.transport
	replacement.transport = nil
	replacement.closed = true
	replacement.mu.Unlock()

	session.mu.Lock()
	if session.closed || session.gameID != assignment.GameID {
		session.mu.Unlock()
		session.reconnectMu.Unlock()
		_ = newTransport.Close()
		return ErrStaleCorrection
	}
	oldTransport := session.transport
	session.transport, session.credential = newTransport, replacement.credential
	session.identity, session.endpoint, session.tlsConfig = expectedIdentity, assignment.Endpoint, tlsConfig.Clone()
	session.Admission, session.HUD, session.World, session.Private, session.Party, session.Events = replacement.Admission, replacement.HUD, replacement.World, replacement.Private, replacement.Party, replacement.Events
	session.reliableHUD, session.reliableWorld = replacement.reliableHUD, replacement.reliableWorld
	session.clock, session.reconnectNonce = replacement.clock, ""
	session.viewRevision++
	session.eventEpoch++
	session.discardAcknowledgedLocked(replacement.Admission.Snapshot.AcknowledgedInput)
	pending := make([]gameserver.CommandIntent, 0, len(session.pending))
	for _, intent := range session.pending {
		pending = append(pending, intent)
	}
	sort.Slice(pending, func(i, j int) bool { return pending[i].Sequence < pending[j].Sequence })
	session.publishPresentationLocked()
	session.mu.Unlock()
	session.reconnectMu.Unlock()
	_ = oldTransport.Close()
	for _, intent := range pending {
		if err := session.Submit(ctx, intent); err != nil {
			return err
		}
	}
	return nil
}

// Refresh fetches one reliable canonical correction and returns the public
// world delta from the previously installed view.
func (session *Session) Refresh(ctx context.Context) (playeradapter.WorldDelta, error) {
	session.reconnectMu.RLock()
	defer session.reconnectMu.RUnlock()
	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()
		return playeradapter.WorldDelta{}, errors.New("client session: closed")
	}
	transport, credential := session.transport, session.credential
	session.mu.Unlock()
	snapshot, err := transport.Refresh(ctx, credential)
	if err != nil {
		return playeradapter.WorldDelta{}, err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return playeradapter.WorldDelta{}, errors.New("client session: closed")
	}
	return session.applyCorrection(snapshot)
}

// Watch streams reliable canonical corrections until cancellation. Only one
// correction is buffered, so a slow consumer propagates backpressure.
func (session *Session) Watch(ctx context.Context) (<-chan playeradapter.WorldDelta, <-chan error, error) {
	watchContext, cancelWatch := context.WithCancel(ctx)
	session.reconnectMu.RLock()
	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()
		session.reconnectMu.RUnlock()
		cancelWatch()
		return nil, nil, errors.New("client session: closed")
	}
	transport, credential := session.transport, session.credential
	session.mu.Unlock()
	snapshots, transportErrors, err := transport.Watch(watchContext, credential)
	if err != nil {
		session.reconnectMu.RUnlock()
		cancelWatch()
		return nil, nil, err
	}
	transforms, transformErrors, err := transport.WatchTransforms(watchContext, credential)
	session.reconnectMu.RUnlock()
	if err != nil {
		cancelWatch()
		return nil, nil, err
	}
	deltas := make(chan playeradapter.WorldDelta, 1)
	errorsOut := make(chan error, 1)
	go func() {
		defer cancelWatch()
		defer close(deltas)
		defer close(errorsOut)
		for snapshots != nil || transportErrors != nil || transforms != nil || transformErrors != nil {
			select {
			case snapshot, open := <-snapshots:
				if !open {
					snapshots = nil
					continue
				}
				view, decodeErr := decodeView(snapshot)
				if decodeErr != nil {
					errorsOut <- decodeErr
					return
				}
				session.mu.Lock()
				delta, applyErr := session.applyDecodedCorrection(snapshot, view)
				session.mu.Unlock()
				if applyErr != nil {
					errorsOut <- applyErr
					return
				}
				publishLatestDelta(ctx, deltas, delta)
			case streamErr, open := <-transportErrors:
				if !open {
					transportErrors = nil
					continue
				}
				if streamErr != nil {
					errorsOut <- streamErr
					return
				}
			case frame, open := <-transforms:
				if !open {
					transforms = nil
					continue
				}
				session.mu.Lock()
				delta, applyErr := session.applyTransform(frame)
				session.mu.Unlock()
				if applyErr != nil {
					errorsOut <- applyErr
					return
				}
				publishLatestDelta(ctx, deltas, delta)
			case streamErr, open := <-transformErrors:
				if !open {
					transformErrors = nil
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
	return session.applyDecodedCorrection(snapshot, view)
}

func (session *Session) applyDecodedCorrection(snapshot gameserver.Snapshot, view playeradapter.ClientView) (playeradapter.WorldDelta, error) {
	if err := validateCorrection(session.Admission.Snapshot, snapshot); err != nil {
		return playeradapter.WorldDelta{}, err
	}
	if err := validateOwnerIdentity(session.reliableHUD.Player, view.HUD.Player); err != nil {
		return playeradapter.WorldDelta{}, err
	}
	previousReliable := session.reliableWorld
	if previousReliable.Version == 0 {
		previousReliable = session.World
	}
	delta := playeradapter.DiffWorldView(previousReliable, view.World)
	session.Admission.Snapshot, session.Private, session.Party, session.Events = snapshot, view.Private, view.Party, view.Events
	session.reliableHUD, session.reliableWorld = view.HUD, view.World
	session.HUD, session.World = mergeReliablePresentation(view.HUD, view.World, session.HUD, session.World)
	session.viewRevision++
	session.observeSnapshotLocked(snapshot, time.Now())
	session.discardAcknowledgedLocked(snapshot.AcknowledgedInput)
	session.publishPresentationLocked()
	return delta, nil
}

func validateOwnerIdentity(expected, actual playeradapter.HUDIdentity) error {
	if expected.PlayerID == "" || expected.CharacterID == "" ||
		actual.PlayerID != expected.PlayerID || actual.CharacterID != expected.CharacterID {
		return ErrAssignment
	}
	return nil
}

// applyTransform requires session.mu. Unknown hashes wait for reliable
// lifecycle metadata; collisions reject the frame rather than aliasing IDs.
func (session *Session) applyTransform(frame sessionquic.TransformFrame) (playeradapter.WorldDelta, error) {
	if frame.Tick <= session.World.Tick {
		return playeradapter.WorldDelta{Version: playeradapter.WorldDeltaVersion, BaseTick: session.World.Tick, Tick: session.World.Tick}, nil
	}
	before := session.World
	world := session.World
	world.Entities = append([]playeradapter.WorldEntity(nil), session.World.Entities...)
	world.Missiles = append([]playeradapter.WorldMissile(nil), session.World.Missiles...)
	indexes := make(map[uint64]int, len(world.Entities)+len(world.Missiles))
	for index, entity := range world.Entities {
		hash := sessionquic.PublicIDHash(entity.ID)
		if _, duplicate := indexes[hash]; duplicate {
			return playeradapter.WorldDelta{}, ErrStaleCorrection
		}
		indexes[hash] = index
	}
	missileIndexes := make(map[uint64]int, len(world.Missiles))
	for index, missile := range world.Missiles {
		hash := sessionquic.PublicIDHash(missile.ID)
		if _, duplicate := indexes[hash]; duplicate {
			return playeradapter.WorldDelta{}, ErrStaleCorrection
		}
		if _, duplicate := missileIndexes[hash]; duplicate {
			return playeradapter.WorldDelta{}, ErrStaleCorrection
		}
		missileIndexes[hash] = index
	}
	for _, transform := range frame.Entities {
		if index, found := indexes[transform.IDHash]; found {
			world.Entities[index].Position = playeradapter.HUDPosition{X: transform.X, Y: transform.Y}
			world.Entities[index].Direction = int64(transform.Direction)
			world.Entities[index].Mode = strings.TrimRight(string(transform.Mode[:]), "\x00")
			world.Entities[index].AnimationStartTick = transform.AnimationStartTick
			continue
		}
		if index, found := missileIndexes[transform.IDHash]; found {
			world.Missiles[index].Position = playeradapter.HUDPosition{X: transform.X, Y: transform.Y}
		}
	}
	world.Tick = frame.Tick
	world.Origin = playeradapter.HUDPosition{X: frame.OwnerX, Y: frame.OwnerY}
	hud := session.HUD
	hud.Tick = frame.Tick
	hud.Position = world.Origin
	hud.Movement.Velocity = playeradapter.HUDPosition{X: frame.VelocityX, Y: frame.VelocityY}
	hud.Animation.Direction = int64(frame.OwnerDirection)
	hud.Animation.Mode = strings.TrimRight(string(frame.OwnerMode[:]), "\x00")
	hud.Animation.StartTick = frame.OwnerAnimationStartTick
	session.HUD, session.World = hud, world
	session.viewRevision++
	session.publishPresentationLocked()
	if session.clock == nil {
		session.clock = networkclock.New(networkclock.Config{UpdateInterval: sessionquic.TransformInterval})
	}
	stats := sessionquic.NetworkStats{}
	if session.transport != nil {
		stats = session.transport.NetworkStats()
	}
	session.clock.Observe(networkclock.Sample{
		Tick: frame.Tick, Step: time.Duration(session.Admission.Snapshot.StepNanos), ReceivedAt: time.Now(),
		SmoothedRTT: stats.SmoothedRTT, RTTVariation: stats.RTTVariation,
	})
	return playeradapter.DiffWorldView(before, world), nil
}

func mergeReliablePresentation(reliableHUD playeradapter.HUD, reliableWorld playeradapter.WorldView, currentHUD playeradapter.HUD, currentWorld playeradapter.WorldView) (playeradapter.HUD, playeradapter.WorldView) {
	if currentWorld.Tick <= reliableWorld.Tick {
		return reliableHUD, reliableWorld
	}
	transforms := make(map[string]playeradapter.WorldEntity, len(currentWorld.Entities))
	for _, entity := range currentWorld.Entities {
		transforms[entity.ID] = entity
	}
	reliableWorld.Entities = append([]playeradapter.WorldEntity(nil), reliableWorld.Entities...)
	for index := range reliableWorld.Entities {
		if current, found := transforms[reliableWorld.Entities[index].ID]; found {
			reliableWorld.Entities[index].Position = current.Position
			reliableWorld.Entities[index].Direction = current.Direction
			reliableWorld.Entities[index].Mode = current.Mode
			reliableWorld.Entities[index].AnimationStartTick = current.AnimationStartTick
		}
	}
	missileTransforms := make(map[string]playeradapter.HUDPosition, len(currentWorld.Missiles))
	for _, missile := range currentWorld.Missiles {
		missileTransforms[missile.ID] = missile.Position
	}
	reliableWorld.Missiles = append([]playeradapter.WorldMissile(nil), reliableWorld.Missiles...)
	for index := range reliableWorld.Missiles {
		if position, found := missileTransforms[reliableWorld.Missiles[index].ID]; found {
			reliableWorld.Missiles[index].Position = position
		}
	}
	reliableWorld.Tick = currentWorld.Tick
	reliableWorld.Origin = currentWorld.Origin
	reliableHUD.Tick = currentHUD.Tick
	reliableHUD.Position = currentHUD.Position
	reliableHUD.Movement.Velocity = currentHUD.Movement.Velocity
	reliableHUD.Animation = currentHUD.Animation
	return reliableHUD, reliableWorld
}

func publishLatestDelta(ctx context.Context, deltas chan playeradapter.WorldDelta, delta playeradapter.WorldDelta) {
	select {
	case deltas <- delta:
		return
	default:
	}
	select {
	case <-deltas:
	default:
	}
	select {
	case deltas <- delta:
	case <-ctx.Done():
	}
}

func (session *Session) discardAcknowledgedLocked(acknowledged uint64) {
	for sequence := range session.pending {
		if sequence <= acknowledged {
			delete(session.pending, sequence)
		}
	}
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
	decoder := json.NewDecoder(bytes.NewReader(snapshot.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&view); err != nil {
		return playeradapter.ClientView{}, fmt.Errorf("%w: invalid ClientView/v%d", ErrAssignment, playeradapter.ClientViewVersion)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) || playeradapter.ValidateClientView(view, snapshot.Tick) != nil {
		return playeradapter.ClientView{}, fmt.Errorf("%w: invalid ClientView/v%d", ErrAssignment, playeradapter.ClientViewVersion)
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
