// Package session owns one transport- and renderer-independent authoritative
// game simulation session.
package session

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

var (
	// ErrClosed rejects work after session authority has been released.
	ErrClosed = errors.New("game session: closed")
	// ErrHandler reports an incomplete command contract.
	ErrHandler = errors.New("game session: command handler is invalid")
	// ErrCommandApply wraps a command that was admitted but could not be applied.
	ErrCommandApply    = errors.New("game session: admitted command failed to apply")
	ErrCommandSequence = errors.New("game session: duplicate command sequence")
)

// Config bounds deterministic fixed stepping and command lead time. Zero values
// select production defaults so tests and offline composition share one policy.
type Config struct {
	Step               time.Duration
	MaxCatchUp         int
	MaxCommandLead     uint64
	CheckpointInterval uint64
	RollbackWindow     uint64
}

// CommandHandler separates untrusted payload validation from trusted mutation.
// Allowed authority classes are explicit; omission means ordinary player only.
type CommandHandler struct {
	Validate simulation.CommandValidator
	Apply    func(*gameecs.Engine, simulation.Command) error
	Allowed  []simulation.Authority
}

// CommandSource samples local intent for exactly one upcoming fixed tick. It is
// not used for network commands, whose arrival is handled through Submit.
type CommandSource func(nextTick uint64) []simulation.Command

// Session serializes command admission, deterministic command ordering, ECS
// updates, and replay recording behind one authority boundary.
type Session struct {
	mu                  sync.Mutex
	engine              *gameecs.Engine
	admitter            *simulation.Admitter
	config              Config
	handlers            map[string]CommandHandler
	participants        []simulation.StateParticipant
	pending             map[uint64][]simulation.Command
	initial             gameecs.Snapshot
	initialParticipants []simulation.ParticipantState
	commands            []simulation.Command
	checkpoints         []simulation.Checkpoint
	lag                 time.Duration
	closed              bool
	history             []rollbackFrame
	processed           map[string]sequenceProgress
	recoveryAccepted    map[string]simulation.Command
}

// sequenceProgress tracks one player's contiguous acknowledgement and completed gaps beyond it.
type sequenceProgress struct {
	ack    uint64
	beyond map[uint64]struct{}
}

// pendingParticipant holds initial state until an entire registration batch has validated successfully.
type pendingParticipant struct {
	participant simulation.StateParticipant
	state       simulation.ParticipantState
}

// RegisterAuthoritativeRuntime pins one rule implementation and its engine-owned
// durable stores into the session. The runtime may be Lua today or another
// adapter later; session determinism depends only on these language-neutral
// participants.
func (session *Session) RegisterAuthoritativeRuntime(
	identity simulation.RuntimeIdentity,
	stores *simulation.StateStore,
	additional ...simulation.StateParticipant,
) error {
	if stores == nil {
		return fmt.Errorf("game session: authoritative state store is required")
	}

	identityParticipant, err := simulation.NewIdentityParticipant(identity)
	if err != nil {
		return err
	}

	participants := []simulation.StateParticipant{identityParticipant, stores}
	participants = append(participants, additional...)

	return session.registerStateParticipants(participants...)
}

// RegisterStateParticipant adds deterministic authoritative state to session
// initial state, checkpoints, replay restoration, and desync verification.
// Register participants before the first command is submitted or tick advances.
func (session *Session) RegisterStateParticipant(participant simulation.StateParticipant) error {
	return session.registerStateParticipants(participant)
}

// registerStateParticipants validates and snapshots a batch atomically so partial registration cannot leak through.
func (session *Session) registerStateParticipants(participants ...simulation.StateParticipant) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	if err := session.validateParticipantRegistrationLocked(); err != nil {
		return err
	}

	pending, err := session.capturePendingParticipantsLocked(participants)
	if err != nil {
		return err
	}

	session.commitPendingParticipantsLocked(pending)

	return nil
}

// validateParticipantRegistrationLocked enforces the startup-only registration boundary under the session lock.
func (session *Session) validateParticipantRegistrationLocked() error {
	if session.closed {
		return ErrClosed
	}

	if session.engine.Tick() != session.initial.Tick || len(session.commands) != 0 || len(session.pending) != 0 {
		return fmt.Errorf("game session: state participants must be registered before commands or ticks")
	}

	return nil
}

