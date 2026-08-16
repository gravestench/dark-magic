package session

import (
	"encoding/json"
	"errors"
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestRecoveryCheckpointPreservesPendingAndDuplicateSuppression(t *testing.T) {
	build := func(identity simulation.RuntimeIdentity) *Session {
		engine := gameecs.New()
		result, err := New(engine, Config{MaxCommandLead: 4, RollbackWindow: 8})
		if err != nil {
			t.Fatal(err)
		}
		stores := simulation.NewStateStore()
		if err := stores.Register("test.state", "test/v1", []byte(`{"value":1}`)); err != nil {
			t.Fatal(err)
		}
		if err := result.RegisterAuthoritativeRuntime(identity, stores); err != nil {
			t.Fatal(err)
		}
		if err := result.Register("move", CommandHandler{Validate: func(simulation.Command) error { return nil },
			Apply: func(*gameecs.Engine, simulation.Command) error { return nil }}); err != nil {
			t.Fatal(err)
		}
		return result
	}
	identity := compatibilityIdentity("recovery")
	source := build(identity)
	defer source.Close()
	applied := simulation.Command{Tick: 1, Player: "player", Sequence: 1, Kind: "move", Payload: json.RawMessage(`{"x":1}`)}
	pending := simulation.Command{Tick: 3, Player: "player", Sequence: 2, Kind: "move", Payload: json.RawMessage(`{"x":2}`)}
	if _, err := source.SubmitNetwork(applied); err != nil {
		t.Fatal(err)
	}
	if err := source.Step(); err != nil {
		t.Fatal(err)
	}
	if _, err := source.SubmitNetwork(pending); err != nil {
		t.Fatal(err)
	}
	recovery, err := source.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if recovery.State.Tick != 1 || len(recovery.Pending) != 1 || len(recovery.Accepted) != 1 || recovery.Sequences["player"].Ack != 1 {
		t.Fatalf("recovery checkpoint = %#v", recovery)
	}

	restored := build(identity)
	defer restored.Close()
	if err := restored.RestoreRecoveryCheckpoint(recovery); err != nil {
		t.Fatal(err)
	}
	if restored.Status().Tick != 1 || restored.ProcessedSequence("player") != 1 {
		t.Fatalf("restored status=%#v sequence=%d", restored.Status(), restored.ProcessedSequence("player"))
	}
	if accepted, found := restored.AcceptedNetworkCommand("player", 1); !found || string(accepted.Payload) != string(applied.Payload) {
		t.Fatalf("restored accepted command=%#v found=%v", accepted, found)
	}
	secondGeneration, err := restored.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	if len(secondGeneration.Accepted) != 1 || string(secondGeneration.Accepted[0].Payload) != string(applied.Payload) {
		t.Fatalf("second-generation accepted commands = %#v", secondGeneration.Accepted)
	}
	if len(secondGeneration.Pending) != 1 || string(secondGeneration.Pending[0].Payload) != string(pending.Payload) {
		t.Fatalf("second-generation pending commands = %#v", secondGeneration.Pending)
	}
	if err := ValidateRecoveryCheckpoint(secondGeneration); err != nil {
		t.Fatalf("second-generation recovery checkpoint: %v", err)
	}
	if _, err := restored.SubmitNetwork(applied); !errors.Is(err, ErrCommandSequence) {
		t.Fatalf("restored duplicate error = %v", err)
	}
	if err := restored.Step(); err != nil {
		t.Fatal(err)
	}
	if err := restored.Step(); err != nil {
		t.Fatal(err)
	}
	if restored.ProcessedSequence("player") != 2 {
		t.Fatalf("pending command sequence = %d", restored.ProcessedSequence("player"))
	}

	tampered := recovery
	tampered.Pending[0].Payload = json.RawMessage(`{"x":99}`)
	if err := ValidateRecoveryCheckpoint(tampered); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("tampered recovery error = %v", err)
	}
}

func TestRecoveryCheckpointRejectsMismatchedComposedRuntime(t *testing.T) {
	engine := gameecs.New()
	source, err := New(engine, Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	stores := simulation.NewStateStore()
	if err := source.RegisterAuthoritativeRuntime(compatibilityIdentity("source"), stores); err != nil {
		t.Fatal(err)
	}
	recovery, err := source.RecoveryCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	targetEngine := gameecs.New()
	target, err := New(targetEngine, Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	if err := target.RegisterAuthoritativeRuntime(compatibilityIdentity("other"), simulation.NewStateStore()); err != nil {
		t.Fatal(err)
	}
	if err := target.RestoreRecoveryCheckpoint(recovery); err == nil {
		t.Fatal("mismatched runtime recovery succeeded")
	}
	if target.Status().Tick != 0 {
		t.Fatalf("failed restore mutated tick to %d", target.Status().Tick)
	}
}
