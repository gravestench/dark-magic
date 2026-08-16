package clientapp

import (
	"context"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/app/realm"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/gravestench/dark-magic/internal/preferences"
)

type fakeRealmAPI struct {
	character     realm.CharacterSummary
	handoff       realm.GameHandoff
	logouts       int
	joins         int
	detailRef     string
	channelErr    error
	createRequest realm.CreateGameRequest
}

func (*fakeRealmAPI) ServiceInfo(context.Context) (realm.ServiceInfo, error) {
	return realm.ServiceInfo{Version: realm.RealmControlPlaneVersion}, nil
}
func (*fakeRealmAPI) Signup(context.Context, string, string, string) (realm.Account, error) {
	return realm.Account{ID: "account", Name: "Alice"}, nil
}
func (*fakeRealmAPI) Authenticate(context.Context, string, string) (realm.RealmSession, error) {
	return realm.RealmSession{Account: realm.Account{ID: "account", Name: "Alice"}}, nil
}
func (api *fakeRealmAPI) Logout(context.Context) error {
	api.logouts++
	return nil
}
func (*fakeRealmAPI) BeginPasswordRecovery(context.Context, string) error { return nil }
func (api *fakeRealmAPI) ListCharacters(context.Context) ([]realm.CharacterSummary, error) {
	if api.character.Character.ID == "" {
		return []realm.CharacterSummary{}, nil
	}
	return []realm.CharacterSummary{api.character}, nil
}
func (api *fakeRealmAPI) CreateCharacter(_ context.Context, request realm.CreateCharacterRequest) (realm.CharacterSummary, error) {
	api.character = realm.CharacterSummary{Character: d2save.Character{ID: "hero", Name: request.Name, Class: request.Class}}
	return api.character, nil
}
func (api *fakeRealmAPI) DeleteCharacter(context.Context, string) error {
	api.character = realm.CharacterSummary{}
	return nil
}
func (api *fakeRealmAPI) SelectCharacter(context.Context, string) (realm.CharacterSummary, error) {
	return api.character, nil
}

func (api *fakeRealmAPI) JoinChannel(context.Context, string) (realm.ChannelView, error) {
	api.joins++
	return realm.ChannelView{Name: "Diablo II"}, nil
}
func (api *fakeRealmAPI) Channel(context.Context) (realm.ChannelView, error) {
	if api.channelErr != nil {
		err := api.channelErr
		api.channelErr = nil
		return realm.ChannelView{}, err
	}
	return realm.ChannelView{Name: "Diablo II"}, nil
}
func (*fakeRealmAPI) ChannelEvents(context.Context, uint64, int) ([]realm.ChatEvent, error) {
	return []realm.ChatEvent{}, nil
}
func (*fakeRealmAPI) SendMessage(context.Context, string) (realm.ChatEvent, error) {
	return realm.ChatEvent{}, nil
}
func (*fakeRealmAPI) ListGames(context.Context) ([]realm.GameDirectoryEntry, error) {
	return []realm.GameDirectoryEntry{{GameID: "game", Name: "Fresh"}}, nil
}
func (api *fakeRealmAPI) GameDetail(_ context.Context, reference string) (realm.GameDetail, error) {
	api.detailRef = reference
	return realm.GameDetail{Entry: realm.GameDirectoryEntry{GameID: "game", Name: "Fresh"},
		Players: []realm.GamePlayer{{CharacterID: "hero", Name: "Hero", Class: "Amazon", Level: 1}}}, nil
}
func (api *fakeRealmAPI) CreateGame(_ context.Context, request realm.CreateGameRequest) (realm.GameHandoff, error) {
	api.createRequest = request
	return api.gameHandoff(), nil
}
func (*fakeRealmAPI) ResolveGame(context.Context, string, string) (string, error) { return "game", nil }
func (api *fakeRealmAPI) JoinGame(context.Context, string, string) (realm.GameHandoff, error) {
	return api.gameHandoff(), nil
}
func (api *fakeRealmAPI) ReconnectGame(context.Context, string) (realm.GameHandoff, error) {
	return api.gameHandoff(), nil
}
func (api *fakeRealmAPI) LeaveGame(context.Context, string) (realm.CharacterSummary, error) {
	api.character.Revision++
	return api.character, nil
}

