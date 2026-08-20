package session

import (
	"errors"
	"fmt"
	"sort"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// rollbackFrame captures the complete authoritative state immediately before a tick begins.
type rollbackFrame struct {
	tick         uint64
	ecs          gameecs.Snapshot
	participants []simulation.ParticipantState
}

// tickTransaction records the in-memory boundaries needed to undo one failed tick without copying unrelated state.
type tickTransaction struct {
	ecs          gameecs.Snapshot
	participants []simulation.ParticipantState
	commandCount int
	historyCount int
	processed    map[string]sequenceProgress
}

// rollbackTransaction owns a full pre-replay copy because late input can rewrite several completed ticks.
type rollbackTransaction struct {
	ecs          gameecs.Snapshot
	participants []simulation.ParticipantState
	pending      map[uint64][]simulation.Command
	commands     []simulation.Command
	checkpoints  []simulation.Checkpoint
	history      []rollbackFrame
	processed    map[string]sequenceProgress
}

// Step executes exactly one authoritative simulation tick while excluding concurrent admission and observation.
func (session *Session) Step() error {
	session.mu.Lock()
	defer session.mu.Unlock()

	return session.stepLocked()
}

// stepLocked applies one tick as a transaction; callers must hold the session lock across the entire operation.
func (session *Session) stepLocked() error {
	if session.closed {
		return ErrClosed
	}

	transaction, err := session.captureTickTransactionLocked()
	if err != nil {
		return err
	}

	// History is captured before command handlers run so late network input can restore the exact prior boundary.
	session.recordRollbackFrameLocked(transaction)

	tick := session.engine.Tick() + 1
	commands := session.pending[tick]
	canonicalizeTickCommands(commands)

	if err := session.applyTickCommandsLocked(commands); err != nil {
		return session.rollbackTickLocked(transaction, err)
	}

	delete(session.pending, tick)

	if err := session.engine.Update(session.config.Step); err != nil {
		// Retain failed-tick commands for diagnosis or retry after the faulty runtime is replaced.
		session.pending[tick] = commands

		return session.rollbackTickLocked(transaction, err)
	}

	session.expireAcceptedCommandsLocked(session.engine.Tick())

	if tick%session.config.CheckpointInterval == 0 {
		// A checkpoint failure occurs after a successful tick and therefore does not roll authoritative state back.
		return session.checkpointLocked()
	}

	return nil
}

// captureTickTransactionLocked snapshots every store a command handler or system may mutate during one tick.
func (session *Session) captureTickTransactionLocked() (tickTransaction, error) {
	ecs, err := session.engine.Snapshot()
	if err != nil {
		return tickTransaction{}, err
	}

	participants, err := session.snapshotParticipantsLocked()
	if err != nil {
		return tickTransaction{}, err
	}

	return tickTransaction{
		ecs:          ecs,
		participants: participants,
		commandCount: len(session.commands),
		historyCount: len(session.history),
		processed:    cloneProcessed(session.processed),
	}, nil
}

// recordRollbackFrameLocked retains only the configured replay window plus the current pre-tick boundary.
func (session *Session) recordRollbackFrameLocked(transaction tickTransaction) {
	frame := rollbackFrame{
		tick:         transaction.ecs.Tick,
		ecs:          transaction.ecs,
		participants: cloneParticipantStates(transaction.participants),
	}
	session.history = append(session.history, frame)

	limit := int(session.config.RollbackWindow) + 1
	if len(session.history) > limit {
		// Copy the retained tail so discarded frames no longer keep their backing array alive.
		session.history = append([]rollbackFrame(nil), session.history[len(session.history)-limit:]...)
	}
}

// rollbackTickLocked restores the pre-tick stores and bookkeeping while preserving the original failure as the cause.
func (session *Session) rollbackTickLocked(transaction tickTransaction, cause error) error {
	if restoreErr := session.engine.Restore(transaction.ecs); restoreErr != nil {
		return errors.Join(cause, fmt.Errorf("game session: roll back ECS: %w", restoreErr))
	}

	if restoreErr := session.restoreParticipantsLocked(transaction.participants); restoreErr != nil {
		return errors.Join(cause, fmt.Errorf("game session: roll back participants: %w", restoreErr))
	}

	session.commands = session.commands[:transaction.commandCount]
	session.history = session.history[:transaction.historyCount]
	session.processed = transaction.processed

	return cause
}

// canonicalizeTickCommands makes execution independent of transport arrival order within the target tick.
func canonicalizeTickCommands(commands []simulation.Command) {
	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Player != commands[j].Player {
			return commands[i].Player < commands[j].Player
		}

		if commands[i].Sequence != commands[j].Sequence {
			return commands[i].Sequence < commands[j].Sequence
		}

		return commands[i].Kind < commands[j].Kind
	})
}

