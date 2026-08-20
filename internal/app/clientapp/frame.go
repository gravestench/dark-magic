package clientapp

import (
	"context"
	"fmt"
	"runtime/pprof"
	"time"

	"github.com/gravestench/dark-magic/internal/inputstate"
)

// attachFramePipeline connects the three things that happen every picture:
// update the game, draw the game, then draw the developer console on top.
func (app *application) attachFramePipeline() error {
	app.lastFrame = time.Now()

	app.stopScene = app.renderer.SubscribeFrame(app.updateFrame)
	if err := app.renderer.AttachAudio(app.mixer); err != nil {
		return wrap("attach audio mixer", err)
	}

	if err := app.renderer.AttachComposer(app.composer); err != nil {
		return wrap("attach render composer", err)
	}

	app.stopOverlay = app.renderer.SubscribeOverlay(app.drawConsole)

	return nil
}

// updateFrame advances native simulation before Lua presentation so rendered
// state belongs to the current authority tick.
func (app *application) updateFrame() {
	started := time.Now()

	var simulationWork, luaWork time.Duration

	scene, ok := app.navigator.Focused()
	if !ok {
		scene = "none"
	}

	frameContext := app.scenes.FrameContext(context.Background())
	pprof.SetGoroutineLabels(frameContext)
	app.publishInput(frameContext)

	// Measure wall-clock interval independently from work duration: the first reveals
	// scheduler/render pacing while the second identifies work owned by this callback.
	now := time.Now()
	elapsed := now.Sub(app.lastFrame)

	app.lastFrame = now
	defer func() { app.frameMetrics.Record(scene, elapsed, time.Since(started), simulationWork, luaWork) }()

	simulationStarted := time.Now()

	if err := app.advanceGame(elapsed); err != nil {
		app.reportSceneError(err)
		return
	}

	simulationWork = time.Since(simulationStarted)

	luaStarted := time.Now()

	updated, err := app.updateAndRenderScenes(frameContext, elapsed)
	if err != nil {
		app.reportSceneError(err)
	}

	if !updated {
		return
	}

	luaWork = time.Since(luaStarted)
}

// updateAndRenderScenes refreshes profiling ownership after update because Lua may replace the focused scene mid-frame.
// The boolean distinguishes update failure, which historically records no Lua work, from a render failure after update.
func (app *application) updateAndRenderScenes(
	frameContext context.Context,
	elapsed time.Duration,
) (bool, error) {
	if err := app.scenes.Update(frameContext, elapsed); err != nil {
		return false, fmt.Errorf("updating Lua scenes: %w", err)
	}

	// A scene may replace itself while updating. Ask again who owns the frame
	// before drawing so profiling labels name the new scene, not the old one.
	renderContext := app.scenes.FrameContext(context.Background())
	pprof.SetGoroutineLabels(renderContext)

	if err := app.scenes.Render(renderContext); err != nil {
		return true, fmt.Errorf("rendering Lua scenes: %w", err)
	}

	return true, nil
}

// publishInput routes one immutable input snapshot to the console or the active
// gameplay scene, never both accidentally.
func (app *application) publishInput(frameContext context.Context) {
	frame := app.input.Snapshot()
	if app.pointerAcceptance != nil {
		x, y, present := app.controlledPlayerPosition()
		frame = app.pointerAcceptance.Frame(frame, x, y, present)
	}

	frame.WorldSplit = float64(app.profile.Width) / 2
	owner := app.focusOwner()

	captured := app.console.Handle(frameContext, frame)
	if captured {
		owner = inputstate.FocusOwner{Domain: inputstate.FocusDebug, ID: "client-console"}
	}

	gameplayScene := app.gameplayInputScene()
	gameplayAllowed, worldView := app.navigator.InputPolicy(gameplayScene)
	app.inputState.Publish(inputstate.Route(frame, owner, captured, gameplayAllowed, worldView))
}

// gameplayInputScene lets gameplay laboratories reuse production input policy without granting it to other dev scenes.
func (app *application) gameplayInputScene() string {
	if focused, ok := app.navigator.Focused(); ok && developmentGameplayScene(focused) {
		return focused
	}

	return "game_world"
}

// focusOwner identifies which scene should receive uncaptured input; an empty navigator intentionally routes nowhere.
func (app *application) focusOwner() inputstate.FocusOwner {
	if focused, ok := app.navigator.Focused(); ok {
		return inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: focused}
	}

	return inputstate.FocusOwner{Domain: inputstate.FocusNone}
}

// advanceGame selects exactly one authority path: connected remote,
// disconnected remote placeholder, or local session.
func (app *application) advanceGame(elapsed time.Duration) error {
	if app.network != nil && app.network.Connected() {
		if err := app.network.Advance(app.ctx, elapsed); err != nil {
			return fmt.Errorf("updating remote game session: %w", err)
		}

		app.syncActiveWorldFromPlayer()

		return nil
	}

	if app.network != nil && !app.network.Local() {
		return nil
	}

	if _, err := app.offlineSession.AdvanceWithSource(elapsed, app.commandSource); err != nil {
		return fmt.Errorf("updating offline game session: %w", err)
	}

	app.syncActiveWorldFromPlayer()

	return nil
}

// drawConsole renders diagnostics after the scene so developer output remains
// visible without changing scene composition.
func (app *application) drawConsole() {
	width, height := app.renderer.WindowSize()
	app.console.Draw(width, height)
}
