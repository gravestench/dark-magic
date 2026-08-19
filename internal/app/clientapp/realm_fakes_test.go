package clientapp

import (
	"context"

	"github.com/gravestench/dark-magic/internal/app/realm"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// fakeRealmAPI records control-plane calls and supplies deterministic Realm data.
type fakeRealmAPI struct {
	character     realm.CharacterSummary
	handoff       realm.GameHandoff
	logouts       int
	joins         int
	detailRef     string
	channelErr    error
	createRequest realm.CreateGameRequest
}

// ServiceInfo reports a compatible control-plane version.
func (*fakeRealmAPI) ServiceInfo(context.Context) (realm.ServiceInfo, error) {
	return realm.ServiceInfo{Version: realm.RealmControlPlaneVersion}, nil
}

// Signup returns a deterministic account without creating a session.
func (*fakeRealmAPI) Signup(context.Context, string, string, string) (realm.Account, error) {
	return realm.Account{ID: "account", Name: "Alice"}, nil
}

// Authenticate returns a deterministic authenticated session.
func (*fakeRealmAPI) Authenticate(context.Context, string, string) (realm.RealmSession, error) {
	return realm.RealmSession{Account: realm.Account{ID: "account", Name: "Alice"}}, nil
}

// Logout records that the client explicitly cleared its Realm presence.
func (api *fakeRealmAPI) Logout(context.Context) error {
	api.logouts++

	return nil
}

// BeginPasswordRecovery accepts every recovery request.
func (*fakeRealmAPI) BeginPasswordRecovery(context.Context, string) error {
	return nil
}

// ListCharacters returns the fake's current character when one exists.
func (api *fakeRealmAPI) ListCharacters(context.Context) ([]realm.CharacterSummary, error) {
	if api.character.Character.ID == "" {
		return []realm.CharacterSummary{}, nil
	}

	return []realm.CharacterSummary{api.character}, nil
}

// CreateCharacter stores a character derived from the supplied request.
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

// DeleteCharacter removes the fake's current character.
func (api *fakeRealmAPI) DeleteCharacter(context.Context, string) error {
	api.character = realm.CharacterSummary{}

	return nil
}

// SelectCharacter returns the fake's current character.
func (api *fakeRealmAPI) SelectCharacter(context.Context, string) (realm.CharacterSummary, error) {
	return api.character, nil
}

// JoinChannel records the renewed membership and returns the default channel.
func (api *fakeRealmAPI) JoinChannel(context.Context, string) (realm.ChannelView, error) {
	api.joins++

	return realm.ChannelView{Name: defaultRealmChannel}, nil
}

// Channel returns either the configured one-shot error or the default channel.
func (api *fakeRealmAPI) Channel(context.Context) (realm.ChannelView, error) {
	if api.channelErr != nil {
		err := api.channelErr
		api.channelErr = nil

		return realm.ChannelView{}, err
	}

	return realm.ChannelView{Name: defaultRealmChannel}, nil
}

// ChannelEvents returns an empty event page.
func (*fakeRealmAPI) ChannelEvents(context.Context, uint64, int) ([]realm.ChatEvent, error) {
	return []realm.ChatEvent{}, nil
}

// SendMessage accepts every chat message.
func (*fakeRealmAPI) SendMessage(context.Context, string) (realm.ChatEvent, error) {
	return realm.ChatEvent{}, nil
}

// ListGames returns one deterministic directory entry.
func (*fakeRealmAPI) ListGames(context.Context) ([]realm.GameDirectoryEntry, error) {
	return []realm.GameDirectoryEntry{{GameID: "game", Name: "Fresh"}}, nil
}

// GameDetail records the requested reference and returns one player.
func (api *fakeRealmAPI) GameDetail(_ context.Context, reference string) (realm.GameDetail, error) {
	api.detailRef = reference

	return realm.GameDetail{
		Entry: realm.GameDirectoryEntry{GameID: "game", Name: "Fresh"},
		Players: []realm.GamePlayer{
			{CharacterID: "hero", Name: "Hero", Class: "Amazon", Level: 1},
		},
	}, nil
}

// CreateGame records the request and returns a private worker handoff.
func (api *fakeRealmAPI) CreateGame(
	_ context.Context,
	request realm.CreateGameRequest,
) (realm.GameHandoff, error) {
	api.createRequest = request

	return api.gameHandoff(), nil
}

// ResolveGame resolves every reference to the deterministic game ID.
func (*fakeRealmAPI) ResolveGame(context.Context, string, string) (string, error) {
	return "game", nil
}

// JoinGame returns a private worker handoff.
func (api *fakeRealmAPI) JoinGame(context.Context, string, string) (realm.GameHandoff, error) {
	return api.gameHandoff(), nil
}

// ReconnectGame returns a renewed private worker handoff.
func (api *fakeRealmAPI) ReconnectGame(context.Context, string) (realm.GameHandoff, error) {
	return api.gameHandoff(), nil
}

// LeaveGame increments the committed character revision.
func (api *fakeRealmAPI) LeaveGame(context.Context, string) (realm.CharacterSummary, error) {
	api.character.Revision++

	return api.character, nil
}

// gameHandoff returns an override or the default private worker assignment.
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

// fakeRealmGameConnector records the private assignment passed to native networking.
type fakeRealmGameConnector struct {
	assignment realm.JoinAssignment
	err        error
}

// ConnectRealm records an initial worker assignment.
func (connector *fakeRealmGameConnector) ConnectRealm(
	_ context.Context,
	assignment realm.JoinAssignment,
) error {
	connector.assignment = assignment

	return connector.err
}

// ReconnectRealm records a renewed worker assignment.
func (connector *fakeRealmGameConnector) ReconnectRealm(
	_ context.Context,
	assignment realm.JoinAssignment,
) error {
	connector.assignment = assignment

	return connector.err
}
