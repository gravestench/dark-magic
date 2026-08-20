// Package distribution owns product packaging defaults. Generic engine and
// mod-cache packages do not know which first-party mods ship with Dark Magic.
package distribution

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

// ModSet is the complete content selection prepared for one application run.
// Close must be called when mounted extensions are present so their archive handles do not outlive the session.
type ModSet struct {
	Profile        modcache.Profile
	ProfileCreated bool
	Resolved       modcache.ResolvedSet
	Packages       simulation.RuntimePackageSet
	Layers         []content.Layer
	mounted        *modcache.MountedSet
	Cache          *modcache.Store
}

// PrepareMods resolves the product-owned base and any selected extensions into runtime and VFS views.
// The optional override is intentionally evaluated before profile access so "none" can recover from broken state.
func PrepareMods(enabledOverride ...string) (*ModSet, error) {
	base, baseFS, err := prepareBuiltinMod()
	if err != nil {
		return nil, err
	}

	if len(enabledOverride) > 0 && strings.EqualFold(strings.TrimSpace(enabledOverride[0]), "none") {
		profile := modcache.Profile{Schema: modcache.ProfileSchema}
		return builtinOnlyModSet(profile, false, base, baseFS), nil
	}

	paths, err := modcache.DefaultPaths()
	if err != nil {
		return nil, err
	}

	profile, created, err := loadExtensionProfile(paths.Profile, base.Manifest.ID, enabledOverride)
	if err != nil {
		return nil, err
	}

	if len(profile.Enabled) == 0 {
		return builtinOnlyModSet(profile, created, base, baseFS), nil
	}

	return prepareExtensionModSet(paths.Cache, profile, created, base, baseFS)
}

// prepareBuiltinMod authenticates the embedded first-party package and applies its namespace policy.
// Keeping both steps together prevents callers from accidentally exposing private mod files at the shared VFS root.
func prepareBuiltinMod() (modcache.LockedPackage, fs.FS, error) {
	baseSource := content.D2Legacy()

	base, err := modcache.DescribeBuiltin(baseSource)
	if err != nil {
		return modcache.LockedPackage{}, nil, fmt.Errorf("describe built-in d2legacy: %w", err)
	}

	baseFS, err := modcache.NewPackageFS(base.Manifest, baseSource)
	if err != nil {
		return modcache.LockedPackage{}, nil, fmt.Errorf("mount built-in d2legacy: %w", err)
	}

	return base, baseFS, nil
}

// loadExtensionProfile migrates legacy base selections before applying any explicit extension override.
// Persisting the migration keeps later launches from repeatedly interpreting the base as an extension.
func loadExtensionProfile(
	profilePath string,
	builtinID string,
	enabledOverride []string,
) (modcache.Profile, bool, error) {
	profile, created, err := modcache.LoadOrCreateProfile(profilePath, nil)
	if err != nil {
		return modcache.Profile{}, false, err
	}

	profile, migrated := removeBuiltinFromProfile(profile, builtinID)
	if migrated {
		if err := modcache.SaveProfile(profilePath, profile); err != nil {
			return modcache.Profile{}, false, fmt.Errorf(
				"migrate built-in d2legacy out of extension profile: %w",
				err,
			)
		}
	}

	if len(enabledOverride) == 0 || strings.TrimSpace(enabledOverride[0]) == "" {
		return profile, created, nil
	}

	profile, err = overrideProfile(enabledOverride[0])
	if err != nil {
		return modcache.Profile{}, false, err
	}

	return profile, created, nil
}

// builtinOnlyModSet builds the fast path without creating a cache store or touching extension archives.
// This is both the default installation state and the recovery behavior promised by the "none" override.
func builtinOnlyModSet(
	profile modcache.Profile,
	created bool,
	base modcache.LockedPackage,
	baseFS fs.FS,
) *ModSet {
	resolved := modcache.ResolvedSet{Base: base, Extensions: emptyExtensionLock()}

	return &ModSet{
		Profile:        profile,
		ProfileCreated: created,
		Resolved:       resolved,
		Packages:       runtimePackages(resolved),
		Layers:         []content.Layer{{Name: "builtin:d2legacy", FS: baseFS}},
	}
}

