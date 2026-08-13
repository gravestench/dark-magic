// Package sessionquic carries the authenticated game-session protocol over
// QUIC. Every operation uses a bounded reliable stream; replaceable datagram
// traffic will be added only with an explicit compact schema and loss tests.
package sessionquic

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/quic-go/quic-go"
)

const (
	ALPN              = "dark-magic-session/1"
	InitialPacketSize = 1200
	// MaxDatagramPayloadBytes leaves conservative room within a 1200-byte QUIC
	// packet for QUIC, packet-number, AEAD, and future message framing overhead.
	// Datagram transport remains disabled until its schemas are implemented.
	MaxDatagramPayloadBytes = 1000
	MaxFrameBytes           = 4 << 20
	MaxCommandPayloadBytes  = 8 << 10
	MaxCredentialBytes      = 4 << 10
	MaxProfileOfferBytes    = 128 << 10
)

var ErrWire = errors.New("game session QUIC: invalid wire message")

type operation string

const (
	operationJoin         operation = "join"
	operationSubmit       operation = "submit"
	operationRefresh      operation = "refresh"
	operationWatch        operation = "watch"
	operationReconnect    operation = "reconnect"
	operationLeave        operation = "leave"
	operationProfileAdmit operation = "profile_admit"
)

type request struct {
	Operation  operation                    `json:"operation"`
	Join       *gameserver.JoinRequest      `json:"join,omitempty"`
	Credential gameserver.SessionCredential `json:"credential,omitempty"`
	Command    *gameserver.CommandIntent    `json:"command,omitempty"`
	Reconnect  *gameserver.ReconnectRequest `json:"reconnect,omitempty"`
	Offer      json.RawMessage              `json:"offer,omitempty"`
}

type response struct {
	Join     *gameserver.JoinResponse `json:"join,omitempty"`
	Snapshot *gameserver.Snapshot     `json:"snapshot,omitempty"`
	Error    string                   `json:"error,omitempty"`
	Ticket   string                   `json:"ticket,omitempty"`
}

type ProfileAdmissions interface {
	Admit(context.Context, string, []byte) (string, error)
}

type Server struct {
	listener *quic.Listener
	endpoint *gameserver.Endpoint
	profiles ProfileAdmissions
}

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

// ListenPacket serves the production protocol on a caller-provided packet
// connection. It permits platform socket configuration and deterministic
// network impairment tests without replacing any protocol behavior.
func ListenPacket(packet net.PacketConn, tlsConfig *tls.Config, endpoint *gameserver.Endpoint) (*Server, error) {
	if packet == nil || tlsConfig == nil || endpoint == nil {
		return nil, errors.New("game session QUIC: packet connection, TLS, and endpoint are required")
	}
	listener, err := quic.Listen(packet, serverTLS(tlsConfig), quicConfig())
	if err != nil {
		return nil, err
	}
	return &Server{listener: listener, endpoint: endpoint}, nil
}

func (server *Server) Addr() string { return server.listener.Addr().String() }

func (server *Server) Close() error { return server.listener.Close() }

// SetProfileAdmissions enables the explicitly self-hosted character-offer
// operation. Realm servers leave it nil and therefore reject that operation.
func (server *Server) SetProfileAdmissions(profiles ProfileAdmissions) { server.profiles = profiles }

func (server *Server) Serve(ctx context.Context) error {
	for {
		connection, err := server.listener.Accept(ctx)
		if err != nil {
			return err
		}
		go server.serveConnection(ctx, connection)
	}
}

func (server *Server) serveConnection(ctx context.Context, connection *quic.Conn) {
	for {
		stream, err := connection.AcceptStream(ctx)
		if err != nil {
			return
		}
		go func() {
			defer stream.Close()
			defer stream.CancelRead(0)
			var message request
			if err := readFrame(stream, &message); err != nil {
				_ = writeFrame(stream, response{Error: err.Error()})
				return
			}
			if message.Operation == operationWatch {
				if !validShape(message) || len(message.Credential) > MaxCredentialBytes {
					_ = writeFrame(stream, response{Error: ErrWire.Error()})
					return
				}
				server.watch(ctx, stream, message.Credential)
				return
			}
			_ = writeFrame(stream, server.dispatch(ctx, message))
		}()
	}
}

