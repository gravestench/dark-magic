package clientapp

import (
	"context"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2legacy "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	"github.com/gravestench/dark-magic/internal/modcache"
)

// ensureModCache returns the configured cache or creates the platform default.
func (app *application) ensureModCache() (*modcache.Store, error) {
	if app.options.ModCache != nil {
		return app.options.ModCache, nil
	}

	paths, err := modcache.DefaultPaths()
	if err != nil {
		return nil, err
	}

	store, err := modcache.New(paths.Cache)
	if err != nil {
		return nil, err
	}

	app.options.ModCache = store

	return store, nil
}

// cloneRuntimePackages copies the extension slice owned by a runtime recipe.
func cloneRuntimePackages(packages simulation.RuntimePackageSet) simulation.RuntimePackageSet {
	packages.Extensions = append([]simulation.RuntimePackage(nil), packages.Extensions...)

	return packages
}

// restoreConfiguredPackages returns to the package recipe selected at startup.
func (app *application) restoreConfiguredPackages(ctx context.Context) error {
	if app.configuredMods.Base.ID == "" {
		return nil
	}

	source, err := app.modSource("d2legacy")
	if err != nil {
		return err
	}

	identity, err := d2legacy.IdentityForPackagesAndData(
		source,
		app.configuredMods,
		app.options.AssetSetID,
		app.gameDataGenerationID(),
		app.sessionInitialData(),
	)
	if err != nil {
		return err
	}

	// A failed recomposition can replace only part of the definition catalog.
	// Clearing digests forces every configured extension through reconciliation.
	app.packageDigests = nil

	return app.recomposeForNetworkRecipe(ctx, identity.Recipe)
}

// recomposeForNetworkRecipe applies one authenticated package recipe in phases.
func (app *application) recomposeForNetworkRecipe(
	ctx context.Context,
	recipe simulation.RuntimeRecipe,
) error {
	app.recomposeMu.Lock()
	defer app.recomposeMu.Unlock()

	if err := app.validateNetworkPackageRecipe(recipe); err != nil {
		return err
	}

	plan, err := app.prepareNetworkPackagePlan(recipe)
	if err != nil {
		return err
	}

	// Old extensions must stop while their modules and content are still mounted.
	if err := app.stopPreviousExtensionComponents(ctx, plan.resolved); err != nil {
		return plan.abort(wrap("stop previous extension components", err))
	}

	if err := app.replacePackageModules(ctx, plan); err != nil {
		return plan.abort(err)
	}

	if err := app.installNetworkPackageContent(plan); err != nil {
		return err
	}

	// Content ownership has transferred to app; later failures are restored by
	// restoreConfiguredPackages rather than closing the active mounted set here.
	app.options.Mods = &plan.resolved
	app.options.Packages = recipe.Packages

	if err := app.refreshPackageDerivedContent(); err != nil {
		return err
	}

	if err := app.reconcileNetworkPackageDefinitions(ctx, plan.resolved); err != nil {
		return err
	}

	app.packageDigests = packageDigestMap(plan.resolved)

	return nil
}
