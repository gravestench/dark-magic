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
	Version     uint32 `json:"version"`
	SessionID   string `json:"session_id"`
	PrincipalID string `json:"principal_id"`
	CharacterID string `json:"character_id"`
	PlayerID    string `json:"player_id"`
	ExpiresUnix int64  `json:"expires_unix"`
	Nonce       string `json:"nonce"`
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

func NewTicketAuthority(key []byte, sessionID string) (*TicketAuthority, error) {
	if len(key) < minimumTicketKeyBytes || strings.TrimSpace(sessionID) == "" {
		return nil, errors.New("game server ticket: a 32-byte key and session ID are required")
	}
	return &TicketAuthority{key: append([]byte(nil), key...), sessionID: sessionID, now: time.Now, consumed: make(map[string]int64)}, nil
}

func (authority *TicketAuthority) Issue(principal Principal, lifetime time.Duration) (string, error) {
	if strings.TrimSpace(principal.ID) == "" || strings.TrimSpace(principal.CharacterID) == "" || strings.TrimSpace(principal.PlayerID) == "" || lifetime <= 0 {
		return "", ErrAuthentication
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return "", err
	}
	ticket := admissionTicket{
		Version: SessionProtocolVersion, SessionID: authority.sessionID,
		PrincipalID: principal.ID, CharacterID: principal.CharacterID, PlayerID: principal.PlayerID,
		ExpiresUnix: authority.now().Add(lifetime).Unix(), Nonce: hex.EncodeToString(nonce[:]),
	}
	payload, err := json.Marshal(ticket)
	if err != nil {
		return "", err
	}
	signature := authority.sign(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (authority *TicketAuthority) Authenticate(_ context.Context, credential string) (Principal, error) {
	parts := strings.Split(credential, ".")
	if len(parts) != 2 {
		return Principal{}, ErrAuthentication
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || len(payload) > MaxTicketPayloadBytes {
		return Principal{}, ErrAuthentication
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, authority.sign(payload)) {
		return Principal{}, ErrAuthentication
	}
	var ticket admissionTicket
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&ticket); err != nil {
		return Principal{}, ErrAuthentication
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Principal{}, ErrAuthentication
	}
	if ticket.Version != SessionProtocolVersion || ticket.SessionID != authority.sessionID || ticket.ExpiresUnix <= authority.now().Unix() || ticket.Nonce == "" || ticket.PrincipalID == "" || ticket.CharacterID == "" || ticket.PlayerID == "" {
		return Principal{}, ErrAuthentication
	}
	authority.mu.Lock()
	defer authority.mu.Unlock()
	for nonce, expiry := range authority.consumed {
		if expiry <= authority.now().Unix() {
			delete(authority.consumed, nonce)
		}
	}
	if _, used := authority.consumed[ticket.Nonce]; used {
		return Principal{}, ErrAuthentication
	}
	authority.consumed[ticket.Nonce] = ticket.ExpiresUnix
	return Principal{ID: ticket.PrincipalID, CharacterID: ticket.CharacterID, PlayerID: ticket.PlayerID}, nil
}

func (authority *TicketAuthority) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, authority.key)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

const MaxTicketPayloadBytes = 2 << 10