// A 10 Hz authoritative stream leaves two to three 25 Hz simulation samples
// between corrections for interpolation without flooding full projections.
const correctionInterval = 100 * time.Millisecond

func (server *Server) watch(ctx context.Context, stream *quic.Stream, credential gameserver.SessionCredential) {
	ticker := time.NewTicker(correctionInterval)
	defer ticker.Stop()
	lastTick, lastChecksum := ^uint64(0), ""
	for {
		snapshot, err := server.endpoint.Refresh(credential)
		if err != nil {
			_ = writeFrame(stream, response{Error: err.Error()})
			return
		}
		if snapshot.Tick != lastTick || snapshot.Checksum != lastChecksum {
			if err := writeFrame(stream, response{Snapshot: &snapshot}); err != nil {
				return
			}
			lastTick, lastChecksum = snapshot.Tick, snapshot.Checksum
		}
		select {
		case <-ctx.Done():
			return
		case <-stream.Context().Done():
			return
		case <-ticker.C:
		}
	}
}

func (server *Server) dispatch(ctx context.Context, message request) response {
	if !validShape(message) {
		return response{Error: ErrWire.Error()}
	}
	switch message.Operation {
	case operationJoin:
		if message.Join == nil || len(message.Join.Credential) > MaxCredentialBytes {
			return response{Error: ErrWire.Error()}
		}
		joined, err := server.endpoint.Join(ctx, *message.Join)
		return joinResponse(joined, err)
	case operationSubmit:
		if message.Command == nil || len(message.Credential) > MaxCredentialBytes || len(message.Command.Payload) > MaxCommandPayloadBytes {
			return response{Error: ErrWire.Error()}
		}
		return errorResponse(server.endpoint.Submit(message.Credential, *message.Command))
	case operationRefresh:
		if len(message.Credential) > MaxCredentialBytes {
			return response{Error: ErrWire.Error()}
		}
		snapshot, err := server.endpoint.Refresh(message.Credential)
		if err != nil {
			return response{Error: err.Error()}
		}
		return response{Snapshot: &snapshot}
	case operationReconnect:
		if message.Reconnect == nil || len(message.Reconnect.Credential) > MaxCredentialBytes {
			return response{Error: ErrWire.Error()}
		}
		joined, err := server.endpoint.Reconnect(*message.Reconnect)
		return joinResponse(joined, err)
	case operationLeave:
		if len(message.Credential) > MaxCredentialBytes {
			return response{Error: ErrWire.Error()}
		}
		return errorResponse(server.endpoint.Leave(message.Credential))
	case operationProfileAdmit:
		if server.profiles == nil || len(message.Credential) > MaxCredentialBytes || len(message.Offer) == 0 || len(message.Offer) > MaxProfileOfferBytes {
			return response{Error: ErrWire.Error()}
		}
		ticket, err := server.profiles.Admit(ctx, message.Credential.String(), message.Offer)
		if err != nil {
			return response{Error: err.Error()}
		}
		return response{Ticket: ticket}
	default:
		return response{Error: ErrWire.Error()}
	}
}

func validShape(message request) bool {
	switch message.Operation {
	case operationJoin:
		return message.Join != nil && message.Command == nil && message.Reconnect == nil && message.Credential == "" && len(message.Offer) == 0
	case operationSubmit:
		return message.Join == nil && message.Command != nil && message.Reconnect == nil && message.Credential != "" && len(message.Offer) == 0
	case operationRefresh, operationWatch, operationLeave:
		return message.Join == nil && message.Command == nil && message.Reconnect == nil && message.Credential != "" && len(message.Offer) == 0
	case operationReconnect:
		return message.Join == nil && message.Command == nil && message.Reconnect != nil && message.Credential == "" && len(message.Offer) == 0
	case operationProfileAdmit:
		return message.Join == nil && message.Command == nil && message.Reconnect == nil && len(message.Offer) > 0
	default:
		return false
	}
}

