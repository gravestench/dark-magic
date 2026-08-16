package realm

import (
	"bytes"
	"context"
	"crypto/hmac"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

const (
	WorkerControlProtocolVersion = "RealmWorkerControl/v1"
	maximumWorkerRequestBytes    = 1 << 20
	maximumWorkerResponseBytes   = maximumGameCheckpointBytes + (1 << 20)
	maximumWorkerTicketLifetime  = 5 * time.Minute
)

var (
	ErrWorkerAuthentication = errors.New("realm: worker authentication failed")
	ErrWorkerProtocol       = errors.New("realm: malformed worker protocol")
)

type workerRequest struct {
	Version string `json:"version"`
}

type workerAdmitRequest struct {
	Version   string          `json:"version"`
	Admission WorkerAdmission `json:"admission"`
}

type workerProjectRequest struct {
	Version  string           `json:"version"`
	PlayerID string           `json:"player_id"`
	Baseline d2save.Character `json:"baseline"`
}

type workerRemoveRequest struct {
	Version  string `json:"version"`
	PlayerID string `json:"player_id"`
}

type workerTicketIssueRequest struct {
	Version        string             `json:"version"`
	Principal      AdmissionPrincipal `json:"principal"`
	LifetimeMillis int64              `json:"lifetime_millis"`
}

type workerTicketRevokeRequest struct {
	Version string `json:"version"`
	Ticket  string `json:"ticket"`
}

func (request *workerRequest) protocolVersion() string             { return request.Version }
func (request *workerAdmitRequest) protocolVersion() string        { return request.Version }
func (request *workerProjectRequest) protocolVersion() string      { return request.Version }
func (request *workerRemoveRequest) protocolVersion() string       { return request.Version }
func (request *workerTicketIssueRequest) protocolVersion() string  { return request.Version }
func (request *workerTicketRevokeRequest) protocolVersion() string { return request.Version }

type workerEnvelope struct {
	Version string          `json:"version"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// NewWorkerHTTPHandler exposes the private process-independent worker
// operations. The caller owns TLS; this handler additionally requires an
// independent high-entropy bearer token on every request.
func NewWorkerHTTPHandler(worker WorkerClient, tickets TicketIssuer, token string, drain func()) (http.Handler, error) {
	if worker == nil || tickets == nil || len(strings.TrimSpace(token)) < 32 {
		return nil, errors.New("realm: worker HTTP service requires worker, tickets, and a 32-byte token")
	}
	server := &workerHTTPServer{worker: worker, tickets: tickets, token: token, drain: drain}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/describe", server.describe)
	mux.HandleFunc("POST /v1/status", server.status)
	mux.HandleFunc("POST /v1/checkpoint", server.checkpoint)
	mux.HandleFunc("POST /v1/admit", server.admit)
	mux.HandleFunc("POST /v1/remove", server.remove)
	mux.HandleFunc("POST /v1/project", server.project)
	mux.HandleFunc("POST /v1/tickets/issue", server.issueTicket)
	mux.HandleFunc("POST /v1/tickets/revoke", server.revokeTicket)
	mux.HandleFunc("POST /v1/drain", server.beginDrain)
	return http.MaxBytesHandler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !validWorkerBearer(request.Header.Get("Authorization"), server.token) {
			writeWorkerResponse(writer, nil, ErrWorkerAuthentication)
			return
		}
		mux.ServeHTTP(writer, request)
	}), maximumWorkerRequestBytes), nil
}

type workerHTTPServer struct {
	worker  WorkerClient
	tickets TicketIssuer
	token   string
	drain   func()
}

func (server *workerHTTPServer) describe(writer http.ResponseWriter, request *http.Request) {
	if !decodeWorkerRequest(writer, request, &workerRequest{Version: WorkerControlProtocolVersion}) {
		return
	}
	value, err := server.worker.Describe(request.Context())
	writeWorkerResponse(writer, value, err)
}

func (server *workerHTTPServer) status(writer http.ResponseWriter, request *http.Request) {
	if !decodeWorkerRequest(writer, request, &workerRequest{Version: WorkerControlProtocolVersion}) {
		return
	}
	value, err := server.worker.Status(request.Context())
	writeWorkerResponse(writer, value, err)
}

func (server *workerHTTPServer) checkpoint(writer http.ResponseWriter, request *http.Request) {
	if !decodeWorkerRequest(writer, request, &workerRequest{Version: WorkerControlProtocolVersion}) {
		return
	}
	value, err := server.worker.Checkpoint(request.Context())
	if err == nil {
		encoded, encodeErr := json.Marshal(value)
		if encodeErr != nil || len(encoded) == 0 || len(encoded) > maximumGameCheckpointBytes {
			err = ErrWorkerProtocol
		}
	}
	writeWorkerResponse(writer, value, err)
}

func (server *workerHTTPServer) admit(writer http.ResponseWriter, request *http.Request) {
	var input workerAdmitRequest
	if !decodeWorkerRequest(writer, request, &input) {
		return
	}
	err := server.worker.AdmitCharacter(request.Context(), input.Admission)
	if err != nil {
		slog.Debug("worker character admission failed", "character_id", input.Admission.Character.ID,
			"player_id", input.Admission.PlayerID, "error", err)
	}
	writeWorkerResponse(writer, struct{}{}, err)
}

func (server *workerHTTPServer) remove(writer http.ResponseWriter, request *http.Request) {
	var input workerRemoveRequest
	if !decodeWorkerRequest(writer, request, &input) {
		return
	}
	if strings.TrimSpace(input.PlayerID) == "" {
		writeWorkerResponse(writer, nil, ErrWorkerProtocol)
		return
	}
	writeWorkerResponse(writer, struct{}{}, server.worker.RemoveCharacter(request.Context(), input.PlayerID))
}

func (server *workerHTTPServer) project(writer http.ResponseWriter, request *http.Request) {
	var input workerProjectRequest
	if !decodeWorkerRequest(writer, request, &input) {
		return
	}
	value, err := server.worker.ProjectCharacter(request.Context(), input.PlayerID, input.Baseline)
	writeWorkerResponse(writer, value, err)
}

func (server *workerHTTPServer) issueTicket(writer http.ResponseWriter, request *http.Request) {
	var input workerTicketIssueRequest
	if !decodeWorkerRequest(writer, request, &input) {
		return
	}
	if input.LifetimeMillis <= 0 || input.LifetimeMillis > maximumWorkerTicketLifetime.Milliseconds() {
		writeWorkerResponse(writer, nil, ErrWorkerProtocol)
		return
	}
	value, err := server.tickets.Issue(request.Context(), input.Principal, time.Duration(input.LifetimeMillis)*time.Millisecond)
	writeWorkerResponse(writer, map[string]string{"ticket": value}, err)
}

func (server *workerHTTPServer) revokeTicket(writer http.ResponseWriter, request *http.Request) {
	var input workerTicketRevokeRequest
	if !decodeWorkerRequest(writer, request, &input) {
		return
	}
	writeWorkerResponse(writer, struct{}{}, server.tickets.Revoke(request.Context(), input.Ticket))
}

func (server *workerHTTPServer) beginDrain(writer http.ResponseWriter, request *http.Request) {
	var input workerRequest
	if !decodeWorkerRequest(writer, request, &input) {
		return
	}
	if server.drain == nil {
		writeWorkerResponse(writer, nil, ErrWorker)
		return
	}
	writeWorkerResponse(writer, struct{}{}, nil)
	if controller := http.NewResponseController(writer); controller != nil {
		_ = controller.Flush()
	}
	server.drain()
}

func decodeWorkerRequest(writer http.ResponseWriter, request *http.Request, destination any) bool {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		writeWorkerResponse(writer, nil, ErrWorkerProtocol)
		return false
	}
	versioned, ok := destination.(interface{ protocolVersion() string })
	if !ok || versioned.protocolVersion() != WorkerControlProtocolVersion {
		writeWorkerResponse(writer, nil, ErrWorkerProtocol)
		return false
	}
	return true
}

func writeWorkerResponse(writer http.ResponseWriter, value any, err error) {
	writer.Header().Set("Content-Type", "application/json")
	envelope := workerEnvelope{Version: WorkerControlProtocolVersion}
	status := http.StatusOK
	if err != nil {
		envelope.Error = "worker_failed"
		status = http.StatusInternalServerError
		if errors.Is(err, ErrWorkerAuthentication) {
			envelope.Error, status = "unauthorized", http.StatusUnauthorized
		} else if errors.Is(err, ErrWorkerProtocol) {
			envelope.Error, status = "invalid_protocol", http.StatusBadRequest
		} else if errors.Is(err, gamesession.ErrCompatibility) {
			envelope.Error = "incompatible_runtime"
		}
	} else {
		envelope.Data, err = json.Marshal(value)
		if err != nil {
			envelope.Error, status = "worker_failed", http.StatusInternalServerError
		}
	}
	payload, encodeErr := json.Marshal(envelope)
	if encodeErr != nil || len(payload) > maximumWorkerResponseBytes {
		envelope = workerEnvelope{Version: WorkerControlProtocolVersion, Error: "worker_failed"}
		status = http.StatusInternalServerError
		payload, _ = json.Marshal(envelope)
	}
	writer.WriteHeader(status)
	_, _ = writer.Write(payload)
}

func validWorkerBearer(header, token string) bool {
	scheme, candidate, found := strings.Cut(header, " ")
	return found && strings.EqualFold(scheme, "Bearer") && hmac.Equal([]byte(strings.TrimSpace(candidate)), []byte(token))
}

// WorkerHTTPClient implements both worker control and ticket issue/revocation
// over the private bounded HTTPS protocol.
type WorkerHTTPClient struct {
	base  *url.URL
	http  *http.Client
	token string
}

func NewWorkerHTTPClient(endpoint, token string, client *http.Client) (*WorkerHTTPClient, error) {
	base, err := url.Parse(strings.TrimRight(strings.TrimSpace(endpoint), "/"))
	if err != nil || base.Scheme != "https" || base.Host == "" || len(strings.TrimSpace(token)) < 32 {
		return nil, ErrWorkerProtocol
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &WorkerHTTPClient{base: base, http: client, token: token}, nil
}

func (client *WorkerHTTPClient) Describe(ctx context.Context) (WorkerDescription, error) {
	var result WorkerDescription
	err := client.call(ctx, "/v1/describe", workerRequest{Version: WorkerControlProtocolVersion}, &result)
	return result, err
}

func (client *WorkerHTTPClient) Status(ctx context.Context) (WorkerStatus, error) {
	var result WorkerStatus
	err := client.call(ctx, "/v1/status", workerRequest{Version: WorkerControlProtocolVersion}, &result)
	return result, err
}

func (client *WorkerHTTPClient) Checkpoint(ctx context.Context) (gamesession.RecoveryCheckpoint, error) {
	var result gamesession.RecoveryCheckpoint
	err := client.call(ctx, "/v1/checkpoint", workerRequest{Version: WorkerControlProtocolVersion}, &result)
	return result, err
}

func (client *WorkerHTTPClient) AdmitCharacter(ctx context.Context, admission WorkerAdmission) error {
	return client.call(ctx, "/v1/admit", workerAdmitRequest{Version: WorkerControlProtocolVersion, Admission: admission}, nil)
}

func (client *WorkerHTTPClient) RemoveCharacter(ctx context.Context, playerID string) error {
	if strings.TrimSpace(playerID) == "" {
		return ErrWorkerProtocol
	}
	return client.call(ctx, "/v1/remove", workerRemoveRequest{Version: WorkerControlProtocolVersion, PlayerID: playerID}, nil)
}

func (client *WorkerHTTPClient) ProjectCharacter(ctx context.Context, playerID string, baseline d2save.Character) (d2save.Character, error) {
	var result d2save.Character
	err := client.call(ctx, "/v1/project", workerProjectRequest{Version: WorkerControlProtocolVersion,
		PlayerID: playerID, Baseline: baseline}, &result)
	return result, err
}

func (client *WorkerHTTPClient) Issue(ctx context.Context, principal AdmissionPrincipal, lifetime time.Duration) (string, error) {
	if lifetime <= 0 || lifetime > maximumWorkerTicketLifetime {
		return "", ErrWorkerProtocol
	}
	var result map[string]string
	err := client.call(ctx, "/v1/tickets/issue", workerTicketIssueRequest{Version: WorkerControlProtocolVersion,
		Principal: principal, LifetimeMillis: lifetime.Milliseconds()}, &result)
	if err != nil || strings.TrimSpace(result["ticket"]) == "" {
		return "", errors.Join(err, ErrWorker)
	}
	return result["ticket"], nil
}

func (client *WorkerHTTPClient) Revoke(ctx context.Context, ticket string) error {
	return client.call(ctx, "/v1/tickets/revoke", workerTicketRevokeRequest{Version: WorkerControlProtocolVersion, Ticket: ticket}, nil)
}

func (client *WorkerHTTPClient) Close(ctx context.Context) error {
	err := client.call(ctx, "/v1/drain", workerRequest{Version: WorkerControlProtocolVersion}, nil)
	if transport, ok := client.http.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return err
}

func (client *WorkerHTTPClient) call(ctx context.Context, path string, input, output any) error {
	if client == nil || client.base == nil || client.http == nil || ctx == nil {
		return ErrWorkerProtocol
	}
	data, err := json.Marshal(input)
	if err != nil || len(data) > maximumWorkerRequestBytes {
		return ErrWorkerProtocol
	}
	relative, err := url.Parse(path)
	if err != nil {
		return ErrWorkerProtocol
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.base.ResolveReference(relative).String(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+client.token)
	response, err := client.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, maximumWorkerResponseBytes+1))
	if err != nil || len(payload) > maximumWorkerResponseBytes {
		return ErrWorkerProtocol
	}
	var envelope workerEnvelope
	if err := strictWorkerJSON(payload, &envelope); err != nil || envelope.Version != WorkerControlProtocolVersion {
		return ErrWorkerProtocol
	}
	if response.StatusCode != http.StatusOK || envelope.Error != "" {
		if envelope.Error == "unauthorized" {
			return ErrWorkerAuthentication
		}
		if envelope.Error == "invalid_protocol" {
			return ErrWorkerProtocol
		}
		if envelope.Error == "incompatible_runtime" {
			return gamesession.ErrCompatibility
		}
		return fmt.Errorf("%w: %s", ErrWorker, envelope.Error)
	}
	if output == nil {
		return nil
	}
	return strictWorkerJSON(envelope.Data, output)
}

func strictWorkerJSON(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrWorkerProtocol
	}
	return nil
}
