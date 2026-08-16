package realm

import (
	"errors"
	"testing"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

func TestMemoryCheckpointsValidateIdentityMonotonicityAndCopies(t *testing.T) {
	identity := orchestrationIdentity()
	digest, err := identity.Digest()
	if err != nil {
		t.Fatal(err)
	}
	initial, err := testCheckpoint(identity, 0)
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewGameCheckpoint("game", "allocation", digest, initial)
	if err != nil {
		t.Fatalf("tick-zero checkpoint should be valid: %v", err)
	}
	store := NewMemoryCheckpoints()
	saved, err := store.Save(t.Context(), record)
	if err != nil {
		t.Fatal(err)
	}
	saved.Checkpoint.State.Participants[0].Data[0] ^= 0xff
	latest, err := store.Latest(t.Context(), "game")
	if err != nil || latest.Checksum != record.Checksum {
		t.Fatalf("latest=%#v error=%v", latest, err)
	}
	if _, err := store.Save(t.Context(), record); err != nil {
		t.Fatalf("exact same-tick retry should be idempotent: %v", err)
	}

	newerCheckpoint, err := testCheckpoint(identity, 2)
	if err != nil {
		t.Fatal(err)
	}
	newer, err := NewGameCheckpoint("game", "allocation", digest, newerCheckpoint)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(t.Context(), newer); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(t.Context(), record); !errors.Is(err, ErrGameCheckpoint) {
		t.Fatalf("tick regression error = %v", err)
	}
	otherAllocation := newer
	otherAllocation.AllocationID = "replacement"
	if _, err := store.Save(t.Context(), otherAllocation); !errors.Is(err, ErrGameCheckpoint) {
		t.Fatalf("allocation replacement error = %v", err)
	}
	conflictCheckpoint, err := testCheckpoint(identity, 2)
	if err != nil {
		t.Fatal(err)
	}
	conflictCheckpoint.State.Snapshot.Entities = []uint64{1}
	conflictCheckpoint.State.Checksum, err = simulation.CompositeChecksum(*conflictCheckpoint.State.Snapshot,
		conflictCheckpoint.State.Participants)
	if err != nil {
		t.Fatal(err)
	}
	conflictCheckpoint, err = gamesession.NewRecoveryCheckpoint(conflictCheckpoint.State, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	conflict, err := NewGameCheckpoint("game", "allocation", digest, conflictCheckpoint)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(t.Context(), conflict); !errors.Is(err, ErrGameCheckpoint) {
		t.Fatalf("same-tick conflict error = %v", err)
	}
}

func TestGameCheckpointRejectsTamperingAndWrongRuntime(t *testing.T) {
	identity := orchestrationIdentity()
	digest, _ := identity.Digest()
	checkpoint, err := testCheckpoint(identity, 7)
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*GameCheckpoint){
		"checksum": func(record *GameCheckpoint) { record.Checksum = "tampered" },
		"snapshot": func(record *GameCheckpoint) { record.Checkpoint.State.Snapshot.Entities = []uint64{99} },
		"identity": func(record *GameCheckpoint) { record.IdentityHash = "wrong-runtime" },
		"tick":     func(record *GameCheckpoint) { record.Tick++ },
	} {
		t.Run(name, func(t *testing.T) {
			record, createErr := NewGameCheckpoint("game", "allocation", digest, checkpoint)
			if createErr != nil {
				t.Fatal(createErr)
			}
			mutate(&record)
			if _, _, validateErr := validateGameCheckpoint(record); !errors.Is(validateErr, ErrGameCheckpoint) {
				t.Fatalf("tamper error = %v", validateErr)
			}
		})
	}
}
