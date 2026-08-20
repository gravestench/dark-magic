package realm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

var (
	ErrCharacterNotFound = errors.New("realm: character not found")
	ErrCharacterOwner    = errors.New("realm: character ownership differs")
	ErrCharacterLeased   = errors.New("realm: character is already leased")
	ErrCharacterOnline   = errors.New("realm: character is already online")
	ErrLease             = errors.New("realm: invalid character lease")
	ErrCharacterCommit   = errors.New("realm: invalid authoritative character commit")
	ErrCharacterExists   = errors.New("realm: character name already exists")
	ErrCharacterLimit    = errors.New("realm: character limit reached")
	ErrCharacterInput    = errors.New("realm: invalid character input")
)

type CharacterRecord struct {
	AccountID     string
	Revision      uint64
	Character     d2save.Character
	Compatibility gamesession.DurableCompatibility
}

// CharacterSummary is the client-safe realm roster projection. Account IDs and
// authoritative package compatibility remain server-side.
type CharacterSummary struct {
	Revision  uint64           `json:"revision"`
	Character d2save.Character `json:"character"`
}

// publicCharacter returns an independent character repository value so callers cannot mutate repository-owned state
// through a returned record.
func publicCharacter(record CharacterRecord) CharacterSummary {
	return CharacterSummary{Revision: record.Revision, Character: cloneCharacter(record.Character)}
}

type CharacterLease struct {
	Token       string
	CharacterID string
	Revision    uint64
	GameID      string
	ExpiresAt   time.Time
}

type CharacterRepository interface {
	Create(context.Context, CharacterRecord) error
	Delete(context.Context, string, string) error
	Get(context.Context, string, string) (CharacterRecord, error)
	List(context.Context, string) ([]CharacterRecord, error)
	Acquire(context.Context, string, string, string, time.Duration) (CharacterRecord, CharacterLease, error)
	BindCompatibility(context.Context, CharacterLease, gamesession.DurableCompatibility) (CharacterRecord, error)
	Renew(context.Context, CharacterLease, time.Duration) (CharacterLease, error)
	Release(context.Context, CharacterLease) error
	ReleaseGame(context.Context, string) (int, error)
	Commit(context.Context, CharacterLease, d2save.Character) (CharacterRecord, error)
}

// BindCompatibility atomically pins an unbound character to the first verified
// authoritative runtime that admits it. Existing bindings are immutable here;
// changing them requires an explicit reviewed migration.
func (repository *MemoryCharacters) BindCompatibility(
	_ context.Context,
	lease CharacterLease,
	compatibility gamesession.DurableCompatibility,
) (CharacterRecord, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	entry, found := repository.records[lease.CharacterID]
	if !found || !sameLease(entry.lease, lease) || !entry.lease.ExpiresAt.After(repository.now()) ||
		compatibility.CharacterID != lease.CharacterID || strings.TrimSpace(compatibility.ModID) == "" ||
		strings.TrimSpace(compatibility.ContractVersion) == "" || strings.TrimSpace(compatibility.IdentityHash) == "" {
		return CharacterRecord{}, ErrCharacterCommit
	}

	if !emptyCompatibility(entry.record.Compatibility) && entry.record.Compatibility != compatibility {
		return CharacterRecord{}, ErrCharacterCommit
	}

	entry.record.Compatibility = compatibility

	return cloneCharacterRecord(entry.record), nil
}

// emptyCompatibility checks the character repository invariant before state changes, keeping invalid values off shared
// paths.
func emptyCompatibility(value gamesession.DurableCompatibility) bool {
	return value == (gamesession.DurableCompatibility{})
}

type memoryCharacter struct {
	record CharacterRecord
	lease  *CharacterLease
}

// MemoryCharacters is a deterministic in-memory implementation of the realm
// persistence boundary. Durable database adapters must preserve its ownership,
// revision, exclusivity, and compare-before-release semantics.
type MemoryCharacters struct {
	mu      sync.Mutex
	now     func() time.Time
	records map[string]*memoryCharacter
}

