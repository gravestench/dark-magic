package simulation

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrCommandIdentity = errors.New("simulation: command identity is required")
	ErrCommandTick     = errors.New("simulation: command tick is outside the admission window")
	ErrCommandSequence = errors.New("simulation: command sequence is not next")
	ErrCommandKind     = errors.New("simulation: command kind is not registered")
	ErrCommandPayload  = errors.New("simulation: command payload is invalid")
)

// Command is the transport-neutral input accepted by an authoritative session.
type Command struct {
	Tick     uint64          `json:"tick"`
	Player   string          `json:"player"`
	Sequence uint64          `json:"sequence"`
	Kind     string          `json:"kind"`
	Payload  json.RawMessage `json:"payload"`
}

type CommandValidator func(Command) error

// Admitter validates identity, timing, monotonic per-player sequence, kind, and
// payload before a command can enter the authoritative event log.
type Admitter struct {
	mu         sync.Mutex
	maxLead    uint64
	sequences  map[string]uint64
	ticks      map[string]uint64
	validators map[string]CommandValidator
}

func NewAdmitter(maxLead uint64) *Admitter {
	return &Admitter{maxLead: maxLead, sequences: make(map[string]uint64), ticks: make(map[string]uint64), validators: make(map[string]CommandValidator)}
}

func (admitter *Admitter) Register(kind string, validator CommandValidator) error {
	kind = strings.TrimSpace(kind)
	if kind == "" || validator == nil {
		return ErrCommandKind
	}
	admitter.mu.Lock()
	defer admitter.mu.Unlock()
	if _, exists := admitter.validators[kind]; exists {
		return fmt.Errorf("%w: %q already registered", ErrCommandKind, kind)
	}
	admitter.validators[kind] = validator
	return nil
}

func (admitter *Admitter) Admit(command Command, currentTick uint64) error {
	command.Player = strings.TrimSpace(command.Player)
	command.Kind = strings.TrimSpace(command.Kind)
	if command.Player == "" {
		return ErrCommandIdentity
	}
	if command.Tick < currentTick || command.Tick-currentTick > admitter.maxLead {
		return fmt.Errorf("%w: current=%d command=%d", ErrCommandTick, currentTick, command.Tick)
	}
	if !json.Valid(command.Payload) {
		return ErrCommandPayload
	}
	admitter.mu.Lock()
	defer admitter.mu.Unlock()
	validator, found := admitter.validators[command.Kind]
	if !found {
		return fmt.Errorf("%w: %q", ErrCommandKind, command.Kind)
	}
	want := admitter.sequences[command.Player] + 1
	if command.Sequence != want {
		return fmt.Errorf("%w: player=%q got=%d want=%d", ErrCommandSequence, command.Player, command.Sequence, want)
	}
	if previous, found := admitter.ticks[command.Player]; found && command.Tick < previous {
		return fmt.Errorf("%w: player=%q previous=%d command=%d", ErrCommandTick, command.Player, previous, command.Tick)
	}
	if err := validator(command); err != nil {
		return fmt.Errorf("%w: %v", ErrCommandPayload, err)
	}
	admitter.sequences[command.Player] = command.Sequence
	admitter.ticks[command.Player] = command.Tick
	return nil
}
