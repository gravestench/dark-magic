package gameserver

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// SessionProtocolVersion identifies the wire contract understood by this endpoint implementation.
const SessionProtocolVersion uint32 = 2

var (
	// ErrAuthentication intentionally collapses credential and identity failures at the trust boundary.
	ErrAuthentication = errors.New("game server protocol: authentication failed")
	// ErrProtocol reports a client whose wire contract cannot be interpreted safely.
	ErrProtocol = errors.New("game server protocol: unsupported version")
	// ErrRateLimit reports a valid membership that has exhausted a bounded operation budget.
	ErrRateLimit = errors.New("game server protocol: rate limit exceeded")
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

// Authenticator converts an opaque admission credential into server-trusted identity.
type Authenticator interface {
	Authenticate(context.Context, string) (Principal, error)
}

// SnapshotProjector builds the public/private semantic view allowed for one
// player. It deliberately prevents the protocol from exposing raw ECS state.
type SnapshotProjector func(playerID string, checkpoint simulation.Checkpoint) (json.RawMessage, error)

// JoinRequest carries an admission credential and the runtime identity the client actually composed.
type JoinRequest struct {
	Version    uint32
	Credential string
	Identity   simulation.RuntimeIdentity
}

// SessionCredential is an opaque bearer secret bound to one admitted membership.
type SessionCredential string

// String exposes the credential only for transport serialization; callers must continue treating it as secret.
func (credential SessionCredential) String() string { return string(credential) }

// Snapshot is the versioned, per-player correction envelope returned by join and observation operations.
type Snapshot struct {
	Version           uint32          `json:"version"`
	Tick              uint64          `json:"tick"`
	StepNanos         int64           `json:"step_nanos"`
	Checksum          string          `json:"checksum"`
	AcknowledgedInput uint64          `json:"acknowledged_input"`
	Payload           json.RawMessage `json:"payload"`
}

// JoinResponse binds an admitted runtime token and first canonical snapshot to a fresh bearer credential.
type JoinResponse struct {
	Credential SessionCredential
	Admission  gamesession.AdmissionToken
	Snapshot   Snapshot
}

// ReconnectRequest proves knowledge of the old bearer credential and supplies an idempotency nonce.
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

// connection is the protocol-owned state transferred atomically when credentials rotate.
type connection struct {
	principal            Principal
	admission            gamesession.AdmissionToken
	commands             tokenBucket
	refreshes            tokenBucket
	connected            bool
	disconnectGeneration uint64
}

// reconnectReplay retains one idempotent rotation result for the old credential's grace window.
type reconnectReplay struct {
	nonce    string
	response JoinResponse
}

// tokenBucket stores per-membership capacity; copying a connection also copies its current rate-limit state.
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
