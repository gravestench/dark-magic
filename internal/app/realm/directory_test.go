package realm

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestGameDirectorySeparatesNamesFromOpaqueIDs verifies game directory separates names from opaque ids. The scenario
// keeps the directory contract visible to maintainers.
func TestGameDirectorySeparatesNamesFromOpaqueIDs(t *testing.T) {
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	owner := authenticatedFixture(t, accounts, "Owner")
	directory := NewGameDirectory()
	now := time.Unix(100, 0)
	directory.now = func() time.Time { now = now.Add(time.Second); return now }

	public, err := directory.Create(t.Context(), owner, CreateGameRequest{Name: " Trist   Run  ", Description: "Act I",
		Difficulty: DifficultyNormal, Maximum: 3, CharacterDifference: 4, Expansion: true})
	if err != nil {
		t.Fatal(err)
	}

	if public.Entry.GameID == "" || public.Entry.GameID == public.Entry.Name || public.Entry.Name != "Trist Run" ||
		public.Entry.Version != GameDirectoryVersion || public.Entry.CharacterDifference != 4 {
		t.Fatalf("public game = %#v", public)
	}

	if _, err := directory.Create(
		t.Context(),
		owner,
		CreateGameRequest{Name: "trist run", Difficulty: DifficultyNormal, Maximum: 8},
	); !errors.Is(
		err,
		ErrGameExists,
	) {
		t.Fatalf("normalized duplicate error = %v", err)
	}

	private, err := directory.Create(t.Context(), owner, CreateGameRequest{Name: "Friends Only", Password: "secret",
		Difficulty: DifficultyNightmare, Maximum: 2, Expansion: true})
	if err != nil {
		t.Fatal(err)
	}

	listed, err := directory.List(t.Context(), GameFilter{})
	if err != nil {
		t.Fatal(err)
	}

	if len(listed) != 1 || listed[0].GameID != public.Entry.GameID || listed[0].PasswordRequired {
		t.Fatalf("public listing = %#v", listed)
	}

	if _, err := directory.Detail(t.Context(), "Friends Only"); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("private detail disclosure error = %v", err)
	}

	if _, err := directory.ResolveJoin(t.Context(), "friends only", "wrong"); !errors.Is(err, ErrGamePassword) {
		t.Fatalf("private password error = %v", err)
	}

	resolved, err := directory.ResolveJoin(t.Context(), " FRIENDS   ONLY ", "secret")
	if err != nil || resolved != private.Entry.GameID {
		t.Fatalf("resolved private ID=%q error=%v", resolved, err)
	}

	resolved, err = directory.ResolveJoin(t.Context(), public.Entry.GameID, "")
	if err != nil || resolved != public.Entry.GameID {
		t.Fatalf("resolved public ID=%q error=%v", resolved, err)
	}
}

// TestGameDirectoryValidatesCharacterDifference verifies game directory validates character difference. The scenario
// keeps the directory contract visible to maintainers.
func TestGameDirectoryValidatesCharacterDifference(t *testing.T) {
	accounts, _ := NewAccounts(time.Hour)
	owner := authenticatedFixture(t, accounts, "Owner")

	directory := NewGameDirectory()
	for _, difference := range []int{-1, 100} {
		_, err := directory.Create(t.Context(), owner, CreateGameRequest{
			Name: "Invalid", Difficulty: DifficultyNormal, Maximum: 8, CharacterDifference: difference,
		})
		if !errors.Is(err, ErrGameDirectoryInput) {
			t.Fatalf("difference %d error = %v", difference, err)
		}
	}
}

