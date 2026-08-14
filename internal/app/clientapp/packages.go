package clientapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	d2legacy "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	"github.com/gravestench/dark-magic/internal/modcache"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

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

func cloneRuntimePackages(packages simulation.RuntimePackageSet) simulation.RuntimePackageSet {
	packages.Extensions = append([]simulation.RuntimePackage(nil), packages.Extensions...)
	return packages
}

func (app *application) restoreConfiguredPackages(ctx context.Context) error {
	if app.configuredMods.Base.ID == "" {
		return nil
	}
	source, err := app.modSource("d2legacy")
	if err != nil {
		return err
	}
	identity, err := d2legacy.IdentityForPackages(source, app.configuredMods, app.sessionInitialData())
	if err != nil {
		return err
	}
	// A failed network recomposition may have replaced only part of a package's
	// definition catalog before returning. Force every configured extension to
	// reload instead of trusting the last fully committed digest map.
	app.packageDigests = nil
	return app.recomposeForNetworkRecipe(ctx, identity.Recipe)
}

// recomposeForNetworkRecipe aligns the live client VFS, Lua namespaces, and
// managed client entrypoints with one authenticated server recipe. The built-in
// game remains distribution-owned; only exact extension blobs come from cache.
func (app *application) recomposeForNetworkRecipe(ctx context.Context, recipe simulation.RuntimeRecipe) error {
	app.recomposeMu.Lock()
	defer app.recomposeMu.Unlock()
	if err := recipe.Validate(); err != nil {
		return err
	}
	if app.options.Mods == nil || app.packageRegistry == nil {
		return errors.New("network package composition is unavailable")
	}
	if recipe.Packages.Base != app.options.Packages.Base {
		return errors.New("network recipe requires a different built-in d2legacy package")
	}
	store := app.options.ModCache
	if len(recipe.Packages.Extensions) > 0 && store == nil {
		return errors.New("network recipe requires an extension cache")
	}
	descriptors := make([]modcache.Descriptor, len(recipe.Packages.Extensions))
	for index, pkg := range recipe.Packages.Extensions {
		descriptors[index] = modcache.Descriptor{ID: pkg.ID, Version: pkg.Version, Digest: pkg.Digest,
			Size: pkg.Size, Redistributable: pkg.Redistributable}
	}
	resolved := modcache.ResolvedSet{Base: app.options.Mods.Base, Extensions: modcache.EmptyLock()}
	var mounted *modcache.MountedSet
	var err error
	if len(descriptors) > 0 {
		resolved, err = store.ResolveExact(descriptors, app.options.Mods.Base)
		if err != nil {
			return fmt.Errorf("resolve authenticated network recipe: %w", err)
		}
		mounted, err = store.Mount(resolved.Extensions, resolved.Base)
		if err != nil {
			return fmt.Errorf("mount authenticated network recipe: %w", err)
		}
	}
	if !resolvedMatchesPackages(resolved, recipe.Packages) {
		if mounted != nil {
			_ = mounted.Close()
		}
		return errors.New("resolved extension manifests differ from the authenticated network recipe")
	}
	// Compile the complete replacement set before changing a live component or
	// VFS layer. Structural definition and lifecycle validation still runs in
	// the production VM after composition, where the real capabilities exist.
	if err := validateMountedPackageSyntax(mounted, resolved); err != nil {
		if mounted != nil {
			_ = mounted.Close()
		}
		return err
	}
	// Stop old extension instances while their modules and content are still
	// mounted. Keep the built-in offline/client pair alive until final network
	// domain activation so the frontend does not lose its base lifecycle midway.
	baseDesired := make(map[string]bool)
	for _, id := range append(append([]string(nil), resolved.Base.Manifest.Entrypoints.ClientComponents...), resolved.Base.Manifest.Entrypoints.AuthorityComponents...) {
		baseDesired[id] = true
	}
	if err := app.components.ApplyDesired(ctx, baseDesired); err != nil {
		if mounted != nil {
			_ = mounted.Close()
		}
		return wrap("stop previous extension components", err)
	}

	changed := changedPackageIDs(app.packageDigests, resolved)
	packageIDs := make([]string, 0, 1+len(resolved.Extensions.Packages))
	for _, pkg := range resolved.Packages() {
		packageIDs = append(packageIDs, pkg.Manifest.ID)
	}
	app.packageRegistry.Replace(packageIDs)
	if len(changed) > 0 {
		if err := modruntime.InvalidatePackageModules(ctx, app.scripts, changed...); err != nil {
			if mounted != nil {
				_ = mounted.Close()
			}
			return err
		}
	}

	for _, pkg := range app.options.Mods.Extensions.Packages {
		app.options.Content.Unmount("mod:" + pkg.Manifest.ID)
	}
	if app.networkMounted != nil {
		_ = app.networkMounted.Close()
		app.networkMounted = nil
	}
	if mounted != nil {
		manifestByID := make(map[string]modcache.Manifest, len(resolved.Extensions.Packages))
		for _, pkg := range resolved.Extensions.Packages {
			manifestByID[pkg.Manifest.ID] = pkg.Manifest
		}
		for _, pkg := range mounted.Packages {
			packageFS, packageErr := modcache.NewPackageFS(manifestByID[pkg.ID], pkg.FS)
			if packageErr != nil {
				_ = mounted.Close()
				return packageErr
			}
			if mountErr := app.options.Content.MountFirst(content.Layer{Name: "mod:" + pkg.ID, FS: packageFS}); mountErr != nil {
				_ = mounted.Close()
				return mountErr
			}
		}
		app.networkMounted = mounted
	}
	app.options.Mods = &resolved
	app.options.Packages = recipe.Packages
	if err := app.refreshPackageDerivedContent(); err != nil {
		return err
	}

	definitionsByPackage, _, err := app.discoverPackageDefinitions(ctx)
	if err != nil {
		return err
	}
	for _, pkg := range resolved.Extensions.Packages {
		if app.packageDigests[pkg.Manifest.ID] == pkg.Descriptor.Digest {
			continue
		}
		if err := app.reconcileManagedDefinitions(ctx, definitionsByPackage[pkg.Manifest.ID]); err != nil {
			return err
		}
	}
	if err := app.activateNetworkClientComponents(ctx); err != nil {
		return err
	}
	app.packageDigests = packageDigestMap(resolved)
	return nil
}

