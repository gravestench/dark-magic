package clientapp

import (
	"errors"
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// networkPackagePlan owns a resolved recipe and unopened live-state changes.
type networkPackagePlan struct {
	resolved modcache.ResolvedSet
	mounted  *modcache.MountedSet
	changed  []string
}

// abort closes package archives that have not transferred to the application.
func (plan *networkPackagePlan) abort(err error) error {
	if plan.mounted != nil {
		_ = plan.mounted.Close()
		plan.mounted = nil
	}

	return err
}

// validateNetworkPackageRecipe checks compatibility before cache access.
func (app *application) validateNetworkPackageRecipe(recipe simulation.RuntimeRecipe) error {
	if err := recipe.Validate(); err != nil {
		return err
	}

	if err := validateLocalAssetSet(recipe, app.options.AssetSetID); err != nil {
		return err
	}

	if app.options.Mods == nil || app.packageRegistry == nil {
		return errors.New("network package composition is unavailable")
	}

	if recipe.Packages.Base != app.options.Packages.Base {
		return errors.New("network recipe requires a different built-in d2legacy package")
	}

	if len(recipe.Packages.Extensions) > 0 && app.options.ModCache == nil {
		return errors.New("network recipe requires an extension cache")
	}

	return nil
}

// prepareNetworkPackagePlan resolves, mounts, and validates extension archives.
func (app *application) prepareNetworkPackagePlan(
	recipe simulation.RuntimeRecipe,
) (*networkPackagePlan, error) {
	plan := &networkPackagePlan{
		resolved: modcache.ResolvedSet{
			Base:       app.options.Mods.Base,
			Extensions: modcache.EmptyLock(),
		},
	}

	descriptors := runtimePackageDescriptors(recipe.Packages.Extensions)
	if len(descriptors) > 0 {
		resolved, err := app.options.ModCache.ResolveExact(descriptors, app.options.Mods.Base)
		if err != nil {
			return nil, fmt.Errorf("resolve authenticated network recipe: %w", err)
		}

		plan.resolved = resolved

		plan.mounted, err = app.options.ModCache.Mount(resolved.Extensions, resolved.Base)
		if err != nil {
			return nil, fmt.Errorf("mount authenticated network recipe: %w", err)
		}
	}

	if !resolvedMatchesPackages(plan.resolved, recipe.Packages) {
		err := errors.New("resolved extension manifests differ from the authenticated network recipe")

		return nil, plan.abort(err)
	}

	// Syntax validation completes before any live component, module, or VFS mutation.
	if err := validateMountedPackageSyntax(plan.mounted, plan.resolved); err != nil {
		return nil, plan.abort(err)
	}

	return plan, nil
}

// runtimePackageDescriptors translates transport-neutral recipe entries for the cache.
func runtimePackageDescriptors(packages []simulation.RuntimePackage) []modcache.Descriptor {
	descriptors := make([]modcache.Descriptor, len(packages))

	for index, pkg := range packages {
		descriptors[index] = modcache.Descriptor{
			ID:              pkg.ID,
			Version:         pkg.Version,
			Digest:          pkg.Digest,
			Size:            pkg.Size,
			Redistributable: pkg.Redistributable,
		}
	}

	return descriptors
}

// validateLocalAssetSet prevents packages from masking incompatible game data.
func validateLocalAssetSet(recipe simulation.RuntimeRecipe, localAssetSetID string) error {
	if recipe.AssetSetID != localAssetSetID {
		return errors.New("network recipe requires a different external game-asset set")
	}

	return nil
}

// validateMountedPackageSyntax compiles every extension before live mutation.
func validateMountedPackageSyntax(
	mounted *modcache.MountedSet,
	resolved modcache.ResolvedSet,
) error {
	if mounted == nil {
		return nil
	}

	manifests := packageManifestsByID(resolved.Extensions.Packages)

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

// resolvedMatchesPackages compares cache output with the authenticated recipe.
func resolvedMatchesPackages(
	resolved modcache.ResolvedSet,
	packages simulation.RuntimePackageSet,
) bool {
	locked := resolved.Packages()

	wanted := append([]simulation.RuntimePackage{packages.Base}, packages.Extensions...)

	if len(locked) != len(wanted) {
		return false
	}

	for index, pkg := range locked {
		candidate := simulation.RuntimePackage{
			ID:              pkg.Manifest.ID,
			Version:         pkg.Manifest.Version,
			Digest:          pkg.Descriptor.Digest,
			Size:            pkg.Descriptor.Size,
			Redistributable: pkg.Descriptor.Redistributable,
		}
		if candidate != wanted[index] {
			return false
		}
	}

	return true
}

// packageManifestsByID indexes resolved manifests for mounted archive wrapping.
func packageManifestsByID(packages []modcache.LockedPackage) map[string]modcache.Manifest {
	manifests := make(map[string]modcache.Manifest, len(packages))

	for _, pkg := range packages {
		manifests[pkg.Manifest.ID] = pkg.Manifest
	}

	return manifests
}
