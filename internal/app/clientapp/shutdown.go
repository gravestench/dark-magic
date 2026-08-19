package clientapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

// errorCollector preserves all cleanup failures because one broken closer must
// not prevent later resources from being released or hide their errors.
type errorCollector struct {
	err error
}

// add ignores nil and joins non-nil failures while retaining Go's error matching behavior.
func (collector *errorCollector) add(err error) {
	collector.err = errors.Join(collector.err, err)
}

// shutdown mirrors assembly in reverse: callbacks stop before targets, clients
// persist/leave before runtimes disappear, and native backends close last.
func (app *application) shutdown() error {
	app.stopSubscriptions()

	errors := &errorCollector{}
	app.closeConnections(errors)

	app.closeUserState(errors)

	app.closeRuntime(errors)

	return errors.err
}

// stopSubscriptions prevents late resize/frame callbacks from reaching video or
// other services while those targets are being torn down.
func (app *application) stopSubscriptions() {
	app.stopCapture()
	app.stopScene()
	app.stopOverlay()
}

// closeConnections commits player state and Realm presence while session and save
// dependencies still exist; closing transports first could lose authoritative exit data.
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

// closeUserState flushes durable user-facing outputs before their content and
// presentation dependencies are released. Optional tools still participate in error reporting.
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

// closeRuntime separates frontend consumers from backend owners and gives each
// phase a bounded context so a stuck component cannot hang process exit forever.
func (app *application) closeRuntime(errors *errorCollector) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	app.closeRuntimeFrontends(ctx, errors)
	app.closeRuntimeBackends(ctx, errors)
}

// closeRuntimeFrontends stops scenes, managed scripts, and Lua before engine host
// services, preventing callbacks into backends that are already closing.
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

// closeRuntimeBackends releases host/native services before ECS engines and loading
// coordination, preserving every backend's opportunity to detach from simulation.
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

// wrap adds the failed composition stage without changing the underlying error,
// making startup/shutdown reports actionable while preserving errors.Is semantics.
func wrap(stage string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", stage, err)
}
