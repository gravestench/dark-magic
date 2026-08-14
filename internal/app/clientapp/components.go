package clientapp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/host"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// loadScriptComponents finds Lua components and turns on the requested ones.
// Think of definitions as labeled toy boxes. The desired map says which boxes
// should be open right now.
func (app *application) loadScriptComponents() error {
	definitions, err := app.discoverScriptDefinitions()
	if err != nil {
		return wrap("discover Lua components", err)
	}
	if err := app.registerManagedDefinitions(definitions); err != nil {
		return err
	}
	return app.activateComponents()
}

func (app *application) discoverScriptDefinitions() ([]modruntime.Definition, error) {
	if app.options.Mods == nil {
		return modruntime.DiscoverDefinitions(context.Background(), app.scripts, app.options.Content)
	}
	var definitions []modruntime.Definition
	for _, pkg := range app.options.Mods.Packages {
		source, err := app.modSource(pkg.Manifest.ID)
		if err != nil {
			return nil, err
		}
		discovered, err := modruntime.DiscoverDefinitions(context.Background(), app.scripts, source)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, discovered...)
	}
	return definitions, nil
}

func (app *application) modSource(id string) (fs.FS, error) {
	if app.options.Content == nil {
		return nil, errors.New("resolve mod source: content filesystem is required")
	}
	if app.options.Mods == nil {
		return app.options.Content, nil
	}
	for _, pkg := range app.options.Mods.Packages {
		if pkg.Manifest.ID == id {
			source, err := fs.Sub(app.options.Content, path.Join("mods", id))
			if err != nil {
				return nil, fmt.Errorf("resolve mod source %q: %w", id, err)
			}
			return source, nil
		}
	}
	return nil, fmt.Errorf("resolve mod source %q: package is not locked", id)
}

func (app *application) registerManagedDefinitions(definitions []modruntime.Definition) error {
	for _, definition := range definitions {
		if err := app.components.Register(definition.Managed()); err != nil {
			return wrap("register Lua component "+definition.ID, err)
		}
	}
	return nil
}

func (app *application) activateComponents() error {
	var defaults []string
	if app.options.Mods != nil {
		defaults = append(app.options.Mods.AuthorityComponents(), app.options.Mods.ClientComponents()...)
	}
	desired, err := host.ParseDesired(os.Getenv("DARK_MAGIC_ENABLED_COMPONENTS"), defaults...)
	if err != nil {
		return wrap("parse enabled components", err)
	}
	if err := app.components.ApplyDesired(context.Background(), desired); err != nil {
		return wrap("start enabled components", err)
	}
	// d2legacy registers its authoritative command handlers while its managed
	// component starts. Queue bootstrap work only after that registration has
	// completed; session admission correctly rejects unknown command kinds.
	if desired == nil || desired["d2legacy.authoritative"] {
		for levelID, worldMap := range app.gameWorlds {
			if err := modruntime.SetWorldMapForLevel(context.Background(), app.scripts,
				"d2legacy.gameplay.systems.init", "set_collision", levelID, worldMap); err != nil {
				return wrap("install authoritative collision map", err)
			}
		}
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
