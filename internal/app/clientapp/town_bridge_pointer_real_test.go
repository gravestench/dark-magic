package clientapp

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	"github.com/gravestench/dark-magic/internal/platform/desktop"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// This shell runs the production game_world scene and its pointer input without opening a native window.
type bridgeTestRenderer struct{}

// Start leaves native rendering inactive while preserving component startup ordering.
func (*bridgeTestRenderer) Start(context.Context) error { return nil }

// Stop has no native resources to release.
func (*bridgeTestRenderer) Stop(context.Context) error { return nil }

// Run is unused because the test advances production frame stages directly.
func (*bridgeTestRenderer) Run(context.Context) error { return nil }

// Name marks renderer diagnostics as headless test output.
func (*bridgeTestRenderer) Name() string { return "bridge-test" }

// SubscribeFrame is inert because the test calls each frame stage itself.
func (*bridgeTestRenderer) SubscribeFrame(func()) func() { return func() {} }

// SubscribePostFrame has no presented pixels to observe.
func (*bridgeTestRenderer) SubscribePostFrame(func()) func() { return func() {} }

// SubscribeOverlay has no native surface on which to draw.
func (*bridgeTestRenderer) SubscribeOverlay(func()) func() { return func() {} }

// SubscribeViewport keeps the manifest-sized logical viewport fixed.
func (*bridgeTestRenderer) SubscribeViewport(func(int, int)) func() { return func() {} }

// AttachComposer lets Lua build its real scene graph without GPU upload.
func (*bridgeTestRenderer) AttachComposer(*render.Composer) error { return nil }

// AttachAudio leaves sound playback outside the movement assertion.
func (*bridgeTestRenderer) AttachAudio(*audio.Mixer) error { return nil }

// CaptureScreenshot is inert because authority is asserted without image pixels.
func (*bridgeTestRenderer) CaptureScreenshot(string) error { return nil }

// WindowSize gives input routing the game-world manifest dimensions.
func (*bridgeTestRenderer) WindowSize() (int, int) { return 800, 600 }

// Resolution keeps the cursor in the ordinary logical viewport.
func (*bridgeTestRenderer) Resolution() (int, int) { return 800, 600 }

// ScreenToGame applies no native scaling to the fixed logical viewport.
func (*bridgeTestRenderer) ScreenToGame(x, y int) (int, int, bool) { return x, y, true }

// SetWindowTitle has no window state to mutate.
func (*bridgeTestRenderer) SetWindowTitle(string) {}

// SetResidencyDebug has no texture cache state to mutate.
func (*bridgeTestRenderer) SetResidencyDebug(bool) {}

// SetTextureUploadBudget has no uploads to throttle.
func (*bridgeTestRenderer) SetTextureUploadBudget(uint64) {}

// SetTextureCacheBudget has no texture cache to limit.
func (*bridgeTestRenderer) SetTextureCacheBudget(uint64) {}

// BackendDiagnostics exposes no native frame metrics.
func (*bridgeTestRenderer) BackendDiagnostics() any { return nil }

// CacheDiagnostics exposes no native texture metrics.
func (*bridgeTestRenderer) CacheDiagnostics() any { return nil }

// bridgeTestInput supplies one logical frame to the production input router.
type bridgeTestInput struct{ frame inputstate.Frame }

// Start has no device initialization because the test owns the frame.
func (*bridgeTestInput) Start(context.Context) error { return nil }

// Stop has no device state to release.
func (*bridgeTestInput) Stop(context.Context) error { return nil }

// Snapshot supplies the frame through the normal desktop input interface.
func (input *bridgeTestInput) Snapshot() inputstate.Frame { return input.frame }

// bridgeTestConsole never captures gameplay clicks, matching a closed developer console.
type bridgeTestConsole struct{}

// LoadFont needs no GPU font because console drawing is outside this test.
func (bridgeTestConsole) LoadFont() error { return nil }

// Handle leaves the pointer event available to the focused gameplay scene.
func (bridgeTestConsole) Handle(context.Context, inputstate.Frame) bool { return false }

