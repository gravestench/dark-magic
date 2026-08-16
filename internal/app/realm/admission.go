package realm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

var ErrAdmission = errors.New("realm: game admission failed")

type GameEndpoint struct {
	Address        string `json:"address"`
	TLSFingerprint string `json:"tls_fingerprint"`
}

type JoinRequest struct {
	AccountID, CharacterID, PlayerID, GameID string
	Destination                              playeradapter.Destination
}

type JoinAssignment struct {
	GameID            string
	Endpoint          GameEndpoint
	Ticket            string
	CharacterRevision uint64
	Runtime           simulation.RuntimeIdentity
}

type admissionGame struct {
	tickets  TicketIssuer
	endpoint GameEndpoint
}

type characterMembership struct {
	lease    CharacterLease
	baseline CharacterRecord
	ticket   string
}

// Admissions coordinates realm-owned leases and trusted character entry with
// a worker-owned ticket authority. It never executes gameplay rules itself.
type Admissions struct {
	mu              sync.RWMutex
	entryMu         sync.Mutex
	workers         WorkerRegistry
	characters      CharacterRepository
	membershipStore MembershipRepository
	leaseLifetime   time.Duration
	ticketLifetime  time.Duration
	now             func() time.Time
	games           map[string]admissionGame
	memberships     map[string]characterMembership
	entrySequences  map[string]uint64
}

func NewAdmissions(workers WorkerRegistry, characters CharacterRepository, leaseLifetime, ticketLifetime time.Duration) (*Admissions, error) {
	memberships, err := NewMemoryMemberships(characters)
	if err != nil {
		return nil, err
	}
	return NewAdmissionsWithMemberships(workers, characters, memberships, leaseLifetime, ticketLifetime)
}

func NewAdmissionsWithMemberships(workers WorkerRegistry, characters CharacterRepository, memberships MembershipRepository, leaseLifetime, ticketLifetime time.Duration) (*Admissions, error) {
	if workers == nil || characters == nil || memberships == nil || leaseLifetime <= 0 || ticketLifetime <= 0 || ticketLifetime > leaseLifetime {
		return nil, errors.New("realm: admissions require workers, characters, and bounded lifetimes")
	}
	return &Admissions{workers: workers, characters: characters, membershipStore: memberships, leaseLifetime: leaseLifetime, ticketLifetime: ticketLifetime,
		now: time.Now, games: make(map[string]admissionGame), memberships: make(map[string]characterMembership),
		entrySequences: make(map[string]uint64)}, nil
}

func (admissions *Admissions) RegisterGame(gameID string, tickets TicketIssuer, endpoint GameEndpoint) error {
	gameID = strings.TrimSpace(gameID)
	if gameID == "" || tickets == nil || strings.TrimSpace(endpoint.Address) == "" || strings.TrimSpace(endpoint.TLSFingerprint) == "" {
		return ErrAdmission
	}
	if _, found := admissions.workers.Game(gameID); !found {
		return ErrGameNotFound
	}
	admissions.mu.Lock()
	defer admissions.mu.Unlock()
	if _, exists := admissions.games[gameID]; exists {
		return ErrGameExists
	}
	admissions.games[gameID] = admissionGame{tickets: tickets, endpoint: endpoint}
	return nil
}

func (admissions *Admissions) ReplaceGame(gameID string, tickets TicketIssuer, endpoint GameEndpoint) error {
	gameID = strings.TrimSpace(gameID)
	if admissions == nil || gameID == "" || tickets == nil || strings.TrimSpace(endpoint.Address) == "" ||
		strings.TrimSpace(endpoint.TLSFingerprint) == "" {
		return ErrAdmission
	}
	if _, found := admissions.workers.Game(gameID); !found {
		return ErrGameNotFound
	}
	admissions.mu.Lock()
	defer admissions.mu.Unlock()
	if _, exists := admissions.games[gameID]; !exists {
		return ErrGameNotFound
	}
	admissions.games[gameID] = admissionGame{tickets: tickets, endpoint: endpoint}
	return nil
}

