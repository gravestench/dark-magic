package realm

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

var ErrWorker = errors.New("realm: game worker operation failed")

// WorkerDescription is the immutable public control-plane description of one
// prepared authority. It contains no live session, ECS, or transport handle.
type WorkerDescription struct {
	GameID           string                     `json:"game_id,omitempty"`
	Runtime          simulation.RuntimeIdentity `json:"runtime"`
	IdentityHash     string                     `json:"identity_hash"`
	EntryDestination playeradapter.Destination  `json:"entry_destination"`
}

type WorkerStatus struct {
	Ready          bool     `json:"ready"`
	Tick           uint64   `json:"tick"`
	ActivePlayers  int      `json:"active_players"`
	ExpiredPlayers []string `json:"expired_players,omitempty"`
}

// WorkerMemberships is the shared process-local bridge between the public
// transport and private worker control adapters. QUIC marks a player only after
// reconnect grace expires; Realm control removes it only after canonical state
// has been projected and committed.
type WorkerMemberships struct {
	mu      sync.Mutex
	players map[string]workerMembership
	expired map[string]struct{}
	now     func() time.Time
}

type workerMembership struct {
	connected     bool
	claimDeadline time.Time
}

// NewWorkerMemberships constructs the worker boundary and validates dependencies before callers can publish or mutate
// shared state.
func NewWorkerMemberships() *WorkerMemberships {
	return &WorkerMemberships{
		players: make(map[string]workerMembership),
		expired: make(map[string]struct{}),
		now:     time.Now,
	}
}

// Admit coordinates admit through the owning worker synchronization boundary so shared state is published only after a
// complete transition.
func (memberships *WorkerMemberships) Admit(playerID string, claimDeadline time.Time) {
	if memberships == nil || strings.TrimSpace(playerID) == "" {
		return
	}

	if claimDeadline.IsZero() {
		claimDeadline = memberships.now().Add(45 * time.Second)
	}

	memberships.mu.Lock()
	memberships.players[playerID] = workerMembership{claimDeadline: claimDeadline}
	delete(memberships.expired, playerID)
	memberships.mu.Unlock()
}

// Connect coordinates connect through the owning worker synchronization boundary so shared state is published only
// after a complete transition.
func (memberships *WorkerMemberships) Connect(playerID string) {
	if memberships == nil || strings.TrimSpace(playerID) == "" {
		return
	}

	memberships.mu.Lock()
	if membership, active := memberships.players[playerID]; active {
		membership.connected = true
		memberships.players[playerID] = membership
		delete(memberships.expired, playerID)
	}
	memberships.mu.Unlock()
}

// Expire coordinates expire through the owning worker synchronization boundary so shared state is published only after
// a complete transition.
func (memberships *WorkerMemberships) Expire(playerID string) {
	if memberships == nil || strings.TrimSpace(playerID) == "" {
		return
	}

	memberships.mu.Lock()
	if _, active := memberships.players[playerID]; active {
		memberships.expired[playerID] = struct{}{}
	}
	memberships.mu.Unlock()
}

// Remove coordinates remove through the owning worker synchronization boundary so shared state is published only after
// a complete transition.
func (memberships *WorkerMemberships) Remove(playerID string) {
	if memberships == nil {
		return
	}

	memberships.mu.Lock()
	delete(memberships.players, playerID)
	delete(memberships.expired, playerID)
	memberships.mu.Unlock()
}

// Status coordinates status through the owning worker synchronization boundary so shared state is published only after
// a complete transition.
func (memberships *WorkerMemberships) Status() (int, []string) {
	if memberships == nil {
		return 0, nil
	}

	memberships.mu.Lock()
	defer memberships.mu.Unlock()

	now := memberships.now()
	for playerID, membership := range memberships.players {
		if !membership.connected && !membership.claimDeadline.After(now) {
			memberships.expired[playerID] = struct{}{}
		}
	}

	expired := make([]string, 0, len(memberships.expired))
	for playerID := range memberships.expired {
		expired = append(expired, playerID)
	}

	sort.Strings(expired)

	return len(memberships.players), expired
}

