package gameserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const SessionProtocolVersion uint32 = 2

var (
	ErrAuthentication = errors.New("game server protocol: authentication failed")
	ErrProtocol       = errors.New("game server protocol: unsupported version")
	ErrRateLimit      = errors.New("game server protocol: rate limit exceeded")
)

const (
	commandBurst     = 64
	commandRate      = 32.0
	refreshBurst     = 4
	refreshRate      = 2.0
	joinReadyTimeout = 2 * time.Second
	joinReadyPoll    = 10 * time.Millisecond
	reconnectGrace   = 10 * time.Second
)

// Principal is the trusted result of authentication. PlayerID is the stable
// command identity used inside the authoritative session; clients never choose
// it in command messages.
type Principal struct {
	ID                  string
	CharacterID         string
	PlayerID            string
	CharacterRevision   uint64
	RuntimeIdentityHash string
}

type Authenticator interface {
	Authenticate(context.Context, string) (Principal, error)
}

// SnapshotProjector builds the public/private semantic view allowed for one
// player. It deliberately prevents the protocol from exposing raw ECS state.
type SnapshotProjector func(playerID string, checkpoint simulation.Checkpoint) (json.RawMessage, error)

type JoinRequest struct {
	Version    uint32
	Credential string
	Identity   simulation.RuntimeIdentity
}

type SessionCredential string

func (credential SessionCredential) String() string { return string(credential) }

type Snapshot struct {
	Version           uint32          `json:"version"`
	Tick              uint64          `json:"tick"`
	StepNanos         int64           `json:"step_nanos"`
	Checksum          string          `json:"checksum"`
	AcknowledgedInput uint64          `json:"acknowledged_input"`
	Payload           json.RawMessage `json:"payload"`
}

type JoinResponse struct {
	Credential SessionCredential
	Admission  gamesession.AdmissionToken
	Snapshot   Snapshot
}

type ReconnectRequest struct {
	Credential SessionCredential
	Identity   simulation.RuntimeIdentity
	Nonce      string `json:"nonce"`
}

// CommandIntent contains only client-controlled gameplay input. Identity and
// authority are supplied by Endpoint from authenticated connection state.
type CommandIntent struct {
	ObservedServerTick uint64          `json:"observed_server_tick"`
	TargetTick         uint64          `json:"target_tick"`
	Sequence           uint64          `json:"sequence"`
	Kind               string          `json:"kind"`
	Payload            json.RawMessage `json:"payload"`
}

type connection struct {
	principal            Principal
	admission            gamesession.AdmissionToken
	commands             tokenBucket
	refreshes            tokenBucket
	connected            bool
	disconnectGeneration uint64
}

type reconnectReplay struct {
	nonce    string
	response JoinResponse
}

type tokenBucket struct {
	tokens, capacity, rate float64
	updated                time.Time
}

// Endpoint is the transport-neutral authenticated boundary around one Host.
// HTTP, UDP, loopback, and legacy protocol adapters can all call this API.
type Endpoint struct {
	mu              sync.RWMutex
	snapshotMu      sync.Mutex
	host            *Host
	auth            Authenticator
	project         SnapshotProjector
	connections     map[string]connection
	now             func() time.Time
	snapshotPending func(error) bool
	checkpoint      simulation.Checkpoint
	leave           func(Principal) error
	connected       func(Principal)
	watches         map[string]bool
	reconnects      map[string]reconnectReplay
	after           func(time.Duration, func())
	reconnectGrace  time.Duration
}

// SetSnapshotPending identifies the one expected projection error while a
// trusted next-tick admission command is waiting to materialize its player.
func (endpoint *Endpoint) SetSnapshotPending(classify func(error) bool) {
	endpoint.snapshotPending = classify
}

// SetLeave installs the mod-owned membership cleanup command. Authentication
// and credential revocation remain protocol policy; the meaning of removing a
// live player and its owned entities belongs to the active game rules.
func (endpoint *Endpoint) SetLeave(leave func(Principal) error) { endpoint.leave = leave }

// SetConnected observes successful initial joins and reconnects. It is used by
// Realm worker lifecycle accounting; authentication remains owned here.
func (endpoint *Endpoint) SetConnected(connected func(Principal)) { endpoint.connected = connected }

func NewEndpoint(host *Host, auth Authenticator, project SnapshotProjector) (*Endpoint, error) {
	if host == nil || host.Session == nil || auth == nil || project == nil {
		return nil, errors.New("game server protocol: host, authenticator, and projector are required")
	}
	return &Endpoint{
		host: host, auth: auth, project: project, connections: make(map[string]connection),
		watches: make(map[string]bool), reconnects: make(map[string]reconnectReplay), now: time.Now, reconnectGrace: reconnectGrace,
		after: func(delay time.Duration, callback func()) { time.AfterFunc(delay, callback) },
	}, nil
}

