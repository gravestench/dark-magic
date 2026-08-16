package realm

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
)

func TestPostgresRepositoriesPreserveAccountCharacterAndLeaseContracts(t *testing.T) {
	connectionString := os.Getenv("DARK_MAGIC_TEST_POSTGRES")
	if connectionString == "" {
		t.Skip("DARK_MAGIC_TEST_POSTGRES is required")
	}
	store, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.Close)
	// Opening a second pool proves concurrent-safe schema initialization is
	// idempotent without introducing production migration history yet.
	second, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	second.Close()
	var schemaTables int
	if err := store.Pool.QueryRow(t.Context(), `SELECT count(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = ANY($1)`, []string{
		"realm_accounts", "realm_sessions", "realm_characters", "realm_character_leases",
		"realm_games", "realm_allocations", "realm_memberships", "realm_game_reservations",
		"realm_game_players", "realm_game_checkpoints", "realm_mail_outbox", "realm_account_challenges", "realm_audit_events",
	}).Scan(&schemaTables); err != nil || schemaTables != 13 {
		t.Fatalf("Realm schema tables = %d, %v", schemaTables, err)
	}
	var nativeAuthorizationTable *string
	if err := store.Pool.QueryRow(t.Context(), `SELECT to_regclass('realm_native_authorizations')::text`).
		Scan(&nativeAuthorizationTable); err != nil || nativeAuthorizationTable != nil {
		t.Fatalf("obsolete native authorization table = %v, %v", nativeAuthorizationTable, err)
	}

	suffix := uuid.New().String()[:8]
	accountName := "Alice_" + suffix
	password := "long enough password"
	account, err := store.Accounts.Create(t.Context(), accountName, password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Accounts.Create(t.Context(), "alice_"+suffix, password); !errors.Is(err, ErrAccountExists) {
		t.Fatalf("normalized duplicate account error = %v", err)
	}
	if _, err := store.Accounts.Authenticate(t.Context(), accountName, "wrong password"); !errors.Is(err, ErrAccountCredentials) {
		t.Fatalf("bad password error = %v", err)
	}
	session, err := store.Accounts.Authenticate(t.Context(), accountName, password)
	if err != nil {
		t.Fatal(err)
	}
	principal, err := store.Accounts.Authorize(t.Context(), session.Token)
	if err != nil || principal.AccountID() != account.ID || principal.SessionID() != session.ID {
		t.Fatalf("principal = %#v, %v", principal, err)
	}
	digest := sha256.Sum256([]byte(session.Token))
	var storedDigest []byte
	if err := store.Pool.QueryRow(t.Context(), `SELECT token_digest FROM realm_sessions WHERE id = $1`, session.ID).Scan(&storedDigest); err != nil {
		t.Fatal(err)
	}
	if string(storedDigest) != string(digest[:]) || string(storedDigest) == session.Token {
		t.Fatal("PostgreSQL session did not retain exactly the token digest")
	}

	control, err := NewControlPlane(ControlPlaneConfig{Accounts: store.Accounts, Characters: store.Characters})
	if err != nil {
		t.Fatal(err)
	}
	created, err := control.CreateCharacter(t.Context(), session.Token, CreateCharacterRequest{Name: "Heroine", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}
	muleSession, err := store.Accounts.Authenticate(t.Context(), accountName, password)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Accounts.SelectCharacter(t.Context(), muleSession.Token, created.Character.ID); !errors.Is(err, ErrCharacterOnline) {
		t.Fatalf("duplicate PostgreSQL character claim error = %v", err)
	}
	mule, err := control.CreateCharacter(t.Context(), muleSession.Token,
		CreateCharacterRequest{Name: "MuleHero", Class: "Barbarian"})
	if err != nil || mule.Character.ID == created.Character.ID {
		t.Fatalf("same-account mule character = %#v, %v", mule, err)
	}
	if err := store.Accounts.Logout(t.Context(), muleSession.Token); err != nil {
		t.Fatal(err)
	}
	selected, err := control.SelectedCharacter(t.Context(), session.Token)
	if err != nil || selected.Character.ID != created.Character.ID {
		t.Fatalf("selected = %#v, %v", selected, err)
	}
	compatibility := gamesession.DurableCompatibility{CharacterID: created.Character.ID, ModID: "d2legacy",
		ContractVersion: "v1", IdentityHash: "identity"}
	record, lease, err := store.Characters.Acquire(t.Context(), account.ID, created.Character.ID, "game-1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Characters.Acquire(t.Context(), account.ID, created.Character.ID, "game-2", time.Minute); !errors.Is(err, ErrCharacterLeased) {
		t.Fatalf("duplicate lease error = %v", err)
	}
	bound, err := store.Characters.BindCompatibility(t.Context(), lease, compatibility)
	if err != nil || bound.Compatibility != compatibility {
		t.Fatalf("bound = %#v, %v", bound, err)
	}
	renewed, err := store.Characters.Renew(t.Context(), lease, 2*time.Minute)
	if err != nil || !renewed.ExpiresAt.After(lease.ExpiresAt) {
		t.Fatalf("renewed = %#v, %v", renewed, err)
	}
	record.Character.Level++
	committed, err := store.Characters.Commit(t.Context(), renewed, record.Character)
	if err != nil || committed.Revision != record.Revision+1 || committed.Character.Level != record.Character.Level {
		t.Fatalf("committed = %#v, %v", committed, err)
	}
	if _, err := store.Characters.Commit(t.Context(), renewed, record.Character); !errors.Is(err, ErrCharacterCommit) {
		t.Fatalf("replayed commit error = %v", err)
	}

	// Exactly one concurrent transaction can acquire the now-unleased character.
	var acquired atomic.Int32
	var winner CharacterLease
	var winnerMu sync.Mutex
	var group sync.WaitGroup
	for index := 0; index < 8; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			_, candidate, err := store.Characters.Acquire(context.Background(), account.ID, created.Character.ID,
				fmt.Sprintf("concurrent-%d", index), time.Minute)
			if err == nil {
				acquired.Add(1)
				winnerMu.Lock()
				winner = candidate
				winnerMu.Unlock()
			} else if !errors.Is(err, ErrCharacterLeased) {
				t.Errorf("concurrent acquire %d: %v", index, err)
			}
		}(index)
	}
	group.Wait()
	if acquired.Load() != 1 {
		t.Fatalf("successful concurrent acquisitions = %d", acquired.Load())
	}
	if err := store.Characters.Release(t.Context(), winner); err != nil {
		t.Fatal(err)
	}
	if err := control.DeleteCharacter(t.Context(), session.Token, created.Character.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.Accounts.Logout(t.Context(), session.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Accounts.Authorize(t.Context(), session.Token); !errors.Is(err, ErrRealmSession) {
		t.Fatalf("logged-out authorization error = %v", err)
	}
}

func TestPostgresGameDirectoryPersistsAndSerializesCapacity(t *testing.T) {
	connectionString := os.Getenv("DARK_MAGIC_TEST_POSTGRES")
	if connectionString == "" {
		t.Skip("DARK_MAGIC_TEST_POSTGRES is required")
	}
	store, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	suffix := uuid.New().String()[:8]
	account, err := store.Accounts.Create(t.Context(), "Directory_"+suffix, "long enough password")
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	session, err := store.Accounts.Authenticate(t.Context(), account.Name, "long enough password")
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	control, err := NewControlPlane(ControlPlaneConfig{Accounts: store.Accounts, Characters: store.Characters,
		Games: store.Games})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	characterA, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "Amazonia", Class: "Amazon", Expansion: true})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	characterB, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "Barbaria", Class: "Barbarian", Expansion: true})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	principal, err := store.Accounts.Authorize(t.Context(), session.Token)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	name := "Durable Game " + suffix
	created, err := store.Games.Create(t.Context(), principal, CreateGameRequest{
		Name: name, Difficulty: DifficultyNormal, Maximum: 1, CharacterDifference: 4, Expansion: true,
	})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	if created.Entry.CharacterDifference != 4 {
		store.Close()
		t.Fatalf("created character difference = %d", created.Entry.CharacterDifference)
	}
	allocationID := "allocation-" + suffix
	if _, err := store.Allocations.Request(t.Context(), created.Entry.GameID, allocationID); err != nil {
		store.Close()
		t.Fatal(err)
	}
	if repeated, err := store.Allocations.Request(t.Context(), created.Entry.GameID, allocationID); err != nil || repeated.AllocationID != allocationID {
		store.Close()
		t.Fatalf("repeated PostgreSQL allocation request = %#v, %v", repeated, err)
	}
	if _, err := store.Allocations.Request(t.Context(), created.Entry.GameID, allocationID+"-other"); !errors.Is(err, ErrGameExists) {
		store.Close()
		t.Fatalf("conflicting PostgreSQL allocation request = %v", err)
	}
	endpoint := GameEndpoint{Address: "127.0.0.1:4000", TLSFingerprint: "sha256:worker"}
	if _, err := store.Allocations.Ready(t.Context(), created.Entry.GameID, endpoint, orchestrationIdentity()); err != nil {
		store.Close()
		t.Fatal(err)
	}
	if repeated, err := store.Allocations.Ready(t.Context(), created.Entry.GameID, endpoint, orchestrationIdentity()); err != nil || repeated.State != AllocationReady {
		store.Close()
		t.Fatalf("repeated PostgreSQL allocation ready = %#v, %v", repeated, err)
	}
	replacementEndpoint := GameEndpoint{Address: "127.0.0.1:4001", TLSFingerprint: "sha256:replacement"}
	if restored, err := store.Allocations.RestoreReady(t.Context(), created.Entry.GameID, allocationID,
		replacementEndpoint, orchestrationIdentity()); err != nil || restored.Endpoint != replacementEndpoint {
		store.Close()
		t.Fatalf("restored PostgreSQL allocation ready = %#v, %v", restored, err)
	}
	endpoint = replacementEndpoint
	store.Close()

	restarted, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	persisted, err := restarted.Games.Detail(t.Context(), strings.ToLower(name))
	if err != nil || persisted.Entry.GameID != created.Entry.GameID || persisted.Entry.Revision != 1 {
		t.Fatalf("persisted game = %#v, %v", persisted, err)
	}
	activeAllocations, err := restarted.Allocations.Active(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	var persistedAllocation AllocationRecord
	for _, record := range activeAllocations {
		if record.GameID == created.Entry.GameID {
			persistedAllocation = record
			break
		}
	}
	if persistedAllocation.AllocationID != allocationID || persistedAllocation.State != AllocationReady ||
		persistedAllocation.Endpoint != endpoint || persistedAllocation.LastHealthyAt == nil {
		t.Fatalf("persisted allocation = %#v", persistedAllocation)
	}

	players := []GamePlayer{
		{CharacterID: characterA.Character.ID, Name: characterA.Character.Name, Class: characterA.Character.Class, Level: 1},
		{CharacterID: characterB.Character.ID, Name: characterB.Character.Name, Class: characterB.Character.Class, Level: 1},
	}
	reservations := make(chan GamePlayerReservation, len(players))
	errorsSeen := make(chan error, len(players))
	var group sync.WaitGroup
	for _, player := range players {
		player := player
		group.Add(1)
		go func() {
			defer group.Done()
			reservation, reserveErr := restarted.Games.ReservePlayer(context.Background(), created.Entry.GameID, player)
			if reserveErr != nil {
				errorsSeen <- reserveErr
				return
			}
			reservations <- reservation
		}()
	}
	group.Wait()
	close(reservations)
	close(errorsSeen)
	if len(reservations) != 1 || len(errorsSeen) != 1 {
		t.Fatalf("reservations=%d errors=%d, want one capacity winner", len(reservations), len(errorsSeen))
	}
	for reserveErr := range errorsSeen {
		if !errors.Is(reserveErr, ErrGameFull) {
			t.Fatalf("capacity loser error = %v", reserveErr)
		}
	}
	winner := <-reservations
	digest := sha256.Sum256([]byte(winner.Token))
	var storedDigest []byte
	if err := restarted.Pool.QueryRow(t.Context(), `SELECT token_digest FROM realm_game_reservations WHERE game_id = $1`,
		created.Entry.GameID).Scan(&storedDigest); err != nil {
		t.Fatal(err)
	}
	if string(storedDigest) != string(digest[:]) || string(storedDigest) == winner.Token {
		t.Fatal("PostgreSQL game reservation did not retain exactly the token digest")
	}
	committed, err := restarted.Games.CommitPlayer(t.Context(), winner)
	if err != nil || committed.Entry.Players != 1 || committed.Entry.Revision != 2 || len(committed.Players) != 1 {
		t.Fatalf("committed game = %#v, %v", committed, err)
	}
	if _, _, err := restarted.Characters.Acquire(t.Context(), account.ID, characterA.Character.ID,
		created.Entry.GameID, time.Minute); err != nil {
		t.Fatal(err)
	}
	recoveryControl, err := NewControlPlane(ControlPlaneConfig{Accounts: restarted.Accounts,
		Characters: restarted.Characters, Games: restarted.Games, Allocations: restarted.Allocations})
	if err != nil {
		t.Fatal(err)
	}
	if recovered, err := recoveryControl.RecoverInterruptedGames(t.Context()); err != nil || recovered != 1 {
		t.Fatalf("PostgreSQL restart recovery = %d, %v", recovered, err)
	}
	if _, err := restarted.Games.Detail(t.Context(), created.Entry.GameID); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("completed game detail error = %v", err)
	}
	if _, lease, err := restarted.Characters.Acquire(t.Context(), account.ID, characterA.Character.ID,
		"replacement-"+suffix, time.Minute); err != nil {
		t.Fatalf("interrupted game lease remained active: %v", err)
	} else if err := restarted.Characters.Release(t.Context(), lease); err != nil {
		t.Fatal(err)
	}
	var allocationState, allocationError string
	if err := restarted.Pool.QueryRow(t.Context(), `SELECT state, COALESCE(last_error, '') FROM realm_allocations WHERE game_id = $1`,
		created.Entry.GameID).Scan(&allocationState, &allocationError); err != nil ||
		allocationState != string(AllocationFailed) || allocationError == "" {
		t.Fatalf("recovered allocation state=%q error=%q query=%v", allocationState, allocationError, err)
	}
	if err := restarted.Allocations.Fail(t.Context(), created.Entry.GameID, errors.New(allocationError)); err != nil {
		t.Fatalf("repeated PostgreSQL allocation failure = %v", err)
	}
	if _, err := restarted.Games.Create(t.Context(), principal, CreateGameRequest{
		Name: name, Difficulty: DifficultyNormal, Maximum: 8, Expansion: true,
	}); err != nil {
		t.Fatalf("reuse completed game name: %v", err)
	}
}

