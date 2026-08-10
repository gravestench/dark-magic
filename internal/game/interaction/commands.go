package interaction

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

const (
	OpenCommand  = "interaction.open"
	CloseCommand = "interaction.close"
)

type Payload struct {
	Owner  string `json:"owner,omitempty"`
	Target string `json:"target,omitempty"`
}

func RegisterCommands(session *gamesession.Session, authority *Authority) error {
	if session == nil || authority == nil {
		return fmt.Errorf("interaction: session and authority are required")
	}
	if err := session.RegisterStateParticipant(authority); err != nil {
		return fmt.Errorf("interaction: register state: %w", err)
	}
	for _, definition := range []struct {
		kind string
		open bool
	}{{OpenCommand, true}, {CloseCommand, false}} {
		kind, opens := definition.kind, definition.open
		if err := session.Register(kind, gamesession.CommandHandler{
			Validate: func(command simulation.Command) error { _, err := decode(command, opens); return err },
			Apply: func(engine *gameecs.Engine, command simulation.Command) error {
				payload, err := decode(command, opens)
				if err != nil {
					return err
				}
				owner := payload.Owner
				if owner == "" {
					owner = command.Player
				}
				if opens {
					return authority.openSpatial(engine, owner, payload.Target)
				}
				return authority.close(owner)
			},
			Allowed: []simulation.Authority{simulation.AuthorityPlayer, simulation.AuthorityAdmin},
		}); err != nil {
			return err
		}
	}
	return nil
}

func decode(command simulation.Command, opens bool) (Payload, error) {
	var payload Payload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return Payload{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Payload{}, fmt.Errorf("interaction: payload has trailing data")
	}
	payload.Owner, payload.Target = strings.TrimSpace(payload.Owner), strings.ToLower(strings.TrimSpace(payload.Target))
	if opens && payload.Target == "" {
		return Payload{}, fmt.Errorf("interaction: target is required")
	}
	if !opens && payload.Target != "" {
		return Payload{}, fmt.Errorf("interaction: close does not accept a target")
	}
	if command.Authority == simulation.AuthorityPlayer && payload.Owner != "" && payload.Owner != command.Player {
		return Payload{}, fmt.Errorf("interaction: player cannot change another owner's context")
	}
	return payload, nil
}

func Command(kind string, payload Payload, actor string, sequence, tick uint64, authority simulation.Authority) (simulation.Command, error) {
	if kind != OpenCommand && kind != CloseCommand {
		return simulation.Command{}, fmt.Errorf("interaction: unsupported command %q", kind)
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return simulation.Command{}, err
	}
	return simulation.Command{Tick: tick, Player: actor, Authority: authority, Sequence: sequence, Kind: kind, Payload: encoded}, nil
}
