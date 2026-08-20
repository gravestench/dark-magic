package session

import (
	"errors"
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

// TestPredictionReconciliationAlwaysAppliesServerCanonicalOutcome proves untrusted prediction cannot mutate authority.
func TestPredictionReconciliationAlwaysAppliesServerCanonicalOutcome(t *testing.T) {
	serverEngine := gameecs.New()

	server, err := New(serverEngine, Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestResource(t, server)

	if err := server.Step(); err != nil {
		t.Fatal(err)
	}

	authoritative, err := server.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	predictedEngine := gameecs.New()
	defer closeTestResource(t, predictedEngine)

	predicted, err := predictedEngine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	predicted.Tick = 99 // an untrusted client claims a different future

	reconciliation, err := ReconcilePrediction(PredictionLimited, &predicted, authoritative)
	if err != nil {
		t.Fatal(err)
	}

	if !reconciliation.Corrected || reconciliation.Difference == "" {
		t.Fatalf("reconciliation = %#v", reconciliation)
	}

	if err := reconciliation.Apply(predictedEngine); err != nil {
		t.Fatal(err)
	}

	corrected, err := predictedEngine.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if difference := gameecs.FirstDifference(*authoritative.Snapshot, corrected); difference != "" {
		t.Fatalf("corrected client differs: %s", difference)
	}

	stillAuthoritative, err := server.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	if stillAuthoritative.Checksum != authoritative.Checksum {
		t.Fatal("untrusted prediction changed the server outcome")
	}
}

// TestPredictionMayBeDisabledOrAlreadyCanonical distinguishes disabled prediction from an already matching client.
func TestPredictionMayBeDisabledOrAlreadyCanonical(t *testing.T) {
	engine := gameecs.New()

	session, err := New(engine, Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestResource(t, session)

	authoritative, err := session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	disabled, err := ReconcilePrediction(PredictionNone, nil, authoritative)
	if err != nil {
		t.Fatal(err)
	}

	if !disabled.Corrected || disabled.Difference != "prediction disabled" {
		t.Fatalf("disabled reconciliation = %#v", disabled)
	}

	matching, err := ReconcilePrediction(PredictionLimited, authoritative.Snapshot, authoritative)
	if err != nil {
		t.Fatal(err)
	}

	if matching.Corrected || matching.Difference != "" {
		t.Fatalf("matching reconciliation = %#v", matching)
	}
}

// TestPredictionReconciliationRejectsUnverifiedServerFrames prevents corrupted authority data from reaching clients.
func TestPredictionReconciliationRejectsUnverifiedServerFrames(t *testing.T) {
	engine := gameecs.New()

	session, err := New(engine, Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer closeTestResource(t, session)

	authoritative, err := session.CanonicalCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	authoritative.Checksum = "tampered"
	if _, err := ReconcilePrediction(PredictionLimited, nil, authoritative); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("tampered frame error = %v", err)
	}
}
