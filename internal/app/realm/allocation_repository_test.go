package realm

import (
	"errors"
	"testing"
)

func TestMemoryAllocationsPreserveLifecycleAndDefensiveIdentity(t *testing.T) {
	store := NewMemoryAllocations()
	requested, err := store.Request(t.Context(), "game", "allocation")
	if err != nil || requested.State != AllocationRequested {
		t.Fatalf("requested = %#v, %v", requested, err)
	}
	if _, err := store.Request(t.Context(), "game", "other"); !errors.Is(err, ErrGameExists) {
		t.Fatalf("duplicate request error = %v", err)
	}
	if repeated, err := store.Request(t.Context(), "game", "allocation"); err != nil || repeated.AllocationID != requested.AllocationID {
		t.Fatalf("repeated request = %#v, %v", repeated, err)
	}
	identity := orchestrationIdentity()
	endpoint := GameEndpoint{Address: "127.0.0.1:4000", TLSFingerprint: "sha256:worker"}
	ready, err := store.Ready(t.Context(), "game", endpoint, identity)
	if err != nil || ready.State != AllocationReady || ready.LastHealthyAt == nil {
		t.Fatalf("ready = %#v, %v", ready, err)
	}
	if repeated, err := store.Ready(t.Context(), "game", endpoint, identity); err != nil || repeated.State != AllocationReady {
		t.Fatalf("repeated ready = %#v, %v", repeated, err)
	}
	replacementEndpoint := GameEndpoint{Address: "127.0.0.1:4001", TLSFingerprint: "sha256:replacement"}
	restored, err := store.RestoreReady(t.Context(), "game", "allocation", replacementEndpoint, identity)
	if err != nil || restored.Endpoint != replacementEndpoint || restored.AllocationID != "allocation" {
		t.Fatalf("restored readiness = %#v, %v", restored, err)
	}
	mismatched := identity
	mismatched.Recipe.ConfigurationHash = "other"
	if _, err := store.RestoreReady(t.Context(), "game", "allocation", endpoint, mismatched); !errors.Is(err, ErrAllocationRecord) {
		t.Fatalf("mismatched restoration error = %v", err)
	}
	ready.Runtime.Recipe.EngineAPI = "mutated"
	active, err := store.Active(t.Context())
	if err != nil || len(active) != 1 || active[0].Runtime.Recipe.EngineAPI != identity.Recipe.EngineAPI {
		t.Fatalf("active = %#v, %v", active, err)
	}
	if err := store.Healthy(t.Context(), "game"); err != nil {
		t.Fatal(err)
	}
	if err := store.Fail(t.Context(), "game", errors.New("worker stopped")); err != nil {
		t.Fatal(err)
	}
	if err := store.Fail(t.Context(), "game", errors.New("worker stopped")); err != nil {
		t.Fatalf("repeated failure = %v", err)
	}
	if err := store.Healthy(t.Context(), "game"); !errors.Is(err, ErrAllocationRecord) {
		t.Fatalf("terminal allocation became healthy: %v", err)
	}
	if active, err := store.Active(t.Context()); err != nil || len(active) != 0 {
		t.Fatalf("failed allocation remained active: %#v, %v", active, err)
	}
	if _, err := store.Ready(t.Context(), "game", endpoint, identity); !errors.Is(err, ErrAllocationRecord) {
		t.Fatalf("terminal allocation became ready: %v", err)
	}
}
