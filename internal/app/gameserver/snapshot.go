package gameserver

import (
	"encoding/json"
	"errors"
	"fmt"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// snapshot projects one player from a canonical checkpoint and defensively owns the returned JSON payload.
func (endpoint *Endpoint) snapshot(playerID string) (Snapshot, error) {
	checkpoint, err := endpoint.canonicalCheckpoint()
	if err != nil {
		return Snapshot{}, err
	}

	payload, err := endpoint.project(playerID, checkpoint)
	if err != nil {
		return Snapshot{}, fmt.Errorf("game server protocol: project snapshot: %w", err)
	}

	if !json.Valid(payload) {
		return Snapshot{}, errors.New("game server protocol: projector returned invalid JSON")
	}

	return Snapshot{
		Version:           SessionProtocolVersion,
		Tick:              checkpoint.Tick,
		Checksum:          checkpoint.Checksum,
		StepNanos:         int64(endpoint.host.Session.StepDuration()),
		AcknowledgedInput: endpoint.host.Session.ProcessedSequence(playerID),
		Payload:           append(json.RawMessage(nil), payload...),
	}, nil
}

// canonicalCheckpoint captures at most once per completed authoritative tick.
// Every watcher projects from the same immutable checkpoint instead of independently snapshotting a live ECS.
func (endpoint *Endpoint) canonicalCheckpoint() (simulation.Checkpoint, error) {
	endpoint.snapshotMu.Lock()
	defer endpoint.snapshotMu.Unlock()

	current := endpoint.host.Session.Status().Tick
	if endpoint.checkpoint.Snapshot != nil && endpoint.checkpoint.Tick == current {
		return cloneCheckpoint(endpoint.checkpoint), nil
	}

	checkpoint, err := endpoint.host.Session.CanonicalCheckpoint()
	if err != nil {
		return simulation.Checkpoint{}, err
	}

	endpoint.checkpoint = cloneCheckpoint(checkpoint)

	return cloneCheckpoint(checkpoint), nil
}

// cloneCheckpoint owns slices and the ECS snapshot shell so concurrent projectors cannot mutate cached state.
func cloneCheckpoint(checkpoint simulation.Checkpoint) simulation.Checkpoint {
	copy := checkpoint
	if checkpoint.Snapshot != nil {
		snapshot := *checkpoint.Snapshot
		snapshot.Entities = append([]uint64(nil), checkpoint.Snapshot.Entities...)
		snapshot.Components = append([]gameecs.ComponentSnapshot(nil), checkpoint.Snapshot.Components...)
		copy.Snapshot = &snapshot
	}

	copy.Participants = append([]simulation.ParticipantState(nil), checkpoint.Participants...)

	return copy
}
