package serverapp

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestRemoteProfileAdmissionThrottleIsPerClientAndRefills verifies that failed
// authentication consumes capacity, client buckets remain isolated, and one
// elapsed second restores exactly one admission opportunity.
func TestRemoteProfileAdmissionThrottleIsPerClientAndRefills(t *testing.T) {
	fixture := newRemoteProfileFixture(t, 1000)
	admissions := newRemoteProfileAdmissionsForTest(t, fixture, RemoteProfileConfig{
		Credential:  "secret",
		PrincipalID: "principal",
		PlayerID:    "player",
		Destination: fixture.destination,
		Lifetime:    time.Minute,
	})
	now := time.Unix(100, 0)
	admissions.now = func() time.Time { return now }

	clientA := WithProfileAdmissionClient(context.Background(), "192.0.2.1:6112")
	for range int(remoteProfileBurst) {
		if _, err := admissions.Admit(clientA, "wrong", fixture.offer); !errors.Is(
			err,
			ErrRemoteProfileAdmission,
		) {
			t.Fatalf("wrong credential error = %v", err)
		}
	}

	if _, err := admissions.Admit(clientA, "secret", fixture.offer); !errors.Is(
		err,
		ErrRemoteProfileAdmission,
	) {
		t.Fatalf("exhausted client error = %v", err)
	}

	clientB := WithProfileAdmissionClient(context.Background(), "192.0.2.2:6112")

	if _, err := admissions.Admit(clientB, "secret", fixture.offer); err != nil {
		t.Fatalf("independent client was throttled: %v", err)
	}

	now = now.Add(time.Second)

	if _, err := admissions.Admit(clientA, "secret", fixture.offer); err != nil {
		t.Fatalf("refilled client error = %v", err)
	}
}
