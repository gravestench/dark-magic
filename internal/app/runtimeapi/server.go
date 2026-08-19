// Package runtimeapi exposes the narrow development-time component management
// surface. It delegates lifecycle serialization to host.Manager.
package runtimeapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/host"
)

const readHeaderTimeout = 5 * time.Second

// Server is an optional local administration adapter. Host.Manager remains the
// sole lifecycle reconciler; HTTP requests express desired transitions only.
type Server struct {
	address string
	manager *host.Manager
	server  *http.Server
	listen  net.Listener
}

// New constructs a stopped server. A blank address disables listening while preserving the handler for tests and
// embedding, so callers do not need a second construction path for development and production configurations.
func New(address string, manager *host.Manager) *Server {
	server := &Server{address: address, manager: manager}
	server.server = &http.Server{
		Addr:              address,
		Handler:           server.Handler(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return server
}

// Start binds the configured address before returning, then serves in the background. Binding synchronously ensures
// the host never reports a successful start for an address that was already unavailable.
func (s *Server) Start(context.Context) error {
	if strings.TrimSpace(s.address) == "" {
		return nil
	}

	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("runtimeapi: listen: %w", err)
	}

	s.listen = listener
	go s.serve(listener)

	return nil
}

// serve owns the background accept loop. Unexpected serving failures trigger defensive listener cleanup at this
// adapter boundary; expected shutdown errors need no additional handling.
func (s *Server) serve(listener net.Listener) {
	if err := s.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// Retain the explicit close so unexpected exits cannot leave adapter ownership ambiguous.
		_ = listener.Close()
	}
}

// Stop performs bounded HTTP shutdown through the host lifecycle context. A server that never listened is already
// stopped, which keeps disabled configurations and failed host startup cleanup idempotent.
func (s *Server) Stop(ctx context.Context) error {
	if s.listen == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}
