package gameserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
)

// SetSnapshotPending identifies the one expected projection error while a
// trusted next-tick admission command is waiting to materialize its player.
func (endpoint *Endpoint) SetSnapshotPending(classify func(error) bool) {
	endpoint.snapshotPending = classify
}

// SetLeave installs the mod-owned membership cleanup command. Authentication
// and credential revocation remain protocol policy; removing live entities belongs to the active game rules.
func (endpoint *Endpoint) SetLeave(leave func(Principal) error) { endpoint.leave = leave }

// SetConnected observes successful initial joins and reconnects for Realm worker lifecycle accounting.
func (endpoint *Endpoint) SetConnected(connected func(Principal)) { endpoint.connected = connected }

// NewEndpoint validates transport-independent dependencies and initializes isolated membership state.
func NewEndpoint(host *Host, auth Authenticator, project SnapshotProjector) (*Endpoint, error) {
	if host == nil || host.Session == nil || auth == nil || project == nil {
		return nil, errors.New("game server protocol: host, authenticator, and projector are required")
	}

	return &Endpoint{
		host:           host,
		auth:           auth,
		project:        project,
		connections:    make(map[string]connection),
		watches:        make(map[string]bool),
		reconnects:     make(map[string]reconnectReplay),
		now:            time.Now,
		reconnectGrace: reconnectGrace,
		after:          func(delay time.Duration, callback func()) { time.AfterFunc(delay, callback) },
	}, nil
}

// Join authenticates identity, admits the exact runtime, and publishes a snapshot only after
// player projection is ready.
func (endpoint *Endpoint) Join(ctx context.Context, request JoinRequest) (JoinResponse, error) {
	if request.Version != SessionProtocolVersion {
		return JoinResponse{}, ErrProtocol
	}

	principal, err := endpoint.authenticateJoin(ctx, request)
	if err != nil {
		return JoinResponse{}, err
	}

	admission, err := endpoint.host.Admit(principal.CharacterID, request.Identity)
	if err != nil {
		return JoinResponse{}, err
	}

	credential, err := newSessionCredential()
	if err != nil {
		return JoinResponse{}, fmt.Errorf("game server protocol: create credential: %w", err)
	}

	endpoint.registerConnection(credential, principal, admission)

	snapshot, err := endpoint.joinSnapshot(ctx, principal.PlayerID)
	if err != nil {
		endpoint.removeConnection(credential)

		return JoinResponse{}, err
	}

	if endpoint.connected != nil {
		endpoint.connected(principal)
	}

	return JoinResponse{Credential: credential, Admission: admission, Snapshot: snapshot}, nil
}

// authenticateJoin rejects incomplete principals and tickets pinned to a different running authority.
func (endpoint *Endpoint) authenticateJoin(ctx context.Context, request JoinRequest) (Principal, error) {
	principal, err := endpoint.auth.Authenticate(ctx, request.Credential)
	if err != nil || strings.TrimSpace(principal.ID) == "" || strings.TrimSpace(principal.CharacterID) == "" ||
		strings.TrimSpace(principal.PlayerID) == "" {
		return Principal{}, ErrAuthentication
	}

	if principal.RuntimeIdentityHash != "" && principal.RuntimeIdentityHash != endpoint.host.Allocation.IdentityHash {
		return Principal{}, ErrAuthentication
	}

	return principal, nil
}

// registerConnection installs full rate-limit capacity before any client can use the fresh credential.
func (endpoint *Endpoint) registerConnection(
	credential SessionCredential,
	principal Principal,
	admission gamesession.AdmissionToken,
) {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()

	now := endpoint.now()
	endpoint.connections[string(credential)] = connection{
		principal: principal,
		admission: admission,
		commands:  newTokenBucket(commandBurst, commandRate, now),
		refreshes: newTokenBucket(refreshBurst, refreshRate, now),
		connected: true,
	}
}

// removeConnection rolls back a credential whose admitted player never became projectable.
func (endpoint *Endpoint) removeConnection(credential SessionCredential) {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()

	delete(endpoint.connections, string(credential))
}

// joinSnapshot polls only the expected admission-pending error and bounds readiness independently from caller timeout.
func (endpoint *Endpoint) joinSnapshot(ctx context.Context, playerID string) (Snapshot, error) {
	deadline := time.NewTimer(joinReadyTimeout)
	defer deadline.Stop()

	for {
		snapshot, err := endpoint.snapshot(playerID)
		if err == nil || endpoint.snapshotPending == nil || !endpoint.snapshotPending(err) {
			return snapshot, err
		}

		timer := time.NewTimer(joinReadyPoll)
		select {
		case <-ctx.Done():
			timer.Stop()

			return Snapshot{}, ctx.Err()
		case <-deadline.C:
			timer.Stop()

			return Snapshot{}, fmt.Errorf("game server protocol: player admission readiness: %w", err)
		case <-timer.C:
		}
	}
}
