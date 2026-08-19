package clientapp

import (
	"math"
	"testing"
	"time"

	"github.com/gravestench/akara"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// warpLabScenario retains authority, presentation, and portal facts needed to prove both sides of a
// world transition remain synchronized.
type warpLabScenario struct {
	fixture     *realD2LegacyFixture
	player      akara.Entity
	location    *akara.DynamicComponent
	warps       *akara.DynamicStore
	residents   *akara.DynamicStore
	locations   *akara.DynamicStore
	positions   *akara.DynamicStore
	selectables *akara.DynamicStore
}

// TestWarpLabUsesProductionMovementAndTransition proves point movement, range validation, paired
// traversal, route clearing, and resumed locomotion through production systems.
func TestWarpLabUsesProductionMovementAndTransition(t *testing.T) {
	fixture := newRealD2LegacyFixture(t, realD2LegacyFixtureConfig{
		startScene:         "warp_lab",
		applySceneDefaults: true,
	})
	app := fixture.app

	if !shouldActivateDevelopmentSession(app.options) {
		t.Fatal("Warp Lab direct start did not request offline-session activation")
	}

	if err := app.network.StartSelected(); err != nil {
		t.Fatal(err)
	}

	fixture.advanceGame(t, 10)

	scenario := newWarpLabScenario(t, fixture)

	townID, townX, townY := scenario.portal(t, app.transitionSeam.Town.LevelID)
	if townID == "" {
		t.Fatal("Warp Lab created no town-side warp endpoint")
	}

	scenario.rejectStaleInteraction(t, townX, townY)
	scenario.moveToPortal(t, townX, townY, "town-side")
	scenario.openPortal(t, townX, townY)
	scenario.assertWildernessEntry(t)

	_, returnX, returnY := scenario.portal(t, app.transitionSeam.Wilderness.LevelID)
	if returnX == 0 && returnY == 0 {
		t.Fatal("Warp Lab created no wilderness-side warp endpoint")
	}

	scenario.moveToPortal(t, returnX, returnY, "wilderness-side")
	scenario.returnToTown(t, returnX, returnY)
	scenario.assertLocomotionResumes(t)
}

// newWarpLabScenario requires authoritative admission and both portal endpoints before movement,
// separating fixture/bootstrap failures from traversal policy failures.
func newWarpLabScenario(t *testing.T, fixture *realD2LegacyFixture) *warpLabScenario {
	t.Helper()

	controls := requireRealStore(
		t,
		fixture,
		"d2legacy.world.player_control",
		"Warp Lab admitted no player control store",
	)
	if controls.Len() != 1 {
		t.Fatalf("Warp Lab admitted players = %d, want 1", controls.Len())
	}

	locations := requireRealStore(
		t,
		fixture,
		"d2legacy.world.location",
		"Warp Lab has no authoritative world locations",
	)
	player := controls.Entities()[0]

	location, found := locations.Get(player)
	if !found {
		t.Fatal("Warp Lab player has no authoritative location")
	}

	townLevel := int64(fixture.app.transitionSeam.Town.LevelID)

	level, _ := location.Get("level_id")
	if level != townLevel {
		t.Fatalf("Warp Lab entry level = %v, want town %d", level, townLevel)
	}

	scenario := &warpLabScenario{
		fixture:   fixture,
		player:    player,
		location:  location,
		locations: locations,
		warps: requireRealStore(
			t,
			fixture,
			"d2legacy.world.warp",
			"Warp Lab is missing authoritative warp presentation stores",
		),
		residents: requireRealStore(
			t,
			fixture,
			"d2legacy.world.room_resident",
			"Warp Lab is missing authoritative warp presentation stores",
		),
		positions: requireRealStore(
			t,
			fixture,
			"d2legacy.world.position",
			"Warp Lab is missing authoritative warp presentation stores",
		),
		selectables: requireRealStore(
			t,
			fixture,
			"d2legacy.world.selectable",
			"Warp Lab is missing authoritative warp presentation stores",
		),
	}
	scenario.assertPortalDirectory(t)

	return scenario
}

// assertPortalDirectory requires both endpoints to carry generated room residency, which transition
// logic uses to keep inactive rooms from accepting interaction.
func (scenario *warpLabScenario) assertPortalDirectory(t *testing.T) {
	t.Helper()

	if scenario.warps.Len() != 2 {
		t.Fatalf(
			"Warp Lab authoritative endpoints = %d, want paired production entities",
			scenario.warps.Len(),
		)
	}

	if scenario.residents.Len() < 2 {
		t.Fatalf("Warp Lab room residents = %d, want both endpoints", scenario.residents.Len())
	}

	for _, entity := range scenario.warps.Entities() {
		if _, present := scenario.residents.Get(entity); !present {
			t.Fatalf("Warp Lab endpoint %d has no room residency", entity)
		}
	}
}

// portal resolves a portal by stable public identity and returns its authoritative coordinates,
// avoiding hard-coded fixture positions that could diverge from generated collision.
func (scenario *warpLabScenario) portal(
	t *testing.T,
	levelID int,
) (string, float64, float64) {
	t.Helper()

	for _, entity := range scenario.warps.Entities() {
		location, present := scenario.locations.Get(entity)
		if !present {
			continue
		}

		level, _ := location.Get("level_id")
		if level != int64(levelID) {
			continue
		}

		x, y := dynamicPosition(scenario.positions, entity)
		selectable, _ := scenario.selectables.Get(entity)
		id, _ := selectable.Get("id")

		return id.(string), x, y
	}

	return "", 0, 0
}

// rejectStaleInteraction proves presentation or predicted proximity cannot authorize a remote portal;
// authority must reject interaction until canonical movement reaches range.
func (scenario *warpLabScenario) rejectStaleInteraction(t *testing.T, x, y float64) {
	t.Helper()

	if err := scenario.submitInteraction(x, y); err != nil {
		t.Fatal(err)
	}

	if err := scenario.fixture.app.advanceGame(time.Second / 25); err != nil {
		t.Fatalf("stale Warp Lab interaction terminated the session: %v", err)
	}

	want := int64(scenario.fixture.app.transitionSeam.Town.LevelID)
	if level := scenario.currentLevel(); level != want {
		t.Fatalf("out-of-range Warp Lab interaction changed level to %v", level)
	}
}

// moveToPortal submits through point-and-click input and advances the full application until the
// production route completes, exercising pathfinding rather than teleporting the player.
func (scenario *warpLabScenario) moveToPortal(
	t *testing.T,
	x float64,
	y float64,
	side string,
) {
	t.Helper()

	control := scenario.fixture.app.playerControl
	if err := control.SetMoveTargetWithRadius(x, y, 3.5); err != nil {
		t.Fatal(err)
	}

	for range 250 {
		scenario.fixture.advanceGame(t, 1)

		if !control.HasMoveTarget() {
			return
		}
	}

	t.Fatalf("Warp Lab player never reached the %s warp", side)
}

// openPortal uses the presentation interaction API at the reached point, preserving the same target
// resolution and authority command boundary used by players.
func (scenario *warpLabScenario) openPortal(t *testing.T, x, y float64) {
	t.Helper()

	if err := scenario.submitInteraction(x, y); err != nil {
		t.Fatal(err)
	}

	scenario.fixture.advanceGame(t, 5)
}

// assertWildernessEntry requires authority and active presentation map to agree after traversal,
// catching transitions that update simulation but leave navigation on the old world.
func (scenario *warpLabScenario) assertWildernessEntry(t *testing.T) {
	t.Helper()

	want := scenario.fixture.app.transitionSeam.Wilderness.LevelID
	if level := scenario.currentLevel(); level != int64(want) {
		t.Fatalf("Warp Lab player did not operate paired warp; final level = %v", level)
	}

	if scenario.fixture.app.activeWorldLevel != want {
		t.Fatalf(
			"active presentation world = %d, want %d",
			scenario.fixture.app.activeWorldLevel,
			want,
		)
	}
}

// returnToTown queues traversal back and proves the previous world's movement route is discarded;
// coordinates from one world must never continue driving movement in another.
func (scenario *warpLabScenario) returnToTown(t *testing.T, x, y float64) {
	t.Helper()

	control := scenario.fixture.app.playerControl
	if err := control.SetMoveTarget(x+8, y); err != nil {
		t.Fatal(err)
	}

	if err := scenario.submitInteraction(x, y); err != nil {
		t.Fatal(err)
	}

	scenario.fixture.advanceGame(t, 5)

	want := scenario.fixture.app.transitionSeam.Town.LevelID
	if scenario.currentLevel() != int64(want) || scenario.fixture.app.activeWorldLevel != want {
		t.Fatalf(
			"Warp Lab return left authority/presentation at %v/%d",
			scenario.currentLevel(),
			scenario.fixture.app.activeWorldLevel,
		)
	}

	if control.HasMoveTarget() {
		t.Fatal("Warp Lab return retained a route target from the previous world")
	}
}

// assertLocomotionResumes starts a fresh post-transition route, proving navigation was rebound to the
// returned world rather than merely clearing stale input.
func (scenario *warpLabScenario) assertLocomotionResumes(t *testing.T) {
	t.Helper()

	playerPosition, _ := scenario.positions.Get(scenario.player)
	startX, startY := dynamicPosition(scenario.positions, scenario.player)
	town := scenario.fixture.app.gameWorlds[scenario.fixture.app.transitionSeam.Town.LevelID]

	goalX, goalY, found := town.OpenPointNearSubtileForRadius(startX-6, startY, 1)
	if !found {
		t.Fatal("Warp Lab return has no footprint-safe locomotion target")
	}

	request := gameworld.PathRequest{
		Start:  gameworld.Point{X: startX, Y: startY},
		Goal:   gameworld.Point{X: goalX, Y: goalY},
		Radius: 1,
	}
	if _, err := town.FindPath(request); err != nil {
		t.Fatalf("Warp Lab return position cannot start production locomotion: %v", err)
	}

	if err := scenario.fixture.app.playerControl.SetMoveTarget(goalX, goalY); err != nil {
		t.Fatal(err)
	}

	for range 100 {
		scenario.fixture.advanceGame(t, 1)

		x, _ := playerPosition.Get("x")

		y, _ := playerPosition.Get("y")
		if math.Hypot(x.(float64)-startX, y.(float64)-startY) > 0.5 {
			return
		}
	}

	t.Fatal("Warp Lab locomotion did not resume after the return warp")
}

// submitInteraction encodes the same point-based operation intent produced by presentation input,
// leaving target selection and range checks to authority.
func (scenario *warpLabScenario) submitInteraction(x, y float64) error {
	return scenario.fixture.app.commandIntents.Submit("interaction.open", map[string]any{
		"at": true,
		"x":  x,
		"y":  y,
	})
}

// currentLevel reads the controlled player's authoritative location instead of the application's
// active presentation cache.
func (scenario *warpLabScenario) currentLevel() int64 {
	level, _ := scenario.location.Get("level_id")

	return level.(int64)
}
