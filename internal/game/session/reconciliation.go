package session

import (
	"errors"
	"fmt"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// CanonicalCheckpoint snapshots the current server outcome without adding a
// replay checkpoint. The checksum includes opaque authoritative participants,
// so transport code cannot substitute an ECS-only client claim.
func (session *Session) CanonicalCheckpoint() (simulation.Checkpoint, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return simulation.Checkpoint{}, ErrClosed
	}
	return session.canonicalCheckpointLocked()
}

func (session *Session) canonicalCheckpointLocked() (simulation.Checkpoint, error) {
	snapshot, err := session.engine.Snapshot()
	if err != nil {
		return simulation.Checkpoint{}, err
	}
	participants, err := session.snapshotParticipantsLocked()
	if err != nil {
		return simulation.Checkpoint{}, err
	}
	checksum, err := simulation.CompositeChecksum(snapshot, participants)
	if err != nil {
		return simulation.Checkpoint{}, err
	}
	return simulation.Checkpoint{Tick: snapshot.Tick, Checksum: checksum, Snapshot: &snapshot, Participants: participants}, nil
}

// Reconciliation is a server-canonical correction for an optional predicted
// client world. Snapshot never contains client-authored state.
type Reconciliation struct {
	Tick       uint64
	Checksum   string
	Corrected  bool
	Difference string
	Snapshot   gameecs.Snapshot
}

// ReconcilePrediction compares an untrusted prediction with a verified server
// checkpoint. Nil prediction represents clients that disable prediction.
func ReconcilePrediction(tier PredictionTier, predicted *gameecs.Snapshot, authoritative simulation.Checkpoint) (Reconciliation, error) {
	if err := ValidatePredictionTier(tier); err != nil {
		return Reconciliation{}, err
	}
	if authoritative.Snapshot == nil {
		return Reconciliation{}, fmt.Errorf("%w: canonical snapshot is required", ErrCompatibility)
	}
	checksum, err := simulation.CompositeChecksum(*authoritative.Snapshot, authoritative.Participants)
	if err != nil {
		return Reconciliation{}, err
	}
	if checksum != authoritative.Checksum || authoritative.Tick != authoritative.Snapshot.Tick {
		return Reconciliation{}, fmt.Errorf("%w: canonical checkpoint integrity differs", ErrCompatibility)
	}
	snapshot, err := cloneSnapshot(*authoritative.Snapshot)
	if err != nil {
		return Reconciliation{}, err
	}
	result := Reconciliation{Tick: authoritative.Tick, Checksum: authoritative.Checksum, Snapshot: snapshot}
	if predicted == nil {
		result.Corrected = true
		result.Difference = "prediction disabled"
		return result, nil
	}
	result.Difference = gameecs.FirstDifference(snapshot, *predicted)
	result.Corrected = result.Difference != ""
	return result, nil
}

// Apply replaces a client-side predicted ECS with the server outcome.
func (reconciliation Reconciliation) Apply(predicted *gameecs.Engine) error {
	if predicted == nil {
		return errors.New("game session: predicted engine is required")
	}
	if reconciliation.Snapshot.Version == 0 {
		return errors.New("game session: canonical reconciliation snapshot is required")
	}
	return predicted.Restore(reconciliation.Snapshot)
}
