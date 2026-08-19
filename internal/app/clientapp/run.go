package clientapp

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// Run is the package ownership boundary: it assembles one application, runs the
// blocking window lifetime, and always attempts reverse-order cleanup.
func Run(options Options) error {
	app := newApplication(options)
	defer app.stop()

	if err := app.assemble(); err != nil {
		return errors.Join(err, app.shutdown())
	}

	return errors.Join(app.runWindow(), app.shutdown())
}

// newApplication establishes safe zero-state callbacks and a signal-scoped root
// context so partial assembly and OS shutdown share the same cleanup path.
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

// runWindow treats native loop exit and asynchronous scene failure as peers.
// Joining both preserves the first actionable cause without abandoning window cleanup.
func (app *application) runWindow() error {
	err := app.renderer.Run(app.ctx)
	select {
	case sceneErr := <-app.sceneErrors:
		return errors.Join(err, sceneErr)
	default:
		return err
	}
}

// reportSceneError uses first-error semantics because later frame failures are
// usually consequences; closing the window promptly prevents repeated mutation after failure.
func (app *application) reportSceneError(err error) {
	select {
	case app.sceneErrors <- err:
		app.stop()
	default:
		// The first error identifies the earliest broken frame stage.
	}
}

// assemble is intentionally top-down: durable policy precedes native services,
// authority precedes Lua components, and components precede scene activation.
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
