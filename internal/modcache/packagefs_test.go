package modcache_test

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/modcache"
)

// TestPackageNamespaceIsPrivateAndDeclaredContentRootsOverlay verifies private
// files remain namespaced while only declared roots enter shared lookup.
func TestPackageNamespaceIsPrivateAndDeclaredContentRootsOverlay(t *testing.T) {
	base := packageFS(t, manifest("base", []string{"assets"}), "base")
	extension := packageFS(t, manifest("extension", []string{"assets"}), "extension")

	contentFS, err := content.New(
		content.Layer{Name: "mod:extension", FS: extension},
		content.Layer{Name: "mod:base", FS: base},
	)
	if err != nil {
		t.Fatal(err)
	}

	assertFile(t, contentFS, "mods/base/boot.lua", "base boot")
	assertFile(t, contentFS, "mods/extension/boot.lua", "extension boot")
	assertFile(t, contentFS, "assets/shared.txt", "extension")

	if _, err := fs.ReadFile(contentFS, "boot.lua"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("private root boot = %v", err)
	}
}

// TestManifestRejectsPrivateRootsAsSharedContent prevents executable and package
// namespaces from being exported into the shared virtual root.
func TestManifestRejectsPrivateRootsAsSharedContent(t *testing.T) {
	for _, root := range []string{"components", "lua", "mods"} {
		candidate := manifest("example", []string{root})
		if err := modcache.ValidateManifest(candidate); err == nil {
			t.Fatalf("reserved root %q was accepted", root)
		}
	}
}

// manifest creates the smallest valid extension manifest needed by filesystem
// projection tests, keeping namespace expectations visible.
func manifest(id string, roots []string) modcache.Manifest {
	return modcache.Manifest{
		Schema: modcache.ManifestSchema, ID: id, Name: id, Version: "1.0.0", Kind: "game",
		EngineAPI: modcache.EngineAPI, Redistributable: true, ContentRoots: roots,
		Entrypoints: modcache.Entrypoints{ClientComponents: []string{id + ".boot"}},
	}
}

// packageFS builds a package with distinct private and exported values so tests
// can detect accidental namespace leakage.
func packageFS(t *testing.T, packageManifest modcache.Manifest, value string) fs.FS {
	t.Helper()

	source := fstest.MapFS{
		"boot.lua":          &fstest.MapFile{Data: []byte(value + " boot")},
		"assets/shared.txt": &fstest.MapFile{Data: []byte(value)},
	}

	result, err := modcache.NewPackageFS(packageManifest, source)
	if err != nil {
		t.Fatal(err)
	}

	return result
}

// assertFile reads through the mounted filesystem boundary and reports both path
// and contents when an overlay resolves incorrectly.
func assertFile(t *testing.T, source fs.FS, name, want string) {
	t.Helper()

	data, err := fs.ReadFile(source, name)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != want {
		t.Fatalf("%s = %q, want %q", name, data, want)
	}
}
