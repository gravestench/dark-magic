package realm

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

type MembershipState string

const (
	MembershipActive    MembershipState = "active"
	MembershipDeparted  MembershipState = "departed"
	MembershipAbandoned MembershipState = "abandoned"
)

var ErrMembership = errors.New("realm: invalid game membership")

type MembershipRecord struct {
	GameID    string
	PlayerID  string
	AccountID string
	Baseline  CharacterRecord
	Lease     CharacterLease
	State     MembershipState
	Departure *departureReceipt
}

// MembershipRepository owns the durable half of player admission and
// departure. Depart must atomically commit the canonical character, consume its
// lease, and persist the retry receipt. Raw lease tokens remain process-private
// and are never stored in membership rows.
type MembershipRepository interface {
	Admit(context.Context, MembershipRecord) error
	Cancel(context.Context, string, string) error
	Depart(context.Context, MembershipRecord, d2save.Character) (departureReceipt, error)
	MarkWorkerRemoved(context.Context, string, string) (departureReceipt, error)
	ByAccount(context.Context, string, string) (MembershipRecord, error)
	ByPlayer(context.Context, string, string) (MembershipRecord, error)
	ActivePlayerIDs(context.Context, string) ([]string, error)
	// DrainPlayerIDs includes both active memberships and durable departure
	// receipts whose worker/roster cleanup may still need to be retried.
	DrainPlayerIDs(context.Context, string) ([]string, error)
	// ResumeGame reconstructs active memberships after Realm authority loss.
	// Durable stores rotate every raw character-lease token atomically rather
	// than persisting process-private bearer secrets in membership rows.
	ResumeGame(context.Context, string, time.Duration) ([]MembershipRecord, error)
	AbandonGame(context.Context, string) error
}

func (store *MemoryMemberships) ActivePlayerIDs(ctx context.Context, gameID string) ([]string, error) {
	return store.playerIDs(ctx, gameID, false)
}

func (store *MemoryMemberships) DrainPlayerIDs(ctx context.Context, gameID string) ([]string, error) {
	return store.playerIDs(ctx, gameID, true)
}

