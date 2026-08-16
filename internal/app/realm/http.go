package realm

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const maximumRealmRequestBytes = 32 << 10

type HTTPServer struct {
	control       *ControlPlane
	operatorToken string
}

func NewHTTPHandler(control *ControlPlane) (http.Handler, error) {
	if control == nil {
		return nil, errors.New("realm: HTTP handler requires a control plane")
	}
	server := &HTTPServer{control: control}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/status", server.status)
	mux.HandleFunc("POST /v1/accounts", server.createAccount)
	mux.HandleFunc("POST /v1/accounts/verify", server.verifyAccount)
	mux.HandleFunc("POST /v1/accounts/recovery", server.beginAccountRecovery)
	mux.HandleFunc("POST /v1/accounts/recovery/complete", server.completeAccountRecovery)
	mux.HandleFunc("POST /v1/sessions", server.authenticate)
	mux.HandleFunc("DELETE /v1/session", server.logout)
	mux.HandleFunc("GET /v1/characters", server.listCharacters)
	mux.HandleFunc("POST /v1/characters", server.createCharacter)
	mux.HandleFunc("DELETE /v1/characters", server.deleteCharacter)
	mux.HandleFunc("POST /v1/characters/select", server.selectCharacter)
	mux.HandleFunc("GET /v1/characters/selected", server.selectedCharacter)
	mux.HandleFunc("POST /v1/channels/join", server.joinChannel)
	mux.HandleFunc("GET /v1/channel", server.channel)
	mux.HandleFunc("GET /v1/channel/events", server.channelEvents)
	mux.HandleFunc("POST /v1/channel/messages", server.sendMessage)
	mux.HandleFunc("GET /v1/games", server.listGames)
	mux.HandleFunc("POST /v1/games", server.createGame)
	mux.HandleFunc("GET /v1/games/detail", server.gameDetail)
	mux.HandleFunc("POST /v1/games/resolve", server.resolveGame)
	mux.HandleFunc("POST /v1/games/join", server.joinGame)
	mux.HandleFunc("POST /v1/games/reconnect", server.reconnectGame)
	mux.HandleFunc("POST /v1/games/leave", server.leaveGame)
	withAuditContext := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := withAuditClientAddress(request.Context(), request.RemoteAddr)
		mux.ServeHTTP(writer, request.WithContext(ctx))
	})
	return http.MaxBytesHandler(withAuditContext, maximumRealmRequestBytes), nil
}

// NewOperatorHTTPHandler is deliberately separate from the player/portal API
// so deployments can bind it to loopback or a private service network.
func NewOperatorHTTPHandler(control *ControlPlane, operatorToken string) (http.Handler, error) {
	operatorToken = strings.TrimSpace(operatorToken)
	if control == nil || len(operatorToken) < 32 {
		return nil, errors.New("realm: operator HTTP handler requires a control plane and 32-byte token")
	}
	server := &HTTPServer{control: control, operatorToken: operatorToken}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/operator/games/drain", server.operatorDrainGame)
	withAuditContext := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := withAuditClientAddress(request.Context(), request.RemoteAddr)
		mux.ServeHTTP(writer, request.WithContext(ctx))
	})
	return http.MaxBytesHandler(withAuditContext, maximumRealmRequestBytes), nil
}

func (server *HTTPServer) operatorDrainGame(writer http.ResponseWriter, request *http.Request) {
	if !validOperatorBearer(request.Header.Get("Authorization"), server.operatorToken) {
		writeResponse(writer, nil, ErrOperatorAuthentication)
		return
	}
	var input struct {
		GameID string `json:"game_id"`
	}
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.DrainGame(request.Context(), input.GameID)
	writeResponse(writer, value, err)
}

func validOperatorBearer(header, token string) bool {
	scheme, candidate, found := strings.Cut(header, " ")
	return found && token != "" && strings.EqualFold(scheme, "Bearer") &&
		hmac.Equal([]byte(strings.TrimSpace(candidate)), []byte(token))
}

// ServiceInfo is the small unauthenticated compatibility response used before
// the client associates its private local identity. It exposes no realm state
// or secrets.
type ServiceInfo struct {
	Version string `json:"version"`
}

func (server *HTTPServer) status(writer http.ResponseWriter, _ *http.Request) {
	writeResponse(writer, ServiceInfo{Version: server.control.Version()}, nil)
}

