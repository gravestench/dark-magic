package clientapp

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// Run assembles one client, runs its window, and releases owned resources.
func Run(options Options) error {
	app := newApplication(options)
	defer app.stop()

	if err := app.assemble(); err != nil {
		return errors.Join(err, app.shutdown())
	}

	return errors.Join(app.runWindow(), app.shutdown())
}

// newApplication applies defaults and installs process-signal cancellation.
func newApplication(options Options) *application {
	options = applyDevelopmentSceneDefaults(options)
	if options.AssetSetID == "" {
		options.AssetSetID = simulation.EmptyAssetSetID
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	return &application{
		options:     options,
		ctx:         ctx,
		stop:        stop,
		sceneErrors: make(chan error, 1),
		stopScene:   noCleanup,
		stopOverlay: noCleanup,
		stopCapture: noCleanup,
	}
}

// runWindow runs the renderer and joins the first asynchronous scene failure.
func (app *application) runWindow() error {
	err := app.renderer.Run(app.ctx)
	select {
	case sceneErr := <-app.sceneErrors:
		return errors.Join(err, sceneErr)
	default:
		return err
	}
}

// reportSceneError records the first frame error and stops the window loop.
func (app *application) reportSceneError(err error) {
	select {
	case app.sceneErrors <- err:
		app.stop()
	default:
		// The first error identifies the earliest broken frame stage.
	}
}

// assemble builds dependencies before the consumers that rely on them.
func (app *application) assemble() error {
	steps := []func() error{
		app.loadSettings,
		app.buildPresentationCore,
		app.loadGameCatalogs,
		app.buildOfflineSession,
		app.registerLuaRuntime,
		app.startEngineHost,
		app.startDeveloperConsole,
		app.attachFramePipeline,
		app.loadScriptComponents,
		app.startCapture,
	}

	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}