type Client struct{ connection *quic.Conn }

func Dial(ctx context.Context, address string, tlsConfig *tls.Config) (*Client, error) {
	if tlsConfig == nil {
		return nil, errors.New("game session QUIC: TLS is required")
	}
	connection, err := quic.DialAddr(ctx, address, clientTLS(tlsConfig), quicConfig())
	if err != nil {
		return nil, err
	}
	return &Client{connection: connection}, nil
}

// DialPacket connects the production protocol through a caller-provided packet
// connection. The packet connection is owned by the returned client after a
// successful call.
func DialPacket(ctx context.Context, packet net.PacketConn, address net.Addr, tlsConfig *tls.Config) (*Client, error) {
	if packet == nil || address == nil || tlsConfig == nil {
		return nil, errors.New("game session QUIC: packet connection, address, and TLS are required")
	}
	connection, err := quic.Dial(ctx, packet, address, clientTLS(tlsConfig), quicConfig())
	if err != nil {
		return nil, err
	}
	return &Client{connection: connection}, nil
}

func (client *Client) Close() error { return client.connection.CloseWithError(0, "closed") }

func (client *Client) Join(ctx context.Context, join gameserver.JoinRequest) (gameserver.JoinResponse, error) {
	result, err := client.call(ctx, request{Operation: operationJoin, Join: &join})
	if err != nil {
		return gameserver.JoinResponse{}, err
	}
	if result.Join == nil {
		return gameserver.JoinResponse{}, ErrWire
	}
	return *result.Join, nil
}

func (client *Client) Submit(ctx context.Context, credential gameserver.SessionCredential, command gameserver.CommandIntent) error {
	_, err := client.call(ctx, request{Operation: operationSubmit, Credential: credential, Command: &command})
	return err
}

func (client *Client) Refresh(ctx context.Context, credential gameserver.SessionCredential) (gameserver.Snapshot, error) {
	result, err := client.call(ctx, request{Operation: operationRefresh, Credential: credential})
	if err != nil {
		return gameserver.Snapshot{}, err
	}
	if result.Snapshot == nil {
		return gameserver.Snapshot{}, ErrWire
	}
	return *result.Snapshot, nil
}

