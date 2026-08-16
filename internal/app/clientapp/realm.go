package clientapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

var errRealmBusy = errors.New("realm client: request already in progress")

type realmAPI interface {
	ServiceInfo(context.Context) (realm.ServiceInfo, error)
	Signup(context.Context, string, string, string) (realm.Account, error)
	Authenticate(context.Context, string, string) (realm.RealmSession, error)
	Logout(context.Context) error
	BeginPasswordRecovery(context.Context, string) error
	ListCharacters(context.Context) ([]realm.CharacterSummary, error)
	CreateCharacter(context.Context, realm.CreateCharacterRequest) (realm.CharacterSummary, error)
	DeleteCharacter(context.Context, string) error
	SelectCharacter(context.Context, string) (realm.CharacterSummary, error)
	JoinChannel(context.Context, string) (realm.ChannelView, error)
	Channel(context.Context) (realm.ChannelView, error)
	ChannelEvents(context.Context, uint64, int) ([]realm.ChatEvent, error)
	SendMessage(context.Context, string) (realm.ChatEvent, error)
	ListGames(context.Context) ([]realm.GameDirectoryEntry, error)
	GameDetail(context.Context, string) (realm.GameDetail, error)
	CreateGame(context.Context, realm.CreateGameRequest) (realm.GameHandoff, error)
	ResolveGame(context.Context, string, string) (string, error)
	JoinGame(context.Context, string, string) (realm.GameHandoff, error)
	ReconnectGame(context.Context, string) (realm.GameHandoff, error)
	LeaveGame(context.Context, string) (realm.CharacterSummary, error)
}

type realmGameConnector interface {
	ConnectRealm(context.Context, realm.JoinAssignment) error
	ReconnectRealm(context.Context, realm.JoinAssignment) error
}

type realmClientState struct {
	Phase          string                     `json:"phase"`
	Endpoint       string                     `json:"endpoint"`
	Gateway        string                     `json:"gateway"`
	Error          string                     `json:"error,omitempty"`
	Account        realm.Account              `json:"account"`
	Characters     []realm.CharacterSummary   `json:"characters"`
	Selected       realm.CharacterSummary     `json:"selected"`
	Channel        realm.ChannelView          `json:"channel"`
	Events         []realm.ChatEvent          `json:"events"`
	Games          []realm.GameDirectoryEntry `json:"games"`
	SelectedGame   realm.GameDetail           `json:"selected_game"`
	ResolvedGameID string                     `json:"resolved_game_id,omitempty"`
}

type realmController struct {
	mu     sync.RWMutex
	app    *application
	client realmAPI
	games  realmGameConnector
	state  realmClientState
	cancel context.CancelFunc
}

func newRealmController(app *application) *realmController {
	gateway := "127.0.0.1"
	if app != nil && app.gameSettings != nil {
		gateway = app.gameSettings.Values().RealmGateway
	}
	controller := &realmController{app: app, state: realmClientState{Phase: "disconnected", Gateway: gateway}}
	if app != nil {
		controller.games = app.network
	}
	return controller
}

func (controller *realmController) Connect(endpoint string) error {
	controller.mu.RLock()
	gateway := controller.state.Gateway
	busy := controller.cancel != nil
	controller.mu.RUnlock()
	if busy {
		return errRealmBusy
	}
	if strings.TrimSpace(endpoint) == "" {
		endpoint = gateway
	}
	gateway = strings.TrimSpace(endpoint)
	normalized, address, err := normalizeRealmEndpoint(endpoint)
	if err != nil {
		return err
	}
	if controller.app == nil || controller.app.networkTrust == nil {
		return errors.New("realm client: network trust is not configured")
	}
	tlsConfig, err := controller.app.networkTrust.ClientTLS(address)
	if err != nil {
		return err
	}
	// Worker allocation includes deterministic world preparation and can take
	// longer than an ordinary lobby request on a cold machine. Cancellation is
	// still owned by the scene/app context; this cap prevents an abandoned
	// request from living forever without imposing a best-case startup budget.
	client, err := realm.NewRealmClient(normalized, &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}, Timeout: 60 * time.Second})
	if err != nil {
		return err
	}
	controller.mu.Lock()
	if controller.cancel != nil {
		controller.mu.Unlock()
		return errRealmBusy
	}
	ctx, cancel := context.WithCancel(controller.app.ctx)
	controller.client = client
	controller.state.Endpoint = normalized
	controller.state.Gateway = gateway
	controller.state.Account = realm.Account{}
	controller.state.Characters = nil
	controller.state.Selected = realm.CharacterSummary{}
	controller.cancel, controller.state.Phase, controller.state.Error = cancel, "checking_versions", ""
	controller.mu.Unlock()
	go controller.finish(ctx, "checking_versions", client, func(ctx context.Context, client realmAPI) error {
		info, err := client.ServiceInfo(ctx)
		if err != nil {
			return fmt.Errorf("contacting realm: %w", err)
		}
		if info.Version != realm.RealmControlPlaneVersion {
			return fmt.Errorf("realm version %q is incompatible with %q", info.Version, realm.RealmControlPlaneVersion)
		}
		controller.update(func(state *realmClientState) { state.Phase = "login" })
		return nil
	})
	return nil
}