func (server *HTTPServer) createAccount(writer http.ResponseWriter, request *http.Request) {
	var input struct{ Name, Email, Password string }
	if !decodeRequest(writer, request, &input) {
		return
	}
	var value Account
	var err error
	if server.control.accountLifecycle != nil {
		value, err = server.control.Signup(request.Context(), SignupRequest{Name: input.Name, Email: input.Email, Password: input.Password})
	} else {
		value, err = server.control.CreateAccount(request.Context(), input.Name, input.Password)
	}
	writeResponse(writer, value, err)
}

func (server *HTTPServer) verifyAccount(writer http.ResponseWriter, request *http.Request) {
	var input struct{ Token string }
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.VerifyEmail(request.Context(), input.Token)
	writeResponse(writer, value, err)
}

func (server *HTTPServer) beginAccountRecovery(writer http.ResponseWriter, request *http.Request) {
	var input struct{ Email string }
	if !decodeRequest(writer, request, &input) {
		return
	}
	writeResponse(writer, struct{}{}, server.control.BeginPasswordRecovery(request.Context(), input.Email))
}

func (server *HTTPServer) completeAccountRecovery(writer http.ResponseWriter, request *http.Request) {
	var input struct{ Token, Password string }
	if !decodeRequest(writer, request, &input) {
		return
	}
	writeResponse(writer, struct{}{}, server.control.CompletePasswordRecovery(request.Context(), input.Token, input.Password))
}

func (server *HTTPServer) authenticate(writer http.ResponseWriter, request *http.Request) {
	var input struct{ Name, Password string }
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.Authenticate(request.Context(), input.Name, input.Password)
	writeResponse(writer, value, err)
}

func (server *HTTPServer) logout(writer http.ResponseWriter, request *http.Request) {
	writeResponse(writer, struct{}{}, server.control.Logout(request.Context(), bearerToken(request)))
}

func (server *HTTPServer) listCharacters(writer http.ResponseWriter, request *http.Request) {
	value, err := server.control.ListCharacters(request.Context(), bearerToken(request))
	result := make([]CharacterSummary, len(value))
	for index, record := range value {
		result[index] = publicCharacter(record)
	}
	writeResponse(writer, result, err)
}

func (server *HTTPServer) createCharacter(writer http.ResponseWriter, request *http.Request) {
	var input CreateCharacterRequest
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.CreateCharacter(request.Context(), bearerToken(request), input)
	writeResponse(writer, publicCharacter(value), err)
}

func (server *HTTPServer) deleteCharacter(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		CharacterID string `json:"character_id"`
	}
	if !decodeRequest(writer, request, &input) {
		return
	}
	writeResponse(writer, struct{}{}, server.control.DeleteCharacter(request.Context(), bearerToken(request), input.CharacterID))
}

func (server *HTTPServer) selectCharacter(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		CharacterID string `json:"character_id"`
	}
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.SelectCharacter(request.Context(), bearerToken(request), input.CharacterID)
	writeResponse(writer, publicCharacter(value), err)
}

func (server *HTTPServer) selectedCharacter(writer http.ResponseWriter, request *http.Request) {
	value, err := server.control.SelectedCharacter(request.Context(), bearerToken(request))
	writeResponse(writer, publicCharacter(value), err)
}

func (server *HTTPServer) joinChannel(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Channel string `json:"channel"`
	}
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.JoinChannel(request.Context(), bearerToken(request), input.Channel)
	writeResponse(writer, value, err)
}

func (server *HTTPServer) channel(writer http.ResponseWriter, request *http.Request) {
	value, err := server.control.ChannelView(request.Context(), bearerToken(request))
	writeResponse(writer, value, err)
}

func (server *HTTPServer) channelEvents(writer http.ResponseWriter, request *http.Request) {
	var after uint64
	var limit int
	if _, err := fmt.Sscan(request.URL.Query().Get("after"), &after); request.URL.Query().Get("after") != "" && err != nil {
		writeResponse(writer, nil, ErrHTTPInput)
		return
	}
	if _, err := fmt.Sscan(request.URL.Query().Get("limit"), &limit); request.URL.Query().Get("limit") != "" && err != nil {
		writeResponse(writer, nil, ErrHTTPInput)
		return
	}
	value, err := server.control.ChannelEvents(request.Context(), bearerToken(request), after, limit)
	writeResponse(writer, value, err)
}

