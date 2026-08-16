package clientsession

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
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
		Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1", NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID, Packages: packages,
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

func TestAcquireExtensionsInterruptedDownloadLeavesNoPackageAndCleanRetrySucceeds(t *testing.T) {
	baseManifest := packageManifest("d2legacy", "game")
	base, err := modcache.DescribeBuiltin(packageSource(baseManifest, map[string]string{"boot.lua": "return {}"}))
	if err != nil {
		t.Fatal(err)
	}
	hostRoot := filepath.Join(t.TempDir(), "host")
	hostStore, err := modcache.New(hostRoot)
	if err != nil {
		t.Fatal(err)
	}
	extensionManifest := packageManifest("interrupted", "extension")
	extensionManifest.Dependencies = []modcache.Dependency{{ID: base.Manifest.ID, Version: base.Manifest.Version}}
	if _, err := hostStore.ReconcileBundled([]modcache.Bundle{{Source: packageSource(extensionManifest,
		map[string]string{"boot.lua": `return {id="interrupted.boot"}`, "payload.bin": "download payload"})}}); err != nil {
		t.Fatal(err)
	}
	resolved, err := hostStore.Resolve(modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{"interrupted"}}, base)
	if err != nil {
		t.Fatal(err)
	}
	descriptor := resolved.Extensions.Packages[0].Descriptor
	archive, err := os.ReadFile(filepath.Join(hostRoot, "blobs", "sha256", descriptor.Digest[len("sha256:"):]+".zip"))
	if err != nil {
		t.Fatal(err)
	}
	baseRuntime := simulation.RuntimePackage{ID: base.Manifest.ID, Version: base.Manifest.Version,
		Digest: base.Descriptor.Digest, Size: base.Descriptor.Size, Redistributable: base.Descriptor.Redistributable}
	extensionRuntime := simulation.RuntimePackage{ID: descriptor.ID, Version: descriptor.Version,
		Digest: descriptor.Digest, Size: descriptor.Size, Redistributable: descriptor.Redistributable}
	recipe := simulation.RuntimeRecipe{Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1",
		NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID,
		Packages:          simulation.RuntimePackageSet{Base: baseRuntime, Extensions: []simulation.RuntimePackage{extensionRuntime}},
		AuthoritativeHash: "rules", ConfigurationHash: "config"}

	clientRoot := filepath.Join(t.TempDir(), "client")
	clientStore, err := modcache.New(clientRoot)
	if err != nil {
		t.Fatal(err)
	}
	interrupted := &fixtureExtensionTransport{recipe: recipe, archive: archive, failAfter: 16}
	if _, err := AcquireExtensions(t.Context(), interrupted, clientStore, baseRuntime); err == nil {
		t.Fatal("interrupted package transfer succeeded")
	}
	if present, err := clientStore.Has(descriptor); err != nil || present {
		t.Fatalf("partial package present=%t error=%v", present, err)
	}
	quarantine, err := os.ReadDir(filepath.Join(clientRoot, "quarantine"))
	if err != nil || len(quarantine) != 0 {
		t.Fatalf("quarantine after interruption=%#v error=%v", quarantine, err)
	}

	retried := &fixtureExtensionTransport{recipe: recipe, archive: archive}
	if _, err := AcquireExtensions(t.Context(), retried, clientStore, baseRuntime); err != nil {
		t.Fatalf("retry package transfer: %v", err)
	}
	if present, err := clientStore.Has(descriptor); err != nil || !present {
		t.Fatalf("retried package present=%t error=%v", present, err)
	}
}

type fixtureExtensionTransport struct {
	recipe    simulation.RuntimeRecipe
	archive   []byte
	failAfter int64
}

func (transport *fixtureExtensionTransport) Recipe(context.Context) (simulation.RuntimeRecipe, error) {
	return transport.recipe, nil
}

func (transport *fixtureExtensionTransport) PackageChunk(_ context.Context, request sessionquic.PackageRequest) (sessionquic.PackageChunk, error) {
	if transport.failAfter > 0 && request.Offset >= transport.failAfter {
		return sessionquic.PackageChunk{}, errors.New("fixture transport interrupted")
	}
	if request.Offset < 0 || request.Offset >= int64(len(transport.archive)) {
		return sessionquic.PackageChunk{}, errors.New("fixture transport offset out of range")
	}
	limit := request.Limit
	if limit > 16 {
		limit = 16
	}
	end := request.Offset + int64(limit)
	if end > int64(len(transport.archive)) {
		end = int64(len(transport.archive))
	}
	return sessionquic.PackageChunk{Total: int64(len(transport.archive)), Data: append([]byte(nil), transport.archive[request.Offset:end]...)}, nil
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
