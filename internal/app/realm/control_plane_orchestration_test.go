package realm

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestControlPlaneAllocatesAndAdmitsEachSelectedRealmCharacter covers the full
// create, join, capacity, departure, and idempotent cleanup lifecycle.
func TestControlPlaneAllocatesAndAdmitsEachSelectedRealmCharacter(t *testing.T) {
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(nil)
	config.Allocator = allocator
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	createPlayer := func(name, characterName, class string) RealmSession {
		t.Helper()
		if _, err := control.CreateAccount(t.Context(), name, "long enough password"); err != nil {
			t.Fatal(err)
		}
		session, err := control.Authenticate(t.Context(), name, "long enough password")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := control.CreateCharacter(t.Context(), session.Token, CreateCharacterRequest{
			Name: characterName, Class: class, Expansion: true}); err != nil {
			t.Fatal(err)
		}
		return session
	}
	alice := createPlayer("Alice", "Alyssa", "Assassin")
	created, err := control.CreateGame(t.Context(), alice.Token, CreateGameRequest{Name: "Trist Run",
		Difficulty: DifficultyNormal, Maximum: 2, Expansion: true})
	if err != nil {
		t.Fatal(err)
	}
	bob := createPlayer("Bob", "Borin", "Barbarian")
	joined, err := control.JoinGame(t.Context(), bob.Token, "trist run", "")
	if err != nil {
		t.Fatal(err)
	}
	if created.Assignment.GameID != joined.Assignment.GameID || created.Assignment.Ticket == joined.Assignment.Ticket ||
		created.Game.Entry.Players != 1 || joined.Game.Entry.Players != 2 || len(joined.Game.Players) != 2 {
		t.Fatalf("created=%#v joined=%#v", created, joined)
	}
	worker := allocator.workers[created.Assignment.GameID]
	if worker == nil || len(worker.admitted) != 2 || worker.admitted[0].Character.Name != "Alyssa" ||
		worker.admitted[1].Character.Name != "Borin" || worker.admitted[0].PlayerID == worker.admitted[1].PlayerID {
		t.Fatalf("worker = %#v", worker)
	}
	if worker.admitted[0].ClaimDeadline.IsZero() || worker.admitted[1].ClaimDeadline.IsZero() {
		t.Fatalf("worker claim deadlines = %s, %s", worker.admitted[0].ClaimDeadline, worker.admitted[1].ClaimDeadline)
	}
	if status, err := worker.Status(t.Context()); err != nil || status.ActivePlayers != 2 {
		t.Fatalf("worker status = %#v, %v", status, err)
	}
	if got := worker.admitted[0].Destination; got != worker.description.EntryDestination {
		t.Fatalf("admission destination = %#v, want worker-owned %#v", got, worker.description.EntryDestination)
	}
	for _, session := range []RealmSession{alice, bob} {
		record, err := control.SelectedCharacter(t.Context(), session.Token)
		if err != nil ||
			emptyCompatibility(record.Compatibility) ||
			record.Compatibility.IdentityHash != worker.description.IdentityHash {
			t.Fatalf("bound character=%#v error=%v", record, err)
		}
	}
	carol := createPlayer("Carol", "Cara", "Amazon")
	if _, err := control.JoinGame(t.Context(), carol.Token, created.Game.Entry.GameID, ""); !errors.Is(err, ErrGameFull) {
		t.Fatalf("full-game join error = %v", err)
	}
	carolRecord, err := control.SelectedCharacter(t.Context(), carol.Token)
	if err != nil {
		t.Fatal(err)
	}
	_, lease, err := control.characters.Acquire(
		t.Context(),
		carol.Account.ID,
		carolRecord.Character.ID,
		"other",
		time.Minute,
	)
	if err != nil {
		t.Fatalf("failed capacity reservation leaked character lease: %v", err)
	}
	if err := control.characters.Release(t.Context(), lease); err != nil {
		t.Fatal(err)
	}
	bobCommitted, err := control.LeaveGame(t.Context(), bob.Token, created.Game.Entry.GameID)
	if err != nil || bobCommitted.Revision != 2 {
		t.Fatalf("Bob leave committed=%#v error=%v", bobCommitted, err)
	}
	detail, err := control.GameDetail(t.Context(), alice.Token, created.Game.Entry.GameID)
	if err != nil || detail.Entry.Players != 1 || len(detail.Players) != 1 || detail.Players[0].Name != "Alyssa" {
		t.Fatalf("detail after Bob leave=%#v error=%v", detail, err)
	}
	if len(worker.removed) != 1 || worker.removed[0] != worker.admitted[1].PlayerID || allocator.releases != 0 {
		t.Fatalf("worker removals=%#v releases=%d", worker.removed, allocator.releases)
	}
	if status, err := worker.Status(t.Context()); err != nil || status.ActivePlayers != 1 {
		t.Fatalf("worker status after leave = %#v, %v", status, err)
	}
	_, bobLease, err := control.characters.Acquire(
		t.Context(),
		bob.Account.ID,
		bobCommitted.Character.ID,
		"another",
		time.Minute,
	)
	if err != nil {
		t.Fatalf("committed Bob character remained leased: %v", err)
	}
	if err := control.characters.Release(t.Context(), bobLease); err != nil {
		t.Fatal(err)
	}
	aliceCommitted, err := control.LeaveGame(t.Context(), alice.Token, created.Game.Entry.GameID)
	if err != nil || aliceCommitted.Revision != 2 || allocator.releases != 1 {
		t.Fatalf("Alice leave committed=%#v releases=%d error=%v", aliceCommitted, allocator.releases, err)
	}
	if games, err := control.ListGames(t.Context(), alice.Token, GameFilter{}); err != nil || len(games) != 0 {
		t.Fatalf("empty game remained listed: %#v error=%v", games, err)
	}
	retried, err := control.LeaveGame(t.Context(), alice.Token, created.Game.Entry.GameID)
	if err != nil || retried.Revision != aliceCommitted.Revision {
		t.Fatalf("duplicate leave = revision %d, %v", retried.Revision, err)
	}
}

