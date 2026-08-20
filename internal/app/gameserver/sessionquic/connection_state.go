package sessionquic

import (
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
)

const (
	packageBytesPerSecond = 2 << 20
	packageBurstBytes     = 4 << 20
)

// connectionMemberships binds credential cleanup and package bandwidth to one physical QUIC connection.
type connectionMemberships struct {
	mu          sync.Mutex
	credentials map[gameserver.SessionCredential]struct{}
	packages    *packageRateLimiter
}

// newConnectionMemberships starts a connection with no players and a full package-transfer burst allowance.
func newConnectionMemberships() *connectionMemberships {
	return &connectionMemberships{
		credentials: make(map[gameserver.SessionCredential]struct{}),
		packages:    newPackageRateLimiter(),
	}
}

// observe transfers cleanup ownership whenever a successful operation creates, rotates, or removes a credential.
func (memberships *connectionMemberships) observe(message request, result response) {
	memberships.mu.Lock()
	defer memberships.mu.Unlock()

	if result.Join != nil && result.Join.Credential != "" {
		memberships.credentials[result.Join.Credential] = struct{}{}
	}

	if message.Operation == operationLeave && result.Error == "" {
		delete(memberships.credentials, message.Credential)
	}

	if message.Operation == operationReconnect && result.Join != nil {
		delete(memberships.credentials, message.Reconnect.Credential)

		if result.Join.Credential != "" {
			memberships.credentials[result.Join.Credential] = struct{}{}
		}
	}
}

// snapshot copies credentials before cleanup calls Endpoint, avoiding a lock inversion across package boundaries.
func (memberships *connectionMemberships) snapshot() []gameserver.SessionCredential {
	memberships.mu.Lock()
	defer memberships.mu.Unlock()

	result := make([]gameserver.SessionCredential, 0, len(memberships.credentials))
	for credential := range memberships.credentials {
		result = append(result, credential)
	}

	return result
}

// packageRateLimiter is a connection-local byte bucket, preventing parallel streams from multiplying bandwidth.
type packageRateLimiter struct {
	mu     sync.Mutex
	tokens float64
	last   time.Time
}

// newPackageRateLimiter permits a bounded initial burst so ordinary extension downloads do not start artificially.
func newPackageRateLimiter() *packageRateLimiter {
	return &packageRateLimiter{tokens: packageBurstBytes}
}

// Allow refills from elapsed wall time and atomically reserves the requested response allowance.
func (limiter *packageRateLimiter) Allow(size int, now time.Time) bool {
	if limiter == nil || size <= 0 {
		return false
	}

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if limiter.last.IsZero() {
		limiter.last = now
	} else if elapsed := now.Sub(limiter.last); elapsed > 0 {
		limiter.tokens = min(
			float64(packageBurstBytes),
			limiter.tokens+elapsed.Seconds()*packageBytesPerSecond,
		)
		limiter.last = now
	}

	if float64(size) > limiter.tokens {
		return false
	}

	limiter.tokens -= float64(size)

	return true
}
