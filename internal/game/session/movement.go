package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/inputstate"
)

const MoveCommand = "player.move"

type MovePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
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
				if err := velocity.Set("x", float64(payload.X)*10); err != nil {
					return err
				}
				if err := velocity.Set("y", float64(payload.Y)*10); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

// MovementSource turns the latest native input snapshot into one replayable
// command per active gameplay tick.
type MovementSource struct {
	engine   *gameecs.Engine
	input    *inputstate.Store
	player   string
	focusID  string
	sequence uint64
}

func NewMovementSource(engine *gameecs.Engine, input *inputstate.Store, player, focusID string) (*MovementSource, error) {
	player = strings.TrimSpace(player)
	focusID = strings.TrimSpace(focusID)
	if engine == nil || input == nil || player == "" || focusID == "" {
		return nil, fmt.Errorf("game session: movement source requires engine, input, player, and focus owner")
	}
	return &MovementSource{engine: engine, input: input, player: player, focusID: focusID}, nil
}

func (source *MovementSource) Commands(tick uint64) []simulation.Command {
	controls, present := akara.GetDynamicStore(source.engine.World(), "dm.world.player_control")
	if !present || controls.Len() == 0 {
		return nil
	}
	x, y := 0, 0
	owner := source.input.Owner()
	if owner.Domain == inputstate.FocusScene && (owner.ID == source.focusID || source.input.Gameplay()) {
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
	source.sequence++
	payload, _ := json.Marshal(MovePayload{X: x, Y: y})
	return []simulation.Command{{Tick: tick, Player: source.player, Authority: simulation.AuthorityPlayer, Sequence: source.sequence, Kind: MoveCommand, Payload: payload}}
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