// TestGameDirectoryProjectsPlayersCapacityAndFilters verifies game directory projects players capacity and filters.
// The scenario keeps the directory contract visible to maintainers.
func TestGameDirectoryProjectsPlayersCapacityAndFilters(t *testing.T) {
	accounts, err := NewAccounts(time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	owner := authenticatedFixture(t, accounts, "Owner")
	directory := NewGameDirectory()

	game, err := directory.Create(t.Context(), owner, CreateGameRequest{Name: "Hardcore", Difficulty: DifficultyHell,
		Maximum: 2, Expansion: true, Hardcore: true})
	if err != nil {
		t.Fatal(err)
	}

	players := []GamePlayer{
		{CharacterID: "one", Name: "Alice", Class: "Assassin", Level: 80},
		{CharacterID: "two", Name: "Bob", Class: "Barbarian", Level: 79},
	}
	if err := directory.SetPlayers(t.Context(), game.Entry.GameID, players); err != nil {
		t.Fatal(err)
	}

	players[0].Name = "mutated"

	detail, err := directory.Detail(t.Context(), "hardcore")
	if err != nil {
		t.Fatal(err)
	}

	if detail.Entry.Players != 2 || detail.Entry.Revision != 2 || detail.Players[0].Name != "Alice" {
		t.Fatalf("detail = %#v", detail)
	}

	if _, err := directory.ResolveJoin(t.Context(), game.Entry.GameID, ""); !errors.Is(err, ErrGameFull) {
		t.Fatalf("full join error = %v", err)
	}

	normal := DifficultyNormal
	if entries, err := directory.List(t.Context(), GameFilter{Difficulty: &normal}); err != nil || len(entries) != 0 {
		t.Fatalf("filtered entries=%#v error=%v", entries, err)
	}

	if err := directory.SetPlayers(
		t.Context(),
		game.Entry.GameID,
		[]GamePlayer{
			{CharacterID: "same", Name: "A", Class: "Amazon", Level: 1},
			{CharacterID: "same", Name: "B", Class: "Amazon", Level: 1},
		},
	); !errors.Is(
		err,
		ErrGameDirectoryInput,
	) {
		t.Fatalf("duplicate character error = %v", err)
	}

	removed, err := directory.RemovePlayer(t.Context(), game.Entry.GameID, "one")
	if err != nil || removed.Entry.Players != 1 || len(removed.Players) != 1 || removed.Players[0].CharacterID != "two" {
		t.Fatalf("roster after removal=%#v error=%v", removed, err)
	}

	if _, err := directory.RemovePlayer(t.Context(), game.Entry.GameID, "missing"); !errors.Is(err, ErrCharacterNotFound) {
		t.Fatalf("missing removal error = %v", err)
	}

	if err := directory.Remove(t.Context(), game.Entry.GameID); err != nil {
		t.Fatal(err)
	}

	if _, err := directory.Detail(t.Context(), game.Entry.GameID); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("removed detail error = %v", err)
	}
}