// Draw has no native overlay to present.
func (bridgeTestConsole) Draw(int, int) {}

// Close has no console resources to release.
func (bridgeTestConsole) Close() {}

// newBridgePointerClient runs the production Lua screen and authority behind a headless input surface.
func newBridgePointerClient(t *testing.T, hosted bool, spawn ...string) (*application, *bridgeTestInput) {
	t.Helper()
	requireRealD2LegacyAssets(t)
	options := realD2LegacyOptions(t)
	options.StartScene = "game_world"
	options.FixtureCharacters = 1
	options.FixtureWorldSpawn = "seam"

	if len(spawn) > 0 {
		options.FixtureWorldSpawn = spawn[0]
	}

	app := newApplication(options)
	input := &bridgeTestInput{}

	t.Cleanup(func() {
		if app.scenes != nil {
			if err := app.scenes.Close(context.Background()); err != nil {
				t.Errorf("close bridge scene: %v", err)
			}
		}
		// #626 tracks root cancellation stopping a host before its final refresh.
		if app.network != nil {
			if err := app.network.Close(); err != nil {
				t.Errorf("close bridge connection: %v", err)
			}
		}

		if err := app.shutdown(); err != nil {
			t.Errorf("shut down bridge client: %v", err)
		}

		app.stop()

		if app.entitySimulation != nil {
			if err := app.entitySimulation.Close(); err != nil {
				t.Errorf("close bridge authority: %v", err)
			}
		}

		if err := content.Close(options.Content); err != nil {
			t.Errorf("close bridge content: %v", err)
		}
	})

	startBridgeClientStep(t, "settings", app.loadSettings)
	startBridgeClientStep(t, "presentation profile", app.loadPresentationConfiguration)
	app.installPresentationServices(&desktop.Bundle{Renderer: &bridgeTestRenderer{}, Input: input},
		desktop.DefaultOptions())
	app.console = bridgeTestConsole{}

	startBridgeClientStep(t, "game catalogs", app.loadGameCatalogs)
	startBridgeClientStep(t, "offline authority", app.buildOfflineSession)
	startBridgeClientStep(t, "Lua capabilities", app.registerLuaRuntime)
	startBridgeClientStep(t, "engine host", app.startEngineHost)

	if hosted {
		// The direct-start scene admits its selected character after the listen host is selected.
		startBridgeClientStep(t, "listen host", app.network.Host)
	}

	startBridgeClientStep(t, "game-world scene", app.loadScriptComponents)

	if hosted {
		deadline := time.Now().Add(30 * time.Second)
		for !app.network.Connected() && time.Now().Before(deadline) {
			if status := app.network.Status(); status["phase"] == "failed" {
				t.Fatalf("listen connection failed: %v", status["error"])
			}

			time.Sleep(10 * time.Millisecond)
		}

		if !app.network.Connected() {
			t.Fatalf("listen connection timed out: %#v", app.network.Status())
		}
	}

	return app, input
}

// startBridgeClientStep names setup failures so they cannot be mistaken for a crossing failure.
func startBridgeClientStep(t *testing.T, name string, start func() error) {
	t.Helper()

	if err := start(); err != nil {
		t.Fatalf("start %s: %v", name, err)
	}
}

// advanceBridgePointerFrame preserves the client frame order so Lua sees input after the authority tick.
func advanceBridgePointerFrame(t *testing.T, app *application) {
	t.Helper()

	ctx := app.scenes.FrameContext(context.Background())
	app.publishInput(ctx)

	if err := app.advanceGame(time.Second / 25); err != nil {
		t.Fatal(err)
	}

	if err := app.scenes.Update(ctx, time.Second/25); err != nil {
		t.Fatal(err)
	}
}

