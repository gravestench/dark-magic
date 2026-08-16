package realm

import (
	"errors"
	"sync"
	"testing"
	"time"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

func TestMemoryMembershipDepartureCommitsOnceAndRetainsReceipt(t *testing.T) {
	record := CharacterRecord{AccountID: "account", Revision: 1,
		Character: d2save.Character{ID: "character", Name: "Hero", Class: "Amazon", Level: 1}}
	characters, err := NewMemoryCharacters(record)
	if err != nil {
		t.Fatal(err)
	}
	baseline, lease, err := characters.Acquire(t.Context(), "account", "character", "game", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewMemoryMemberships(characters)
	if err != nil {
		t.Fatal(err)
	}
	membership := MembershipRecord{GameID: "game", PlayerID: "player", AccountID: "account",
		Baseline: baseline, Lease: lease, State: MembershipActive}
	if err := store.Admit(t.Context(), membership); err != nil {
		t.Fatal(err)
	}
	if err := store.Admit(t.Context(), MembershipRecord{GameID: "game", PlayerID: "other", AccountID: "account",
		Baseline: baseline, Lease: lease, State: MembershipActive}); !errors.Is(err, ErrCharacterLeased) {
		t.Fatalf("duplicate account membership = %v", err)
	}
	canonical := cloneCharacter(baseline.Character)
	canonical.Level = 2
	var group sync.WaitGroup
	receipts := make(chan departureReceipt, 2)
	errorsSeen := make(chan error, 2)
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			receipt, departErr := store.Depart(t.Context(), membership, canonical)
			receipts <- receipt
			errorsSeen <- departErr
		}()
	}
	group.Wait()
	close(receipts)
	close(errorsSeen)
	for departErr := range errorsSeen {
		if departErr != nil {
			t.Fatal(departErr)
		}
	}
	for receipt := range receipts {
		if receipt.Record.Revision != 2 || receipt.Record.Character.Level != 2 || receipt.WorkerRemoved {
			t.Fatalf("departure receipt = %#v", receipt)
		}
	}
	persisted, err := store.ByAccount(t.Context(), "game", "account")
	if err != nil || persisted.State != MembershipDeparted || persisted.Departure == nil {
		t.Fatalf("persisted membership = %#v, %v", persisted, err)
	}
	if err := store.AbandonGame(t.Context(), "game"); err != nil {
		t.Fatal(err)
	}
	abandoned, err := store.ByAccount(t.Context(), "game", "account")
	if err != nil || abandoned.Departure == nil || !abandoned.Departure.WorkerRemoved {
		t.Fatalf("abandoned departure = %#v, %v", abandoned, err)
	}
	removed, err := store.MarkWorkerRemoved(t.Context(), "game", "player")
	if err != nil || !removed.WorkerRemoved {
		t.Fatalf("worker removal receipt = %#v, %v", removed, err)
	}
	if repeated, err := store.MarkWorkerRemoved(t.Context(), "game", "player"); err != nil || !repeated.WorkerRemoved {
		t.Fatalf("repeated worker removal = %#v, %v", repeated, err)
	}
	committed, err := characters.Get(t.Context(), "account", "character")
	if err != nil || committed.Revision != 2 {
		t.Fatalf("committed character = %#v, %v", committed, err)
	}
}

func TestMemoryMembershipResumeRenewsActiveLease(t *testing.T) {
	record := CharacterRecord{AccountID: "account", Revision: 1,
		Character: d2save.Character{ID: "character", Name: "Hero", Class: "Amazon", Level: 1}}
	characters, err := NewMemoryCharacters(record)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.August, 15, 9, 0, 0, 0, time.UTC)
	characters.now = func() time.Time { return now }
	baseline, lease, err := characters.Acquire(t.Context(), "account", "character", "game", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewMemoryMemberships(characters)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Admit(t.Context(), MembershipRecord{GameID: "game", PlayerID: "player", AccountID: "account",
		Baseline: baseline, Lease: lease, State: MembershipActive}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(30 * time.Second)
	resumed, err := store.ResumeGame(t.Context(), "game", 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(resumed) != 1 || resumed[0].PlayerID != "player" || resumed[0].Lease.Token != lease.Token ||
		!resumed[0].Lease.ExpiresAt.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("resumed memberships = %#v", resumed)
	}
}
