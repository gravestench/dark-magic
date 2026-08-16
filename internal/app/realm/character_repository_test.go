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

func TestMemoryCharactersCommitsOnlyThroughActiveLease(t *testing.T) {
	repository, err := NewMemoryCharacters(CharacterRecord{
		AccountID: "account", Revision: 3,
		Character: d2save.Character{ID: "character", Name: "Before", Stats: &d2save.Stats{Health: 10}},
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(100, 0)
	repository.now = func() time.Time { return now }
	_, lease, err := repository.Acquire(context.Background(), "account", "character", "game", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Commit(context.Background(), CharacterLease{}, d2save.Character{ID: "character"}); !errors.Is(err, ErrCharacterCommit) {
		t.Fatalf("unleased commit error = %v", err)
	}
	if _, err := repository.Commit(context.Background(), lease, d2save.Character{ID: "offline-character"}); !errors.Is(err, ErrCharacterCommit) {
		t.Fatalf("foreign character commit error = %v", err)
	}
	committed, err := repository.Commit(context.Background(), lease, d2save.Character{
		ID: "character", Name: "After", Stats: &d2save.Stats{Health: 25},
	})
	if err != nil {
		t.Fatal(err)
	}
	if committed.Revision != 4 || committed.Character.Name != "After" {
		t.Fatalf("committed = %#v", committed)
	}
	committed.Character.Stats.Health = 0
	if _, err := repository.Commit(context.Background(), lease, d2save.Character{ID: "character"}); !errors.Is(err, ErrCharacterCommit) {
		t.Fatalf("replayed lease commit error = %v", err)
	}
	reloaded, current, err := repository.Acquire(context.Background(), "account", "character", "next-game", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Revision != 4 || reloaded.Character.Stats.Health != 25 || current.Revision != 4 {
		t.Fatalf("reloaded = %#v lease=%#v", reloaded, current)
	}
}

func TestMemoryCharactersListsOnlyOwnedDefensiveRecords(t *testing.T) {
	repository, err := NewMemoryCharacters(
		CharacterRecord{AccountID: "account:a", Revision: 1, Character: d2save.Character{ID: "character:b", Name: "B", Stats: &d2save.Stats{Health: 20}}},
		CharacterRecord{AccountID: "account:a", Revision: 1, Character: d2save.Character{ID: "character:a", Name: "A"}},
		CharacterRecord{AccountID: "account:b", Revision: 1, Character: d2save.Character{ID: "character:c", Name: "C"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	listed, err := repository.List(t.Context(), "account:a")
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 || listed[0].Character.ID != "character:a" || listed[1].Character.ID != "character:b" {
		t.Fatalf("listed = %#v", listed)
	}
	listed[1].Character.Stats.Health = 0
	loaded, err := repository.Get(t.Context(), "account:a", "character:b")
	if err != nil || loaded.Character.Stats.Health != 20 {
		t.Fatalf("loaded=%#v error=%v", loaded, err)
	}
	if _, err := repository.Get(t.Context(), "account:b", "character:b"); !errors.Is(err, ErrCharacterOwner) {
		t.Fatalf("foreign owner error = %v", err)
	}
}

func TestMemoryCharactersDeletesOnlyOwnedIdleCharacter(t *testing.T) {
	repository, err := NewMemoryCharacters(CharacterRecord{
		AccountID: "account:a", Revision: 1, Character: d2save.Character{ID: "character:a", Name: "A"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Delete(t.Context(), "account:b", "character:a"); !errors.Is(err, ErrCharacterOwner) {
		t.Fatalf("foreign delete error = %v", err)
	}
	_, lease, err := repository.Acquire(t.Context(), "account:a", "character:a", "game", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Delete(t.Context(), "account:a", "character:a"); !errors.Is(err, ErrCharacterLeased) {
		t.Fatalf("leased delete error = %v", err)
	}
	if err := repository.Release(t.Context(), lease); err != nil {
		t.Fatal(err)
	}
	if err := repository.Delete(t.Context(), "account:a", "character:a"); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Get(t.Context(), "account:a", "character:a"); !errors.Is(err, ErrCharacterNotFound) {
		t.Fatalf("deleted get error = %v", err)
	}
}

func TestMemoryCharactersReleaseGameFailsClosedWithoutCommitting(t *testing.T) {
	repository, err := NewMemoryCharacters(
		CharacterRecord{AccountID: "account", Revision: 3, Character: d2save.Character{ID: "one", Name: "One"}},
		CharacterRecord{AccountID: "account", Revision: 5, Character: d2save.Character{ID: "two", Name: "Two"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.Acquire(t.Context(), "account", "one", "interrupted", time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, other, err := repository.Acquire(t.Context(), "account", "two", "healthy", time.Minute); err != nil {
		t.Fatal(err)
	} else {
		t.Cleanup(func() { _ = repository.Release(context.Background(), other) })
	}
	released, err := repository.ReleaseGame(t.Context(), "interrupted")
	if err != nil || released != 1 {
		t.Fatalf("released = %d, %v", released, err)
	}
	record, lease, err := repository.Acquire(t.Context(), "account", "one", "replacement", time.Minute)
	if err != nil || record.Revision != 3 {
		t.Fatalf("replacement record = %#v lease=%#v error=%v", record, lease, err)
	}
}