// clickBridgeEndpoint projects a visible subtile into a normal screen click and requires Lua to admit it.
func clickBridgeEndpoint(t *testing.T, app *application, input *bridgeTestInput, target gameworld.Point) {
	t.Helper()

	x, y, present := app.controlledPlayerPosition()
	if !present {
		t.Fatal("player was not admitted")
	}

	cursor := app.currentWorld().Map.Coordinates().SubtileToScreen(target,
		gameworld.Point{X: x, Y: y}, gameworld.Point{X: 400, Y: 330})
	if cursor.X < 0 || cursor.X >= 800 || cursor.Y < 0 || cursor.Y >= 480 {
		t.Fatalf("bridge target is outside the viewport: %v", cursor)
	}

	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Pressed: true, Down: true}}}

	advanceBridgePointerFrame(t, app)

	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y}

	if !app.playerControl.HasMoveTarget() {
		t.Fatal("game_world did not turn the pointer click into a movement target")
	}
}

// awaitBridgeLevel bounds a crossing wait; hosted authority advances on a real timer between client frames.
func awaitBridgeLevel(t *testing.T, app *application, target int) {
	t.Helper()

	for range 400 {
		if app.network.Connected() {
			time.Sleep(40 * time.Millisecond)
		}

		advanceBridgePointerFrame(t, app)

		if app.activeWorldLevel == target {
			return
		}
	}

	t.Fatalf("pointer click did not reach level %d; remained at %d", target, app.activeWorldLevel)
}

// testTownBridgePointerRoundTrip crosses the displayed east opening and returns in one live session.
func testTownBridgePointerRoundTrip(t *testing.T, hosted bool) {
	t.Helper()

	app, input := newBridgePointerClient(t, hosted)
	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	seam := app.transitionSeams[0]
	if seam.First.LevelID != 1 || seam.Second.LevelID != 2 || seam.First.X < 250 {
		t.Fatalf("selected town seam does not meet the visible bridge: %#v", seam)
	}

	clickBridgeEndpoint(t, app, input, gameworld.Point{X: seam.First.X, Y: seam.First.Y})
	awaitBridgeLevel(t, app, 2)
	assertBridgePointerLevel(t, app, 2, seam.Second)

	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	clickBridgeEndpoint(t, app, input, gameworld.Point{X: seam.Second.X, Y: seam.Second.Y})
	awaitBridgeLevel(t, app, 1)
	assertBridgePointerLevel(t, app, 1, seam.First)
}

// assertBridgePointerLevel checks the canonical level, and the connected projection when present.
func assertBridgePointerLevel(t *testing.T, app *application, levelID int,
	arrival gametransition.SeamEndpoint) {
	t.Helper()

	if app.network.Connected() {
		view, _ := app.network.client.View()
		assertConnectedColdPlainsArrival(t, app, view, levelID, arrival)

		return
	}

	if level, found := controlledPlayerLevel(app.entitySimulation.World(), "local-player"); !found || level != levelID {
		t.Fatalf("offline authority level = %d/%t, want %d", level, found, levelID)
	}
}

// TestTownBridgePointerRoundTripOffline proves the local game-world scene routes both ordinary bridge clicks.
func TestTownBridgePointerRoundTripOffline(t *testing.T) {
	testTownBridgePointerRoundTrip(t, false)
}

// TestTownBridgePointerRoundTripConnected proves the listen host commits both ordinary bridge clicks.
func TestTownBridgePointerRoundTripConnected(t *testing.T) {
	testTownBridgePointerRoundTrip(t, true)
}

// testTownBridgePointerBeyondEastLip checks a normal outside-map click and its return trip.
func testTownBridgePointerBeyondEastLip(t *testing.T, hosted bool) {
	t.Helper()

	app, input := newBridgePointerClient(t, hosted)
	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	seam := app.transitionSeams[0]
	town := app.currentWorld().Map
	target := gameworld.Point{X: float64(town.WidthSubtiles + 1), Y: seam.First.Y}

	x, y, present := app.controlledPlayerPosition()
	if !present {
		t.Fatal("bridge player was not admitted")
	}

	if math.Hypot(x-seam.First.X, y-seam.First.Y) <= seam.TriggerRadius {
		t.Fatal("bridge fixture began inside the transition trigger")
	}

	cursor := town.Coordinates().SubtileToScreen(target, gameworld.Point{X: x, Y: y},
		gameworld.Point{X: 400, Y: 330})
	if target.X < float64(town.WidthSubtiles) || cursor.X < 0 || cursor.X >= 800 ||
		cursor.Y < 0 || cursor.Y >= 480 {
		t.Fatalf("outside bridge target %v projects outside gameplay viewport at %v", target, cursor)
	}

	clickBridgeEndpoint(t, app, input, target)
	awaitBridgeLevel(t, app, 2)
	assertBridgePointerLevel(t, app, 2, seam.Second)

	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	clickBridgeEndpoint(t, app, input, gameworld.Point{X: seam.Second.X, Y: seam.Second.Y})
	awaitBridgeLevel(t, app, 1)
	assertBridgePointerLevel(t, app, 1, seam.First)
}

