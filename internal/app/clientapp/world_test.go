package clientapp

import (
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	"github.com/gravestench/dark-magic/internal/inputstate"
	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// TestEntryWorldSpawnsKeepGameplayAndSeamCaptureDistinct prevents screenshot fixture positioning from
// changing normal character admission while preserving generated wilderness arrival.
func TestEntryWorldSpawnsKeepGameplayAndSeamCaptureDistinct(t *testing.T) {
	seam := gametransition.Seam{
		First:  gametransition.SeamEndpoint{ArrivalX: 11, ArrivalY: 12},
		Second: gametransition.SeamEndpoint{ArrivalX: 21, ArrivalY: 22},
	}

	entry, err := entryWorldSpawns("entry", seam, 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	if entry[1] != [2]float64{1, 2} || entry[2] != [2]float64{21, 22} {
		t.Fatalf("ordinary entry spawns = %#v", entry)
	}

	capture, err := entryWorldSpawns("seam", seam, 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	if capture[1] != [2]float64{11, 12} || capture[2] != [2]float64{21, 22} {
		t.Fatalf("seam capture spawns = %#v", capture)
	}

	if _, err := entryWorldSpawns("somewhere", seam, 1, 2); err == nil {
		t.Fatal("accepted an unknown fixture spawn")
	}
}

// TestActivateCurrentConnectedWorldKeepsClickRoute proves connected correction refresh can rebind the
// same map without clearing an in-progress click route.
func TestActivateCurrentConnectedWorldKeepsClickRoute(t *testing.T) {
	engine := gameecs.New()

	t.Cleanup(func() {
		_ = engine.Close()
	})

	controller := &d2movement.MovementController{}

	source, err := d2movement.NewMovementSource(engine, &inputstate.Store{}, "realm-player", "game_world", controller)
	if err != nil {
		t.Fatal(err)
	}

	world := &gameworld.Map{}
	app := &application{
		gameWorlds:       map[int]*gameworld.Map{2: world},
		movementSource:   source,
		activeWorldLevel: 2,
	}

	app.activateWorld(2)

	if err := controller.SetMoveTarget(30, 20); err != nil {
		t.Fatal(err)
	}

	// Connected correction projection reasserts the authoritative level. That
	// is not a transition and must not consume the client's accepted route.
	app.activateWorld(2)

	if !controller.HasMoveTarget() {
		t.Fatal("same-level connected correction cleared the active click route")
	}
}

func TestPresentationWorldIncludesOnlyPreparedLevels(t *testing.T) {
	for _, count := range []int{2, 3} {
		app := &application{activeWorldLevel: 2, gameWorlds: map[int]*gameworld.Map{}, gameWorldZones: map[int]*worldgen.Zone{}}
		for level := 1; level <= count; level++ {
			zone, err := worldgen.NewZone(worldgen.Definition{
				Request: worldgen.Request{Version: worldgen.ContractVersion, Act: 1, LevelID: level},
				Kind:    worldgen.Kind("test"), Bounds: worldgen.Bounds{Width: 1, Height: 1},
				Stamps: []worldgen.Stamp{{ID: 1, Width: 1, Height: 1, DS1Path: "test.ds1", TilePaths: []string{"test.dt1"}}},
			})
			if err != nil {
				t.Fatal(err)
			}
			app.gameWorlds[level] = &gameworld.Map{WidthTiles: level}
			app.gameWorldZones[level] = zone
		}
		app.gameWorlds[99] = &gameworld.Map{}
		snapshot := app.presentationWorld()
		if snapshot.ActiveLevel != 2 || len(snapshot.Maps) != count {
			t.Fatalf("snapshot = %#v", snapshot)
		}
		for index, value := range snapshot.Maps {
			if value.LevelID != index+1 || value.Map != app.gameWorlds[index+1] || value.DS1 != "test.ds1" || value.DT1[0] != "test.dt1" {
				t.Fatalf("map = %#v", value)
			}
		}
		snapshot.Maps[0].DT1[0] = "changed"
		if app.presentationWorld().Maps[0].DT1[0] != "test.dt1" {
			t.Fatal("snapshot mutated recipe")
		}
	}
	if snapshot := (&application{}).presentationWorld(); snapshot.ActiveLevel != 0 || len(snapshot.Maps) != 0 {
		t.Fatalf("empty = %#v", snapshot)
	}
}
