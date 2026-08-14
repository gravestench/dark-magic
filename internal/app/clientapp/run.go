package clientapp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// Run builds one client, lets it run, and then takes it apart in reverse.
// Each helper below owns one small part of that story.
func Run(options Options) error {
	options = applyDevelopmentSceneDefaults(options)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	app := &application{
		options:     options,
		ctx:         ctx,
		stop:        stop,
		sceneErrors: make(chan error, 1),
		stopScene:   noCleanup,
		stopOverlay: noCleanup,
		stopCapture: noCleanup,
	}
	defer stop()

	if err := app.assemble(); err != nil {
		return errors.Join(err, app.shutdown())
	}
	runErr := app.runWindow()
	return errors.Join(runErr, app.shutdown())
}

// assemble follows the order shown on a simple stack of blocks: foundations
// first, things that use the foundations second.
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

func (app *application) runWindow() error {
	err := app.renderer.Run(app.ctx)
	select {
	case sceneErr := <-app.sceneErrors:
		err = errors.Join(err, sceneErr)
	default:
	}
	return err
}

// reportSceneError remembers the first frame error and asks the window loop to
// stop. Later errors are ignored because the first broken block is the useful
// one to report.
func (app *application) reportSceneError(err error) {
	select {
	case app.sceneErrors <- err:
		app.stop()
	default:
	}
}

func (app *application) shutdown() error {
	// Subscriptions are little callbacks held by the renderer. Remove them before
	// closing the objects they point at.
	app.stopCapture()
	app.stopScene()
	app.stopOverlay()

	var err error
	if app.network != nil {
		if app.network.Local() {
			err = errors.Join(err, persistOfflineCharacter(app.saves, app.offlineSession, "local-player"))
		}
		err = errors.Join(err, app.network.Close())
	}
	if app.saves != nil && app.playerProfilePath != "" {
		err = errors.Join(err, d2save.WriteProfileFile(app.playerProfilePath, app.saves.Profile()))
	}
	if app.capture != nil {
		err = errors.Join(err, app.capture.Close())
	}
	if app.console != nil {
		app.console.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if app.scenes != nil {
		err = errors.Join(err, app.scenes.Close(ctx))
	}
	if app.shellSession != nil {
		err = errors.Join(err, app.shellSession.Close())
	}
	if app.components != nil {
		err = errors.Join(err, app.components.ApplyDesired(ctx, map[string]bool{}))
	}
	if app.engineHost != nil && !app.hostStopped {
		err = errors.Join(err, app.engineHost.Stop(ctx))
		app.hostStopped = true
	}
	if app.loading != nil {
		app.loading.Close()
	}
	if app.offlineSession != nil {
		err = errors.Join(err, app.offlineSession.Close())
	}
	if app.clientSimulation != nil {
		err = errors.Join(err, app.clientSimulation.Close())
	}
	return err
}

func wrap(stage string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", stage, err)
}
