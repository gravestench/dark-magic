// Package movement adapts native pointer/keyboard input into d2legacy player
// movement commands. Pathfinding and command transport are generic engine
// mechanisms; schema names and player.move payload policy belong to the mod.
package movement

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
)

const MoveCommand = "player.move"

type MovePayload struct {
	X       int         `json:"x"`
	Y       int         `json:"y"`
	Running bool        `json:"running"`
	Target  *MoveTarget `json:"target,omitempty"`
}

// MoveTarget is a wire value consumed by d2legacy Lua. Keep explicit JSON names:
// Go's default capitalized field names are different keys in Lua tables.
type MoveTarget struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	StopRadius float64 `json:"stop_radius"`
}

// MovementController is the thread-safe local intent mailbox shared by Lua UI
// and the fixed-tick movement command source. It never mutates ECS state.
type MovementController struct {
	running  atomic.Bool
	sequence atomic.Uint64
	mu       sync.Mutex
	target   *MoveTarget
}

func (controller *MovementController) SetRunning(running bool) { controller.running.Store(running) }
func (controller *MovementController) Running() bool           { return controller.running.Load() }
func (controller *MovementController) HasMoveTarget() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.target != nil
}
func (controller *MovementController) nextSequence() uint64 { return controller.sequence.Add(1) }

func (controller *MovementController) SetMoveTarget(x, y float64) error {
	return controller.SetMoveTargetWithRadius(x, y, 0)
}

func (controller *MovementController) SetMoveTargetWithRadius(x, y, stopRadius float64) error {
	if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
		return fmt.Errorf("d2legacy movement: target must be finite")
	}
	if stopRadius < 0 || math.IsNaN(stopRadius) || math.IsInf(stopRadius, 0) {
		return fmt.Errorf("d2legacy movement: stop radius must be non-negative and finite")
	}
	controller.mu.Lock()
	defer controller.mu.Unlock()
	controller.target = &MoveTarget{X: x, Y: y, StopRadius: stopRadius}
	return nil
}

func (controller *MovementController) moveTarget() *MoveTarget {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.target == nil {
		return nil
	}
	result := *controller.target
	return &result
}

func (controller *MovementController) clearMoveTarget() {
	controller.mu.Lock()
	controller.target = nil
	controller.mu.Unlock()
}

func (controller *MovementController) restoreMoveTarget(target *MoveTarget) {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if target == nil {
		controller.target = nil
		return
	}
	copyTarget := *target
	controller.target = &copyTarget
}

// MovementSource turns the latest native input snapshot into one replayable
// command per active gameplay tick.
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

func (source *MovementSource) SetPlayer(player string) {
	player = strings.TrimSpace(player)
	if player == "" {
		return
	}
	source.mu.Lock()
	source.player = player
	source.mu.Unlock()
}

// SetEngine changes only the read-only presentation source used for local
// position/path sampling. Connected clients bind this to their replica ECS;
// commands are still admitted and simulated exclusively by the server.
func (source *MovementSource) SetEngine(engine *gameecs.Engine) {
	if engine == nil {
		return
	}
	source.mu.Lock()
	source.engine = engine
	source.path = nil
	source.pathTarget = nil
	source.mu.Unlock()
}

type movementPathFinder interface {
	FindPath(gameworld.PathRequest) ([]gameworld.Point, error)
}

func (source *MovementSource) SetNavigation(world *gameworld.Map) {
	source.mu.Lock()
	defer source.mu.Unlock()
	// Reliable connected corrections reinstall presentation for the current
	// level. Rebinding the identical map is not a world transition and must not
	// consume an in-progress click route; otherwise every correction turns into
	// an implicit stop command. A new map pointer still invalidates coordinates
	// and route state, including a regenerated instance of the same level ID.
	if current, ok := source.navigation.(*gameworld.Map); ok && current == world {
		return
	}
	source.navigation = world
	source.path = nil
	source.pathTarget = nil
	// Pointer targets and their accepted route are coordinates in one world.
	// Keeping the target while swapping maps can replan an old town coordinate
	// in the wilderness (or vice versa), where it then competes with fresh input.
	// Treat every navigation replacement as an explicit route invalidation.
	source.control.clearMoveTarget()
}

func NewMovementSource(engine *gameecs.Engine, input *inputstate.Store, player, focusID string, controllers ...*MovementController) (*MovementSource, error) {
	player = strings.TrimSpace(player)
	focusID = strings.TrimSpace(focusID)
	if engine == nil || input == nil || player == "" || focusID == "" {
		return nil, fmt.Errorf("d2legacy movement: source requires engine, input, player, and focus owner")
	}
	control := &MovementController{}
	if len(controllers) > 0 && controllers[0] != nil {
		control = controllers[0]
	}
	return &MovementSource{engine: engine, input: input, player: player, focusID: focusID, control: control}, nil
}

