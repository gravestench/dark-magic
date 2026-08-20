package clientapp

import (
	"math"

	"github.com/gravestench/akara"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
)

const pointerAcceptanceStableFrames = 12

// pointerMovementAcceptance is a development-only black-box probe. It presses
// the ordinary logical mouse action once, then watches copied authoritative
// position facts until the player has moved and stopped. It cannot submit a
// command or change ECS state, so the production input path remains the test.
type pointerMovementAcceptance struct {
	cursorX, cursorY float64
	originX, originY float64
	lastX, lastY     float64
	clicked, moved   bool
	stableFrames     int
	done             bool
}

// newPointerMovementAcceptance converts a reachable world target into the screen-space click consumed by normal input.
func newPointerMovementAcceptance(
	world *gameworld.Map,
	spawnX float64,
	spawnY float64,
	screenWidth int,
	screenHeight int,
) *pointerMovementAcceptance {
	targetX, targetY := acceptanceMovementTarget(world, spawnX, spawnY)
	spawnPixelX, spawnPixelY := world.SubtileToPixel(spawnX, spawnY)
	targetPixelX, targetPixelY := world.SubtileToPixel(targetX, targetY)

	return &pointerMovementAcceptance{
		cursorX: float64(screenWidth)/2 + targetPixelX - spawnPixelX,
		cursorY: float64(screenHeight)/2 + targetPixelY - spawnPixelY,
	}
}

// acceptanceMovementTarget picks a reachable open point, not a convenient
// screen pixel that might happen to land behind a wall in another town layout.
func acceptanceMovementTarget(world *gameworld.Map, spawnX, spawnY float64) (float64, float64) {
	for radius := 18; radius >= 6; radius -= 3 {
		distance := float64(radius)

		offsets := [][2]float64{
			{distance, 0},
			{0, distance},
			{-distance, 0},
			{0, -distance},
			{distance, distance},
			{-distance, distance},
		}
		for _, offset := range offsets {
			x, y, found := world.OpenPointNearSubtile(spawnX+offset[0], spawnY+offset[1])
			if !found {
				continue
			}

			path, err := world.FindPath(gameworld.PathRequest{
				Start:  gameworld.Point{X: spawnX, Y: spawnY},
				Goal:   gameworld.Point{X: x, Y: y},
				Radius: 1.5,
			})
			if err == nil && len(path) > 2 {
				return x, y
			}
		}
	}

	return spawnX, spawnY
}

// Frame injects one ordinary pointer press, then passively observes authority movement until it settles.
func (fixture *pointerMovementAcceptance) Frame(
	frame inputstate.Frame,
	x float64,
	y float64,
	playerPresent bool,
) inputstate.Frame {
	if fixture == nil || fixture.done || !playerPresent {
		return frame
	}

	if !fixture.clicked {
		return fixture.pressPointer(frame, x, y)
	}

	fixture.observeMovement(x, y)

	return frame
}

// pressPointer clones actions before injection so the acceptance probe cannot mutate the input backend's snapshot.
func (fixture *pointerMovementAcceptance) pressPointer(
	frame inputstate.Frame,
	x float64,
	y float64,
) inputstate.Frame {
	fixture.originX, fixture.originY = x, y
	fixture.lastX, fixture.lastY = x, y
	fixture.clicked = true
	frame.CursorX, frame.CursorY = fixture.cursorX, fixture.cursorY
	frame.Actions = cloneActions(frame.Actions)
	frame.Actions["pointer_primary"] = inputstate.ActionState{Pressed: true, Down: true}

	return frame
}

// observeMovement requires several stationary authority frames so a transient
// pause cannot complete the capture early.
func (fixture *pointerMovementAcceptance) observeMovement(x, y float64) {
	distance := math.Hypot(x-fixture.originX, y-fixture.originY)

	fixture.moved = fixture.moved || distance > 0.75
	if fixture.moved && math.Hypot(x-fixture.lastX, y-fixture.lastY) < 0.01 {
		fixture.stableFrames++
	} else {
		fixture.stableFrames = 0
	}

	fixture.lastX, fixture.lastY = x, y
	fixture.done = fixture.moved && fixture.stableFrames >= pointerAcceptanceStableFrames
}

// Busy reports whether capture must keep waiting for the injected movement to complete.
func (fixture *pointerMovementAcceptance) Busy() bool { return fixture != nil && !fixture.done }

// cloneActions gives synthetic input private map storage while retaining every physical action unchanged.
func cloneActions(source map[string]inputstate.ActionState) map[string]inputstate.ActionState {
	result := make(map[string]inputstate.ActionState, len(source)+1)
	for name, state := range source {
		result[name] = state
	}

	return result
}

// controlledPlayerPosition reads presentation state only; the acceptance probe
// must never reach into authority ECS state.
func (app *application) controlledPlayerPosition() (float64, float64, bool) {
	engine := app.presentationSimulation()
	if engine == nil {
		return 0, 0, false
	}

	controls, found := akara.GetDynamicStore(engine.World(), "d2legacy.world.player_control")
	if !found {
		return 0, 0, false
	}

	positions, found := akara.GetDynamicStore(engine.World(), "d2legacy.world.position")
	if !found {
		return 0, 0, false
	}

	wanted := "local-player"

	if app.network != nil {
		if player, ok := app.network.Status()["player_id"].(string); ok && player != "" {
			wanted = player
		}
	}

	for _, entity := range controls.Entities() {
		control, present := controls.Get(entity)
		if !present {
			continue
		}

		player, _ := control.Get("player")
		if player != wanted {
			continue
		}

		position, present := positions.Get(entity)
		if !present {
			return 0, 0, false
		}

		x, _ := position.Get("x")
		y, _ := position.Get("y")

		return x.(float64), y.(float64), true
	}

	return 0, 0, false
}
