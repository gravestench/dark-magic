package clientapp

import (
	"context"

	"github.com/gravestench/dark-magic/internal/app/realm"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// fakeRealmAPI is a deterministic control-plane recorder that keeps public data and private handoffs
// independently configurable for trust-boundary tests.
type fakeRealmAPI struct {
	character     realm.CharacterSummary
	handoff       realm.GameHandoff
	logouts       int
	joins         int
	detailRef     string
	channelErr    error
	createRequest realm.CreateGameRequest
}

// ServiceInfo reports compatibility so tests can focus on the operation following connection.
func (*fakeRealmAPI) ServiceInfo(context.Context) (realm.ServiceInfo, error) {
	return realm.ServiceInfo{Version: realm.RealmControlPlaneVersion}, nil
}

// Signup returns public account data without authentication, preserving the real API's explicit-login
// contract.
func (*fakeRealmAPI) Signup(context.Context, string, string, string) (realm.Account, error) {
	return realm.Account{ID: "account", Name: "Alice"}, nil
}

// Authenticate returns the stable session identity used across character and lobby state tests.
func (*fakeRealmAPI) Authenticate(context.Context, string, string) (realm.RealmSession, error) {
	return realm.RealmSession{Account: realm.Account{ID: "account", Name: "Alice"}}, nil
}

// Logout records the explicit server-side cleanup obligation rather than merely clearing local state.
func (api *fakeRealmAPI) Logout(context.Context) error {
	api.logouts++

	return nil
}

// BeginPasswordRecovery succeeds without session mutation so phase behavior can be asserted in isolation.
func (*fakeRealmAPI) BeginPasswordRecovery(context.Context, string) error {
	return nil
}

// ListCharacters returns a fresh directory view from fake authority, matching reload-after-mutation flows.
func (api *fakeRealmAPI) ListCharacters(context.Context) ([]realm.CharacterSummary, error) {
	if api.character.Character.ID == "" {
		return []realm.CharacterSummary{}, nil
	}

	return []realm.CharacterSummary{api.character}, nil
}

// CreateCharacter stores normalized server-style identity so tests do not accidentally trust the
// frontend request as the canonical record.
func (api *fakeRealmAPI) CreateCharacter(
	_ context.Context,
	request realm.CreateCharacterRequest,
) (realm.CharacterSummary, error) {
	api.character = realm.CharacterSummary{
		Character: d2save.Character{
			ID:    "hero",
			Name:  request.Name,
			Class: request.Class,
		},
	}

	return api.character, nil
}

// DeleteCharacter clears fake authority, allowing the controller's directory and selection cleanup to
// be observed.
func (api *fakeRealmAPI) DeleteCharacter(context.Context, string) error {
	api.character = realm.CharacterSummary{}

	return nil
}

// SelectCharacter returns canonical fake state rather than echoing the requested identifier.
func (api *fakeRealmAPI) SelectCharacter(context.Context, string) (realm.CharacterSummary, error) {
	return api.character, nil
}

// JoinChannel records renewal count so tests can distinguish initial membership from prune recovery.
func (api *fakeRealmAPI) JoinChannel(context.Context, string) (realm.ChannelView, error) {
	api.joins++

	return realm.ChannelView{Name: defaultRealmChannel}, nil
}

// Channel consumes a configured one-shot failure to model server-side membership pruning deterministically.
func (api *fakeRealmAPI) Channel(context.Context) (realm.ChannelView, error) {
	if api.channelErr != nil {
		err := api.channelErr
		api.channelErr = nil

		return realm.ChannelView{}, err
	}

	return realm.ChannelView{Name: defaultRealmChannel}, nil
}

// ChannelEvents returns no changes by default, keeping lobby tests focused on cursor and membership flow.
func (*fakeRealmAPI) ChannelEvents(context.Context, uint64, int) ([]realm.ChatEvent, error) {
	return []realm.ChatEvent{}, nil
}

// SendMessage succeeds while leaving canonical installation to the subsequent refresh, as production does.
func (*fakeRealmAPI) SendMessage(context.Context, string) (realm.ChatEvent, error) {
	return realm.ChatEvent{}, nil
}

// ListGames returns a stable directory entry shared by selection and handoff scenarios.
func (*fakeRealmAPI) ListGames(context.Context) ([]realm.GameDirectoryEntry, error) {
	return []realm.GameDirectoryEntry{{GameID: "game", Name: "Fresh"}}, nil
}

// GameDetail records resolution input and returns richer player detail than the directory, proving the
// controller performs the second request.
func (api *fakeRealmAPI) GameDetail(_ context.Context, reference string) (realm.GameDetail, error) {
	api.detailRef = reference

	return realm.GameDetail{
		Entry: realm.GameDirectoryEntry{GameID: "game", Name: "Fresh"},
		Players: []realm.GamePlayer{
			{CharacterID: "hero", Name: "Hero", Class: "Amazon", Level: 1},
		},
	}, nil
}

// CreateGame records typed option conversion and returns the same private handoff shape as join.
func (api *fakeRealmAPI) CreateGame(
	_ context.Context,
	request realm.CreateGameRequest,
) (realm.GameHandoff, error) {
	api.createRequest = request

	return api.gameHandoff(), nil
}

// ResolveGame models Realm-owned reference resolution instead of letting tests treat UI labels as IDs.
func (*fakeRealmAPI) ResolveGame(context.Context, string, string) (string, error) {
	return "game", nil
}

// JoinGame returns ticket-bearing data whose handling is asserted at the native connector boundary.
func (api *fakeRealmAPI) JoinGame(context.Context, string, string) (realm.GameHandoff, error) {
	return api.gameHandoff(), nil
}

// ReconnectGame returns a fresh handoff for the same durable game identity used by recovery tests.
func (api *fakeRealmAPI) ReconnectGame(context.Context, string) (realm.GameHandoff, error) {
	return api.gameHandoff(), nil
}

// LeaveGame increments authority revision so stale cached character copies are detectable after exit.
func (api *fakeRealmAPI) LeaveGame(context.Context, string) (realm.CharacterSummary, error) {
	api.character.Revision++

	return api.character, nil
}

// gameHandoff permits targeted malformed or replacement assignments while centralizing the normal secret
// values used to detect presentation leakage.
func (api *fakeRealmAPI) gameHandoff() realm.GameHandoff {
	if api.handoff.Assignment.GameID != "" {
		return api.handoff
	}

	return realm.GameHandoff{
		Game: realm.GameDetail{
			Entry: realm.GameDirectoryEntry{GameID: "game"},
		},
		Assignment: realm.JoinAssignment{
			GameID: "game",
			Ticket: "private-ticket",
			Endpoint: realm.GameEndpoint{
				Address:        "game.internal:6112",
				TLSFingerprint: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
		},
	}
}

// fakeRealmGameConnector records assignments outside realmClientState, mirroring the production
// native-only trust path.
type fakeRealmGameConnector struct {
	assignment realm.JoinAssignment
	err        error
}

// ConnectRealm captures the initial private handoff without introducing transport behavior into controller tests.
func (connector *fakeRealmGameConnector) ConnectRealm(
	_ context.Context,
	assignment realm.JoinAssignment,
) error {
	connector.assignment = assignment

	return connector.err
}

// ReconnectRealm captures replacement credentials separately from the initial handoff.
func (connector *fakeRealmGameConnector) ReconnectRealm(
	_ context.Context,
	assignment realm.JoinAssignment,
) error {
	connector.assignment = assignment

	return connector.err
}