// applyTickCommandsLocked executes canonical commands and records only those whose handler completed successfully.
func (session *Session) applyTickCommandsLocked(commands []simulation.Command) error {
	for _, command := range commands {
		handler, found := session.handlers[command.Kind]
		if !found {
			return fmt.Errorf("%w: missing handler %q", ErrCommandApply, command.Kind)
		}

		if err := handler.Apply(session.engine, command); err != nil {
			return fmt.Errorf(
				"%w: %s player=%q sequence=%d: %v",
				ErrCommandApply,
				command.Kind,
				command.Player,
				command.Sequence,
				err,
			)
		}

		session.commands = append(session.commands, command)
		session.markProcessedLocked(command.Player, command.Sequence)
	}

	return nil
}

// expireAcceptedCommandsLocked bounds idempotency evidence to the same time window available for rollback.
func (session *Session) expireAcceptedCommandsLocked(current uint64) {
	for key, accepted := range session.recoveryAccepted {
		if current >= accepted.Tick && current-accepted.Tick > session.config.RollbackWindow {
			delete(session.recoveryAccepted, key)
		}
	}
}

// rollbackInsertLocked inserts late input by restoring the preceding frame and replaying through the current tick.
func (session *Session) rollbackInsertLocked(command simulation.Command, current uint64) error {
	baseTick := command.Tick - 1

	base := session.findRollbackFrameLocked(baseTick)
	if base == nil {
		return fmt.Errorf("%w: rollback history missing tick %d", simulation.ErrCommandTick, baseTick)
	}

	transaction, err := session.captureRollbackTransactionLocked()
	if err != nil {
		return err
	}

	kept, replay := partitionCommandsForReplay(session.commands, command.Tick)
	replay = append(replay, command)
	session.prepareReplayTimelineLocked(command.Tick, current, baseTick, kept)

	if err := session.engine.Restore(base.ecs); err != nil {
		return session.restoreRollbackTransactionLocked(transaction, err)
	}

	if err := session.restoreParticipantsLocked(base.participants); err != nil {
		return session.restoreRollbackTransactionLocked(transaction, err)
	}

	if err := session.replayCommandsLocked(command.Tick, current, replay); err != nil {
		return session.restoreRollbackTransactionLocked(transaction, err)
	}

	return nil
}

// findRollbackFrameLocked locates the exact pre-tick boundary required for deterministic late-input replay.
func (session *Session) findRollbackFrameLocked(tick uint64) *rollbackFrame {
	for index := range session.history {
		if session.history[index].tick == tick {
			return &session.history[index]
		}
	}

	return nil
}

// partitionCommandsForReplay separates immutable earlier history from commands that the late input can reorder.
func partitionCommandsForReplay(
	commands []simulation.Command,
	firstReplayTick uint64,
) ([]simulation.Command, []simulation.Command) {
	kept := make([]simulation.Command, 0, len(commands))
	replay := make([]simulation.Command, 0)

	for _, command := range commands {
		if command.Tick >= firstReplayTick {
			replay = append(replay, command)
			continue
		}

		kept = append(kept, command)
	}

	return kept, replay
}

// prepareReplayTimelineLocked removes derived future state and rebuilds sequence progress from retained commands.
func (session *Session) prepareReplayTimelineLocked(firstTick, current, baseTick uint64, kept []simulation.Command) {
	session.commands = kept

	for tick := firstTick; tick <= current; tick++ {
		delete(session.pending, tick)
	}

	for len(session.history) > 0 && session.history[len(session.history)-1].tick > baseTick {
		session.history = session.history[:len(session.history)-1]
	}

	for len(session.checkpoints) > 0 && session.checkpoints[len(session.checkpoints)-1].Tick > baseTick {
		session.checkpoints = session.checkpoints[:len(session.checkpoints)-1]
	}

	session.processed = make(map[string]sequenceProgress)
	for _, command := range kept {
		session.markProcessedLocked(command.Player, command.Sequence)
	}
}

