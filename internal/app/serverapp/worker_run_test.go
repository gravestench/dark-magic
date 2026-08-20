package serverapp

import (
	"context"
	"errors"
	"testing"
)

// TestWrapWorkerRunErrorTreatsCancellationAsShutdown ensures the shared cancel
// signal stays quiet while genuine service failures retain their cause.
func TestWrapWorkerRunErrorTreatsCancellationAsShutdown(t *testing.T) {
	if err := wrapWorkerRunError("QUIC", context.Canceled); err != nil {
		t.Fatalf("canceled service = %v", err)
	}

	cause := errors.New("listener failed")
	if err := wrapWorkerRunError("QUIC", cause); !errors.Is(err, cause) {
		t.Fatalf("service failure = %v", err)
	}
}

// TestRealmWorkerResultDistinguishesDrainFromServiceFailure protects the
// boundary between deliberate listener closure and unexpected service exit.
func TestRealmWorkerResultDistinguishesDrainFromServiceFailure(t *testing.T) {
	cause := errors.New("listener closed")
	if err := realmWorkerResult(true, cause); err != nil {
		t.Fatalf("controlled drain = %v", err)
	}

	if err := realmWorkerResult(false, cause); !errors.Is(err, cause) {
		t.Fatalf("service failure = %v", err)
	}

	if err := realmWorkerResult(false, nil); err == nil {
		t.Fatal("unexpected clean service exit succeeded")
	}
}
