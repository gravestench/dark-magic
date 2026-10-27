package clientapp

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// The same development pointer path must target the outbound and return seam sides in separate captures.
func TestSeamPointerFixtureChoosesCurrentLevelEndpoint(t *testing.T) {
	world, _ := gameworld.NewOpenMap(40, 40)
	seam := gametransition.Seam{
		First:  gametransition.SeamEndpoint{LevelID: 1, X: 35, Y: 20},
		Second: gametransition.SeamEndpoint{LevelID: 2, X: 5, Y: 20},
	}

	app := &application{options: Options{FixturePointerMove: true, FixtureWorldSpawn: "seam"},
		profile: content.PresentationProfile{Width: 800, Height: 600}}
	for _, test := range []struct {
		level   int
		spawnX  float64
		targetX float64
	}{
		{1, 29, 35}, {2, 11, 5},
	} {
		fixture := app.newFixturePointerAcceptance(world, [2]float64{test.spawnX, 20}, test.level, seam)
		frame := fixture.Frame(inputstate.Frame{}, test.spawnX, 20, true)
		spawnPixelX, spawnPixelY := world.SubtileToPixel(test.spawnX, 20)

		targetPixelX, targetPixelY := world.SubtileToPixel(test.targetX, 20)
		if frame.CursorX != 400+targetPixelX-spawnPixelX ||
			frame.CursorY != 330+targetPixelY-spawnPixelY {
			t.Fatalf("level %d seam click = (%v,%v), want endpoint %v", test.level,
				frame.CursorX, frame.CursorY, test.targetX)
		}
	}
}

// The near-exit fixture must click the resolved bridge seam through the same screen cursor used by ordinary input.
func TestPointerMovementAcceptanceTargetsResolvedSeam(t *testing.T) {
	world, _ := gameworld.NewOpenMap(40, 40)
	fixture := newPointerMovementAcceptanceAt(world, 29, 20, 35, 20, 800, 330)
	spawnPixelX, spawnPixelY := world.SubtileToPixel(29, 20)
	targetPixelX, targetPixelY := world.SubtileToPixel(35, 20)

	frame := fixture.Frame(inputstate.Frame{}, 29, 20, true)
	if frame.CursorX != 400+targetPixelX-spawnPixelX ||
		frame.CursorY != 330+targetPixelY-spawnPixelY || !frame.Actions["pointer_primary"].Pressed {
		t.Fatalf("seam pointer = (%v,%v), pressed=%v", frame.CursorX, frame.CursorY,
			frame.Actions["pointer_primary"].Pressed)
	}
}

// TestPointerMovementAcceptanceUsesOneClickAndWaitsForSettledAuthority verifies the probe observes production input.
func TestPointerMovementAcceptanceUsesOneClickAndWaitsForSettledAuthority(t *testing.T) {
	fixture := &pointerMovementAcceptance{cursorX: 500, cursorY: 250}
	original := inputstate.Frame{Actions: map[string]inputstate.ActionState{"inventory": {Pressed: true}}}

	clicked := fixture.Frame(original, 10, 20, true)
	if clicked.CursorX != 500 || clicked.CursorY != 250 || !clicked.Actions["pointer_primary"].Pressed {
		t.Fatalf("injected pointer frame = %#v", clicked)
	}

	if !original.Actions["inventory"].Pressed || original.Actions["pointer_primary"].Pressed {
		t.Fatal("fixture mutated the native input snapshot")
	}

	fixture.Frame(inputstate.Frame{}, 12, 20, true)

	if !fixture.moved || !fixture.Busy() {
		t.Fatal("authoritative displacement did not begin acceptance")
	}

	for range pointerAcceptanceStableFrames {
		frame := fixture.Frame(inputstate.Frame{}, 12, 20, true)
		if frame.Actions["pointer_primary"].Pressed {
			t.Fatal("fixture injected more than one click")
		}
	}

	if fixture.Busy() {
		t.Fatal("settled authoritative movement did not complete acceptance")
	}
}

// TestPointerMovementAcceptanceWaitsForPlayerAdmission prevents a synthetic click before authority exposes the player.
func TestPointerMovementAcceptanceWaitsForPlayerAdmission(t *testing.T) {
	fixture := &pointerMovementAcceptance{cursorX: 500, cursorY: 250}

	frame := fixture.Frame(inputstate.Frame{}, 0, 0, false)
	if fixture.clicked || frame.Actions["pointer_primary"].Pressed || !fixture.Busy() {
		t.Fatalf("pre-admission fixture = %#v, %#v", fixture, frame)
	}
}
