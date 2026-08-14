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

func TestPrepareModsAlwaysMountsBuiltinD2LegacyAndStartsWithNoExtensions(t *testing.T) {
	cacheDirectory, profilePath := filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "mods.json")
	t.Setenv("DARK_MAGIC_MOD_CACHE", cacheDirectory)
	t.Setenv("DARK_MAGIC_MOD_PROFILE", profilePath)

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

	empty := modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{}}
	data, _ := json.Marshal(empty)
	if err := os.WriteFile(profilePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	second, err := PrepareMods()
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if second.ProfileCreated || len(second.Resolved.Extensions.Packages) != 0 || len(second.Layers) != 1 || second.Resolved.Base.Manifest.ID != "d2legacy" {
		t.Fatalf("explicit vanilla set = %#v", second)
	}
}

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
	_ = set.Close()
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

func TestPrepareModsPersistsLegacyBuiltinProfileMigration(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "mods.json")
	t.Setenv("DARK_MAGIC_MOD_CACHE", filepath.Join(t.TempDir(), "cache"))
	t.Setenv("DARK_MAGIC_MOD_PROFILE", profilePath)
	legacy := modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{"d2legacy"}}
	data, _ := json.Marshal(legacy)
	if err := os.WriteFile(profilePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	set, err := PrepareMods()
	if err != nil {
		t.Fatal(err)
	}
	defer set.Close()
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