func (server *HTTPServer) sendMessage(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Message string `json:"message"`
	}
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.SendChannelMessage(request.Context(), bearerToken(request), input.Message)
	writeResponse(writer, value, err)
}

func (server *HTTPServer) listGames(writer http.ResponseWriter, request *http.Request) {
	value, err := server.control.ListGames(request.Context(), bearerToken(request), GameFilter{})
	writeResponse(writer, value, err)
}

func (server *HTTPServer) createGame(writer http.ResponseWriter, request *http.Request) {
	var input CreateGameRequest
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.CreateGame(request.Context(), bearerToken(request), input)
	writeResponse(writer, value, err)
}

func (server *HTTPServer) gameDetail(writer http.ResponseWriter, request *http.Request) {
	value, err := server.control.GameDetail(request.Context(), bearerToken(request), request.URL.Query().Get("reference"))
	writeResponse(writer, value, err)
}

func (server *HTTPServer) resolveGame(writer http.ResponseWriter, request *http.Request) {
	var input struct{ Reference, Password string }
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.ResolveGameJoin(request.Context(), bearerToken(request), input.Reference, input.Password)
	writeResponse(writer, map[string]string{"game_id": value}, err)
}

func (server *HTTPServer) joinGame(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		Reference string `json:"reference"`
		Password  string `json:"password"`
	}
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.JoinGame(request.Context(), bearerToken(request), input.Reference, input.Password)
	writeResponse(writer, value, err)
}

func (server *HTTPServer) reconnectGame(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		GameID string `json:"game_id"`
	}
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.ReconnectGame(request.Context(), bearerToken(request), input.GameID)
	writeResponse(writer, value, err)
}

func (server *HTTPServer) leaveGame(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		GameID string `json:"game_id"`
	}
	if !decodeRequest(writer, request, &input) {
		return
	}
	value, err := server.control.LeaveGame(request.Context(), bearerToken(request), input.GameID)
	writeResponse(writer, publicCharacter(value), err)
}

var ErrHTTPInput = errors.New("realm: malformed request")

type httpEnvelope struct {
	Data  json.RawMessage `json:"data,omitempty"`
	Error *realmHTTPError `json:"error,omitempty"`
}

type realmHTTPError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func decodeRequest(writer http.ResponseWriter, request *http.Request, destination any) bool {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		writeResponse(writer, nil, ErrHTTPInput)
		return false
	}
	return true
}