// ResumeGame rebuilds Admissions' process-local index after durable authority
// recovery. It must run before public traffic begins. The repository rotates
// durable lease secrets first; only the newly returned raw tokens enter memory.
func (admissions *Admissions) ResumeGame(ctx context.Context, gameID string, tickets TicketIssuer, endpoint GameEndpoint) ([]MembershipRecord, error) {
	gameID = strings.TrimSpace(gameID)
	if admissions == nil || ctx == nil || gameID == "" || tickets == nil ||
		strings.TrimSpace(endpoint.Address) == "" || strings.TrimSpace(endpoint.TLSFingerprint) == "" {
		return nil, ErrAdmission
	}
	if _, found := admissions.workers.Game(gameID); !found {
		return nil, ErrGameNotFound
	}
	admissions.entryMu.Lock()
	defer admissions.entryMu.Unlock()
	admissions.mu.RLock()
	_, gameExists := admissions.games[gameID]
	admissions.mu.RUnlock()
	if gameExists {
		return nil, ErrGameExists
	}
	records, err := admissions.membershipStore.ResumeGame(ctx, gameID, admissions.leaseLifetime)
	if err != nil {
		return nil, err
	}
	indexed := make(map[string]characterMembership, len(records))
	for _, record := range records {
		if record.GameID != gameID || record.State != MembershipActive || validateActiveMembership(record) != nil {
			return nil, ErrMembership
		}
		membershipID := membershipKey(gameID, record.PlayerID)
		if _, exists := indexed[membershipID]; exists {
			return nil, ErrMembership
		}
		indexed[membershipID] = characterMembership{lease: record.Lease, baseline: record.Baseline}
	}
	admissions.mu.Lock()
	defer admissions.mu.Unlock()
	if _, exists := admissions.games[gameID]; exists {
		return nil, ErrGameExists
	}
	for membershipID := range indexed {
		if _, exists := admissions.memberships[membershipID]; exists {
			return nil, ErrMembership
		}
	}
	admissions.games[gameID] = admissionGame{tickets: tickets, endpoint: endpoint}
	for membershipID, membership := range indexed {
		admissions.memberships[membershipID] = membership
	}
	result := make([]MembershipRecord, len(records))
	for index, record := range records {
		result[index] = cloneMembershipRecord(record)
	}
	return result, nil
}

// ReconnectAssignment issues a fresh one-use join ticket for an already
// admitted durable membership. This is used when a replacement worker has the
// canonical player entity but the old transport credential and endpoint are no
// longer valid.
func (admissions *Admissions) ReconnectAssignment(ctx context.Context, gameID, accountID string) (JoinAssignment, error) {
	gameID, accountID = strings.TrimSpace(gameID), strings.TrimSpace(accountID)
	if admissions == nil || ctx == nil || gameID == "" || accountID == "" {
		return JoinAssignment{}, ErrAdmission
	}
	admissions.mu.RLock()
	game, configured := admissions.games[gameID]
	var playerID string
	var membership characterMembership
	for membershipID, candidate := range admissions.memberships {
		if candidate.lease.GameID == gameID && candidate.baseline.AccountID == accountID {
			playerID = strings.TrimPrefix(membershipID, gameID+"\x00")
			membership = candidate
			break
		}
	}
	admissions.mu.RUnlock()
	if !configured || playerID == "" {
		return JoinAssignment{}, ErrLease
	}
	worker, found := admissions.workers.Game(gameID)
	if !found {
		return JoinAssignment{}, ErrGameNotFound
	}
	description, err := worker.Describe(ctx)
	if err != nil {
		return JoinAssignment{}, err
	}
	identityHash, err := description.Runtime.Digest()
	if err != nil || identityHash != description.IdentityHash || membership.baseline.Compatibility.IdentityHash != identityHash {
		return JoinAssignment{}, errors.Join(ErrWorker, err)
	}
	principal := AdmissionPrincipal{AccountID: accountID, CharacterID: membership.baseline.Character.ID,
		PlayerID: playerID, CharacterRevision: membership.baseline.Revision, RuntimeIdentityHash: identityHash}
	ticket, err := game.tickets.Issue(ctx, principal, admissions.ticketLifetime)
	if err != nil {
		return JoinAssignment{}, err
	}
	return JoinAssignment{GameID: gameID, Endpoint: game.endpoint, Ticket: ticket,
		CharacterRevision: membership.baseline.Revision, Runtime: cloneRuntimeIdentity(description.Runtime)}, nil
}