func validateMountedPackageSyntax(mounted *modcache.MountedSet, resolved modcache.ResolvedSet) error {
	if mounted == nil {
		return nil
	}
	manifests := make(map[string]modcache.Manifest, len(resolved.Extensions.Packages))
	for _, pkg := range resolved.Extensions.Packages {
		manifests[pkg.Manifest.ID] = pkg.Manifest
	}
	for _, pkg := range mounted.Packages {
		packageFS, err := modcache.NewPackageFS(manifests[pkg.ID], pkg.FS)
		if err != nil {
			return fmt.Errorf("prepare network extension %q: %w", pkg.ID, err)
		}
		if err := modruntime.ValidatePackageSyntax(packageFS); err != nil {
			return fmt.Errorf("validate network extension %q: %w", pkg.ID, err)
		}
	}
	return nil
}

func (app *application) refreshPackageDerivedContent() error {
	if _, err := app.options.Content.Invalidate("."); err != nil {
		return err
	}
	app.records.InvalidateAll()
	app.questCatalog.Invalidate()
	app.locale.Invalidate()
	recoveredData, err := app.questCatalog.Snapshot()
	if err != nil {
		return wrap("reload recovered data after package change", err)
	}
	if err := app.worldObjectResolver.Update(recoveredData, app.records); err != nil {
		return err
	}
	if err := app.buildEntryWorld(); err != nil {
		return wrap("rebuild client world after package change", err)
	}
	return nil
}

func resolvedMatchesPackages(resolved modcache.ResolvedSet, packages simulation.RuntimePackageSet) bool {
	locked := resolved.Packages()
	wanted := append([]simulation.RuntimePackage{packages.Base}, packages.Extensions...)
	if len(locked) != len(wanted) {
		return false
	}
	for index, pkg := range locked {
		candidate := simulation.RuntimePackage{ID: pkg.Manifest.ID, Version: pkg.Manifest.Version, Digest: pkg.Descriptor.Digest,
			Size: pkg.Descriptor.Size, Redistributable: pkg.Descriptor.Redistributable}
		if candidate != wanted[index] {
			return false
		}
	}
	return true
}

func changedPackageIDs(previous map[string]string, next modcache.ResolvedSet) []string {
	now := packageDigestMap(next)
	var changed []string
	for id, digest := range previous {
		if now[id] != digest {
			changed = append(changed, id)
		}
	}
	for id, digest := range now {
		if previous[id] != digest {
			changed = append(changed, id)
		}
	}
	return uniqueStrings(changed)
}

func packageDigestMap(set modcache.ResolvedSet) map[string]string {
	result := make(map[string]string, 1+len(set.Extensions.Packages))
	for _, pkg := range set.Packages() {
		result[pkg.Manifest.ID] = pkg.Descriptor.Digest
	}
	return result
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