func (api *fakeRealmAPI) gameHandoff() realm.GameHandoff {
	if api.handoff.Assignment.GameID != "" {
		return api.handoff
	}
	return realm.GameHandoff{Game: realm.GameDetail{Entry: realm.GameDirectoryEntry{GameID: "game"}},
		Assignment: realm.JoinAssignment{GameID: "game", Ticket: "private-ticket",
			Endpoint: realm.GameEndpoint{Address: "game.internal:6112", TLSFingerprint: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}}
}

type fakeRealmGameConnector struct {
	assignment realm.JoinAssignment
	err        error
}

func (connector *fakeRealmGameConnector) ConnectRealm(_ context.Context, assignment realm.JoinAssignment) error {
	connector.assignment = assignment
	return connector.err
}

func (connector *fakeRealmGameConnector) ReconnectRealm(_ context.Context, assignment realm.JoinAssignment) error {
	connector.assignment = assignment
	return connector.err
}

func TestNormalizeRealmEndpointUsesTLSAndLegacyPort(t *testing.T) {
	endpoint, address, err := normalizeRealmEndpoint("127.0.0.1")
	if err != nil || endpoint != "https://127.0.0.1:6112" || address != "127.0.0.1:6112" {
		t.Fatalf("endpoint=%q address=%q error=%v", endpoint, address, err)
	}
	if _, _, err := normalizeRealmEndpoint("http://realm.example"); err == nil {
		t.Fatal("plaintext realm endpoint accepted")
	}
}

func TestRealmControllerRequiresExplicitLoginAfterCompatibilityCheck(t *testing.T) {
	control, err := realm.NewControlPlane(realm.ControlPlaneConfig{})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := realm.NewHTTPHandler(control)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(handler)
	defer server.Close()
	trust, err := networktrust.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	controller := newRealmController(&application{
		ctx: context.Background(), networkTrust: trust, gameSettings: preferences.NewTransient(),
	})
	if _, err := control.CreateAccount(t.Context(), "Alice", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	if err := controller.Connect(server.URL); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "login")
	status := controller.Status()
	if status["endpoint"] != server.URL || status["gateway"] != server.URL {
		t.Fatalf("status=%#v", status)
	}
	if account := status["account"].(map[string]any); account["id"] != "" {
		t.Fatalf("compatibility check implicitly logged in: %#v", status)
	}
	if err := controller.Login("Alice", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "characters")
	if account := controller.Status()["account"].(map[string]any); account["name"] != "Alice" {
		t.Fatalf("logged in account = %#v", account)
	}
}

func TestRealmControllerSignupAndRecoveryDoNotImplicitlyLogin(t *testing.T) {
	controller := newRealmController(&application{ctx: context.Background()})
	controller.client = &fakeRealmAPI{}
	controller.state.Phase = "login"
	if err := controller.Signup("Alice", "alice@example.test", "correct horse battery staple"); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "verification_required")
	if err := controller.RecoverPassword("alice@example.test"); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "recovery_sent")
}

func TestRealmControllerPersistsSelectedGateway(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")
	settings, err := preferences.New(path)
	if err != nil {
		t.Fatal(err)
	}
	controller := newRealmController(&application{ctx: context.Background(), gameSettings: settings})
	if err := controller.SetGateway("realm.example"); err != nil {
		t.Fatal(err)
	}
	reloaded, err := preferences.New(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Values().RealmGateway; got != "realm.example" {
		t.Fatalf("gateway=%q", got)
	}
}

func TestRealmControllerRunsCharacterThenLobbyFlowAsynchronously(t *testing.T) {
	controller := newRealmController(&application{ctx: context.Background()})
	controller.client = &fakeRealmAPI{}
	controller.state.Phase = "characters"
	if err := controller.CreateCharacter("Hero", "Amazon", true, false); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "characters")
	status := controller.Status()
	if status["selected"].(map[string]any)["character"].(map[string]any)["id"] != "hero" {
		t.Fatalf("status=%#v", status)
	}
	if err := controller.DeleteCharacter("hero"); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "characters")
	if characters := controller.Status()["characters"].([]any); len(characters) != 0 {
		t.Fatalf("characters after delete=%#v", characters)
	}
	if err := controller.CreateCharacter("Hero", "Amazon", true, false); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "characters")
	if err := controller.JoinChannel("Diablo II"); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "lobby")
}

