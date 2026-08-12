// Package transition owns trusted movement between authoritative world zones.
package transition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"
	"sync/atomic"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const CommandKind = "system.world.transition"
const triggerRadius = 2.0

type Payload struct {
	DestinationLevel int `json:"destination_level"`
}
type Authority struct {
	seam     gameworld.Seam
	observer func(int)
}

func NewAuthority(seam gameworld.Seam) (*Authority, error) {
	if seam.Town.LevelID != 1 || seam.Wilderness.LevelID != 2 {
		return nil, fmt.Errorf("transition: Act I seam is incomplete")
	}
	return &Authority{seam: seam}, nil
}

// SetObserver installs an adapter notification after authoritative ECS state
// has committed. The callback may switch caches/render inputs; it decides no
// gameplay fact and therefore cannot veto the transition.
func (authority *Authority) SetObserver(observer func(int)) { authority.observer = observer }

func Register(session *gamesession.Session, authority *Authority) error {
	if session == nil || authority == nil {
		return fmt.Errorf("transition: session and authority are required")
	}
	return session.Register(CommandKind, gamesession.CommandHandler{
		Allowed:  []simulation.Authority{simulation.AuthoritySystem, simulation.AuthorityAdmin},
		Validate: func(command simulation.Command) error { _, err := decode(command.Payload); return err },
		Apply:    authority.apply,
	})
}

func (authority *Authority) apply(engine *gameecs.Engine, command simulation.Command) error {
	payload, err := decode(command.Payload)
	if err != nil {
		return err
	}
	controls, ok := akara.GetDynamicStore(engine.World(), "d2.world.player_control")
	if !ok {
		return fmt.Errorf("transition: player control is unavailable")
	}
	locations, ok := akara.GetDynamicStore(engine.World(), "d2.world.location")
	if !ok {
		return fmt.Errorf("transition: player location is unavailable")
	}
	positions, ok := akara.GetDynamicStore(engine.World(), "d2.world.position")
	if !ok {
		return fmt.Errorf("transition: player position is unavailable")
	}
	bounds, _ := akara.GetDynamicStore(engine.World(), "d2.world.bounds")
	velocities, _ := akara.GetDynamicStore(engine.World(), "d2.world.velocity")
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		owner, _ := control.Get("player")
		if owner != command.Player {
			continue
		}
		location, _ := locations.Get(entity)
		current, _ := location.Get("level_id")
		source, destination, err := authority.endpoints(current.(int64), payload.DestinationLevel)
		if err != nil {
			return err
		}
		position, _ := positions.Get(entity)
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		if math.Hypot(x.(float64)-source.X, y.(float64)-source.Y) > triggerRadius {
			return fmt.Errorf("transition: player is outside the authored seam")
		}
		if err := location.Set("level_id", int64(destination.LevelID)); err != nil {
			return err
		}
		if err := position.Set("x", destination.ArrivalX); err != nil {
			return err
		}
		if err := position.Set("y", destination.ArrivalY); err != nil {
			return err
		}
		if bounds != nil {
			if value, found := bounds.Get(entity); found {
				if err := value.Set("width", destination.Width); err != nil {
					return err
				}
				if err := value.Set("height", destination.Height); err != nil {
					return err
				}
			}
		}
		if velocities != nil {
			if value, found := velocities.Get(entity); found {
				if err := value.Set("x", float64(0)); err != nil {
					return err
				}
				if err := value.Set("y", float64(0)); err != nil {
					return err
				}
			}
		}
		if authority.observer != nil {
			authority.observer(destination.LevelID)
		}
		return nil
	}
	return fmt.Errorf("transition: player %q is unavailable", command.Player)
}

func (authority *Authority) endpoints(current int64, destination int) (gameworld.SeamEndpoint, gameworld.SeamEndpoint, error) {
	if current == 1 && destination == 2 {
		return authority.seam.Town, authority.seam.Wilderness, nil
	}
	if current == 2 && destination == 1 {
		return authority.seam.Wilderness, authority.seam.Town, nil
	}
	return gameworld.SeamEndpoint{}, gameworld.SeamEndpoint{}, fmt.Errorf("transition: invalid level change %d -> %d", current, destination)
}

type Source struct {
	engine    *gameecs.Engine
	player    string
	authority *Authority
	sequence  atomic.Uint64
}

func NewSource(engine *gameecs.Engine, player string, authority *Authority) (*Source, error) {
	player = strings.TrimSpace(player)
	if engine == nil || player == "" || authority == nil {
		return nil, fmt.Errorf("transition: source requires engine, player, and authority")
	}
	return &Source{engine: engine, player: player, authority: authority}, nil
}

func (source *Source) Commands(tick uint64) []simulation.Command {
	controls, ok := akara.GetDynamicStore(source.engine.World(), "d2.world.player_control")
	if !ok {
		return nil
	}
	locations, ok := akara.GetDynamicStore(source.engine.World(), "d2.world.location")
	if !ok {
		return nil
	}
	positions, ok := akara.GetDynamicStore(source.engine.World(), "d2.world.position")
	if !ok {
		return nil
	}
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		owner, _ := control.Get("player")
		if owner != source.player {
			continue
		}
		location, _ := locations.Get(entity)
		level, _ := location.Get("level_id")
		position, _ := positions.Get(entity)
		x, _ := position.Get("x")
		y, _ := position.Get("y")
		endpoint, destination := source.authority.seam.Town, 2
		if level.(int64) == 2 {
			endpoint, destination = source.authority.seam.Wilderness, 1
		}
		if math.Hypot(x.(float64)-endpoint.X, y.(float64)-endpoint.Y) > triggerRadius {
			return nil
		}
		payload, _ := json.Marshal(Payload{DestinationLevel: destination})
		return []simulation.Command{{Tick: tick, Player: source.player, Authority: simulation.AuthoritySystem, Sequence: source.sequence.Add(1), Kind: CommandKind, Payload: payload}}
	}
	return nil
}

func decode(encoded []byte) (Payload, error) {
	var payload Payload
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return Payload{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Payload{}, fmt.Errorf("transition: trailing payload data")
	}
	if payload.DestinationLevel != 1 && payload.DestinationLevel != 2 {
		return Payload{}, fmt.Errorf("transition: destination must be town or Blood Moor")
	}
	return payload, nil
}
