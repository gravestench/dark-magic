package gameserver

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// Leave immediately revokes a credential; Realm layers remain responsible for leases and durable save commits.
func (endpoint *Endpoint) Leave(credential SessionCredential) error {
	return endpoint.leaveCredential(credential, true)
}

// Disconnect suspends a membership for a bounded reconnect lease instead of deleting its player immediately.
func (endpoint *Endpoint) Disconnect(credential SessionCredential) {
	key := string(credential)

	endpoint.mu.Lock()

	member, found := endpoint.connections[key]
	if !found || credential == "" || !member.connected {
		endpoint.mu.Unlock()

		return
	}

	member.connected = false
	member.disconnectGeneration++
	generation := member.disconnectGeneration
	endpoint.connections[key] = member
	delete(endpoint.watches, key)
	endpoint.mu.Unlock()

	endpoint.after(endpoint.reconnectGrace, func() { endpoint.expireDisconnect(credential, generation) })
}

// expireDisconnect removes only the disconnected generation that scheduled this callback.
// A reconnect or later disconnect therefore makes an older timer harmless.
func (endpoint *Endpoint) expireDisconnect(credential SessionCredential, generation uint64) {
	key := string(credential)

	endpoint.mu.Lock()

	member, found := endpoint.connections[key]
	if !found || member.connected || member.disconnectGeneration != generation {
		endpoint.mu.Unlock()

		return
	}

	delete(endpoint.connections, key)
	delete(endpoint.watches, key)
	endpoint.mu.Unlock()

	if endpoint.leave != nil {
		_ = endpoint.leave(member.principal)
	}
}

// leaveCredential centralizes explicit and cleanup-driven revocation while preserving their different strictness.
func (endpoint *Endpoint) leaveCredential(credential SessionCredential, strict bool) error {
	endpoint.mu.Lock()

	member, found := endpoint.connections[string(credential)]
	if !found || credential == "" {
		endpoint.mu.Unlock()

		if strict {
			return ErrAuthentication
		}

		return nil
	}

	delete(endpoint.connections, string(credential))
	delete(endpoint.watches, string(credential))
	endpoint.mu.Unlock()

	if endpoint.leave != nil {
		return endpoint.leave(member.principal)
	}

	return nil
}

// Reconnect validates the retained membership, rotates its bearer credential, and caches one idempotent response.
func (endpoint *Endpoint) Reconnect(request ReconnectRequest) (JoinResponse, error) {
	if len(request.Nonce) < 32 || len(request.Nonce) > 128 {
		return JoinResponse{}, ErrAuthentication
	}

	if replay, found := endpoint.replayedReconnect(request); found {
		return replay, nil
	}

	member, err := endpoint.membership(request.Credential, true)
	if err != nil {
		return JoinResponse{}, err
	}

	if err := endpoint.host.ValidateReconnect(member.admission, request.Identity); err != nil {
		return JoinResponse{}, err
	}

	credential, err := newSessionCredential()
	if err != nil {
		return JoinResponse{}, fmt.Errorf("game server protocol: rotate credential: %w", err)
	}

	snapshot, err := endpoint.snapshot(member.principal.PlayerID)
	if err != nil {
		return JoinResponse{}, err
	}

	response, principal, replayed, err := endpoint.rotateReconnectCredential(request, credential, member, snapshot)
	if err != nil {
		return JoinResponse{}, err
	}

	if replayed {
		return response, nil
	}

	if endpoint.connected != nil {
		endpoint.connected(principal)
	}

	endpoint.scheduleReconnectReplayExpiry(request)

	return response, nil
}

// rotateReconnectCredential revalidates old state under lock so concurrent capacity debits and rotations are preserved.
func (endpoint *Endpoint) rotateReconnectCredential(
	request ReconnectRequest,
	credential SessionCredential,
	member connection,
	snapshot Snapshot,
) (JoinResponse, Principal, bool, error) {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()

	current, found := endpoint.connections[string(request.Credential)]
	if !found || current.principal.ID != member.principal.ID {
		if replay, replayed := endpoint.reconnects[string(request.Credential)]; replayed && replay.nonce == request.Nonce {
			return replay.response, Principal{}, true, nil
		}

		return JoinResponse{}, Principal{}, false, ErrAuthentication
	}

	// Transfer the state observed while holding the lock. A concurrent request
	// may have consumed capacity since membership() returned.
	member = current
	member.connected = true

	delete(endpoint.connections, string(request.Credential))
	delete(endpoint.watches, string(request.Credential))
	endpoint.connections[string(credential)] = member

	response := JoinResponse{Credential: credential, Admission: member.admission, Snapshot: snapshot}
	endpoint.reconnects[string(request.Credential)] = reconnectReplay{nonce: request.Nonce, response: response}

	return response, member.principal, false, nil
}

// scheduleReconnectReplayExpiry bounds old-credential idempotency to the same grace window as disconnected membership.
func (endpoint *Endpoint) scheduleReconnectReplayExpiry(request ReconnectRequest) {
	endpoint.after(endpoint.reconnectGrace, func() {
		endpoint.mu.Lock()
		defer endpoint.mu.Unlock()

		if replay, found := endpoint.reconnects[string(request.Credential)]; found && replay.nonce == request.Nonce {
			delete(endpoint.reconnects, string(request.Credential))
		}
	})
}

// replayedReconnect returns only a nonce-matched replay so a reused old credential cannot obtain another rotation.
func (endpoint *Endpoint) replayedReconnect(request ReconnectRequest) (JoinResponse, bool) {
	endpoint.mu.RLock()
	defer endpoint.mu.RUnlock()

	replay, found := endpoint.reconnects[string(request.Credential)]
	if !found || replay.nonce != request.Nonce {
		return JoinResponse{}, false
	}

	return replay.response, true
}

// connection requires an actively connected membership for commands and observation.
func (endpoint *Endpoint) connection(credential SessionCredential) (connection, error) {
	return endpoint.membership(credential, false)
}

// membership reads one credential and optionally admits its disconnected reconnect-grace state.
func (endpoint *Endpoint) membership(
	credential SessionCredential,
	allowDisconnected bool,
) (connection, error) {
	endpoint.mu.RLock()
	defer endpoint.mu.RUnlock()

	member, found := endpoint.connections[string(credential)]
	if !found || credential == "" || (!allowDisconnected && !member.connected) {
		return connection{}, ErrAuthentication
	}

	return member, nil
}

// newSessionCredential uses 256 bits of randomness so credentials can serve directly as bearer secrets.
func newSessionCredential() (SessionCredential, error) {
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}

	return SessionCredential(hex.EncodeToString(value[:])), nil
}