func TestRealmControllerLogoutAndCloseClearLivePresence(t *testing.T) {
	api := &fakeRealmAPI{}
	controller := newRealmController(&application{ctx: context.Background()})
	controller.client = api
	controller.state = realmClientState{Phase: "lobby", Gateway: "realm.example", Endpoint: "https://realm.example:6112",
		Account: realm.Account{ID: "account", Name: "Alice"}}
	if err := controller.Logout(); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "login")
	if api.logouts != 1 || controller.state.Account.ID != "" {
		t.Fatalf("logouts=%d state=%#v", api.logouts, controller.state)
	}

	controller.state.Account = realm.Account{ID: "account", Name: "Alice"}
	if err := controller.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
	if api.logouts != 2 || controller.state.Phase != "disconnected" {
		t.Fatalf("logouts=%d state=%#v", api.logouts, controller.state)
	}
}

func TestRealmControllerRenewsPrunedChannelPresence(t *testing.T) {
	api := &fakeRealmAPI{channelErr: realm.ErrChannelMember}
	controller := newRealmController(&application{ctx: context.Background()})
	controller.client = api
	controller.state = realmClientState{Phase: "lobby", Channel: realm.ChannelView{Name: "Diablo II"}}
	if err := controller.Refresh(); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "lobby")
	if api.joins != 1 || controller.state.Channel.Name != "Diablo II" {
		t.Fatalf("joins=%d state=%#v", api.joins, controller.state)
	}
}

func TestRealmControllerLoadsSelectedGameDetail(t *testing.T) {
	api := &fakeRealmAPI{}
	controller := newRealmController(&application{ctx: context.Background()})
	controller.client = api
	controller.state.Phase = "lobby"
	if err := controller.SelectGame("game"); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "lobby")
	if api.detailRef != "game" || controller.state.SelectedGame.Entry.Name != "Fresh" ||
		len(controller.state.SelectedGame.Players) != 1 {
		t.Fatalf("reference=%q selected=%#v", api.detailRef, controller.state.SelectedGame)
	}
}

func TestRealmControllerHandsPrivateAssignmentDirectlyToNetwork(t *testing.T) {
	api := &fakeRealmAPI{character: realm.CharacterSummary{Revision: 1,
		Character: d2save.Character{ID: "hero", Name: "Hero", Class: "Amazon"}}}
	games := &fakeRealmGameConnector{}
	controller := newRealmController(&application{ctx: context.Background()})
	controller.client, controller.games, controller.state.Phase = api, games, "lobby"
	if err := controller.CreateGame(map[string]any{
		"name": "Fresh", "maximum_players": float64(6), "character_difference": float64(4),
	}); err != nil {
		t.Fatal(err)
	}
	waitRealmPhase(t, controller, "game_connected")
	if games.assignment.Ticket != "private-ticket" || games.assignment.Endpoint.Address != "game.internal:6112" {
		t.Fatalf("assignment = %#v", games.assignment)
	}
	if api.createRequest.Maximum != 6 || api.createRequest.CharacterDifference != 4 {
		t.Fatalf("create request = %#v", api.createRequest)
	}
	status := controller.Status()
	encoded := fmt.Sprintf("%#v", status)
	if strings.Contains(encoded, "private-ticket") || strings.Contains(encoded, "game.internal") {
		t.Fatalf("Lua-visible Realm status leaked private handoff: %s", encoded)
	}
	controller.state.Characters = []realm.CharacterSummary{api.character}
	controller.state.Selected = api.character
	if err := controller.LeaveConnectedGame(t.Context()); err != nil {
		t.Fatal(err)
	}
	status = controller.Status()
	if status["phase"] != "lobby" || status["resolved_game_id"] != nil {
		t.Fatalf("status after leave = %#v", status)
	}
	if selected := status["selected"].(map[string]any); selected["revision"] != float64(2) {
		t.Fatalf("selected after leave = %#v", selected)
	}
}

func waitRealmPhase(t *testing.T, controller *realmController, phase string) {
	t.Helper()
	// bcrypt and TLS are intentionally exercised here and are substantially
	// slower under the race detector and loaded CI hosts.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if controller.Status()["phase"] == phase {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("phase=%#v, want %s", controller.Status(), phase)
}