// Login always establishes a new Realm session from credentials entered by the
// player. Email verification and browser account authorization never act as an
// implicit game login, and no password is copied into Realm status.
func (controller *realmController) Login(name, password string) error {
	return controller.start("authenticating_account", func(ctx context.Context, client realmAPI) error {
		session, err := client.Authenticate(ctx, name, password)
		if err != nil {
			return fmt.Errorf("authenticate Realm account: %w", err)
		}
		controller.update(func(state *realmClientState) { state.Phase = "loading_characters" })
		characters, err := client.ListCharacters(ctx)
		if err != nil {
			return fmt.Errorf("list Realm characters: %w", err)
		}
		controller.update(func(state *realmClientState) {
			state.Account = session.Account
			state.Characters = characters
			state.Phase = "characters"
		})
		return nil
	})
}

func (controller *realmController) Signup(name, email, password string) error {
	return controller.start("creating_account", func(ctx context.Context, client realmAPI) error {
		account, err := client.Signup(ctx, name, email, password)
		if err != nil {
			return err
		}
		controller.update(func(state *realmClientState) {
			state.Account = account
			state.Phase = "verification_required"
		})
		return nil
	})
}

func (controller *realmController) RecoverPassword(email string) error {
	return controller.start("requesting_password_recovery", func(ctx context.Context, client realmAPI) error {
		if err := client.BeginPasswordRecovery(ctx, email); err != nil {
			return err
		}
		controller.update(func(state *realmClientState) { state.Phase = "recovery_sent" })
		return nil
	})
}

// Logout explicitly removes channel presence and invalidates the bearer
// session. Merely navigating away from the lobby must never leave a character
// looking online until the session-expiry maintenance pass.
func (controller *realmController) Logout() error {
	return controller.start("logging_out", func(ctx context.Context, client realmAPI) error {
		if err := client.Logout(ctx); err != nil {
			return err
		}
		controller.update(func(state *realmClientState) {
			gateway, endpoint := state.Gateway, state.Endpoint
			*state = realmClientState{Phase: "login", Gateway: gateway, Endpoint: endpoint}
		})
		return nil
	})
}

// Close is the process-lifetime counterpart to Logout. It is synchronous so
// window close and SIGTERM get a bounded chance to clear live Realm presence.
func (controller *realmController) Close(ctx context.Context) error {
	if controller == nil || ctx == nil {
		return nil
	}
	controller.mu.Lock()
	if controller.cancel != nil {
		controller.cancel()
		controller.cancel = nil
	}
	client := controller.client
	authenticated := controller.state.Account.ID != ""
	controller.mu.Unlock()
	if client == nil || !authenticated {
		return nil
	}
	if err := client.Logout(ctx); err != nil {
		return err
	}
	controller.update(func(state *realmClientState) {
		gateway := state.Gateway
		*state = realmClientState{Phase: "disconnected", Gateway: gateway}
	})
	return nil
}

func (controller *realmController) SetGateway(endpoint string) error {
	_, _, err := normalizeRealmEndpoint(endpoint)
	if err != nil {
		return err
	}
	gateway := strings.TrimSpace(endpoint)
	controller.mu.Lock()
	controller.state.Gateway = gateway
	controller.mu.Unlock()
	if controller.app == nil || controller.app.gameSettings == nil {
		return nil
	}
	values := controller.app.gameSettings.Values()
	values.RealmGateway = gateway
	if err := controller.app.gameSettings.Update(values); err != nil {
		return err
	}
	if controller.app.gameSettings.Path() == "" {
		return nil
	}
	return controller.app.gameSettings.Save()
}

