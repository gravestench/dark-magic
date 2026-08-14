// Package distribution owns product packaging defaults. Generic engine and
// mod-cache packages do not know which first-party mods ship with Dark Magic.
package distribution

import (
	"fmt"
	"strings"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/modcache"
)

type ModSet struct {
	Profile        modcache.Profile
	ProfileCreated bool
	Lock           modcache.Lock
	Layers         []content.Layer
	mounted        *modcache.MountedSet
}

func PrepareMods(enabledOverride ...string) (*ModSet, error) {
	paths, err := modcache.DefaultPaths()
	if err != nil {
		return nil, err
	}
	store, err := modcache.New(paths.Cache)
	if err != nil {
		return nil, err
	}
	defaults, err := store.ReconcileBundled([]modcache.Bundle{{Source: content.D2Legacy(), DefaultEnabled: true}})
	if err != nil {
		return nil, fmt.Errorf("prepare bundled mods: %w", err)
	}
	profile, created, err := modcache.LoadOrCreateProfile(paths.Profile, defaults)
	if err != nil {
		return nil, err
	}
	if len(enabledOverride) > 0 && strings.TrimSpace(enabledOverride[0]) != "" {
		profile, err = overrideProfile(enabledOverride[0])
		if err != nil {
			return nil, err
		}
	}
	lock, err := store.Resolve(profile)
	if err != nil {
		return nil, err
	}
	mounted, err := store.Mount(lock)
	if err != nil {
		return nil, err
	}
	result := &ModSet{Profile: profile, ProfileCreated: created, Lock: lock, mounted: mounted}
	for _, pkg := range mounted.LookupOrder() {
		var manifest modcache.Manifest
		for _, locked := range lock.Packages {
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
	return result, nil
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
