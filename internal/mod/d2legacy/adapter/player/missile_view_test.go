package player

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/gravestench/akara"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// TestProjectWorldViewCopiesOnlyNearbyMissilePresentation verifies that visual
// missile fields cross the boundary while distant authority stays private.
func TestProjectWorldViewCopiesOnlyNearbyMissilePresentation(t *testing.T) {
	checkpoint := missileCheckpoint()
	owner := strings.Repeat("p", maxMissileSoundOwnerBytes)
	checkpoint.Snapshot.Components = append(checkpoint.Snapshot.Components, gameecs.ComponentSnapshot{
		Name: "d2legacy.presentation.missile_sound_owner", Version: 1,
		Fields:    []gameecs.FieldSnapshot{{Name: "projectile_id", Kind: akara.FieldString}},
		Instances: []gameecs.InstanceSnapshot{{Entity: 2, Values: []gameecs.ValueSnapshot{{String: &owner}}}},
	})

	payload, err := ProjectWorldView("alice", checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	var view WorldView
	if err := json.Unmarshal(payload, &view); err != nil {
		t.Fatal(err)
	}

	if err := validateDecodedWorldView(view); err != nil {
		t.Fatal(err)
	}

	if len(view.Missiles) != 1 {
		t.Fatalf("nearby visible missiles = %+v", view.Missiles)
	}

	missile := view.Missiles[0]
	if missile.SoundOwnerID != owner || missile.ID != "missile:2" || missile.Kind != "projectile" ||
		missile.MissileID != "fireball" ||
		missile.DCC != "data/global/missiles/Fireball.dcc" || missile.Position != (HUDPosition{X: 12, Y: 10}) ||
		missile.Velocity != (HUDPosition{X: 1, Y: 0}) || missile.LogicalDirection != -1 ||
		missile.Directions != 16 || missile.FramesPerSecond != 16 || missile.TransparencyMode != 1 {
		t.Fatalf("projected fireball = %+v", missile)
	}
}

// TestValidateWorldViewRejectsUntrustedMissileShape covers limits that prevent
// malformed remote projectiles from reaching rendering code.
func TestValidateWorldViewRejectsUntrustedMissileShape(t *testing.T) {
	valid := WorldView{Version: WorldViewVersion, Tick: 1, Missiles: []WorldMissile{{
		ID: "missile:1", Kind: "effect", MissileID: "fireexplosion", Position: HUDPosition{},
		Act: 1, LevelID: 1, DCC: "data/global/missiles/FireExplosion.dcc", LogicalDirection: 0,
		Directions: 1, FramesPerSecond: 16, TransparencyMode: 1,
	}}}
	if err := validateDecodedWorldView(valid); err != nil {
		t.Fatalf("valid missile view: %v", err)
	}

	for name, mutate := range map[string]func(*WorldMissile){
		"oversized sound owner": func(value *WorldMissile) {
			value.SoundOwnerID = strings.Repeat("a", maxMissileSoundOwnerBytes+1)
		},
		"authority kind": func(value *WorldMissile) { value.Kind = "damage" },
		"bad directions": func(value *WorldMissile) { value.Directions = 3 },
		"projectile pins frame": func(value *WorldMissile) {
			value.Kind, value.LogicalDirection = "projectile", 0
		},
		"unsupported blend": func(value *WorldMissile) { value.TransparencyMode = 3 },
		"non-finite offset": func(value *WorldMissile) { value.OffsetY = math.NaN() },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := valid
			candidate.Missiles = append([]WorldMissile(nil), valid.Missiles...)
			mutate(&candidate.Missiles[0])

			if err := validateDecodedWorldView(candidate); err == nil {
				t.Fatal("invalid missile projection accepted")
			}
		})
	}
}

