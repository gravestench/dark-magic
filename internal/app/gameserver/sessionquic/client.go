package sessionquic

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/quic-go/quic-go"
)

// Client owns one authenticated QUIC connection and serializes access to its shared datagram receive queue.
type Client struct {
	connection           *quic.Conn
	datagramMu           sync.Mutex
	datagramActive       bool
	transformsReceived   atomic.Uint64
	transformsSuperseded atomic.Uint64
}

// NetworkStats combines QUIC path timing with latest-wins transform delivery counters.
type NetworkStats struct {
	SmoothedRTT          time.Duration
	RTTVariation         time.Duration
	TransformsReceived   uint64
	TransformsSuperseded uint64
}

// Dial opens a production UDP connection after cloning and constraining the caller's TLS configuration.
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

// DialPacket connects through a caller-provided socket whose ownership transfers after a successful handshake.
func DialPacket(
	ctx context.Context,
	packet net.PacketConn,
	address net.Addr,
	tlsConfig *tls.Config,
) (*Client, error) {
	if packet == nil || address == nil || tlsConfig == nil {
		return nil, errors.New("game session QUIC: packet connection, address, and TLS are required")
	}

	connection, err := quic.Dial(ctx, packet, address, clientTLS(tlsConfig), quicConfig())
	if err != nil {
		return nil, err
	}

	return &Client{connection: connection}, nil
}

// Close gracefully terminates the owned QUIC connection and all active stream watches.
func (client *Client) Close() error { return client.connection.CloseWithError(0, "closed") }

// NetworkStats takes atomic transform counters so presentation telemetry never blocks datagram delivery.
func (client *Client) NetworkStats() NetworkStats {
	if client == nil || client.connection == nil {
		return NetworkStats{}
	}

	stats := client.connection.ConnectionStats()

	return NetworkStats{
		SmoothedRTT:          stats.SmoothedRTT,
		RTTVariation:         stats.MeanDeviation,
		TransformsReceived:   client.transformsReceived.Load(),
		TransformsSuperseded: client.transformsSuperseded.Load(),
	}
}

// Join performs the admission handshake and rejects a successful envelope without its required payload.
func (client *Client) Join(
	ctx context.Context,
	join gameserver.JoinRequest,
) (gameserver.JoinResponse, error) {
	result, err := client.call(ctx, request{Operation: operationJoin, Join: &join})
	if err != nil {
		return gameserver.JoinResponse{}, err
	}

	if result.Join == nil {
		return gameserver.JoinResponse{}, ErrWire
	}

	return *result.Join, nil
}

// Submit sends one authenticated command intent on an isolated reliable stream.
func (client *Client) Submit(
	ctx context.Context,
	credential gameserver.SessionCredential,
	command gameserver.CommandIntent,
) error {
	_, err := client.call(ctx, request{
		Operation:  operationSubmit,
		Credential: credential,
		Command:    &command,
	})

	return err
}

// Refresh requests a client-paced correction and requires exactly one snapshot payload on success.
func (client *Client) Refresh(
	ctx context.Context,
	credential gameserver.SessionCredential,
) (gameserver.Snapshot, error) {
	result, err := client.call(ctx, request{Operation: operationRefresh, Credential: credential})
	if err != nil {
		return gameserver.Snapshot{}, err
	}

	if result.Snapshot == nil {
		return gameserver.Snapshot{}, ErrWire
	}

	return *result.Snapshot, nil
}

// Reconnect rotates a retained membership credential and requires a complete replacement admission response.
func (client *Client) Reconnect(
	ctx context.Context,
	reconnect gameserver.ReconnectRequest,
) (gameserver.JoinResponse, error) {
	result, err := client.call(ctx, request{Operation: operationReconnect, Reconnect: &reconnect})
	if err != nil {
		return gameserver.JoinResponse{}, err
	}

	if result.Join == nil {
		return gameserver.JoinResponse{}, ErrWire
	}

	return *result.Join, nil
}

// Leave explicitly revokes a membership so connection cleanup does not leave it in reconnect grace.
func (client *Client) Leave(ctx context.Context, credential gameserver.SessionCredential) error {
	_, err := client.call(ctx, request{Operation: operationLeave, Credential: credential})

	return err
}

