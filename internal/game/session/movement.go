package session

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

type MoveTarget struct{ X, Y, StopRadius float64 }

// MovementController is the thread-safe local intent mailbox shared by Lua UI
// and the fixed-tick movement command source. It never mutates ECS state.
type MovementController struct {
	running   atomic.Bool
	sequence  atomic.Uint64
	mu        sync.Mutex
	skills    map[string]int64
	skillUses []UseSkillPayload
	target    *MoveTarget
}

func (controller *MovementController) SetRunning(running bool) { controller.running.Store(running) }
func (controller *MovementController) Running() bool           { return controller.running.Load() }
func (controller *MovementController) HasMoveTarget() bool {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.target != nil
}
func (controller *MovementController) nextSequence() uint64 { return controller.sequence.Add(1) }

func (controller *MovementController) AssignSkill(slot string, skillID int64) error {
	if slot != "left" && slot != "right" || skillID < 0 {
		return fmt.Errorf("game session: skill assignment requires left/right slot and non-negative skill ID")
	}
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.skills == nil {
		controller.skills = make(map[string]int64)
	}
	controller.skills[slot] = skillID
	return nil
}

func (controller *MovementController) UseSkill(side string, x, y float64, targetID string) error {
	side = strings.ToLower(strings.TrimSpace(side))
	if side != "left" && side != "right" || math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
		return fmt.Errorf("game session: skill use requires left/right side and finite target")
	}
	controller.mu.Lock()
	controller.skillUses = append(controller.skillUses, UseSkillPayload{Side: side, TargetX: x, TargetY: y, TargetID: strings.TrimSpace(targetID)})
	controller.mu.Unlock()
	return nil
}

func (controller *MovementController) drainSkillUses() []UseSkillPayload {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	result := controller.skillUses
	controller.skillUses = nil
	return result
}

func (controller *MovementController) SetMoveTarget(x, y float64) error {
	return controller.SetMoveTargetWithRadius(x, y, 0)
}