func (source *MovementSource) Commands(tick uint64) []simulation.Command {
	source.mu.Lock()
	defer source.mu.Unlock()
	controls, present := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.player_control")
	if !present || controls.Len() == 0 {
		return nil
	}
	x, y := 0, 0
	target := source.control.moveTarget()
	if target == nil {
		source.path = nil
		source.pathTarget = nil
	}
	if target != nil && source.navigation != nil {
		target = source.pathWaypoint(target)
	}
	owner := source.input.Owner()
	if owner.Domain == inputstate.FocusScene && (owner.ID == source.focusID || source.input.Gameplay()) {
		if source.input.Action("toggle_run").Pressed {
			source.control.SetRunning(!source.control.Running())
		}
		if action := source.input.Action("left"); action.Down || action.Pressed {
			x--
		}
		if action := source.input.Action("right"); action.Down || action.Pressed {
			x++
		}
		if action := source.input.Action("up"); action.Down || action.Pressed {
			y--
		}
		if action := source.input.Action("down"); action.Down || action.Pressed {
			y++
		}
		if x != 0 || y != 0 {
			source.control.clearMoveTarget()
			target = nil
		}
	}
	payload, _ := json.Marshal(MovePayload{X: x, Y: y, Running: source.control.Running(), Target: target})
	return []simulation.Command{{Tick: tick, Player: source.player, Authority: simulation.AuthorityPlayer, Sequence: source.control.nextSequence(), Kind: MoveCommand, Payload: payload}}
}

func (source *MovementSource) pathWaypoint(target *MoveTarget) *MoveTarget {
	positions, ok := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.position")
	if !ok {
		return target
	}
	controls, ok := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.player_control")
	if !ok {
		return target
	}
	var current gameworld.Point
	found := false
	var radius float64
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		player, _ := control.Get("player")
		if player != source.player {
			continue
		}
		position, present := positions.Get(entity)
		if !present {
			continue
		}
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		current = gameworld.Point{X: x.(float64), Y: y.(float64)}
		found = true
		if colliders, present := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.collider"); present {
			if collider, exists := colliders.Get(entity); exists {
				value, _ := collider.Get("radius")
				radius = value.(float64)
			}
		}
		break
	}
	if !found {
		return target
	}
	changed := source.pathTarget == nil || source.pathTarget.StopRadius != target.StopRadius ||
		gameworld.CollisionCell(source.pathTarget.X) != gameworld.CollisionCell(target.X) ||
		gameworld.CollisionCell(source.pathTarget.Y) != gameworld.CollisionCell(target.Y)
	if changed {
		path, err := source.navigation.FindPath(gameworld.PathRequest{Start: current, Goal: gameworld.Point{X: target.X, Y: target.Y}, Radius: radius, StopRadius: target.StopRadius})
		if err != nil {
			// A new click is only a proposed route replacement. If the player
			// already has an accepted route, keep it moving rather than letting
			// an unreachable wall/void click act like an implicit stop command.
			if source.pathTarget != nil && len(source.path) > 1 {
				oldTarget := *source.pathTarget
				source.control.restoreMoveTarget(&oldTarget)
				target = &oldTarget
			} else {
				source.control.clearMoveTarget()
				source.path = nil
				source.pathTarget = nil
				return &MoveTarget{X: current.X, Y: current.Y}
			}
		} else {
			source.path = path
			copyTarget := *target
			source.pathTarget = &copyTarget
		}
	} else if source.pathTarget.X != target.X || source.pathTarget.Y != target.Y {
		// Camera following changes the world coordinate beneath a held pointer by
		// a fraction every frame. Navigation is collision-cell based, so preserve
		// the accepted route and only move its exact final destination until the
		// pointer actually crosses into another cell.
		copyTarget := *target
		source.pathTarget = &copyTarget
		if len(source.path) > 0 {
			source.path[len(source.path)-1] = gameworld.Point{X: target.X, Y: target.Y}
		}
	}
	for len(source.path) > 1 && math.Hypot(current.X-source.path[1].X, current.Y-source.path[1].Y) <= 0.3 {
		source.path = source.path[1:]
	}
	if len(source.path) <= 1 {
		source.control.clearMoveTarget()
		source.path = nil
		source.pathTarget = nil
		return &MoveTarget{X: current.X, Y: current.Y}
	}
	return &MoveTarget{X: source.path[1].X, Y: source.path[1].Y}
}

func decodeMove(encoded []byte) (MovePayload, error) {
	var payload MovePayload
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return MovePayload{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return MovePayload{}, fmt.Errorf("movement payload has trailing data")
	}
	if payload.X < -1 || payload.X > 1 || payload.Y < -1 || payload.Y > 1 {
		return MovePayload{}, fmt.Errorf("movement axes must be between -1 and 1")
	}
	if payload.Target != nil && (math.IsNaN(payload.Target.X) || math.IsNaN(payload.Target.Y) || math.IsInf(payload.Target.X, 0) || math.IsInf(payload.Target.Y, 0)) {
		return MovePayload{}, fmt.Errorf("movement target must be finite")
	}
	return payload, nil
}
