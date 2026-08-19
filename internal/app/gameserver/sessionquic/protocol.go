package sessionquic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const (
	// ALPN prevents this protocol from sharing a TLS connection with an incompatible QUIC application.
	ALPN = "dark-magic-session/1"
	// InitialPacketSize stays within the minimum QUIC path MTU so handshakes do not depend on fragmentation.
	InitialPacketSize = 1200
	// CorrectionInterval is the reliable authoritative snapshot cadence used by client interpolation clocks.
	CorrectionInterval = 100 * time.Millisecond
	// TransformInterval matches the 25 Hz simulation cadence for disposable latest-wins samples.
	TransformInterval = 40 * time.Millisecond
	// MaxDatagramPayloadBytes leaves room for QUIC, packet-number, AEAD, and future framing overhead.
	MaxDatagramPayloadBytes = 1000
	// MaxFrameBytes bounds every length-prefixed reliable message before allocation.
	MaxFrameBytes = 4 << 20
	// MaxCommandPayloadBytes bounds client-controlled command JSON independently from the outer frame.
	MaxCommandPayloadBytes = 8 << 10
	// MaxCredentialBytes prevents opaque bearer values from consuming the general frame allowance.
	MaxCredentialBytes = 4 << 10
	// MaxProfileOfferBytes bounds self-hosted character offers before profile admission parses them.
	MaxProfileOfferBytes = 128 << 10
	// MaxPackageChunkBytes keeps package responses within a bounded streaming working set.
	MaxPackageChunkBytes = 32 << 10
)

var (
	// ErrWire reports a structurally invalid or ambiguous message without exposing parser internals.
	ErrWire = errors.New("game session QUIC: invalid wire message")
)

// PackageRateLimitMessage is stable because package clients distinguish retryable throttling from terminal failures.
const PackageRateLimitMessage = "game session QUIC: package distribution rate limit exceeded"

// RemoteError is a semantic rejection returned by a live server; it does not imply transport failure.
type RemoteError struct{ Message string }

// Error exposes the server's stable rejection text through the error interface.
func (err *RemoteError) Error() string { return err.Message }

// remoteError preserves the concrete error type so callers can choose whether a request is retryable.
func remoteError(message string) error {
	return &RemoteError{Message: message}
}

// operation selects one exact request shape on the shared reliable-stream envelope.
type operation string

const (
	operationJoin         operation = "join"
	operationSubmit       operation = "submit"
	operationRefresh      operation = "refresh"
	operationWatch        operation = "watch"
	operationReconnect    operation = "reconnect"
	operationLeave        operation = "leave"
	operationProfileAdmit operation = "profile_admit"
	operationRecipe       operation = "recipe"
	operationPackageChunk operation = "package_chunk"
)

// PackageRequest identifies one bounded range from an immutable runtime package.
type PackageRequest struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
	Offset int64  `json:"offset"`
	Limit  int    `json:"limit"`
}

// PackageChunk returns the exact identity and offset needed to reject substituted or reordered bytes.
type PackageChunk struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
	Offset int64  `json:"offset"`
	Total  int64  `json:"total"`
	Data   []byte `json:"data"`
}

// request is a tagged union whose unused fields must remain empty for every operation.
type request struct {
	Operation  operation                    `json:"operation"`
	Join       *gameserver.JoinRequest      `json:"join,omitempty"`
	Credential gameserver.SessionCredential `json:"credential,omitempty"`
	Command    *gameserver.CommandIntent    `json:"command,omitempty"`
	Reconnect  *gameserver.ReconnectRequest `json:"reconnect,omitempty"`
	Offer      json.RawMessage              `json:"offer,omitempty"`
	Package    *PackageRequest              `json:"package,omitempty"`
}

// response is a tagged union validated by each client operation before payload use.
type response struct {
	Join     *gameserver.JoinResponse  `json:"join,omitempty"`
	Snapshot *gameserver.Snapshot      `json:"snapshot,omitempty"`
	Error    string                    `json:"error,omitempty"`
	Ticket   string                    `json:"ticket,omitempty"`
	Recipe   *simulation.RuntimeRecipe `json:"recipe,omitempty"`
	Package  *PackageChunk             `json:"package,omitempty"`
}

// ProfileAdmissions converts a self-hosted profile offer into the same short-lived ticket used by ordinary joins.
type ProfileAdmissions interface {
	Admit(context.Context, string, []byte) (string, error)
}

// PackageProvider serves the authority's exact runtime recipe and immutable package bytes.
type PackageProvider interface {
	Recipe() simulation.RuntimeRecipe
	ReadChunk(context.Context, PackageRequest) (PackageChunk, error)
}

// profileClientContext lets admission implementations attach the remote address without widening their public API.
type profileClientContext interface {
	WithClient(context.Context, string) context.Context
}

// validShape rejects ambiguous tagged unions before any operation-specific logic can observe extra fields.
func validShape(message request) bool {
	switch message.Operation {
	case operationJoin:
		return message.Join != nil && message.Command == nil && message.Reconnect == nil &&
			message.Package == nil && message.Credential == "" && len(message.Offer) == 0
	case operationSubmit:
		return message.Join == nil && message.Command != nil && message.Reconnect == nil &&
			message.Package == nil && message.Credential != "" && len(message.Offer) == 0
	case operationRefresh, operationWatch, operationLeave:
		return message.Join == nil && message.Command == nil && message.Reconnect == nil &&
			message.Package == nil && message.Credential != "" && len(message.Offer) == 0
	case operationReconnect:
		return message.Join == nil && message.Command == nil && message.Reconnect != nil &&
			message.Package == nil && message.Credential == "" && len(message.Offer) == 0
	case operationProfileAdmit:
		return message.Join == nil && message.Command == nil && message.Reconnect == nil &&
			message.Package == nil && len(message.Offer) > 0
	case operationRecipe:
		return message.Join == nil && message.Command == nil && message.Reconnect == nil &&
			message.Package == nil && message.Credential == "" && len(message.Offer) == 0
	case operationPackageChunk:
		return message.Join == nil && message.Command == nil && message.Reconnect == nil &&
			message.Package != nil && message.Credential == "" && len(message.Offer) == 0 &&
			message.Package.Offset >= 0 && message.Package.Limit > 0 &&
			message.Package.Limit <= MaxPackageChunkBytes
	default:
		return false
	}
}

// joinResponse maps endpoint errors into the wire envelope without changing their stable message text.
func joinResponse(joined gameserver.JoinResponse, err error) response {
	if err != nil {
		return response{Error: err.Error()}
	}

	return response{Join: &joined}
}

// errorResponse maps an error-only operation into an otherwise empty success envelope.
func errorResponse(err error) response {
	if err != nil {
		return response{Error: err.Error()}
	}

	return response{}
}
