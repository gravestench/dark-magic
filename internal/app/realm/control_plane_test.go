package realm

import (
	"errors"
	"testing"
	"time"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// TestControlPlaneReauthorizesEveryLobbyOperation verifies ordinary traffic
// observes session expiry and removes the corresponding channel presence.
func TestControlPlaneReauthorizesEveryLobbyOperation(t *testing.T) {
	config := orchestratedControlConfig(nil)
	config.SessionLifetime, config.ChatHistory = time.Minute, 4

	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Unix(100, 0)
	control.accounts.(*Accounts).now = func() time.Time { return now }

	account, err := control.CreateAccount(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	if err := control.characters.Create(t.Context(), CharacterRecord{
		AccountID: account.ID,
		Revision:  1,
		Character: d2save.Character{
			ID:        "character:alice",
			Name:      "Alyssa",
			Class:     "Assassin",
			Level:     1,
			Expansion: true,
		},
	}); err != nil {
		t.Fatal(err)
	}

	session, err := control.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.SelectCharacter(t.Context(), session.Token, "character:alice"); err != nil {
		t.Fatal(err)
	}

	if _, err := control.JoinChannel(t.Context(), session.Token, "Diablo II"); err != nil {
		t.Fatal(err)
	}

	if _, err := control.SendChannelMessage(t.Context(), session.Token, "hello"); err != nil {
		t.Fatal(err)
	}

	game, err := control.CreateGame(t.Context(), session.Token, CreateGameRequest{
		Name:       "Fresh Game",
		Difficulty: DifficultyNormal,
		Maximum:    8,
	})
	if err != nil {
		t.Fatal(err)
	}

	if game.Game.Entry.GameID == "" || game.Assignment.GameID != game.Game.Entry.GameID || game.Assignment.Ticket == "" {
		t.Fatalf("handoff = %#v", game)
	}

	listed, err := control.ListGames(t.Context(), session.Token, GameFilter{})
	if err != nil || len(listed) != 1 || listed[0].GameID != game.Game.Entry.GameID || listed[0].Players != 1 {
		t.Fatalf("games=%#v error=%v", listed, err)
	}

	if _, err := control.ChannelView(t.Context(), session.Token); !errors.Is(err, ErrChannelMember) {
		t.Fatalf("admitted player remained visible in public channel: %v", err)
	}

	now = now.Add(time.Minute)

	if _, err := control.ChannelView(t.Context(), session.Token); !errors.Is(err, ErrRealmSession) {
		t.Fatalf("expired channel view error = %v", err)
	}

	if _, present := control.channels.bySession[session.ID]; present {
		t.Fatal("expired session remains in channel presence")
	}

	if _, err := control.ListGames(t.Context(), session.Token, GameFilter{}); !errors.Is(err, ErrRealmSession) {
		t.Fatalf("expired game list error = %v", err)
	}
}

// TestControlPlaneLogoutRemovesChannelPresence ensures presence disappears
// before the session token becomes invalid.
func TestControlPlaneLogoutRemovesChannelPresence(t *testing.T) {
	control, err := NewControlPlane(ControlPlaneConfig{})
	if err != nil {
		t.Fatal(err)
	}

	account, err := control.CreateAccount(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	if err := control.characters.Create(t.Context(), CharacterRecord{AccountID: account.ID, Revision: 1,
		Character: d2save.Character{ID: "character", Name: "Hero", Class: "Amazon", Level: 1}}); err != nil {
		t.Fatal(err)
	}

	session, err := control.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.SelectCharacter(t.Context(), session.Token, "character"); err != nil {
		t.Fatal(err)
	}

	if _, err := control.JoinChannel(t.Context(), session.Token, "Diablo II"); err != nil {
		t.Fatal(err)
	}

	principal, err := control.accounts.Authorize(t.Context(), session.Token)
	if err != nil {
		t.Fatal(err)
	}

	if err := control.Logout(t.Context(), session.Token); err != nil {
		t.Fatal(err)
	}

	if _, err := control.accounts.Authorize(t.Context(), session.Token); !errors.Is(err, ErrRealmSession) {
		t.Fatalf("authorization after logout error = %v", err)
	}

	if _, err := control.channels.View(t.Context(), principal); !errors.Is(err, ErrChannelMember) {
		t.Fatalf("channel presence after logout error = %v", err)
	}
}

// TestControlPlaneCreatesAndSelectsRealmOwnedCharacter verifies the realm owns
// defaults, identity, class normalization, and duplicate-name enforcement.
func TestControlPlaneCreatesAndSelectsRealmOwnedCharacter(t *testing.T) {
	control, err := NewControlPlane(ControlPlaneConfig{})
	if err != nil {
		t.Fatal(err)
	}

	account, err := control.CreateAccount(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	session, err := control.Authenticate(t.Context(), account.Name, "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	record, err := control.CreateCharacter(t.Context(), session.Token, CreateCharacterRequest{
		Name: "RealmHero", Class: "barbarian", Expansion: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if record.AccountID != account.ID || record.Revision != 1 || record.Character.ID == "" ||
		record.Character.Class != "Barbarian" || record.Character.Stats == nil || record.Character.Stats.MaxHealth != 55 {
		t.Fatalf("record = %#v", record)
	}

	selected, err := control.SelectedCharacter(t.Context(), session.Token)
	if err != nil || selected.Character.ID != record.Character.ID {
		t.Fatalf("selected=%#v error=%v", selected, err)
	}

	listed, err := control.ListCharacters(t.Context(), session.Token)
	if err != nil || len(listed) != 1 || listed[0].Character.ID != record.Character.ID {
		t.Fatalf("listed=%#v error=%v", listed, err)
	}

	if _, err := control.CreateCharacter(
		t.Context(),
		session.Token,
		CreateCharacterRequest{Name: "realmhero", Class: "Amazon"},
	); !errors.Is(err, ErrCharacterExists) {
		t.Fatalf("duplicate error = %v", err)
	}
}

// TestControlPlaneRequiresSelectionBeforeChannelJoin prevents sessions from
// publishing arbitrary presence before selecting an owned character.
func TestControlPlaneRequiresSelectionBeforeChannelJoin(t *testing.T) {
	control, err := NewControlPlane(ControlPlaneConfig{})
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

	if _, err := control.JoinChannel(t.Context(), session.Token, "Diablo II"); !errors.Is(err, ErrCharacterNotFound) {
		t.Fatalf("join error = %v", err)
	}
}