// capturePendingParticipantsLocked validates IDs and copies initial state without mutating the live participant set.
func (session *Session) capturePendingParticipantsLocked(
	participants []simulation.StateParticipant,
) ([]pendingParticipant, error) {
	known := make(map[string]struct{}, len(session.participants)+len(participants))
	for _, existing := range session.participants {
		known[existing.StateID()] = struct{}{}
	}

	pending := make([]pendingParticipant, 0, len(participants))
	for _, participant := range participants {
		if participant == nil || strings.TrimSpace(participant.StateID()) == "" {
			return nil, fmt.Errorf("game session: state participant and ID are required")
		}

		if _, exists := known[participant.StateID()]; exists {
			return nil, fmt.Errorf("game session: duplicate state participant %q", participant.StateID())
		}

		data, err := participant.SnapshotState()
		if err != nil {
			return nil, fmt.Errorf("game session: snapshot initial participant %q: %w", participant.StateID(), err)
		}

		known[participant.StateID()] = struct{}{}
		pending = append(pending, pendingParticipant{
			participant: participant,
			state: simulation.ParticipantState{
				ID:   participant.StateID(),
				Data: append([]byte(nil), data...),
			},
		})
	}

	return pending, nil
}

// commitPendingParticipantsLocked publishes a validated batch in stable ID order for deterministic checksums.
func (session *Session) commitPendingParticipantsLocked(pending []pendingParticipant) {
	for _, item := range pending {
		session.participants = append(session.participants, item.participant)
		session.initialParticipants = append(session.initialParticipants, item.state)
	}

	sort.Slice(session.participants, func(i, j int) bool {
		return session.participants[i].StateID() < session.participants[j].StateID()
	})
	sort.Slice(session.initialParticipants, func(i, j int) bool {
		return session.initialParticipants[i].ID < session.initialParticipants[j].ID
	})

	if len(session.history) == 1 && session.history[0].tick == session.initial.Tick {
		// Keep the initial rollback boundary aligned with participants registered after session construction.
		session.history[0].participants = cloneParticipantStates(session.initialParticipants)
	}
}

// Status is an observational administration snapshot, never a mutation API.
type Status struct {
	Tick        uint64
	Pending     int
	Commands    int
	Privileged  int
	Checkpoints int
}

// New takes ownership of engine and captures its initial replay state. Non-ECS
// state participants must be registered before commands are queued or time moves.
func New(engine *gameecs.Engine, config Config) (*Session, error) {
	if engine == nil {
		return nil, fmt.Errorf("%w: nil engine", ErrHandler)
	}

	config = configWithDefaults(config)

	initial, err := engine.Snapshot()
	if err != nil {
		return nil, err
	}

	return &Session{
		engine:           engine,
		admitter:         simulation.NewAdmitter(config.MaxCommandLead),
		config:           config,
		handlers:         make(map[string]CommandHandler),
		pending:          make(map[uint64][]simulation.Command),
		initial:          initial,
		processed:        make(map[string]sequenceProgress),
		history:          []rollbackFrame{{tick: initial.Tick, ecs: initial}},
		recoveryAccepted: make(map[string]simulation.Command),
	}, nil
}

// configWithDefaults applies production timing and retention policy while preserving every explicit nonzero value.
func configWithDefaults(config Config) Config {
	if config.Step <= 0 {
		config.Step = gameecs.DefaultStep
	}

	if config.MaxCatchUp <= 0 {
		config.MaxCatchUp = gameecs.DefaultMaxCatchUp
	}

	if config.MaxCommandLead == 0 {
		config.MaxCommandLead = 2
	}

	if config.CheckpointInterval == 0 {
		config.CheckpointInterval = 25
	}

	if config.RollbackWindow == 0 {
		config.RollbackWindow = 8
	}

	return config
}

// Register binds one stable command kind to validation, authority policy, and
// trusted mutation. Command kinds are replay-format identities, not UI labels.
func (session *Session) Register(kind string, handler CommandHandler) error {
	if handler.Validate == nil || handler.Apply == nil {
		return ErrHandler
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return ErrClosed
	}

	allowed := handler.Allowed
	if len(allowed) == 0 {
		allowed = []simulation.Authority{simulation.AuthorityPlayer}
	}

	if err := session.admitter.RegisterAuthorities(kind, handler.Validate, allowed...); err != nil {
		return err
	}

	session.handlers[kind] = handler

	return nil
}

