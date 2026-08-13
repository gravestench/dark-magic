package gameserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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
	commandBurst     = 32
	commandRate      = 16.0
	refreshBurst     = 4
	refreshRate      = 2.0
	joinReadyTimeout = 2 * time.Second
	joinReadyPoll    = 10 * time.Millisecond
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
	principal Principal
	admission gamesession.AdmissionToken
	commands  tokenBucket
	refreshes tokenBucket
}

type tokenBucket struct {
	tokens, capacity, rate float64
	updated                time.Time
}

// Endpoint is the transport-neutral authenticated boundary around one Host.
// HTTP, UDP, loopback, and legacy protocol adapters can all call this API.
type Endpoint struct {
	mu              sync.RWMutex
	host            *Host
	auth            Authenticator
	project         SnapshotProjector
	connections     map[string]connection
	now             func() time.Time
	snapshotPending func(error) bool
}

// SetSnapshotPending identifies the one expected projection error while a
// trusted next-tick admission command is waiting to materialize its player.
func (endpoint *Endpoint) SetSnapshotPending(classify func(error) bool) {
	endpoint.snapshotPending = classify
}

func NewEndpoint(host *Host, auth Authenticator, project SnapshotProjector) (*Endpoint, error) {
	if host == nil || host.Session == nil || auth == nil || project == nil {
		return nil, errors.New("game server protocol: host, authenticator, and projector are required")
	}
	return &Endpoint{host: host, auth: auth, project: project, connections: make(map[string]connection), now: time.Now}, nil
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
		commands: newTokenBucket(commandBurst, commandRate, now), refreshes: newTokenBucket(refreshBurst, refreshRate, now)}
	endpoint.mu.Unlock()
	snapshot, err := endpoint.joinSnapshot(ctx, principal.PlayerID)
	if err != nil {
		endpoint.mu.Lock()
		delete(endpoint.connections, string(credential))
		endpoint.mu.Unlock()
		return JoinResponse{}, err
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
	_, err = endpoint.host.Session.SubmitNetwork(simulation.Command{Tick: target, Player: member.principal.PlayerID,
		Authority: simulation.AuthorityPlayer, Sequence: intent.Sequence, Kind: intent.Kind, Payload: intent.Payload})
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

func (endpoint *Endpoint) consume(credential SessionCredential, refresh bool) (connection, error) {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()
	member, found := endpoint.connections[string(credential)]
	if !found || credential == "" {
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
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()
	if _, found := endpoint.connections[string(credential)]; !found || credential == "" {
		return ErrAuthentication
	}
	delete(endpoint.connections, string(credential))
	return nil
}

// Reconnect verifies that both the opaque connection and runtime identity are
// still valid, rotates the bearer credential, and returns a canonical
// correction snapshot. The old credential cannot be replayed after success.
func (endpoint *Endpoint) Reconnect(request ReconnectRequest) (JoinResponse, error) {
	member, err := endpoint.connection(request.Credential)
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
		endpoint.mu.Unlock()
		return JoinResponse{}, ErrAuthentication
	} else {
		// Transfer the state observed while holding the lock. A concurrent
		// request may have consumed capacity since connection() returned.
		member = current
	}
	delete(endpoint.connections, string(request.Credential))
	endpoint.connections[string(credential)] = member
	endpoint.mu.Unlock()
	return JoinResponse{Credential: credential, Admission: member.admission, Snapshot: snapshot}, nil
}

func (endpoint *Endpoint) connection(credential SessionCredential) (connection, error) {
	endpoint.mu.RLock()
	member, found := endpoint.connections[string(credential)]
	endpoint.mu.RUnlock()
	if !found || credential == "" {
		return connection{}, ErrAuthentication
	}
	return member, nil
}

func (endpoint *Endpoint) snapshot(playerID string) (Snapshot, error) {
	checkpoint, err := endpoint.host.Session.CanonicalCheckpoint()
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

func newSessionCredential() (SessionCredential, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return SessionCredential(hex.EncodeToString(value[:])), nil
}
