package clientapp

import (
	"encoding/json"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/modcache"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// TestComponentDiscoveryKeepsEachResolvedPackageRoot verifies that discovery
// evaluates every package from its own mounted content subtree.
func TestComponentDiscoveryKeepsEachResolvedPackageRoot(t *testing.T) {
	firstManifest := discoveryManifest("first")
	secondManifest := discoveryManifest("second")
	first := packageLayer(t, firstManifest, `return {id="first.boot",api=1}`)
	second := packageLayer(t, secondManifest, `return {id="second.boot",api=1}`)

	contentFS, err := content.New(
		content.Layer{Name: "first", FS: first},
		content.Layer{Name: "second", FS: second},
	)
	if err != nil {
		t.Fatal(err)
	}

	runtime := modruntime.New()
	if err := runtime.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = runtime.Stop(t.Context())
	}()

	resolved := &modcache.ResolvedSet{
		Base: modcache.LockedPackage{Manifest: firstManifest},
		Extensions: modcache.Lock{
			Packages: []modcache.LockedPackage{{Manifest: secondManifest}},
		},
	}
	app := &application{options: Options{Content: contentFS, Mods: resolved}, scripts: runtime}
	definitions, err := app.discoverScriptDefinitions()
	if err != nil {
		t.Fatal(err)
	}

	ids := make([]string, len(definitions))
	for index, definition := range definitions {
		ids[index] = definition.ID
	}

	if !reflect.DeepEqual(ids, []string{"first.boot", "second.boot"}) {
		t.Fatalf("definitions = %v", ids)
	}
}

// discoveryManifest returns one minimal package manifest for discovery tests.
func discoveryManifest(id string) modcache.Manifest {
	return modcache.Manifest{
		Schema:          modcache.ManifestSchema,
		ID:              id,
		Name:            id,
		Version:         "1.0.0",
		Kind:            "game",
		EngineAPI:       modcache.EngineAPI,
		Redistributable: true,
		Entrypoints: modcache.Entrypoints{
			ClientComponents: []string{id + ".boot"},
		},
	}
}

// packageLayer wraps a manifest and boot script in the production package FS.
func packageLayer(t *testing.T, manifest modcache.Manifest, boot string) *modcache.PackageFS {
	t.Helper()

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}

	packageFS, err := modcache.NewPackageFS(manifest, fstest.MapFS{
		"mod.json": &fstest.MapFile{Data: data},
		"boot.lua": &fstest.MapFile{Data: []byte(boot)},
	})
	if err != nil {
		t.Fatal(err)
	}

	return packageFS
}
