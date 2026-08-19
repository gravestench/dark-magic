package clientapp

import (
	"context"
	"errors"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

const defaultRealmChannel = "Diablo II"

// realmRefreshCursor snapshots the state needed to incrementally refresh the lobby.
type realmRefreshCursor struct {
	channelName   string
	selectedGame  string
	afterSequence uint64
}

// JoinChannel joins a lobby channel and loads its initial events and game directory.
func (controller *realmController) JoinChannel(channel string) error {
	return controller.start("joining_channel", func(ctx context.Context, client realmAPI) error {
		view, events, games, err := loadRealmLobby(ctx, client, channel)
		if err != nil {
			return err
		}

		controller.update(func(state *realmClientState) {
			state.Channel = view
			state.Events = events
			state.Games = games
			state.Phase = "lobby"
		})

		return nil
	})
}

// loadRealmLobby loads a newly joined channel in the order expected by the UI.
func loadRealmLobby(
	ctx context.Context,
	client realmAPI,
	channel string,
) (realm.ChannelView, []realm.ChatEvent, []realm.GameDirectoryEntry, error) {
	view, err := client.JoinChannel(ctx, channel)
	if err != nil {
		return realm.ChannelView{}, nil, nil, err
	}

	events, err := client.ChannelEvents(ctx, 0, 0)
	if err != nil {
		return realm.ChannelView{}, nil, nil, err
	}

	games, err := client.ListGames(ctx)
	if err != nil {
		return realm.ChannelView{}, nil, nil, err
	}

	return view, events, games, nil
}

// SendMessage publishes a chat event and then refreshes the lobby view.
func (controller *realmController) SendMessage(message string) error {
	return controller.start("sending_message", func(ctx context.Context, client realmAPI) error {
		if _, err := client.SendMessage(ctx, message); err != nil {
			return err
		}

		return controller.refresh(ctx, client)
	})
}

// Refresh asynchronously reloads the active lobby's mutable state.
func (controller *realmController) Refresh() error {
	return controller.start("refreshing", controller.refresh)
}

// SelectGame loads a game detail view or clears the current selection.
func (controller *realmController) SelectGame(reference string) error {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		controller.update(func(state *realmClientState) {
			state.SelectedGame = realm.GameDetail{}
		})

		return nil
	}

	return controller.start("loading_game_detail", func(ctx context.Context, client realmAPI) error {
		detail, err := client.GameDetail(ctx, reference)
		if err != nil {
			return err
		}

		controller.update(func(state *realmClientState) {
			state.SelectedGame = detail
			state.Phase = "lobby"
		})

		return nil
	})
}

// refresh reloads channel presence, new chat events, games, and selected-game detail.
func (controller *realmController) refresh(ctx context.Context, client realmAPI) error {
	cursor := controller.refreshCursor()

	view, after, rejoined, err := refreshRealmChannel(ctx, client, cursor)
	if err != nil {
		return err
	}

	games, err := client.ListGames(ctx)
	if err != nil {
		return err
	}

	events, err := client.ChannelEvents(ctx, after, 0)
	if err != nil {
		return err
	}

	controller.applyLobbyRefresh(view, games, events, rejoined)

	return controller.refreshSelectedGame(ctx, client, cursor.selectedGame)
}

// refreshCursor captures the incremental event cursor before issuing network requests.
func (controller *realmController) refreshCursor() realmRefreshCursor {
	controller.mu.RLock()
	defer controller.mu.RUnlock()

	cursor := realmRefreshCursor{
		channelName:  controller.state.Channel.Name,
		selectedGame: controller.state.SelectedGame.Entry.GameID,
	}
	if count := len(controller.state.Events); count > 0 {
		cursor.afterSequence = controller.state.Events[count-1].Sequence
	}

	return cursor
}

// refreshRealmChannel renews presence when server-side pruning removed the member.
func refreshRealmChannel(
	ctx context.Context,
	client realmAPI,
	cursor realmRefreshCursor,
) (realm.ChannelView, uint64, bool, error) {
	view, err := client.Channel(ctx)
	if !errors.Is(err, realm.ErrChannelMember) {
		return view, cursor.afterSequence, false, err
	}

	channelName := cursor.channelName
	if channelName == "" {
		channelName = defaultRealmChannel
	}

	view, err = client.JoinChannel(ctx, channelName)
	if err != nil {
		return realm.ChannelView{}, 0, false, err
	}

	// Rejoining creates a new membership view, so replay events from its start
	// instead of applying the stale membership's sequence cursor.
	return view, 0, true, nil
}

// applyLobbyRefresh atomically publishes a complete channel refresh.
func (controller *realmController) applyLobbyRefresh(
	view realm.ChannelView,
	games []realm.GameDirectoryEntry,
	events []realm.ChatEvent,
	rejoined bool,
) {
	controller.update(func(state *realmClientState) {
		state.Channel = view
		state.Games = games
		state.Phase = "lobby"

		if rejoined {
			state.Events = events

			return
		}

		state.Events = append(state.Events, events...)
	})
}

// refreshSelectedGame keeps the open detail view synchronized with the directory.
func (controller *realmController) refreshSelectedGame(
	ctx context.Context,
	client realmAPI,
	gameID string,
) error {
	if gameID == "" {
		return nil
	}

	detail, err := client.GameDetail(ctx, gameID)
	if err == nil {
		controller.update(func(state *realmClientState) {
			state.SelectedGame = detail
		})

		return nil
	}

	if errors.Is(err, realm.ErrGameNotFound) {
		controller.update(func(state *realmClientState) {
			state.SelectedGame = realm.GameDetail{}
		})

		return nil
	}

	return err
}
