package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const MoveCommand = "item.move"

// MovePayload describes intent, not trusted results. In particular, Displaced
// is computed by authority during a held-item swap and never supplied here.
type MovePayload struct {
	Owner       string    `json:"owner,omitempty"`
	ItemID      string    `json:"item_id"`
	Destination Placement `json:"destination"`
	PlaceHeld   bool      `json:"place_held,omitempty"`
}

func RegisterCommands(session *gamesession.Session, authority *Authority) error {
	if session == nil || authority == nil {
		return fmt.Errorf("item: session and authority are required")
	}
	return session.Register(MoveCommand, gamesession.CommandHandler{
		Validate: validateMoveCommand,
		Apply: func(_ *gameecs.Engine, command simulation.Command) error {
			payload, err := decodeMove(command.Payload)
			if err != nil {
				return err
			}
			owner := payload.Owner
			if owner == "" {
				owner = command.Player
			}
			_, err = authority.move(owner, payload.ItemID, payload.Destination, payload.PlaceHeld)
			return err
		},
		Allowed: []simulation.Authority{simulation.AuthorityPlayer, simulation.AuthorityAdmin},
	})
}

func validateMoveCommand(command simulation.Command) error {
	payload, err := decodeMove(command.Payload)
	if err != nil {
		return err
	}
	if command.Authority == simulation.AuthorityPlayer && payload.Owner != "" && payload.Owner != command.Player {
		return fmt.Errorf("item: player cannot move another owner's items")
	}
	if payload.PlaceHeld && !isHeldDestination(payload.Destination.Container) {
		return fmt.Errorf("item: held placement requires a grid, equipment, hireling, or belt destination")
	}
	return nil
}

func isHeldDestination(container Container) bool {
	return isGrid(container) || container == ContainerEquipment || container == ContainerHireling || container == ContainerBelt
}

func decodeMove(encoded []byte) (MovePayload, error) {
	var payload MovePayload
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return MovePayload{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return MovePayload{}, fmt.Errorf("item: move payload has trailing data")
	}
	payload.Owner = strings.TrimSpace(payload.Owner)
	payload.ItemID = strings.TrimSpace(payload.ItemID)
	if payload.ItemID == "" {
		return MovePayload{}, fmt.Errorf("item: item identity is required")
	}
	return payload, nil
}

func Command(payload MovePayload, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: MoveCommand, Payload: encoded}, nil
}
