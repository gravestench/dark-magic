// Package distribution owns product packaging defaults. Generic engine and
// mod-cache packages do not know which first-party mods ship with Dark Magic.
package distribution

import (
	"fmt"
	"strings"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

type ModSet struct {
	Profile        modcache.Profile
	ProfileCreated bool
	Resolved       modcache.ResolvedSet
	Packages       simulation.RuntimePackageSet
	Layers         []content.Layer
	mounted        *modcache.MountedSet
	Cache          *modcache.Store
}

func PrepareMods(enabledOverride ...string) (*ModSet, error) {
	baseSource := content.D2Legacy()
	base, err := modcache.DescribeBuiltin(baseSource)
	if err != nil {
		return nil, fmt.Errorf("describe built-in d2legacy: %w", err)
	}
	baseFS, err := modcache.NewPackageFS(base.Manifest, baseSource)
	if err != nil {
		return nil, fmt.Errorf("mount built-in d2legacy: %w", err)
	}
	empty := func(profile modcache.Profile, created bool) *ModSet {
		resolved := modcache.ResolvedSet{Base: base, Extensions: emptyExtensionLock()}
		return &ModSet{Profile: profile, ProfileCreated: created,
			Resolved: resolved, Packages: runtimePackages(resolved),
			Layers: []content.Layer{{Name: "builtin:d2legacy", FS: baseFS}}}
	}
	if len(enabledOverride) > 0 && strings.EqualFold(strings.TrimSpace(enabledOverride[0]), "none") {
		return empty(modcache.Profile{Schema: modcache.ProfileSchema}, false), nil
	}
	paths, err := modcache.DefaultPaths()
	if err != nil {
		return nil, err
	}
	profile, created, err := modcache.LoadOrCreateProfile(paths.Profile, nil)
	if err != nil {
		return nil, err
	}
	profile, migrated := removeBuiltinFromProfile(profile, base.Manifest.ID)
	if migrated {
		if err := modcache.SaveProfile(paths.Profile, profile); err != nil {
			return nil, fmt.Errorf("migrate built-in d2legacy out of extension profile: %w", err)
		}
	}
	if len(enabledOverride) > 0 && strings.TrimSpace(enabledOverride[0]) != "" {
		profile, err = overrideProfile(enabledOverride[0])
		if err != nil {
			return nil, err
		}
	}
	if len(profile.Enabled) == 0 {
		return empty(profile, created), nil
	}
	store, err := modcache.New(paths.Cache)
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
	result := &ModSet{Profile: profile, ProfileCreated: created, Resolved: resolved, Packages: runtimePackages(resolved), mounted: mounted, Cache: store}
	for _, pkg := range mounted.LookupOrder() {
		var manifest modcache.Manifest
		for _, locked := range resolved.Extensions.Packages {
			if locked.Manifest.ID == pkg.ID {
				manifest = locked.Manifest
				break
			}
		}
		packageFS, err := modcache.NewPackageFS(manifest, pkg.FS)
		if err != nil {
			_ = mounted.Close()
			return nil, err
		}
		result.Layers = append(result.Layers, content.Layer{Name: "mod:" + pkg.ID, FS: packageFS})
	}
	result.Layers = append(result.Layers, content.Layer{Name: "builtin:d2legacy", FS: baseFS})
	return result, nil
}

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

func runtimePackages(resolved modcache.ResolvedSet) simulation.RuntimePackageSet {
	convert := func(pkg modcache.LockedPackage) simulation.RuntimePackage {
		return simulation.RuntimePackage{ID: pkg.Manifest.ID, Version: pkg.Manifest.Version, Digest: pkg.Descriptor.Digest,
			Size: pkg.Descriptor.Size, Redistributable: pkg.Descriptor.Redistributable}
	}
	result := simulation.RuntimePackageSet{Base: convert(resolved.Base)}
	for _, extension := range resolved.Extensions.Packages {
		result.Extensions = append(result.Extensions, convert(extension))
	}
	return result
}

func emptyExtensionLock() modcache.Lock {
	// Resolve emits the digest of this exact empty representation. Keeping the
	// value explicit lets --mods none bypass all cache and profile I/O.
	return modcache.EmptyLock()
}

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

func (set *ModSet) Close() error {
	if set == nil || set.mounted == nil {
		return nil
	}
	err := set.mounted.Close()
	set.mounted = nil
	set.Layers = nil
	return err
}
