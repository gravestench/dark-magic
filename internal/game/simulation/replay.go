package simulation

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

// ReplayVersion identifies the composite ECS-plus-participant wire format.
const ReplayVersion = 2

// ErrReplay identifies structurally invalid or incompatible replay input.
var ErrReplay = errors.New("simulation: invalid replay")

// Replay is the portable deterministic evidence for one completed session span.
// Initial state makes the command log self-contained; checkpoints make drift
// detectable before the end rather than serving as mutable save points.
type Replay struct {
	Version             uint32             `json:"version"`
	StepNanos           int64              `json:"step_nanos"`
	Initial             gameecs.Snapshot   `json:"initial"`
	InitialParticipants []ParticipantState `json:"initial_participants,omitempty"`
	Commands            []Command          `json:"commands"`
	Checkpoints         []Checkpoint       `json:"checkpoints"`
}

// Checkpoint records the expected composite state after one completed tick.
// Embedded snapshots are retained for diagnostics; Checksum is authoritative.
type Checkpoint struct {
	Tick         uint64             `json:"tick"`
	Checksum     string             `json:"checksum"`
	Snapshot     *gameecs.Snapshot  `json:"snapshot,omitempty"`
	Participants []ParticipantState `json:"participants,omitempty"`
}

// StateParticipant contributes deterministic authoritative state that is not
// naturally stored in the ECS. IDs are stable replay-format identities.
type StateParticipant interface {
	StateID() string
	SnapshotState() ([]byte, error)
	RestoreState([]byte) error
}

// ParticipantState is one opaque, stable-ID non-ECS state contribution.
type ParticipantState struct {
	ID   string `json:"id"`
	Data []byte `json:"data"`
}

// ReplayPrepare reinstalls deterministic systems that snapshots intentionally
// omit before commands and ticks are evaluated.
type ReplayPrepare func(*gameecs.Engine) error

// ReplayApply applies one already-recorded command through trusted semantics.
type ReplayApply func(*gameecs.Engine, Command) error

// DesyncError identifies the first checkpoint whose reconstructed state differs.
type DesyncError struct {
	Tick     uint64
	Expected string
	Actual   string
	Detail   string
}

func (err *DesyncError) Error() string {
	if err.Detail == "" {
		return fmt.Sprintf("simulation: replay desync at tick %d: expected %s, got %s", err.Tick, err.Expected, err.Actual)
	}
	return fmt.Sprintf("simulation: replay desync at tick %d: %s (expected %s, got %s)", err.Tick, err.Detail, err.Expected, err.Actual)
}

// VerifyReplay restores the initial world, applies commands before their target
// ticks, advances the fixed simulation clock, and verifies every checkpoint.
func VerifyReplay(replay Replay, prepare ReplayPrepare, apply ReplayApply, participants ...StateParticipant) error {
	if replay.Version != ReplayVersion || replay.StepNanos <= 0 || apply == nil {
		return ErrReplay
	}
	engine, err := gameecs.RestoreSnapshot(replay.Initial)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrReplay, err)
	}
	defer engine.Close()
	participantByID, err := indexParticipants(participants)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrReplay, err)
	}
	if err := restoreParticipants(replay.InitialParticipants, participantByID); err != nil {
		return fmt.Errorf("%w: restore initial participants: %v", ErrReplay, err)
	}
	if prepare != nil {
		if err := prepare(engine); err != nil {
			return fmt.Errorf("%w: prepare: %v", ErrReplay, err)
		}
	}
	commandIndex := 0
	currentTick := replay.Initial.Tick
	for _, checkpoint := range replay.Checkpoints {
		if checkpoint.Tick <= currentTick {
			return fmt.Errorf("%w: checkpoint tick %d follows %d", ErrReplay, checkpoint.Tick, currentTick)
		}
		for currentTick < checkpoint.Tick {
			nextTick := currentTick + 1
			for commandIndex < len(replay.Commands) && replay.Commands[commandIndex].Tick == nextTick {
				if err := apply(engine, replay.Commands[commandIndex]); err != nil {
					return fmt.Errorf("%w: apply command %d at tick %d: %v", ErrReplay, commandIndex, nextTick, err)
				}
				commandIndex++
			}
			if commandIndex < len(replay.Commands) && replay.Commands[commandIndex].Tick < nextTick {
				return fmt.Errorf("%w: commands are not ordered by tick", ErrReplay)
			}
			if err := engine.Update(time.Duration(replay.StepNanos)); err != nil {
				return fmt.Errorf("%w: update tick %d: %v", ErrReplay, nextTick, err)
			}
			currentTick = nextTick
		}
		snapshot, err := engine.Snapshot()
		if err != nil {
			return err
		}
		participantStates, err := snapshotParticipants(participants)
		if err != nil {
			return err
		}
		checksum, err := CompositeChecksum(snapshot, participantStates)
		if err != nil {
			return err
		}
		if checksum != checkpoint.Checksum {
			detail := ""
			if checkpoint.Snapshot != nil {
				detail = gameecs.FirstDifference(*checkpoint.Snapshot, snapshot)
			}
			if detail == "" {
				detail = firstParticipantDifference(checkpoint.Participants, participantStates)
			}
			return &DesyncError{Tick: checkpoint.Tick, Expected: checkpoint.Checksum, Actual: checksum, Detail: detail}
		}
	}
	if commandIndex != len(replay.Commands) {
		return fmt.Errorf("%w: %d commands occur after final checkpoint", ErrReplay, len(replay.Commands)-commandIndex)
	}
	return nil
}

