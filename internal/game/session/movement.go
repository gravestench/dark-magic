package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/inputstate"
)

const MoveCommand = "player.move"

type MovePayload struct {
	X       int  `json:"x"`
	Y       int  `json:"y"`
	Running bool `json:"running"`
}

// MovementController is the thread-safe local intent mailbox shared by Lua UI
// and the fixed-tick movement command source. It never mutates ECS state.
type MovementController struct {
	running  atomic.Bool
	sequence atomic.Uint64
	mu       sync.Mutex
	skills   map[string]int64
}

func (controller *MovementController) SetRunning(running bool) { controller.running.Store(running) }
func (controller *MovementController) Running() bool           { return controller.running.Load() }
func (controller *MovementController) nextSequence() uint64    { return controller.sequence.Add(1) }

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
				if err := velocity.Set("x", float64(payload.X)*speed); err != nil {
					return err
				}
				if err := velocity.Set("y", float64(payload.Y)*speed); err != nil {
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
						moving := payload.X != 0 || payload.Y != 0
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
							if err := animation.Set("direction", movementDirection(payload.X, payload.Y)); err != nil {
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
	engine  *gameecs.Engine
	input   *inputstate.Store
	player  string
	focusID string
	control *MovementController
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
	}
	payload, _ := json.Marshal(MovePayload{X: x, Y: y, Running: source.control.Running()})
	return []simulation.Command{{Tick: tick, Player: source.player, Authority: simulation.AuthorityPlayer, Sequence: source.control.nextSequence(), Kind: MoveCommand, Payload: payload}}
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
	return payload, nil
}