func bearerToken(request *http.Request) string {
	scheme, token, found := strings.Cut(request.Header.Get("Authorization"), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func writeResponse(writer http.ResponseWriter, value any, err error) {
	writer.Header().Set("Content-Type", "application/json")
	if err != nil {
		status, code := realmHTTPStatus(err)
		writer.WriteHeader(status)
		_ = json.NewEncoder(writer).Encode(map[string]any{"error": realmHTTPError{Code: code, Message: code}})
		return
	}
	_ = json.NewEncoder(writer).Encode(map[string]any{"data": value})
}

func realmHTTPStatus(err error) (int, string) {
	switch {
	case errors.Is(err, ErrRealmSession), errors.Is(err, ErrAccountCredentials), errors.Is(err, ErrOperatorAuthentication):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, ErrChannelMember):
		return http.StatusConflict, "channel_membership_required"
	case errors.Is(err, ErrAccountUnverified):
		return http.StatusForbidden, "email_unverified"
	case errors.Is(err, ErrAccountChallenge):
		return http.StatusBadRequest, "invalid_challenge"
	case errors.Is(err, ErrAccountExists), errors.Is(err, ErrCharacterExists), errors.Is(err, ErrGameExists):
		return http.StatusConflict, "already_exists"
	case errors.Is(err, ErrCharacterNotFound), errors.Is(err, ErrGameNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, ErrLease):
		return http.StatusConflict, "not_in_game"
	case errors.Is(err, ErrGameFull), errors.Is(err, ErrCharacterLimit):
		return http.StatusConflict, "capacity"
	case errors.Is(err, ErrGamePassword):
		return http.StatusForbidden, "invalid_password"
	case errors.Is(err, ErrGameLevelRange):
		return http.StatusForbidden, "level_restricted"
	case errors.Is(err, ErrGameUnavailable), errors.Is(err, ErrWorker), errors.Is(err, ErrAdmission):
		return http.StatusServiceUnavailable, "unavailable"
	case errors.Is(err, ErrAccountInput), errors.Is(err, ErrCharacterInput), errors.Is(err, ErrGameDirectoryInput), errors.Is(err, ErrHTTPInput):
		return http.StatusBadRequest, "invalid_input"
	default:
		return http.StatusInternalServerError, "internal"
	}
}

// RealmClient is the typed transport used by presentation controllers. It
// owns the bearer token and never exposes it through Lua snapshots.
type RealmClient struct {
	base  *url.URL
	http  *http.Client
	token string
}

func NewRealmClient(endpoint string, client *http.Client) (*RealmClient, error) {
	base, err := url.Parse(strings.TrimRight(strings.TrimSpace(endpoint), "/"))
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return nil, ErrHTTPInput
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &RealmClient{base: base, http: client}, nil
}

func (client *RealmClient) ServiceInfo(ctx context.Context) (ServiceInfo, error) {
	var result ServiceInfo
	err := client.call(ctx, http.MethodGet, "/v1/status", nil, &result)
	return result, err
}

func (client *RealmClient) CreateAccount(ctx context.Context, name, password string) (Account, error) {
	var result Account
	err := client.call(ctx, http.MethodPost, "/v1/accounts", map[string]string{"name": name, "password": password}, &result)
	return result, err
}

func (client *RealmClient) Signup(ctx context.Context, name, email, password string) (Account, error) {
	var result Account
	err := client.call(ctx, http.MethodPost, "/v1/accounts",
		map[string]string{"name": name, "email": email, "password": password}, &result)
	return result, err
}

func (client *RealmClient) VerifyEmail(ctx context.Context, token string) (Account, error) {
	var result Account
	err := client.call(ctx, http.MethodPost, "/v1/accounts/verify", map[string]string{"token": token}, &result)
	return result, err
}

func (client *RealmClient) BeginPasswordRecovery(ctx context.Context, email string) error {
	return client.call(ctx, http.MethodPost, "/v1/accounts/recovery", map[string]string{"email": email}, nil)
}

func (client *RealmClient) CompletePasswordRecovery(ctx context.Context, token, password string) error {
	return client.call(ctx, http.MethodPost, "/v1/accounts/recovery/complete",
		map[string]string{"token": token, "password": password}, nil)
}

func (client *RealmClient) Authenticate(ctx context.Context, name, password string) (RealmSession, error) {
	var result RealmSession
	if err := client.call(ctx, http.MethodPost, "/v1/sessions", map[string]string{"name": name, "password": password}, &result); err != nil {
		return RealmSession{}, err
	}
	client.token = result.Token
	return result, nil
}
func (client *RealmClient) Logout(ctx context.Context) error {
	if client.token == "" {
		return nil
	}
	if err := client.call(ctx, http.MethodDelete, "/v1/session", nil, nil); err != nil {
		return err
	}
	client.token = ""
	return nil
}
func (client *RealmClient) ListCharacters(ctx context.Context) ([]CharacterSummary, error) {
	var result []CharacterSummary
	err := client.call(ctx, http.MethodGet, "/v1/characters", nil, &result)
	return result, err
}
func (client *RealmClient) CreateCharacter(ctx context.Context, request CreateCharacterRequest) (CharacterSummary, error) {
	var result CharacterSummary
	err := client.call(ctx, http.MethodPost, "/v1/characters", request, &result)
	return result, err
}
func (client *RealmClient) DeleteCharacter(ctx context.Context, id string) error {
	return client.call(ctx, http.MethodDelete, "/v1/characters", map[string]string{"character_id": id}, nil)
}
func (client *RealmClient) SelectCharacter(ctx context.Context, id string) (CharacterSummary, error) {
	var result CharacterSummary
	err := client.call(ctx, http.MethodPost, "/v1/characters/select", map[string]string{"character_id": id}, &result)
	return result, err
}
func (client *RealmClient) JoinChannel(ctx context.Context, channel string) (ChannelView, error) {
	var result ChannelView
	err := client.call(ctx, http.MethodPost, "/v1/channels/join", map[string]string{"channel": channel}, &result)
	return result, err
}
func (client *RealmClient) Channel(ctx context.Context) (ChannelView, error) {
	var result ChannelView
	err := client.call(ctx, http.MethodGet, "/v1/channel", nil, &result)
	return result, err
}
func (client *RealmClient) ChannelEvents(ctx context.Context, after uint64, limit int) ([]ChatEvent, error) {
	var result []ChatEvent
	query := url.Values{"after": {strconv.FormatUint(after, 10)}, "limit": {strconv.Itoa(limit)}}
	err := client.call(ctx, http.MethodGet, "/v1/channel/events?"+query.Encode(), nil, &result)
	return result, err
}
func (client *RealmClient) SendMessage(ctx context.Context, message string) (ChatEvent, error) {
	var result ChatEvent
	err := client.call(ctx, http.MethodPost, "/v1/channel/messages", map[string]string{"message": message}, &result)
	return result, err
}
func (client *RealmClient) ListGames(ctx context.Context) ([]GameDirectoryEntry, error) {
	var result []GameDirectoryEntry
	err := client.call(ctx, http.MethodGet, "/v1/games", nil, &result)
	return result, err
}
func (client *RealmClient) GameDetail(ctx context.Context, reference string) (GameDetail, error) {
	var result GameDetail
	query := url.Values{"reference": {reference}}
	err := client.call(ctx, http.MethodGet, "/v1/games/detail?"+query.Encode(), nil, &result)
	return result, err
}
func (client *RealmClient) CreateGame(ctx context.Context, request CreateGameRequest) (GameHandoff, error) {
	var result GameHandoff
	err := client.call(ctx, http.MethodPost, "/v1/games", request, &result)
	return result, err
}
func (client *RealmClient) ResolveGame(ctx context.Context, reference, password string) (string, error) {
	var result map[string]string
	err := client.call(ctx, http.MethodPost, "/v1/games/resolve", map[string]string{"reference": reference, "password": password}, &result)
	return result["game_id"], err
}

func (client *RealmClient) JoinGame(ctx context.Context, reference, password string) (GameHandoff, error) {
	var result GameHandoff
	err := client.call(ctx, http.MethodPost, "/v1/games/join", map[string]string{"reference": reference, "password": password}, &result)
	return result, err
}

func (client *RealmClient) ReconnectGame(ctx context.Context, gameID string) (GameHandoff, error) {
	var result GameHandoff
	err := client.call(ctx, http.MethodPost, "/v1/games/reconnect", map[string]string{"game_id": gameID}, &result)
	return result, err
}

func (client *RealmClient) LeaveGame(ctx context.Context, gameID string) (CharacterSummary, error) {
	var result CharacterSummary
	err := client.call(ctx, http.MethodPost, "/v1/games/leave", map[string]string{"game_id": gameID}, &result)
	return result, err
}

func (client *RealmClient) call(ctx context.Context, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	relative, err := url.Parse(path)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, method, client.base.ResolveReference(relative).String(), body)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if client.token != "" {
		request.Header.Set("Authorization", "Bearer "+client.token)
	}
	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, maximumRealmRequestBytes+1))
	if err != nil || len(data) > maximumRealmRequestBytes {
		return ErrHTTPInput
	}
	var envelope httpEnvelope
	if json.Unmarshal(data, &envelope) != nil {
		return ErrHTTPInput
	}
	if envelope.Error != nil {
		return realmHTTPCodeError(envelope.Error.Code)
	}
	if output == nil {
		return nil
	}
	return json.Unmarshal(envelope.Data, output)
}

func realmHTTPCodeError(code string) error {
	var cause error
	switch code {
	case "unauthorized":
		cause = ErrAccountCredentials
	case "email_unverified":
		cause = ErrAccountUnverified
	case "invalid_challenge":
		cause = ErrAccountChallenge
	case "already_exists":
		cause = errors.Join(ErrAccountExists, ErrCharacterExists, ErrGameExists)
	case "not_found":
		cause = errors.Join(ErrCharacterNotFound, ErrGameNotFound)
	case "capacity":
		cause = errors.Join(ErrCharacterLimit, ErrGameFull)
	case "invalid_password":
		cause = ErrGamePassword
	case "invalid_input":
		cause = ErrHTTPInput
	case "unavailable":
		cause = ErrGameUnavailable
	case "not_in_game":
		cause = ErrLease
	case "channel_membership_required":
		cause = ErrChannelMember
	default:
		cause = errors.New("realm: request failed")
	}
	return fmt.Errorf("%w: %s", cause, code)
}
