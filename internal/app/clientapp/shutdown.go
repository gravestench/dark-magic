package clientapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// errorCollector retains every shutdown failure in the order it occurred.
type errorCollector struct {
	err error
}

// add joins one resource error into the accumulated shutdown result.
func (collector *errorCollector) add(err error) {
	collector.err = errors.Join(collector.err, err)
}

// shutdown releases resources in reverse ownership order and joins failures.
func (app *application) shutdown() error {
	app.stopSubscriptions()

	errors := &errorCollector{}
	app.closeConnections(errors)

	app.closeUserState(errors)

	app.closeRuntime(errors)

	return errors.err
}

// stopSubscriptions detaches renderer callbacks before closing their targets.
func (app *application) stopSubscriptions() {
	app.stopCapture()
	app.stopScene()
	app.stopOverlay()
}

// closeConnections persists local play before closing network and realm clients.
func (app *application) closeConnections(errors *errorCollector) {
	if app.network != nil {
		if app.network.Local() {
			errors.add(persistOfflineCharacter(app.saves, app.offlineSession, "local-player"))
		}

		errors.add(app.network.Close())
	}

	if app.realm == nil {
		return
	}

	// Realm closure may perform remote cleanup, so bound it independently.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	errors.add(app.realm.Close(ctx))
	cancel()
}

// closeUserState writes profiles and closes optional capture and console tools.
func (app *application) closeUserState(errors *errorCollector) {
	if app.saves != nil && app.playerProfilePath != "" {
		errors.add(d2save.WriteProfileFile(app.playerProfilePath, app.saves.Profile()))
	}

	if app.capture != nil {
		errors.add(app.capture.Close())
	}

	if app.console != nil {
		app.console.Close()
	}
}

// closeRuntime stops scene, script, component, host, and simulation resources.
func (app *application) closeRuntime(errors *errorCollector) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	app.closeRuntimeFrontends(ctx, errors)
	app.closeRuntimeBackends(ctx, errors)
}

// closeRuntimeFrontends closes consumers before their underlying host services.
func (app *application) closeRuntimeFrontends(ctx context.Context, errors *errorCollector) {
	if app.scenes != nil {
		errors.add(app.scenes.Close(ctx))
	}

	if app.shellSession != nil {
		errors.add(app.shellSession.Close())
	}

	if app.components != nil {
		errors.add(app.components.ApplyDesired(ctx, map[string]bool{}))
	}

	if app.networkMounted != nil {
		errors.add(app.networkMounted.Close())
		app.networkMounted = nil
	}
}

// closeRuntimeBackends closes the host and its remaining owned state.
func (app *application) closeRuntimeBackends(ctx context.Context, errors *errorCollector) {
	if app.engineHost != nil && !app.hostStopped {
		errors.add(app.engineHost.Stop(ctx))
		app.hostStopped = true
	}

	if app.loading != nil {
		app.loading.Close()
	}

	if app.offlineSession != nil {
		errors.add(app.offlineSession.Close())
	}

	if app.clientSimulation != nil {
		errors.add(app.clientSimulation.Close())
	}
}

// wrap identifies the composition stage that returned an error.
func wrap(stage string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", stage, err)
}