func TestPostgresGameDirectoryPersistsDrainingAdmissionBarrier(t *testing.T) {
	connectionString := os.Getenv("DARK_MAGIC_TEST_POSTGRES")
	if connectionString == "" {
		t.Skip("DARK_MAGIC_TEST_POSTGRES is required")
	}
	store, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	suffix := uuid.New().String()[:8]
	account, err := store.Accounts.Create(t.Context(), "Drain_"+suffix, "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.Accounts.Authenticate(t.Context(), account.Name, "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	principal, err := store.Accounts.Authorize(t.Context(), session.Token)
	if err != nil {
		t.Fatal(err)
	}
	game, err := store.Games.Create(t.Context(), principal,
		CreateGameRequest{Name: "Drain " + suffix, Difficulty: DifficultyNormal, Maximum: 2})
	if err != nil {
		t.Fatal(err)
	}
	control, err := NewControlPlane(ControlPlaneConfig{Accounts: store.Accounts, Characters: store.Characters, Games: store.Games})
	if err != nil {
		t.Fatal(err)
	}
	character, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "Drainhero", Class: "Amazon"})
	if err != nil {
		t.Fatal(err)
	}
	reservation, err := store.Games.ReservePlayer(t.Context(), game.Entry.GameID,
		GamePlayer{CharacterID: character.Character.ID, Name: character.Character.Name, Class: character.Character.Class, Level: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Games.CommitPlayer(t.Context(), reservation); err != nil {
		t.Fatal(err)
	}
	if err := store.Games.BeginDrain(t.Context(), game.Entry.GameID); err != nil {
		t.Fatal(err)
	}
	if err := store.Games.BeginDrain(t.Context(), game.Entry.GameID); err != nil {
		t.Fatalf("retry drain: %v", err)
	}
	var state string
	if err := store.Pool.QueryRow(t.Context(), `SELECT state FROM realm_games WHERE id = $1`, game.Entry.GameID).Scan(&state); err != nil || state != drainingRealmGameState {
		t.Fatalf("state=%q error=%v", state, err)
	}
	if listed, err := store.Games.List(t.Context(), GameFilter{}); err != nil {
		t.Fatal(err)
	} else {
		for _, entry := range listed {
			if entry.GameID == game.Entry.GameID {
				t.Fatal("draining PostgreSQL game remains listed")
			}
		}
	}
	if _, err := store.Games.ResolveJoin(t.Context(), game.Entry.GameID, ""); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("draining join error=%v", err)
	}
	departed, err := store.Games.RemovePlayer(t.Context(), game.Entry.GameID, character.Character.ID)
	if err != nil || departed.Entry.Players != 0 {
		t.Fatalf("departure=%#v error=%v", departed, err)
	}
	if err := store.Games.Remove(t.Context(), game.Entry.GameID); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresMembershipDepartureIsAtomicRetryableAndDurable(t *testing.T) {
	connectionString := os.Getenv("DARK_MAGIC_TEST_POSTGRES")
	if connectionString == "" {
		t.Skip("DARK_MAGIC_TEST_POSTGRES is required")
	}
	store, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	suffix := uuid.New().String()[:8]
	account, err := store.Accounts.Create(t.Context(), "Membership_"+suffix, "long enough password")
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	session, err := store.Accounts.Authenticate(t.Context(), account.Name, "long enough password")
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	control, err := NewControlPlane(ControlPlaneConfig{Accounts: store.Accounts, Characters: store.Characters,
		Games: store.Games, Memberships: store.Memberships})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	character, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "Member", Class: "Amazon", Expansion: true})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	principal, err := store.Accounts.Authorize(t.Context(), session.Token)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	game, err := store.Games.Create(t.Context(), principal, CreateGameRequest{
		Name: "Membership Game " + suffix, Difficulty: DifficultyNormal, Maximum: 8, Expansion: true})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	baseline, lease, err := store.Characters.Acquire(t.Context(), account.ID, character.Character.ID,
		game.Entry.GameID, time.Minute)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	membership := MembershipRecord{GameID: game.Entry.GameID, PlayerID: "player-" + suffix,
		AccountID: account.ID, Baseline: baseline, Lease: lease, State: MembershipActive}
	if err := store.Memberships.Admit(t.Context(), membership); err != nil {
		store.Close()
		t.Fatal(err)
	}
	canonical := cloneCharacter(baseline.Character)
	canonical.Level = 2
	receipts := make(chan departureReceipt, 2)
	errorsSeen := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			receipt, departErr := store.Memberships.Depart(context.Background(), membership, canonical)
			receipts <- receipt
			errorsSeen <- departErr
		}()
	}
	group.Wait()
	close(receipts)
	close(errorsSeen)
	for departErr := range errorsSeen {
		if departErr != nil {
			store.Close()
			t.Fatal(departErr)
		}
	}
	for receipt := range receipts {
		if receipt.Record.Revision != 2 || receipt.Record.Character.Level != 2 {
			store.Close()
			t.Fatalf("PostgreSQL membership receipt = %#v", receipt)
		}
	}
	store.Close()

	restarted, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	persisted, err := restarted.Memberships.ByAccount(t.Context(), game.Entry.GameID, account.ID)
	if err != nil || persisted.State != MembershipDeparted || persisted.Departure == nil ||
		persisted.Departure.Record.Revision != 2 {
		t.Fatalf("persisted PostgreSQL membership = %#v, %v", persisted, err)
	}
	if repeated, err := restarted.Memberships.Depart(t.Context(), membership, canonical); err != nil || repeated.Record.Revision != 2 {
		t.Fatalf("restarted PostgreSQL departure retry = %#v, %v", repeated, err)
	}
	if err := restarted.Memberships.AbandonGame(t.Context(), game.Entry.GameID); err != nil {
		t.Fatal(err)
	}
	abandoned, err := restarted.Memberships.ByPlayer(t.Context(), game.Entry.GameID, membership.PlayerID)
	if err != nil || abandoned.Departure == nil || !abandoned.Departure.WorkerRemoved {
		t.Fatalf("abandoned PostgreSQL departure = %#v, %v", abandoned, err)
	}
	if _, err := restarted.Memberships.MarkWorkerRemoved(t.Context(), game.Entry.GameID, membership.PlayerID); err != nil {
		t.Fatal(err)
	}
	if repeated, err := restarted.Memberships.MarkWorkerRemoved(t.Context(), game.Entry.GameID, membership.PlayerID); err != nil || !repeated.WorkerRemoved {
		t.Fatalf("repeated PostgreSQL worker removal = %#v, %v", repeated, err)
	}
	committed, err := restarted.Characters.Get(t.Context(), account.ID, character.Character.ID)
	if err != nil || committed.Revision != 2 || committed.Character.Level != 2 {
		t.Fatalf("atomically committed PostgreSQL character = %#v, %v", committed, err)
	}
}

