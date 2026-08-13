package session

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// IntentController is a thread-safe mailbox between presentation code and the
// fixed-tick command stream. It knows nothing about Diablo commands or rules.
// A loaded mod chooses the command kind and payload; the session supplies the
// authenticated player, tick, authority, and deterministic sequence later.
type IntentController struct {
	mu       sync.Mutex
	requests []CommandIntent
	sequence uint64
}

// CommandIntent is an immutable command request waiting for the next tick.
type CommandIntent struct {
	Kind    string
	Payload any
}

// Submit copies a serializable request into the mailbox.
func (controller *IntentController) Submit(kind string, payload any) error {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return fmt.Errorf("command intent: kind is required")
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("command intent %q payload: %w", kind, err)
	}
	var copied any
	if err := json.Unmarshal(encoded, &copied); err != nil {
		return fmt.Errorf("command intent %q copy: %w", kind, err)
	}
	controller.mu.Lock()
	controller.requests = append(controller.requests, CommandIntent{
		Kind:    kind,
		Payload: copied,
	})
	controller.mu.Unlock()
	return nil
}

func (controller *IntentController) drain() []CommandIntent {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	requests := controller.requests
	controller.requests = nil
	return requests
}

// Drain returns the queued copied intents for an authenticated remote-session
// adapter. Offline sessions use IntentSource; both paths consume the same
// presentation mailbox exactly once.
func (controller *IntentController) Drain() []CommandIntent { return controller.drain() }

// IntentSource turns queued requests into transport-neutral player commands.
type IntentSource struct {
	controller *IntentController
	player     string
}

func NewIntentSource(controller *IntentController, player string) (*IntentSource, error) {
	player = strings.TrimSpace(player)
	if controller == nil || player == "" {
		return nil, fmt.Errorf("command intent source requires controller and player")
	}
	return &IntentSource{controller: controller, player: player}, nil
}

func (source *IntentSource) Commands(tick uint64) []simulation.Command {
	requests := source.controller.drain()
	commands := make([]simulation.Command, 0, len(requests))
	for _, request := range requests {
		payload, err := json.Marshal(request.Payload)
		if err != nil {
			continue
		}
		source.controller.sequence++
		commands = append(commands, simulation.Command{
			Tick:      tick,
			Player:    source.player,
			Authority: simulation.AuthorityPlayer,
			Sequence:  source.controller.sequence,
			Kind:      request.Kind,
			Payload:   payload,
		})
	}
	return commands
}
