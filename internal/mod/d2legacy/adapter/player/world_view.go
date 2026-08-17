package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

const (
	WorldViewVersion     uint32 = 2
	WorldViewRadius             = 80.0
	MaxWorldViewEntities        = 256
	maxWorldIDBytes             = 128
	maxWorldKindBytes           = 32
	maxWorldLabelBytes          = 256
)

var ErrWorldView = errors.New("player world view: invalid public projection")

type WorldView struct {
	Version   uint32        `json:"version"`
	Tick      uint64        `json:"tick"`
	Origin    HUDPosition   `json:"origin"`
	Entities  []WorldEntity `json:"entities"`
	Truncated bool          `json:"truncated"`
}

type WorldEntity struct {
	ID                     string      `json:"id"`
	Kind                   string      `json:"kind"`
	Label                  string      `json:"label,omitempty"`
	Owner                  string      `json:"owner,omitempty"`
	Position               HUDPosition `json:"position"`
	Radius                 float64     `json:"radius"`
	Priority               int64       `json:"priority"`
	Health                 *int64      `json:"health,omitempty"`
	MaxHealth              *int64      `json:"max_health,omitempty"`
	Class                  string      `json:"class,omitempty"`
	Token                  string      `json:"token,omitempty"`
	Mode                   string      `json:"mode,omitempty"`
	Direction              int64       `json:"direction,omitempty"`
	AnimationStartTick     uint64      `json:"animation_start_tick,omitempty"`
	VelocityPercent        int64       `json:"velocitypercent,omitempty"`
	ItemFasterMoveVelocity int64       `json:"item_fastermovevelocity,omitempty"`
	Act                    int64       `json:"act,omitempty"`
	LevelID                int64       `json:"level_id,omitempty"`
	distance2              float64
}

