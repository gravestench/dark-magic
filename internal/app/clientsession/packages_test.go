package clientsession

import (
	"context"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/modcache"
)

func TestPrepareSelfHostedExtensionsDownloadsAndVerifiesExactRecipe(t *testing.T) {
	baseManifest := packageManifest("d2legacy", "game")
	base, err := modcache.DescribeBuiltin(packageSource(baseManifest, map[string]string{"boot.lua": "return {}"}))
	if err != nil {
		t.Fatal(err)
	}
	hostStore, _ := modcache.New(filepath.Join(t.TempDir(), "host"))
	extensionManifest := packageManifest("example", "extension")
	extensionManifest.Dependencies = []modcache.Dependency{{ID: base.Manifest.ID, Version: base.Manifest.Version}}
	if _, err := hostStore.ReconcileBundled([]modcache.Bundle{{Source: packageSource(extensionManifest, map[string]string{"boot.lua": `return {id="example.boot"}`})}}); err != nil {
		t.Fatal(err)
	}
	resolved, err := hostStore.Resolve(modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{"example"}}, base)
	if err != nil {
		t.Fatal(err)
	}
	toRuntime := func(pkg modcache.LockedPackage) simulation.RuntimePackage {
		return simulation.RuntimePackage{ID: pkg.Manifest.ID, Version: pkg.Manifest.Version, Digest: pkg.Descriptor.Digest,
			Size: pkg.Descriptor.Size, Redistributable: pkg.Descriptor.Redistributable}
	}
	packages := simulation.RuntimePackageSet{Base: toRuntime(base), Extensions: []simulation.RuntimePackage{toRuntime(resolved.Extensions.Packages[0])}}
	identity := simulation.RuntimeIdentity{Recipe: simulation.RuntimeRecipe{
		Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", Packages: packages,
		AuthoritativeHash: "rules", ConfigurationHash: "config",
	}}
	allocation, err := gamesession.Allocate("game", identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}
	engine := gameecs.New()
	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })
	endpoint, err := gameserver.NewEndpoint(&gameserver.Host{Engine: engine, Session: session, Allocation: allocation}, rejectingAuthenticator{},
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.RawMessage(`{}`), nil })
	if err != nil {
		t.Fatal(err)
	}
	serverTLS, clientTLS, fingerprint := connectTLS(t)
	server, err := sessionquic.Listen("127.0.0.1:0", serverTLS, endpoint)
	if err != nil {
		t.Fatal(err)
	}
	provider, err := serverapp.NewPackageProvider(identity.Recipe, hostStore)
	if err != nil {
		t.Fatal(err)
	}
	server.SetPackageProvider(provider)
	t.Cleanup(func() { _ = server.Close() })
	serveContext, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	go func() { _ = server.Serve(serveContext) }()

	clientStore, _ := modcache.New(filepath.Join(t.TempDir(), "client"))
	ctx, stop := context.WithTimeout(t.Context(), 5*time.Second)
	defer stop()
	recipe, err := PrepareSelfHostedExtensions(ctx, SelfHostedAssignment{
		GameID: "game", Endpoint: realm.GameEndpoint{Address: server.Addr(), TLSFingerprint: fingerprint}, Runtime: identity,
	}, clientTLS, clientStore, packages.Base)
	if err != nil {
		t.Fatal(err)
	}
	if recipe.Packages.Extensions[0] != packages.Extensions[0] {
		t.Fatalf("downloaded recipe = %#v", recipe)
	}
	descriptor := resolved.Extensions.Packages[0].Descriptor
	if present, err := clientStore.Has(descriptor); err != nil || !present {
		t.Fatalf("downloaded extension present=%t error=%v", present, err)
	}
	if _, err := clientStore.ResolveExact([]modcache.Descriptor{descriptor}, base); err != nil {
		t.Fatalf("resolve downloaded exact extension: %v", err)
	}
}

type rejectingAuthenticator struct{}

func (rejectingAuthenticator) Authenticate(context.Context, string) (gameserver.Principal, error) {
	return gameserver.Principal{}, gameserver.ErrAuthentication
}

func packageManifest(id, kind string) modcache.Manifest {
	return modcache.Manifest{Schema: modcache.ManifestSchema, ID: id, Name: id, Version: "1.0.0", Kind: kind,
		EngineAPI: modcache.EngineAPI, Redistributable: true,
		Entrypoints: modcache.Entrypoints{ClientComponents: []string{id + ".boot"}}}
}

func packageSource(manifest modcache.Manifest, files map[string]string) fs.FS {
	encoded, _ := json.Marshal(manifest)
	result := fstest.MapFS{"mod.json": &fstest.MapFile{Data: encoded}}
	for name, value := range files {
		result[name] = &fstest.MapFile{Data: []byte(value)}
	}
	return result
}
