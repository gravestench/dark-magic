package clientapp

import (
	"context"
	"errors"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

const defaultRealmChannel = "Diablo II"

// realmRefreshCursor freezes membership, selection, and event progress before network I/O. Refresh
// results can then be applied consistently even though the controller lock is not held during I/O.
type realmRefreshCursor struct {
	channelName   string
	selectedGame  string
	afterSequence uint64
}

// JoinChannel publishes channel, history, and directory together after all three requests succeed.
// The frontend never observes a joined channel paired with stale lobby data.
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

// loadRealmLobby establishes membership before requesting member-scoped history and directory data.
// Failure returns no partial view for the controller to publish.
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

// SendMessage relies on the subsequent incremental refresh to install the canonical event. Locally
// appending the response could duplicate it when the event stream catches up.
func (controller *realmController) SendMessage(message string) error {
	return controller.start("sending_message", func(ctx context.Context, client realmAPI) error {
		if _, err := client.SendMessage(ctx, message); err != nil {
			return err
		}

		return controller.refresh(ctx, client)
	})
}

// Refresh uses the shared serialized operation runner so periodic UI refresh cannot overlap account,
// channel, or game transitions.
func (controller *realmController) Refresh() error {
	return controller.start("refreshing", controller.refresh)
}

// SelectGame clears locally for an empty reference, otherwise stores only server-resolved detail.
// This keeps player rosters and join metadata tied to a current directory entry.
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

// refresh renews membership if necessary, obtains directory and incremental events, publishes them
// atomically, then synchronizes the optional detail panel against the refreshed directory.
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

// refreshCursor uses the last installed event sequence, not wall-clock time, so refresh neither
// duplicates nor skips events under clock skew.
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

// refreshRealmChannel treats missing membership as recoverable and resets the event cursor after
// rejoin. A cursor from the old membership cannot safely index the new channel view.
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

// applyLobbyRefresh replaces history after rejoin but appends it for normal incremental refresh.
// Both cases publish channel, games, events, and phase inside one state mutation.
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

// refreshSelectedGame refreshes an open detail panel and clears it when the game disappears. Other
// failures remain visible because silently clearing on transport errors would misrepresent state.
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