func TestPostgresMembershipsAllowDifferentCharactersFromSameAccount(t *testing.T) {
	connectionString := os.Getenv("DARK_MAGIC_TEST_POSTGRES")
	if connectionString == "" {
		t.Skip("DARK_MAGIC_TEST_POSTGRES is required")
	}
	store, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.Close)
	suffix := uuid.New().String()[:8]
	nameSuffix := make([]byte, 8)
	randomName := uuid.New()
	for index := range nameSuffix {
		nameSuffix[index] = 'a' + randomName[index]%26
	}
	account, err := store.Accounts.Create(t.Context(), "Muling_"+suffix, "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	firstSession, err := store.Accounts.Authenticate(t.Context(), account.Name, "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	secondSession, err := store.Accounts.Authenticate(t.Context(), account.Name, "long enough password")
	if err != nil {
		t.Fatal(err)
	}
	control, err := NewControlPlane(ControlPlaneConfig{Accounts: store.Accounts, Characters: store.Characters,
		Games: store.Games, Memberships: store.Memberships})
	if err != nil {
		t.Fatal(err)
	}
	first, err := control.CreateCharacter(t.Context(), firstSession.Token,
		CreateCharacterRequest{Name: "First" + string(nameSuffix), Class: "Amazon", Expansion: true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := control.CreateCharacter(t.Context(), secondSession.Token,
		CreateCharacterRequest{Name: "Second" + string(nameSuffix), Class: "Barbarian", Expansion: true})
	if err != nil {
		t.Fatal(err)
	}
	principal, err := store.Accounts.Authorize(t.Context(), firstSession.Token)
	if err != nil {
		t.Fatal(err)
	}
	game, err := store.Games.Create(t.Context(), principal, CreateGameRequest{
		Name: "Muling Game " + suffix, Difficulty: DifficultyNormal, Maximum: 8, Expansion: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for index, character := range []CharacterRecord{first, second} {
		baseline, lease, err := store.Characters.Acquire(t.Context(), account.ID, character.Character.ID,
			game.Entry.GameID, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.Memberships.Admit(t.Context(), MembershipRecord{
			GameID: game.Entry.GameID, PlayerID: fmt.Sprintf("player-%d-%s", index+1, suffix),
			AccountID: account.ID, Baseline: baseline, Lease: lease, State: MembershipActive,
		}); err != nil {
			t.Fatalf("admit character %q: %v", character.Character.ID, err)
		}
	}
	players, err := store.Memberships.ActivePlayerIDs(t.Context(), game.Entry.GameID)
	if err != nil || len(players) != 2 {
		t.Fatalf("same-account PostgreSQL memberships = %#v, %v", players, err)
	}
}

func TestPostgresMembershipResumeRotatesLostLeaseTokenAtomically(t *testing.T) {
	connectionString := os.Getenv("DARK_MAGIC_TEST_POSTGRES")
	if connectionString == "" {
		t.Skip("DARK_MAGIC_TEST_POSTGRES is required")
	}
	store, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	suffix := uuid.New().String()[:8]
	account, err := store.Accounts.Create(t.Context(), "Resume_"+suffix, "long enough password")
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	session, err := store.Accounts.Authenticate(t.Context(), account.Name, "long enough password")
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	control, err := NewControlPlane(ControlPlaneConfig{Accounts: store.Accounts, Characters: store.Characters,
		Games: store.Games, Memberships: store.Memberships})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	character, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "ResumeHero", Class: "Amazon", Expansion: true})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	principal, err := store.Accounts.Authorize(t.Context(), session.Token)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	game, err := store.Games.Create(t.Context(), principal, CreateGameRequest{
		Name: "Resume Game " + suffix, Difficulty: DifficultyNormal, Maximum: 8, Expansion: true})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	baseline, lostLease, err := store.Characters.Acquire(t.Context(), account.ID, character.Character.ID,
		game.Entry.GameID, time.Minute)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	membership := MembershipRecord{GameID: game.Entry.GameID, PlayerID: "player-" + suffix,
		AccountID: account.ID, Baseline: baseline, Lease: lostLease, State: MembershipActive}
	if err := store.Memberships.Admit(t.Context(), membership); err != nil {
		store.Close()
		t.Fatal(err)
	}
	lostDigest := sha256.Sum256([]byte(lostLease.Token))
	store.Close()

	restarted, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	resumed, err := restarted.Memberships.ResumeGame(t.Context(), game.Entry.GameID, 2*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(resumed) != 1 || resumed[0].PlayerID != membership.PlayerID || resumed[0].Lease.Token == "" ||
		resumed[0].Lease.Token == lostLease.Token || !resumed[0].Lease.ExpiresAt.After(time.Now().Add(time.Minute)) {
		t.Fatalf("resumed memberships = %#v", resumed)
	}
	newDigest := sha256.Sum256([]byte(resumed[0].Lease.Token))
	var storedDigest []byte
	if err := restarted.Pool.QueryRow(t.Context(), `SELECT token_digest FROM realm_character_leases WHERE character_id = $1`,
		character.Character.ID).Scan(&storedDigest); err != nil {
		t.Fatal(err)
	}
	if string(storedDigest) != string(newDigest[:]) || string(storedDigest) == string(lostDigest[:]) {
		t.Fatal("resumed membership did not rotate the durable lease digest")
	}
	if _, err := restarted.Characters.Renew(t.Context(), lostLease, time.Minute); !errors.Is(err, ErrLease) {
		t.Fatalf("lost lease renewal error = %v", err)
	}
	if _, err := restarted.Characters.Renew(t.Context(), resumed[0].Lease, time.Minute); err != nil {
		t.Fatalf("resumed lease renewal: %v", err)
	}
}

func TestPostgresVerifiedAccountRecoveryAndTransactionalMail(t *testing.T) {
	connectionString := os.Getenv("DARK_MAGIC_TEST_POSTGRES")
	if connectionString == "" {
		t.Skip("DARK_MAGIC_TEST_POSTGRES is required")
	}
	store, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.Close)
	if err := store.Accounts.SetAccountBaseURL("https://accounts.dark-magic.test"); err != nil {
		t.Fatal(err)
	}
	suffix := uuid.New().String()[:8]
	name, email := "Verified_"+suffix, "owner-"+suffix+"@example.test"
	oldPassword, newPassword := "first long password", "replacement long password"
	account, err := store.Accounts.Signup(t.Context(), SignupRequest{Name: name, Email: email, Password: oldPassword})
	if err != nil || account.EmailVerified {
		t.Fatalf("signup account=%#v error=%v", account, err)
	}
	if _, err := store.Accounts.Signup(t.Context(), SignupRequest{Name: "Other_" + suffix,
		Email: strings.ToUpper(email), Password: oldPassword}); !errors.Is(err, ErrAccountExists) {
		t.Fatalf("normalized email duplicate error = %v", err)
	}
	if _, err := store.Accounts.Authenticate(t.Context(), name, oldPassword); !errors.Is(err, ErrAccountUnverified) {
		t.Fatalf("unverified login error = %v", err)
	}
	verification, err := store.Mail.ClaimMail(t.Context(), "test-mailer", time.Minute)
	if err != nil || verification.Kind != "verify_email" || verification.Recipient != email {
		t.Fatalf("verification mail=%#v error=%v", verification, err)
	}
	verificationToken := accountMailToken(t, verification, "verification_url")
	verified, err := store.Accounts.VerifyEmail(t.Context(), verificationToken)
	if err != nil || !verified.EmailVerified || verified.ID != account.ID {
		t.Fatalf("verified account=%#v error=%v", verified, err)
	}
	if _, err := store.Accounts.VerifyEmail(t.Context(), verificationToken); !errors.Is(err, ErrAccountChallenge) {
		t.Fatalf("replayed verification error = %v", err)
	}
	if err := store.Mail.CompleteMail(t.Context(), "test-mailer", verification.ID); err != nil {
		t.Fatal(err)
	}
	session, err := store.Accounts.Authenticate(t.Context(), name, oldPassword)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Accounts.BeginPasswordRecovery(t.Context(), "missing-"+suffix+"@example.test"); err != nil {
		t.Fatalf("non-enumerating missing recovery error = %v", err)
	}
	if _, err := store.Mail.ClaimMail(t.Context(), "test-mailer", time.Minute); !errors.Is(err, ErrMailUnavailable) {
		t.Fatalf("missing account produced mail: %v", err)
	}
	if err := store.Accounts.BeginPasswordRecovery(t.Context(), strings.ToUpper(email)); err != nil {
		t.Fatal(err)
	}
	recovery, err := store.Mail.ClaimMail(t.Context(), "test-mailer", time.Minute)
	if err != nil || recovery.Kind != "reset_password" {
		t.Fatalf("recovery mail=%#v error=%v", recovery, err)
	}
	recoveryToken := accountMailToken(t, recovery, "recovery_url")
	if err := store.Accounts.CompletePasswordRecovery(t.Context(), recoveryToken, newPassword); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Accounts.Authorize(t.Context(), session.Token); !errors.Is(err, ErrRealmSession) {
		t.Fatalf("pre-recovery session error = %v", err)
	}
	if _, err := store.Accounts.Authenticate(t.Context(), name, oldPassword); !errors.Is(err, ErrAccountCredentials) {
		t.Fatalf("old password error = %v", err)
	}
	if _, err := store.Accounts.Authenticate(t.Context(), name, newPassword); err != nil {
		t.Fatalf("new password error = %v", err)
	}
	if err := store.Accounts.CompletePasswordRecovery(t.Context(), recoveryToken, newPassword); !errors.Is(err, ErrAccountChallenge) {
		t.Fatalf("replayed recovery error = %v", err)
	}
}

func accountMailToken(t *testing.T, job MailJob, field string) string {
	t.Helper()
	value, ok := job.Payload[field].(string)
	if !ok {
		t.Fatalf("mail field %q = %#v", field, job.Payload[field])
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Query().Get("token") == "" {
		t.Fatalf("mail URL %q error=%v", value, err)
	}
	return parsed.Query().Get("token")
}

func TestPostgresControlPlanePersistsCanonicalLeaveAndSessionAcrossRestart(t *testing.T) {
	connectionString := os.Getenv("DARK_MAGIC_TEST_POSTGRES")
	if connectionString == "" {
		t.Skip("DARK_MAGIC_TEST_POSTGRES is required")
	}
	store, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	suffix := uuid.New().String()[:8]
	accountName := "Restart_" + suffix
	password := "long enough password"
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(store.Audit)
	config.Accounts, config.Characters, config.Allocator = store.Accounts, store.Characters, allocator
	config.Games, config.Allocations, config.Memberships = store.Games, store.Allocations, store.Memberships
	control, err := NewControlPlane(config)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	if _, err := control.CreateAccount(t.Context(), accountName, password); err != nil {
		store.Close()
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), accountName, password)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	created, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "RestartHero", Class: "Amazon"})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	handoff, err := control.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Persistent Leave " + suffix, Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	committed, err := control.LeaveGame(t.Context(), session.Token, handoff.Game.Entry.GameID)
	if err != nil || committed.Revision != created.Revision+1 {
		store.Close()
		t.Fatalf("committed = %#v, %v", committed, err)
	}
	var auditOperation, auditCharacterID string
	if err := store.Pool.QueryRow(t.Context(), `SELECT event->>'operation', event->>'character_id'
		FROM realm_audit_events WHERE event->>'account_id' = $1 ORDER BY id DESC LIMIT 1`, session.Account.ID).
		Scan(&auditOperation, &auditCharacterID); err != nil || auditOperation != AuditGameLeave || auditCharacterID != created.Character.ID {
		store.Close()
		t.Fatalf("audit operation=%q character=%q error=%v", auditOperation, auditCharacterID, err)
	}
	var leases int
	if err := store.Pool.QueryRow(t.Context(), `SELECT count(*) FROM realm_character_leases WHERE character_id = $1`,
		created.Character.ID).Scan(&leases); err != nil || leases != 0 {
		store.Close()
		t.Fatalf("leases = %d, %v", leases, err)
	}
	store.Close()

	restarted, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	principal, err := restarted.Accounts.Authorize(t.Context(), session.Token)
	if err != nil || principal.AccountID() != session.Account.ID {
		t.Fatalf("restarted principal = %#v, %v", principal, err)
	}
	records, err := restarted.Characters.List(t.Context(), session.Account.ID)
	if err != nil || len(records) != 1 || records[0].Revision != committed.Revision ||
		records[0].Character.ID != created.Character.ID {
		t.Fatalf("restarted characters = %#v, %v", records, err)
	}
	restartedControl, err := NewControlPlane(ControlPlaneConfig{Accounts: restarted.Accounts,
		Characters: restarted.Characters, Games: restarted.Games, Allocations: restarted.Allocations,
		Memberships: restarted.Memberships})
	if err != nil {
		t.Fatal(err)
	}
	retried, err := restartedControl.LeaveGame(t.Context(), session.Token, handoff.Game.Entry.GameID)
	if err != nil || retried.Revision != committed.Revision {
		t.Fatalf("restarted departure retry = %#v, %v", retried, err)
	}
}

func TestPostgresGameCheckpointSurvivesRestartRejectsRegressionAndDetectsTampering(t *testing.T) {
	connectionString := os.Getenv("DARK_MAGIC_TEST_POSTGRES")
	if connectionString == "" {
		t.Skip("DARK_MAGIC_TEST_POSTGRES is required")
	}
	store, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	suffix := uuid.New().String()[:8]
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(store.Audit)
	config.Accounts, config.Characters, config.Allocator = store.Accounts, store.Characters, allocator
	config.Games, config.Allocations, config.Memberships = store.Games, store.Allocations, store.Memberships
	config.Checkpoints = store.Checkpoints
	control, err := NewControlPlane(config)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	if _, err := control.CreateAccount(t.Context(), "Checkpoint_"+suffix, "long enough password"); err != nil {
		store.Close()
		t.Fatal(err)
	}
	session, err := control.Authenticate(t.Context(), "Checkpoint_"+suffix, "long enough password")
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	if _, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "CheckpointHero", Class: "Amazon"}); err != nil {
		store.Close()
		t.Fatal(err)
	}
	handoff, err := control.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Checkpoint " + suffix, Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	gameID := handoff.Game.Entry.GameID
	if reconciled, err := control.ReconcileGames(t.Context()); err != nil || reconciled != 0 {
		store.Close()
		t.Fatalf("reconcile=%d error=%v", reconciled, err)
	}
	initial, err := store.Checkpoints.Latest(t.Context(), gameID)
	if err != nil || initial.Tick != 0 {
		store.Close()
		t.Fatalf("initial checkpoint=%#v error=%v", initial, err)
	}
	newerState, err := testCheckpoint(allocator.workers[gameID].description.Runtime, 3)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	newer, err := NewGameCheckpoint(gameID, initial.AllocationID, initial.IdentityHash, newerState)
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	var group sync.WaitGroup
	errorsSeen := make(chan error, 8)
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			_, saveErr := store.Checkpoints.Save(t.Context(), newer)
			errorsSeen <- saveErr
		}()
	}
	group.Wait()
	close(errorsSeen)
	for saveErr := range errorsSeen {
		if saveErr != nil {
			store.Close()
			t.Fatalf("concurrent idempotent save: %v", saveErr)
		}
	}
	if _, err := store.Checkpoints.Save(t.Context(), initial); !errors.Is(err, ErrGameCheckpoint) {
		store.Close()
		t.Fatalf("checkpoint regression error = %v", err)
	}
	store.Close()

	restarted, err := OpenPostgres(t.Context(), connectionString, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	persisted, err := restarted.Checkpoints.Latest(t.Context(), gameID)
	if err != nil || persisted.Tick != 3 || persisted.Checksum != newer.Checksum {
		t.Fatalf("persisted checkpoint=%#v error=%v", persisted, err)
	}
	if _, err := restarted.Pool.Exec(t.Context(), `UPDATE realm_game_checkpoints
		SET payload_digest = decode(repeat('00', 32), 'hex') WHERE game_id = $1`, gameID); err != nil {
		t.Fatal(err)
	}
	if _, err := restarted.Checkpoints.Latest(t.Context(), gameID); !errors.Is(err, ErrGameCheckpoint) {
		t.Fatalf("tampered checkpoint error = %v", err)
	}
	if _, err := restarted.Pool.Exec(t.Context(), `DELETE FROM realm_games WHERE id = $1`, gameID); err != nil {
		t.Fatal(err)
	}
}