// TestTownBridgePointerBeyondEastLipOffline exercises the prepared bridge from a local scene.
func TestTownBridgePointerBeyondEastLipOffline(t *testing.T) {
	testTownBridgePointerBeyondEastLip(t, false)
}

// TestTownBridgePointerBeyondEastLipConnected exercises the same click through a listen host.
func TestTownBridgePointerBeyondEastLipConnected(t *testing.T) {
	testTownBridgePointerBeyondEastLip(t, true)
}

// bridgeCameraPosition reads the camera used by the live scene, which can lag behind the hero.
func bridgeCameraPosition(t *testing.T, app *application) gameworld.Point {
	t.Helper()
	world := app.presentationSimulation().World()
	follows, found := akara.GetDynamicStore(world, "d2legacy.world.camera_follow")
	positions, hasPositions := akara.GetDynamicStore(world, "d2legacy.world.position")
	if !found || !hasPositions {
		t.Fatalf("bridge camera stores unavailable: follows=%v positions=%v", found, hasPositions)
	}
	if follows.Len() != 1 {
		t.Fatalf("bridge camera count = %d, want 1", follows.Len())
	}
	position, found := positions.Get(follows.Entities()[0])
	if !found {
		t.Fatal("bridge camera has no position")
	}
	x, errX := position.Get("x")
	y, errY := position.Get("y")
	if errX != nil || errY != nil {
		t.Fatalf("read bridge camera position: x=%v y=%v", errX, errY)
	}
	return gameworld.Point{X: x.(float64), Y: y.(float64)}
}

// bridgeClickAnchor follows the rendered canvas clamp using the live camera position.
func bridgeClickAnchor(t *testing.T, town *gameworld.Map, camera gameworld.Point) gameworld.Point {
	t.Helper()
	width := float64((town.WidthTiles+town.HeightTiles)*gameworld.TilePixelWidth/2 + gameworld.PreviewMargin*2)
	height := float64((town.WidthTiles+town.HeightTiles)*gameworld.TilePixelHeight/2 + gameworld.PreviewMargin*2)
	cameraPixel := town.Coordinates().SubtileToWorldPixel(camera)
	clamp := func(wanted, size, viewport float64) float64 {
		if size <= viewport {
			return viewport / 2
		}
		return math.Max(viewport-size/2, math.Min(size/2, wanted))
	}
	rootX := clamp(400+width/2-cameraPixel.X, width, 800)
	rootY := clamp(330+height/2-cameraPixel.Y, height, 600)
	anchor := gameworld.Point{X: rootX - width/2 + cameraPixel.X, Y: rootY - height/2 + cameraPixel.Y}
	if math.Abs(anchor.X-400) > 0.01 || math.Abs(anchor.Y-330) > 0.01 {
		t.Fatalf("bridge shoulder camera unexpectedly clamped: camera=%v anchor=%v", camera, anchor)
	}
	return anchor
}

