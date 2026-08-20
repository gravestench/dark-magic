package realm

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"
)

const RealmAuditVersion = "RealmAudit/v1"

const (
	AuditAccountCreate           = "account.create"
	AuditAccountLogin            = "account.login"
	AuditAccountLogout           = "account.logout"
	AuditAccountExpire           = "account.expire"
	AuditAccountSignup           = "account.signup"
	AuditAccountVerify           = "account.verify"
	AuditAccountRecoveryBegin    = "account.recovery.begin"
	AuditAccountRecoveryComplete = "account.recovery.complete"
	AuditCharacterCreate         = "character.create"
	AuditCharacterSelect         = "character.select"
	AuditCharacterDelete         = "character.delete"
	AuditChannelJoin             = "channel.join"
	AuditChannelMessage          = "channel.message"
	AuditGameCreate              = "game.create"
	AuditGameResolve             = "game.resolve"
	AuditGameJoin                = "game.join"
	AuditGameReconnect           = "game.reconnect"
	AuditGameRestore             = "game.restore"
	AuditGameLeave               = "game.leave"
	AuditGameDrain               = "game.drain"
	AuditGameReconcile           = "game.reconcile"
)

// AuditEvent is the deliberately bounded security and operations record emitted
// by the realm control plane. It has no fields capable of carrying passwords,
// bearer tokens, password hashes, chat contents, or character save payloads.
type AuditEvent struct {
	Version       string `json:"version"`
	Operation     string `json:"operation"`
	Outcome       string `json:"outcome"`
	ErrorCode     string `json:"error_code,omitempty"`
	ClientAddress string `json:"client_address,omitempty"`
	AccountID     string `json:"account_id,omitempty"`
	AccountName   string `json:"account_name,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
	CharacterID   string `json:"character_id,omitempty"`
	CharacterName string `json:"character_name,omitempty"`
	Channel       string `json:"channel,omitempty"`
	GameID        string `json:"game_id,omitempty"`
	GameName      string `json:"game_name,omitempty"`
	GameReference string `json:"game_reference,omitempty"`
	MessageBytes  int    `json:"message_bytes,omitempty"`
}

// AuditSink receives semantic realm events independently of their transport.
// Implementations must be safe for concurrent use.
type AuditSink interface {
	Record(context.Context, AuditEvent)
}

type AuditSinkFunc func(context.Context, AuditEvent)

// Record contains record within the audit boundary so callers do not duplicate its domain-specific policy.
func (function AuditSinkFunc) Record(ctx context.Context, event AuditEvent) {
	if function != nil {
		function(ctx, event)
	}
}

type auditFanout []AuditSink

// NewAuditFanout delivers the same bounded semantic event to durable and
// observational sinks without allowing either to see credentials or saves.
func NewAuditFanout(sinks ...AuditSink) AuditSink {
	filtered := make(auditFanout, 0, len(sinks))
	for _, sink := range sinks {
		if sink != nil {
			filtered = append(filtered, sink)
		}
	}

	return filtered
}

// Record contains record within the audit boundary so callers do not duplicate its domain-specific policy.
func (fanout auditFanout) Record(ctx context.Context, event AuditEvent) {
	for _, sink := range fanout {
		sink.Record(ctx, event)
	}
}

type slogAuditSink struct {
	logger *slog.Logger
}

// NewSlogAuditSink emits one structured log entry per realm audit event. A nil
// logger follows slog.Default at record time so a headless administration shell
// can install its process-wide observer after constructing the control plane.
func NewSlogAuditSink(logger *slog.Logger) AuditSink {
	return slogAuditSink{logger: logger}
}

// Record contains record within the audit boundary so callers do not duplicate its domain-specific policy.
func (sink slogAuditSink) Record(ctx context.Context, event AuditEvent) {
	logger := sink.logger
	if logger == nil {
		logger = slog.Default()
	}

	level := slog.LevelInfo
	if event.Outcome != "success" {
		level = slog.LevelWarn
	}

	attributes := []slog.Attr{
		slog.String("audit_version", event.Version),
		slog.String("operation", event.Operation),
		slog.String("outcome", event.Outcome),
	}
	attributes = appendAuditString(attributes, "error_code", event.ErrorCode)
	attributes = appendAuditString(attributes, "client_address", event.ClientAddress)
	attributes = appendAuditString(attributes, "account_id", event.AccountID)
	attributes = appendAuditString(attributes, "account_name", event.AccountName)
	attributes = appendAuditString(attributes, "session_id", event.SessionID)
	attributes = appendAuditString(attributes, "character_id", event.CharacterID)
	attributes = appendAuditString(attributes, "character_name", event.CharacterName)
	attributes = appendAuditString(attributes, "channel", event.Channel)
	attributes = appendAuditString(attributes, "game_id", event.GameID)
	attributes = appendAuditString(attributes, "game_name", event.GameName)

	attributes = appendAuditString(attributes, "game_reference", event.GameReference)
	if event.MessageBytes > 0 {
		attributes = append(attributes, slog.Int("message_bytes", event.MessageBytes))
	}

	logger.LogAttrs(ctx, level, "realm audit", attributes...)
}

// appendAuditString contains append audit string within the audit boundary so callers do not duplicate its
// domain-specific policy.
func appendAuditString(attributes []slog.Attr, key, value string) []slog.Attr {
	if value == "" {
		return attributes
	}

	return append(attributes, slog.String(key, value))
}

// recordAudit contains record audit within the audit boundary so callers do not duplicate its domain-specific policy.
func (control *ControlPlane) recordAudit(ctx context.Context, event AuditEvent, err error) {
	if control == nil || control.audit == nil {
		return
	}

	event.Version = RealmAuditVersion
	event.ClientAddress = auditClientAddress(ctx)
	event.Outcome, event.ErrorCode = auditResult(err)
	control.audit.Record(ctx, event)
}

// auditResult contains audit result within the audit boundary so callers do not duplicate its domain-specific policy.
func auditResult(err error) (string, string) {
	if err == nil {
		return "success", ""
	}

	switch {
	case errors.Is(err, ErrRealmSession), errors.Is(err, ErrAccountCredentials), errors.Is(err, ErrAccountUnverified):
		return "denied", "unauthorized"
	case errors.Is(err, ErrAccountExists), errors.Is(err, ErrCharacterExists), errors.Is(err, ErrGameExists):
		return "denied", "already_exists"
	case errors.Is(err, ErrCharacterNotFound), errors.Is(err, ErrGameNotFound):
		return "denied", "not_found"
	case errors.Is(err, ErrGameFull), errors.Is(err, ErrCharacterLimit):
		return "denied", "capacity"
	case errors.Is(err, ErrGamePassword):
		return "denied", "invalid_password"
	case errors.Is(err, ErrGameLevelRange):
		return "denied", "level_restricted"
	case errors.Is(err, ErrChannelMember):
		return "denied", "membership_required"
	case errors.Is(err, ErrCharacterOwner):
		return "denied", "ownership"
	case errors.Is(err, ErrCharacterLeased):
		return "denied", "leased"
	case errors.Is(err, ErrAccountInput),
		errors.Is(err, ErrAccountChallenge),
		errors.Is(err, ErrCharacterInput),
		errors.Is(err, ErrChannelInput),
		errors.Is(err, ErrGameDirectoryInput),
		errors.Is(err, ErrHTTPInput):
		return "denied", "invalid_input"
	case errors.Is(err, context.Canceled):
		return "failed", "canceled"
	case errors.Is(err, context.DeadlineExceeded):
		return "failed", "deadline_exceeded"
	case errors.Is(err, ErrGameUnavailable), errors.Is(err, ErrWorker), errors.Is(err, ErrAdmission):
		return "failed", "unavailable"
	default:
		return "failed", "internal"
	}
}

type auditContextKey struct{}

// withAuditClientAddress contains with audit client address within the audit boundary so callers do not duplicate its
// domain-specific policy.
func withAuditClientAddress(ctx context.Context, remoteAddress string) context.Context {
	address := strings.TrimSpace(remoteAddress)
	if host, _, err := net.SplitHostPort(address); err == nil {
		address = host
	}

	return context.WithValue(ctx, auditContextKey{}, address)
}

// auditClientAddress contains audit client address within the audit boundary so callers do not duplicate its
// domain-specific policy.
func auditClientAddress(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	address, _ := ctx.Value(auditContextKey{}).(string)

	return address
}