// ProjectWorldView exposes only nearby entities carrying the mod's explicit
// public selectable contract. Raw ECS identity and non-allowlisted components
// never enter the result.
func ProjectWorldView(playerID string, checkpoint simulation.Checkpoint) (json.RawMessage, error) {
	if checkpoint.Snapshot == nil || strings.TrimSpace(playerID) == "" {
		return nil, ErrWorldView
	}
	snapshot := *checkpoint.Snapshot
	identities, found := findComponent(snapshot, "d2legacy.player.identity")
	if !found {
		return nil, ErrWorldView
	}
	playerEntity, _, found := findString(identities, "player", playerID)
	if !found {
		return nil, ErrWorldView
	}
	positions, found := findComponent(snapshot, "d2legacy.world.position")
	if !found {
		return nil, ErrWorldView
	}
	originFields, found := findInstance(positions, playerEntity)
	if !found {
		return nil, ErrWorldView
	}
	origin := HUDPosition{X: floatField(originFields, "x"), Y: floatField(originFields, "y")}
	locations, found := findComponent(snapshot, "d2legacy.world.location")
	if !found {
		return nil, ErrWorldView
	}
	originLocation, found := findInstance(locations, playerEntity)
	if !found {
		return nil, ErrWorldView
	}
	originAct, originLevel := intField(originLocation, "act"), intField(originLocation, "level_id")
	selectables, found := findComponent(snapshot, "d2legacy.world.selectable")
	if !found {
		return json.Marshal(WorldView{Version: WorldViewVersion, Tick: checkpoint.Tick, Origin: origin, Entities: []WorldEntity{}})
	}
	inactive, _ := findComponent(snapshot, "d2legacy.world.inactive")
	monsters, _ := findComponent(snapshot, "d2legacy.monster.stats")
	players, _ := findComponent(snapshot, "d2legacy.player.identity")
	appearances, _ := findComponent(snapshot, "d2legacy.player.appearance")
	animations, _ := findComponent(snapshot, "d2legacy.player.animation")
	movementStats, _ := findComponent(snapshot, "d2legacy.player.movement_stats")
	facings, _ := findComponent(snapshot, "d2legacy.world.facing")
	view := WorldView{Version: WorldViewVersion, Tick: checkpoint.Tick, Origin: origin, Entities: []WorldEntity{}}
	seen := make(map[string]struct{})
	for _, instance := range selectables.Instances {
		if instance.Entity == playerEntity {
			continue
		}
		if _, dormant := findInstance(inactive, instance.Entity); dormant {
			continue
		}
		public, ok := findInstance(selectables, instance.Entity)
		position, positioned := findInstance(positions, instance.Entity)
		if !ok || !positioned {
			continue
		}
		entity := WorldEntity{ID: stringField(public, "id"), Kind: stringField(public, "kind"), Label: stringField(public, "label"), Owner: stringField(public, "owner"), Position: HUDPosition{X: floatField(position, "x"), Y: floatField(position, "y")}, Radius: floatField(public, "radius"), Priority: intField(public, "priority")}
		location, located := findInstance(locations, instance.Entity)
		if !located {
			continue
		}
		entity.Act, entity.LevelID = intField(location, "act"), intField(location, "level_id")
		if entity.Act != originAct || entity.LevelID != originLevel {
			continue
		}
		if err := validateWorldEntity(entity); err != nil {
			return nil, err
		}
		if _, duplicate := seen[entity.ID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate public ID %q", ErrWorldView, entity.ID)
		}
		seen[entity.ID] = struct{}{}
		dx, dy := entity.Position.X-origin.X, entity.Position.Y-origin.Y
		entity.distance2 = dx*dx + dy*dy
		if entity.distance2 > WorldViewRadius*WorldViewRadius {
			continue
		}
		if stats, found := findInstance(monsters, instance.Entity); found {
			health, maximum := intField(stats, "health"), intField(stats, "max_health")
			entity.Health, entity.MaxHealth = &health, &maximum
		}
		if appearance, found := findInstance(appearances, instance.Entity); found {
			entity.Token = stringField(appearance, "token")
			entity.Mode = stringField(appearance, "mode")
		}
		if identity, found := findInstance(players, instance.Entity); found {
			entity.Class = stringField(identity, "class")
			if animation, ok := findInstance(animations, instance.Entity); ok {
				entity.Mode = stringField(animation, "mode")
				entity.AnimationStartTick = uint64(max(0, intField(animation, "start_tick")))
			}
			if facing, ok := findInstance(facings, instance.Entity); ok {
				entity.Direction = intField(facing, "direction")
			}
			if movement, ok := findInstance(movementStats, instance.Entity); ok {
				entity.VelocityPercent = intField(movement, "velocitypercent")
				entity.ItemFasterMoveVelocity = intField(movement, "item_fastermovevelocity")
			}
		}
		view.Entities = append(view.Entities, entity)
	}
	sort.Slice(view.Entities, func(i, j int) bool {
		if view.Entities[i].distance2 != view.Entities[j].distance2 {
			return view.Entities[i].distance2 < view.Entities[j].distance2
		}
		return view.Entities[i].ID < view.Entities[j].ID
	})
	if len(view.Entities) > MaxWorldViewEntities {
		view.Entities, view.Truncated = view.Entities[:MaxWorldViewEntities], true
	}
	return json.Marshal(view)
}

func validateWorldEntity(entity WorldEntity) error {
	if entity.ID == "" || entity.Kind == "" || len(entity.ID) > maxWorldIDBytes || len(entity.Kind) > maxWorldKindBytes || len(entity.Label) > maxWorldLabelBytes || len(entity.Owner) > maxWorldIDBytes || math.IsNaN(entity.Position.X) || math.IsNaN(entity.Position.Y) || math.IsInf(entity.Position.X, 0) || math.IsInf(entity.Position.Y, 0) || entity.Radius < 0 || math.IsNaN(entity.Radius) || math.IsInf(entity.Radius, 0) {
		return ErrWorldView
	}
	return nil
}