// TestControlPlaneEnforcesCreateGameCharacterDifference verifies the configured
// level boundary at both its inclusive edge and first rejected value.
func TestControlPlaneEnforcesCreateGameCharacterDifference(t *testing.T) {
	control, err := NewControlPlane(orchestratedControlConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
	createPlayer := func(accountName, characterName string, level int) RealmSession {
		t.Helper()
		if _, err := control.CreateAccount(t.Context(), accountName, "long enough password"); err != nil {
			t.Fatal(err)
		}
		session, err := control.Authenticate(t.Context(), accountName, "long enough password")
		if err != nil {
			t.Fatal(err)
		}
		record, err := control.CreateCharacter(t.Context(), session.Token, CreateCharacterRequest{
			Name: characterName, Class: "Amazon", Expansion: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		repository := control.characters.(*MemoryCharacters)
		repository.mu.Lock()
		repository.records[record.Character.ID].record.Character.Level = level
		repository.mu.Unlock()
		return session
	}

	owner := createPlayer("LevelOwner", "LevelOwnerHero", 20)
	created, err := control.CreateGame(t.Context(), owner.Token, CreateGameRequest{
		Name: "Narrow Levels", Difficulty: DifficultyNormal, Maximum: 8,
		CharacterDifference: 4, Expansion: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	near := createPlayer("NearPlayer", "NearHero", 24)
	if _, err := control.JoinGame(t.Context(), near.Token, created.Game.Entry.GameID, ""); err != nil {
		t.Fatalf("boundary level rejected: %v", err)
	}
	far := createPlayer("FarPlayer", "FarHero", 25)
	if _, err := control.JoinGame(
		t.Context(),
		far.Token,
		created.Game.Entry.GameID,
		"",
	); !errors.Is(err, ErrGameLevelRange) {
		t.Fatalf("out-of-range level error = %v", err)
	}
}

// TestControlPlaneRejectsDuplicateActiveCharacterSessionWithoutDisturbingOriginal
// ensures a failed second admission leaves both existing games unchanged.
func TestControlPlaneRejectsDuplicateActiveCharacterSessionWithoutDisturbingOriginal(t *testing.T) {
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(nil)
	config.Allocator = allocator
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	createPlayer := func(accountName, characterName string) RealmSession {
		t.Helper()
		if _, err := control.CreateAccount(t.Context(), accountName, "long enough password"); err != nil {
			t.Fatal(err)
		}
		session, err := control.Authenticate(t.Context(), accountName, "long enough password")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := control.CreateCharacter(t.Context(), session.Token,
			CreateCharacterRequest{Name: characterName, Class: "Amazon"}); err != nil {
			t.Fatal(err)
		}
		return session
	}

	first := createPlayer("DuplicateAlice", "OnlyAlice")
	firstGame, err := control.CreateGame(t.Context(), first.Token,
		CreateGameRequest{Name: "Original Session", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	second := createPlayer("DuplicateBob", "OnlyBob")
	secondGame, err := control.CreateGame(t.Context(), second.Token,
		CreateGameRequest{Name: "Other Session", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := control.JoinGame(
		t.Context(),
		first.Token,
		secondGame.Game.Entry.GameID,
		"",
	); !errors.Is(err, ErrCharacterLeased) {
		t.Fatalf("duplicate active character session error = %v", err)
	}
	original, err := control.ReconnectGame(t.Context(), first.Token, firstGame.Game.Entry.GameID)
	if err != nil || original.Assignment.GameID != firstGame.Game.Entry.GameID || original.Assignment.Ticket == "" {
		t.Fatalf("original session after duplicate attempt=%#v error=%v", original, err)
	}
	other, err := control.GameDetail(t.Context(), second.Token, secondGame.Game.Entry.GameID)
	if err != nil || other.Entry.Players != 1 || len(other.Players) != 1 || other.Players[0].Name != "OnlyBob" {
		t.Fatalf("other game after duplicate attempt=%#v error=%v", other, err)
	}
}

// TestControlPlaneOperatorDrainCommitsCanonicalCharactersAndRetiresWorker
// verifies the trusted drain path completes every durable lifecycle phase.
func TestControlPlaneOperatorDrainCommitsCanonicalCharactersAndRetiresWorker(t *testing.T) {
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(nil)
	config.Allocator = allocator
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := control.CreateAccount(t.Context(), "DrainOwner", "long enough password"); err != nil {
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), "DrainOwner", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	created, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "DrainHero", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := control.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Drain Game", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	result, err := control.DrainGame(t.Context(), handoff.Game.Entry.GameID)
	if err != nil || result.GameID != handoff.Game.Entry.GameID || result.CommittedCharacters != 1 {
		t.Fatalf("drain=%#v error=%v", result, err)
	}
	if allocator.releases != 1 {
		t.Fatalf("worker releases=%d", allocator.releases)
	}
	if games, err := control.ListGames(t.Context(), session.Token, GameFilter{}); err != nil || len(games) != 0 {
		t.Fatalf("games after drain=%#v error=%v", games, err)
	}
	persisted, err := control.SelectedCharacter(t.Context(), session.Token)
	if err != nil || persisted.Character.ID != created.Character.ID || persisted.Revision != created.Revision+1 {
		t.Fatalf("character after drain=%#v error=%v", persisted, err)
	}
	if active, err := control.allocations.Active(t.Context()); err != nil || len(active) != 0 {
		t.Fatalf("allocations after drain=%#v error=%v", active, err)
	}
}

// TestControlPlaneOperatorDrainRetriesCleanupAfterDurableCommit ensures retrying
// worker cleanup cannot advance the character revision a second time.
func TestControlPlaneOperatorDrainRetriesCleanupAfterDurableCommit(t *testing.T) {
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(nil)
	config.Allocator = allocator
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := control.CreateAccount(t.Context(), "DrainRetryOwner", "long enough password"); err != nil {
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), "DrainRetryOwner", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	created, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "Retryhero", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := control.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Retry Drain", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	worker := allocator.workers[handoff.Game.Entry.GameID]
	worker.removeErr = errors.New("temporary worker control failure")
	first, err := control.DrainGame(t.Context(), handoff.Game.Entry.GameID)
	if err == nil || first.CommittedCharacters != 0 {
		t.Fatalf("first drain=%#v error=%v", first, err)
	}
	persisted, err := control.SelectedCharacter(t.Context(), session.Token)
	if err != nil || persisted.Revision != created.Revision+1 {
		t.Fatalf("character after partial drain=%#v error=%v", persisted, err)
	}
	worker.removeErr = nil
	retried, err := control.DrainGame(t.Context(), handoff.Game.Entry.GameID)
	if err != nil || retried.CommittedCharacters != 0 {
		t.Fatalf("retried drain=%#v error=%v", retried, err)
	}
	if persisted, err = control.SelectedCharacter(t.Context(), session.Token); err != nil ||
		persisted.Revision != created.Revision+1 {
		t.Fatalf("retry recommitted character=%#v error=%v", persisted, err)
	}
	if allocator.releases != 1 || len(worker.removed) != 2 {
		t.Fatalf("releases=%d removals=%#v", allocator.releases, worker.removed)
	}
}

// TestControlPlaneRollsBackDirectoryWorkerReservationAndLease checks that each
// failed allocation phase releases all resources reserved by earlier phases.
func TestControlPlaneRollsBackDirectoryWorkerReservationAndLease(t *testing.T) {
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(nil)
	config.Allocator = allocator
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	account, err := control.CreateAccount(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	record, err := control.CreateCharacter(
		t.Context(),
		session.Token,
		CreateCharacterRequest{Name: "Hero", Class: "Amazon"},
	)
	if err != nil {
		t.Fatal(err)
	}
	allocator.fail = errors.New("allocator unavailable")
	request := CreateGameRequest{Name: "Rollback", Difficulty: DifficultyNormal, Maximum: 8}
	if _, err := control.CreateGame(t.Context(), session.Token, request); err == nil {
		t.Fatal("allocator failure succeeded")
	}
	if games, _ := control.ListGames(t.Context(), session.Token, GameFilter{}); len(games) != 0 {
		t.Fatalf("failed allocation left directory entry: %#v", games)
	}
	allocator.fail, allocator.admitErr = nil, errors.New("admission rejected")
	if _, err := control.CreateGame(t.Context(), session.Token, request); err == nil {
		t.Fatal("admission failure succeeded")
	}
	if allocator.releases != 1 {
		t.Fatalf("worker releases = %d", allocator.releases)
	}
	if games, _ := control.ListGames(t.Context(), session.Token, GameFilter{}); len(games) != 0 {
		t.Fatalf("failed admission left directory entry: %#v", games)
	}
	_, lease, err := control.characters.Acquire(
		context.Background(),
		account.ID,
		record.Character.ID,
		"other",
		time.Minute,
	)
	if err != nil {
		t.Fatalf("failed admission leaked character lease: %v", err)
	}
	if err := control.characters.Release(t.Context(), lease); err != nil {
		t.Fatal(err)
	}
}

// TestControlPlaneRetriesDepartureCleanupWithoutRecommittingCharacter verifies
// a durable receipt resumes worker cleanup without duplicating its commit.
func TestControlPlaneRetriesDepartureCleanupWithoutRecommittingCharacter(t *testing.T) {
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(nil)
	config.Allocator = allocator
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := control.CreateAccount(t.Context(), "Alice", "long enough password"); err != nil {
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := control.CreateCharacter(
		t.Context(),
		session.Token,
		CreateCharacterRequest{Name: "Hero", Class: "Amazon"},
	); err != nil {
		t.Fatal(err)
	}
	created, err := control.CreateGame(t.Context(), session.Token, CreateGameRequest{Name: "Retry Leave",
		Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	worker := allocator.workers[created.Game.Entry.GameID]
	worker.removeErr = errors.New("temporary worker control failure")
	committed, err := control.LeaveGame(t.Context(), session.Token, created.Game.Entry.GameID)
	if err == nil || committed.Revision != 2 {
		t.Fatalf("first leave = %#v, %v", committed, err)
	}
	worker.removeErr = nil
	retried, err := control.LeaveGame(t.Context(), session.Token, created.Game.Entry.GameID)
	if err != nil || retried.Revision != committed.Revision {
		t.Fatalf("retried leave = %#v, %v", retried, err)
	}
	if len(worker.removed) != 2 || allocator.releases != 1 {
		t.Fatalf("worker removals=%#v releases=%d", worker.removed, allocator.releases)
	}
}

// TestControlPlaneCommitsWorkerReportedReconnectExpiry verifies trusted expiry
// notifications follow the canonical departure and audit path.
func TestControlPlaneCommitsWorkerReportedReconnectExpiry(t *testing.T) {
	capture := &capturedAudit{}
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(capture)
	config.Allocator = allocator
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := control.CreateAccount(t.Context(), "Alice", "long enough password"); err != nil {
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	createdCharacter, err := control.CreateCharacter(
		t.Context(),
		session.Token,
		CreateCharacterRequest{Name: "Hero", Class: "Amazon"},
	)
	if err != nil {
		t.Fatal(err)
	}
	created, err := control.CreateGame(t.Context(), session.Token, CreateGameRequest{Name: "Expired Client",
		Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	worker := allocator.workers[created.Game.Entry.GameID]
	worker.expired = []string{worker.admitted[0].PlayerID}
	if reconciled, err := control.ReconcileGames(t.Context()); err != nil || reconciled != 0 {
		t.Fatalf("reconcile expired membership = %d, %v", reconciled, err)
	}
	committed, err := control.SelectedCharacter(t.Context(), session.Token)
	if err != nil || committed.Revision != createdCharacter.Revision+1 {
		t.Fatalf("committed character = %#v, %v", committed, err)
	}
	if allocator.releases != 1 || len(worker.removed) != 1 {
		t.Fatalf("releases=%d removals=%#v", allocator.releases, worker.removed)
	}
	events := capture.snapshot()
	if len(events) == 0 ||
		events[len(events)-1].Operation != AuditGameLeave ||
		events[len(events)-1].Outcome != "success" ||
		events[len(events)-1].CharacterID != createdCharacter.Character.ID {
		t.Fatalf("last audit event = %#v", events)
	}
}

// TestControlPlaneReconcilesWorkerOnlyAfterConsecutiveHealthFailures protects
// healthy games from retirement after an isolated control-plane timeout.
func TestControlPlaneReconcilesWorkerOnlyAfterConsecutiveHealthFailures(t *testing.T) {
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(nil)
	config.Allocator = allocator
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	account, err := control.CreateAccount(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	record, err := control.CreateCharacter(
		t.Context(),
		session.Token,
		CreateCharacterRequest{Name: "Hero", Class: "Amazon"},
	)
	if err != nil {
		t.Fatal(err)
	}
	created, err := control.CreateGame(t.Context(), session.Token, CreateGameRequest{Name: "Health Check",
		Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	worker := allocator.workers[created.Game.Entry.GameID]
	worker.healthErr = errors.New("temporary control-plane timeout")
	for attempt := 0; attempt < workerFailureThreshold-1; attempt++ {
		if reconciled, err := control.ReconcileGames(t.Context()); err != nil || reconciled != 0 {
			t.Fatalf("transient reconciliation attempt %d = %d, %v", attempt+1, reconciled, err)
		}
	}
	worker.healthErr = nil
	if reconciled, err := control.ReconcileGames(t.Context()); err != nil || reconciled != 0 {
		t.Fatalf("healthy reset = %d, %v", reconciled, err)
	}
	worker.healthErr = errors.New("worker stopped")
	for attempt := 1; attempt <= workerFailureThreshold; attempt++ {
		reconciled, err := control.ReconcileGames(t.Context())
		if err != nil {
			t.Fatalf("reconciliation attempt %d: %v", attempt, err)
		}
		want := 0
		if attempt == workerFailureThreshold {
			want = 1
		}
		if reconciled != want {
			t.Fatalf("reconciliation attempt %d = %d, want %d", attempt, reconciled, want)
		}
	}
	if allocator.releases != 1 {
		t.Fatalf("worker releases = %d", allocator.releases)
	}
	if games, err := control.ListGames(t.Context(), session.Token, GameFilter{}); err != nil || len(games) != 0 {
		t.Fatalf("reconciled games = %#v, %v", games, err)
	}
	persisted, err := control.SelectedCharacter(t.Context(), session.Token)
	if err != nil || persisted.Revision != record.Revision {
		t.Fatalf("persisted character = %#v, %v", persisted, err)
	}
	_, lease, err := control.characters.Acquire(t.Context(), account.ID, record.Character.ID, "replacement", time.Minute)
	if err != nil {
		t.Fatalf("reconciled character remained leased: %v", err)
	}
	if err := control.characters.Release(t.Context(), lease); err != nil {
		t.Fatal(err)
	}
}

type restoringOrchestrationAllocator struct {
	*orchestrationAllocator
	recovery GameRecovery
	fences   int
}

// Restore captures the recovery payload before delegating replacement worker
// construction to the ordinary allocation fixture.
func (allocator *restoringOrchestrationAllocator) Restore(
	ctx context.Context,
	spec GameSpec,
	recovery GameRecovery,
) (WorkerAllocation, error) {
	allocator.recovery = recovery
	return allocator.Allocate(ctx, spec)
}

// Fence removes the exact fixture worker generation so startup restoration can
// assert that replacement never overlaps surviving authority.
func (allocator *restoringOrchestrationAllocator) Fence(_ context.Context, spec GameSpec) error {
	allocator.mu.Lock()
	defer allocator.mu.Unlock()
	if allocator.workers[spec.GameID] == nil {
		return ErrWorker
	}
	delete(allocator.workers, spec.GameID)
	allocator.fences++
	return nil
}

// TestControlPlaneRestoresFailedWorkerAndKeepsGameMembership verifies live
// recovery preserves directory, membership, and character revision state.
func TestControlPlaneRestoresFailedWorkerAndKeepsGameMembership(t *testing.T) {
	allocator := &restoringOrchestrationAllocator{orchestrationAllocator: newOrchestrationAllocator()}
	config := orchestratedControlConfig(nil)
	config.Allocator = allocator
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := control.CreateAccount(t.Context(), "RestoreAccount", "long enough password"); err != nil {
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), "RestoreAccount", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	character, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "RestoreHero", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := control.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Restored Game", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	gameID := handoff.Game.Entry.GameID
	if _, err := control.ReconcileGames(t.Context()); err != nil {
		t.Fatal(err)
	}
	failed := allocator.workers[gameID]
	failed.healthErr = errors.New("worker lost")
	for attempt := 1; attempt <= workerFailureThreshold; attempt++ {
		reconciled, err := control.ReconcileGames(t.Context())
		if err != nil {
			t.Fatalf("restore attempt %d: %v", attempt, err)
		}
		want := 0
		if attempt == workerFailureThreshold {
			want = 1
		}
		if reconciled != want {
			t.Fatalf("restore attempt %d reconciled=%d want=%d", attempt, reconciled, want)
		}
	}
	if allocator.recovery.Version != GameRecoveryVersion || len(allocator.recovery.PlayerIDs) != 1 ||
		allocator.recovery.Checkpoint.Checksum == "" {
		t.Fatalf("worker recovery = %#v", allocator.recovery)
	}
	if allocator.releases != 1 || allocator.workers[gameID] == failed {
		t.Fatalf("releases=%d replacement=%p failed=%p", allocator.releases, allocator.workers[gameID], failed)
	}
	games, err := control.ListGames(t.Context(), session.Token, GameFilter{})
	if err != nil || len(games) != 1 || games[0].GameID != gameID {
		t.Fatalf("games after restore=%#v error=%v", games, err)
	}
	reconnect, err := control.ReconnectGame(t.Context(), session.Token, gameID)
	if err != nil || reconnect.Assignment.Ticket == "" || reconnect.Assignment.GameID != gameID {
		t.Fatalf("reconnect after restore=%#v error=%v", reconnect, err)
	}
	persisted, err := control.SelectedCharacter(t.Context(), session.Token)
	if err != nil || persisted.Character.ID != character.Character.ID || persisted.Revision != character.Revision {
		t.Fatalf("character after restore=%#v error=%v", persisted, err)
	}
}

// TestControlPlaneFencesAndRestoresInterruptedAllocationOnStartup ensures a
// restart fences the surviving generation before installing its replacement.
func TestControlPlaneFencesAndRestoresInterruptedAllocationOnStartup(t *testing.T) {
	allocator := &restoringOrchestrationAllocator{orchestrationAllocator: newOrchestrationAllocator()}
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	characters, err := NewMemoryCharacters()
	if err != nil {
		t.Fatal(err)
	}
	memberships, err := NewMemoryMemberships(characters)
	if err != nil {
		t.Fatal(err)
	}
	config := orchestratedControlConfig(nil)
	config.Accounts, config.Characters, config.Games = accounts, characters, NewGameDirectory()
	config.Allocations = NewMemoryAllocations()
	config.Memberships = memberships
	config.Checkpoints = NewMemoryCheckpoints()
	config.Allocator = allocator
	first, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.CreateAccount(t.Context(), "StartupRestore", "long enough password"); err != nil {
		t.Fatal(err)
	}
	session, err := first.Authenticate(t.Context(), "StartupRestore", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "RestartHero", Class: "Amazon"}); err != nil {
		t.Fatal(err)
	}
	handoff, err := first.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Restart Game", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.ReconcileGames(t.Context()); err != nil {
		t.Fatal(err)
	}
	restarted, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := restarted.RecoverInterruptedGames(t.Context())
	if err != nil || recovered != 1 {
		t.Fatalf("startup recovery=%d error=%v", recovered, err)
	}
	if allocator.fences != 1 || allocator.recovery.Version != GameRecoveryVersion {
		t.Fatalf("fences=%d recovery=%#v", allocator.fences, allocator.recovery)
	}
	if active, err := config.Allocations.Active(t.Context()); err != nil ||
		len(active) != 1 ||
		active[0].GameID != handoff.Game.Entry.GameID {
		t.Fatalf("active allocations=%#v error=%v", active, err)
	}
	reconnect, err := restarted.ReconnectGame(t.Context(), session.Token, handoff.Game.Entry.GameID)
	if err != nil || reconnect.Assignment.Ticket == "" {
		t.Fatalf("reconnect=%#v error=%v", reconnect, err)
	}
}

type tamperingCheckpointRepository struct{ CheckpointRepository }

// Latest corrupts a copied checkpoint payload so recovery validation can prove
// tampered state never reaches the worker restorer.
func (repository tamperingCheckpointRepository) Latest(
	ctx context.Context,
	gameID string,
) (GameCheckpoint, error) {
	record, err := repository.CheckpointRepository.Latest(ctx, gameID)
	if err == nil {
		record.Checkpoint.State.Snapshot.Entities = []uint64{99}
	}
	return record, err
}

// TestControlPlaneStartupRejectsTamperedCheckpointAndReleasesCharacter verifies
// fail-closed recovery cleans every durable resource after integrity failure.
func TestControlPlaneStartupRejectsTamperedCheckpointAndReleasesCharacter(t *testing.T) {
	allocator := &restoringOrchestrationAllocator{orchestrationAllocator: newOrchestrationAllocator()}
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	characters, err := NewMemoryCharacters()
	if err != nil {
		t.Fatal(err)
	}
	memberships, err := NewMemoryMemberships(characters)
	if err != nil {
		t.Fatal(err)
	}
	checkpoints := NewMemoryCheckpoints()
	config := orchestratedControlConfig(nil)
	config.Accounts, config.Characters, config.Games = accounts, characters, NewGameDirectory()
	config.Allocations, config.Memberships, config.Checkpoints = NewMemoryAllocations(), memberships, checkpoints
	config.Allocator = allocator
	first, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	account, err := first.CreateAccount(t.Context(), "TamperAccount", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	session, err := first.Authenticate(t.Context(), "TamperAccount", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	character, err := first.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "TamperHero", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := first.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Tampered Restart", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.ReconcileGames(t.Context()); err != nil {
		t.Fatal(err)
	}

	config.Checkpoints = tamperingCheckpointRepository{CheckpointRepository: checkpoints}
	restarted, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	processed, err := restarted.RecoverInterruptedGames(t.Context())
	if err != nil || processed != 1 {
		t.Fatalf("tampered startup recovery processed=%d error=%v", processed, err)
	}
	if allocator.fences != 0 || allocator.recovery.Version != "" {
		t.Fatalf(
			"tampered checkpoint reached worker restoration: fences=%d recovery=%#v",
			allocator.fences,
			allocator.recovery,
		)
	}
	if active, err := config.Allocations.Active(t.Context()); err != nil || len(active) != 0 {
		t.Fatalf("active allocation after tamper=%#v error=%v", active, err)
	}
	if games, err := restarted.ListGames(t.Context(), session.Token, GameFilter{}); err != nil || len(games) != 0 {
		t.Fatalf("listed games after tamper=%#v error=%v", games, err)
	}
	record, lease, err := characters.Acquire(t.Context(), account.ID, character.Character.ID, "replacement", time.Minute)
	if err != nil || record.Character.ID != character.Character.ID {
		t.Fatalf("released character after tamper=%#v lease=%#v error=%v", record, lease, err)
	}
	if err := characters.Release(t.Context(), lease); err != nil {
		t.Fatal(err)
	}
	if allocation, err := config.Allocations.Get(t.Context(), handoff.Game.Entry.GameID); err != nil ||
		allocation.State != AllocationFailed || allocation.LastError == "" {
		t.Fatalf("tampered allocation=%#v error=%v", allocation, err)
	}
}

// TestControlPlanePersistsHealthyCheckpointAndRemovesItAfterCleanCompletion
// covers checkpoint cadence and clean-retirement deletion together.
func TestControlPlanePersistsHealthyCheckpointAndRemovesItAfterCleanCompletion(t *testing.T) {
	allocator := newOrchestrationAllocator()
	checkpoints := NewMemoryCheckpoints()
	config := orchestratedControlConfig(nil)
	config.Allocator, config.Checkpoints = allocator, checkpoints
	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := control.CreateAccount(t.Context(), "Alice", "long enough password"); err != nil {
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "Hero", Class: "Amazon"}); err != nil {
		t.Fatal(err)
	}
	handoff, err := control.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Checkpoint", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}
	if reconciled, err := control.ReconcileGames(t.Context()); err != nil || reconciled != 0 {
		t.Fatalf("reconcile=%d error=%v", reconciled, err)
	}
	stored, err := checkpoints.Latest(t.Context(), handoff.Game.Entry.GameID)
	if err != nil || stored.AllocationID == "" || stored.IdentityHash == "" || stored.Tick != 0 {
		t.Fatalf("stored checkpoint=%#v error=%v", stored, err)
	}
	if reconciled, err := control.ReconcileGames(t.Context()); err != nil || reconciled != 0 {
		t.Fatalf("second reconcile=%d error=%v", reconciled, err)
	}
	if calls := allocator.workers[handoff.Game.Entry.GameID].checkpoints; calls != 1 {
		t.Fatalf("checkpoint calls inside cadence = %d", calls)
	}
	if _, err := control.LeaveGame(t.Context(), session.Token, handoff.Game.Entry.GameID); err != nil {
		t.Fatal(err)
	}
	if _, err := checkpoints.Latest(t.Context(), handoff.Game.Entry.GameID); !errors.Is(err, ErrGameCheckpoint) {
		t.Fatalf("completed checkpoint error = %v", err)
	}
}

// TestControlPlaneFailsClosedInterruptedAllocationOnStartup verifies allocators
// without fencing and restore support release leases instead of guessing state.
func TestControlPlaneFailsClosedInterruptedAllocationOnStartup(t *testing.T) {
	allocator := newOrchestrationAllocator()
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	characters, err := NewMemoryCharacters()
	if err != nil {
		t.Fatal(err)
	}
	games := NewGameDirectory()
	allocations := NewMemoryAllocations()
	config := orchestratedControlConfig(nil)
	config.Accounts, config.Characters, config.Games = accounts, characters, games
	config.Allocations, config.Allocator = allocations, allocator
	first, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	account, err := first.CreateAccount(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	session, err := first.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	character, err := first.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "Hero", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := first.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Interrupted", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}

	restarted, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := restarted.RecoverInterruptedGames(t.Context())
	if err != nil || recovered != 1 {
		t.Fatalf("recovered = %d, %v", recovered, err)
	}
	if allocator.releases != 1 {
		t.Fatalf("worker releases = %d", allocator.releases)
	}
	if active, err := allocations.Active(t.Context()); err != nil || len(active) != 0 {
		t.Fatalf("active allocations = %#v, %v", active, err)
	}
	if record := allocations.records[handoff.Game.Entry.GameID]; record.State != AllocationFailed ||
		record.LastError == "" {
		t.Fatalf("recovered allocation = %#v", record)
	}
	if games, err := restarted.ListGames(t.Context(), session.Token, GameFilter{}); err != nil || len(games) != 0 {
		t.Fatalf("recovered directory = %#v, %v", games, err)
	}
	record, lease, err := characters.Acquire(t.Context(), account.ID, character.Character.ID, "replacement", time.Minute)
	if err != nil || record.Revision != character.Revision {
		t.Fatalf("released character = %#v lease=%#v error=%v", record, lease, err)
	}
}
