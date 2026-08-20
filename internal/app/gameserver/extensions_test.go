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

// TestAuthorityStartsLockedExtensionEntrypoints proves resolved extension code and identity enter
// one authority runtime.
func TestAuthorityStartsLockedExtensionEntrypoints(t *testing.T) {
	fixture := newLockedExtensionFixture(t)

	host, err := Start(t.Context(), fixture.source, fixtureRecords{}, Config{
		Mode:       ModeListen,
		SessionID:  "extensions",
		Prediction: gamesession.PredictionLimited,
		Packages:   fixture.packages,
		Content:    fixture.content,
		Mods:       &fixture.resolved,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = host.Close(context.Background()) })

	if state, found := host.Authority.State.Read("example.started"); !found || string(state.Data) != `{"value":true}` {
		t.Fatalf("authority extension state = %#v found=%t", state, found)
	}

	if len(host.Allocation.Identity.Recipe.Packages.Extensions) != 1 ||
		host.Allocation.Identity.Recipe.Packages.Extensions[0].ID != "example" {
		t.Fatalf("authority recipe = %#v", host.Allocation.Identity.Recipe)
	}
}

// lockedExtensionFixture retains mounted content and the matching runtime recipe used to start the host.
type lockedExtensionFixture struct {
	source   fs.FS
	content  fs.FS
	packages simulation.RuntimePackageSet
	resolved modcache.ResolvedSet
}

// newLockedExtensionFixture resolves and mounts one authority extension without hiding its package identity.
func newLockedExtensionFixture(t *testing.T) lockedExtensionFixture {
	t.Helper()

	baseSource := content.D2Legacy()

	base, err := modcache.DescribeBuiltin(baseSource)
	if err != nil {
		t.Fatal(err)
	}

	manifest, extensionSource := extensionFixtureSource(base.Manifest.Version)
	resolved, mounted := resolveExtensionFixture(t, base, extensionSource)
	contentFS := mountExtensionFixtureContent(t, base, baseSource, manifest, mounted)

	d2source, err := fs.Sub(contentFS, "mods/d2legacy")
	if err != nil {
		t.Fatal(err)
	}

	packages := simulation.RuntimePackageSet{Base: runtimePackage(base)}
	packages.Extensions = append(packages.Extensions, runtimePackage(resolved.Extensions.Packages[0]))

	return lockedExtensionFixture{
		source:   d2source,
		content:  contentFS,
		packages: packages,
		resolved: resolved,
	}
}

// extensionFixtureSource keeps the declared dependency and the executable extension files together.
func extensionFixtureSource(baseVersion string) (modcache.Manifest, fstest.MapFS) {
	manifest := modcache.Manifest{
		Schema: modcache.ManifestSchema, ID: "example", Name: "Example", Version: "1.0.0",
		Kind: "extension", EngineAPI: modcache.EngineAPI, Redistributable: true,
		Entrypoints:  modcache.Entrypoints{AuthorityComponents: []string{"example.authority"}},
		Dependencies: []modcache.Dependency{{ID: "d2legacy", Version: baseVersion}},
	}
	manifestJSON, _ := json.Marshal(manifest)
	source := fstest.MapFS{
		"mod.json": {Data: manifestJSON},
		"boot.lua": {Data: []byte(`return {id="example.boot",api=1}`)},
		"components/authority.lua": {Data: []byte(
			`local state=require("engine.authority_state/v1"); ` +
				`return {id="example.authority",api=1,start=function() ` +
				`state.register("example.started","example/v1",{value=true}) end}`,
		)},
	}

	return manifest, source
}

// resolveExtensionFixture exercises the real cache path so the test cannot accidentally bypass lock metadata.
func resolveExtensionFixture(
	t *testing.T,
	base modcache.LockedPackage,
	extensionSource fs.FS,
) (modcache.ResolvedSet, *modcache.MountedSet) {
	t.Helper()

	store, err := modcache.New(filepath.Join(t.TempDir(), "cache"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := store.ReconcileBundled([]modcache.Bundle{{Source: extensionSource}}); err != nil {
		t.Fatal(err)
	}

	profile := modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{"example"}}

	resolved, err := store.Resolve(profile, base)
	if err != nil {
		t.Fatal(err)
	}

	mounted, err := store.Mount(resolved.Extensions, base)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = mounted.Close() })

	return resolved, mounted
}

// mountExtensionFixtureContent assembles the same layered namespace that production hosts receive.
func mountExtensionFixtureContent(
	t *testing.T,
	base modcache.LockedPackage,
	baseSource fs.FS,
	manifest modcache.Manifest,
	mounted *modcache.MountedSet,
) fs.FS {
	t.Helper()

	extensionFS, err := modcache.NewPackageFS(manifest, mounted.Packages[0].FS)
	if err != nil {
		t.Fatal(err)
	}

	baseFS, err := modcache.NewPackageFS(base.Manifest, baseSource)
	if err != nil {
		t.Fatal(err)
	}

	contentFS, err := content.New(
		content.Layer{Name: "mod:example", FS: extensionFS},
		content.Layer{Name: "builtin:d2legacy", FS: baseFS},
	)
	if err != nil {
		t.Fatal(err)
	}

	return contentFS
}

// runtimePackage converts cache identity into the exact recipe entry authenticated by clients.
func runtimePackage(pkg modcache.LockedPackage) simulation.RuntimePackage {
	return simulation.RuntimePackage{
		ID:              pkg.Manifest.ID,
		Version:         pkg.Manifest.Version,
		Digest:          pkg.Descriptor.Digest,
		Size:            pkg.Descriptor.Size,
		Redistributable: pkg.Descriptor.Redistributable,
	}
}