// WorkerAdmission is the complete trusted character-entry request sent from
// the Realm to a prepared worker. The worker validates durable compatibility
// and chooses the next authoritative tick; clients cannot construct it.
type WorkerAdmission struct {
	Character     d2save.Character                 `json:"character"`
	Compatibility gamesession.DurableCompatibility `json:"compatibility"`
	PlayerID      string                           `json:"player_id"`
	Destination   playeradapter.Destination        `json:"destination"`
	Actor         string                           `json:"actor"`
	Sequence      uint64                           `json:"sequence"`
	ClaimDeadline time.Time                        `json:"claim_deadline"`
}

// WorkerClient is the process-independent Realm control boundary for one game
// authority. An in-process adapter implements it today; a child process and a
// future cluster allocator must preserve these semantics without exposing a
// gameserver.Host to Realm business logic.
type WorkerClient interface {
	Describe(context.Context) (WorkerDescription, error)
	Status(context.Context) (WorkerStatus, error)
	Checkpoint(context.Context) (gamesession.RecoveryCheckpoint, error)
	AdmitCharacter(context.Context, WorkerAdmission) error
	RemoveCharacter(context.Context, string) error
	ProjectCharacter(context.Context, string, d2save.Character) (d2save.Character, error)
	Close(context.Context) error
}

// WorkerRegistry resolves a durable GameID to its allocated control client.
// Admissions depends on this narrow contract rather than the local Manager.
type WorkerRegistry interface {
	Game(string) (WorkerClient, bool)
}

// GameSpec is the topology-neutral request for one authoritative game. Runtime
// and asset policy are pinned by the allocator configuration and revalidated
// through WorkerDescription before the game becomes joinable.
type GameSpec struct {
	GameID         string
	AllocationID   string
	Difficulty     GameDifficulty
	Hardcore       bool
	Ladder         bool
	MaximumPlayers int
}

// WorkerAllocation is the complete private result of preparing one game. It is
// consumed by Realm orchestration and is never exposed through the directory.
type WorkerAllocation struct {
	GameID       string
	AllocationID string
	Worker       WorkerClient
	Tickets      TicketIssuer
	Endpoint     GameEndpoint
}

// GameAllocator is implemented by the supervised local-process allocator and,
// later, by a cluster allocator without changing Realm business logic.
type GameAllocator interface {
	WorkerRegistry
	Allocate(context.Context, GameSpec) (WorkerAllocation, error)
	Release(context.Context, string) error
}

// GameRestorer is the optional allocation capability for starting a replacement
// authority from a validated durable recovery checkpoint.
type GameRestorer interface {
	Restore(context.Context, GameSpec, GameRecovery) (WorkerAllocation, error)
}

// GameFencer proves that a surviving authority for an interrupted allocation
// has stopped accepting game traffic before a replacement is started. A
// durable startup recovery path must never restore without this proof.
type GameFencer interface {
	Fence(context.Context, GameSpec) error
}

const GameRecoveryVersion = "RealmGameRecovery/v1"

// GameRecovery is the complete trusted replacement-worker handoff. Player IDs
// are Realm-owned membership identities, not client claims. They seed the new
// worker's reconnect tracker while the session checkpoint restores gameplay.
type GameRecovery struct {
	Version    string                         `json:"version"`
	Checkpoint gamesession.RecoveryCheckpoint `json:"checkpoint"`
	PlayerIDs  []string                       `json:"player_ids,omitempty"`
}

// NewGameRecovery constructs the worker boundary and validates dependencies before callers can publish or mutate
// shared state.
func NewGameRecovery(checkpoint gamesession.RecoveryCheckpoint, playerIDs []string) (GameRecovery, error) {
	recovery := GameRecovery{Version: GameRecoveryVersion, Checkpoint: checkpoint,
		PlayerIDs: append([]string(nil), playerIDs...)}
	if err := ValidateGameRecovery(recovery); err != nil {
		return GameRecovery{}, err
	}

	sort.Strings(recovery.PlayerIDs)

	return recovery, nil
}

// ValidateGameRecovery checks the worker invariant before state changes, keeping invalid values off shared paths.
func ValidateGameRecovery(recovery GameRecovery) error {
	if recovery.Version != GameRecoveryVersion || gamesession.ValidateRecoveryCheckpoint(recovery.Checkpoint) != nil ||
		len(recovery.PlayerIDs) > 8 {
		return ErrWorker
	}

	seen := make(map[string]struct{}, len(recovery.PlayerIDs))
	for _, playerID := range recovery.PlayerIDs {
		playerID = strings.TrimSpace(playerID)
		if playerID == "" || len(playerID) > 255 {
			return ErrWorker
		}

		if _, exists := seen[playerID]; exists {
			return ErrWorker
		}

		seen[playerID] = struct{}{}
	}

	return nil
}

