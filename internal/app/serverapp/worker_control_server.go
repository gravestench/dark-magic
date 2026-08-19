package serverapp

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"
)

// WorkerControlServer owns the private HTTPS listener used by a Realm process
// allocator. It is not a public client or administration endpoint.
type WorkerControlServer struct {
	listener    net.Listener
	server      *http.Server
	fingerprint string
}

// newWorkerControlServer wraps the already-bound listener in TLS 1.3 and
// applies strict HTTP bounds before ownership passes to the caller.
func newWorkerControlServer(
	listener net.Listener,
	certificate tls.Certificate,
	fingerprint string,
	handler http.Handler,
) *WorkerControlServer {
	tlsListener := tls.NewListener(listener, &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS13,
	})
	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    8 << 10,
	}

	return &WorkerControlServer{
		listener:    tlsListener,
		server:      httpServer,
		fingerprint: fingerprint,
	}
}

// TLSFingerprint returns the leaf-certificate pin advertised to the allocator;
// a nil server reports no usable identity.
func (server *WorkerControlServer) TLSFingerprint() string {
	if server == nil {
		return ""
	}

	return server.fingerprint
}

// Addr returns the bound private endpoint, including the kernel-selected port
// used when configuration requested port zero.
func (server *WorkerControlServer) Addr() net.Addr {
	if server == nil || server.listener == nil {
		return nil
	}

	return server.listener.Addr()
}

// Serve runs until the listener fails or cancellation completes a bounded
// graceful shutdown. Expected http.Server shutdown is reported as success.
func (server *WorkerControlServer) Serve(ctx context.Context) error {
	if server == nil || server.listener == nil || server.server == nil || ctx == nil {
		return errors.New("server: invalid worker-control server")
	}

	shutdownDone := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			// WithoutCancel retains context values while guaranteeing that the
			// shutdown timeout, rather than the already-canceled parent, is used.
			shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			_ = server.server.Shutdown(shutdownCtx)

			cancel()
		case <-shutdownDone:
		}
	}()

	err := server.server.Serve(server.listener)

	close(shutdownDone)

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

// Close gracefully stops the control server; a nil context intentionally falls
// back to an unbounded background shutdown for compatibility with callers.
func (server *WorkerControlServer) Close(ctx context.Context) error {
	if server == nil || server.server == nil {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return server.server.Shutdown(ctx)
}