// AdmitProfile exchanges a bounded self-hosted character offer for a short-lived ordinary join ticket.
func (client *Client) AdmitProfile(ctx context.Context, credential string, offer []byte) (string, error) {
	if len(offer) == 0 || len(offer) > MaxProfileOfferBytes {
		return "", ErrWire
	}

	result, err := client.call(ctx, request{
		Operation:  operationProfileAdmit,
		Credential: gameserver.SessionCredential(credential),
		Offer:      append(json.RawMessage(nil), offer...),
	})
	if err != nil {
		return "", err
	}

	if result.Ticket == "" {
		return "", ErrWire
	}

	return result.Ticket, nil
}

// Recipe fetches and validates the exact runtime identity before any extension bytes are trusted.
func (client *Client) Recipe(ctx context.Context) (simulation.RuntimeRecipe, error) {
	result, err := client.call(ctx, request{Operation: operationRecipe})
	if err != nil {
		return simulation.RuntimeRecipe{}, err
	}

	if !validRecipeResponse(result) {
		return simulation.RuntimeRecipe{}, ErrWire
	}

	return *result.Recipe, nil
}

// PackageChunk validates request bounds and response identity so bytes cannot be substituted or reordered.
func (client *Client) PackageChunk(ctx context.Context, packageRequest PackageRequest) (PackageChunk, error) {
	if packageRequest.Offset < 0 ||
		packageRequest.Limit <= 0 ||
		packageRequest.Limit > MaxPackageChunkBytes {
		return PackageChunk{}, ErrWire
	}

	result, err := client.call(ctx, packageRequestMessage(packageRequest))
	if err != nil {
		return PackageChunk{}, err
	}

	if !validPackageResponse(result, packageRequest) {
		return PackageChunk{}, ErrWire
	}

	return *result.Package, nil
}

// validRecipeResponse rejects mixed envelopes and malformed recipes even when the server reported success.
func validRecipeResponse(result response) bool {
	return result.Recipe != nil && result.Recipe.Validate() == nil &&
		result.Join == nil && result.Snapshot == nil && result.Package == nil &&
		result.Ticket == "" && result.Error == ""
}

// validPackageResponse binds a non-empty bounded chunk to the exact package range the client requested.
func validPackageResponse(result response, packageRequest PackageRequest) bool {
	if result.Package == nil ||
		result.Join != nil ||
		result.Snapshot != nil ||
		result.Recipe != nil ||
		result.Ticket != "" ||
		result.Error != "" {
		return false
	}

	chunk := result.Package
	if chunk.ID != packageRequest.ID ||
		chunk.Digest != packageRequest.Digest ||
		chunk.Offset != packageRequest.Offset ||
		chunk.Total <= 0 ||
		len(chunk.Data) == 0 ||
		len(chunk.Data) > packageRequest.Limit ||
		chunk.Offset >= chunk.Total {
		return false
	}

	return int64(len(chunk.Data)) <= chunk.Total-chunk.Offset
}

// packageRequestMessage takes an addressable copy so the request owns its serialized package range.
func packageRequestMessage(packageRequest PackageRequest) request {
	return request{Operation: operationPackageChunk, Package: &packageRequest}
}

// call exchanges one bounded request and response on a fresh stream so failures cannot desynchronize later calls.
func (client *Client) call(ctx context.Context, message request) (response, error) {
	stream, err := client.connection.OpenStreamSync(ctx)
	if err != nil {
		return response{}, err
	}
	defer func() { _ = stream.Close() }()
	defer stream.CancelRead(0)

	if deadline, ok := ctx.Deadline(); ok {
		if err := stream.SetDeadline(deadline); err != nil {
			return response{}, err
		}
	}

	if err := writeFrame(stream, message); err != nil {
		return response{}, err
	}

	var result response
	if err := readFrame(stream, &result); err != nil {
		return response{}, err
	}

	if result.Error != "" {
		return result, remoteError(result.Error)
	}

	return result, nil
}
