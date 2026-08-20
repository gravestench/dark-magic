package clientapp

import (
	"errors"
	"fmt"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// networkPackagePlan holds validated mounts before ownership transfers to application.
// Until installation succeeds, abort must close every archive it owns.
type networkPackagePlan struct {
	resolved modcache.ResolvedSet
	mounted  *modcache.MountedSet
	changed  []string
}

// abort preserves the triggering error while releasing untransferred archives,
// preventing preparation failures from leaking file handles or cache leases.
func (plan *networkPackagePlan) abort(err error) error {
	if plan.mounted != nil {
		_ = plan.mounted.Close()
		plan.mounted = nil
	}

	return err
}

// validateNetworkPackageRecipe rejects identity and asset incompatibility before
// downloading or opening extension archives, keeping untrusted recipes cheap to reject.
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

// prepareNetworkPackagePlan performs every fallible non-live operation first:
// cache resolution, mount construction, metadata comparison, and Lua syntax validation.
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

// runtimePackageDescriptors converts authenticated wire metadata without adding
// local defaults that could change the server-declared package identity.
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

// validateLocalAssetSet ensures extension compatibility cannot hide a mismatch in
// the player's external game assets, which packages do not carry or replace.
func validateLocalAssetSet(recipe simulation.RuntimeRecipe, localAssetSetID string) error {
	if recipe.AssetSetID != localAssetSetID {
		return errors.New("network recipe requires a different external game-asset set")
	}

	return nil
}

// validateMountedPackageSyntax catches broken entrypoints while the previous
// component set is still running, avoiding a preventable half-recomposed client.
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

// resolvedMatchesPackages requires exact ordered ID, version, digest, and archive
// metadata so a cache hit cannot substitute merely namespace-compatible content.
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

// packageManifestsByID lets each mounted archive retain its own manifest-derived
// filesystem policy rather than consulting the aggregate resolved set.
func packageManifestsByID(packages []modcache.LockedPackage) map[string]modcache.Manifest {
	manifests := make(map[string]modcache.Manifest, len(packages))

	for _, pkg := range packages {
		manifests[pkg.Manifest.ID] = pkg.Manifest
	}

	return manifests
}
