package realm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
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
	authority *gameserver.TicketAuthority
	endpoint  GameEndpoint
}

// Admissions coordinates realm-owned leases and trusted character entry with
// a worker-owned ticket authority. It never executes gameplay rules itself.
type Admissions struct {
	mu             sync.RWMutex
	entryMu        sync.Mutex
	manager        *Manager
	characters     CharacterRepository
	leaseLifetime  time.Duration
	ticketLifetime time.Duration
	games          map[string]admissionGame
	memberships    map[string]CharacterLease
	entrySequences map[string]uint64
}

func NewAdmissions(manager *Manager, characters CharacterRepository, leaseLifetime, ticketLifetime time.Duration) (*Admissions, error) {
	if manager == nil || characters == nil || leaseLifetime <= 0 || ticketLifetime <= 0 || ticketLifetime > leaseLifetime {
		return nil, errors.New("realm: admissions require manager, characters, and bounded lifetimes")
	}
	return &Admissions{manager: manager, characters: characters, leaseLifetime: leaseLifetime, ticketLifetime: ticketLifetime,
		games: make(map[string]admissionGame), memberships: make(map[string]CharacterLease), entrySequences: make(map[string]uint64)}, nil
}

func (admissions *Admissions) RegisterGame(gameID string, authority *gameserver.TicketAuthority, endpoint GameEndpoint) error {
	gameID = strings.TrimSpace(gameID)
	if gameID == "" || authority == nil || strings.TrimSpace(endpoint.Address) == "" || strings.TrimSpace(endpoint.TLSFingerprint) == "" {
		return ErrAdmission
	}
	if _, found := admissions.manager.Game(gameID); !found {
		return ErrGameNotFound
	}
	admissions.mu.Lock()
	defer admissions.mu.Unlock()
	if _, exists := admissions.games[gameID]; exists {
		return ErrGameExists
	}
	admissions.games[gameID] = admissionGame{authority: authority, endpoint: endpoint}
	return nil
}

func (admissions *Admissions) Join(ctx context.Context, request JoinRequest) (JoinAssignment, error) {
	if strings.TrimSpace(request.AccountID) == "" || strings.TrimSpace(request.CharacterID) == "" || strings.TrimSpace(request.PlayerID) == "" || strings.TrimSpace(request.GameID) == "" {
		return JoinAssignment{}, ErrAdmission
	}
	host, found := admissions.manager.Game(request.GameID)
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
		if ticket != "" {
			cause = errors.Join(cause, game.authority.Revoke(ticket))
		}
		cause = errors.Join(cause, admissions.characters.Release(ctx, lease))
		return JoinAssignment{}, fmt.Errorf("%w: %v", ErrAdmission, cause)
	}
	if err := host.Allocation.ValidateDurable(record.Compatibility); err != nil {
		return rollback(err, "")
	}
	principal := gameserver.Principal{ID: request.AccountID, CharacterID: request.CharacterID, PlayerID: request.PlayerID,
		CharacterRevision: record.Revision, RuntimeIdentityHash: host.Allocation.IdentityHash}
	ticket, err := game.authority.Issue(principal, admissions.ticketLifetime)
	if err != nil {
		return rollback(err, "")
	}
	admissions.entryMu.Lock()
	defer admissions.entryMu.Unlock()
	checkpoint, err := host.Session.CanonicalCheckpoint()
	if err != nil {
		return rollback(err, ticket)
	}
	actor := "realm:entry:" + request.PlayerID
	sequence := admissions.entrySequences[request.GameID+"\x00"+actor] + 1
	command, err := playeradapter.AdmissionCommand(record.Character, request.PlayerID, request.Destination,
		actor, sequence, checkpoint.Tick+1, simulation.AuthoritySystem)
	if err != nil {
		return rollback(err, ticket)
	}
	if err := host.Session.Submit(command); err != nil {
		return rollback(err, ticket)
	}
	admissions.entrySequences[request.GameID+"\x00"+actor] = sequence
	admissions.mu.Lock()
	admissions.memberships[membershipID] = lease
	admissions.mu.Unlock()
	return JoinAssignment{GameID: request.GameID, Endpoint: game.endpoint, Ticket: ticket,
		CharacterRevision: record.Revision, Runtime: cloneRuntimeIdentity(host.Allocation.Identity)}, nil
}

func (admissions *Admissions) RenewMembership(ctx context.Context, gameID, playerID string) (CharacterLease, error) {
	membershipID := strings.TrimSpace(gameID) + "\x00" + strings.TrimSpace(playerID)
	admissions.mu.RLock()
	lease, found := admissions.memberships[membershipID]
	admissions.mu.RUnlock()
	if !found {
		return CharacterLease{}, ErrLease
	}
	renewed, err := admissions.characters.Renew(ctx, lease, admissions.leaseLifetime)
	if err != nil {
		return CharacterLease{}, err
	}
	admissions.mu.Lock()
	if current, exists := admissions.memberships[membershipID]; !exists || current.Token != lease.Token {
		admissions.mu.Unlock()
		return CharacterLease{}, ErrLease
	}
	admissions.memberships[membershipID] = renewed
	admissions.mu.Unlock()
	return renewed, nil
}

// CommitMembership is the trusted worker-to-realm persistence boundary. The
// opaque lease stays inside Admissions and is never included in JoinAssignment.
func (admissions *Admissions) CommitMembership(ctx context.Context, gameID, playerID string, character d2save.Character) (CharacterRecord, error) {
	membershipID := strings.TrimSpace(gameID) + "\x00" + strings.TrimSpace(playerID)
	admissions.mu.RLock()
	lease, found := admissions.memberships[membershipID]
	admissions.mu.RUnlock()
	if !found {
		return CharacterRecord{}, ErrLease
	}
	committed, err := admissions.characters.Commit(ctx, lease, character)
	if err != nil {
		return CharacterRecord{}, err
	}
	admissions.mu.Lock()
	if current, exists := admissions.memberships[membershipID]; !exists || current.Token != lease.Token {
		admissions.mu.Unlock()
		return CharacterRecord{}, ErrLease
	}
	delete(admissions.memberships, membershipID)
	admissions.mu.Unlock()
	return committed, nil
}

func cloneRuntimeIdentity(identity simulation.RuntimeIdentity) simulation.RuntimeIdentity {
	identity.Dependencies = cloneStrings(identity.Dependencies)
	identity.CapabilityVersions = cloneStrings(identity.CapabilityVersions)
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