// Submit admits and queues a command. Network arrival order does not define
// execution order; Step sorts commands canonically within their target tick.
func (session *Session) Submit(command simulation.Command) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	return session.submitLocked(command)
}

// SubmitNext builds and admits a command for the next authoritative tick while
// holding the session lock. Trusted asynchronous producers use it when deriving
// a command from current session time; observing Tick and calling Submit in two
// operations would race the fixed-step runner.
func (session *Session) SubmitNext(build func(tick uint64) (simulation.Command, error)) error {
	if build == nil {
		return ErrHandler
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return ErrClosed
	}

	command, err := build(session.engine.Tick() + 1)
	if err != nil {
		return err
	}

	return session.submitLocked(command)
}

// SubmitNetwork admits a sequenced client input for its intended tick. Inputs
// inside the bounded history window trigger deterministic restore and replay;
// future inputs use ordinary admission. The server never trusts player or
// authority fields supplied by the client-facing protocol.
func (session *Session) SubmitNetwork(command simulation.Command) (uint64, error) {
	session.mu.Lock()
	defer session.mu.Unlock()

	current := session.engine.Tick()
	if command.Tick == 0 {
		return 0, fmt.Errorf("%w: command tick must be positive", simulation.ErrCommandTick)
	}

	if command.Sequence == 0 || session.hasSequenceLocked(command.Player, command.Sequence) {
		return 0, fmt.Errorf("%w: player=%q sequence=%d", ErrCommandSequence, command.Player, command.Sequence)
	}

	command.Player = strings.TrimSpace(command.Player)

	command.Kind = strings.TrimSpace(command.Kind)
	if command.Authority == "" {
		command.Authority = simulation.AuthorityPlayer
	}

	if err := session.admitter.ValidateNetwork(command, current); err != nil {
		return 0, err
	}

	if command.Tick > current {
		session.pending[command.Tick] = append(session.pending[command.Tick], command)
		return command.Tick, nil
	}

	if current-command.Tick >= session.config.RollbackWindow {
		return 0, fmt.Errorf("%w: rollback window current=%d command=%d", simulation.ErrCommandTick, current, command.Tick)
	}

	if err := session.rollbackInsertLocked(command, current); err != nil {
		return 0, err
	}

	return command.Tick, nil
}

// ProcessedSequence returns the highest contiguous sequence applied for a player under the session lock.
func (session *Session) ProcessedSequence(player string) uint64 {
	session.mu.Lock()
	defer session.mu.Unlock()

	return session.processed[player].ack
}

// AcceptedNetworkCommand returns the exact already-admitted input for an
// idempotency check at the transport boundary. A client may retransmit after
// the authority accepted a request but its response was lost; only the same
// command is safe to acknowledge as success.
func (session *Session) AcceptedNetworkCommand(player string, sequence uint64) (simulation.Command, bool) {
	session.mu.Lock()
	defer session.mu.Unlock()

	recoveryKey := recoveryCommandKey(simulation.Command{Player: player, Sequence: sequence})
	if command, found := session.recoveryAccepted[recoveryKey]; found {
		return command, true
	}

	for _, commands := range session.pending {
		for _, command := range commands {
			if command.Player == player && command.Sequence == sequence {
				return command, true
			}
		}
	}

	for index := len(session.commands) - 1; index >= 0; index-- {
		command := session.commands[index]
		if command.Player == player && command.Sequence == sequence {
			return command, true
		}
	}

	return simulation.Command{}, false
}

// submitLocked normalizes, validates, and queues future input while its caller holds the session lock.
func (session *Session) submitLocked(command simulation.Command) error {
	if session.closed {
		return ErrClosed
	}

	command.Player = strings.TrimSpace(command.Player)

	command.Kind = strings.TrimSpace(command.Kind)
	if command.Authority == "" {
		command.Authority = simulation.AuthorityPlayer
	}

	currentTick := session.engine.Tick()
	if command.Tick <= currentTick {
		return fmt.Errorf("%w: current=%d command=%d", simulation.ErrCommandTick, currentTick, command.Tick)
	}

	if err := session.admitter.Admit(command, currentTick); err != nil {
		return err
	}

	session.pending[command.Tick] = append(session.pending[command.Tick], command)

	return nil
}

