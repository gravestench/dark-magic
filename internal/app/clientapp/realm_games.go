package clientapp

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

// CreateGame validates UI options, creates a Realm game, and connects to its worker.
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

// realmCreateGameRequest converts loosely typed Lua options into a Realm request.
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

// JoinGame resolves a directory reference and connects to the assigned worker.
func (controller *realmController) JoinGame(reference, password string) error {
	return controller.start("resolving_game", func(ctx context.Context, client realmAPI) error {
		handoff, err := client.JoinGame(ctx, reference, password)
		if err != nil {
			return err
		}

		return controller.connectGame(ctx, handoff)
	})
}

// connectGame passes a private assignment directly to the native network layer.
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

// ReconnectConnectedGame obtains a fresh private assignment for the active game.
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

// reconnectAssignment renews a worker credential without allowing game identity to change.
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

// LeaveConnectedGame commits the player's final character state before transport close.
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

// connectedGame snapshots the private client and public game identity under one lock.
func (controller *realmController) connectedGame() (realmAPI, string) {
	controller.mu.RLock()
	defer controller.mu.RUnlock()

	return controller.client, controller.state.ResolvedGameID
}

// applyCommittedCharacter replaces every cached view of a character after game exit.
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

// realmOptionString returns a string-valued Lua option or its zero value.
func realmOptionString(options map[string]any, name string) string {
	value, _ := options[name].(string)

	return value
}

// realmOptionBool returns a Boolean-valued Lua option or the supplied fallback.
func realmOptionBool(options map[string]any, name string, fallback bool) bool {
	value, found := options[name].(bool)
	if !found {
		return fallback
	}

	return value
}

// realmOptionInt accepts integral JSON numbers and decimal strings from Lua options.
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
