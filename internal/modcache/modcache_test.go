package modcache

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
)

func TestExtensionIsCachedByDigestAndResolvedAgainstBuiltinBase(t *testing.T) {
	store, err := New(filepath.Join(t.TempDir(), "cache"))
	if err != nil {
		t.Fatal(err)
	}
	base := testBuiltin(t)
	bundle := testBundle(testManifest("example", "extension"), map[string]string{"boot.lua": "return {}"})
	defaults, err := store.ReconcileBundled([]Bundle{{Source: bundle, DefaultEnabled: true}})
	if err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(t.TempDir(), "mods.json")
	profile, created, err := LoadOrCreateProfile(profilePath, defaults)
	if err != nil {
		t.Fatal(err)
	}
	if !created || !reflect.DeepEqual(profile.Enabled, []string{"example"}) {
		t.Fatalf("new profile = %#v created=%t", profile, created)
	}
	resolved, err := store.Resolve(profile, base)
	if err != nil {
		t.Fatal(err)
	}
	lock := resolved.Extensions
	if resolved.Base.Manifest.ID != "d2legacy" || len(lock.Packages) != 1 || lock.Packages[0].Manifest.ID != "example" || !validDigest(lock.Digest) {
		t.Fatalf("resolved set = %#v", resolved)
	}
	descriptor := lock.Packages[0].Descriptor
	if _, err := os.Stat(store.blobPath(descriptor.Digest)); err != nil {
		t.Fatalf("content-addressed blob: %v", err)
	}
	mounted, err := store.Mount(lock, base)
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
	resolved, err = store.Resolve(profile, base)
	if err != nil || len(resolved.Extensions.Packages) != 0 {
		t.Fatalf("empty extension set = %#v, %v", resolved, err)
	}
}

func TestResolverOrdersDependenciesBeforeDependentsAndLookupInReverse(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	base := testBuiltin(t)
	foundation := testManifest("foundation", "extension")
	foundation.Dependencies = []Dependency{{ID: base.Manifest.ID, Version: base.Manifest.Version}}
	extension := testManifest("extension", "extension")
	extension.Dependencies = []Dependency{{ID: "foundation", Version: foundation.Version}}
	if _, err := store.ReconcileBundled([]Bundle{
		{Source: testBundle(foundation, map[string]string{"shared.txt": "foundation"})},
		{Source: testBundle(extension, map[string]string{"shared.txt": "extension"})},
	}); err != nil {
		t.Fatal(err)
	}
	resolved, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"extension"}}, base)
	if err != nil {
		t.Fatal(err)
	}
	lock := resolved.Extensions
	ids := []string{lock.Packages[0].Manifest.ID, lock.Packages[1].Manifest.ID}
	if !reflect.DeepEqual(ids, []string{"foundation", "extension"}) {
		t.Fatalf("activation order = %v", ids)
	}
	mounted, err := store.Mount(lock, base)
	if err != nil {
		t.Fatal(err)
	}
	defer mounted.Close()
	lookup := mounted.LookupOrder()
	if lookup[0].ID != "extension" || lookup[1].ID != "foundation" {
		t.Fatalf("lookup order = %#v", lookup)
	}
}

func TestResolverRejectsMissingCyclesAndTamperedBlobs(t *testing.T) {
	base := testBuiltin(t)
	t.Run("missing", func(t *testing.T) {
		store, _ := New(t.TempDir())
		if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"missing"}}, base); err == nil {
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
		if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"first"}}, base); err == nil || !strings.Contains(err.Error(), "cycle") {
			t.Fatalf("cycle error = %v", err)
		}
	})
	t.Run("tamper", func(t *testing.T) {
		store, _ := New(t.TempDir())
		if _, err := store.ReconcileBundled([]Bundle{{Source: testBundle(testManifest("enabled", "extension"), nil)}}); err != nil {
			t.Fatal(err)
		}
		catalog, err := store.readIndex()
		if err != nil {
			t.Fatal(err)
		}
		descriptor := catalog.Packages["enabled"]
		if err := os.WriteFile(store.blobPath(descriptor.Digest), []byte("tampered"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"enabled"}}, base); err == nil {
			t.Fatal("tampered package was accepted")
		}
	})
}

