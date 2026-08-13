package realm

import (
	"context"
	"errors"
	"testing"
	"time"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestMemoryCharactersEnforcesOwnerExclusiveRevisionedLease(t *testing.T) {
	repository, err := NewMemoryCharacters(CharacterRecord{
		AccountID: "account:a", Revision: 7,
		Character: d2save.Character{ID: "character:a", Name: "Alice", Stats: &d2save.Stats{Health: 20}},
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	repository.now = func() time.Time { return now }
	if _, _, err := repository.Acquire(context.Background(), "account:b", "character:a", "game:1", time.Minute); !errors.Is(err, ErrCharacterOwner) {
		t.Fatalf("owner error = %v", err)
	}
	record, lease, err := repository.Acquire(context.Background(), "account:a", "character:a", "game:1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if record.Revision != 7 || lease.Revision != 7 || lease.Token == "" {
		t.Fatalf("record/lease = %#v/%#v", record, lease)
	}
	record.Character.Stats.Health = 0
	if _, _, err := repository.Acquire(context.Background(), "account:a", "character:a", "game:2", time.Minute); !errors.Is(err, ErrCharacterLeased) {
		t.Fatalf("exclusive error = %v", err)
	}
	renewed, err := repository.Renew(context.Background(), lease, 2*time.Minute)
	if err != nil || !renewed.ExpiresAt.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("renewed=%#v error=%v", renewed, err)
	}
	if err := repository.Release(context.Background(), lease); err != nil {
		t.Fatal(err)
	}
	second, _, err := repository.Acquire(context.Background(), "account:a", "character:a", "game:2", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if second.Character.Stats.Health != 20 {
		t.Fatal("repository record was mutated through acquired copy")
	}
}

func TestMemoryCharactersExpiresAndRejectsStaleLease(t *testing.T) {
	repository, err := NewMemoryCharacters(CharacterRecord{AccountID: "account", Revision: 1, Character: d2save.Character{ID: "character"}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	repository.now = func() time.Time { return now }
	_, stale, err := repository.Acquire(context.Background(), "account", "character", "game:1", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Second)
	_, current, err := repository.Acquire(context.Background(), "account", "character", "game:2", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Release(context.Background(), stale); !errors.Is(err, ErrLease) {
		t.Fatalf("stale release error = %v", err)
	}
	if _, err := repository.Renew(context.Background(), stale, time.Minute); !errors.Is(err, ErrLease) {
		t.Fatalf("stale renew error = %v", err)
	}
	if err := repository.Release(context.Background(), current); err != nil {
		t.Fatal(err)
	}
}
