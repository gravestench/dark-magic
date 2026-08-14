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

func TestPrepareModsInstallsAndEnablesD2LegacyOnlyForANewProfile(t *testing.T) {
	cacheDirectory, profilePath := filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "mods.json")
	t.Setenv("DARK_MAGIC_MOD_CACHE", cacheDirectory)
	t.Setenv("DARK_MAGIC_MOD_PROFILE", profilePath)

	first, err := PrepareMods()
	if err != nil {
		t.Fatal(err)
	}
	if !first.ProfileCreated || !reflect.DeepEqual(first.Profile.Enabled, []string{"d2legacy"}) ||
		len(first.Lock.Packages) != 1 || first.Lock.Packages[0].Manifest.ID != "d2legacy" ||
		len(first.Layers) != 1 || first.Layers[0].Name != "mod:d2legacy" {
		t.Fatalf("default bundled set = %#v", first)
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
	if second.ProfileCreated || len(second.Lock.Packages) != 0 || len(second.Layers) != 0 {
		t.Fatalf("explicit empty bundled set = %#v", second)
	}
}

func TestModOverrideDoesNotRewritePersistentProfile(t *testing.T) {
	cacheDirectory, profilePath := filepath.Join(t.TempDir(), "cache"), filepath.Join(t.TempDir(), "mods.json")
	t.Setenv("DARK_MAGIC_MOD_CACHE", cacheDirectory)
	t.Setenv("DARK_MAGIC_MOD_PROFILE", profilePath)
	set, err := PrepareMods("none")
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Lock.Packages) != 0 {
		t.Fatalf("override lock = %#v", set.Lock)
	}
	_ = set.Close()
	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	var persisted modcache.Profile
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(persisted.Enabled, []string{"d2legacy"}) {
		t.Fatalf("persisted profile = %#v", persisted)
	}
}
