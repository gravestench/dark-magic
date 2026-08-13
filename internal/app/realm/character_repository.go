package realm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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
	ErrLease             = errors.New("realm: invalid character lease")
)

type CharacterRecord struct {
	AccountID     string
	Revision      uint64
	Character     d2save.Character
	Compatibility gamesession.DurableCompatibility
}

type CharacterLease struct {
	Token       string
	CharacterID string
	Revision    uint64
	GameID      string
	ExpiresAt   time.Time
}

type CharacterRepository interface {
	Acquire(context.Context, string, string, string, time.Duration) (CharacterRecord, CharacterLease, error)
	Renew(context.Context, CharacterLease, time.Duration) (CharacterLease, error)
	Release(context.Context, CharacterLease) error
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

func NewMemoryCharacters(records ...CharacterRecord) (*MemoryCharacters, error) {
	repository := &MemoryCharacters{now: time.Now, records: make(map[string]*memoryCharacter)}
	for _, record := range records {
		if strings.TrimSpace(record.AccountID) == "" || strings.TrimSpace(record.Character.ID) == "" || record.Revision == 0 {
			return nil, errors.New("realm: character record requires account, character, and revision")
		}
		if _, exists := repository.records[record.Character.ID]; exists {
			return nil, fmt.Errorf("realm: duplicate character %q", record.Character.ID)
		}
		repository.records[record.Character.ID] = &memoryCharacter{record: cloneCharacterRecord(record)}
	}
	return repository, nil
}

func (repository *MemoryCharacters) Acquire(_ context.Context, accountID, characterID, gameID string, lifetime time.Duration) (CharacterRecord, CharacterLease, error) {
	if strings.TrimSpace(accountID) == "" || strings.TrimSpace(characterID) == "" || strings.TrimSpace(gameID) == "" || lifetime <= 0 {
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

func (repository *MemoryCharacters) Renew(_ context.Context, lease CharacterLease, lifetime time.Duration) (CharacterLease, error) {
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

func newCharacterLease(characterID string, revision uint64, gameID string, expiry time.Time) (CharacterLease, error) {
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return CharacterLease{}, err
	}
	return CharacterLease{Token: hex.EncodeToString(nonce[:]), CharacterID: characterID, Revision: revision, GameID: gameID, ExpiresAt: expiry}, nil
}

func sameLease(current *CharacterLease, candidate CharacterLease) bool {
	return current != nil && current.Token != "" && current.Token == candidate.Token && current.CharacterID == candidate.CharacterID && current.Revision == candidate.Revision && current.GameID == candidate.GameID
}

func cloneCharacterRecord(record CharacterRecord) CharacterRecord {
	copyStore := d2save.New(record.Character)
	characters := copyStore.Characters()
	record.Character = characters[0]
	return record
}
