package player

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gravestench/akara"
	gamecombat "github.com/gravestench/dark-magic/internal/game/combat"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameitem "github.com/gravestench/dark-magic/internal/game/item"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/game/targeting"
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

func TestEquippingInventoryWeaponChangesAuthoritativeMeleeImpact(t *testing.T) {
	engine := gameecs.New()
	defer engine.Close()
	authority := gameitem.NewAuthority()
	weapon := gameitem.Melee{
		Range: 4, PhysicalMinRaw: gamecombat.MustWhole(6).Raw(),
		PhysicalMaxRaw: gamecombat.MustWhole(6).Raw(), WeaponClass: "1hs",
	}
	state, err := gameitem.NewState(
		gameitem.Layout{Grids: map[gameitem.Container]gameitem.Grid{gameitem.ContainerInventory: {Width: 10, Height: 4}}},
		[]gameitem.Item{{ID: "short-sword", Code: "ssd", Width: 1, Height: 3, BodySlots: []string{"rarm", "larm"}, Melee: weapon}},
		map[string]gameitem.Placement{"short-sword": {Container: gameitem.ContainerInventory}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := authority.Register("hero", state); err != nil {
		t.Fatal(err)
	}
	if err := RegisterEquipmentProfile(engine, authority); err != nil {
		t.Fatal(err)
	}
	if err := gamecombat.RegisterBasicMelee(engine, gamecombat.BasicMeleePolicy{HitChance: 100}); err != nil {
		t.Fatal(err)
	}
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	if err := gameitem.RegisterCommands(session, authority); err != nil {
		t.Fatal(err)
	}

	controls, profiles, appearances, err := equipmentStores(engine)
	if err != nil {
		t.Fatal(err)
	}
	selectables, _ := akara.GetDynamicStore(engine.World(), targeting.Component)
	positions, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	locations, _ := akara.GetDynamicStore(engine.World(), "d2legacy.world.location")
	vitals, _ := akara.GetDynamicStore(engine.World(), "d2legacy.player.vitals")
	monsterStats, _ := akara.GetDynamicStore(engine.World(), "d2legacy.monster.stats")
	attackRequests, _ := akara.GetDynamicStore(engine.World(), gamecombat.BasicAttackRequest)
	player := engine.World().MustCreateEntity()
	hostile := engine.World().MustCreateEntity()
	setEquipmentComponent(t, controls, player, map[string]any{"player": "hero"})
	setEquipmentComponent(t, profiles, player, map[string]any{"range": 2.0, "physical_min": gamecombat.MustWhole(1).Raw(), "physical_max": gamecombat.MustWhole(2).Raw()})
	setEquipmentComponent(t, appearances, player, map[string]any{"cof": "", "token": "AM", "palette": "units", "weapon_class": "HTH"})
	setEquipmentComponent(t, selectables, player, map[string]any{"id": "player:hero", "kind": targeting.KindPlayer, "label": "Hero", "owner": "hero", "radius": 1.0, "priority": int64(10)})
	setEquipmentComponent(t, positions, player, map[string]any{"x": 2.0, "y": 2.0})
	setEquipmentComponent(t, locations, player, map[string]any{"act": int64(1), "level_id": int64(2)})
	setEquipmentComponent(t, vitals, player, map[string]any{"health": int64(10), "max_health": int64(10), "mana": int64(0), "max_mana": int64(0), "mana_raw": int64(0), "max_mana_raw": int64(0)})
	setEquipmentComponent(t, selectables, hostile, map[string]any{"id": "monster:fallen", "kind": targeting.KindHostile, "label": "Fallen", "owner": "", "radius": 0.5, "priority": int64(20)})
	setEquipmentComponent(t, positions, hostile, map[string]any{"x": 5.0, "y": 2.0})
	setEquipmentComponent(t, locations, hostile, map[string]any{"act": int64(1), "level_id": int64(2)})
	setEquipmentComponent(t, monsterStats, hostile, map[string]any{"level": int64(1), "health": gamecombat.MustWhole(10).Raw(), "max_health": gamecombat.MustWhole(10).Raw(), "defense": int64(0), "attack_rating": int64(0), "physical_min": int64(0), "physical_max": int64(0), "experience": int64(0)})
	setEquipmentComponent(t, attackRequests, player, map[string]any{"target_id": "monster:fallen", "request_tick": int64(1)})

	payload, _ := json.Marshal(gameitem.MovePayload{ItemID: "short-sword", Destination: gameitem.Placement{Container: gameitem.ContainerEquipment, Slot: "rarm", WeaponSet: 0}})
	if err := session.Submit(simulation.Command{Tick: 1, Player: "hero", Authority: simulation.AuthorityPlayer, Sequence: 1, Kind: gameitem.MoveCommand, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if err := session.Step(); err != nil {
		t.Fatal(err)
	}
	profile, _ := profiles.Get(player)
	if got, _ := profile.Get("range"); got != weapon.Range {
		t.Fatalf("equipped range = %v, want %v", got, weapon.Range)
	}
	if got, _ := appearances.Get(player); got != nil {
		weaponClass, _ := got.Get("weapon_class")
		if weaponClass != "1HS" {
			t.Fatalf("equipped weapon class = %v, want 1HS", weaponClass)
		}
	}
	stats, _ := monsterStats.Get(hostile)
	health, _ := stats.Get("health")
	if health != gamecombat.MustWhole(4).Raw() {
		t.Fatalf("equipped melee health = %v, want %d", health, gamecombat.MustWhole(4).Raw())
	}
}

func setEquipmentComponent(t *testing.T, store *akara.DynamicStore, entity akara.Entity, values map[string]any) {
	t.Helper()
	if _, err := store.Set(entity, values); err != nil {
		t.Fatal(err)
	}
}