func (admissions *Admissions) UnregisterGame(gameID string) error {
	gameID = strings.TrimSpace(gameID)
	if admissions == nil || gameID == "" {
		return ErrAdmission
	}
	admissions.mu.Lock()
	defer admissions.mu.Unlock()
	if _, found := admissions.games[gameID]; !found {
		return ErrGameNotFound
	}
	for membershipID := range admissions.memberships {
		if strings.HasPrefix(membershipID, gameID+"\x00") {
			return ErrCharacterLeased
		}
	}
	delete(admissions.games, gameID)
	return nil
}

func (admissions *Admissions) Join(ctx context.Context, request JoinRequest) (JoinAssignment, error) {
	if ctx == nil || strings.TrimSpace(request.AccountID) == "" || strings.TrimSpace(request.CharacterID) == "" || strings.TrimSpace(request.PlayerID) == "" || strings.TrimSpace(request.GameID) == "" {
		return JoinAssignment{}, ErrAdmission
	}
	worker, found := admissions.workers.Game(request.GameID)
	if !found {
		return JoinAssignment{}, ErrGameNotFound
	}
	admissions.mu.RLock()
	game, configured := admissions.games[request.GameID]
	admissions.mu.RUnlock()
	if !configured {
		return JoinAssignment{}, ErrAdmission
	}
	membershipID := request.GameID + "\x00" + request.PlayerID
	admissions.mu.RLock()
	_, alreadyJoined := admissions.memberships[membershipID]
	admissions.mu.RUnlock()
	if alreadyJoined {
		return JoinAssignment{}, ErrCharacterLeased
	}
	record, lease, err := admissions.characters.Acquire(ctx, request.AccountID, request.CharacterID, request.GameID, admissions.leaseLifetime)
	if err != nil {
		return JoinAssignment{}, err
	}
	rollback := func(cause error, ticket string) (JoinAssignment, error) {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if ticket != "" {
			cause = errors.Join(cause, game.tickets.Revoke(cleanupCtx, ticket))
		}
		cause = errors.Join(cause, admissions.characters.Release(cleanupCtx, lease))
		return JoinAssignment{}, fmt.Errorf("%w: %v", ErrAdmission, cause)
	}
	description, err := worker.Describe(ctx)
	if err != nil {
		return rollback(err, "")
	}
	identityHash, err := description.Runtime.Digest()
	if err != nil || identityHash != description.IdentityHash {
		return rollback(errors.Join(ErrWorker, err), "")
	}
	expectedCompatibility := gamesession.DurableCompatibility{CharacterID: record.Character.ID,
		ModID: description.Runtime.Recipe.Packages.Base.ID, ContractVersion: description.Runtime.Recipe.EngineAPI,
		IdentityHash: description.IdentityHash}
	if emptyCompatibility(record.Compatibility) {
		record, err = admissions.characters.BindCompatibility(ctx, lease, expectedCompatibility)
		if err != nil {
			return rollback(err, "")
		}
	} else if record.Compatibility != expectedCompatibility {
		return rollback(gamesession.ErrCompatibility, "")
	}
	destination := request.Destination
	if _, err := playeradapter.NewDestination(description.EntryDestination.X, description.EntryDestination.Y,
		description.EntryDestination.Width, description.EntryDestination.Height, description.EntryDestination.Act,
		description.EntryDestination.LevelID); err == nil {
		destination = description.EntryDestination
	}
	principal := AdmissionPrincipal{AccountID: request.AccountID, CharacterID: request.CharacterID, PlayerID: request.PlayerID,
		CharacterRevision: record.Revision, RuntimeIdentityHash: description.IdentityHash}
	ticket, err := game.tickets.Issue(ctx, principal, admissions.ticketLifetime)
	if err != nil {
		return rollback(err, "")
	}
	admissions.entryMu.Lock()
	defer admissions.entryMu.Unlock()
	actor := "realm:entry:" + request.PlayerID
	sequence := admissions.entrySequences[request.GameID+"\x00"+actor] + 1
	if err := worker.AdmitCharacter(ctx, WorkerAdmission{Character: record.Character, Compatibility: record.Compatibility,
		PlayerID: request.PlayerID, Destination: destination, Actor: actor, Sequence: sequence,
		ClaimDeadline: admissions.now().Add(admissions.ticketLifetime)}); err != nil {
		return rollback(err, ticket)
	}
	membership := characterMembership{lease: lease, baseline: record, ticket: ticket}
	if err := admissions.membershipStore.Admit(ctx, MembershipRecord{GameID: request.GameID, PlayerID: request.PlayerID,
		AccountID: request.AccountID, Baseline: record, Lease: lease, State: MembershipActive}); err != nil {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		removeErr := worker.RemoveCharacter(cleanupCtx, request.PlayerID)
		cancel()
		return rollback(errors.Join(err, removeErr), ticket)
	}
	admissions.entrySequences[request.GameID+"\x00"+actor] = sequence
	admissions.mu.Lock()
	admissions.memberships[membershipID] = membership
	admissions.mu.Unlock()
	return JoinAssignment{GameID: request.GameID, Endpoint: game.endpoint, Ticket: ticket,
		CharacterRevision: record.Revision, Runtime: cloneRuntimeIdentity(description.Runtime)}, nil
}