// testTownBridgeShoulderPointerCrosses reproduces a visible eastward click from the bridge shoulder.
func testTownBridgeShoulderPointerCrosses(t *testing.T, hosted bool) {
	t.Helper()
	app, input := newBridgePointerClient(t, hosted)
	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	shoulder := gameworld.Point{X: 269, Y: 80}
	if err := app.playerControl.SetMoveTarget(shoulder.X, shoulder.Y); err != nil {
		t.Fatal(err)
	}
	for range 200 {
		if hosted {
			time.Sleep(40 * time.Millisecond)
		}
		advanceBridgePointerFrame(t, app)
		x, y, present := app.controlledPlayerPosition()
		if present && math.Hypot(x-shoulder.X, y-shoulder.Y) < 0.35 {
			break
		}
	}
	x, y, present := app.controlledPlayerPosition()
	if !present || math.Hypot(x-shoulder.X, y-shoulder.Y) >= 0.35 || app.activeWorldLevel != 1 {
		t.Fatalf("bridge shoulder setup stopped at (%.2f,%.2f), level %d", x, y, app.activeWorldLevel)
	}
	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	town := app.currentWorld().Map
	seam := app.transitionSeams[0]
	camera := bridgeCameraPosition(t, app)
	anchor := bridgeClickAnchor(t, town, camera)
	outside := gameworld.Point{X: float64(town.WidthSubtiles + 1), Y: 80}
	projected := town.Coordinates().SubtileToScreen(outside, camera, anchor)
	cursor := gameworld.Point{X: math.Round(projected.X), Y: math.Round(projected.Y)}
	if cursor.X < 0 || cursor.X >= 800 || cursor.Y < 0 || cursor.Y >= 480 {
		t.Fatalf("bridge shoulder outside click is outside gameplay viewport: %v", cursor)
	}
	raw := town.Coordinates().ScreenToSubtile(cursor, camera, anchor)
	if gameworld.CollisionCell(raw.X) < town.WidthSubtiles || gameworld.CollisionCell(raw.Y) != 80 {
		t.Fatalf("bridge shoulder cursor %v reverse-projects to %v, want east edge at lane 80", cursor, raw)
	}
	start := gameworld.Point{X: x, Y: y}
	_, rawPathErr := town.FindPath(gameworld.PathRequest{Start: start, Goal: raw, Radius: 1})
	_, safePathErr := town.FindPath(gameworld.PathRequest{
		Start: start, Goal: gameworld.Point{X: seam.First.X, Y: 82}, Radius: 1,
	})
	t.Logf("bridge shoulder camera=%v anchor=%v cursor=%v raw=%v start=%v raw route=%v safe route=%v strip=[%d,%d]",
		camera, anchor, cursor, raw, start, rawPathErr, safePathErr,
		seam.First.TriggerStrip.LaneMin, seam.First.TriggerStrip.LaneMax)
	if rawPathErr == nil || safePathErr != nil {
		t.Fatalf("bridge shoulder route preconditions changed: raw=%v safe=%v", rawPathErr, safePathErr)
	}

	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Pressed: true, Down: true}}}
	advanceBridgePointerFrame(t, app)
	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y}
	if !app.playerControl.HasMoveTarget() {
		t.Fatal("bridge shoulder click did not become a movement request")
	}
	awaitBridgeLevel(t, app, 2)
	assertBridgePointerLevel(t, app, 2, seam.Second)
	x, y, _ = app.controlledPlayerPosition()
	t.Logf("bridge shoulder crossing committed level=%d position=(%.2f,%.2f)", app.activeWorldLevel, x, y)
}

func TestTownBridgeShoulderPointerCrossesOffline(t *testing.T) {
	testTownBridgeShoulderPointerCrosses(t, false)
}

func TestTownBridgeShoulderPointerCrossesConnected(t *testing.T) {
	testTownBridgeShoulderPointerCrosses(t, true)
}

// clickAndAwaitBridgePosition requires the scene click to produce a movement target and reach it.
func clickAndAwaitBridgePosition(t *testing.T, app *application, input *bridgeTestInput,
	target gameworld.Point, frames int) {
	t.Helper()
	clickBridgeEndpoint(t, app, input, target)

	for range frames {
		if app.network.Connected() {
			time.Sleep(40 * time.Millisecond)
		}

		advanceBridgePointerFrame(t, app)

		x, y, present := app.controlledPlayerPosition()
		if present && math.Hypot(x-target.X, y-target.Y) < 0.5 {
			return
		}
	}

	x, y, _ := app.controlledPlayerPosition()
	t.Fatalf("normal pointer route stopped at (%.2f,%.2f), target %v", x, y, target)
}