func (controller *MovementController) SetMoveTargetWithRadius(x, y, stopRadius float64) error {
	if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
		return fmt.Errorf("game session: movement target must be finite")
	}
	if stopRadius < 0 || math.IsNaN(stopRadius) || math.IsInf(stopRadius, 0) {
		return fmt.Errorf("game session: movement stop radius must be non-negative and finite")
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

func (controller *MovementController) drainSkills() map[string]int64 {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	result := controller.skills
	controller.skills = nil
	return result
}

// RegisterMovement installs the authoritative adapter from normalized movement
// intent to Lua-defined world velocity components.
func RegisterMovement(session *Session) error {
	return session.Register(MoveCommand, CommandHandler{
		Validate: func(command simulation.Command) error {
			_, err := decodeMove(command.Payload)
			return err
		},
		Apply: func(engine *gameecs.Engine, command simulation.Command) error {
			payload, err := decodeMove(command.Payload)
			if err != nil {
				return err
			}
			controls, present := akara.GetDynamicStore(engine.World(), "dm.world.player_control")
			if !present {
				return nil
			}
			velocities, present := akara.GetDynamicStore(engine.World(), "dm.world.velocity")
			if !present {
				return nil
			}
			modes, modesPresent := akara.GetDynamicStore(engine.World(), "dm.player.movement_mode")
			animations, animationsPresent := akara.GetDynamicStore(engine.World(), "dm.player.animation")
			for _, entity := range controls.Entities() {
				control, found := controls.Get(entity)
				if !found {
					continue
				}
				player, err := control.Get("player")
				if err != nil {
					return err
				}
				if player != command.Player {
					continue
				}
				velocity, found := velocities.Get(entity)
				if !found {
					continue
				}
				speed := 10.0
				if payload.Running {
					speed = 15
				}
				// Two full-speed axes would make diagonal movement sqrt(2) times
				// faster. Normalize the vector before simulation sees it.
				x, y := float64(payload.X), float64(payload.Y)
				if payload.Target != nil {
					positions, positionsPresent := akara.GetDynamicStore(engine.World(), "dm.world.position")
					if positionsPresent {
						if position, found := positions.Get(entity); found {
							currentX, _ := position.Get("x")
							currentY, _ := position.Get("y")
							x, y = payload.Target.X-currentX.(float64), payload.Target.Y-currentY.(float64)
							distance := math.Hypot(x, y)
							if distance <= 0.2 {
								x, y = 0, 0
							} else {
								x, y = x/distance, y/distance
							}
						}
					}
				} else if payload.X != 0 && payload.Y != 0 {
					const inverseSquareRootTwo = 0.7071067811865476
					x *= inverseSquareRootTwo
					y *= inverseSquareRootTwo
				}
				if err := velocity.Set("x", x*speed); err != nil {
					return err
				}
				if err := velocity.Set("y", y*speed); err != nil {
					return err
				}
				if modesPresent {
					if mode, found := modes.Get(entity); found {
						if err := mode.Set("running", payload.Running); err != nil {
							return err
						}
					}
				}
				if animationsPresent {
					if animation, found := animations.Get(entity); found {
						moving := x != 0 || y != 0
						mode := "NU"
						if moving && payload.Running {
							mode = "RN"
						} else if moving {
							mode = "WL"
						}
						if err := animation.Set("mode", mode); err != nil {
							return err
						}
						if moving {
							directionX, directionY := sign(x), sign(y)
							if err := animation.Set("direction", movementDirection(directionX, directionY)); err != nil {
								return err
							}
						}
					}
				}
			}
			return nil
		},
	})
}

// movementDirection converts the eight normalized world-space input vectors to
// a readable logical direction order. The presentation adapter converts this
// authoritative value to each legacy asset's encoded 8/16-direction order.
// Stopping does not call this function, so an idle player keeps looking the way
// they last moved.
func movementDirection(x, y int) int64 {
	directions := map[[2]int]int64{
		{0, 1}: 0, {-1, 0}: 1, {0, -1}: 2, {1, 0}: 3,
		{1, 1}: 4, {-1, 1}: 5, {-1, -1}: 6, {1, -1}: 7,
	}
	return directions[[2]int{x, y}]
}

// MovementSource turns the latest native input snapshot into one replayable
// command per active gameplay tick.
type MovementSource struct {
	engine     *gameecs.Engine
	input      *inputstate.Store
	player     string
	focusID    string
	control    *MovementController
	navigation *gameworld.Map
	path       []gameworld.Point
	pathTarget *MoveTarget
}

func (source *MovementSource) SetNavigation(world *gameworld.Map) {
	source.navigation = world
	source.path = nil
	source.pathTarget = nil
}

func NewMovementSource(engine *gameecs.Engine, input *inputstate.Store, player, focusID string, controllers ...*MovementController) (*MovementSource, error) {
	player = strings.TrimSpace(player)
	focusID = strings.TrimSpace(focusID)
	if engine == nil || input == nil || player == "" || focusID == "" {
		return nil, fmt.Errorf("game session: movement source requires engine, input, player, and focus owner")
	}
	control := &MovementController{}
	if len(controllers) > 0 && controllers[0] != nil {
		control = controllers[0]
	}
	return &MovementSource{engine: engine, input: input, player: player, focusID: focusID, control: control}, nil
}

func (source *MovementSource) Commands(tick uint64) []simulation.Command {
	controls, present := akara.GetDynamicStore(source.engine.World(), "dm.world.player_control")
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
	positions, ok := akara.GetDynamicStore(source.engine.World(), "dm.world.position")
	if !ok {
		return target
	}
	controls, ok := akara.GetDynamicStore(source.engine.World(), "dm.world.player_control")
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
		if colliders, present := akara.GetDynamicStore(source.engine.World(), "dm.world.collider"); present {
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
	changed := source.pathTarget == nil || source.pathTarget.X != target.X || source.pathTarget.Y != target.Y || source.pathTarget.StopRadius != target.StopRadius
	if changed {
		path, err := source.navigation.FindPath(gameworld.PathRequest{Start: current, Goal: gameworld.Point{X: target.X, Y: target.Y}, Radius: radius, StopRadius: target.StopRadius})
		if err != nil {
			source.control.clearMoveTarget()
			source.path = nil
			source.pathTarget = nil
			return &MoveTarget{X: current.X, Y: current.Y}
		}
		source.path = path
		copyTarget := *target
		source.pathTarget = &copyTarget
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

func sign(value float64) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}
