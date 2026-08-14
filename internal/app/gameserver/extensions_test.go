package gameserver

import (
	"context"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/gravestench/dark-magic/internal/content"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

func TestAuthorityStartsLockedExtensionEntrypoints(t *testing.T) {
	baseSource := content.D2Legacy()
	base, err := modcache.DescribeBuiltin(baseSource)
	if err != nil {
		t.Fatal(err)
	}
	manifest := modcache.Manifest{
		Schema: modcache.ManifestSchema, ID: "example", Name: "Example", Version: "1.0.0",
		Kind: "extension", EngineAPI: modcache.EngineAPI, Redistributable: true,
		Entrypoints:  modcache.Entrypoints{AuthorityComponents: []string{"example.authority"}},
		Dependencies: []modcache.Dependency{{ID: "d2legacy", Version: base.Manifest.Version}},
	}
	manifestJSON, _ := json.Marshal(manifest)
	extensionSource := fstest.MapFS{
		"mod.json":                 {Data: manifestJSON},
		"boot.lua":                 {Data: []byte(`return {id="example.boot",api=1}`)},
		"components/authority.lua": {Data: []byte(`local state=require("engine.authority_state/v1"); return {id="example.authority",api=1,start=function() state.register("example.started","example/v1",{value=true}) end}`)},
	}
	store, err := modcache.New(filepath.Join(t.TempDir(), "cache"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ReconcileBundled([]modcache.Bundle{{Source: extensionSource}}); err != nil {
		t.Fatal(err)
	}
	resolved, err := store.Resolve(modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{"example"}}, base)
	if err != nil {
		t.Fatal(err)
	}
	mounted, err := store.Mount(resolved.Extensions, base)
	if err != nil {
		t.Fatal(err)
	}
	defer mounted.Close()
	extensionFS, err := modcache.NewPackageFS(manifest, mounted.Packages[0].FS)
	if err != nil {
		t.Fatal(err)
	}
	baseFS, err := modcache.NewPackageFS(base.Manifest, baseSource)
	if err != nil {
		t.Fatal(err)
	}
	contentFS, err := content.New(content.Layer{Name: "mod:example", FS: extensionFS}, content.Layer{Name: "builtin:d2legacy", FS: baseFS})
	if err != nil {
		t.Fatal(err)
	}
	d2source, err := fs.Sub(contentFS, "mods/d2legacy")
	if err != nil {
		t.Fatal(err)
	}
	packages := simulation.RuntimePackageSet{Base: runtimePackage(base)}
	packages.Extensions = append(packages.Extensions, runtimePackage(resolved.Extensions.Packages[0]))
	host, err := Start(t.Context(), d2source, fixtureRecords{}, Config{
		Mode: ModeListen, SessionID: "extensions", Prediction: gamesession.PredictionLimited,
		Packages: packages, Content: contentFS, Mods: &resolved,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close(context.Background())
	if state, found := host.Authority.State.Read("example.started"); !found || string(state.Data) != `{"value":true}` {
		t.Fatalf("authority extension state = %#v found=%t", state, found)
	}
	if len(host.Allocation.Identity.Recipe.Packages.Extensions) != 1 || host.Allocation.Identity.Recipe.Packages.Extensions[0].ID != "example" {
		t.Fatalf("authority recipe = %#v", host.Allocation.Identity.Recipe)
	}
}

func runtimePackage(pkg modcache.LockedPackage) simulation.RuntimePackage {
	return simulation.RuntimePackage{ID: pkg.Manifest.ID, Version: pkg.Manifest.Version, Digest: pkg.Descriptor.Digest,
		Size: pkg.Descriptor.Size, Redistributable: pkg.Descriptor.Redistributable}
}
