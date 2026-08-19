package clientapp

import (
	"testing"

	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// TestEntryWorldSpawnsKeepGameplayAndSeamCaptureDistinct verifies both spawn policies.
func TestEntryWorldSpawnsKeepGameplayAndSeamCaptureDistinct(t *testing.T) {
	seam := gametransition.Seam{
		Town:       gametransition.SeamEndpoint{ArrivalX: 11, ArrivalY: 12},
		Wilderness: gametransition.SeamEndpoint{ArrivalX: 21, ArrivalY: 22},
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

// TestActivateCurrentConnectedWorldKeepsClickRoute guards same-level navigation state.
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
