package sessionquic

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/quic-go/quic-go"
)

// Server owns a QUIC listener and adapts reliable streams and datagrams to one transport-neutral endpoint.
type Server struct {
	listener *quic.Listener
	endpoint *gameserver.Endpoint
	profiles ProfileAdmissions
	packages PackageProvider
}

// Listen opens a production UDP listener after cloning and constraining the caller's TLS configuration.
func Listen(address string, tlsConfig *tls.Config, endpoint *gameserver.Endpoint) (*Server, error) {
	if tlsConfig == nil || endpoint == nil {
		return nil, errors.New("game session QUIC: TLS and endpoint are required")
	}

	listener, err := quic.ListenAddr(address, serverTLS(tlsConfig), quicConfig())
	if err != nil {
		return nil, err
	}

	return &Server{listener: listener, endpoint: endpoint}, nil
}

// ListenPacket serves through a caller-provided socket so platforms and tests can control packet behavior.
func ListenPacket(
	packet net.PacketConn,
	tlsConfig *tls.Config,
	endpoint *gameserver.Endpoint,
) (*Server, error) {
	if packet == nil || tlsConfig == nil || endpoint == nil {
		return nil, errors.New("game session QUIC: packet connection, TLS, and endpoint are required")
	}

	listener, err := quic.Listen(packet, serverTLS(tlsConfig), quicConfig())
	if err != nil {
		return nil, err
	}

	return &Server{listener: listener, endpoint: endpoint}, nil
}

// Addr returns the bound listener address used by clients and test fixtures.
func (server *Server) Addr() string { return server.listener.Addr().String() }

// Close stops admission of new QUIC connections; active handlers then unwind through their contexts.
func (server *Server) Close() error { return server.listener.Close() }

// SetProfileAdmissions enables the explicitly self-hosted character-offer operation.
func (server *Server) SetProfileAdmissions(profiles ProfileAdmissions) { server.profiles = profiles }

// SetPackageProvider enables distribution of the exact recipe and extension bytes used by this authority.
func (server *Server) SetPackageProvider(packages PackageProvider) { server.packages = packages }

// Serve accepts connections until cancellation or listener failure and gives each connection independent ownership.
func (server *Server) Serve(ctx context.Context) error {
	for {
		connection, err := server.listener.Accept(ctx)
		if err != nil {
			return err
		}

		go server.serveConnection(ctx, connection)
	}
}

// serveConnection tracks every credential created on one transport so unexpected closure starts reconnect leases.
func (server *Server) serveConnection(ctx context.Context, connection *quic.Conn) {
	memberships := newConnectionMemberships()
	defer server.disconnectMemberships(memberships)

	for {
		stream, err := connection.AcceptStream(ctx)
		if err != nil {
			return
		}

		go server.serveStream(ctx, connection, stream, memberships)
	}
}

// disconnectMemberships suspends rather than removes members because transport loss may be recoverable.
func (server *Server) disconnectMemberships(memberships *connectionMemberships) {
	for _, credential := range memberships.snapshot() {
		server.endpoint.Disconnect(credential)
	}
}

// serveStream reads exactly one request because independent QUIC streams isolate durable operation failures.
func (server *Server) serveStream(
	ctx context.Context,
	connection *quic.Conn,
	stream *quic.Stream,
	memberships *connectionMemberships,
) {
	defer func() { _ = stream.Close() }()
	defer stream.CancelRead(0)

	var message request
	if err := readFrame(stream, &message); err != nil {
		_ = writeFrame(stream, response{Error: err.Error()})

		return
	}

	if message.Operation == operationWatch {
		server.serveWatch(ctx, connection, stream, message)

		return
	}

	requestContext := server.clientContext(ctx, connection)
	result := server.dispatchRateLimited(requestContext, message, memberships.packages, time.Now())
	memberships.observe(message, result)
	_ = writeFrame(stream, result)
}

// serveWatch validates the long-lived stream before handing ownership to the correction loop.
func (server *Server) serveWatch(
	ctx context.Context,
	connection *quic.Conn,
	stream *quic.Stream,
	message request,
) {
	if !validShape(message) || len(message.Credential) > MaxCredentialBytes {
		_ = writeFrame(stream, response{Error: ErrWire.Error()})

		return
	}

	server.watch(ctx, connection, stream, message.Credential)
}

// clientContext lets profile admission record the peer without making ordinary endpoint operations network-aware.
func (server *Server) clientContext(ctx context.Context, connection *quic.Conn) context.Context {
	contextSource, ok := server.profiles.(profileClientContext)
	if !ok {
		return ctx
	}

	return contextSource.WithClient(ctx, connection.RemoteAddr().String())
}

// dispatchRateLimited applies package bandwidth policy per connection before touching the shared package provider.
func (server *Server) dispatchRateLimited(
	ctx context.Context,
	message request,
	packages *packageRateLimiter,
	now time.Time,
) response {
	if message.Operation == operationPackageChunk &&
		message.Package != nil &&
		!packages.Allow(message.Package.Limit, now) {
		return response{Error: PackageRateLimitMessage}
	}

	return server.dispatch(ctx, message)
}
