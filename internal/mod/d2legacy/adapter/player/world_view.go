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
	WorldViewVersion        uint32 = 4
	WorldViewRadius                = 80.0
	MaxWorldViewEntities           = 256
	MaxWorldViewMissiles           = 512
	maxWorldIDBytes                = 128
	maxWorldKindBytes              = 32
	maxWorldLabelBytes             = 256
	maxWorldAssetBytes             = 512
	maxWorldComponentsBytes        = 4096
)

var ErrWorldView = errors.New("player world view: invalid public projection")

type WorldView struct {
	Version   uint32         `json:"version"`
	Tick      uint64         `json:"tick"`
	Origin    HUDPosition    `json:"origin"`
	Entities  []WorldEntity  `json:"entities"`
	Missiles  []WorldMissile `json:"missiles"`
	Truncated bool           `json:"truncated"`
}

// WorldMissile is the bounded visual subset of an authoritative projectile or
// aftermath entity. It deliberately omits damage, collision, ownership, hit
// locks, and lifetime policy: a connected client can render these values but
// cannot reconstruct or advance gameplay from them.
type WorldMissile struct {
	ID               string      `json:"id"`
	Kind             string      `json:"kind"`
	MissileID        string      `json:"missile_id"`
	Position         HUDPosition `json:"position"`
	Velocity         HUDPosition `json:"velocity"`
	Act              int64       `json:"act"`
	LevelID          int64       `json:"level_id"`
	DCC              string      `json:"dcc"`
	Palette          string      `json:"palette,omitempty"`
	LogicalDirection int64       `json:"logical_direction"`
	Directions       int64       `json:"directions"`
	FramesPerSecond  int64       `json:"frames_per_second"`
	Loop             bool        `json:"loop"`
	TransparencyMode int64       `json:"transparency_mode"`
	OffsetX          float64     `json:"offset_x"`
	OffsetY          float64     `json:"offset_y"`
	OffsetZ          float64     `json:"offset_z"`
	distance2        float64
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
	SpawnID                string      `json:"spawn_id,omitempty"`
	DefinitionID           string      `json:"definition_id,omitempty"`
	Token                  string      `json:"token,omitempty"`
	Mode                   string      `json:"mode,omitempty"`
	WeaponClass            string      `json:"weapon_class,omitempty"`
	Components             string      `json:"components,omitempty"`
	DeathSound             string      `json:"death_sound,omitempty"`
	OverlayHeight          int64       `json:"overlay_height,omitempty"`
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
	selectables, _ := findComponent(snapshot, "d2legacy.world.selectable")
	inactive, _ := findComponent(snapshot, "d2legacy.world.inactive")
	monsters, _ := findComponent(snapshot, "d2legacy.monster.stats")
	monsterIdentities, _ := findComponent(snapshot, "d2legacy.monster.identity")
	monsterAppearances, _ := findComponent(snapshot, "d2legacy.monster.appearance")
	monsterDeaths, _ := findComponent(snapshot, "d2legacy.monster.death")
	monsterAI, _ := findComponent(snapshot, "d2legacy.monster.ai")
	velocities, _ := findComponent(snapshot, "d2legacy.world.velocity")
	players, _ := findComponent(snapshot, "d2legacy.player.identity")
	appearances, _ := findComponent(snapshot, "d2legacy.player.appearance")
	animations, _ := findComponent(snapshot, "d2legacy.player.animation")
	movementStats, _ := findComponent(snapshot, "d2legacy.player.movement_stats")
	facings, _ := findComponent(snapshot, "d2legacy.world.facing")
	view := WorldView{Version: WorldViewVersion, Tick: checkpoint.Tick, Origin: origin, Entities: []WorldEntity{}, Missiles: []WorldMissile{}}
	seen := make(map[string]struct{})
	decorateMonster := func(entity *WorldEntity, source uint64) {
		if identity, found := findInstance(monsterIdentities, source); found {
			entity.SpawnID = stringField(identity, "spawn_id")
			entity.DefinitionID = stringField(identity, "definition_id")
		}
		if appearance, found := findInstance(monsterAppearances, source); found {
			if nameKey := stringField(appearance, "name_key"); nameKey != "" {
				entity.Label = nameKey
			}
			entity.Token = stringField(appearance, "token")
			entity.Mode = stringField(appearance, "mode")
			entity.WeaponClass = stringField(appearance, "weapon_class")
			entity.Components = stringField(appearance, "components")
			entity.DeathSound = stringField(appearance, "death_sound")
			entity.OverlayHeight = intField(appearance, "overlay_height")
			brain, _ := findInstance(monsterAI, source)
			velocity, _ := findInstance(velocities, source)
			entity.Mode = monsterPresentationMode(
				entity.Mode,
				stringField(brain, "state"),
				floatField(velocity, "x"),
				floatField(velocity, "y"),
			)
		}
	}
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
		decorateMonster(&entity, instance.Entity)
		if appearance, found := findInstance(appearances, instance.Entity); found {
			entity.Token = stringField(appearance, "token")
			entity.Mode = stringField(appearance, "mode")
		}
		if facing, ok := findInstance(facings, instance.Entity); ok {
			entity.Direction = intField(facing, "direction")
		}
		if identity, found := findInstance(players, instance.Entity); found {
			entity.Class = stringField(identity, "class")
			if animation, ok := findInstance(animations, instance.Entity); ok {
				entity.Mode = stringField(animation, "mode")
				entity.AnimationStartTick = uint64(max(0, intField(animation, "start_tick")))
			}
			if movement, ok := findInstance(movementStats, instance.Entity); ok {
				entity.VelocityPercent = intField(movement, "velocitypercent")
				entity.ItemFasterMoveVelocity = intField(movement, "item_fastermovevelocity")
			}
		}
		if err := validateWorldEntity(entity); err != nil {
			return nil, err
		}
		view.Entities = append(view.Entities, entity)
	}
	// Death removes selection and collision from authority, but the same entity
	// remains a visible corpse. Project only its presentation identity and
	// spatial facts; loot, XP, corpse usability, and attribution stay server-only.
	for _, instance := range monsterDeaths.Instances {
		if _, dormant := findInstance(inactive, instance.Entity); dormant {
			continue
		}
		identity, identified := findInstance(monsterIdentities, instance.Entity)
		position, positioned := findInstance(positions, instance.Entity)
		location, located := findInstance(locations, instance.Entity)
		if !identified || !positioned || !located {
			continue
		}
		spawnID := stringField(identity, "spawn_id")
		entity := WorldEntity{
			ID: "monster:" + spawnID, Kind: "corpse", Label: stringField(identity, "definition_id"),
			Position: HUDPosition{X: floatField(position, "x"), Y: floatField(position, "y")},
			Act:      intField(location, "act"), LevelID: intField(location, "level_id"),
		}
		if _, alreadyProjected := seen[entity.ID]; alreadyProjected {
			continue
		}
		if entity.Act != originAct || entity.LevelID != originLevel {
			continue
		}
		if stats, found := findInstance(monsters, instance.Entity); found {
			health, maximum := intField(stats, "health"), intField(stats, "max_health")
			entity.Health, entity.MaxHealth = &health, &maximum
		}
		decorateMonster(&entity, instance.Entity)
		if facing, found := findInstance(facings, instance.Entity); found {
			entity.Direction = intField(facing, "direction")
		}
		if err := validateWorldEntity(entity); err != nil {
			return nil, err
		}
		seen[entity.ID] = struct{}{}
		dx, dy := entity.Position.X-origin.X, entity.Position.Y-origin.Y
		entity.distance2 = dx*dx + dy*dy
		if entity.distance2 <= WorldViewRadius*WorldViewRadius {
			view.Entities = append(view.Entities, entity)
		}
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
	view.Missiles, found = projectWorldMissiles(snapshot, origin, originAct, originLevel)
	if !found {
		return nil, ErrWorldView
	}
	if len(view.Missiles) > MaxWorldViewMissiles {
		view.Missiles, view.Truncated = view.Missiles[:MaxWorldViewMissiles], true
	}
	return json.Marshal(view)
}