// StepDuration is the authoritative simulation cadence advertised to clients.
func (session *Session) StepDuration() time.Duration {
	session.mu.Lock()
	defer session.mu.Unlock()

	return session.config.Step
}

// Advance converts elapsed host time into bounded fixed simulation steps.
func (session *Session) Advance(elapsed time.Duration) (int, error) {
	return session.AdvanceWithSource(elapsed, nil)
}

// AdvanceWithSource samples local authoritative inputs exactly once for every
// fixed tick that will execute. Remotely submitted commands use Submit instead.
func (session *Session) AdvanceWithSource(elapsed time.Duration, source CommandSource) (int, error) {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return 0, ErrClosed
	}

	if elapsed < 0 {
		elapsed = 0
	}

	maximumLag := session.config.Step * time.Duration(session.config.MaxCatchUp)
	session.lag = min(session.lag+elapsed, maximumLag)
	steps := 0

	for session.lag >= session.config.Step {
		if source != nil {
			for _, command := range source(session.engine.Tick() + 1) {
				if err := session.submitLocked(command); err != nil {
					return steps, err
				}
			}
		}

		if err := session.stepLocked(); err != nil {
			return steps, err
		}

		session.lag -= session.config.Step
		steps++
	}

	return steps, nil
}

// Run advances the session from elapsed monotonic time until cancellation.
// Ticker notifications are only wakeups: Go may coalesce them while the host is
// busy, so advancing exactly one step per notification would silently slow the
// game clock under load. Advance retains bounded residual lag and applies the
// same maximum catch-up policy used by local sessions.
func (session *Session) Run(ctx context.Context) error {
	ticker := time.NewTicker(session.config.Step)
	defer ticker.Stop()

	last := time.Now()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			now := time.Now()
			_, err := session.Advance(now.Sub(last))
			last = now

			if err != nil {
				return err
			}
		}
	}
}

// checkpointLocked appends a composite state checkpoint while its caller holds the session lock.
func (session *Session) checkpointLocked() error {
	snapshot, err := session.engine.Snapshot()
	if err != nil {
		return err
	}

	participants, err := session.snapshotParticipantsLocked()
	if err != nil {
		return err
	}

	checksum, err := simulation.CompositeChecksum(snapshot, participants)
	if err != nil {
		return err
	}

	copy := snapshot
	session.checkpoints = append(session.checkpoints, simulation.Checkpoint{
		Tick:         snapshot.Tick,
		Checksum:     checksum,
		Snapshot:     &copy,
		Participants: participants,
	})

	return nil
}

// Replay exports a defensive replay ending at the current completed tick.
func (session *Session) Replay() (simulation.Replay, error) {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return simulation.Replay{}, ErrClosed
	}

	tick := session.engine.Tick()

	checkpointMissing := len(session.checkpoints) == 0
	if !checkpointMissing {
		checkpointMissing = session.checkpoints[len(session.checkpoints)-1].Tick != tick
	}

	if tick > session.initial.Tick && checkpointMissing {
		// Always end an exported replay at a verifiable boundary, even between periodic checkpoints.
		if err := session.checkpointLocked(); err != nil {
			return simulation.Replay{}, err
		}
	}

	initial, err := cloneSnapshot(session.initial)
	if err != nil {
		return simulation.Replay{}, err
	}

	commands := make([]simulation.Command, len(session.commands))
	for index, command := range session.commands {
		commands[index] = command
		commands[index].Payload = append([]byte(nil), command.Payload...)
	}

	checkpoints := make([]simulation.Checkpoint, len(session.checkpoints))
	for index, checkpoint := range session.checkpoints {
		checkpoints[index] = checkpoint
		if checkpoint.Snapshot != nil {
			copy, err := cloneSnapshot(*checkpoint.Snapshot)
			if err != nil {
				return simulation.Replay{}, err
			}

			checkpoints[index].Snapshot = &copy
		}

		checkpoints[index].Participants = cloneParticipantStates(checkpoint.Participants)
	}

	return simulation.Replay{
		Version:             simulation.ReplayVersion,
		StepNanos:           int64(session.config.Step),
		Initial:             initial,
		InitialParticipants: cloneParticipantStates(session.initialParticipants),
		Commands:            commands,
		Checkpoints:         checkpoints,
	}, nil
}