// prepareExtensionModSet resolves, mounts, and namespaces selected extensions as one ownership boundary.
// A layer-construction failure closes every opened archive before returning, preventing descriptor leaks.
func prepareExtensionModSet(
	cachePath string,
	profile modcache.Profile,
	created bool,
	base modcache.LockedPackage,
	baseFS fs.FS,
) (*ModSet, error) {
	store, err := modcache.New(cachePath)
	if err != nil {
		return nil, err
	}

	resolved, err := store.Resolve(profile, base)
	if err != nil {
		return nil, err
	}

	mounted, err := store.Mount(resolved.Extensions, base)
	if err != nil {
		return nil, err
	}

	layers, err := extensionLayers(resolved.Extensions.Packages, mounted, baseFS)
	if err != nil {
		_ = mounted.Close()
		return nil, err
	}

	result := &ModSet{
		Profile:        profile,
		ProfileCreated: created,
		Resolved:       resolved,
		Packages:       runtimePackages(resolved),
		Layers:         layers,
		mounted:        mounted,
		Cache:          store,
	}

	return result, nil
}

// extensionLayers converts mounted archives into lookup layers while preserving dependency override order.
// The built-in layer remains last so extensions always get the first opportunity to provide shared content.
func extensionLayers(
	lockedPackages []modcache.LockedPackage,
	mounted *modcache.MountedSet,
	baseFS fs.FS,
) ([]content.Layer, error) {
	manifests := make(map[string]modcache.Manifest, len(lockedPackages))
	for _, locked := range lockedPackages {
		manifests[locked.Manifest.ID] = locked.Manifest
	}

	layers := make([]content.Layer, 0, len(lockedPackages)+1)

	for _, pkg := range mounted.LookupOrder() {
		packageFS, err := modcache.NewPackageFS(manifests[pkg.ID], pkg.FS)
		if err != nil {
			return nil, err
		}

		layers = append(layers, content.Layer{Name: "mod:" + pkg.ID, FS: packageFS})
	}

	return append(layers, content.Layer{Name: "builtin:d2legacy", FS: baseFS}), nil
}

// removeBuiltinFromProfile strips a legacy base-package entry while preserving extension order.
// The boolean lets callers avoid rewriting already-migrated profiles.
func removeBuiltinFromProfile(profile modcache.Profile, builtinID string) (modcache.Profile, bool) {
	enabled := make([]string, 0, len(profile.Enabled))
	removed := false

	for _, id := range profile.Enabled {
		if id == builtinID {
			removed = true
			continue
		}

		enabled = append(enabled, id)
	}

	if removed {
		profile.Enabled = enabled
	}

	return profile, removed
}

// runtimePackages projects cache locks into the transport-safe identity used by simulations and replays.
// Copying only immutable descriptor fields prevents cache implementation details from crossing that boundary.
func runtimePackages(resolved modcache.ResolvedSet) simulation.RuntimePackageSet {
	result := simulation.RuntimePackageSet{Base: runtimePackage(resolved.Base)}
	for _, extension := range resolved.Extensions.Packages {
		result.Extensions = append(result.Extensions, runtimePackage(extension))
	}

	return result
}

// runtimePackage converts one authenticated cache package without retaining references to mutable cache state.
func runtimePackage(pkg modcache.LockedPackage) simulation.RuntimePackage {
	return simulation.RuntimePackage{
		ID:              pkg.Manifest.ID,
		Version:         pkg.Manifest.Version,
		Digest:          pkg.Descriptor.Digest,
		Size:            pkg.Descriptor.Size,
		Redistributable: pkg.Descriptor.Redistributable,
	}
}

// emptyExtensionLock returns the canonical digest-bearing empty lock expected by runtime identity checks.
func emptyExtensionLock() modcache.Lock {
	// Resolve emits the digest of this exact empty representation. Keeping the
	// value explicit lets --mods none bypass all cache and profile I/O.
	return modcache.EmptyLock()
}

// overrideProfile parses a command-line extension list without consulting or mutating the persisted profile.
// Validation rejects empty or duplicate identifiers before cache resolution can produce confusing lookup failures.
func overrideProfile(value string) (modcache.Profile, error) {
	profile := modcache.Profile{Schema: modcache.ProfileSchema}
	if strings.EqualFold(strings.TrimSpace(value), "none") {
		return profile, nil
	}

	for _, raw := range strings.Split(value, ",") {
		profile.Enabled = append(profile.Enabled, strings.TrimSpace(raw))
	}

	if err := modcache.ValidateProfile(profile); err != nil {
		return modcache.Profile{}, fmt.Errorf("invalid mod override: %w", err)
	}

	return profile, nil
}

// Close releases extension archives and clears dependent layers so a closed set cannot be reused accidentally.
// Built-in-only sets own no external handles and therefore close as a no-op.
func (set *ModSet) Close() error {
	if set == nil || set.mounted == nil {
		return nil
	}

	err := set.mounted.Close()
	set.mounted = nil
	set.Layers = nil

	return err
}
