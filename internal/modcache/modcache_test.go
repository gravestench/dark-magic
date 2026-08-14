package modcache

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBundledPackageIsCachedByDigestAndEnabledOnlyForNewProfile(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "cache"))
	if err != nil {
		t.Fatal(err)
	}
	bundle := testBundle(testManifest("base", "game"), map[string]string{"boot.lua": "return {}"})
	defaults, err := store.ReconcileBundled([]Bundle{{Source: bundle, DefaultEnabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(t.TempDir(), "mods.json")
	profile, created, err := LoadOrCreateProfile(profilePath, defaults)
	if err != nil {
		t.Fatal(err)
	}
	if !created || !reflect.DeepEqual(profile.Enabled, []string{"base"}) {
		t.Fatalf("new profile = %#v created=%t", profile, created)
	}
	lock, err := store.Resolve(profile)
	if err != nil {
		t.Fatal(err)
	}
	if len(lock.Packages) != 1 || lock.Packages[0].Manifest.ID != "base" || !validDigest(lock.Digest) {
		t.Fatalf("resolved lock = %#v", lock)
	}
	descriptor := lock.Packages[0].Descriptor
	if _, err := os.Stat(store.blobPath(descriptor.Digest)); err != nil {
		t.Fatalf("content-addressed blob: %v", err)
	}
	mounted, err := store.Mount(lock)
	if err != nil {
		t.Fatal(err)
	}
	defer mounted.Close()
	data, err := fs.ReadFile(mounted.Packages[0].FS, "boot.lua")
	if err != nil || string(data) != "return {}" {
		t.Fatalf("mounted boot = %q, %v", data, err)
	}

	// An existing empty profile is an intentional user choice. Reconciliation
	// keeps the bundled package installed but does not silently re-enable it.
	empty := Profile{Schema: ProfileSchema, Enabled: []string{}}
	data, _ = json.Marshal(empty)
	if err := os.WriteFile(profilePath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	profile, created, err = LoadOrCreateProfile(profilePath, defaults)
	if err != nil {
		t.Fatal(err)
	}
	if created || len(profile.Enabled) != 0 {
		t.Fatalf("existing empty profile = %#v created=%t", profile, created)
	}
	lock, err = store.Resolve(profile)
	if err != nil || len(lock.Packages) != 0 {
		t.Fatalf("empty lock = %#v, %v", lock, err)
	}
}

func TestResolverOrdersDependenciesBeforeDependentsAndLookupInReverse(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	base := testManifest("base", "game")
	extension := testManifest("extension", "extension")
	extension.Dependencies = []Dependency{{ID: "base", Version: base.Version}}
	if _, err := store.ReconcileBundled([]Bundle{
		{Source: testBundle(base, map[string]string{"shared.txt": "base"})},
		{Source: testBundle(extension, map[string]string{"shared.txt": "extension"})},
	}); err != nil {
		t.Fatal(err)
	}
	lock, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"extension"}})
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{lock.Packages[0].Manifest.ID, lock.Packages[1].Manifest.ID}
	if !reflect.DeepEqual(ids, []string{"base", "extension"}) {
		t.Fatalf("activation order = %v", ids)
	}
	mounted, err := store.Mount(lock)
	if err != nil {
		t.Fatal(err)
	}
	defer mounted.Close()
	lookup := mounted.LookupOrder()
	if lookup[0].ID != "extension" || lookup[1].ID != "base" {
		t.Fatalf("lookup order = %#v", lookup)
	}
}

func TestResolverRejectsMissingCyclesAndTamperedBlobs(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		store, _ := New(t.TempDir())
		if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"missing"}}); err == nil {
			t.Fatal("missing enabled package was accepted")
		}
	})
	t.Run("cycle", func(t *testing.T) {
		store, _ := New(t.TempDir())
		first, second := testManifest("first", "extension"), testManifest("second", "extension")
		first.Dependencies = []Dependency{{ID: "second"}}
		second.Dependencies = []Dependency{{ID: "first"}}
		if _, err := store.ReconcileBundled([]Bundle{{Source: testBundle(first, nil)}, {Source: testBundle(second, nil)}}); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"first"}}); err == nil || !strings.Contains(err.Error(), "cycle") {
			t.Fatalf("cycle error = %v", err)
		}
	})
	t.Run("tamper", func(t *testing.T) {
		store, _ := New(t.TempDir())
		if _, err := store.ReconcileBundled([]Bundle{{Source: testBundle(testManifest("base", "game"), nil)}}); err != nil {
			t.Fatal(err)
		}
		catalog, err := store.readIndex()
		if err != nil {
			t.Fatal(err)
		}
		descriptor := catalog.Packages["base"]
		if err := os.WriteFile(store.blobPath(descriptor.Digest), []byte("tampered"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"base"}}); err == nil {
			t.Fatal("tampered package was accepted")
		}
	})
}