// ReplayContainer exports the current replay with defensive copies of the
// caller's versioned manifests and semantic event evidence. The returned value
// can be encoded or atomically persisted by simulation's replay-container API.
func (session *Session) ReplayContainer(manifests map[string]simulation.ReplayManifest,
	events []simulation.ReplayEvent,
) (simulation.ReplayContainer, error) {
	replay, err := session.Replay()
	if err != nil {
		return simulation.ReplayContainer{}, err
	}

	container := simulation.NewReplayContainer(replay)

	container.Manifests = make(map[string]simulation.ReplayManifest, len(manifests))
	for name, manifest := range manifests {
		manifest.Data = append([]byte(nil), manifest.Data...)
		container.Manifests[name] = manifest
	}

	container.Events = make([]simulation.ReplayEvent, len(events))
	for index, event := range events {
		container.Events[index] = event
		container.Events[index].Payload = append([]byte(nil), event.Payload...)
	}

	return container, nil
}

// snapshotParticipantsLocked captures participant state in stable registration order and owns every returned byte.
func (session *Session) snapshotParticipantsLocked() ([]simulation.ParticipantState, error) {
	result := make([]simulation.ParticipantState, 0, len(session.participants))
	for _, participant := range session.participants {
		data, err := participant.SnapshotState()
		if err != nil {
			return nil, fmt.Errorf("game session: snapshot participant %q: %w", participant.StateID(), err)
		}

		result = append(result, simulation.ParticipantState{
			ID:   participant.StateID(),
			Data: append([]byte(nil), data...),
		})
	}

	return result, nil
}

// restoreParticipantsLocked restores by stable ID and copies bytes so participants cannot retain caller-owned buffers.
func (session *Session) restoreParticipantsLocked(states []simulation.ParticipantState) error {
	if len(states) != len(session.participants) {
		return fmt.Errorf("game session: participant count changed during tick")
	}

	byID := make(map[string]simulation.StateParticipant, len(session.participants))
	for _, participant := range session.participants {
		byID[participant.StateID()] = participant
	}

	for _, state := range states {
		participant, found := byID[state.ID]
		if !found {
			return fmt.Errorf("game session: missing participant %q", state.ID)
		}

		if err := participant.RestoreState(append([]byte(nil), state.Data...)); err != nil {
			return fmt.Errorf("game session: restore participant %q: %w", state.ID, err)
		}
	}

	return nil
}

// cloneParticipantStates deep-copies opaque bytes so exported or transactional state cannot alias session ownership.
func cloneParticipantStates(states []simulation.ParticipantState) []simulation.ParticipantState {
	result := make([]simulation.ParticipantState, len(states))
	for index, state := range states {
		result[index] = simulation.ParticipantState{ID: state.ID, Data: append([]byte(nil), state.Data...)}
	}

	return result
}

// Status returns a compact observational snapshot for administration surfaces.
func (session *Session) Status() Status {
	session.mu.Lock()
	defer session.mu.Unlock()

	pending := 0
	for _, commands := range session.pending {
		pending += len(commands)
	}

	privileged := 0

	for _, command := range session.commands {
		if command.Authority == simulation.AuthorityAdmin || command.Authority == simulation.AuthoritySystem {
			privileged++
		}
	}

	return Status{
		Tick:        session.engine.Tick(),
		Pending:     pending,
		Commands:    len(session.commands),
		Privileged:  privileged,
		Checkpoints: len(session.checkpoints),
	}
}

// Audit returns accepted privileged commands in canonical execution order.
func (session *Session) Audit() []simulation.Command {
	session.mu.Lock()
	defer session.mu.Unlock()

	result := make([]simulation.Command, 0)

	for _, command := range session.commands {
		if command.Authority != simulation.AuthorityAdmin && command.Authority != simulation.AuthoritySystem {
			continue
		}

		command.Payload = append([]byte(nil), command.Payload...)
		result = append(result, command)
	}

	return result
}

// cloneSnapshot round-trips the canonical encoding to preserve all nested ownership and serialization semantics.
func cloneSnapshot(snapshot gameecs.Snapshot) (gameecs.Snapshot, error) {
	encoded, err := snapshot.Marshal()
	if err != nil {
		return gameecs.Snapshot{}, err
	}

	return gameecs.UnmarshalSnapshot(encoded)
}

// Close rejects future work and releases the owned ECS exactly once.
func (session *Session) Close() error {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.closed {
		return nil
	}

	session.closed = true

	return session.engine.Close()
}
