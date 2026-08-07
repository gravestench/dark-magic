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
	ErrClosed       = errors.New("game session: closed")
	ErrHandler      = errors.New("game session: command handler is invalid")
	ErrCommandApply = errors.New("game session: admitted command failed to apply")
)

type Config struct {
	Step               time.Duration
	MaxCatchUp         int
	MaxCommandLead     uint64
	CheckpointInterval uint64
}

type CommandHandler struct {
	Validate simulation.CommandValidator
	Apply    func(*gameecs.Engine, simulation.Command) error
}

// Session serializes command admission, deterministic command ordering, ECS
// updates, and replay recording behind one authority boundary.
type Session struct {
	mu          sync.Mutex
	engine      *gameecs.Engine
	admitter    *simulation.Admitter
	config      Config
	handlers    map[string]CommandHandler
	pending     map[uint64][]simulation.Command
	initial     gameecs.Snapshot
	commands    []simulation.Command
	checkpoints []simulation.Checkpoint
	lag         time.Duration
	closed      bool
}

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

func (session *Session) Register(kind string, handler CommandHandler) error {
	if handler.Validate == nil || handler.Apply == nil {
		return ErrHandler
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return ErrClosed
	}
	if err := session.admitter.Register(kind, handler.Validate); err != nil {
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
	if session.closed {
		return ErrClosed
	}
	command.Player = strings.TrimSpace(command.Player)
	command.Kind = strings.TrimSpace(command.Kind)
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
			return fmt.Errorf("%w: missing handler %q", ErrCommandApply, command.Kind)
		}
		if err := handler.Apply(session.engine, command); err != nil {
			return fmt.Errorf("%w: %s player=%q sequence=%d: %v", ErrCommandApply, command.Kind, command.Player, command.Sequence, err)
		}
		session.commands = append(session.commands, command)
	}
	delete(session.pending, tick)
	if err := session.engine.Update(session.config.Step); err != nil {
		return err
	}
	if tick%session.config.CheckpointInterval == 0 {
		return session.checkpointLocked()
	}
	return nil
}

// Advance converts elapsed host time into bounded fixed simulation steps.
func (session *Session) Advance(elapsed time.Duration) (int, error) {
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
	checksum, err := snapshot.Checksum()
	if err != nil {
		return err
	}
	copy := snapshot
	session.checkpoints = append(session.checkpoints, simulation.Checkpoint{Tick: snapshot.Tick, Checksum: checksum, Snapshot: &copy})
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
	}
	return simulation.Replay{Version: simulation.ReplayVersion, StepNanos: int64(session.config.Step), Initial: initial, Commands: commands, Checkpoints: checkpoints}, nil
}

func cloneSnapshot(snapshot gameecs.Snapshot) (gameecs.Snapshot, error) {
	encoded, err := snapshot.Marshal()
	if err != nil {
		return gameecs.Snapshot{}, err
	}
	return gameecs.UnmarshalSnapshot(encoded)
}

func (session *Session) Close() error {
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		return nil
	}
	session.closed = true
	return session.engine.Close()
}
