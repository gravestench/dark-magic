package gameserver

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"time"
)

const minimumTicketKeyBytes = 32

type admissionTicket struct {
	Version             uint32 `json:"version"`
	SessionID           string `json:"session_id"`
	PrincipalID         string `json:"principal_id"`
	CharacterID         string `json:"character_id"`
	PlayerID            string `json:"player_id"`
	CharacterRevision   uint64 `json:"character_revision"`
	RuntimeIdentityHash string `json:"runtime_identity_hash"`
	ExpiresUnix         int64  `json:"expires_unix"`
	Nonce               string `json:"nonce"`
}

// TicketAuthority issues and consumes short-lived, session-bound admission
// tickets. Consumption is atomic so a captured ticket cannot open two sessions.
type TicketAuthority struct {
	mu        sync.Mutex
	key       []byte
	sessionID string
	now       func() time.Time
	consumed  map[string]int64
}

// NewTicketAuthority defensively owns its signing key so caller mutation cannot invalidate outstanding tickets.
func NewTicketAuthority(key []byte, sessionID string) (*TicketAuthority, error) {
	if len(key) < minimumTicketKeyBytes || strings.TrimSpace(sessionID) == "" {
		return nil, errors.New("game server ticket: a 32-byte key and session ID are required")
	}

	return &TicketAuthority{
		key:       append([]byte(nil), key...),
		sessionID: sessionID,
		now:       time.Now,
		consumed:  make(map[string]int64),
	}, nil
}

// Issue signs a short-lived principal claim whose random nonce becomes the atomic one-time-use key.
func (authority *TicketAuthority) Issue(principal Principal, lifetime time.Duration) (string, error) {
	if strings.TrimSpace(principal.ID) == "" || strings.TrimSpace(principal.CharacterID) == "" ||
		strings.TrimSpace(principal.PlayerID) == "" || lifetime <= 0 {
		return "", ErrAuthentication
	}

	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}

	ticket := admissionTicket{
		Version:             SessionProtocolVersion,
		SessionID:           authority.sessionID,
		PrincipalID:         principal.ID,
		CharacterID:         principal.CharacterID,
		PlayerID:            principal.PlayerID,
		CharacterRevision:   principal.CharacterRevision,
		RuntimeIdentityHash: principal.RuntimeIdentityHash,
		ExpiresUnix:         authority.now().Add(lifetime).Unix(),
		Nonce:               hex.EncodeToString(nonce[:]),
	}

	payload, err := json.Marshal(ticket)
	if err != nil {
		return "", err
	}

	signature := authority.sign(payload)

	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

// Authenticate verifies expiry before atomically consuming the nonce, preventing a captured ticket from joining twice.
func (authority *TicketAuthority) Authenticate(_ context.Context, credential string) (Principal, error) {
	ticket, err := authority.verify(credential)
	if err != nil {
		return Principal{}, err
	}

	if ticket.ExpiresUnix <= authority.now().Unix() {
		return Principal{}, ErrAuthentication
	}

	authority.mu.Lock()
	defer authority.mu.Unlock()

	authority.purgeLocked()

	if _, used := authority.consumed[ticket.Nonce]; used {
		return Principal{}, ErrAuthentication
	}

	authority.consumed[ticket.Nonce] = ticket.ExpiresUnix

	return Principal{
		ID:                  ticket.PrincipalID,
		CharacterID:         ticket.CharacterID,
		PlayerID:            ticket.PlayerID,
		CharacterRevision:   ticket.CharacterRevision,
		RuntimeIdentityHash: ticket.RuntimeIdentityHash,
	}, nil
}

// Revoke invalidates an issued ticket during admission rollback.
func (authority *TicketAuthority) Revoke(credential string) error {
	ticket, err := authority.verify(credential)
	if err != nil {
		return err
	}

	authority.mu.Lock()
	defer authority.mu.Unlock()

	authority.purgeLocked()
	authority.consumed[ticket.Nonce] = ticket.ExpiresUnix

	return nil
}

// verify authenticates the envelope and strictly decodes the session-bound claim without consuming it.
func (authority *TicketAuthority) verify(credential string) (admissionTicket, error) {
	parts := strings.Split(credential, ".")
	if len(parts) != 2 {
		return admissionTicket{}, ErrAuthentication
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || len(payload) > MaxTicketPayloadBytes {
		return admissionTicket{}, ErrAuthentication
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, authority.sign(payload)) {
		return admissionTicket{}, ErrAuthentication
	}

	var ticket admissionTicket

	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&ticket); err != nil {
		return admissionTicket{}, ErrAuthentication
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return admissionTicket{}, ErrAuthentication
	}

	if ticket.Version != SessionProtocolVersion || ticket.SessionID != authority.sessionID || ticket.Nonce == "" ||
		ticket.PrincipalID == "" || ticket.CharacterID == "" || ticket.PlayerID == "" {
		return admissionTicket{}, ErrAuthentication
	}

	return ticket, nil
}

// purgeLocked bounds replay memory to unexpired nonces; callers must hold authority.mu.
func (authority *TicketAuthority) purgeLocked() {
	for nonce, expiry := range authority.consumed {
		if expiry <= authority.now().Unix() {
			delete(authority.consumed, nonce)
		}
	}
}

// sign authenticates ticket bytes exactly as serialized, making unknown-field or payload rewrites detectable.
func (authority *TicketAuthority) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, authority.key)
	_, _ = mac.Write(payload)

	return mac.Sum(nil)
}

// MaxTicketPayloadBytes bounds decoded claims before JSON parsing allocates attacker-controlled structures.
const MaxTicketPayloadBytes = 2 << 10