// missileCheckpoint creates both projectile and lingering-effect rows to test
// their distinct logical-direction contracts.
func missileCheckpoint() simulation.Checkpoint {
	// These constructors retain distinct addresses for positional snapshot values.
	stringValue := func(value string) gameecs.ValueSnapshot {
		return gameecs.ValueSnapshot{String: &value}
	}
	intValue := func(value int64) gameecs.ValueSnapshot {
		return gameecs.ValueSnapshot{Int: &value}
	}
	floatValue := func(value float64) gameecs.ValueSnapshot {
		bits := math.Float64bits(value)
		return gameecs.ValueSnapshot{Float: &bits}
	}
	boolValue := func(value bool) gameecs.ValueSnapshot {
		return gameecs.ValueSnapshot{Bool: &value}
	}
	// component creates the positional schema used by all missile fixture rows.
	component := func(
		name string,
		fields []string,
		instances ...gameecs.InstanceSnapshot,
	) gameecs.ComponentSnapshot {
		result := gameecs.ComponentSnapshot{Name: name, Version: 1, Instances: instances}
		for _, field := range fields {
			result.Fields = append(result.Fields, gameecs.FieldSnapshot{Name: field, Kind: akara.FieldString})
		}

		return result
	}
	player, fireball, farEffect, invisible := uint64(1), uint64(2), uint64(3), uint64(4)
	missileFields := []string{
		"missile_id", "dcc", "palette", "velocity_x", "velocity_y", "directions",
		"frames_per_second", "loop", "transparency_mode", "offset_x", "offset_y", "offset_z",
	}
	// missileValues centralizes the shared visual defaults so each row emphasizes
	// only identity, asset visibility, and velocity.
	missileValues := func(id, dcc string, velocityX float64) []gameecs.ValueSnapshot {
		return []gameecs.ValueSnapshot{
			stringValue(id), stringValue(dcc), stringValue("data/global/palette/units/pal.dat"),
			floatValue(velocityX), floatValue(0), intValue(16), intValue(16), boolValue(true), intValue(1),
			floatValue(0), floatValue(0), floatValue(0),
		}
	}

	return simulation.Checkpoint{Tick: 20, Snapshot: &gameecs.Snapshot{
		Version:  gameecs.SnapshotVersion,
		Tick:     20,
		Entities: []uint64{player, fireball, farEffect, invisible},
		Components: []gameecs.ComponentSnapshot{
			component(
				"d2legacy.player.identity",
				[]string{"player"},
				gameecs.InstanceSnapshot{
					Entity: player, Values: []gameecs.ValueSnapshot{stringValue("alice")},
				},
			),
			component("d2legacy.world.position", []string{"x", "y"},
				gameecs.InstanceSnapshot{
					Entity: player, Values: []gameecs.ValueSnapshot{floatValue(10), floatValue(10)},
				},
				gameecs.InstanceSnapshot{
					Entity: fireball, Values: []gameecs.ValueSnapshot{floatValue(12), floatValue(10)},
				},
				gameecs.InstanceSnapshot{
					Entity: farEffect, Values: []gameecs.ValueSnapshot{floatValue(200), floatValue(10)},
				},
				gameecs.InstanceSnapshot{
					Entity: invisible, Values: []gameecs.ValueSnapshot{floatValue(13), floatValue(10)},
				}),
			component("d2legacy.world.location", []string{"act", "level_id"},
				gameecs.InstanceSnapshot{
					Entity: player, Values: []gameecs.ValueSnapshot{intValue(1), intValue(2)},
				},
				gameecs.InstanceSnapshot{
					Entity: fireball, Values: []gameecs.ValueSnapshot{intValue(1), intValue(2)},
				},
				gameecs.InstanceSnapshot{
					Entity: farEffect, Values: []gameecs.ValueSnapshot{intValue(1), intValue(2)},
				},
				gameecs.InstanceSnapshot{
					Entity: invisible, Values: []gameecs.ValueSnapshot{intValue(1), intValue(2)},
				}),
			component("d2legacy.missile.projectile", missileFields,
				gameecs.InstanceSnapshot{
					Entity: fireball,
					Values: missileValues("fireball", "data/global/missiles/Fireball.dcc", 1),
				},
				gameecs.InstanceSnapshot{
					Entity: farEffect,
					Values: missileValues("far", "data/global/missiles/Far.dcc", 0),
				},
				gameecs.InstanceSnapshot{Entity: invisible, Values: missileValues("helper", "", 0)}),
		},
	}}
}

func TestMissileSoundOwnerProjectionRejectsMalformedPresentSidecars(t *testing.T) {
	for _, source := range []struct {
		name       string
		checkpoint func() simulation.Checkpoint
		project    func(simulation.Checkpoint) error
	}{
		{"world", missileCheckpoint, func(checkpoint simulation.Checkpoint) error {
			_, err := ProjectWorldView("alice", checkpoint)
			return err
		}},
		{"event", semanticCheckpoint, func(checkpoint simulation.Checkpoint) error {
			_, err := ProjectEventView("alice", checkpoint)
			return err
		}},
	} {
		t.Run(source.name, func(t *testing.T) {
			if err := source.project(source.checkpoint()); err != nil {
				t.Fatalf("absent optional sidecar: %v", err)
			}

			for name, mutate := range map[string]func(*gameecs.ComponentSnapshot){
				"wrong version":    func(c *gameecs.ComponentSnapshot) { c.Version = 2 },
				"missing field":    func(c *gameecs.ComponentSnapshot) { c.Fields = nil },
				"wrong field":      func(c *gameecs.ComponentSnapshot) { c.Fields[0].Name = "owner" },
				"wrong type":       func(c *gameecs.ComponentSnapshot) { c.Fields[0].Kind = akara.FieldInt64 },
				"missing value":    func(c *gameecs.ComponentSnapshot) { c.Instances[0].Values = nil },
				"non-string value": func(c *gameecs.ComponentSnapshot) { c.Instances[0].Values[0] = gameecs.ValueSnapshot{} },
				"empty owner":      func(c *gameecs.ComponentSnapshot) { empty := ""; c.Instances[0].Values[0].String = &empty },
				"oversized owner": func(c *gameecs.ComponentSnapshot) {
					value := strings.Repeat("a", maxMissileSoundOwnerBytes+1)
					c.Instances[0].Values[0].String = &value
				},
				"duplicate entity": func(c *gameecs.ComponentSnapshot) { c.Instances = append(c.Instances, c.Instances[0]) },
			} {
				t.Run(name, func(t *testing.T) {
					checkpoint := source.checkpoint()
					owner := "projectile:2"
					component := gameecs.ComponentSnapshot{
						Name: "d2legacy.presentation.missile_sound_owner", Version: 1,
						Fields:    []gameecs.FieldSnapshot{{Name: "projectile_id", Kind: akara.FieldString}},
						Instances: []gameecs.InstanceSnapshot{{Entity: 2, Values: []gameecs.ValueSnapshot{{String: &owner}}}},
					}
					mutate(&component)
					checkpoint.Snapshot.Components = append(checkpoint.Snapshot.Components, component)

					if err := source.project(checkpoint); err == nil {
						t.Fatal("malformed present sidecar accepted")
					}
				})
			}
		})
	}
}
