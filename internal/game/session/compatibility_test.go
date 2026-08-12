package session

import (
	"errors"
	"testing"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func compatibilityIdentity(packageHash string) simulation.RuntimeIdentity {
	return simulation.RuntimeIdentity{ModID: "d2legacy", ContractVersion: "v1",
		PackageHash: packageHash, AuthoritativeHash: packageHash, ConfigurationHash: "config",
		Dependencies: map[string]string{"engine": "hash"}, CapabilityVersions: map[string]string{"engine.ecs": "v1"}}
}

func identityState(t *testing.T, identity simulation.RuntimeIdentity) []simulation.ParticipantState {
	t.Helper()
	participant, err := simulation.NewIdentityParticipant(identity)
	if err != nil {
		t.Fatal(err)
	}
	data, err := participant.SnapshotState()
	if err != nil {
		t.Fatal(err)
	}
	return []simulation.ParticipantState{{ID: participant.StateID(), Data: data}}
}

func TestHeadlessRealmAndGameServerPinSessionCompatibility(t *testing.T) {
	server := compatibilityIdentity("package-a")
	allocation, err := Allocate("realm-session-7", server, PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	token, err := allocation.Admit("character-alice", server)
	if err != nil {
		t.Fatal(err)
	}
	if token.Prediction != PredictionLimited {
		t.Fatalf("prediction tier = %q", token.Prediction)
	}
	if err := allocation.ValidateReconnect(token, server); err != nil {
		t.Fatal(err)
	}
	if err := allocation.ValidateDurable(allocation.Durable("character-alice")); err != nil {
		t.Fatal(err)
	}

	mismatch := compatibilityIdentity("package-b")
	if _, err := allocation.Admit("character-bob", mismatch); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched client error = %v", err)
	}
	if err := allocation.ValidateReconnect(token, mismatch); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched reconnect error = %v", err)
	}
	badDurable := allocation.Durable("character-alice")
	badDurable.IdentityHash = "different"
	if err := allocation.ValidateDurable(badDurable); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched durable error = %v", err)
	}

	checkpoint := simulation.Checkpoint{Tick: 1, Participants: identityState(t, server)}
	replay := simulation.Replay{Version: simulation.ReplayVersion, InitialParticipants: identityState(t, server),
		Checkpoints: []simulation.Checkpoint{checkpoint}}
	if err := allocation.ValidateCheckpoint(checkpoint, nil); err != nil {
		t.Fatal(err)
	}
	if err := allocation.ValidateReplay(replay, nil); err != nil {
		t.Fatal(err)
	}
	badCheckpoint := checkpoint
	badCheckpoint.Participants = identityState(t, mismatch)
	if err := allocation.ValidateCheckpoint(badCheckpoint, nil); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched checkpoint error = %v", err)
	}
	badReplay := replay
	badReplay.InitialParticipants = identityState(t, mismatch)
	if err := allocation.ValidateReplay(badReplay, nil); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("mismatched replay error = %v", err)
	}
}

func TestChangedPackageAppliesOnlyToNewAllocationWithoutExplicitMigration(t *testing.T) {
	oldIdentity, newIdentity := compatibilityIdentity("package-old"), compatibilityIdentity("package-new")
	oldAllocation, _ := Allocate("old-session", oldIdentity, PredictionNone)
	newAllocation, _ := Allocate("new-session", newIdentity, PredictionSharedRules)
	oldCheckpoint := simulation.Checkpoint{Tick: 4, Participants: identityState(t, oldIdentity)}
	if err := newAllocation.ValidateCheckpoint(oldCheckpoint, nil); !errors.Is(err, ErrCompatibility) {
		t.Fatalf("implicit migration error = %v", err)
	}
	oldHash, _ := oldIdentity.Digest()
	newHash, _ := newIdentity.Digest()
	migration := &StateMigration{FromIdentityHash: oldHash, ToIdentityHash: newHash}
	if err := newAllocation.ValidateCheckpoint(oldCheckpoint, migration); err != nil {
		t.Fatal(err)
	}
	if err := oldAllocation.ValidateCheckpoint(oldCheckpoint, nil); err != nil {
		t.Fatal(err)
	}
}
