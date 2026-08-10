package item

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// Controller is a small mailbox between presentation and the fixed simulation
// clock. Lua can post intent, but it cannot mutate item authority directly.
type Controller struct {
	mu       sync.Mutex
	requests []MovePayload
	sequence uint64
}

func (controller *Controller) Move(payload MovePayload) error {
	payload.ItemID = strings.TrimSpace(payload.ItemID)
	if payload.ItemID == "" {
		return fmt.Errorf("item: item identity is required")
	}
	controller.mu.Lock()
	defer controller.mu.Unlock()
	controller.requests = append(controller.requests, payload)
	return nil
}

func (controller *Controller) drain() []MovePayload {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	requests := controller.requests
	controller.requests = nil
	return requests
}

type Source struct {
	controller *Controller
	player     string
}

func NewSource(controller *Controller, player string) (*Source, error) {
	player = strings.TrimSpace(player)
	if controller == nil || player == "" {
		return nil, fmt.Errorf("item: command source requires controller and player")
	}
	return &Source{controller: controller, player: player}, nil
}

func (source *Source) Commands(tick uint64) []simulation.Command {
	requests := source.controller.drain()
	commands := make([]simulation.Command, 0, len(requests))
	for _, request := range requests {
		source.controller.sequence++
		command, err := Command(request, source.player, source.controller.sequence, tick, simulation.AuthorityPlayer)
		if err == nil {
			commands = append(commands, command)
		}
	}
	return commands
}