// CancelMembership rolls back an admission that cannot be handed to a client.
// It revokes the one-use ticket, removes the queued player entity, and releases
// the Realm-owned character lease under an independent bounded context.
func (admissions *Admissions) CancelMembership(ctx context.Context, gameID, playerID string) error {
	if admissions == nil || ctx == nil {
		return ErrAdmission
	}
	gameID, playerID = strings.TrimSpace(gameID), strings.TrimSpace(playerID)
	membershipID := gameID + "\x00" + playerID
	admissions.mu.RLock()
	membership, found := admissions.memberships[membershipID]
	game, configured := admissions.games[gameID]
	admissions.mu.RUnlock()
	if !found || !configured {
		return ErrLease
	}
	worker, found := admissions.workers.Game(gameID)
	if !found {
		return ErrGameNotFound
	}
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	var revokeErr error
	if membership.ticket != "" {
		revokeErr = game.tickets.Revoke(cleanupCtx, membership.ticket)
	}
	if err := worker.RemoveCharacter(cleanupCtx, playerID); err != nil {
		return errors.Join(revokeErr, err)
	}
	admissions.mu.Lock()
	if current, exists := admissions.memberships[membershipID]; exists && current.lease.Token == membership.lease.Token {
		delete(admissions.memberships, membershipID)
	}
	admissions.mu.Unlock()
	return errors.Join(revokeErr, admissions.characters.Release(cleanupCtx, membership.lease),
		admissions.membershipStore.Cancel(cleanupCtx, gameID, playerID))
}

func (admissions *Admissions) RenewMembership(ctx context.Context, gameID, playerID string) (CharacterLease, error) {
	membershipID := strings.TrimSpace(gameID) + "\x00" + strings.TrimSpace(playerID)
	admissions.mu.RLock()
	membership, found := admissions.memberships[membershipID]
	admissions.mu.RUnlock()
	if !found {
		return CharacterLease{}, ErrLease
	}
	renewed, err := admissions.characters.Renew(ctx, membership.lease, admissions.leaseLifetime)
	if err != nil {
		return CharacterLease{}, err
	}
	admissions.mu.Lock()
	if current, exists := admissions.memberships[membershipID]; !exists || current.lease.Token != membership.lease.Token {
		admissions.mu.Unlock()
		return CharacterLease{}, ErrLease
	}
	membership.lease = renewed
	admissions.memberships[membershipID] = membership
	admissions.mu.Unlock()
	return renewed, nil
}