func TestDisabledBrokenPackageCannotPreventStartup(t *testing.T) {
	store, _ := New(t.TempDir())
	if _, err := store.ReconcileBundled([]Bundle{
		{Source: testBundle(testManifest("enabled", "game"), nil)},
		{Source: testBundle(testManifest("disabled", "extension"), nil)},
	}); err != nil {
		t.Fatal(err)
	}
	catalog, _ := store.readIndex()
	disabled := catalog.Packages["disabled"]
	if err := os.WriteFile(store.blobPath(disabled.Digest), []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	lock, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"enabled"}})
	if err != nil || len(lock.Packages) != 1 || lock.Packages[0].Manifest.ID != "enabled" {
		t.Fatalf("enabled lock = %#v, %v", lock, err)
	}
}

func TestResolvedSetRequiresOneGamePackage(t *testing.T) {
	t.Run("extension only", func(t *testing.T) {
		store, _ := New(t.TempDir())
		if _, err := store.ReconcileBundled([]Bundle{{Source: testBundle(testManifest("extension", "extension"), nil)}}); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"extension"}}); err == nil {
			t.Fatal("extension-only set was accepted")
		}
	})
	t.Run("multiple games", func(t *testing.T) {
		store, _ := New(t.TempDir())
		if _, err := store.ReconcileBundled([]Bundle{
			{Source: testBundle(testManifest("first", "game"), nil)},
			{Source: testBundle(testManifest("second", "game"), nil)},
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"first", "second"}}); err == nil {
			t.Fatal("multiple game packages were accepted")
		}
	})
}

func TestBundledArchiveDigestIsDeterministic(t *testing.T) {
	store, _ := New(t.TempDir())
	bundle := testBundle(testManifest("base", "game"), map[string]string{"z.txt": "z", "a.txt": "a"})
	if _, err := store.ReconcileBundled([]Bundle{{Source: bundle}}); err != nil {
		t.Fatal(err)
	}
	first, _ := store.readIndex()
	if _, err := store.ReconcileBundled([]Bundle{{Source: bundle}}); err != nil {
		t.Fatal(err)
	}
	second, _ := store.readIndex()
	if first.Packages["base"].Digest != second.Packages["base"].Digest {
		t.Fatalf("bundle digest changed: %s != %s", first.Packages["base"].Digest, second.Packages["base"].Digest)
	}
}

func TestMountRejectsLockMetadataChangedAfterResolution(t *testing.T) {
	store, _ := New(t.TempDir())
	if _, err := store.ReconcileBundled([]Bundle{{Source: testBundle(testManifest("base", "game"), nil)}}); err != nil {
		t.Fatal(err)
	}
	lock, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"base"}})
	if err != nil {
		t.Fatal(err)
	}
	lock.Packages[0].Manifest.Entrypoints.ClientComponents = []string{"attacker.boot"}
	if _, err := store.Mount(lock); err == nil || !strings.Contains(err.Error(), "digest") {
		t.Fatalf("tampered lock error = %v", err)
	}
}

func TestArchiveValidationRejectsTraversalAndDuplicateEntries(t *testing.T) {
	for name, entries := range map[string][]string{
		"traversal": {"mod.json", "../outside.lua"},
		"duplicate": {"mod.json", "mod.json"},
	} {
		t.Run(name, func(t *testing.T) {
			var data bytes.Buffer
			writer := zip.NewWriter(&data)
			for _, entry := range entries {
				file, err := writer.Create(entry)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = file.Write([]byte("{}"))
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			archive, err := zip.NewReader(bytes.NewReader(data.Bytes()), int64(data.Len()))
			if err != nil {
				t.Fatal(err)
			}
			if err := validateArchive(archive); err == nil {
				t.Fatal("unsafe archive was accepted")
			}
		})
	}
}

func testManifest(id, kind string) Manifest {
	return Manifest{Schema: ManifestSchema, ID: id, Name: id, Version: "1.0.0", Kind: kind,
		EngineAPI: EngineAPI, Redistributable: true, Entrypoints: Entrypoints{ClientComponents: []string{id + ".boot"}}}
}

func testBundle(manifest Manifest, files map[string]string) fstest.MapFS {
	data, _ := json.Marshal(manifest)
	result := fstest.MapFS{manifestPath: &fstest.MapFile{Data: data, Mode: 0o600}}
	for name, value := range files {
		result[name] = &fstest.MapFile{Data: []byte(value), Mode: 0o600}
	}
	return result
}
