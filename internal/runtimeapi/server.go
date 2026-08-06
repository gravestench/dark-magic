// Package runtimeapi exposes the narrow development-time component management
// surface. It delegates lifecycle serialization to host.Manager.
package runtimeapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/host"
)

type Server struct {
	address string
	manager *host.Manager
	server  *http.Server
	listen  net.Listener
}

func New(address string, manager *host.Manager) *Server {
	server := &Server{address: address, manager: manager}
	server.server = &http.Server{Addr: address, Handler: server.Handler(), ReadHeaderTimeout: 5 * time.Second}
	return server
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/components", s.list)
	mux.HandleFunc("POST /v1/components/{id}/{action}", s.transition)
	return mux
}

func (s *Server) Start(context.Context) error {
	if strings.TrimSpace(s.address) == "" {
		return nil
	}
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("runtimeapi: listen: %w", err)
	}
	s.listen = listener
	go func() {
		if err := s.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Runtime serving errors surface on shutdown through Server.Shutdown;
			// lifecycle transitions themselves remain synchronous HTTP responses.
			_ = listener.Close()
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.listen == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *Server) list(writer http.ResponseWriter, _ *http.Request) {
	type status struct {
		ID      string     `json:"id"`
		Desired bool       `json:"desired"`
		State   host.State `json:"state"`
		Error   string     `json:"error,omitempty"`
	}
	entries := s.manager.Statuses()
	result := make([]status, 0, len(entries))
	for _, entry := range entries {
		item := status{ID: entry.ID, Desired: entry.Desired, State: entry.State}
		if entry.Err != nil {
			item.Error = entry.Err.Error()
		}
		result = append(result, item)
	}
	writeJSON(writer, http.StatusOK, result)
}

func (s *Server) transition(writer http.ResponseWriter, request *http.Request) {
	id, action := request.PathValue("id"), request.PathValue("action")
	var err error
	switch action {
	case "enable":
		err = s.manager.Enable(request.Context(), id)
	case "disable":
		if request.URL.Query().Get("cascade") == "true" {
			err = s.manager.DisableCascade(request.Context(), id)
		} else {
			err = s.manager.Disable(request.Context(), id)
		}
	case "restart":
		err = s.manager.Restart(request.Context(), id)
	default:
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "unknown action"})
		return
	}
	if err != nil {
		writeJSON(writer, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	entry, _ := s.manager.Status(id)
	writeJSON(writer, http.StatusOK, entry)
}

func writeJSON(writer http.ResponseWriter, code int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)
	_ = json.NewEncoder(writer).Encode(value)
}