func TestDisabledBrokenPackageCannotPreventStartup(t *testing.T) {
	store, _ := New(t.TempDir())
	if _, err := store.ReconcileBundled([]Bundle{
		{Source: testBundle(testManifest("enabled", "extension"), nil)},
		{Source: testBundle(testManifest("disabled", "extension"), nil)},
	}); err != nil {
		t.Fatal(err)
	}
	catalog, _ := store.readIndex()
	disabled := catalog.Packages["disabled"]
	if err := os.WriteFile(store.blobPath(disabled.Digest), []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolved, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"enabled"}}, testBuiltin(t))
	if err != nil || len(resolved.Extensions.Packages) != 1 || resolved.Extensions.Packages[0].Manifest.ID != "enabled" {
		t.Fatalf("enabled set = %#v, %v", resolved, err)
	}
}

func TestResolvedSetRejectsCachedGameAndExplicitBuiltinProfileEntry(t *testing.T) {
	store, _ := New(t.TempDir())
	if _, err := store.ReconcileBundled([]Bundle{{Source: testBundle(testManifest("other_game", "game"), nil)}}); err != nil {
		t.Fatal(err)
	}
	base := testBuiltin(t)
	if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"other_game"}}, base); err == nil || !strings.Contains(err.Error(), "not an extension") {
		t.Fatalf("cached game error = %v", err)
	}
	if _, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"d2legacy"}}, base); err == nil || !strings.Contains(err.Error(), "always enabled") {
		t.Fatalf("built-in profile error = %v", err)
	}
}

func TestResolvedSetRejectsOverlappingPackageNamespaces(t *testing.T) {
	store, _ := New(t.TempDir())
	parent := testManifest("example", "extension")
	child := testManifest("example.feature", "extension")
	if _, err := store.ReconcileBundled([]Bundle{
		{Source: testBundle(parent, nil)}, {Source: testBundle(child, nil)},
	}); err != nil {
		t.Fatal(err)
	}
	profile := Profile{Schema: ProfileSchema, Enabled: []string{parent.ID, child.ID}}
	if _, err := store.Resolve(profile, testBuiltin(t)); err == nil || !strings.Contains(err.Error(), "overlap") {
		t.Fatalf("overlapping namespace error = %v", err)
	}
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
	if _, err := store.ReconcileBundled([]Bundle{{Source: testBundle(testManifest("example", "extension"), nil)}}); err != nil {
		t.Fatal(err)
	}
	base := testBuiltin(t)
	resolved, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"example"}}, base)
	if err != nil {
		t.Fatal(err)
	}
	lock := resolved.Extensions
	lock.Packages[0].Manifest.Entrypoints.ClientComponents = []string{"attacker.boot"}
	if _, err := store.Mount(lock, base); err == nil || !strings.Contains(err.Error(), "digest") {
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

func TestInstallVerifiedPromotesOnlyExactExtensionBytes(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	manifest := testManifest("downloaded", "extension")
	archive := archiveBytes(t, testBundle(manifest, map[string]string{"boot.lua": "return {}"}))
	digest := sha256.Sum256(archive)
	descriptor := Descriptor{ID: manifest.ID, Version: manifest.Version,
		Digest: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(archive)), Redistributable: true}
	installed, err := store.InstallVerified(context.Background(), bytes.NewReader(archive), descriptor)
	if err != nil {
		t.Fatal(err)
	}
	if installed.ID != manifest.ID {
		t.Fatalf("installed manifest = %#v", installed)
	}
	base := testBuiltin(t)
	resolved, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{manifest.ID}}, base)
	if err != nil || len(resolved.Extensions.Packages) != 1 {
		t.Fatalf("resolved installed extension = %#v, %v", resolved, err)
	}
	tampered := append([]byte(nil), archive...)
	tampered[len(tampered)-1] ^= 0xff
	if _, err := store.InstallVerified(context.Background(), bytes.NewReader(tampered), descriptor); err == nil {
		t.Fatal("tampered download was installed")
	}
}

func TestInstallVerifiedRejectsManifestThatDiffersFromDescriptor(t *testing.T) {
	store, _ := New(t.TempDir())
	manifest := testManifest("actual", "extension")
	archive := archiveBytes(t, testBundle(manifest, nil))
	digest := sha256.Sum256(archive)
	descriptor := Descriptor{ID: "advertised", Version: manifest.Version,
		Digest: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(archive)), Redistributable: true}
	if _, err := store.InstallVerified(t.Context(), bytes.NewReader(archive), descriptor); err == nil {
		t.Fatal("archive with a different manifest identity was accepted")
	}
}

