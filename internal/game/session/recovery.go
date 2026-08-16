package session

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const RecoveryCheckpointVersion = "GameSessionRecovery/v1"

// SequenceCheckpoint retains duplicate-suppression state across authority
// replacement. Beyond contains processed out-of-order sequences above Ack.
type SequenceCheckpoint struct {
	Ack    uint64   `json:"ack"`
	Beyond []uint64 `json:"beyond,omitempty"`
}

// RecoveryCheckpoint is the complete restart boundary for one authoritative
// session. State alone is insufficient: accepted future inputs and sequence
// acknowledgements must survive or a restored worker could lose or reapply an
// input whose network acknowledgement raced the crash.
type RecoveryCheckpoint struct {
	Version   string                        `json:"version"`
	State     simulation.Checkpoint         `json:"state"`
	Pending   []simulation.Command          `json:"pending,omitempty"`
	Accepted  []simulation.Command          `json:"accepted,omitempty"`
	Sequences map[string]SequenceCheckpoint `json:"sequences,omitempty"`
	Checksum  string                        `json:"checksum"`
}

type recoveryChecksumPayload struct {
	Version   string                        `json:"version"`
	State     simulation.Checkpoint         `json:"state"`
	Pending   []simulation.Command          `json:"pending,omitempty"`
	Accepted  []simulation.Command          `json:"accepted,omitempty"`
	Sequences map[string]SequenceCheckpoint `json:"sequences,omitempty"`
}

// NewRecoveryCheckpoint seals an externally composed recovery boundary. It is
// primarily used by trusted persistence and compatibility fixtures; ordinary
// authorities should capture through Session.RecoveryCheckpoint.
func NewRecoveryCheckpoint(state simulation.Checkpoint, pending, accepted []simulation.Command,
	sequences map[string]SequenceCheckpoint) (RecoveryCheckpoint, error) {
	recovery := RecoveryCheckpoint{Version: RecoveryCheckpointVersion, State: state,
		Pending: append([]simulation.Command(nil), pending...), Accepted: append([]simulation.Command(nil), accepted...),
		Sequences: make(map[string]SequenceCheckpoint, len(sequences))}
	for player, sequence := range sequences {
		recovery.Sequences[player] = SequenceCheckpoint{Ack: sequence.Ack, Beyond: append([]uint64(nil), sequence.Beyond...)}
	}
	canonicalizeRecoveryCommands(recovery.Pending)
	canonicalizeRecoveryCommands(recovery.Accepted)
	checksum, err := recoveryChecksum(recovery)
	if err != nil {
		return RecoveryCheckpoint{}, err
	}
	recovery.Checksum = checksum
	if err := ValidateRecoveryCheckpoint(recovery); err != nil {
		return RecoveryCheckpoint{}, err
	}
	return cloneRecoveryCheckpoint(recovery)
}

// RecoveryCheckpoint captures canonical simulation state and network admission
// bookkeeping under the same session lock.
func (session *Session) RecoveryCheckpoint() (RecoveryCheckpoint, error) {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return RecoveryCheckpoint{}, ErrClosed
	}
	state, err := session.canonicalCheckpointLocked()
	if err != nil {
		return RecoveryCheckpoint{}, err
	}
	recovery := RecoveryCheckpoint{Version: RecoveryCheckpointVersion, State: state,
		Sequences: make(map[string]SequenceCheckpoint, len(session.processed))}
	for _, commands := range session.pending {
		recovery.Pending = append(recovery.Pending, commands...)
	}
	accepted := make(map[string]simulation.Command)
	for key, command := range session.recoveryAccepted {
		if state.Tick-command.Tick <= session.config.RollbackWindow {
			accepted[key] = command
		}
	}
	for _, command := range session.commands {
		if state.Tick-command.Tick <= session.config.RollbackWindow {
			accepted[recoveryCommandKey(command)] = command
		}
	}
	for _, command := range accepted {
		recovery.Accepted = append(recovery.Accepted, command)
	}
	for player, progress := range session.processed {
		sequence := SequenceCheckpoint{Ack: progress.ack}
		for value := range progress.beyond {
			sequence.Beyond = append(sequence.Beyond, value)
		}
		sort.Slice(sequence.Beyond, func(i, j int) bool { return sequence.Beyond[i] < sequence.Beyond[j] })
		recovery.Sequences[player] = sequence
	}
	canonicalizeRecoveryCommands(recovery.Pending)
	canonicalizeRecoveryCommands(recovery.Accepted)
	checksum, err := recoveryChecksum(recovery)
	if err != nil {
		return RecoveryCheckpoint{}, err
	}
	recovery.Checksum = checksum
	return cloneRecoveryCheckpoint(recovery)
}