func (endpoint *Endpoint) Join(ctx context.Context, request JoinRequest) (JoinResponse, error) {
	if request.Version != SessionProtocolVersion {
		return JoinResponse{}, ErrProtocol
	}
	principal, err := endpoint.auth.Authenticate(ctx, request.Credential)
	if err != nil || strings.TrimSpace(principal.ID) == "" || strings.TrimSpace(principal.CharacterID) == "" || strings.TrimSpace(principal.PlayerID) == "" {
		return JoinResponse{}, ErrAuthentication
	}
	if principal.RuntimeIdentityHash != "" && principal.RuntimeIdentityHash != endpoint.host.Allocation.IdentityHash {
		return JoinResponse{}, ErrAuthentication
	}
	admission, err := endpoint.host.Admit(principal.CharacterID, request.Identity)
	if err != nil {
		return JoinResponse{}, err
	}
	credential, err := newSessionCredential()
	if err != nil {
		return JoinResponse{}, fmt.Errorf("game server protocol: create credential: %w", err)
	}
	endpoint.mu.Lock()
	now := endpoint.now()
	endpoint.connections[string(credential)] = connection{principal: principal, admission: admission,
		commands: newTokenBucket(commandBurst, commandRate, now), refreshes: newTokenBucket(refreshBurst, refreshRate, now), connected: true}
	endpoint.mu.Unlock()
	snapshot, err := endpoint.joinSnapshot(ctx, principal.PlayerID)
	if err != nil {
		endpoint.mu.Lock()
		delete(endpoint.connections, string(credential))
		endpoint.mu.Unlock()
		return JoinResponse{}, err
	}
	if endpoint.connected != nil {
		endpoint.connected(principal)
	}
	return JoinResponse{Credential: credential, Admission: admission, Snapshot: snapshot}, nil
}

func (endpoint *Endpoint) joinSnapshot(ctx context.Context, playerID string) (Snapshot, error) {
	deadline := time.NewTimer(joinReadyTimeout)
	defer deadline.Stop()
	for {
		snapshot, err := endpoint.snapshot(playerID)
		if err == nil || endpoint.snapshotPending == nil || !endpoint.snapshotPending(err) {
			return snapshot, err
		}
		timer := time.NewTimer(joinReadyPoll)
		select {
		case <-ctx.Done():
			timer.Stop()
			return Snapshot{}, ctx.Err()
		case <-deadline.C:
			timer.Stop()
			return Snapshot{}, fmt.Errorf("game server protocol: player admission readiness: %w", err)
		case <-timer.C:
		}
	}
}

func (endpoint *Endpoint) Submit(credential SessionCredential, intent CommandIntent) error {
	member, err := endpoint.consume(credential, false)
	if err != nil {
		return err
	}
	target := intent.TargetTick
	if target == 0 {
		target = intent.ObservedServerTick + 2
	}
	command := simulation.Command{Tick: target, Player: member.principal.PlayerID,
		Authority: simulation.AuthorityPlayer, Sequence: intent.Sequence, Kind: intent.Kind, Payload: intent.Payload}
	_, err = endpoint.host.Session.SubmitNetwork(command)
	if errors.Is(err, gamesession.ErrCommandSequence) {
		if accepted, found := endpoint.host.Session.AcceptedNetworkCommand(command.Player, command.Sequence); found &&
			accepted.Tick == command.Tick && accepted.Kind == command.Kind && bytes.Equal(accepted.Payload, command.Payload) {
			return nil
		}
	}
	return err
}

// Refresh returns the latest canonical per-player correction projection.
func (endpoint *Endpoint) Refresh(credential SessionCredential) (Snapshot, error) {
	member, err := endpoint.consume(credential, true)
	if err != nil {
		return Snapshot{}, err
	}
	return endpoint.snapshot(member.principal.PlayerID)
}

// Observe returns a correction for an authenticated long-lived server-paced
// watch stream. Unlike client-paced Refresh, it does not consume the unary
// refresh token bucket: the server's correction ticker is already the rate
// limiter, and charging both would terminate every healthy watch stream.
func (endpoint *Endpoint) Observe(credential SessionCredential) (Snapshot, error) {
	member, err := endpoint.connection(credential)
	if err != nil {
		return Snapshot{}, err
	}
	return endpoint.snapshot(member.principal.PlayerID)
}

// BeginWatch permits one server-paced correction stream per membership. A
// client can still issue bounded unary Refresh calls, but cannot multiply
// projection work by opening arbitrary concurrent long-lived streams.
func (endpoint *Endpoint) BeginWatch(credential SessionCredential) error {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()
	key := string(credential)
	if member, found := endpoint.connections[key]; !found || credential == "" || !member.connected {
		return ErrAuthentication
	}
	if endpoint.watches[key] {
		return ErrRateLimit
	}
	endpoint.watches[key] = true
	return nil
}

