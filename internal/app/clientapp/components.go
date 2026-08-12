package clientapp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gravestench/dark-magic/internal/app/filewatch"
	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/hotreload"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// loadScriptComponents finds Lua components and turns on the requested ones.
// Think of definitions as labeled toy boxes. The desired map says which boxes
// should be open right now.
func (app *application) loadScriptComponents() error {
	definitions, err := modruntime.DiscoverDefinitions(context.Background(), app.scripts, app.options.Content)
	if err != nil {
		return wrap("discover Lua components", err)
	}
	if err := app.registerManagedDefinitions(definitions); err != nil {
		return err
	}
	modDirectory, err := app.modDirectory()
	if err != nil {
		return err
	}
	if modDirectory != "" {
		if err := app.registerHotReload(modDirectory, definitions); err != nil {
			return err
		}
	}
	return app.activateComponents(modDirectory)
}

func (app *application) registerManagedDefinitions(definitions []modruntime.Definition) error {
	for _, definition := range definitions {
		if err := app.components.Register(definition.Managed()); err != nil {
			return wrap("register Lua component "+definition.ID, err)
		}
	}
	return nil
}

func (app *application) modDirectory() (string, error) {
	path := os.Getenv("DARK_MAGIC_MOD_DIRECTORY")
	if path == "" {
		return "", nil
	}
	expanded, err := darkpaths.ExpandHost(path)
	return expanded, wrap("expand mod directory", err)
}

func (app *application) registerHotReload(directory string, definitions []modruntime.Definition) error {
	coordinator := hotreload.New(app.options.Content, app.scripts, app.components, app.records, definitions)
	definition := host.ManagedDefinition{
		ID: "engine.hot-reload",
		New: func(context.Context) (host.Component, error) {
			return filewatch.New(directory, 250*time.Millisecond, coordinator.Reload), nil
		},
	}
	return wrap("register hot reload", app.components.Register(definition))
}

func (app *application) activateComponents(modDirectory string) error {
	desired, err := host.ParseDesired(os.Getenv("DARK_MAGIC_ENABLED_COMPONENTS"), "d2legacy.authoritative", "d2legacy.boot")
	if err != nil {
		return wrap("parse enabled components", err)
	}
	if modDirectory != "" && desired != nil {
		desired["engine.hot-reload"] = true
	}
	if err := app.components.ApplyDesired(context.Background(), desired); err != nil {
		return wrap("start enabled components", err)
	}
	// d2legacy registers its authoritative command handlers while its managed
	// component starts. Queue bootstrap work only after that registration has
	// completed; session admission correctly rejects unknown command kinds.
	if desired == nil || desired["d2legacy.authoritative"] {
		if err := app.queueEntryPopulation(); err != nil {
			return err
		}
	}
	if err := app.scenes.Flush(context.Background()); err != nil {
		return wrap("flush initial scene requests", err)
	}
	if app.options.StartScene == "" {
		if len(requestedOverlays(app.options.StartOverlays)) > 0 {
			return fmt.Errorf("open requested overlays: --start-overlays requires --start-scene")
		}
		return nil
	}
	if err := app.navigator.Replace(context.Background(), app.options.StartScene); err != nil {
		return wrap("open requested start scene", err)
	}
	for _, overlay := range requestedOverlays(app.options.StartOverlays) {
		if err := app.navigator.Push(context.Background(), overlay); err != nil {
			return wrap("open requested start overlay "+overlay, err)
		}
	}
	return nil
}

func requestedOverlays(value string) []string {
	var result []string
	for _, candidate := range strings.Split(value, ",") {
		if candidate = strings.TrimSpace(candidate); candidate != "" {
			result = append(result, candidate)
		}
	}
	return result
}