func TestExactSessionVersionsCoexistWithoutChangingProfileSelection(t *testing.T) {
	store, _ := New(t.TempDir())
	first := testManifest("shared", "extension")
	first.Version = "1.0.0"
	second := first
	second.Version = "2.0.0"
	install := func(manifest Manifest, marker string) Descriptor {
		archive := archiveBytes(t, testBundle(manifest, map[string]string{"marker.txt": marker}))
		digest := sha256.Sum256(archive)
		descriptor := Descriptor{ID: manifest.ID, Version: manifest.Version,
			Digest: "sha256:" + hex.EncodeToString(digest[:]), Size: int64(len(archive)), Redistributable: true}
		if _, err := store.InstallVerified(t.Context(), bytes.NewReader(archive), descriptor); err != nil {
			t.Fatal(err)
		}
		return descriptor
	}
	firstDescriptor := install(first, "first")
	secondDescriptor := install(second, "second")
	base := testBuiltin(t)
	profile, err := store.Resolve(Profile{Schema: ProfileSchema, Enabled: []string{"shared"}}, base)
	if err != nil || profile.Extensions.Packages[0].Descriptor != firstDescriptor {
		t.Fatalf("profile selection changed = %#v, %v", profile, err)
	}
	exact, err := store.ResolveExact([]Descriptor{secondDescriptor}, base)
	if err != nil || exact.Extensions.Packages[0].Descriptor != secondDescriptor {
		t.Fatalf("exact session version = %#v, %v", exact, err)
	}
	if present, err := store.Has(firstDescriptor); err != nil || !present {
		t.Fatalf("first version present=%t error=%v", present, err)
	}
}

func TestConcurrentCacheUpdatesDoNotLosePackages(t *testing.T) {
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	const count = 12
	var wait sync.WaitGroup
	errorsFound := make(chan error, count)
	for index := 0; index < count; index++ {
		index := index
		wait.Add(1)
		go func() {
			defer wait.Done()
			id := fmt.Sprintf("extension_%02d", index)
			_, err := store.ReconcileBundled([]Bundle{{Source: testBundle(testManifest(id, "extension"), nil)}})
			errorsFound <- err
		}()
	}
	wait.Wait()
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatal(err)
		}
	}
	catalog, err := store.readIndex()
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Packages) != count {
		t.Fatalf("concurrent index contains %d packages, want %d", len(catalog.Packages), count)
	}
}

func TestDescribeBuiltinCanonicalizesTextLineEndings(t *testing.T) {
	manifest := testManifest("d2legacy", "game")
	lf := testBundle(manifest, map[string]string{
		"boot.lua":  "local value = 1\nreturn value\n",
		"README.md": "first line\nsecond line\n",
	})
	crlf := testBundle(manifest, map[string]string{
		"boot.lua":  "local value = 1\r\nreturn value\r\n",
		"README.md": "first line\r\nsecond line\r\n",
	})

	left, err := DescribeBuiltin(lf)
	if err != nil {
		t.Fatal(err)
	}
	right, err := DescribeBuiltin(crlf)
	if err != nil {
		t.Fatal(err)
	}
	if left.Descriptor != right.Descriptor {
		t.Fatalf("equivalent text checkouts have different descriptors:\nLF:   %#v\nCRLF: %#v", left.Descriptor, right.Descriptor)
	}
}

func TestDescribeBuiltinKeepsBinaryBytesExact(t *testing.T) {
	manifest := testManifest("d2legacy", "game")
	lf := testBundle(manifest, map[string]string{"asset.bin": "first\nsecond"})
	crlf := testBundle(manifest, map[string]string{"asset.bin": "first\r\nsecond"})

	left, err := DescribeBuiltin(lf)
	if err != nil {
		t.Fatal(err)
	}
	right, err := DescribeBuiltin(crlf)
	if err != nil {
		t.Fatal(err)
	}
	if left.Descriptor.Digest == right.Descriptor.Digest {
		t.Fatal("different binary bytes produced the same built-in digest")
	}
}

func TestCachedExtensionArchiveKeepsTextBytesExact(t *testing.T) {
	manifest := testManifest("example", "extension")
	lf := archiveBytes(t, testBundle(manifest, map[string]string{"boot.lua": "first\nsecond"}))
	crlf := archiveBytes(t, testBundle(manifest, map[string]string{"boot.lua": "first\r\nsecond"}))
	if bytes.Equal(lf, crlf) {
		t.Fatal("extension archives canonicalized distinct text bytes")
	}
}

func archiveBytes(t *testing.T, source fs.FS) []byte {
	t.Helper()
	var archive bytes.Buffer
	if err := writeArchive(&archive, source); err != nil {
		t.Fatal(err)
	}
	return archive.Bytes()
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

func testBuiltin(t *testing.T) LockedPackage {
	t.Helper()
	base, err := DescribeBuiltin(testBundle(testManifest("d2legacy", "game"), map[string]string{"boot.lua": "return {}"}))
	if err != nil {
		t.Fatal(err)
	}
	return base
}