func (endpoint *Endpoint) EndWatch(credential SessionCredential) {
	endpoint.mu.Lock()
	delete(endpoint.watches, string(credential))
	endpoint.mu.Unlock()
}

func (endpoint *Endpoint) consume(credential SessionCredential, refresh bool) (connection, error) {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()
	member, found := endpoint.connections[string(credential)]
	if !found || credential == "" || !member.connected {
		return connection{}, ErrAuthentication
	}
	bucket := &member.commands
	if refresh {
		bucket = &member.refreshes
	}
	if !bucket.take(endpoint.now()) {
		return connection{}, ErrRateLimit
	}
	endpoint.connections[string(credential)] = member
	return member, nil
}

func newTokenBucket(capacity, rate float64, now time.Time) tokenBucket {
	return tokenBucket{tokens: capacity, capacity: capacity, rate: rate, updated: now}
}

func (bucket *tokenBucket) take(now time.Time) bool {
	if now.Before(bucket.updated) {
		now = bucket.updated
	}
	bucket.tokens = min(bucket.capacity, bucket.tokens+now.Sub(bucket.updated).Seconds()*bucket.rate)
	bucket.updated = now
	if bucket.tokens < 1 {
		return false
	}
	bucket.tokens--
	return true
}

// Leave revokes a connection credential. Character lease release and durable
// save commit remain realm responsibilities layered above this endpoint.
func (endpoint *Endpoint) Leave(credential SessionCredential) error {
	return endpoint.leaveCredential(credential, true)
}

// Disconnect suspends a membership for a short reconnect lease. A QUIC
// connection may disappear without sending its final unary request, and
// deleting the player immediately would make the reconnect protocol useless
// precisely when it is needed. Explicit Leave remains immediate.
func (endpoint *Endpoint) Disconnect(credential SessionCredential) {
	key := string(credential)
	endpoint.mu.Lock()
	member, found := endpoint.connections[key]
	if !found || credential == "" || !member.connected {
		endpoint.mu.Unlock()
		return
	}
	member.connected = false
	member.disconnectGeneration++
	generation := member.disconnectGeneration
	endpoint.connections[key] = member
	delete(endpoint.watches, key)
	endpoint.mu.Unlock()
	endpoint.after(endpoint.reconnectGrace, func() { endpoint.expireDisconnect(credential, generation) })
}

func (endpoint *Endpoint) expireDisconnect(credential SessionCredential, generation uint64) {
	key := string(credential)
	endpoint.mu.Lock()
	member, found := endpoint.connections[key]
	if !found || member.connected || member.disconnectGeneration != generation {
		endpoint.mu.Unlock()
		return
	}
	delete(endpoint.connections, key)
	delete(endpoint.watches, key)
	endpoint.mu.Unlock()
	if endpoint.leave != nil {
		_ = endpoint.leave(member.principal)
	}
}

func (endpoint *Endpoint) leaveCredential(credential SessionCredential, strict bool) error {
	endpoint.mu.Lock()
	member, found := endpoint.connections[string(credential)]
	if !found || credential == "" {
		endpoint.mu.Unlock()
		if strict {
			return ErrAuthentication
		}
		return nil
	}
	delete(endpoint.connections, string(credential))
	delete(endpoint.watches, string(credential))
	endpoint.mu.Unlock()
	if endpoint.leave != nil {
		return endpoint.leave(member.principal)
	}
	return nil
}