// TestGameDirectoryReservationsArePrivateAtomicAndCancelable verifies game directory reservations are private atomic
// and cancelable. The scenario keeps the directory contract visible to maintainers.
func TestGameDirectoryReservationsArePrivateAtomicAndCancelable(t *testing.T) {
	accounts, _ := NewAccounts(time.Hour)
	owner := authenticatedFixture(t, accounts, "Owner")
	directory := NewGameDirectory()

	game, err := directory.Create(
		t.Context(),
		owner,
		CreateGameRequest{Name: "Two Seats", Difficulty: DifficultyNormal, Maximum: 2},
	)
	if err != nil {
		t.Fatal(err)
	}

	first, err := directory.ReservePlayer(
		t.Context(),
		game.Entry.GameID,
		GamePlayer{CharacterID: "one", Name: "One", Class: "Amazon", Level: 1},
	)
	if err != nil {
		t.Fatal(err)
	}

	second, err := directory.ReservePlayer(
		t.Context(),
		game.Entry.GameID,
		GamePlayer{CharacterID: "two", Name: "Two", Class: "Barbarian", Level: 1},
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := directory.ReservePlayer(
		t.Context(),
		game.Entry.GameID,
		GamePlayer{CharacterID: "three", Name: "Three", Class: "Amazon", Level: 1},
	); !errors.Is(
		err,
		ErrGameFull,
	) {
		t.Fatalf("third reservation error = %v", err)
	}

	visible, err := directory.Detail(t.Context(), game.Entry.GameID)
	if err != nil || visible.Entry.Players != 0 || len(visible.Players) != 0 {
		t.Fatalf("pending reservations leaked publicly: %#v error=%v", visible, err)
	}

	committed, err := directory.CommitPlayer(t.Context(), first)
	if err != nil || committed.Entry.Players != 1 || len(committed.Players) != 1 {
		t.Fatalf("committed=%#v error=%v", committed, err)
	}

	if err := directory.CancelPlayer(t.Context(), second); err != nil {
		t.Fatal(err)
	}

	if _, err := directory.ReservePlayer(
		t.Context(),
		game.Entry.GameID,
		GamePlayer{CharacterID: "three", Name: "Three", Class: "Amazon", Level: 1},
	); err != nil {
		t.Fatalf("canceled capacity was not returned: %v", err)
	}
}

// TestGameDirectoryDrainClosesDiscoveryAndAdmissionButAllowsDeparture verifies game directory drain closes discovery
// and admission but allows departure. The scenario keeps the directory contract visible to maintainers.
func TestGameDirectoryDrainClosesDiscoveryAndAdmissionButAllowsDeparture(t *testing.T) {
	accounts, _ := NewAccounts(time.Hour)
	owner := authenticatedFixture(t, accounts, "Owner")
	directory := NewGameDirectory()

	game, err := directory.Create(t.Context(), owner,
		CreateGameRequest{Name: "Drain Me", Difficulty: DifficultyNormal, Maximum: 2})
	if err != nil {
		t.Fatal(err)
	}

	reservation, err := directory.ReservePlayer(t.Context(), game.Entry.GameID,
		GamePlayer{CharacterID: "one", Name: "One", Class: "Amazon", Level: 1})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := directory.CommitPlayer(t.Context(), reservation); err != nil {
		t.Fatal(err)
	}

	if err := directory.BeginDrain(t.Context(), game.Entry.GameID); err != nil {
		t.Fatal(err)
	}

	if err := directory.BeginDrain(t.Context(), game.Entry.GameID); err != nil {
		t.Fatalf("retry drain: %v", err)
	}

	if games, err := directory.List(t.Context(), GameFilter{}); err != nil || len(games) != 0 {
		t.Fatalf("draining games=%#v error=%v", games, err)
	}

	if _, err := directory.Detail(t.Context(), game.Entry.GameID); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("draining detail error=%v", err)
	}

	if _, err := directory.ResolveJoin(t.Context(), game.Entry.GameID, ""); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("draining resolve error=%v", err)
	}

	if _, err := directory.ReservePlayer(t.Context(), game.Entry.GameID,
		GamePlayer{CharacterID: "two", Name: "Two", Class: "Amazon", Level: 1}); !errors.Is(err, ErrGameNotFound) {
		t.Fatalf("draining reservation error=%v", err)
	}

	departed, err := directory.RemovePlayer(t.Context(), game.Entry.GameID, "one")
	if err != nil || departed.Entry.Players != 0 {
		t.Fatalf("draining departure=%#v error=%v", departed, err)
	}
}

// TestRealmServicesRejectCancellationAndInvalidPrincipals verifies realm services reject cancellation and invalid
// principals. The scenario keeps the directory contract visible to maintainers.
func TestRealmServicesRejectCancellationAndInvalidPrincipals(t *testing.T) {
	control, err := NewControlPlane(ControlPlaneConfig{})
	if err != nil {
		t.Fatal(err)
	}

	if control.Version() != RealmControlPlaneVersion || control.accounts == nil || control.channels == nil ||
		control.games == nil ||
		control.characters == nil {
		t.Fatalf("control plane = %#v", control)
	}

	if _, err := control.games.Create(
		t.Context(),
		AuthenticatedPrincipal{},
		CreateGameRequest{Name: "Game", Difficulty: DifficultyNormal, Maximum: 8},
	); !errors.Is(
		err,
		ErrGameDirectoryInput,
	) {
		t.Fatalf("fabricated principal error = %v", err)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := control.games.List(cancelled, GameFilter{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled list error = %v", err)
	}
}