// testTownBridgePointerFromEntry walks from normal admission with ordinary scene clicks.
func testTownBridgePointerFromEntry(t *testing.T, hosted bool) {
	t.Helper()

	app, input := newBridgePointerClient(t, hosted, "entry")
	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	seam := app.transitionSeams[0]
	town := app.currentWorld().Map

	x, y, present := app.controlledPlayerPosition()
	if !present {
		t.Fatal("normal character was not admitted")
	}

	approach := gameworld.Point{X: seam.First.X - 6, Y: seam.First.Y}

	path, err := town.FindPath(gameworld.PathRequest{
		Start: gameworld.Point{X: x, Y: y}, Goal: approach, Radius: 1,
	})
	if err != nil {
		t.Fatalf("normal town spawn cannot route to bridge: %v", err)
	}

	if len(path) <= 96 {
		t.Fatalf("normal town route is unexpectedly short: %d", len(path))
	}
	// These indices follow the owned seed-one route with clicks above the HUD.
	// The final direct approach avoids a selectable beside the bridge.
	for _, next := range []int{12, 24, 36, 48, 60, 72, 80, 88, 96} {
		clickAndAwaitBridgePosition(t, app, input, path[next], 200)

		for range 10 {
			advanceBridgePointerFrame(t, app)
		}
	}
	// This clear intermediate point brings the bridge approach back above the HUD boundary.
	visibilityStep := gameworld.Point{X: 252, Y: 84}
	clickAndAwaitBridgePosition(t, app, input, visibilityStep, 100)
	clickAndAwaitBridgePosition(t, app, input, approach, 200)
	clickBridgeEndpoint(t, app, input, gameworld.Point{X: float64(town.WidthSubtiles + 1), Y: seam.First.Y})
	awaitBridgeLevel(t, app, 2)
	assertBridgePointerLevel(t, app, 2, seam.Second)

	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	clickBridgeEndpoint(t, app, input, gameworld.Point{X: seam.Second.X, Y: seam.Second.Y})
	awaitBridgeLevel(t, app, 1)
	assertBridgePointerLevel(t, app, 1, seam.First)
}

// TestTownBridgePointerFromEntryOffline covers the normal local admission route.
func TestTownBridgePointerFromEntryOffline(t *testing.T) {
	testTownBridgePointerFromEntry(t, false)
}

// TestTownBridgePointerFromEntryConnected covers normal admission through a listen host.
func TestTownBridgePointerFromEntryConnected(t *testing.T) {
	testTownBridgePointerFromEntry(t, true)
}

// TestTownBridgePointerUnrelatedBorderStaysInTown rejects an outside-map click away from the bridge.
func TestTownBridgePointerUnrelatedBorderStaysInTown(t *testing.T) {
	app, input := newBridgePointerClient(t, false)
	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	town := app.currentWorld().Map

	x, y, present := app.controlledPlayerPosition()
	if !present {
		t.Fatal("bridge player was not admitted")
	}

	target := gameworld.Point{X: float64(town.WidthSubtiles + 1), Y: y - 12}

	cursor := town.Coordinates().SubtileToScreen(target, gameworld.Point{X: x, Y: y},
		gameworld.Point{X: 400, Y: 330})
	if cursor.X < 0 || cursor.X >= 800 || cursor.Y < 0 || cursor.Y >= 480 {
		t.Fatalf("unrelated border target is outside gameplay viewport: %v", cursor)
	}

	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Pressed: true, Down: true}}}

	advanceBridgePointerFrame(t, app)

	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y}

	for range 100 {
		advanceBridgePointerFrame(t, app)
	}

	if app.activeWorldLevel != 1 || app.playerControl.HasMoveTarget() {
		t.Fatalf("non-exit border click changed level to %d or retained impossible route", app.activeWorldLevel)
	}
}

