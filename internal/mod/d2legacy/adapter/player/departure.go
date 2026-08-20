package player

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const LeaveCommand = "system.player.leave"

// DepartureCommand creates the trusted membership command that removes one
// player. Fixed authority and actor values prevent clients from forging exits.
func DepartureCommand(player string, sequence, tick uint64) (simulation.Command, error) {
	player = strings.TrimSpace(player)
	if player == "" || sequence == 0 || tick == 0 {
		return simulation.Command{}, fmt.Errorf("player: departure requires player, sequence, and tick")
	}

	payload, err := json.Marshal(map[string]string{"player": player})
	if err != nil {
		return simulation.Command{}, err
	}

	return simulation.Command{
		Tick:      tick,
		Player:    "system:membership",
		Authority: simulation.AuthoritySystem,
		Sequence:  sequence,
		Kind:      LeaveCommand,
		Payload:   payload,
	}, nil
}

// DepartureQueue serializes host-trusted membership removals onto the ordinary
// deterministic session command stream. Sharing it across server topologies
// ensures transport details cannot change cleanup ordering.
type DepartureQueue struct {
	mu       sync.Mutex
	sequence uint64
}

// Submit reserves a unique departure sequence and schedules it for the next
// session tick. The lock spans scheduling so concurrent disconnects cannot
// publish commands in an order different from their assigned sequences.
func (queue *DepartureQueue) Submit(session *gamesession.Session, player string) error {
	if queue == nil || session == nil {
		return fmt.Errorf("player: departure queue requires a session")
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	queue.sequence++

	return session.SubmitNext(func(tick uint64) (simulation.Command, error) {
		return DepartureCommand(player, queue.sequence, tick)
	})
}