// Watch opens one long-lived reliable correction stream. The bounded channel
// applies backpressure directly to QUIC; corrections are never queued without
// limit in application memory.
func (client *Client) Watch(ctx context.Context, credential gameserver.SessionCredential) (<-chan gameserver.Snapshot, <-chan error, error) {
	stream, err := client.connection.OpenStreamSync(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := writeFrame(stream, request{Operation: operationWatch, Credential: credential}); err != nil {
		stream.CancelRead(1)
		stream.CancelWrite(1)
		return nil, nil, err
	}
	snapshots := make(chan gameserver.Snapshot, 1)
	errorsOut := make(chan error, 1)
	go func() {
		defer close(snapshots)
		defer close(errorsOut)
		defer stream.Close()
		done := make(chan struct{})
		defer close(done)
		go func() {
			select {
			case <-ctx.Done():
				stream.CancelRead(0)
				stream.CancelWrite(0)
			case <-done:
			}
		}()
		for {
			var result response
			if err := readFrame(stream, &result); err != nil {
				if ctx.Err() == nil {
					errorsOut <- err
				}
				return
			}
			if result.Error != "" {
				errorsOut <- errors.New(result.Error)
				return
			}
			if result.Snapshot == nil {
				errorsOut <- ErrWire
				return
			}
			select {
			case snapshots <- *result.Snapshot:
			case <-ctx.Done():
				return
			}
		}
	}()
	return snapshots, errorsOut, nil
}

func (client *Client) Reconnect(ctx context.Context, reconnect gameserver.ReconnectRequest) (gameserver.JoinResponse, error) {
	result, err := client.call(ctx, request{Operation: operationReconnect, Reconnect: &reconnect})
	if err != nil {
		return gameserver.JoinResponse{}, err
	}
	if result.Join == nil {
		return gameserver.JoinResponse{}, ErrWire
	}
	return *result.Join, nil
}

func (client *Client) Leave(ctx context.Context, credential gameserver.SessionCredential) error {
	_, err := client.call(ctx, request{Operation: operationLeave, Credential: credential})
	return err
}

func (client *Client) AdmitProfile(ctx context.Context, credential string, offer []byte) (string, error) {
	if len(offer) == 0 || len(offer) > MaxProfileOfferBytes {
		return "", ErrWire
	}
	result, err := client.call(ctx, request{Operation: operationProfileAdmit, Credential: gameserver.SessionCredential(credential), Offer: append(json.RawMessage(nil), offer...)})
	if err != nil {
		return "", err
	}
	if result.Ticket == "" {
		return "", ErrWire
	}
	return result.Ticket, nil
}

func (client *Client) call(ctx context.Context, message request) (response, error) {
	stream, err := client.connection.OpenStreamSync(ctx)
	if err != nil {
		return response{}, err
	}
	defer stream.Close()
	defer stream.CancelRead(0)
	if err := writeFrame(stream, message); err != nil {
		return response{}, err
	}
	var result response
	if err := readFrame(stream, &result); err != nil {
		return response{}, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}

func quicConfig() *quic.Config {
	return &quic.Config{
		HandshakeIdleTimeout: 5 * time.Second, MaxIdleTimeout: 30 * time.Second,
		InitialPacketSize: InitialPacketSize, MaxIncomingStreams: 16,
		MaxIncomingUniStreams: -1, MaxStreamReceiveWindow: MaxFrameBytes,
		MaxConnectionReceiveWindow: 2 * MaxFrameBytes,
	}
}

func serverTLS(config *tls.Config) *tls.Config {
	clone := config.Clone()
	clone.NextProtos = []string{ALPN}
	clone.MinVersion = tls.VersionTLS13
	return clone
}

func clientTLS(config *tls.Config) *tls.Config {
	clone := config.Clone()
	clone.NextProtos = []string{ALPN}
	clone.MinVersion = tls.VersionTLS13
	return clone
}

func writeFrame(writer io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(data) == 0 || len(data) > MaxFrameBytes {
		return ErrWire
	}
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(data)))
	if err := writeAll(writer, size[:]); err != nil {
		return err
	}
	return writeAll(writer, data)
}

func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := writer.Write(data)
		if err != nil {
			return err
		}
		if written <= 0 || written > len(data) {
			return io.ErrShortWrite
		}
		data = data[written:]
	}
	return nil
}

func readFrame(reader io.Reader, value any) error {
	buffered := bufio.NewReaderSize(reader, 4<<10)
	var size [4]byte
	if _, err := io.ReadFull(buffered, size[:]); err != nil {
		return err
	}
	length := binary.BigEndian.Uint32(size[:])
	if length == 0 || length > MaxFrameBytes {
		return ErrWire
	}
	limited := &io.LimitedReader{R: buffered, N: int64(length)}
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return fmt.Errorf("%w: %v", ErrWire, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) || limited.N != 0 {
		return ErrWire
	}
	return nil
}

func joinResponse(joined gameserver.JoinResponse, err error) response {
	if err != nil {
		return response{Error: err.Error()}
	}
	return response{Join: &joined}
}

func errorResponse(err error) response {
	if err != nil {
		return response{Error: err.Error()}
	}
	return response{}
}
