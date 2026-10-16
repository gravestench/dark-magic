package player

import (
	"fmt"
	"math"
	"sort"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
)

const maxMissileAssetBytes = 512

// Generated projectile IDs append cast, tick and instance suffixes to source identities.
const maxMissileSoundOwnerBytes = 512

// projectWorldMissiles joins live authority with its spatial components and
// copies only presentation-safe fields. Returning false means the checkpoint
// contained a malformed component shape, which must fail the entire client
// projection rather than silently produce a different world for one player.
func projectWorldMissiles(
	snapshot gameecs.Snapshot,
	origin HUDPosition,
	act, levelID int64,
) ([]WorldMissile, bool) {
	positions, positioned := findComponent(snapshot, "d2legacy.world.position")

	locations, located := findComponent(snapshot, "d2legacy.world.location")
	if !positioned || !located {
		return nil, false
	}

	positionByEntity, ok := indexEventComponent(positions)
	if !ok {
		return nil, false
	}

	locationByEntity, ok := indexEventComponent(locations)
	if !ok {
		return nil, false
	}

	ownerByEntity, ok := projectMissileSoundOwners(snapshot)
	if !ok {
		return nil, false
	}

	result := []WorldMissile{}
	seen := make(map[string]struct{})

	for _, source := range []struct {
		component string
		kind      string
		logical   bool
	}{
		{component: "d2legacy.missile.projectile", kind: "projectile"},
		{component: "d2legacy.missile.effect", kind: "effect", logical: true},
	} {
		component, found := findComponent(snapshot, source.component)
		if !found {
			continue
		}

		for _, instance := range component.Instances {
			fields, valid := eventInstanceFields(component, instance)
			position, hasPosition := positionByEntity[instance.Entity]

			location, hasLocation := locationByEntity[instance.Entity]
			if !valid || !hasPosition || !hasLocation {
				return nil, false
			}

			entryAct, entryLevel := intField(location, "act"), intField(location, "level_id")
			if entryAct != act || entryLevel != levelID {
				continue
			}

			x, y := floatField(position, "x"), floatField(position, "y")

			dx, dy := x-origin.X, y-origin.Y
			if dx*dx+dy*dy > WorldViewRadius*WorldViewRadius {
				continue
			}
			// Authority can own invisible helper missiles. They remain gameplay
			// facts and do not consume the client's bounded visual collection.
			dcc := stringField(fields, "dcc")
			if dcc == "" {
				continue
			}

			logicalDirection := int64(-1)
			if source.logical {
				logicalDirection = intField(fields, "logical_direction")
			}

			entry := WorldMissile{
				SoundOwnerID:     ownerByEntity[instance.Entity],
				ID:               fmt.Sprintf("missile:%d", instance.Entity),
				Kind:             source.kind,
				MissileID:        stringField(fields, "missile_id"),
				Position:         HUDPosition{X: x, Y: y},
				Velocity:         HUDPosition{X: floatField(fields, "velocity_x"), Y: floatField(fields, "velocity_y")},
				Act:              entryAct,
				LevelID:          entryLevel,
				DCC:              dcc,
				Palette:          stringField(fields, "palette"),
				LogicalDirection: logicalDirection,
				Directions:       intField(fields, "directions"),
				FramesPerSecond:  intField(fields, "frames_per_second"),
				Loop:             boolField(fields, "loop"),
				TransparencyMode: intField(fields, "transparency_mode"),
				OffsetX:          floatField(fields, "offset_x"),
				OffsetY:          floatField(fields, "offset_y"),
				OffsetZ:          floatField(fields, "offset_z"),
				distance2:        dx*dx + dy*dy,
			}
			if _, duplicate := seen[entry.ID]; duplicate {
				return nil, false
			}

			if validateWorldMissile(entry) != nil {
				return nil, false
			}

			seen[entry.ID] = struct{}{}
			result = append(result, entry)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].distance2 != result[j].distance2 {
			return result[i].distance2 < result[j].distance2
		}

		return result[i].ID < result[j].ID
	})

	return result, true
}

// validateWorldMissile enforces the visual protocol's numeric and asset bounds.
// It deliberately does not validate gameplay fields because none are projected.
func validateWorldMissile(missile WorldMissile) error {
	validDirections := missile.Directions == 1 || missile.Directions == 4 || missile.Directions == 8 ||
		missile.Directions == 16 || missile.Directions == 32
	if !bounded(missile.SoundOwnerID, maxMissileSoundOwnerBytes) || !boundedRequired(missile.ID, maxWorldIDBytes) ||
		(missile.Kind != "projectile" && missile.Kind != "effect") ||
		!boundedRequired(missile.MissileID, maxWorldKindBytes) || !boundedRequired(missile.DCC, maxMissileAssetBytes) ||
		!bounded(missile.Palette, maxMissileAssetBytes) || missile.Act < 1 || missile.Act > maxEventAct ||
		missile.LevelID < 0 || missile.LevelID > maxEventLevelID || !validDirections ||
		missile.FramesPerSecond < 1 || missile.FramesPerSecond > 1000 ||
		missile.TransparencyMode < 0 || missile.TransparencyMode > 1 ||
		missile.LogicalDirection < -1 || missile.LogicalDirection >= missile.Directions ||
		!finiteView(missile.Position.X, missile.Position.Y, missile.Velocity.X, missile.Velocity.Y,
			missile.OffsetX, missile.OffsetY, missile.OffsetZ) ||
		math.Hypot(missile.Position.X, missile.Position.Y) > math.MaxFloat32 {
		return ErrClientView
	}

	if missile.Kind == "projectile" && missile.LogicalDirection != -1 ||
		missile.Kind == "effect" && missile.LogicalDirection < 0 {
		return ErrClientView
	}

	return nil
}

// projectMissileSoundOwners permits absent sidecars but rejects unreadable present schemas or values.
func projectMissileSoundOwners(snapshot gameecs.Snapshot) (map[uint64]string, bool) {
	component, found := findComponent(snapshot, "d2legacy.presentation.missile_sound_owner")
	if !found {
		return nil, true
	}
	if component.Version != 1 || len(component.Fields) != 1 || component.Fields[0].Name != "projectile_id" || component.Fields[0].Kind != akara.FieldString {
		return nil, false
	}
	fields, ok := indexEventComponent(component)
	if !ok {
		return nil, false
	}
	owners := make(map[uint64]string, len(fields))
	for entity, row := range fields {
		value := row["projectile_id"]
		if value.String == nil || value.Bool != nil || value.Int != nil || value.Uint != nil || value.Float != nil || value.Entity != nil || !boundedRequired(*value.String, maxMissileSoundOwnerBytes) {
			return nil, false
		}
		owners[entity] = *value.String
	}
	return owners, true
}
