package clientapp

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/modcache"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// stopPreviousExtensionComponents removes extension behavior before its Lua
// modules and VFS layers change, preventing old code from executing against new content.
func (app *application) stopPreviousExtensionComponents(
	ctx context.Context,
	resolved modcache.ResolvedSet,
) error {
	desired := make(map[string]bool)

	entrypoints := append(
		append([]string(nil), resolved.Base.Manifest.Entrypoints.ClientComponents...),
		resolved.Base.Manifest.Entrypoints.AuthorityComponents...,
	)

	for _, id := range entrypoints {
		desired[id] = true
	}

	return app.components.ApplyDesired(ctx, desired)
}

// replacePackageModules swaps the require allowlist first, then invalidates only
// changed namespaces so unchanged modules preserve state and startup cost.
func (app *application) replacePackageModules(
	ctx context.Context,
	plan *networkPackagePlan,
) error {
	plan.changed = changedPackageIDs(app.packageDigests, plan.resolved)
	app.packageRegistry.Replace(resolvedPackageIDs(plan.resolved))

	if len(plan.changed) == 0 {
		return nil
	}

	return modruntime.InvalidatePackageModules(ctx, app.scripts, plan.changed...)
}

// resolvedPackageIDs preserves resolved order because module visibility and
// component activation must agree with the authenticated lock recipe.
func resolvedPackageIDs(resolved modcache.ResolvedSet) []string {
	packages := resolved.Packages()
	ids := make([]string, 0, len(packages))

	for _, pkg := range packages {
		ids = append(ids, pkg.Manifest.ID)
	}

	return ids
}

// installNetworkPackageContent is the ownership handoff: after it succeeds,
// application—not the preparation plan—must unmount and close extension archives.
func (app *application) installNetworkPackageContent(plan *networkPackagePlan) error {
	app.unmountPreviousExtensionContent()

	if plan.mounted == nil {
		return nil
	}

	manifests := packageManifestsByID(plan.resolved.Extensions.Packages)

	for _, pkg := range plan.mounted.Packages {
		packageFS, err := modcache.NewPackageFS(manifests[pkg.ID], pkg.FS)
		if err != nil {
			return plan.abort(err)
		}

		layer := content.Layer{Name: "mod:" + pkg.ID, FS: packageFS}
		if err := app.options.Content.MountFirst(layer); err != nil {
			return plan.abort(err)
		}
	}

	// From this point shutdown and restoreConfiguredPackages own the mounted set.
	app.networkMounted = plan.mounted
	plan.mounted = nil

	return nil
}

// unmountPreviousExtensionContent removes stale extension roots before installing
// replacements so identical paths cannot resolve nondeterministically across generations.
func (app *application) unmountPreviousExtensionContent() {
	for _, pkg := range app.options.Mods.Extensions.Packages {
		app.options.Content.Unmount("mod:" + pkg.Manifest.ID)
	}

	if app.networkMounted != nil {
		_ = app.networkMounted.Close()
		app.networkMounted = nil
	}
}

// reconcileNetworkPackageDefinitions replaces definitions after the new VFS is
// live, then starts only client-domain components permitted in a connected process.
func (app *application) reconcileNetworkPackageDefinitions(
	ctx context.Context,
	resolved modcache.ResolvedSet,
) error {
	definitionsByPackage, _, err := app.discoverPackageDefinitions(ctx)
	if err != nil {
		return err
	}

	for _, pkg := range resolved.Extensions.Packages {
		if app.packageDigests[pkg.Manifest.ID] == pkg.Descriptor.Digest {
			continue
		}

		if err := app.reconcileManagedDefinitions(
			ctx,
			definitionsByPackage[pkg.Manifest.ID],
		); err != nil {
			return err
		}
	}

	return app.activateNetworkClientComponents(ctx)
}

// refreshPackageDerivedContent rebuilds records, maps, and bootstrap data from one
// newly installed VFS generation before presentation resumes.
func (app *application) refreshPackageDerivedContent() error {
	if _, err := app.options.Content.Invalidate("."); err != nil {
		return err
	}

	if err := app.refreshPackageRecords(); err != nil {
		return err
	}

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

// refreshPackageRecords closes the previous record generation only after the
// replacement opens, avoiding a window where consumers have no catalog.
func (app *application) refreshPackageRecords() error {
	pinned, _, err := recordstore.Pin(app.options.Content)
	if err != nil && !errors.Is(err, recordstore.ErrNoAuthoritativeTables) {
		return err
	}

	if err == nil {
		pinned.SetLogger(slog.Default().With("component", "records"))
		app.records = pinned

		return nil
	}

	app.records = recordstore.New(app.options.Content)

	return nil
}

// changedPackageIDs includes additions, removals, and digest replacements because
// all three can leave stale require results under the same namespace.
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

// packageDigestMap captures the minimal comparison state needed for future module invalidation.
func packageDigestMap(set modcache.ResolvedSet) map[string]string {
	result := make(map[string]string, 1+len(set.Extensions.Packages))

	for _, pkg := range set.Packages() {
		result[pkg.Manifest.ID] = pkg.Descriptor.Digest
	}

	return result
}

// uniqueStrings preserves first-seen lock order while preventing repeated stop,
// invalidation, or activation work for a namespace listed through multiple paths.
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
