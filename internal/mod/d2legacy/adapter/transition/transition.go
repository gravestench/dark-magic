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
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

const CommandKind = "system.world.transition"
const triggerRadius = 2.0

type Payload struct {
	DestinationLevel int     `json:"destination_level"`
	SourceLevel      int     `json:"source_level"`
	SourceX          float64 `json:"source_x"`
	SourceY          float64 `json:"source_y"`
	ArrivalX         float64 `json:"arrival_x"`
	ArrivalY         float64 `json:"arrival_y"`
	WorldWidth       float64 `json:"world_width"`
	WorldHeight      float64 `json:"world_height"`
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

type Source struct {
	engine    interface{ World() *akara.World }
	player    string
	authority *Authority
	sequence  atomic.Uint64
}

func NewSource(engine interface{ World() *akara.World }, player string, authority *Authority) (*Source, error) {
	player = strings.TrimSpace(player)
	if engine == nil || player == "" || authority == nil {
		return nil, fmt.Errorf("transition: source requires engine, player, and authority")
	}
	return &Source{engine: engine, player: player, authority: authority}, nil
}

func (source *Source) Commands(tick uint64) []simulation.Command {
	controls, ok := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.player_control")
	if !ok {
		return nil
	}
	locations, ok := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.location")
	if !ok {
		return nil
	}
	positions, ok := akara.GetDynamicStore(source.engine.World(), "d2legacy.world.position")
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
		destinationEndpoint := source.authority.seam.Wilderness
		if destination == 1 {
			destinationEndpoint = source.authority.seam.Town
		}
		payload, _ := json.Marshal(Payload{DestinationLevel: destination, SourceLevel: int(level.(int64)),
			SourceX: endpoint.X, SourceY: endpoint.Y, ArrivalX: destinationEndpoint.ArrivalX,
			ArrivalY: destinationEndpoint.ArrivalY, WorldWidth: destinationEndpoint.Width, WorldHeight: destinationEndpoint.Height})
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
	if payload.SourceLevel != 1 && payload.SourceLevel != 2 || payload.SourceLevel == payload.DestinationLevel {
		return Payload{}, fmt.Errorf("transition: source and destination are invalid")
	}
	for _, value := range []float64{payload.SourceX, payload.SourceY, payload.ArrivalX, payload.ArrivalY, payload.WorldWidth, payload.WorldHeight} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return Payload{}, fmt.Errorf("transition: geometry must be finite")
		}
	}
	if payload.WorldWidth <= 0 || payload.WorldHeight <= 0 || payload.ArrivalX < 0 || payload.ArrivalY < 0 || payload.ArrivalX >= payload.WorldWidth || payload.ArrivalY >= payload.WorldHeight {
		return Payload{}, fmt.Errorf("transition: destination geometry is invalid")
	}
	return payload, nil
}
