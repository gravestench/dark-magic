package clientapp

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

// CreateGame converts the untyped Lua boundary before contacting Realm, then passes the private
// worker handoff directly into native networking without exposing it in controller status.
func (controller *realmController) CreateGame(options map[string]any) error {
	request, err := realmCreateGameRequest(options)
	if err != nil {
		return err
	}

	return controller.start("creating_game", func(ctx context.Context, client realmAPI) error {
		handoff, err := client.CreateGame(ctx, request)
		if err != nil {
			return err
		}

		return controller.connectGame(ctx, handoff)
	})
}

// realmCreateGameRequest centralizes defaults and integer validation at the Lua boundary. Native
// Realm code can then operate on a fully typed request instead of repeating coercion rules.
func realmCreateGameRequest(options map[string]any) (realm.CreateGameRequest, error) {
	maximum, err := realmOptionInt(options, "maximum_players", 8)
	if err != nil {
		return realm.CreateGameRequest{}, err
	}

	characterDifference, err := realmOptionInt(options, "character_difference", 0)
	if err != nil {
		return realm.CreateGameRequest{}, err
	}

	difficulty := realmOptionString(options, "difficulty")
	if difficulty == "" {
		difficulty = string(realm.DifficultyNormal)
	}

	return realm.CreateGameRequest{
		Name:                realmOptionString(options, "name"),
		Password:            realmOptionString(options, "password"),
		Description:         realmOptionString(options, "description"),
		Difficulty:          realm.GameDifficulty(difficulty),
		Maximum:             maximum,
		CharacterDifference: characterDifference,
		Expansion:           realmOptionBool(options, "expansion", true),
		Hardcore:            realmOptionBool(options, "hardcore", false),
	}, nil
}

// JoinGame lets Realm resolve names and IDs and verify any password before the client receives a
// private worker assignment. The directory reference itself is not treated as authority.
func (controller *realmController) JoinGame(reference, password string) error {
	return controller.start("resolving_game", func(ctx context.Context, client realmAPI) error {
		handoff, err := client.JoinGame(ctx, reference, password)
		if err != nil {
			return err
		}

		return controller.connectGame(ctx, handoff)
	})
}

// connectGame publishes only the public game ID and phase, while the ticket and pinned endpoint move
// directly to native networking. Successful connection is the sole transition to game_connected.
func (controller *realmController) connectGame(ctx context.Context, handoff realm.GameHandoff) error {
	if controller.games == nil || handoff.Assignment.GameID == "" {
		return errors.New("realm client: game assignment is unavailable")
	}

	controller.update(func(state *realmClientState) {
		state.ResolvedGameID = handoff.Game.Entry.GameID
		state.Phase = "game_connecting"
	})

	if err := controller.games.ConnectRealm(ctx, handoff.Assignment); err != nil {
		return err
	}

	controller.update(func(state *realmClientState) {
		state.Phase = "game_connected"
	})

	return nil
}

// ReconnectConnectedGame obtains a fresh ticket for the existing game and delegates the atomic
// endpoint swap to networking. It cannot create or select a different logical game.
func (controller *realmController) ReconnectConnectedGame(ctx context.Context) error {
	if controller == nil || ctx == nil || controller.games == nil {
		return errors.New("realm client: reconnect is unavailable")
	}

	assignment, err := controller.reconnectAssignment(ctx)
	if err != nil {
		return err
	}

	return controller.games.ReconnectRealm(ctx, assignment)
}

// reconnectAssignment rejects a control-plane response that changes GameID, even if its ticket and
// endpoint are otherwise valid. Recovery must preserve the player's current durable session.
func (controller *realmController) reconnectAssignment(ctx context.Context) (realm.JoinAssignment, error) {
	client, gameID := controller.connectedGame()
	if client == nil || gameID == "" {
		return realm.JoinAssignment{}, errors.New("realm client: no connected game")
	}

	handoff, err := client.ReconnectGame(ctx, gameID)
	if err != nil {
		return realm.JoinAssignment{}, err
	}

	if handoff.Assignment.GameID != gameID {
		return realm.JoinAssignment{}, errors.New("realm client: reconnect assignment changed game identity")
	}

	return handoff.Assignment, nil
}

// LeaveConnectedGame asks Realm to commit authority state while the session is still identifiable,
// then replaces cached character views and returns to lobby. Transport teardown happens afterward.
func (controller *realmController) LeaveConnectedGame(ctx context.Context) error {
	if controller == nil || ctx == nil {
		return nil
	}

	client, gameID := controller.connectedGame()
	if client == nil || gameID == "" {
		return nil
	}

	committed, err := client.LeaveGame(ctx, gameID)
	if err != nil {
		return err
	}

	controller.update(func(state *realmClientState) {
		applyCommittedCharacter(state, committed)
		state.ResolvedGameID = ""
		state.Phase = "lobby"
	})

	return nil
}

// connectedGame snapshots the API owner and current durable game ID together so callers cannot pair
// a newly installed client with stale game identity.
func (controller *realmController) connectedGame() (realmAPI, string) {
	controller.mu.RLock()
	defer controller.mu.RUnlock()

	return controller.client, controller.state.ResolvedGameID
}

// applyCommittedCharacter updates both directory and selection copies because the frontend retains
// each independently. Missing either copy would show stale progression or admit an old revision.
func applyCommittedCharacter(state *realmClientState, committed realm.CharacterSummary) {
	for index := range state.Characters {
		if state.Characters[index].Character.ID == committed.Character.ID {
			state.Characters[index] = committed
		}
	}

	if state.Selected.Character.ID == committed.Character.ID {
		state.Selected = committed
	}
}

// realmOptionString accepts only actual strings; silently formatting arbitrary Lua values would hide
// malformed UI requests from validation in the control plane.
func realmOptionString(options map[string]any, name string) string {
	value, _ := options[name].(string)

	return value
}

// realmOptionBool distinguishes omission from an explicit false so defaults remain predictable.
func realmOptionBool(options map[string]any, name string, fallback bool) bool {
	value, found := options[name].(bool)
	if !found {
		return fallback
	}

	return value
}

// realmOptionInt accepts the two integer representations produced by Lua/JSON bridges but rejects
// fractional numbers. Truncation would silently change game admission policy.
func realmOptionInt(options map[string]any, name string, fallback int) (int, error) {
	value, found := options[name]
	if !found {
		return fallback, nil
	}

	switch typed := value.(type) {
	case float64:
		if typed == float64(int(typed)) {
			return int(typed), nil
		}
	case string:
		parsed, err := strconv.Atoi(typed)

		return parsed, err
	}

	return 0, fmt.Errorf("realm client: %s must be an integer", name)
}
