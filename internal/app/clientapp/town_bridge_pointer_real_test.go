package clientapp

import (
	"context"
	"strings"
	"testing"
	"time"

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
func newBridgePointerClient(t *testing.T, hosted bool) (*application, *bridgeTestInput) {
	t.Helper()
	requireRealD2LegacyAssets(t)
	options := realD2LegacyOptions(t)
	options.StartScene = "game_world"
	options.FixtureCharacters = 1
	options.FixtureWorldSpawn = "seam"
	app := newApplication(options)
	input := &bridgeTestInput{}

	t.Cleanup(func() {
		// #627 tracks camera-follow cleanup after connected replica rebinding.
		if app.scenes != nil {
			if err := app.scenes.Close(context.Background()); err != nil {
				if hosted && strings.Contains(err.Error(), `unregister "d2legacy.world.camera_follow"`) &&
					!strings.Contains(err.Error(), "\n") {
					t.Logf("known #627 scene cleanup failure: %v", err)
				} else {
					t.Errorf("close bridge scene: %v", err)
				}
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
	if cursor.X < 0 || cursor.X >= 800 || cursor.Y < 0 || cursor.Y >= 600 {
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
