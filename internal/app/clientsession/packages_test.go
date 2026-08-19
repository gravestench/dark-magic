package clientsession

import (
	"context"
	"crypto/tls"
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

// TestPrepareSelfHostedExtensionsDownloadsAndVerifiesExactRecipe proves authenticated preparation
// installs the exact redistributable recipe without consuming profile admission.
func TestPrepareSelfHostedExtensionsDownloadsAndVerifiesExactRecipe(t *testing.T) {
	fixture := newExtensionRecipeFixture(t, "example", map[string]string{
		"boot.lua": `return {id="example.boot"}`,
	})
	endpoint, clientTLS := startPackageFixtureServer(t, fixture)

	clientStore, _ := modcache.New(filepath.Join(t.TempDir(), "client"))

	ctx, stop := context.WithTimeout(t.Context(), 5*time.Second)
	defer stop()

	recipe, err := PrepareSelfHostedExtensions(ctx, SelfHostedAssignment{
		GameID: "game", Endpoint: endpoint, Runtime: fixture.identity,
	}, clientTLS, clientStore, fixture.packages.Base)
	if err != nil {
		t.Fatal(err)
	}

	if recipe.Packages.Extensions[0] != fixture.packages.Extensions[0] {
		t.Fatalf("downloaded recipe = %#v", recipe)
	}

	if present, err := clientStore.Has(fixture.descriptor); err != nil || !present {
		t.Fatalf("downloaded extension present=%t error=%v", present, err)
	}

	if _, err := clientStore.ResolveExact([]modcache.Descriptor{fixture.descriptor}, fixture.base); err != nil {
		t.Fatalf("resolve downloaded exact extension: %v", err)
	}
}

// TestAcquireExtensionsInterruptedDownloadLeavesNoPackageAndCleanRetrySucceeds protects quarantine:
// partial bytes never become a cache hit and a later complete transfer remains possible.
func TestAcquireExtensionsInterruptedDownloadLeavesNoPackageAndCleanRetrySucceeds(t *testing.T) {
	fixture := newExtensionRecipeFixture(t, "interrupted", map[string]string{
		"boot.lua":    `return {id="interrupted.boot"}`,
		"payload.bin": "download payload",
	})

	archive, err := os.ReadFile(filepath.Join(
		fixture.hostRoot,
		"blobs",
		"sha256",
		fixture.descriptor.Digest[len("sha256:"):]+".zip",
	))
	if err != nil {
		t.Fatal(err)
	}

	clientRoot := filepath.Join(t.TempDir(), "client")

	clientStore, err := modcache.New(clientRoot)
	if err != nil {
		t.Fatal(err)
	}

	interrupted := &fixtureExtensionTransport{recipe: fixture.identity.Recipe, archive: archive, failAfter: 16}
	if _, err := AcquireExtensions(t.Context(), interrupted, clientStore, fixture.packages.Base); err == nil {
		t.Fatal("interrupted package transfer succeeded")
	}

	assertInterruptedPackageQuarantined(t, clientStore, clientRoot, fixture.descriptor)

	retried := &fixtureExtensionTransport{recipe: fixture.identity.Recipe, archive: archive}
	if _, err := AcquireExtensions(t.Context(), retried, clientStore, fixture.packages.Base); err != nil {
		t.Fatalf("retry package transfer: %v", err)
	}

	if present, err := clientStore.Has(fixture.descriptor); err != nil || !present {
		t.Fatalf("retried package present=%t error=%v", present, err)
	}
}

// extensionRecipeFixture keeps the canonical cache artifacts and their network representation together.
// This prevents tests from silently describing different bytes to the host store and the remote recipe.
type extensionRecipeFixture struct {
	base       modcache.LockedPackage
	descriptor modcache.Descriptor
	hostRoot   string
	hostStore  *modcache.Store
	identity   simulation.RuntimeIdentity
	packages   simulation.RuntimePackageSet
}

// newExtensionRecipeFixture installs one redistributable extension and derives its authenticated recipe.
func newExtensionRecipeFixture(
	t *testing.T,
	extensionID string,
	files map[string]string,
) extensionRecipeFixture {
	t.Helper()

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

	extensionManifest := packageManifest(extensionID, "extension")

	extensionManifest.Dependencies = []modcache.Dependency{{ID: base.Manifest.ID, Version: base.Manifest.Version}}
	if _, err := hostStore.ReconcileBundled([]modcache.Bundle{{
		Source: packageSource(extensionManifest, files),
	}}); err != nil {
		t.Fatal(err)
	}

	profile := modcache.Profile{Schema: modcache.ProfileSchema, Enabled: []string{extensionID}}

	resolved, err := hostStore.Resolve(profile, base)
	if err != nil {
		t.Fatal(err)
	}

	descriptor := resolved.Extensions.Packages[0].Descriptor
	baseRuntime := runtimePackage(base)
	extensionRuntime := runtimePackage(resolved.Extensions.Packages[0])
	packages := simulation.RuntimePackageSet{
		Base: baseRuntime, Extensions: []simulation.RuntimePackage{extensionRuntime},
	}
	recipe := simulation.RuntimeRecipe{Schema: simulation.RuntimeRecipeSchema, EngineAPI: "v1",
		NetworkProtocol: "test/v1", AssetSetID: simulation.EmptyAssetSetID,
		GameDataGenerationID: simulation.GameDataGenerationIDForAssetSet(simulation.EmptyAssetSetID),
		Packages:             packages,
		AuthoritativeHash:    "rules", ConfigurationHash: "config"}

	return extensionRecipeFixture{
		base:       base,
		descriptor: descriptor,
		hostRoot:   hostRoot,
		hostStore:  hostStore,
		identity:   simulation.RuntimeIdentity{Recipe: recipe},
		packages:   packages,
	}
}

// runtimePackage converts a resolved cache record into the exact package identity sent over the wire.
func runtimePackage(pkg modcache.LockedPackage) simulation.RuntimePackage {
	return simulation.RuntimePackage{
		ID:              pkg.Manifest.ID,
		Version:         pkg.Manifest.Version,
		Digest:          pkg.Descriptor.Digest,
		Size:            pkg.Descriptor.Size,
		Redistributable: pkg.Descriptor.Redistributable,
	}
}

// startPackageFixtureServer exposes only recipe and package services.
// Its rejecting authenticator ensures preparation cannot accidentally consume gameplay admission.
func startPackageFixtureServer(
	t *testing.T,
	fixture extensionRecipeFixture,
) (realm.GameEndpoint, *tls.Config) {
	t.Helper()

	allocation, err := gamesession.Allocate("game", fixture.identity, gamesession.PredictionLimited)
	if err != nil {
		t.Fatal(err)
	}

	engine := gameecs.New()

	session, err := gamesession.New(engine, gamesession.Config{})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = session.Close(); _ = engine.Close() })

	host := &gameserver.Host{Engine: engine, Session: session, Allocation: allocation}

	endpoint, err := gameserver.NewEndpoint(host, rejectingAuthenticator{},
		func(string, simulation.Checkpoint) (json.RawMessage, error) { return json.RawMessage(`{}`), nil })
	if err != nil {
		t.Fatal(err)
	}

	serverTLS, clientTLS, fingerprint := connectTLS(t)

	server, err := sessionquic.Listen("127.0.0.1:0", serverTLS, endpoint)
	if err != nil {
		t.Fatal(err)
	}

	provider, err := serverapp.NewPackageProvider(fixture.identity.Recipe, fixture.hostStore)
	if err != nil {
		t.Fatal(err)
	}

	server.SetPackageProvider(provider)
	t.Cleanup(func() { _ = server.Close() })
	serveContext, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	go func() { _ = server.Serve(serveContext) }()

	return realm.GameEndpoint{Address: server.Addr(), TLSFingerprint: fingerprint}, clientTLS
}