func normalizeRealmEndpoint(endpoint string) (string, string, error) {
	value := strings.TrimSpace(endpoint)
	if value == "" {
		return "", "", realm.ErrHTTPInput
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.Path != "" && parsed.Path != "/" {
		return "", "", realm.ErrHTTPInput
	}
	if parsed.Port() == "" {
		parsed.Host = parsed.Host + ":6112"
	}
	parsed.Path, parsed.RawQuery, parsed.Fragment = "", "", ""
	return strings.TrimRight(parsed.String(), "/"), parsed.Host, nil
}

func (controller *realmController) CreateCharacter(name, class string, expansion, hardcore bool) error {
	return controller.start("creating_character", func(ctx context.Context, client realmAPI) error {
		record, err := client.CreateCharacter(ctx, realm.CreateCharacterRequest{Name: name, Class: class, Expansion: expansion, Hardcore: hardcore})
		if err != nil {
			return err
		}
		characters, err := client.ListCharacters(ctx)
		if err != nil {
			return err
		}
		controller.update(func(state *realmClientState) {
			state.Characters, state.Selected, state.Phase = characters, record, "characters"
		})
		return nil
	})
}

func (controller *realmController) DeleteCharacter(id string) error {
	return controller.start("deleting_character", func(ctx context.Context, client realmAPI) error {
		if err := client.DeleteCharacter(ctx, id); err != nil {
			return err
		}
		characters, err := client.ListCharacters(ctx)
		if err != nil {
			return err
		}
		controller.update(func(state *realmClientState) {
			state.Characters, state.Selected, state.Phase = characters, realm.CharacterSummary{}, "characters"
		})
		return nil
	})
}

func (controller *realmController) SelectCharacter(id string) error {
	return controller.start("selecting_character", func(ctx context.Context, client realmAPI) error {
		record, err := client.SelectCharacter(ctx, id)
		if err == nil {
			controller.update(func(state *realmClientState) { state.Selected, state.Phase = record, "character_selected" })
		}
		return err
	})
}

func (controller *realmController) JoinChannel(channel string) error {
	return controller.start("joining_channel", func(ctx context.Context, client realmAPI) error {
		view, err := client.JoinChannel(ctx, channel)
		if err != nil {
			return err
		}
		events, err := client.ChannelEvents(ctx, 0, 0)
		if err != nil {
			return err
		}
		games, err := client.ListGames(ctx)
		if err != nil {
			return err
		}
		controller.update(func(state *realmClientState) {
			state.Channel, state.Events, state.Games, state.Phase = view, events, games, "lobby"
		})
		return nil
	})
}

func (controller *realmController) SendMessage(message string) error {
	return controller.start("sending_message", func(ctx context.Context, client realmAPI) error {
		if _, err := client.SendMessage(ctx, message); err != nil {
			return err
		}
		return controller.refresh(ctx, client)
	})
}

func (controller *realmController) Refresh() error {
	return controller.start("refreshing", controller.refresh)
}

func (controller *realmController) SelectGame(reference string) error {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		controller.update(func(state *realmClientState) { state.SelectedGame = realm.GameDetail{} })
		return nil
	}
	return controller.start("loading_game_detail", func(ctx context.Context, client realmAPI) error {
		detail, err := client.GameDetail(ctx, reference)
		if err != nil {
			return err
		}
		controller.update(func(state *realmClientState) {
			state.SelectedGame, state.Phase = detail, "lobby"
		})
		return nil
	})
}

func (controller *realmController) refresh(ctx context.Context, client realmAPI) error {
	controller.mu.RLock()
	after := uint64(0)
	channelName := controller.state.Channel.Name
	selectedGameID := controller.state.SelectedGame.Entry.GameID
	if count := len(controller.state.Events); count > 0 {
		after = controller.state.Events[count-1].Sequence
	}
	controller.mu.RUnlock()
	rejoined := false
	view, err := client.Channel(ctx)
	if errors.Is(err, realm.ErrChannelMember) {
		if channelName == "" {
			channelName = "Diablo II"
		}
		view, err = client.JoinChannel(ctx, channelName)
		after = 0
		rejoined = err == nil
	}
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
	controller.update(func(state *realmClientState) {
		state.Channel, state.Games, state.Phase = view, games, "lobby"
		if rejoined {
			state.Events = events
		} else {
			state.Events = append(state.Events, events...)
		}
	})
	if selectedGameID != "" {
		detail, detailErr := client.GameDetail(ctx, selectedGameID)
		if detailErr == nil {
			controller.update(func(state *realmClientState) { state.SelectedGame = detail })
		} else if errors.Is(detailErr, realm.ErrGameNotFound) {
			controller.update(func(state *realmClientState) { state.SelectedGame = realm.GameDetail{} })
		} else {
			return detailErr
		}
	}
	return nil
}

func (controller *realmController) CreateGame(options map[string]any) error {
	maximum, err := realmOptionInt(options, "maximum_players", 8)
	if err != nil {
		return err
	}
	characterDifference, err := realmOptionInt(options, "character_difference", 0)
	if err != nil {
		return err
	}
	difficulty := realmOptionString(options, "difficulty")
	if difficulty == "" {
		difficulty = string(realm.DifficultyNormal)
	}
	request := realm.CreateGameRequest{Name: realmOptionString(options, "name"), Password: realmOptionString(options, "password"),
		Description: realmOptionString(options, "description"), Difficulty: realm.GameDifficulty(difficulty),
		Maximum: maximum, CharacterDifference: characterDifference,
		Expansion: realmOptionBool(options, "expansion", true), Hardcore: realmOptionBool(options, "hardcore", false)}
	return controller.start("creating_game", func(ctx context.Context, client realmAPI) error {
		handoff, err := client.CreateGame(ctx, request)
		if err != nil {
			return err
		}
		return controller.connectGame(ctx, handoff)
	})
}