// ValidateRecoveryCheckpoint verifies both simulation state and the session
// admission metadata that prevents input loss or duplicate application.
func ValidateRecoveryCheckpoint(recovery RecoveryCheckpoint) error {
	if recovery.Version != RecoveryCheckpointVersion || recovery.State.Snapshot == nil ||
		recovery.State.Tick != recovery.State.Snapshot.Tick || recovery.State.Checksum == "" || recovery.Checksum == "" {
		return ErrCompatibility
	}
	stateChecksum, err := simulation.CompositeChecksum(*recovery.State.Snapshot, recovery.State.Participants)
	if err != nil || stateChecksum != recovery.State.Checksum {
		return ErrCompatibility
	}
	processed := make(map[string]SequenceCheckpoint, len(recovery.Sequences))
	for player, sequence := range recovery.Sequences {
		if strings.TrimSpace(player) == "" {
			return ErrCompatibility
		}
		previous := sequence.Ack
		for _, value := range sequence.Beyond {
			if value <= sequence.Ack || value <= previous {
				return ErrCompatibility
			}
			previous = value
		}
		processed[player] = sequence
	}
	seen := make(map[string]struct{}, len(recovery.Pending)+len(recovery.Accepted))
	for _, command := range recovery.Accepted {
		if !validRecoveryCommand(command) || command.Tick > recovery.State.Tick {
			return ErrCompatibility
		}
		if !recoverySequenceContains(processed[command.Player], command.Sequence) {
			return ErrCompatibility
		}
		key := recoveryCommandKey(command)
		if _, duplicate := seen[key]; duplicate {
			return ErrCompatibility
		}
		seen[key] = struct{}{}
	}
	for _, command := range recovery.Pending {
		if !validRecoveryCommand(command) || command.Tick <= recovery.State.Tick {
			return ErrCompatibility
		}
		if recoverySequenceContains(processed[command.Player], command.Sequence) {
			return ErrCompatibility
		}
		key := recoveryCommandKey(command)
		if _, duplicate := seen[key]; duplicate {
			return ErrCompatibility
		}
		seen[key] = struct{}{}
	}
	want, err := recoveryChecksum(recovery)
	if err != nil || want != recovery.Checksum {
		return ErrCompatibility
	}
	return nil
}

// RestoreRecoveryCheckpoint replaces the newly composed runtime's state before
// it starts ticking. Systems and participants remain registered from the exact
// validated package identity; replay history restarts at this recovery boundary.
func (session *Session) RestoreRecoveryCheckpoint(recovery RecoveryCheckpoint) error {
	if err := ValidateRecoveryCheckpoint(recovery); err != nil {
		return err
	}
	initial, err := cloneSnapshot(*recovery.State.Snapshot)
	if err != nil {
		return err
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return ErrClosed
	}
	if session.engine.Tick() != session.initial.Tick || len(session.commands) != 0 || len(session.pending) != 0 {
		return fmt.Errorf("%w: recovery must occur before commands or ticks", ErrCompatibility)
	}
	beforeECS, err := session.engine.Snapshot()
	if err != nil {
		return err
	}
	beforeParticipants, err := session.snapshotParticipantsLocked()
	if err != nil {
		return err
	}
	if err := session.engine.Restore(*recovery.State.Snapshot); err != nil {
		return err
	}
	if err := session.restoreParticipantsLocked(recovery.State.Participants); err != nil {
		return errors.Join(err, session.engine.Restore(beforeECS), session.restoreParticipantsLocked(beforeParticipants))
	}
	session.initial = initial
	session.initialParticipants = cloneParticipantStates(recovery.State.Participants)
	session.pending = make(map[uint64][]simulation.Command)
	for _, command := range recovery.Pending {
		session.pending[command.Tick] = append(session.pending[command.Tick], command)
	}
	session.commands = nil
	session.checkpoints = nil
	session.history = []rollbackFrame{{tick: initial.Tick, ecs: initial, participants: cloneParticipantStates(recovery.State.Participants)}}
	session.processed = make(map[string]sequenceProgress, len(recovery.Sequences))
	for player, sequence := range recovery.Sequences {
		progress := sequenceProgress{ack: sequence.Ack, beyond: make(map[uint64]struct{}, len(sequence.Beyond))}
		for _, value := range sequence.Beyond {
			progress.beyond[value] = struct{}{}
		}
		session.processed[player] = progress
	}
	session.recoveryAccepted = make(map[string]simulation.Command, len(recovery.Accepted))
	for _, command := range recovery.Accepted {
		session.recoveryAccepted[recoveryCommandKey(command)] = command
	}
	session.lag = 0
	return nil
}

func recoveryChecksum(recovery RecoveryCheckpoint) (string, error) {
	payload := recoveryChecksumPayload{Version: recovery.Version, State: recovery.State,
		Pending: recovery.Pending, Accepted: recovery.Accepted, Sequences: recovery.Sequences}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func cloneRecoveryCheckpoint(recovery RecoveryCheckpoint) (RecoveryCheckpoint, error) {
	encoded, err := json.Marshal(recovery)
	if err != nil {
		return RecoveryCheckpoint{}, err
	}
	var clone RecoveryCheckpoint
	if err := json.Unmarshal(encoded, &clone); err != nil {
		return RecoveryCheckpoint{}, err
	}
	return clone, nil
}

func canonicalizeRecoveryCommands(commands []simulation.Command) {
	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Tick != commands[j].Tick {
			return commands[i].Tick < commands[j].Tick
		}
		if commands[i].Player != commands[j].Player {
			return commands[i].Player < commands[j].Player
		}
		return commands[i].Sequence < commands[j].Sequence
	})
}

func validRecoveryCommand(command simulation.Command) bool {
	return command.Tick > 0 && command.Sequence > 0 && strings.TrimSpace(command.Player) != "" && strings.TrimSpace(command.Kind) != ""
}

func recoveryCommandKey(command simulation.Command) string {
	return fmt.Sprintf("%s\x00%d", command.Player, command.Sequence)
}

func recoverySequenceContains(sequence SequenceCheckpoint, value uint64) bool {
	if value > 0 && value <= sequence.Ack {
		return true
	}
	index := sort.Search(len(sequence.Beyond), func(index int) bool { return sequence.Beyond[index] >= value })
	return index < len(sequence.Beyond) && sequence.Beyond[index] == value
}
