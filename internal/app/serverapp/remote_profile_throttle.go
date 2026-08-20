package serverapp

import (
	"context"
	"net/netip"
	"time"
)

const (
	remoteProfileBurst = 8.0
	remoteProfileRate  = 1.0
)

type profileAdmissionBucket struct {
	tokens  float64
	updated time.Time
}

type profileAdmissionClientKey struct{}

// take consumes one token from the client's bucket while the caller holds the
// admissions lock. Authentication failures intentionally consume capacity so
// repeated secret guesses are throttled too.
func (admissions *RemoteProfileAdmissions) take(client string) bool {
	now := admissions.now()

	bucket, found := admissions.clients[client]
	if !found {
		bucket = profileAdmissionBucket{tokens: remoteProfileBurst, updated: now}
	}
	// A clock moving backward must not subtract tokens or move the refill
	// watermark behind the last observed time.
	if now.Before(bucket.updated) {
		now = bucket.updated
	}

	bucket.tokens = min(remoteProfileBurst, bucket.tokens+now.Sub(bucket.updated).Seconds()*remoteProfileRate)
	bucket.updated = now

	if bucket.tokens < 1 {
		admissions.clients[client] = bucket
		return false
	}

	bucket.tokens--
	admissions.clients[client] = bucket

	return true
}

// WithProfileAdmissionClient lets transports bind admission throttling to a
// normalized remote IP without expanding the profile interface or trusting a
// client-supplied identifier.
func WithProfileAdmissionClient(ctx context.Context, address string) context.Context {
	host := address
	if parsed, err := netip.ParseAddrPort(address); err == nil {
		host = parsed.Addr().Unmap().String()
	}

	return context.WithValue(ctx, profileAdmissionClientKey{}, host)
}

// WithClient satisfies the transport admission interface while keeping the
// context key and normalization policy private to this package.
func (admissions *RemoteProfileAdmissions) WithClient(ctx context.Context, address string) context.Context {
	return WithProfileAdmissionClient(ctx, address)
}

// profileAdmissionClient returns a shared fallback bucket when a transport did
// not attach an address, so missing metadata cannot bypass throttling.
func profileAdmissionClient(ctx context.Context) string {
	if ctx == nil {
		return "unknown"
	}

	value, ok := ctx.Value(profileAdmissionClientKey{}).(string)
	if !ok || value == "" {
		return "unknown"
	}

	return value
}
