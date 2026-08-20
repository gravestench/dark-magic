// Package movement adapts native pointer and keyboard input into d2legacy player movement commands.
// Pathfinding and command transport remain generic engine services; command policy belongs to this adapter.
package movement

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
)

// MovementSource converts the latest native input snapshot into one replayable command per active gameplay tick.
type MovementSource struct {
	mu         sync.Mutex
	engine     *gameecs.Engine
	input      *inputstate.Store
	player     string
	focusID    string
	control    *MovementController
	navigation movementPathFinder
	path       []gameworld.Point
	pathTarget *MoveTarget
}

// NewMovementSource binds local input and presentation state to one authoritative player identity.
// Optional controller injection lets Lua UI and native input share the same thread-safe intent mailbox.
func NewMovementSource(
	engine *gameecs.Engine,
	input *inputstate.Store,
	player string,
	focusID string,
	controllers ...*MovementController,
) (*MovementSource, error) {
	player = strings.TrimSpace(player)

	focusID = strings.TrimSpace(focusID)
	if engine == nil || input == nil || player == "" || focusID == "" {
		return nil, fmt.Errorf("d2legacy movement: source requires engine, input, player, and focus owner")
	}

	control := &MovementController{}
	if len(controllers) > 0 && controllers[0] != nil {
		control = controllers[0]
	}

	return &MovementSource{
		engine:  engine,
		input:   input,
		player:  player,
		focusID: focusID,
		control: control,
	}, nil
}

// SetPlayer changes the command identity while rejecting blank replacements that would lose authority attribution.
func (source *MovementSource) SetPlayer(player string) {
	player = strings.TrimSpace(player)
	if player == "" {
		return
	}

	source.mu.Lock()
	source.player = player
	source.mu.Unlock()
}

// SetEngine changes only the read-only presentation source used for local position and path sampling.
// Connected clients bind their replica ECS here; command admission and simulation remain server-owned.
func (source *MovementSource) SetEngine(engine *gameecs.Engine) {
	if engine == nil {
		return
	}

	source.mu.Lock()
	source.engine = engine
	source.resetAcceptedPath()
	source.mu.Unlock()
}

// Commands samples input and route state atomically so a tick cannot combine values from two world bindings.
func (source *MovementSource) Commands(tick uint64) []simulation.Command {
	source.mu.Lock()
	defer source.mu.Unlock()

	if !source.hasPlayerControl() {
		return nil
	}

	target := source.resolvePointerTarget()
	x, y, target := source.applyFocusedInput(target)
	payload := source.encodeMovePayload(x, y, target)

	return []simulation.Command{{
		Tick:      tick,
		Player:    source.player,
		Authority: simulation.AuthorityPlayer,
		Sequence:  source.control.nextSequence(),
		Kind:      MoveCommand,
		Payload:   payload,
	}}
}

// hasPlayerControl suppresses commands until the presentation replica contains an active control component.
func (source *MovementSource) hasPlayerControl() bool {
	controls, present := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.player_control")
	return present && controls.Len() > 0
}

// resolvePointerTarget advances accepted routes but clears stale path state when pointer intent has ended.
func (source *MovementSource) resolvePointerTarget() *MoveTarget {
	target := source.control.moveTarget()
	if target == nil {
		source.resetAcceptedPath()
		return nil
	}

	if source.navigation == nil {
		return target
	}

	return source.pathWaypoint(target)
}

// applyFocusedInput admits only gameplay-owned keys and gives directional input precedence over pointer routing.
func (source *MovementSource) applyFocusedInput(target *MoveTarget) (int, int, *MoveTarget) {
	owner := source.input.Owner()
	if owner.Domain != inputstate.FocusScene || (owner.ID != source.focusID && !source.input.Gameplay()) {
		return 0, 0, target
	}

	if source.input.Action("toggle_run").Pressed {
		source.control.SetRunning(!source.control.Running())
	}

	x := movementAxis(source.input.Action("left"), source.input.Action("right"))

	y := movementAxis(source.input.Action("up"), source.input.Action("down"))
	if x == 0 && y == 0 {
		return x, y, target
	}

	// Keyboard motion explicitly replaces click-to-move so an old route cannot resume after key release.
	source.control.clearMoveTarget()

	return x, y, nil
}

// movementAxis collapses two opposing actions into the same -1/0/1 wire range used by replay commands.
func movementAxis(negative, positive inputstate.ActionState) int {
	axis := 0
	if negative.Down || negative.Pressed {
		axis--
	}

	if positive.Down || positive.Pressed {
		axis++
	}

	return axis
}

// encodeMovePayload serializes a closed value-only schema whose fields cannot fail JSON encoding.
func (source *MovementSource) encodeMovePayload(x, y int, target *MoveTarget) []byte {
	payload, _ := json.Marshal(MovePayload{
		X:       x,
		Y:       y,
		Running: source.control.Running(),
		Target:  target,
	})

	return payload
}