// Reconnect verifies that both the opaque connection and runtime identity are
// still valid, rotates the bearer credential, and returns a canonical
// correction snapshot. The old credential cannot be replayed after success.
func (endpoint *Endpoint) Reconnect(request ReconnectRequest) (JoinResponse, error) {
	if len(request.Nonce) < 32 || len(request.Nonce) > 128 {
		return JoinResponse{}, ErrAuthentication
	}
	if replay, found := endpoint.replayedReconnect(request); found {
		return replay, nil
	}
	member, err := endpoint.membership(request.Credential, true)
	if err != nil {
		return JoinResponse{}, err
	}
	if err := endpoint.host.ValidateReconnect(member.admission, request.Identity); err != nil {
		return JoinResponse{}, err
	}
	credential, err := newSessionCredential()
	if err != nil {
		return JoinResponse{}, fmt.Errorf("game server protocol: rotate credential: %w", err)
	}
	snapshot, err := endpoint.snapshot(member.principal.PlayerID)
	if err != nil {
		return JoinResponse{}, err
	}
	endpoint.mu.Lock()
	if current, found := endpoint.connections[string(request.Credential)]; !found || current.principal.ID != member.principal.ID {
		if replay, replayed := endpoint.reconnects[string(request.Credential)]; replayed && replay.nonce == request.Nonce {
			endpoint.mu.Unlock()
			return replay.response, nil
		}
		endpoint.mu.Unlock()
		return JoinResponse{}, ErrAuthentication
	} else {
		// Transfer the state observed while holding the lock. A concurrent
		// request may have consumed capacity since connection() returned.
		member = current
	}
	member.connected = true
	delete(endpoint.connections, string(request.Credential))
	delete(endpoint.watches, string(request.Credential))
	endpoint.connections[string(credential)] = member
	response := JoinResponse{Credential: credential, Admission: member.admission, Snapshot: snapshot}
	endpoint.reconnects[string(request.Credential)] = reconnectReplay{nonce: request.Nonce, response: response}
	endpoint.mu.Unlock()
	if endpoint.connected != nil {
		endpoint.connected(member.principal)
	}
	endpoint.after(endpoint.reconnectGrace, func() {
		endpoint.mu.Lock()
		if replay, found := endpoint.reconnects[string(request.Credential)]; found && replay.nonce == request.Nonce {
			delete(endpoint.reconnects, string(request.Credential))
		}
		endpoint.mu.Unlock()
	})
	return response, nil
}

func (endpoint *Endpoint) replayedReconnect(request ReconnectRequest) (JoinResponse, bool) {
	endpoint.mu.RLock()
	replay, found := endpoint.reconnects[string(request.Credential)]
	endpoint.mu.RUnlock()
	if !found || replay.nonce != request.Nonce {
		return JoinResponse{}, false
	}
	return replay.response, true
}

func (endpoint *Endpoint) connection(credential SessionCredential) (connection, error) {
	return endpoint.membership(credential, false)
}

func (endpoint *Endpoint) membership(credential SessionCredential, allowDisconnected bool) (connection, error) {
	endpoint.mu.RLock()
	member, found := endpoint.connections[string(credential)]
	endpoint.mu.RUnlock()
	if !found || credential == "" || (!allowDisconnected && !member.connected) {
		return connection{}, ErrAuthentication
	}
	return member, nil
}

func (endpoint *Endpoint) snapshot(playerID string) (Snapshot, error) {
	checkpoint, err := endpoint.canonicalCheckpoint()
	if err != nil {
		return Snapshot{}, err
	}
	payload, err := endpoint.project(playerID, checkpoint)
	if err != nil {
		return Snapshot{}, fmt.Errorf("game server protocol: project snapshot: %w", err)
	}
	if !json.Valid(payload) {
		return Snapshot{}, errors.New("game server protocol: projector returned invalid JSON")
	}
	return Snapshot{Version: SessionProtocolVersion, Tick: checkpoint.Tick, Checksum: checkpoint.Checksum,
		StepNanos: int64(endpoint.host.Session.StepDuration()), AcknowledgedInput: endpoint.host.Session.ProcessedSequence(playerID), Payload: append(json.RawMessage(nil), payload...)}, nil
}

// canonicalCheckpoint captures at most once per completed authoritative tick.
// Every watcher then projects from the same immutable checkpoint instead of
// independently snapshotting a live ECS at slightly different instants.
func (endpoint *Endpoint) canonicalCheckpoint() (simulation.Checkpoint, error) {
	endpoint.snapshotMu.Lock()
	defer endpoint.snapshotMu.Unlock()
	current := endpoint.host.Session.Status().Tick
	if endpoint.checkpoint.Snapshot != nil && endpoint.checkpoint.Tick == current {
		return cloneCheckpoint(endpoint.checkpoint), nil
	}
	checkpoint, err := endpoint.host.Session.CanonicalCheckpoint()
	if err != nil {
		return simulation.Checkpoint{}, err
	}
	endpoint.checkpoint = cloneCheckpoint(checkpoint)
	return cloneCheckpoint(checkpoint), nil
}

func cloneCheckpoint(checkpoint simulation.Checkpoint) simulation.Checkpoint {
	copy := checkpoint
	if checkpoint.Snapshot != nil {
		snapshot := *checkpoint.Snapshot
		snapshot.Entities = append([]uint64(nil), checkpoint.Snapshot.Entities...)
		snapshot.Components = append([]gameecs.ComponentSnapshot(nil), checkpoint.Snapshot.Components...)
		copy.Snapshot = &snapshot
	}
	copy.Participants = append([]simulation.ParticipantState(nil), checkpoint.Participants...)
	return copy
}

func newSessionCredential() (SessionCredential, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return SessionCredential(hex.EncodeToString(value[:])), nil
}
