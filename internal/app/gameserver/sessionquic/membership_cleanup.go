package sessionquic

import (
	"context"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
)

// membershipCleanup serializes one membership's rotation and socket cleanup, not other players' readiness.
// Only the current credential and its last reconnect input can name the retained membership.
type membershipCleanup struct {
	mu       sync.Mutex
	current  gameserver.SessionCredential
	previous gameserver.SessionCredential
	owner    *connectionMemberships
}

func (server *Server) dispatchConnection(
	ctx context.Context, message request, memberships *connectionMemberships,
) response {
	if !validShape(message) {
		return response{Error: ErrWire.Error()}
	}

	if message.Operation != operationJoin && message.Operation != operationReconnect &&
		message.Operation != operationLeave {
		return server.dispatchRateLimited(ctx, message, memberships.packages, time.Now())
	}

	memberships.operationMu.Lock()
	defer memberships.operationMu.Unlock()

	if memberships.closed {
		return response{Error: gameserver.ErrAuthentication.Error()}
	}

	if message.Operation == operationJoin {
		result := server.dispatchRateLimited(ctx, message, memberships.packages, time.Now())
		if result.Join != nil {
			cleanup := &membershipCleanup{current: result.Join.Credential, owner: memberships}
			server.publishCleanup(cleanup, "", "")
			memberships.observe(message, result)
		}

		return result
	}

	credential := message.Credential
	if message.Operation == operationReconnect {
		credential = message.Reconnect.Credential
	}

	cleanup := server.lockCleanup(credential)
	defer cleanup.mu.Unlock()

	if cleanup.owner == nil {
		defer func() {
			if cleanup.owner == nil {
				server.removeCleanup(cleanup)
			}
		}()
	}

	if memberships.isSuperseded(credential) {
		return response{Error: gameserver.ErrAuthentication.Error()}
	}

	result := server.dispatchRateLimited(ctx, message, memberships.packages, time.Now())
	if result.Join != nil {
		current, previous := cleanup.current, cleanup.previous
		if cleanup.owner != nil && cleanup.owner != memberships {
			cleanup.owner.retire(current, previous, credential, result.Join.Credential)
		}

		if result.Join.Credential != current {
			cleanup.previous = credential
			cleanup.current = result.Join.Credential
		}

		cleanup.owner = memberships
		server.publishCleanup(cleanup, current, previous)
	} else if message.Operation == operationLeave && result.Error == "" {
		if cleanup.owner != nil {
			cleanup.owner.observe(message, result)
		}

		cleanup.owner = nil
		server.removeCleanup(cleanup)
	}

	memberships.observe(message, result)

	return result
}

// lockCleanup retries only when another operation retired the alias while its lock was acquired.
func (server *Server) lockCleanup(credential gameserver.SessionCredential) *membershipCleanup {
	for {
		server.membershipMu.Lock()
		if server.memberships == nil {
			server.memberships = make(map[gameserver.SessionCredential]*membershipCleanup)
		}

		cleanup := server.memberships[credential]
		if cleanup == nil {
			cleanup = &membershipCleanup{current: credential}
			server.memberships[credential] = cleanup
		}
		server.membershipMu.Unlock()

		cleanup.mu.Lock()
		server.membershipMu.Lock()
		current := server.memberships[credential] == cleanup
		server.membershipMu.Unlock()

		if current {
			return cleanup
		}
		cleanup.mu.Unlock()
	}
}

func (server *Server) publishCleanup(cleanup *membershipCleanup, current, previous gameserver.SessionCredential) {
	server.membershipMu.Lock()
	defer server.membershipMu.Unlock()

	if server.memberships == nil {
		server.memberships = make(map[gameserver.SessionCredential]*membershipCleanup)
	}

	for _, credential := range []gameserver.SessionCredential{current, previous} {
		if server.memberships[credential] == cleanup {
			delete(server.memberships, credential)
		}
	}

	server.memberships[cleanup.current] = cleanup
	if cleanup.previous != "" {
		server.memberships[cleanup.previous] = cleanup
	}
}

func (server *Server) removeCleanup(cleanup *membershipCleanup) {
	server.membershipMu.Lock()
	defer server.membershipMu.Unlock()

	for _, credential := range []gameserver.SessionCredential{cleanup.current, cleanup.previous} {
		if server.memberships[credential] == cleanup {
			delete(server.memberships, credential)
		}
	}
}