// CompositeChecksum covers the ECS and every non-ECS participant with explicit
// length framing, so concatenation cannot create ambiguous state identities.
func CompositeChecksum(snapshot gameecs.Snapshot, participants []ParticipantState) (string, error) {
	ecs, err := snapshot.Marshal()
	if err != nil {
		return "", err
	}
	ordered := cloneParticipantStates(participants)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	hash := sha256.New()
	writeFramed(hash, []byte("ecs"))
	writeFramed(hash, ecs)
	for _, participant := range ordered {
		if participant.ID == "" {
			return "", fmt.Errorf("simulation: participant ID is required")
		}
		writeFramed(hash, []byte(participant.ID))
		writeFramed(hash, participant.Data)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type byteWriter interface{ Write([]byte) (int, error) }

func writeFramed(destination byteWriter, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = destination.Write(length[:])
	_, _ = destination.Write(value)
}

func indexParticipants(participants []StateParticipant) (map[string]StateParticipant, error) {
	result := make(map[string]StateParticipant, len(participants))
	for _, participant := range participants {
		if participant == nil || participant.StateID() == "" {
			return nil, fmt.Errorf("simulation: participant and ID are required")
		}
		if _, exists := result[participant.StateID()]; exists {
			return nil, fmt.Errorf("simulation: duplicate participant %q", participant.StateID())
		}
		result[participant.StateID()] = participant
	}
	return result, nil
}

func snapshotParticipants(participants []StateParticipant) ([]ParticipantState, error) {
	result := make([]ParticipantState, 0, len(participants))
	for _, participant := range participants {
		data, err := participant.SnapshotState()
		if err != nil {
			return nil, fmt.Errorf("simulation: snapshot participant %q: %w", participant.StateID(), err)
		}
		result = append(result, ParticipantState{ID: participant.StateID(), Data: append([]byte(nil), data...)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func restoreParticipants(states []ParticipantState, participants map[string]StateParticipant) error {
	if len(states) != len(participants) {
		return fmt.Errorf("participant count is %d, want %d", len(participants), len(states))
	}
	for _, state := range states {
		participant, found := participants[state.ID]
		if !found {
			return fmt.Errorf("missing participant %q", state.ID)
		}
		if err := participant.RestoreState(append([]byte(nil), state.Data...)); err != nil {
			return fmt.Errorf("participant %q: %w", state.ID, err)
		}
	}
	return nil
}

func cloneParticipantStates(states []ParticipantState) []ParticipantState {
	result := make([]ParticipantState, len(states))
	for index, state := range states {
		result[index] = ParticipantState{ID: state.ID, Data: append([]byte(nil), state.Data...)}
	}
	return result
}

func firstParticipantDifference(expected, actual []ParticipantState) string {
	left, right := cloneParticipantStates(expected), cloneParticipantStates(actual)
	sort.Slice(left, func(i, j int) bool { return left[i].ID < left[j].ID })
	sort.Slice(right, func(i, j int) bool { return right[i].ID < right[j].ID })
	if len(left) != len(right) {
		return fmt.Sprintf("participant count differs: expected %d, got %d", len(left), len(right))
	}
	for index := range left {
		if left[index].ID != right[index].ID {
			return fmt.Sprintf("participant ID differs: expected %q, got %q", left[index].ID, right[index].ID)
		}
		if !equalBytes(left[index].Data, right[index].Data) {
			return fmt.Sprintf("participant %q state differs", left[index].ID)
		}
	}
	return ""
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
