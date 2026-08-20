package distribution

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/modcache"
)

// TestPrepareModsAlwaysMountsBuiltinD2LegacyAndStartsWithNoExtensions verifies both new and existing empty profiles.
// The shared assertions protect the base package's namespace boundary as well as its position in lookup order.
func TestPrepareModsAlwaysMountsBuiltinD2LegacyAndStartsWithNoExtensions(t *testing.T) {
	cacheDirectory, profilePath := filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "mods.json")
	t.Setenv("DARK_MAGIC_MOD_CACHE", cacheDirectory)
	t.Setenv("DARK_MAGIC_MOD_PROFILE", profilePath)

	// A missing profile is created, but it must not turn the built-in package into a cache-backed extension.
	first, err := PrepareMods()
	if err != nil {
		t.Fatal(err)
	}

	if !first.ProfileCreated || len(first.Profile.Enabled) != 0 ||
		first.Resolved.Base.Manifest.ID != "d2legacy" || len(first.Resolved.Extensions.Packages) != 0 ||
		len(first.Layers) != 1 || first.Layers[0].Name != "builtin:d2legacy" {
		t.Fatalf("default built-in set = %#v", first)
	}

	contentFS, err := content.New(first.Layers...)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := fs.ReadFile(contentFS, "mods/d2legacy/boot.lua"); err != nil {
		t.Fatalf("read namespaced boot: %v", err)
	}

	if _, err := fs.ReadFile(contentFS, "boot.lua"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("private boot leaked into shared VFS root: %v", err)
	}

	if _, err := fs.ReadFile(contentFS, "manifests/presentation.v1.json"); err != nil {
		t.Fatalf("read declared shared content: %v", err)
	}

	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	// An explicitly saved empty profile follows the same runtime path without being reported as newly created.
	empty := modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{}}

	data, err := json.Marshal(empty)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(profilePath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	second, err := PrepareMods()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := second.Close(); err != nil {
			t.Errorf("close explicit vanilla set: %v", err)
		}
	})

	if second.ProfileCreated || len(second.Resolved.Extensions.Packages) != 0 || len(second.Layers) != 1 ||
		second.Resolved.Base.Manifest.ID != "d2legacy" {
		t.Fatalf("explicit vanilla set = %#v", second)
	}
}

// TestNoneOverrideBypassesBrokenPersistentExtensionState ensures the recovery override performs no profile or cache
// I/O.
// Operators can therefore launch the built-in game even when persistent extension state is corrupt.
func TestNoneOverrideBypassesBrokenPersistentExtensionState(t *testing.T) {
	cacheDirectory, profilePath := filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "mods.json")
	t.Setenv("DARK_MAGIC_MOD_CACHE", cacheDirectory)
	t.Setenv("DARK_MAGIC_MOD_PROFILE", profilePath)

	if err := os.WriteFile(profilePath, []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}

	set, err := PrepareMods("none")
	if err != nil {
		t.Fatal(err)
	}

	if len(set.Resolved.Extensions.Packages) != 0 || set.Resolved.Base.Manifest.ID != "d2legacy" || len(set.Layers) != 1 {
		t.Fatalf("override set = %#v", set)
	}

	if err := set.Close(); err != nil {
		t.Fatal(err)
	}

	// Reading the original sentinel proves the override did not silently repair or replace operator state.
	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "broken" {
		t.Fatalf("none override rewrote profile: %q", data)
	}

	if _, err := os.Stat(cacheDirectory); !os.IsNotExist(err) {
		t.Fatalf("none override touched extension cache: %v", err)
	}
}

// TestLegacyBuiltinProfileEntryMigratesToExtensionsOnly verifies migration is ordered and idempotent.
// Preserving extension order is significant because later entries have higher content lookup priority.
func TestLegacyBuiltinProfileEntryMigratesToExtensionsOnly(t *testing.T) {
	profile := modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{"d2legacy", "example"}}

	migrated, changed := removeBuiltinFromProfile(profile, "d2legacy")
	if !changed || !reflect.DeepEqual(migrated.Enabled, []string{"example"}) {
		t.Fatalf("migrated profile = %#v, changed=%t", migrated, changed)
	}

	unchanged, changed := removeBuiltinFromProfile(migrated, "d2legacy")
	if changed || !reflect.DeepEqual(unchanged, migrated) {
		t.Fatalf("idempotent migration = %#v, changed=%t", unchanged, changed)
	}
}

// TestPrepareModsPersistsLegacyBuiltinProfileMigration ensures migration survives beyond the current process.
// Without persistence, every launch would continue treating a product-owned base as user-selected extension state.
func TestPrepareModsPersistsLegacyBuiltinProfileMigration(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "mods.json")
	t.Setenv("DARK_MAGIC_MOD_CACHE", filepath.Join(t.TempDir(), "cache"))
	t.Setenv("DARK_MAGIC_MOD_PROFILE", profilePath)

	legacy := modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{"d2legacy"}}

	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(profilePath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	set, err := PrepareMods()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := set.Close(); err != nil {
			t.Errorf("close migrated set: %v", err)
		}
	})

	if len(set.Profile.Enabled) != 0 || len(set.Resolved.Extensions.Packages) != 0 {
		t.Fatalf("migrated runtime set = %#v", set)
	}

	persisted, _, err := modcache.LoadOrCreateProfile(profilePath, []string{"unexpected"})
	if err != nil {
		t.Fatal(err)
	}

	if len(persisted.Enabled) != 0 {
		t.Fatalf("persisted migrated profile = %#v", persisted)
	}
}
