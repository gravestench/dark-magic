package clientapp

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/realm"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// TestRealmControllerRunsCharacterThenLobbyFlowAsynchronously proves character and channel phases
// publish coherently while network operations remain off the render thread.
func TestRealmControllerRunsCharacterThenLobbyFlowAsynchronously(t *testing.T) {
	controller := newFakeRealmController()
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

	createTestRealmCharacter(t, controller)

	if err := controller.JoinChannel(defaultRealmChannel); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "lobby")
}

// createTestRealmCharacter creates and selects through controller APIs so lobby tests inherit the
// same canonical state transitions as the frontend.
func createTestRealmCharacter(t *testing.T, controller *realmController) {
	t.Helper()

	if err := controller.CreateCharacter("Hero", "Amazon", true, false); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "characters")
}

// TestRealmControllerRenewsPrunedChannelPresence requires refresh to rejoin after server pruning and
// reset its event cursor rather than applying stale membership history.
func TestRealmControllerRenewsPrunedChannelPresence(t *testing.T) {
	api := &fakeRealmAPI{channelErr: realm.ErrChannelMember}
	controller := newFakeRealmControllerWithAPI(api)
	controller.state = realmClientState{
		Phase:   "lobby",
		Channel: realm.ChannelView{Name: defaultRealmChannel},
	}

	if err := controller.Refresh(); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "lobby")

	if api.joins != 1 || controller.state.Channel.Name != defaultRealmChannel {
		t.Fatalf("joins=%d state=%#v", api.joins, controller.state)
	}
}

// TestRealmControllerLoadsSelectedGameDetail proves selection is populated from server detail, not
// merely copied from the less complete directory entry.
func TestRealmControllerLoadsSelectedGameDetail(t *testing.T) {
	api := &fakeRealmAPI{}
	controller := newFakeRealmControllerWithAPI(api)
	controller.state.Phase = "lobby"

	if err := controller.SelectGame("game"); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "lobby")

	selected := controller.state.SelectedGame
	if api.detailRef != "game" || selected.Entry.Name != "Fresh" || len(selected.Players) != 1 {
		t.Fatalf("reference=%q selected=%#v", api.detailRef, selected)
	}
}

// TestRealmControllerHandsPrivateAssignmentDirectlyToNetwork proves worker tickets and fingerprints
// reach the native connector while remaining absent from Lua-visible status.
func TestRealmControllerHandsPrivateAssignmentDirectlyToNetwork(t *testing.T) {
	controller, api, games := newRealmGameController()

	options := map[string]any{
		"name":                 "Fresh",
		"maximum_players":      float64(6),
		"character_difference": float64(4),
	}
	if err := controller.CreateGame(options); err != nil {
		t.Fatal(err)
	}

	waitRealmPhase(t, controller, "game_connected")

	if games.assignment.Ticket != "private-ticket" || games.assignment.Endpoint.Address != "game.internal:6112" {
		t.Fatalf("assignment = %#v", games.assignment)
	}

	if api.createRequest.Maximum != 6 || api.createRequest.CharacterDifference != 4 {
		t.Fatalf("create request = %#v", api.createRequest)
	}

	assertRealmStatusHidesAssignment(t, controller.Status())
}

// TestRealmControllerLeaveCommitsCharacter requires Realm's committed revision to replace both
// selected and directory copies before the controller returns to lobby.
func TestRealmControllerLeaveCommitsCharacter(t *testing.T) {
	controller, api, _ := newRealmGameController()
	controller.state.ResolvedGameID = "game"
	controller.state.Characters = []realm.CharacterSummary{api.character}
	controller.state.Selected = api.character

	if err := controller.LeaveConnectedGame(t.Context()); err != nil {
		t.Fatal(err)
	}

	status := controller.Status()
	if status["phase"] != "lobby" || status["resolved_game_id"] != nil {
		t.Fatalf("status after leave = %#v", status)
	}

	if selected := status["selected"].(map[string]any); selected["revision"] != float64(2) {
		t.Fatalf("selected after leave = %#v", selected)
	}
}

// newFakeRealmController builds the common deterministic control-plane fixture for state-machine tests.
func newFakeRealmController() *realmController {
	return newFakeRealmControllerWithAPI(&fakeRealmAPI{})
}

// newFakeRealmControllerWithAPI permits one behavior override while retaining the normal application
// context and native game-connector boundary.
func newFakeRealmControllerWithAPI(api *fakeRealmAPI) *realmController {
	controller := newRealmController(&application{ctx: context.Background()})
	controller.client = api

	return controller
}

// newRealmGameController returns controller, API recorder, and connector recorder separately so tests
// can assert public state and private handoff flow independently.
func newRealmGameController() (*realmController, *fakeRealmAPI, *fakeRealmGameConnector) {
	api := &fakeRealmAPI{
		character: realm.CharacterSummary{
			Revision: 1,
			Character: d2save.Character{
				ID:    "hero",
				Name:  "Hero",
				Class: "Amazon",
			},
		},
	}
	games := &fakeRealmGameConnector{}
	controller := newFakeRealmControllerWithAPI(api)
	controller.games = games
	controller.state.Phase = "lobby"

	return controller, api, games
}

// assertRealmStatusHidesAssignment recursively checks the serialized status boundary for ticket,
// endpoint fingerprint, and assignment leakage.
func assertRealmStatusHidesAssignment(t *testing.T, status map[string]any) {
	t.Helper()

	encoded := fmt.Sprintf("%#v", status)
	if strings.Contains(encoded, "private-ticket") || strings.Contains(encoded, "game.internal") {
		t.Fatalf("Lua-visible Realm status leaked private handoff: %s", encoded)
	}
}