// assertInterruptedPackageQuarantined proves incomplete bytes are neither addressable nor left on disk.
func assertInterruptedPackageQuarantined(
	t *testing.T,
	clientStore *modcache.Store,
	clientRoot string,
	descriptor modcache.Descriptor,
) {
	t.Helper()

	if present, err := clientStore.Has(descriptor); err != nil || present {
		t.Fatalf("partial package present=%t error=%v", present, err)
	}

	quarantine, err := os.ReadDir(filepath.Join(clientRoot, "quarantine"))
	if err != nil || len(quarantine) != 0 {
		t.Fatalf("quarantine after interruption=%#v error=%v", quarantine, err)
	}
}

// fixtureExtensionTransport records chunk progress and can interrupt one deterministic byte offset.
type fixtureExtensionTransport struct {
	recipe    simulation.RuntimeRecipe
	archive   []byte
	failAfter int64
}

// Recipe returns the authenticated fixture recipe advertised by the transport.
func (transport *fixtureExtensionTransport) Recipe(context.Context) (simulation.RuntimeRecipe, error) {
	return transport.recipe, nil
}

// PackageChunk enforces requested ranges while injecting an optional one-shot interruption.
func (transport *fixtureExtensionTransport) PackageChunk(
	_ context.Context,
	request sessionquic.PackageRequest,
) (sessionquic.PackageChunk, error) {
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

	return sessionquic.PackageChunk{
		Total: int64(len(transport.archive)),
		Data:  append([]byte(nil), transport.archive[request.Offset:end]...),
	}, nil
}

// rejectingAuthenticator proves package preparation never attempts gameplay authentication.
type rejectingAuthenticator struct{}

// Authenticate fails every call so accidental ticket consumption is immediately visible.
func (rejectingAuthenticator) Authenticate(context.Context, string) (gameserver.Principal, error) {
	return gameserver.Principal{}, gameserver.ErrAuthentication
}

// packageManifest creates a deterministic cache manifest for package-transfer scenarios.
func packageManifest(id, kind string) modcache.Manifest {
	return modcache.Manifest{Schema: modcache.ManifestSchema, ID: id, Name: id, Version: "1.0.0", Kind: kind,
		EngineAPI: modcache.EngineAPI, Redistributable: true,
		Entrypoints: modcache.Entrypoints{ClientComponents: []string{id + ".boot"}}}
}

// packageSource serializes a manifest and its files into the bundle layout consumed by modcache.
func packageSource(manifest modcache.Manifest, files map[string]string) fs.FS {
	encoded, _ := json.Marshal(manifest)

	result := fstest.MapFS{"mod.json": &fstest.MapFile{Data: encoded}}
	for name, value := range files {
		result[name] = &fstest.MapFile{Data: []byte(value)}
	}

	return result
}