// NewMemoryCharacters constructs the character repository boundary and validates dependencies before callers can
// publish or mutate shared state.
func NewMemoryCharacters(records ...CharacterRecord) (*MemoryCharacters, error) {
	repository := &MemoryCharacters{now: time.Now, records: make(map[string]*memoryCharacter)}
	for _, record := range records {
		if err := repository.Create(context.Background(), record); err != nil {
			return nil, err
		}
	}

	return repository, nil
}

// Create coordinates create through the owning character repository synchronization boundary so shared state is
// published only after a complete transition.
func (repository *MemoryCharacters) Create(ctx context.Context, record CharacterRecord) error {
	if err := contextErr(ctx); err != nil {
		return err
	}

	if repository == nil || strings.TrimSpace(record.AccountID) == "" || strings.TrimSpace(record.Character.ID) == "" ||
		record.Revision == 0 {
		return errors.New("realm: character record requires account, character, and revision")
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	if _, exists := repository.records[record.Character.ID]; exists {
		return fmt.Errorf("realm: duplicate character %q", record.Character.ID)
	}

	repository.records[record.Character.ID] = &memoryCharacter{record: cloneCharacterRecord(record)}

	return nil
}

// Delete coordinates delete through the owning character repository synchronization boundary so shared state is
// published only after a complete transition.
func (repository *MemoryCharacters) Delete(ctx context.Context, accountID, characterID string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}

	if repository == nil || strings.TrimSpace(accountID) == "" || strings.TrimSpace(characterID) == "" {
		return ErrCharacterNotFound
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	entry := repository.records[characterID]
	if entry == nil {
		return ErrCharacterNotFound
	}

	if entry.record.AccountID != accountID {
		return ErrCharacterOwner
	}

	if entry.lease != nil && entry.lease.ExpiresAt.After(repository.now()) {
		return ErrCharacterLeased
	}

	delete(repository.records, characterID)

	return nil
}

// Get coordinates get through the owning character repository synchronization boundary so shared state is published
// only after a complete transition.
func (repository *MemoryCharacters) Get(ctx context.Context, accountID, characterID string) (CharacterRecord, error) {
	if err := contextErr(ctx); err != nil {
		return CharacterRecord{}, err
	}

	if repository == nil || strings.TrimSpace(accountID) == "" || strings.TrimSpace(characterID) == "" {
		return CharacterRecord{}, ErrCharacterNotFound
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	entry := repository.records[characterID]
	if entry == nil {
		return CharacterRecord{}, ErrCharacterNotFound
	}

	if entry.record.AccountID != accountID {
		return CharacterRecord{}, ErrCharacterOwner
	}

	return cloneCharacterRecord(entry.record), nil
}

// List coordinates list through the owning character repository synchronization boundary so shared state is published
// only after a complete transition.
func (repository *MemoryCharacters) List(ctx context.Context, accountID string) ([]CharacterRecord, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}

	if repository == nil || strings.TrimSpace(accountID) == "" {
		return nil, ErrCharacterOwner
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	result := make([]CharacterRecord, 0)

	for _, entry := range repository.records {
		if entry.record.AccountID == accountID {
			result = append(result, cloneCharacterRecord(entry.record))
		}
	}

	sort.Slice(result, func(i, j int) bool { return result[i].Character.ID < result[j].Character.ID })

	return result, nil
}

// Acquire coordinates acquire through the owning character repository synchronization boundary so shared state is
// published only after a complete transition.
func (repository *MemoryCharacters) Acquire(
	_ context.Context,
	accountID, characterID, gameID string,
	lifetime time.Duration,
) (CharacterRecord, CharacterLease, error) {
	if strings.TrimSpace(accountID) == "" || strings.TrimSpace(characterID) == "" || strings.TrimSpace(gameID) == "" ||
		lifetime <= 0 {
		return CharacterRecord{}, CharacterLease{}, ErrLease
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	entry, found := repository.records[characterID]
	if !found {
		return CharacterRecord{}, CharacterLease{}, ErrCharacterNotFound
	}

	if entry.record.AccountID != accountID {
		return CharacterRecord{}, CharacterLease{}, ErrCharacterOwner
	}

	if entry.lease != nil && entry.lease.ExpiresAt.After(repository.now()) {
		return CharacterRecord{}, CharacterLease{}, ErrCharacterLeased
	}

	lease, err := newCharacterLease(characterID, entry.record.Revision, gameID, repository.now().Add(lifetime))
	if err != nil {
		return CharacterRecord{}, CharacterLease{}, err
	}

	entry.lease = &lease

	return cloneCharacterRecord(entry.record), lease, nil
}

// Renew coordinates renew through the owning character repository synchronization boundary so shared state is
// published only after a complete transition.
func (repository *MemoryCharacters) Renew(
	_ context.Context,
	lease CharacterLease,
	lifetime time.Duration,
) (CharacterLease, error) {
	if lifetime <= 0 {
		return CharacterLease{}, ErrLease
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	entry, found := repository.records[lease.CharacterID]
	if !found || !sameLease(entry.lease, lease) || !entry.lease.ExpiresAt.After(repository.now()) {
		return CharacterLease{}, ErrLease
	}

	entry.lease.ExpiresAt = repository.now().Add(lifetime)

	return *entry.lease, nil
}

// Release coordinates release through the owning character repository synchronization boundary so shared state is
// published only after a complete transition.
func (repository *MemoryCharacters) Release(_ context.Context, lease CharacterLease) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	entry, found := repository.records[lease.CharacterID]
	if !found || !sameLease(entry.lease, lease) {
		return ErrLease
	}

	entry.lease = nil

	return nil
}

// ReleaseGame is the fail-closed restart path used only after Realm proves the
// former allocation authority is unavailable. It never commits replacement
// character state; the last durable revision remains canonical.
func (repository *MemoryCharacters) ReleaseGame(ctx context.Context, gameID string) (int, error) {
	if err := contextErr(ctx); err != nil {
		return 0, err
	}

	gameID = strings.TrimSpace(gameID)
	if repository == nil || gameID == "" {
		return 0, ErrLease
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	released := 0

	for _, entry := range repository.records {
		if entry.lease != nil && entry.lease.GameID == gameID {
			entry.lease = nil
			released++
		}
	}

	return released, nil
}

// Commit atomically replaces realm-owned character state, advances its
// revision, and consumes the active worker lease. Offline/client stores never
// receive this lease and therefore cannot write trusted realm state.
func (repository *MemoryCharacters) Commit(
	_ context.Context,
	lease CharacterLease,
	character d2save.Character,
) (CharacterRecord, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	entry, found := repository.records[lease.CharacterID]
	if !found || !sameLease(entry.lease, lease) || !entry.lease.ExpiresAt.After(repository.now()) ||
		strings.TrimSpace(character.ID) == "" || character.ID != entry.record.Character.ID {
		return CharacterRecord{}, ErrCharacterCommit
	}

	entry.record.Character = cloneCharacter(character)
	entry.record.Revision++
	entry.lease = nil

	return cloneCharacterRecord(entry.record), nil
}

// newCharacterLease constructs the character repository boundary and validates dependencies before callers can publish
// or mutate shared state.
func newCharacterLease(characterID string, revision uint64, gameID string, expiry time.Time) (CharacterLease, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return CharacterLease{}, err
	}

	return CharacterLease{
		Token:       hex.EncodeToString(nonce[:]),
		CharacterID: characterID,
		Revision:    revision,
		GameID:      gameID,
		ExpiresAt:   expiry,
	}, nil
}

// sameLease checks the character repository invariant before state changes, keeping invalid values off shared paths.
func sameLease(current *CharacterLease, candidate CharacterLease) bool {
	return current != nil && current.Token != "" && current.Token == candidate.Token &&
		current.CharacterID == candidate.CharacterID &&
		current.Revision == candidate.Revision &&
		current.GameID == candidate.GameID
}

// cloneCharacterRecord returns an independent character repository value so callers cannot mutate repository-owned
// state through a returned record.
func cloneCharacterRecord(record CharacterRecord) CharacterRecord {
	record.Character = cloneCharacter(record.Character)
	return record
}

// cloneCharacter returns an independent character repository value so callers cannot mutate repository-owned state
// through a returned record.
func cloneCharacter(character d2save.Character) d2save.Character {
	copyStore := d2save.New(character)
	return copyStore.Characters()[0]
}