func (store *MemoryMemberships) playerIDs(ctx context.Context, gameID string, includeDeparted bool) ([]string, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	gameID = strings.TrimSpace(gameID)
	if store == nil || gameID == "" {
		return nil, ErrMembership
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	result := make([]string, 0)
	for _, record := range store.records {
		if record.GameID == gameID && (record.State == MembershipActive ||
			(includeDeparted && record.State == MembershipDeparted)) {
			result = append(result, record.PlayerID)
		}
	}
	sort.Strings(result)
	return result, nil
}

type MemoryMemberships struct {
	mu         sync.Mutex
	characters CharacterRepository
	records    map[string]MembershipRecord
}

func NewMemoryMemberships(characters CharacterRepository) (*MemoryMemberships, error) {
	if characters == nil {
		return nil, ErrMembership
	}
	return &MemoryMemberships{characters: characters, records: make(map[string]MembershipRecord)}, nil
}

func (store *MemoryMemberships) Admit(ctx context.Context, record MembershipRecord) error {
	if err := validateActiveMembership(record); err != nil {
		return err
	}
	if err := contextErr(ctx); err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	for key, existing := range store.records {
		if existing.GameID == record.GameID && existing.AccountID == record.AccountID {
			if existing.State == MembershipActive {
				return ErrCharacterLeased
			}
			delete(store.records, key)
		}
	}
	key := membershipKey(record.GameID, record.PlayerID)
	if _, exists := store.records[key]; exists {
		return ErrMembership
	}
	record.State = MembershipActive
	store.records[key] = cloneMembershipRecord(record)
	return nil
}

func (store *MemoryMemberships) Cancel(ctx context.Context, gameID, playerID string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	key := membershipKey(gameID, playerID)
	record, found := store.records[key]
	if !found {
		return nil
	}
	if record.State != MembershipActive {
		return ErrMembership
	}
	delete(store.records, key)
	return nil
}

func (store *MemoryMemberships) Depart(ctx context.Context, wanted MembershipRecord, character d2save.Character) (departureReceipt, error) {
	if err := contextErr(ctx); err != nil {
		return departureReceipt{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	key := membershipKey(wanted.GameID, wanted.PlayerID)
	record, found := store.records[key]
	if !found {
		return departureReceipt{}, ErrMembership
	}
	if record.State == MembershipDeparted && record.Departure != nil {
		return cloneDepartureReceipt(*record.Departure), nil
	}
	if record.State != MembershipActive || !sameMembership(record, wanted) {
		return departureReceipt{}, ErrMembership
	}
	committed, err := store.characters.Commit(ctx, record.Lease, character)
	if err != nil {
		return departureReceipt{}, err
	}
	receipt := departureReceipt{Record: committed, PlayerID: record.PlayerID}
	record.State, record.Departure = MembershipDeparted, &receipt
	store.records[key] = cloneMembershipRecord(record)
	return cloneDepartureReceipt(receipt), nil
}

func (store *MemoryMemberships) MarkWorkerRemoved(ctx context.Context, gameID, playerID string) (departureReceipt, error) {
	if err := contextErr(ctx); err != nil {
		return departureReceipt{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	key := membershipKey(gameID, playerID)
	record, found := store.records[key]
	if !found || record.State != MembershipDeparted || record.Departure == nil {
		return departureReceipt{}, ErrMembership
	}
	record.Departure.WorkerRemoved = true
	store.records[key] = cloneMembershipRecord(record)
	return cloneDepartureReceipt(*record.Departure), nil
}

func (store *MemoryMemberships) ByAccount(ctx context.Context, gameID, accountID string) (MembershipRecord, error) {
	if err := contextErr(ctx); err != nil {
		return MembershipRecord{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	for _, record := range store.records {
		if record.GameID == strings.TrimSpace(gameID) && record.AccountID == strings.TrimSpace(accountID) {
			return cloneMembershipRecord(record), nil
		}
	}
	return MembershipRecord{}, ErrMembership
}

func (store *MemoryMemberships) ByPlayer(ctx context.Context, gameID, playerID string) (MembershipRecord, error) {
	if err := contextErr(ctx); err != nil {
		return MembershipRecord{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	record, found := store.records[membershipKey(gameID, playerID)]
	if !found {
		return MembershipRecord{}, ErrMembership
	}
	return cloneMembershipRecord(record), nil
}

func (store *MemoryMemberships) ResumeGame(ctx context.Context, gameID string, lifetime time.Duration) ([]MembershipRecord, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	gameID = strings.TrimSpace(gameID)
	if store == nil || gameID == "" || lifetime <= 0 {
		return nil, ErrMembership
	}
	store.mu.Lock()
	candidates := make([]MembershipRecord, 0)
	for _, record := range store.records {
		if record.GameID != gameID || record.State != MembershipActive {
			continue
		}
		candidates = append(candidates, cloneMembershipRecord(record))
	}
	store.mu.Unlock()
	result := make([]MembershipRecord, 0, len(candidates))
	for _, record := range candidates {
		previousToken := record.Lease.Token
		renewed, err := store.characters.Renew(ctx, record.Lease, lifetime)
		if err != nil {
			return nil, err
		}
		record.Lease = renewed
		store.mu.Lock()
		key := membershipKey(record.GameID, record.PlayerID)
		current, found := store.records[key]
		if !found || current.State != MembershipActive || current.Lease.Token != previousToken {
			store.mu.Unlock()
			return nil, ErrMembership
		}
		store.records[key] = cloneMembershipRecord(record)
		store.mu.Unlock()
		result = append(result, cloneMembershipRecord(record))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].PlayerID < result[j].PlayerID })
	return result, nil
}

func (store *MemoryMemberships) AbandonGame(ctx context.Context, gameID string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	for key, record := range store.records {
		if record.GameID != strings.TrimSpace(gameID) {
			continue
		}
		if record.State == MembershipActive {
			record.State = MembershipAbandoned
			record.Lease = CharacterLease{}
		}
		if record.State == MembershipDeparted && record.Departure != nil {
			record.Departure.WorkerRemoved = true
		}
		store.records[key] = record
	}
	return nil
}

func validateActiveMembership(record MembershipRecord) error {
	if strings.TrimSpace(record.GameID) == "" || strings.TrimSpace(record.PlayerID) == "" ||
		strings.TrimSpace(record.AccountID) == "" || strings.TrimSpace(record.Baseline.Character.ID) == "" ||
		record.Lease.Token == "" || record.Lease.GameID != record.GameID ||
		record.Lease.CharacterID != record.Baseline.Character.ID || record.Lease.Revision != record.Baseline.Revision {
		return ErrMembership
	}
	return nil
}

func sameMembership(current, wanted MembershipRecord) bool {
	return current.GameID == wanted.GameID && current.PlayerID == wanted.PlayerID &&
		current.AccountID == wanted.AccountID && sameLease(&current.Lease, wanted.Lease)
}

func membershipKey(gameID, playerID string) string {
	return strings.TrimSpace(gameID) + "\x00" + strings.TrimSpace(playerID)
}

func cloneMembershipRecord(record MembershipRecord) MembershipRecord {
	record.Baseline = cloneCharacterRecord(record.Baseline)
	if record.Departure != nil {
		receipt := cloneDepartureReceipt(*record.Departure)
		record.Departure = &receipt
	}
	return record
}

func cloneDepartureReceipt(receipt departureReceipt) departureReceipt {
	receipt.Record = cloneCharacterRecord(receipt.Record)
	return receipt
}

var _ MembershipRepository = (*MemoryMemberships)(nil)
