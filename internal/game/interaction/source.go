package interaction

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

type Controller struct {
	mu       sync.Mutex
	requests []Payload
	sequence uint64
}

func (controller *Controller) Open(target string) error {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		return fmt.Errorf("interaction: target is required")
	}
	controller.mu.Lock()
	controller.requests = append(controller.requests, Payload{Target: target})
	controller.mu.Unlock()
	return nil
}

func (controller *Controller) OpenAt(x, y float64) error {
	if !finite(x) || !finite(y) {
		return fmt.Errorf("interaction: coordinates must be finite")
	}
	controller.mu.Lock()
	controller.requests = append(controller.requests, Payload{At: true, X: x, Y: y})
	controller.mu.Unlock()
	return nil
}

func (controller *Controller) Close() {
	controller.mu.Lock()
	controller.requests = append(controller.requests, Payload{})
	controller.mu.Unlock()
}

type Source struct {
	controller *Controller
	player     string
}

func NewSource(controller *Controller, player string) (*Source, error) {
	player = strings.TrimSpace(player)
	if controller == nil || player == "" {
		return nil, fmt.Errorf("interaction: source requires controller and player")
	}
	return &Source{controller: controller, player: player}, nil
}

func (source *Source) Commands(tick uint64) []simulation.Command {
	source.controller.mu.Lock()
	requests := source.controller.requests
	source.controller.requests = nil
	source.controller.mu.Unlock()
	commands := make([]simulation.Command, 0, len(requests))
	for _, payload := range requests {
		source.controller.sequence++
		kind := OpenCommand
		if payload.Target == "" && !payload.At {
			kind = CloseCommand
		}
		command, err := Command(kind, payload, source.player, source.controller.sequence, tick, simulation.AuthorityPlayer)
		if err == nil {
			commands = append(commands, command)
		}
	}
	return commands
}
