package clientapp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"

	"github.com/gravestench/dark-magic/internal/modcache"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// loadScriptComponents completes the definition lifecycle before activation,
// ensuring dependency/ownership errors fail without leaving a partially started set.
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

// discoverScriptDefinitions uses locked package roots when available so each mod
// executes against authenticated mounted content rather than the aggregate filesystem.
func (app *application) discoverScriptDefinitions() ([]modruntime.Definition, error) {
	if app.options.Mods == nil {
		return modruntime.DiscoverDefinitions(
			context.Background(),
			app.scripts,
			app.options.Content,
		)
	}

	_, definitions, err := app.discoverPackageDefinitions(context.Background())

	return definitions, err
}

// discoverPackageDefinitions evaluates packages independently and preserves lock
// order, which is also the deterministic order used for component registration.
func (app *application) discoverPackageDefinitions(
	ctx context.Context,
) (map[string][]modruntime.Definition, []modruntime.Definition, error) {
	byPackage := make(map[string][]modruntime.Definition)

	var definitions []modruntime.Definition

	for _, pkg := range app.options.Mods.Packages() {
		discovered, err := app.discoverDefinitionsForPackage(ctx, pkg)
		if err != nil {
			return nil, nil, err
		}

		byPackage[pkg.Manifest.ID] = discovered
		definitions = append(definitions, discovered...)
	}

	if err := modruntime.ValidateDefinitionDomains(
		definitions,
		app.options.Mods.ClientComponents(),
		app.options.Mods.AuthorityComponents(),
	); err != nil {
		return nil, nil, err
	}

	return byPackage, definitions, nil
}

// discoverDefinitionsForPackage binds a definition to its package root and manifest
// dependencies, preventing one archive from claiming another package's scripts.
func (app *application) discoverDefinitionsForPackage(
	ctx context.Context,
	pkg modcache.LockedPackage,
) ([]modruntime.Definition, error) {
	source, err := app.modSource(pkg.Manifest.ID)
	if err != nil {
		return nil, err
	}

	discovered, err := modruntime.DiscoverOwnedDefinitions(
		ctx,
		app.scripts,
		source,
		pkg.Manifest.ID,
	)
	if err != nil {
		return nil, err
	}

	// Validate package boundaries before any definition reaches the live manager.
	if err := modruntime.ValidateDefinitionDependencies(
		discovered,
		pkg.Manifest.ID,
		dependencyIDs(pkg.Manifest),
	); err != nil {
		return nil, err
	}

	if err := modruntime.ValidateDefinitionEntrypoints(
		discovered,
		pkg.Manifest.Entrypoints.ClientComponents,
		pkg.Manifest.Entrypoints.AuthorityComponents,
	); err != nil {
		return nil, err
	}

	return discovered, nil
}

// dependencyIDs preserves manifest order because dependency lists also influence
// deterministic discovery diagnostics and component planning.
func dependencyIDs(manifest modcache.Manifest) []string {
	result := make([]string, len(manifest.Dependencies))
	for index, dependency := range manifest.Dependencies {
		result[index] = dependency.ID
	}

	return result
}

// modSource enforces package filesystem isolation. Falling back to distribution
// content is permitted only when no locked package set owns the requested namespace.
func (app *application) modSource(id string) (fs.FS, error) {
	if app.options.Content == nil {
		return nil, errors.New("resolve mod source: content filesystem is required")
	}

	if app.options.Mods == nil {
		return app.options.Content, nil
	}

	for _, pkg := range app.options.Mods.Packages() {
		if pkg.Manifest.ID != id {
			continue
		}
		// Package roots are mounted beneath mods/<id> in the composed content FS.
		source, err := fs.Sub(app.options.Content, path.Join("mods", id))
		if err != nil {
			return nil, fmt.Errorf("resolve mod source %q: %w", id, err)
		}

		return source, nil
	}

	return nil, fmt.Errorf("resolve mod source %q: package is not locked", id)
}

// registerManagedDefinitions creates the initial registry and identity index
// together so later hot recomposition can distinguish replacement from addition.
func (app *application) registerManagedDefinitions(definitions []modruntime.Definition) error {
	app.ensureComponentIDs()

	for _, definition := range definitions {
		if err := app.components.Register(definition.Managed()); err != nil {
			return wrap("register Lua component "+definition.ID, err)
		}

		app.componentIDs[definition.ID] = true
	}

	return nil
}

// reconcileManagedDefinitions updates definitions without restarting unchanged
// namespaces, limiting hot-recomposition disruption to packages whose digest changed.
func (app *application) reconcileManagedDefinitions(
	ctx context.Context,
	definitions []modruntime.Definition,
) error {
	app.ensureComponentIDs()

	for _, definition := range definitions {
		if err := app.reconcileManagedDefinition(ctx, definition); err != nil {
			return err
		}
	}

	return nil
}

// reconcileManagedDefinition chooses replace versus register from the explicit ID
// index; relying on registry errors would blur expected replacement with real failure.
func (app *application) reconcileManagedDefinition(
	ctx context.Context,
	definition modruntime.Definition,
) error {
	if app.componentIDs[definition.ID] {
		return wrap(
			"replace Lua component "+definition.ID,
			app.components.Replace(ctx, definition.Managed()),
		)
	}

	if err := app.components.Register(definition.Managed()); err != nil {
		return wrap("register Lua component "+definition.ID, err)
	}

	app.componentIDs[definition.ID] = true

	return nil
}

// ensureComponentIDs keeps zero-value application safe during tests and partial
// assembly while preserving one authoritative set of known component identities.
func (app *application) ensureComponentIDs() {
	if app.componentIDs == nil {
		app.componentIDs = make(map[string]bool)
	}
}
