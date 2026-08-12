package player

import (
	"fmt"
	"strings"

	"github.com/gravestench/akara"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameitem "github.com/gravestench/dark-magic/internal/game/item"
)

const EquipmentProfileSystemID = "player.equipment_melee_profile"

// EquipmentSource is the narrow item-authority view needed by player combat.
// Combat never learns about grids, held items, vendors, or cursor behavior.
type EquipmentSource interface {
	ActiveMelee(owner string) (gameitem.Melee, bool, error)
}

// RegisterEquipmentProfile projects active equipment before combat runs.
func RegisterEquipmentProfile(engine *gameecs.Engine, source EquipmentSource) error {
	if engine == nil || source == nil {
		return fmt.Errorf("player: equipment profile requires engine and item authority")
	}
	controls, profiles, appearances, err := equipmentStores(engine)
	if err != nil {
		return err
	}
	return engine.Register(gameecs.Definition{
		ID: EquipmentProfileSystemID, Phase: gameecs.PhasePreSimulate,
		All:  []akara.ComponentType{controls, profiles, appearances},
		Read: []akara.ComponentType{controls}, Write: []akara.ComponentType{profiles, appearances},
		Update: func(_ gameecs.Context, entities []akara.Entity, _ *akara.CommandBuffer) error {
			for _, entity := range entities {
				control, _ := controls.Get(entity)
				owner, _ := control.Get("player")
				weapon, equipped, err := source.ActiveMelee(owner.(string))
				if err != nil {
					return err
				}
				if !equipped {
					weapon = gameitem.Melee{Range: 2, PhysicalMinRaw: gamecombat.MustWhole(1).Raw(), PhysicalMaxRaw: gamecombat.MustWhole(2).Raw(), WeaponClass: "HTH"}
				}
				if err := applyEquipmentProfile(entity, weapon, profiles, appearances); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

func equipmentStores(engine *gameecs.Engine) (*akara.DynamicStore, *akara.DynamicStore, *akara.DynamicStore, error) {
	controls, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.world.player_control", Version: 1, Fields: []akara.Field{{Name: "player", Kind: akara.FieldString}}})
	if err != nil {
		return nil, nil, nil, err
	}
	profiles, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: gamecombat.MeleeProfile, Version: 1, Fields: []akara.Field{{Name: "range", Kind: akara.FieldFloat64}, {Name: "physical_min", Kind: akara.FieldInt64}, {Name: "physical_max", Kind: akara.FieldInt64}}})
	if err != nil {
		return nil, nil, nil, err
	}
	appearances, err := akara.RegisterSchema(engine.World(), akara.Schema{Name: "d2.player.appearance", Version: 1, Fields: []akara.Field{{Name: "cof", Kind: akara.FieldString}, {Name: "token", Kind: akara.FieldString}, {Name: "palette", Kind: akara.FieldString}, {Name: "weapon_class", Kind: akara.FieldString}}})
	return controls, profiles, appearances, err
}

func applyEquipmentProfile(entity akara.Entity, weapon gameitem.Melee, profiles, appearances *akara.DynamicStore) error {
	if weapon.Range <= 0 || weapon.PhysicalMinRaw < 0 || weapon.PhysicalMaxRaw < weapon.PhysicalMinRaw || strings.TrimSpace(weapon.WeaponClass) == "" {
		return fmt.Errorf("player: invalid equipped melee profile")
	}
	profile, _ := profiles.Get(entity)
	appearance, _ := appearances.Get(entity)
	for name, value := range map[string]any{"range": weapon.Range, "physical_min": weapon.PhysicalMinRaw, "physical_max": weapon.PhysicalMaxRaw} {
		if err := profile.Set(name, value); err != nil {
			return err
		}
	}
	return appearance.Set("weapon_class", strings.ToUpper(strings.TrimSpace(weapon.WeaponClass)))
}