// RenewGameMemberships keeps active Realm characters exclusively leased while
// their worker is healthy. Renewal is serialized with canonical leave and only
// writes after half the lease lifetime has elapsed.
func (admissions *Admissions) RenewGameMemberships(ctx context.Context, gameID string) (int, error) {
	if admissions == nil || ctx == nil {
		return 0, ErrAdmission
	}
	admissions.entryMu.Lock()
	defer admissions.entryMu.Unlock()
	gameID = strings.TrimSpace(gameID)
	now := admissions.now()
	admissions.mu.RLock()
	memberships := make(map[string]characterMembership)
	for membershipID, membership := range admissions.memberships {
		if membership.lease.GameID == gameID && membership.lease.ExpiresAt.Sub(now) <= admissions.leaseLifetime/2 {
			memberships[membershipID] = membership
		}
	}
	admissions.mu.RUnlock()
	renewedCount := 0
	var result error
	for membershipID, membership := range memberships {
		renewed, err := admissions.characters.Renew(ctx, membership.lease, admissions.leaseLifetime)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		admissions.mu.Lock()
		current, exists := admissions.memberships[membershipID]
		if exists && current.lease.Token == membership.lease.Token {
			current.lease = renewed
			admissions.memberships[membershipID] = current
			renewedCount++
		}
		admissions.mu.Unlock()
	}
	return renewedCount, result
}

// CommitMembership is the trusted worker-to-realm persistence boundary. The
// opaque lease stays inside Admissions and is never included in JoinAssignment.
func (admissions *Admissions) CommitMembership(ctx context.Context, gameID, playerID string, character d2save.Character) (CharacterRecord, error) {
	membershipID := strings.TrimSpace(gameID) + "\x00" + strings.TrimSpace(playerID)
	admissions.mu.RLock()
	membership, found := admissions.memberships[membershipID]
	admissions.mu.RUnlock()
	if !found {
		return CharacterRecord{}, ErrLease
	}
	receipt, err := admissions.membershipStore.Depart(ctx, MembershipRecord{GameID: gameID, PlayerID: playerID,
		AccountID: membership.baseline.AccountID, Baseline: membership.baseline, Lease: membership.lease,
		State: MembershipActive}, character)
	if err != nil {
		return CharacterRecord{}, err
	}
	admissions.mu.Lock()
	if current, exists := admissions.memberships[membershipID]; !exists || current.lease.Token != membership.lease.Token {
		admissions.mu.Unlock()
		return CharacterRecord{}, ErrLease
	}
	delete(admissions.memberships, membershipID)
	admissions.mu.Unlock()
	return receipt.Record, nil
}

// CommitCanonicalMembership projects the worker's canonical checkpoint into
// the leased durable baseline, then commits it through the realm-only lease.
func (admissions *Admissions) CommitCanonicalMembership(ctx context.Context, gameID, playerID string) (CharacterRecord, error) {
	gameID, playerID = strings.TrimSpace(gameID), strings.TrimSpace(playerID)
	membershipID := gameID + "\x00" + playerID
	admissions.mu.RLock()
	membership, found := admissions.memberships[membershipID]
	admissions.mu.RUnlock()
	if !found {
		return CharacterRecord{}, ErrLease
	}
	worker, found := admissions.workers.Game(gameID)
	if !found {
		return CharacterRecord{}, ErrGameNotFound
	}
	character, err := worker.ProjectCharacter(ctx, playerID, membership.baseline.Character)
	if err != nil {
		return CharacterRecord{}, err
	}
	return admissions.CommitMembership(ctx, gameID, playerID, character)
}

// AccountMembership resolves only Realm-owned membership state. The account
// identity comes from an authenticated Realm session, never a client-supplied
// player ID.
func (admissions *Admissions) AccountMembership(gameID, accountID string) (string, CharacterRecord, error) {
	if admissions == nil {
		return "", CharacterRecord{}, ErrLease
	}
	gameID, accountID = strings.TrimSpace(gameID), strings.TrimSpace(accountID)
	admissions.mu.RLock()
	defer admissions.mu.RUnlock()
	for membershipID, membership := range admissions.memberships {
		if membership.lease.GameID != gameID || membership.baseline.AccountID != accountID {
			continue
		}
		_, playerID, found := strings.Cut(membershipID, "\x00")
		if !found || playerID == "" {
			return "", CharacterRecord{}, ErrLease
		}
		return playerID, cloneCharacterRecord(membership.baseline), nil
	}
	return "", CharacterRecord{}, ErrLease
}

