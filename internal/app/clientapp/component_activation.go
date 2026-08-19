package clientapp

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/host"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// activateNetworkClientComponents starts presentation/client pieces after a
// network recipe is installed; authoritative/server components must never run locally.
func (app *application) activateNetworkClientComponents(ctx context.Context) error {
	desired := make(map[string]bool)

	if app.options.Mods != nil {
		for _, id := range app.options.Mods.ClientComponents() {
			desired[id] = true
		}
	}

	return wrap(
		"activate network client components",
		app.components.ApplyDesired(ctx, desired),
	)
}

// activateComponents separates authoritative bootstrap from general component
// startup so population handlers exist before the initial world command is queued.
func (app *application) activateComponents() error {
	desired, err := app.desiredComponents()
	if err != nil {
		return err
	}

	if err := app.components.ApplyDesired(context.Background(), desired); err != nil {
		return wrap("start enabled components", err)
	}

	// Component startup registers command handlers used by bootstrap submissions.
	if err := app.bootstrapAuthoritativeComponents(desired); err != nil {
		return err
	}

	if err := app.activateDevelopmentSession(); err != nil {
		return err
	}

	if err := app.scenes.Flush(context.Background()); err != nil {
		return wrap("flush initial scene requests", err)
	}

	return app.openInitialScene()
}

// desiredComponents starts from manifest policy and applies the process override
// last, making explicit developer selection authoritative without mutating manifests.
func (app *application) desiredComponents() (map[string]bool, error) {
	var defaults []string
	if app.options.Mods != nil {
		defaults = append(
			app.options.Mods.AuthorityComponents(),
			app.options.Mods.ClientComponents()...,
		)
	}

	desired, err := host.ParseDesired(
		os.Getenv("DARK_MAGIC_ENABLED_COMPONENTS"),
		defaults...,
	)

	return desired, wrap("parse enabled components", err)
}

// bootstrapAuthoritativeComponents starts the d2legacy authority first and queues
// population only after its command handlers exist; reversing this order drops bootstrap work.
func (app *application) bootstrapAuthoritativeComponents(desired map[string]bool) error {
	if desired != nil && !desired["d2legacy.authoritative"] {
		return nil
	}

	for levelID, worldMap := range app.gameWorlds {
		err := modruntime.SetWorldMapForLevel(
			context.Background(),
			app.scripts,
			"d2legacy.gameplay.systems.init",
			"set_collision",
			levelID,
			worldMap,
		)
		if err != nil {
			return wrap("install authoritative collision map", err)
		}
	}

	return app.queueEntryPopulation()
}

// activateDevelopmentSession explicitly admits direct-start fixtures that skipped
// frontend character selection; normal player flows must still enter through the roster.
func (app *application) activateDevelopmentSession() error {
	if !shouldActivateDevelopmentSession(app.options) {
		return nil
	}

	// Direct scene entry bypasses the roster flow that normally starts selection.
	return wrap(
		"activate direct-start development session",
		app.network.StartSelected(),
	)
}

// openInitialScene opens the base scene before overlays so overlay navigation has
// a valid underlying stack and preserves caller order for deterministic captures.
func (app *application) openInitialScene() error {
	overlays := requestedOverlays(app.options.StartOverlays)
	if app.options.StartScene == "" {
		if len(overlays) > 0 {
			return fmt.Errorf("open requested overlays: --start-overlays requires --start-scene")
		}

		return nil
	}

	if err := app.navigator.Replace(context.Background(), app.options.StartScene); err != nil {
		return wrap("open requested start scene", err)
	}

	for _, overlay := range overlays {
		if err := app.navigator.Push(context.Background(), overlay); err != nil {
			return wrap("open requested start overlay "+overlay, err)
		}
	}

	return nil
}

// requestedOverlays removes blank entries without sorting because overlay order
// controls stacking and therefore visible/input behavior.
func requestedOverlays(value string) []string {
	var result []string

	for _, candidate := range strings.Split(value, ",") {
		if candidate = strings.TrimSpace(candidate); candidate != "" {
			result = append(result, candidate)
		}
	}

	return result
}
