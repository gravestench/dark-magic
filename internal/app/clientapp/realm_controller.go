package clientapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

var errRealmBusy = errors.New("realm client: request already in progress")

// realmAPI describes the control-plane operations used by the client application.
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

// realmGameConnector passes private game assignments to the native network layer.
type realmGameConnector interface {
	ConnectRealm(context.Context, realm.JoinAssignment) error
	ReconnectRealm(context.Context, realm.JoinAssignment) error
}

// realmClientState contains the Realm data that may be exposed to Lua.
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

// realmController owns the active Realm client, its cancellable request, and its presentation state.
type realmController struct {
	mu     sync.RWMutex
	app    *application
	client realmAPI
	games  realmGameConnector
	state  realmClientState
	cancel context.CancelFunc
}

// newRealmController creates a disconnected controller using the configured gateway.
func newRealmController(app *application) *realmController {
	gateway := "127.0.0.1"
	if app != nil && app.gameSettings != nil {
		gateway = app.gameSettings.Values().RealmGateway
	}

	controller := &realmController{
		app:   app,
		state: realmClientState{Phase: "disconnected", Gateway: gateway},
	}
	if app != nil {
		controller.games = app.network
	}

	return controller
}

// Connect configures a Realm client and asynchronously verifies control-plane compatibility.
func (controller *realmController) Connect(endpoint string) error {
	gateway, err := controller.connectionGateway(endpoint)
	if err != nil {
		return err
	}

	normalized, address, err := normalizeRealmEndpoint(gateway)
	if err != nil {
		return err
	}

	client, err := controller.newRealmAPI(normalized, address)
	if err != nil {
		return err
	}

	return controller.beginConnection(client, normalized, gateway)
}

// connectionGateway selects the requested endpoint or falls back to the saved gateway.
func (controller *realmController) connectionGateway(endpoint string) (string, error) {
	controller.mu.RLock()
	gateway := controller.state.Gateway
	busy := controller.cancel != nil
	controller.mu.RUnlock()

	if busy {
		return "", errRealmBusy
	}

	if strings.TrimSpace(endpoint) == "" {
		endpoint = gateway
	}

	return strings.TrimSpace(endpoint), nil
}

// newRealmAPI builds the HTTPS client used for all Realm control-plane requests.
func (controller *realmController) newRealmAPI(endpoint, address string) (realmAPI, error) {
	if controller.app == nil || controller.app.networkTrust == nil {
		return nil, errors.New("realm client: network trust is not configured")
	}

	tlsConfig, err := controller.app.networkTrust.ClientTLS(address)
	if err != nil {
		return nil, err
	}

	// Cold worker allocation includes deterministic world preparation. The request
	// remains app-cancellable, while this cap prevents an abandoned request from
	// living forever without imposing a best-case startup budget.
	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   60 * time.Second,
	}

	return realm.NewRealmClient(endpoint, httpClient)
}

// beginConnection installs a prepared client and starts its compatibility check.
func (controller *realmController) beginConnection(client realmAPI, endpoint, gateway string) error {
	controller.mu.Lock()
	if controller.cancel != nil {
		controller.mu.Unlock()

		return errRealmBusy
	}

	ctx, cancel := context.WithCancel(controller.app.ctx)
	controller.client = client
	controller.state.Endpoint = endpoint
	controller.state.Gateway = gateway
	controller.state.Account = realm.Account{}
	controller.state.Characters = nil
	controller.state.Selected = realm.CharacterSummary{}
	controller.cancel = cancel
	controller.state.Phase = "checking_versions"
	controller.state.Error = ""
	controller.mu.Unlock()

	go controller.finish(ctx, "checking_versions", client, controller.checkRealmCompatibility)

	return nil
}

// checkRealmCompatibility verifies the server protocol before enabling login.
func (controller *realmController) checkRealmCompatibility(ctx context.Context, client realmAPI) error {
	info, err := client.ServiceInfo(ctx)
	if err != nil {
		return fmt.Errorf("contacting realm: %w", err)
	}

	if info.Version != realm.RealmControlPlaneVersion {
		return fmt.Errorf(
			"realm version %q is incompatible with %q",
			info.Version,
			realm.RealmControlPlaneVersion,
		)
	}

	controller.update(func(state *realmClientState) {
		state.Phase = "login"
	})

	return nil
}