// PlayerMembership resolves private Realm ownership for a trusted worker
// lifecycle notification. Player IDs come from authenticated worker status and
// are never accepted from a public client request.
func (admissions *Admissions) PlayerMembership(gameID, playerID string) (string, CharacterRecord, error) {
	if admissions == nil {
		return "", CharacterRecord{}, ErrAdmission
	}
	gameID, playerID = strings.TrimSpace(gameID), strings.TrimSpace(playerID)
	if gameID == "" || playerID == "" {
		return "", CharacterRecord{}, ErrAdmission
	}
	admissions.mu.RLock()
	membership, found := admissions.memberships[gameID+"\x00"+playerID]
	admissions.mu.RUnlock()
	if !found {
		return "", CharacterRecord{}, ErrLease
	}
	return membership.baseline.AccountID, cloneCharacterRecord(membership.baseline), nil
}

// LeaveCanonicalMembership serializes the final projection/commit/removal
// sequence. Once Commit succeeds, the durable character and lease are safe
// even if worker removal reports an error; callers must still remove the public
// roster and may drain an empty or unhealthy allocation.
func (admissions *Admissions) LeaveCanonicalMembership(ctx context.Context, gameID, playerID string) (CharacterRecord, error) {
	if admissions == nil || ctx == nil {
		return CharacterRecord{}, ErrAdmission
	}
	admissions.entryMu.Lock()
	defer admissions.entryMu.Unlock()
	gameID, playerID = strings.TrimSpace(gameID), strings.TrimSpace(playerID)
	membershipID := gameID + "\x00" + playerID
	admissions.mu.RLock()
	membership, found := admissions.memberships[membershipID]
	admissions.mu.RUnlock()
	if !found {
		return CharacterRecord{}, ErrLease
	}
	worker, found := admissions.workers.Game(gameID)
	if !found {
		return CharacterRecord{}, ErrGameNotFound
	}
	character, err := worker.ProjectCharacter(ctx, playerID, membership.baseline.Character)
	if err != nil {
		return CharacterRecord{}, err
	}
	receipt, err := admissions.membershipStore.Depart(ctx, MembershipRecord{GameID: gameID, PlayerID: playerID,
		AccountID: membership.baseline.AccountID, Baseline: membership.baseline, Lease: membership.lease,
		State: MembershipActive}, character)
	if err != nil {
		return CharacterRecord{}, err
	}
	admissions.mu.Lock()
	if current, exists := admissions.memberships[membershipID]; !exists || current.lease.Token != membership.lease.Token {
		admissions.mu.Unlock()
		return CharacterRecord{}, ErrLease
	}
	delete(admissions.memberships, membershipID)
	admissions.mu.Unlock()
	return receipt.Record, nil
}

// AbandonGame releases leases after the allocation is proven unavailable.
// Canonical projection is impossible in this path, so the last durable Realm
// revision remains unchanged and no client-authored fallback is accepted.
func (admissions *Admissions) AbandonGame(ctx context.Context, gameID string) error {
	if admissions == nil || ctx == nil {
		return ErrAdmission
	}
	admissions.entryMu.Lock()
	defer admissions.entryMu.Unlock()
	gameID = strings.TrimSpace(gameID)
	admissions.mu.Lock()
	if _, found := admissions.games[gameID]; !found {
		admissions.mu.Unlock()
		return ErrGameNotFound
	}
	var leases []CharacterLease
	for membershipID, membership := range admissions.memberships {
		if membership.lease.GameID == gameID {
			leases = append(leases, membership.lease)
			delete(admissions.memberships, membershipID)
		}
	}
	delete(admissions.games, gameID)
	admissions.mu.Unlock()
	var result error
	for _, lease := range leases {
		result = errors.Join(result, admissions.characters.Release(ctx, lease))
	}
	result = errors.Join(result, admissions.membershipStore.AbandonGame(ctx, gameID))
	return result
}

func cloneRuntimeIdentity(identity simulation.RuntimeIdentity) simulation.RuntimeIdentity {
	identity.Recipe.Packages.Extensions = append([]simulation.RuntimePackage(nil), identity.Recipe.Packages.Extensions...)
	identity.Recipe.CapabilityVersions = cloneStrings(identity.Recipe.CapabilityVersions)
	return identity
}

func cloneStrings(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