// AdmissionPrincipal is the Realm-owned identity bound into a one-use worker
// ticket. It deliberately excludes credentials and character payloads.
type AdmissionPrincipal struct {
	AccountID           string `json:"account_id"`
	CharacterID         string `json:"character_id"`
	PlayerID            string `json:"player_id"`
	CharacterRevision   uint64 `json:"character_revision"`
	RuntimeIdentityHash string `json:"runtime_identity_hash"`
}

// TicketIssuer abstracts the authority that creates and revokes one-use game
// admission tickets. A remote implementation can cross the private worker
// protocol without changing Admissions.
type TicketIssuer interface {
	Issue(context.Context, AdmissionPrincipal, time.Duration) (string, error)
	Revoke(context.Context, string) error
}

type inProcessWorker struct {
	host        *gameserver.Host
	departures  playeradapter.DepartureQueue
	destination playeradapter.Destination
	memberships *WorkerMemberships
}

// newInProcessWorker constructs the worker boundary and validates dependencies before callers can publish or mutate
// shared state.
func newInProcessWorker(host *gameserver.Host) (WorkerClient, error) {
	return NewInProcessWorker(host)
}

// NewInProcessWorker contains the concrete host inside the local adapter.
// Server composition uses it to expose the same semantic boundary remotely.
func NewInProcessWorker(host *gameserver.Host) (WorkerClient, error) {
	return NewInProcessWorkerWithDestination(host, playeradapter.Destination{})
}

// NewInProcessWorkerWithDestination publishes the exact spawn prepared by the
// worker. Realm may provide a fallback only for legacy in-process fixtures;
// production admissions always use this worker-owned geometry.
func NewInProcessWorkerWithDestination(
	host *gameserver.Host,
	destination playeradapter.Destination,
) (WorkerClient, error) {
	return NewInProcessWorkerWithMemberships(host, destination, NewWorkerMemberships())
}

// NewInProcessWorkerWithMemberships binds worker control to the same transport
// membership tracker used by a realm-worker QUIC endpoint.
func NewInProcessWorkerWithMemberships(
	host *gameserver.Host,
	destination playeradapter.Destination,
	memberships *WorkerMemberships,
) (WorkerClient, error) {
	if host == nil || host.Session == nil || strings.TrimSpace(host.Allocation.SessionID) == "" ||
		strings.TrimSpace(host.Allocation.IdentityHash) == "" {
		return nil, ErrWorker
	}

	if memberships == nil {
		return nil, ErrWorker
	}

	return &inProcessWorker{host: host, destination: destination, memberships: memberships}, nil
}

// Status returns the worker observation through its owning boundary so callers receive a consistent snapshot and error
// contract.
func (worker *inProcessWorker) Status(ctx context.Context) (WorkerStatus, error) {
	if err := worker.ready(ctx); err != nil {
		return WorkerStatus{}, err
	}

	checkpoint, err := worker.host.Session.CanonicalCheckpoint()
	if err != nil {
		return WorkerStatus{}, err
	}

	active, expired := worker.memberships.Status()

	return WorkerStatus{Ready: true, Tick: checkpoint.Tick, ActivePlayers: active, ExpiredPlayers: expired}, nil
}

// Checkpoint contains checkpoint within the worker boundary so callers do not duplicate its domain-specific policy.
func (worker *inProcessWorker) Checkpoint(ctx context.Context) (gamesession.RecoveryCheckpoint, error) {
	if err := worker.ready(ctx); err != nil {
		return gamesession.RecoveryCheckpoint{}, err
	}

	return worker.host.Session.RecoveryCheckpoint()
}

// Describe contains describe within the worker boundary so callers do not duplicate its domain-specific policy.
func (worker *inProcessWorker) Describe(ctx context.Context) (WorkerDescription, error) {
	if err := worker.ready(ctx); err != nil {
		return WorkerDescription{}, err
	}

	return WorkerDescription{
		GameID:           worker.host.Allocation.SessionID,
		Runtime:          cloneRuntimeIdentity(worker.host.Allocation.Identity),
		IdentityHash:     worker.host.Allocation.IdentityHash,
		EntryDestination: worker.destination,
	}, nil
}

