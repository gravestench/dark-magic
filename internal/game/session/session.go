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
	ErrCommandApply = errors.New("game session: admitted command failed to apply")
)

// Config bounds deterministic fixed stepping and command lead time. Zero values
// select production defaults so tests and offline composition share one policy.
type Config struct {
	Step               time.Duration
	MaxCatchUp         int
	MaxCommandLead     uint64
	CheckpointInterval uint64
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
}

// RegisterAuthoritativeRuntime pins one rule implementation and its engine-owned
// durable stores into the session. The runtime may be Lua today or another
// adapter later; session determinism depends only on these language-neutral
// participants.
func (session *Session) RegisterAuthoritativeRuntime(identity simulation.RuntimeIdentity, stores *simulation.StateStore, additional ...simulation.StateParticipant) error {
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

func (session *Session) registerStateParticipants(participants ...simulation.StateParticipant) error {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return ErrClosed
	}
	if session.engine.Tick() != session.initial.Tick || len(session.commands) != 0 || len(session.pending) != 0 {
		return fmt.Errorf("game session: state participants must be registered before commands or ticks")
	}
	known := make(map[string]struct{}, len(session.participants)+len(participants))
	for _, existing := range session.participants {
		known[existing.StateID()] = struct{}{}
	}
	type pendingParticipant struct {
		participant simulation.StateParticipant
		state       simulation.ParticipantState
	}
	pending := make([]pendingParticipant, 0, len(participants))
	for _, participant := range participants {
		if participant == nil || strings.TrimSpace(participant.StateID()) == "" {
			return fmt.Errorf("game session: state participant and ID are required")
		}
		if _, exists := known[participant.StateID()]; exists {
			return fmt.Errorf("game session: duplicate state participant %q", participant.StateID())
		}
		data, err := participant.SnapshotState()
		if err != nil {
			return fmt.Errorf("game session: snapshot initial participant %q: %w", participant.StateID(), err)
		}
		known[participant.StateID()] = struct{}{}
		pending = append(pending, pendingParticipant{participant: participant, state: simulation.ParticipantState{ID: participant.StateID(), Data: append([]byte(nil), data...)}})
	}
	for _, item := range pending {
		session.participants = append(session.participants, item.participant)
		session.initialParticipants = append(session.initialParticipants, item.state)
	}
	sort.Slice(session.participants, func(i, j int) bool { return session.participants[i].StateID() < session.participants[j].StateID() })
	sort.Slice(session.initialParticipants, func(i, j int) bool { return session.initialParticipants[i].ID < session.initialParticipants[j].ID })
	return nil
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
	initial, err := engine.Snapshot()
	if err != nil {
		return nil, err
	}
	return &Session{
		engine: engine, admitter: simulation.NewAdmitter(config.MaxCommandLead), config: config,
		handlers: make(map[string]CommandHandler), pending: make(map[uint64][]simulation.Command), initial: initial,
	}, nil
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

// Step executes exactly one authoritative simulation tick.
func (session *Session) Step() error {
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.stepLocked()
}

func (session *Session) stepLocked() error {
	if session.closed {
		return ErrClosed
	}
	// A tick is one transaction. Command handlers and systems may touch several
	// ECS components and script-owned state values before returning an error.
	// Capture both stores before doing any work so failure cannot expose half of
	// a command or half of a system schedule to the next tick.
	beforeECS, err := session.engine.Snapshot()
	if err != nil {
		return err
	}
	beforeParticipants, err := session.snapshotParticipantsLocked()
	if err != nil {
		return err
	}
	beforeCommands := len(session.commands)
	rollback := func(cause error) error {
		if restoreErr := session.engine.Restore(beforeECS); restoreErr != nil {
			return errors.Join(cause, fmt.Errorf("game session: roll back ECS: %w", restoreErr))
		}
		if restoreErr := session.restoreParticipantsLocked(beforeParticipants); restoreErr != nil {
			return errors.Join(cause, fmt.Errorf("game session: roll back participants: %w", restoreErr))
		}
		session.commands = session.commands[:beforeCommands]
		return cause
	}
	tick := session.engine.Tick() + 1
	commands := session.pending[tick]
	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Player != commands[j].Player {
			return commands[i].Player < commands[j].Player
		}
		if commands[i].Sequence != commands[j].Sequence {
			return commands[i].Sequence < commands[j].Sequence
		}
		return commands[i].Kind < commands[j].Kind
	})
	for _, command := range commands {
		handler, found := session.handlers[command.Kind]
		if !found {
			return rollback(fmt.Errorf("%w: missing handler %q", ErrCommandApply, command.Kind))
		}
		if err := handler.Apply(session.engine, command); err != nil {
			return rollback(fmt.Errorf("%w: %s player=%q sequence=%d: %v", ErrCommandApply, command.Kind, command.Player, command.Sequence, err))
		}
		session.commands = append(session.commands, command)
	}
	delete(session.pending, tick)
	if err := session.engine.Update(session.config.Step); err != nil {
		// Keep the queued commands available for an administrator to inspect or
		// retry after the faulty runtime is replaced.
		session.pending[tick] = commands
		return rollback(err)
	}
	if tick%session.config.CheckpointInterval == 0 {
		return session.checkpointLocked()
	}
	return nil
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

// Run advances the session from a wall-clock ticker until cancellation.
func (session *Session) Run(ctx context.Context) error {
	ticker := time.NewTicker(session.config.Step)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := session.Step(); err != nil {
				return err
			}
		}
	}
}

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
	session.checkpoints = append(session.checkpoints, simulation.Checkpoint{Tick: snapshot.Tick, Checksum: checksum, Snapshot: &copy, Participants: participants})
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
	if tick > session.initial.Tick && (len(session.checkpoints) == 0 || session.checkpoints[len(session.checkpoints)-1].Tick != tick) {
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
	return simulation.Replay{Version: simulation.ReplayVersion, StepNanos: int64(session.config.Step), Initial: initial, InitialParticipants: cloneParticipantStates(session.initialParticipants), Commands: commands, Checkpoints: checkpoints}, nil
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

func (session *Session) snapshotParticipantsLocked() ([]simulation.ParticipantState, error) {
	result := make([]simulation.ParticipantState, 0, len(session.participants))
	for _, participant := range session.participants {
		data, err := participant.SnapshotState()
		if err != nil {
			return nil, fmt.Errorf("game session: snapshot participant %q: %w", participant.StateID(), err)
		}
		result = append(result, simulation.ParticipantState{ID: participant.StateID(), Data: append([]byte(nil), data...)})
	}
	return result, nil
}

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
	return Status{Tick: session.engine.Tick(), Pending: pending, Commands: len(session.commands), Privileged: privileged, Checkpoints: len(session.checkpoints)}
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
