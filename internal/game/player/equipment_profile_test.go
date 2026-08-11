package player

import (
	"testing"
	"time"

	"github.com/gravestench/akara"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameitem "github.com/gravestench/dark-magic/internal/game/item"
)

type equipmentFixture struct {
	profile  gameitem.Melee
	equipped bool
}

func (fixture equipmentFixture) ActiveMelee(string) (gameitem.Melee, bool, error) {
	return fixture.profile, fixture.equipped, nil
}

func TestEquipmentProfileProjectsActiveWeaponBeforeCombat(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	want := gameitem.Melee{
		Range: 4, PhysicalMinRaw: gamecombat.MustWhole(3).Raw(),
		PhysicalMaxRaw: gamecombat.MustWhole(7).Raw(), WeaponClass: "2hs",
	}
	if err := RegisterEquipmentProfile(engine, equipmentFixture{profile: want, equipped: true}); err != nil {
		t.Fatal(err)
	}
	controls, profiles, appearances, err := equipmentStores(engine)
	if err != nil {
		t.Fatal(err)
	}
	entity, err := engine.World().CreateEntity()
	if err != nil {
		t.Fatal(err)
	}
	setEquipmentComponent(t, controls, entity, map[string]any{"player": "player-1"})
	setEquipmentComponent(t, profiles, entity, map[string]any{"range": 2.0, "physical_min": int64(256), "physical_max": int64(512)})
	setEquipmentComponent(t, appearances, entity, map[string]any{"cof": "", "token": "AM", "palette": "units", "weapon_class": "HTH"})
	if err := engine.Update(40 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	profile, _ := profiles.Get(entity)
	appearance, _ := appearances.Get(entity)
	for name, expected := range map[string]any{"range": want.Range, "physical_min": want.PhysicalMinRaw, "physical_max": want.PhysicalMaxRaw} {
		if got, _ := profile.Get(name); got != expected {
			t.Errorf("%s = %v, want %v", name, got, expected)
		}
	}
	if got, _ := appearance.Get("weapon_class"); got != "2HS" {
		t.Errorf("weapon class = %v, want 2HS", got)
	}
}

func setEquipmentComponent(t *testing.T, store *akara.DynamicStore, entity akara.Entity, values map[string]any) {
	t.Helper()
	if _, err := store.Set(entity, values); err != nil {
		t.Fatal(err)
	}
}