// replayCommandsLocked queues commands at their original ticks and reuses normal stepping to preserve all invariants.
func (session *Session) replayCommandsLocked(firstTick, current uint64, commands []simulation.Command) error {
	for tick := firstTick; tick <= current; tick++ {
		for _, command := range commands {
			if command.Tick == tick {
				session.pending[tick] = append(session.pending[tick], command)
			}
		}

		if err := session.stepLocked(); err != nil {
			return err
		}
	}

	return nil
}

// captureRollbackTransactionLocked copies all state that a multi-tick replay can replace.
func (session *Session) captureRollbackTransactionLocked() (rollbackTransaction, error) {
	ecs, err := session.engine.Snapshot()
	if err != nil {
		return rollbackTransaction{}, err
	}

	participants, err := session.snapshotParticipantsLocked()
	if err != nil {
		return rollbackTransaction{}, err
	}

	return rollbackTransaction{
		ecs:          ecs,
		participants: participants,
		pending:      clonePending(session.pending),
		commands:     append([]simulation.Command(nil), session.commands...),
		checkpoints:  append([]simulation.Checkpoint(nil), session.checkpoints...),
		history:      append([]rollbackFrame(nil), session.history...),
		processed:    cloneProcessed(session.processed),
	}, nil
}

// restoreRollbackTransactionLocked restores a failed replay and joins any restoration errors with its original cause.
func (session *Session) restoreRollbackTransactionLocked(transaction rollbackTransaction, cause error) error {
	var restoreErrors []error
	if err := session.engine.Restore(transaction.ecs); err != nil {
		restoreErrors = append(restoreErrors, fmt.Errorf("restore ECS: %w", err))
	}

	if err := session.restoreParticipantsLocked(transaction.participants); err != nil {
		restoreErrors = append(restoreErrors, fmt.Errorf("restore participants: %w", err))
	}

	session.pending = transaction.pending
	session.commands = transaction.commands
	session.checkpoints = transaction.checkpoints
	session.history = transaction.history
	session.processed = transaction.processed

	return errors.Join(append([]error{cause}, restoreErrors...)...)
}

// clonePending copies command slices so rollback restoration cannot alias subsequently rewritten queues.
func clonePending(source map[uint64][]simulation.Command) map[uint64][]simulation.Command {
	result := make(map[uint64][]simulation.Command, len(source))
	for tick, commands := range source {
		result[tick] = append([]simulation.Command(nil), commands...)
	}

	return result
}

// cloneProcessed copies nested sequence sets so failed replay bookkeeping can be restored independently.
func cloneProcessed(source map[string]sequenceProgress) map[string]sequenceProgress {
	result := make(map[string]sequenceProgress, len(source))
	for player, progress := range source {
		copy := sequenceProgress{ack: progress.ack, beyond: make(map[uint64]struct{}, len(progress.beyond))}
		for sequence := range progress.beyond {
			copy.beyond[sequence] = struct{}{}
		}

		result[player] = copy
	}

	return result
}

// hasSequenceLocked checks completed and queued input so retransmission cannot apply a sequence twice.
func (session *Session) hasSequenceLocked(player string, sequence uint64) bool {
	progress := session.processed[player]
	if sequence <= progress.ack {
		return true
	}

	if _, found := progress.beyond[sequence]; found {
		return true
	}

	for _, commands := range session.pending {
		for _, command := range commands {
			if command.Player == player && command.Sequence == sequence {
				return true
			}
		}
	}

	return false
}

// markProcessedLocked advances the contiguous acknowledgement while retaining out-of-order completed sequences.
func (session *Session) markProcessedLocked(player string, sequence uint64) {
	progress := session.processed[player]
	if progress.beyond == nil {
		progress.beyond = make(map[uint64]struct{})
	}

	progress.beyond[sequence] = struct{}{}
	for {
		if _, found := progress.beyond[progress.ack+1]; !found {
			break
		}

		progress.ack++
		delete(progress.beyond, progress.ack)
	}

	session.processed[player] = progress
}