// Close synchronously clears live Realm presence during process shutdown.
func (controller *realmController) Close(ctx context.Context) error {
	if controller == nil || ctx == nil {
		return nil
	}

	client, authenticated := controller.prepareToClose()
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

// prepareToClose cancels in-flight work and snapshots the session that needs logout.
func (controller *realmController) prepareToClose() (realmAPI, bool) {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.cancel != nil {
		controller.cancel()
		controller.cancel = nil
	}

	return controller.client, controller.state.Account.ID != ""
}

// SetGateway validates and persists the player's preferred Realm gateway.
func (controller *realmController) SetGateway(endpoint string) error {
	if _, _, err := normalizeRealmEndpoint(endpoint); err != nil {
		return err
	}

	gateway := strings.TrimSpace(endpoint)
	controller.updateGateway(gateway)

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

// updateGateway stores the gateway without clearing an existing error message.
func (controller *realmController) updateGateway(gateway string) {
	controller.mu.Lock()
	controller.state.Gateway = gateway
	controller.mu.Unlock()
}

// normalizeRealmEndpoint converts a gateway into its canonical TLS URL and dial address.
func normalizeRealmEndpoint(endpoint string) (string, string, error) {
	value := strings.TrimSpace(endpoint)
	if value == "" {
		return "", "", realm.ErrHTTPInput
	}

	if !strings.Contains(value, "://") {
		value = "https://" + value
	}

	parsed, err := url.Parse(value)
	if err != nil || !validRealmURL(parsed) {
		return "", "", realm.ErrHTTPInput
	}

	if parsed.Port() == "" {
		parsed.Host += ":6112"
	}

	parsed.Path = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return strings.TrimRight(parsed.String(), "/"), parsed.Host, nil
}

// validRealmURL accepts only HTTPS origins with no application path.
func validRealmURL(parsed *url.URL) bool {
	if parsed == nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return false
	}

	return parsed.Path == "" || parsed.Path == "/"
}

// start reserves the controller and runs one Realm operation asynchronously.
func (controller *realmController) start(
	phase string,
	operation func(context.Context, realmAPI) error,
) error {
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
	client := controller.client
	controller.cancel = cancel
	controller.state.Phase = phase
	controller.state.Error = ""
	controller.mu.Unlock()

	go controller.finish(ctx, phase, client, operation)

	return nil
}

// finish records a background operation's failure and releases the controller.
func (controller *realmController) finish(
	ctx context.Context,
	phase string,
	client realmAPI,
	operation func(context.Context, realmAPI) error,
) {
	err := operation(ctx, client)

	controller.mu.Lock()
	defer controller.mu.Unlock()

	if err != nil && !errors.Is(err, context.Canceled) {
		slog.Debug("realm request failed", "phase", phase, "error", err)

		controller.state.Phase = "error"
		controller.state.Error = err.Error()
	}

	controller.cancel = nil
}

// update applies an atomic state change and clears any prior request error.
func (controller *realmController) update(update func(*realmClientState)) {
	controller.mu.Lock()
	update(&controller.state)
	controller.state.Error = ""
	controller.mu.Unlock()
}

// Cancel stops the active request and returns the controller to an available phase.
func (controller *realmController) Cancel() {
	controller.mu.Lock()
	defer controller.mu.Unlock()

	if controller.cancel != nil {
		controller.cancel()
	}

	controller.cancel = nil

	if controller.client == nil {
		controller.state.Phase = "disconnected"

		return
	}

	controller.state.Phase = "ready"
}

// Status returns a JSON-shaped copy of the Realm state for Lua consumers.
func (controller *realmController) Status() map[string]any {
	controller.mu.RLock()
	state := controller.state
	controller.mu.RUnlock()

	// The JSON round trip deliberately applies the same field names, omissions,
	// and number representation that the Lua bridge observes.
	data, _ := json.Marshal(state)

	var result map[string]any

	_ = json.Unmarshal(data, &result)

	return result
}