// monsterPresentationMode collapses server-only AI and velocity facts into
// the two-byte animation vocabulary already carried by the public transform
// stream. Its death -> attack -> movement -> authored precedence matches the
// offline monster snapshot adapter, so the client learns what to draw without
// learning why the monster chose it.
func monsterPresentationMode(authoredMode, aiState string, velocityX, velocityY float64) string {
	if authoredMode == "DT" {
		return authoredMode
	}
	if aiState == "attack" {
		return "A1"
	}
	if velocityX != 0 || velocityY != 0 {
		return "WL"
	}
	return authoredMode
}

func validateWorldEntity(entity WorldEntity) error {
	if entity.ID == "" || entity.Kind == "" || len(entity.ID) > maxWorldIDBytes || len(entity.Kind) > maxWorldKindBytes ||
		len(entity.Label) > maxWorldLabelBytes || len(entity.Owner) > maxWorldIDBytes || len(entity.SpawnID) > maxWorldIDBytes ||
		len(entity.DefinitionID) > maxWorldIDBytes || len(entity.Token) > maxWorldAssetBytes || len(entity.Mode) > maxWorldKindBytes ||
		len(entity.WeaponClass) > maxWorldKindBytes || len(entity.Components) > maxWorldComponentsBytes || len(entity.DeathSound) > maxWorldAssetBytes ||
		entity.OverlayHeight < 0 || entity.OverlayHeight > 4 || math.IsNaN(entity.Position.X) || math.IsNaN(entity.Position.Y) ||
		math.IsInf(entity.Position.X, 0) || math.IsInf(entity.Position.Y, 0) || entity.Radius < 0 || math.IsNaN(entity.Radius) || math.IsInf(entity.Radius, 0) {
		return ErrWorldView
	}
	if entity.Kind == "corpse" && (entity.SpawnID == "" || entity.Token == "" || entity.Mode != "DT") {
		return ErrWorldView
	}
	return nil
}