// AdmitCharacter contains admit character within the worker boundary so callers do not duplicate its domain-specific
// policy.
func (worker *inProcessWorker) AdmitCharacter(ctx context.Context, request WorkerAdmission) error {
	if err := worker.ready(ctx); err != nil {
		return err
	}

	if strings.TrimSpace(request.PlayerID) == "" || strings.TrimSpace(request.Actor) == "" || request.Sequence == 0 {
		return ErrWorker
	}

	if err := worker.host.Allocation.ValidateDurable(request.Compatibility); err != nil {
		return err
	}

	checkpoint, err := worker.host.Session.CanonicalCheckpoint()
	if err != nil {
		return err
	}

	command, err := playeradapter.AdmissionCommand(request.Character, request.PlayerID, request.Destination,
		request.Actor, request.Sequence, checkpoint.Tick+1, simulation.AuthoritySystem)
	if err != nil {
		return err
	}

	if err := worker.host.Session.Submit(command); err != nil {
		return err
	}

	worker.memberships.Admit(request.PlayerID, request.ClaimDeadline)

	return nil
}

// RemoveCharacter owns the worker cleanup transition so resources and durable state are retired in the required order.
func (worker *inProcessWorker) RemoveCharacter(ctx context.Context, playerID string) error {
	if err := worker.ready(ctx); err != nil {
		return err
	}

	if err := worker.departures.Submit(worker.host.Session, playerID); err != nil {
		return err
	}

	worker.memberships.Remove(playerID)

	return nil
}

// ProjectCharacter contains project character within the worker boundary so callers do not duplicate its
// domain-specific policy.
func (worker *inProcessWorker) ProjectCharacter(
	ctx context.Context,
	playerID string,
	baseline d2save.Character,
) (d2save.Character, error) {
	if err := worker.ready(ctx); err != nil {
		return d2save.Character{}, err
	}

	checkpoint, err := worker.host.Session.CanonicalCheckpoint()
	if err != nil {
		return d2save.Character{}, err
	}

	return playeradapter.ProjectCharacter(playerID, baseline, checkpoint)
}

// Close owns the worker cleanup transition so resources and durable state are retired in the required order.
func (worker *inProcessWorker) Close(ctx context.Context) error {
	if worker == nil || worker.host == nil {
		return nil
	}

	return worker.host.Close(ctx)
}

// ready decodes the worker representation at one boundary so malformed data fails before it becomes shared state.
func (worker *inProcessWorker) ready(ctx context.Context) error {
	if worker == nil || worker.host == nil || worker.host.Session == nil {
		return ErrWorker
	}

	if ctx == nil {
		return ErrWorker
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// localTicketIssuer adapts the current in-process ticket implementation to the
// same context-aware semantic boundary a remote worker issuer will implement.
type localTicketIssuer struct {
	authority *gameserver.TicketAuthority
}

// newLocalTicketIssuer constructs the worker boundary and validates dependencies before callers can publish or mutate
// shared state.
func newLocalTicketIssuer(authority *gameserver.TicketAuthority) (TicketIssuer, error) {
	return NewLocalTicketIssuer(authority)
}

// NewLocalTicketIssuer contains the concrete ticket authority inside the
// local adapter used by the worker service.
func NewLocalTicketIssuer(authority *gameserver.TicketAuthority) (TicketIssuer, error) {
	if authority == nil {
		return nil, ErrWorker
	}

	return &localTicketIssuer{authority: authority}, nil
}

// Issue contains issue within the worker boundary so callers do not duplicate its domain-specific policy.
func (issuer *localTicketIssuer) Issue(
	ctx context.Context,
	principal AdmissionPrincipal,
	lifetime time.Duration,
) (string, error) {
	if issuer == nil || issuer.authority == nil || ctx == nil {
		return "", ErrWorker
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	return issuer.authority.Issue(gameserver.Principal{ID: principal.AccountID, CharacterID: principal.CharacterID,
		PlayerID: principal.PlayerID, CharacterRevision: principal.CharacterRevision,
		RuntimeIdentityHash: principal.RuntimeIdentityHash}, lifetime)
}

// Revoke contains revoke within the worker boundary so callers do not duplicate its domain-specific policy.
func (issuer *localTicketIssuer) Revoke(ctx context.Context, ticket string) error {
	if issuer == nil || issuer.authority == nil || ctx == nil {
		return ErrWorker
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return issuer.authority.Revoke(ticket)
}