// TestTownBridgeHeldPointerBeyondEastLip keeps targeting the exit as the camera follows the hero.
func TestTownBridgeHeldPointerBeyondEastLip(t *testing.T) {
	app, input := newBridgePointerClient(t, false)
	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	seam := app.transitionSeams[0]
	town := app.currentWorld().Map

	x, y, present := app.controlledPlayerPosition()
	if !present {
		t.Fatal("bridge player was not admitted")
	}

	cursor := town.Coordinates().SubtileToScreen(
		gameworld.Point{X: float64(town.WidthSubtiles + 1), Y: seam.First.Y},
		gameworld.Point{X: x, Y: y}, gameworld.Point{X: 400, Y: 330})
	if cursor.X < 0 || cursor.X >= 800 || cursor.Y < 0 || cursor.Y >= 480 {
		t.Fatalf("beyond-lip held cursor is outside gameplay viewport: %v", cursor)
	}

	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Pressed: true, Down: true}}}

	advanceBridgePointerFrame(t, app)

	input.frame = inputstate.Frame{CursorX: cursor.X, CursorY: cursor.Y,
		Actions: map[string]inputstate.ActionState{"pointer_primary": {Down: true}}}

	for range 200 {
		advanceBridgePointerFrame(t, app)

		if app.activeWorldLevel == 2 {
			input.frame = inputstate.Frame{}

			assertBridgePointerLevel(t, app, 2, seam.Second)

			return
		}
	}

	x, y, _ = app.controlledPlayerPosition()
	t.Fatalf("held beyond-lip click stopped in town at (%.2f,%.2f)", x, y)
}

// The northmost traversable bridge lane must enter the published transition trigger.
func TestTownBridgeNorthLaneCrossesOffline(t *testing.T) {
	app, input := newBridgePointerClient(t, false)
	for range 10 {
		advanceBridgePointerFrame(t, app)
	}

	seam := app.transitionSeams[0]
	town := app.currentWorld().Map
	start := gameworld.Point{X: 269, Y: 82}
	outer := gameworld.Point{X: float64(town.WidthSubtiles - 2), Y: 82}

	path, err := town.FindPath(gameworld.PathRequest{Start: start, Goal: outer, Radius: 1})
	if err != nil {
		t.Fatalf("north bridge lane is not walkable: %v", err)
	}

	for _, point := range path {
		if point.Y != 82 {
			t.Fatalf("north bridge route left its lane at %v", point)
		}
	}

	if math.Hypot(start.X-seam.First.X, start.Y-seam.First.Y) <= seam.TriggerRadius {
		t.Fatal("north bridge lane begins inside the transition trigger")
	}

	if err := app.playerControl.SetMoveTarget(start.X, start.Y); err != nil {
		t.Fatal(err)
	}

	for range 200 {
		advanceBridgePointerFrame(t, app)

		x, y, present := app.controlledPlayerPosition()
		if present && math.Hypot(x-start.X, y-start.Y) < 0.35 {
			break
		}
	}

	x, y, present := app.controlledPlayerPosition()
	if !present || math.Hypot(x-start.X, y-start.Y) >= 0.35 || app.activeWorldLevel != 1 {
		t.Fatalf("north lane setup stopped at (%.2f,%.2f), level %d", x, y, app.activeWorldLevel)
	}

	clickBridgeEndpoint(t, app, input, outer)

	maxX, minDistance := x, math.Hypot(x-seam.First.X, y-seam.First.Y)

	for range 200 {
		advanceBridgePointerFrame(t, app)

		if app.activeWorldLevel == 2 {
			assertBridgePointerLevel(t, app, 2, seam.Second)
			return
		}

		x, y, _ = app.controlledPlayerPosition()
		maxX = math.Max(maxX, x)

		minDistance = math.Min(minDistance, math.Hypot(x-seam.First.X, y-seam.First.Y))
		if math.Abs(y-82) > 0.5 {
			t.Fatalf("north bridge movement left lane at (%.2f,%.2f)", x, y)
		}
	}

	t.Fatalf("north bridge lane stayed in town through X %.2f; closest trigger distance %.2f, radius %.2f",
		maxX, minDistance, seam.TriggerRadius)
}
