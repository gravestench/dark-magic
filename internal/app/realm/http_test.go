package realm

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRealmHTTPClientCompletesLobbyFlow verifies realm httpclient completes lobby flow. The scenario keeps the http
// contract visible to maintainers.
func TestRealmHTTPClientCompletesLobbyFlow(t *testing.T) {
	control, err := NewControlPlane(orchestratedControlConfig(nil))
	if err != nil {
		t.Fatal(err)
	}

	handler, err := NewHTTPHandler(control)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewRealmClient(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}

	info, err := client.ServiceInfo(t.Context())
	if err != nil || info.Version != RealmControlPlaneVersion {
		t.Fatalf("service info=%#v error=%v", info, err)
	}

	if _, err := client.ListCharacters(t.Context()); !errors.Is(err, ErrAccountCredentials) {
		t.Fatalf("unauthenticated character list error = %v", err)
	}

	if _, err := client.CreateAccount(t.Context(), "Alice", "long enough password"); err != nil {
		t.Fatal(err)
	}

	session, err := client.Authenticate(t.Context(), "Alice", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	character, err := client.CreateCharacter(
		t.Context(),
		CreateCharacterRequest{Name: "RealmHero", Class: "Amazon", Expansion: true},
	)
	if err != nil {
		t.Fatal(err)
	}

	if character.Character.Stats == nil || character.Character.Stats.MaxHealth != 50 {
		t.Fatalf("character = %#v", character)
	}

	retired, err := client.CreateCharacter(
		t.Context(),
		CreateCharacterRequest{Name: "Retired", Class: "Druid", Expansion: true},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := client.DeleteCharacter(t.Context(), retired.Character.ID); err != nil {
		t.Fatal(err)
	}

	characters, err := client.ListCharacters(t.Context())
	if err != nil || len(characters) != 1 || characters[0].Character.ID != character.Character.ID {
		t.Fatalf("characters after delete=%#v error=%v", characters, err)
	}

	if _, err := client.SelectCharacter(t.Context(), character.Character.ID); err != nil {
		t.Fatal(err)
	}

	view, err := client.JoinChannel(t.Context(), "Diablo II")
	if err != nil || len(view.Members) != 1 || view.Members[0].Character.CharacterID != character.Character.ID {
		t.Fatalf("view=%#v error=%v", view, err)
	}

	if _, err := client.SendMessage(t.Context(), "hello realm"); err != nil {
		t.Fatal(err)
	}

	game, err := client.CreateGame(t.Context(), CreateGameRequest{Name: "Tal Set4ur Ber", Difficulty: DifficultyNormal,
		Maximum: 8, CharacterDifference: 4, Expansion: true})
	if err != nil || game.Assignment.Ticket == "" || game.Game.Entry.CharacterDifference != 4 {
		t.Fatal(err)
	}

	games, err := client.ListGames(t.Context())
	if err != nil || len(games) != 1 || games[0].GameID != game.Game.Entry.GameID {
		t.Fatalf("games=%#v error=%v", games, err)
	}

	resolved, err := client.ResolveGame(t.Context(), game.Game.Entry.Name, "")
	if err != nil || resolved != game.Game.Entry.GameID {
		t.Fatalf("resolved=%q error=%v", resolved, err)
	}

	reconnected, err := client.ReconnectGame(t.Context(), game.Game.Entry.GameID)
	if err != nil || reconnected.Assignment.Ticket == "" ||
		reconnected.Assignment.GameID != game.Game.Entry.GameID {
		t.Fatalf("reconnect handoff=%#v error=%v", reconnected, err)
	}

	committed, err := client.LeaveGame(t.Context(), game.Game.Entry.GameID)
	if err != nil || committed.Character.ID != character.Character.ID || committed.Revision != 2 {
		t.Fatalf("committed=%#v error=%v", committed, err)
	}

	if games, err := client.ListGames(t.Context()); err != nil || len(games) != 0 {
		t.Fatalf("games after leave=%#v error=%v", games, err)
	}

	retried, err := client.LeaveGame(t.Context(), game.Game.Entry.GameID)
	if err != nil || retried.Revision != committed.Revision {
		t.Fatalf("retried leave=%#v error=%v", retried, err)
	}

	if err := client.Logout(t.Context()); err != nil {
		t.Fatal(err)
	}

	if _, present := control.channels.bySession[session.ID]; present {
		t.Fatal("logged-out Realm HTTP client remained in live channel presence")
	}

	if _, err := client.ListCharacters(t.Context()); !errors.Is(err, ErrAccountCredentials) {
		t.Fatalf("logged-out character list error = %v", err)
	}
}

// TestRealmHTTPRejectsUnknownInput verifies realm httprejects unknown input. The scenario keeps the http contract
// visible to maintainers.
func TestRealmHTTPRejectsUnknownInput(t *testing.T) {
	control, _ := NewControlPlane(ControlPlaneConfig{})
	handler, _ := NewHTTPHandler(control)
	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/accounts",
		strings.NewReader(`{"name":"Alice","password":"long enough password","admin":true}`),
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

// TestRealmOperatorDrainRequiresIndependentCredential verifies realm operator drain requires independent credential.
// The scenario keeps the http contract visible to maintainers.
func TestRealmOperatorDrainRequiresIndependentCredential(t *testing.T) {
	allocator := newOrchestrationAllocator()
	config := orchestratedControlConfig(nil)
	config.Allocator = allocator

	control, err := NewControlPlane(config)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.CreateAccount(t.Context(), "OperatorDrain", "long enough password"); err != nil {
		t.Fatal(err)
	}

	session, err := control.Authenticate(t.Context(), "OperatorDrain", "long enough password")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := control.CreateCharacter(t.Context(), session.Token,
		CreateCharacterRequest{Name: "OperatorHero", Class: "Amazon"}); err != nil {
		t.Fatal(err)
	}

	handoff, err := control.CreateGame(t.Context(), session.Token,
		CreateGameRequest{Name: "Operator Game", Difficulty: DifficultyNormal, Maximum: 8})
	if err != nil {
		t.Fatal(err)
	}

	const operatorToken = "operator-token-0123456789abcdef0123456789"

	handler, err := NewOperatorHTTPHandler(control, operatorToken)
	if err != nil {
		t.Fatal(err)
	}

	payload := `{"game_id":"` + handoff.Game.Entry.GameID + `"}`
	request := httptest.NewRequest(http.MethodPost, "/v1/operator/games/drain", strings.NewReader(payload))
	request.Header.Set("Authorization", "Bearer "+session.Token)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || allocator.releases != 0 {
		t.Fatalf(
			"player-token drain status=%d releases=%d body=%s",
			response.Code,
			allocator.releases,
			response.Body.String(),
		)
	}

	request = httptest.NewRequest(http.MethodPost, "/v1/operator/games/drain", strings.NewReader(payload))
	request.Header.Set("Authorization", "Bearer "+operatorToken)

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK || allocator.releases != 1 {
		t.Fatalf("operator drain status=%d releases=%d body=%s", response.Code, allocator.releases, response.Body.String())
	}
}