func (controller *realmController) JoinGame(reference, password string) error {
	return controller.start("resolving_game", func(ctx context.Context, client realmAPI) error {
		handoff, err := client.JoinGame(ctx, reference, password)
		if err != nil {
			return err
		}
		return controller.connectGame(ctx, handoff)
	})
}

func (controller *realmController) connectGame(ctx context.Context, handoff realm.GameHandoff) error {
	if controller.games == nil || handoff.Assignment.GameID == "" {
		return errors.New("realm client: game assignment is unavailable")
	}
	controller.update(func(state *realmClientState) {
		state.ResolvedGameID, state.Phase = handoff.Game.Entry.GameID, "game_connecting"
	})
	if err := controller.games.ConnectRealm(ctx, handoff.Assignment); err != nil {
		return err
	}
	controller.update(func(state *realmClientState) { state.Phase = "game_connected" })
	return nil
}

// ReconnectConnectedGame obtains a fresh assignment after the original worker
// endpoint can no longer honor its transport credential. It is native-only;
// Lua never receives the ticket or worker address.
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

func (controller *realmController) reconnectAssignment(ctx context.Context) (realm.JoinAssignment, error) {
	controller.mu.RLock()
	client, gameID := controller.client, controller.state.ResolvedGameID
	controller.mu.RUnlock()
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

// LeaveConnectedGame is called by the native network lifecycle before its
// transport closes. Lua can request navigation, but never supplies a player ID,
// lease, revision, or replacement character record.
func (controller *realmController) LeaveConnectedGame(ctx context.Context) error {
	if controller == nil || ctx == nil {
		return nil
	}
	controller.mu.RLock()
	client, gameID := controller.client, controller.state.ResolvedGameID
	controller.mu.RUnlock()
	if client == nil || gameID == "" {
		return nil
	}
	committed, err := client.LeaveGame(ctx, gameID)
	if err != nil {
		return err
	}
	controller.update(func(state *realmClientState) {
		for index := range state.Characters {
			if state.Characters[index].Character.ID == committed.Character.ID {
				state.Characters[index] = committed
			}
		}
		if state.Selected.Character.ID == committed.Character.ID {
			state.Selected = committed
		}
		state.ResolvedGameID, state.Phase = "", "lobby"
	})
	return nil
}

func (controller *realmController) start(phase string, operation func(context.Context, realmAPI) error) error {
	controller.mu.Lock()
	if controller.cancel != nil {
		controller.mu.Unlock()
		return errRealmBusy
	}
	if controller.client == nil {
		controller.mu.Unlock()
		return errors.New("realm client: endpoint is not configured")
	}
	ctx, cancel := context.WithCancel(controller.app.ctx)
	controller.cancel, controller.state.Phase, controller.state.Error = cancel, phase, ""
	client := controller.client
	controller.mu.Unlock()
	go controller.finish(ctx, phase, client, operation)
	return nil
}

func (controller *realmController) finish(ctx context.Context, phase string, client realmAPI, operation func(context.Context, realmAPI) error) {
	err := operation(ctx, client)
	controller.mu.Lock()
	if err != nil && !errors.Is(err, context.Canceled) {
		slog.Debug("realm request failed", "phase", phase, "error", err)
		controller.state.Phase, controller.state.Error = "error", err.Error()
	}
	controller.cancel = nil
	controller.mu.Unlock()
}

func (controller *realmController) update(update func(*realmClientState)) {
	controller.mu.Lock()
	update(&controller.state)
	controller.state.Error = ""
	controller.mu.Unlock()
}
func (controller *realmController) Cancel() {
	controller.mu.Lock()
	if controller.cancel != nil {
		controller.cancel()
	}
	controller.cancel = nil
	if controller.client == nil {
		controller.state.Phase = "disconnected"
	} else {
		controller.state.Phase = "ready"
	}
	controller.mu.Unlock()
}

func (controller *realmController) Status() map[string]any {
	controller.mu.RLock()
	state := controller.state
	controller.mu.RUnlock()
	data, _ := json.Marshal(state)
	var result map[string]any
	_ = json.Unmarshal(data, &result)
	return result
}

func realmOptionString(options map[string]any, name string) string {
	value, _ := options[name].(string)
	return value
}
func realmOptionBool(options map[string]any, name string, fallback bool) bool {
	value, found := options[name].(bool)
	if !found {
		return fallback
	}
	return value
}
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
