package clientapp

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gravestench/akara"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/distribution"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/localization"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// realD2LegacyFixtureConfig describes the production scene setup required by one acceptance scenario.
type realD2LegacyFixtureConfig struct {
	startScene         string
	fixtureCharacters  int
	fixtureWorldLevel  int
	applySceneDefaults bool
}

// requireRealStore returns a required dynamic component store with a scenario-specific failure.
func requireRealStore(
	t *testing.T,
	fixture *realD2LegacyFixture,
	name string,
	missing string,
) *akara.DynamicStore {
	t.Helper()

	store, found := akara.GetDynamicStore(fixture.app.entitySimulation.World(), name)
	if !found {
		t.Fatal(missing)
	}

	return store
}

// realD2LegacyFixture owns a production-data client application for one acceptance test.
type realD2LegacyFixture struct {
	app     *application
	options Options
}

// newRealD2LegacyFixture starts the real d2legacy authority and registers complete cleanup.
func newRealD2LegacyFixture(t *testing.T, config realD2LegacyFixtureConfig) *realD2LegacyFixture {
	t.Helper()
	requireRealD2LegacyAssets(t)

	options := realD2LegacyOptions(t)
	options.StartScene = config.startScene
	options.FixtureCharacters = config.fixtureCharacters

	options.FixtureWorldLevel = config.fixtureWorldLevel
	if config.applySceneDefaults {
		options = applyDevelopmentSceneDefaults(options)
	}

	app := &application{
		options:    options,
		inputState: &inputstate.Store{},
		locale:     localization.New(options.Content, "English"),
		scripts:    modruntime.New(),
	}
	if err := app.loadGameCatalogs(); err != nil {
		t.Fatal(err)
	}

	if err := app.buildOfflineSession(); err != nil {
		t.Fatal(err)
	}

	startTestD2LegacyAuthority(t, app)
	t.Cleanup(func() {
		app.loading.Close()
		_ = app.offlineSession.Close()
		_ = app.entitySimulation.Close()
		_ = content.Close(options.Content)
	})

	return &realD2LegacyFixture{app: app, options: options}
}

// requireRealD2LegacyAssets skips production-data acceptance tests without an MPQ source.
func requireRealD2LegacyAssets(t *testing.T) {
	t.Helper()

	if os.Getenv("MPQ_DIRECTORY") == "" {
		t.Skip("MPQ_DIRECTORY is not configured")
	}
}

// startTestD2LegacyAuthority loads and starts the real managed d2legacy component.
func startTestD2LegacyAuthority(t *testing.T, app *application) {
	t.Helper()

	if err := app.scripts.Start(t.Context()); err != nil {
		t.Fatal(err)
	}

	source, err := app.modSource("d2legacy")
	if err != nil {
		t.Fatal(err)
	}

	definition, err := modruntime.LoadDefinition(
		t.Context(),
		app.scripts,
		source,
		"components/d2legacy.lua",
	)
	if err != nil {
		t.Fatal(err)
	}

	component, err := definition.Managed().New(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if err := component.Start(t.Context()); err != nil {
		t.Fatal(err)
	}

	if err := app.queueEntryPopulation(); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = component.Stop(context.Background())
		_ = app.scripts.Stop(context.Background())
	})
}

// realD2LegacyOptions resolves production mod layers and pinned asset identity.
func realD2LegacyOptions(t *testing.T) Options {
	t.Helper()

	mods, err := distribution.PrepareMods("none")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = mods.Close()
	})

	assets, err := content.FromEnvironment(mods.Layers...)
	if err != nil {
		t.Fatal(err)
	}

	assetSetID, err := content.AssetSetIdentityFromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	return Options{
		Content:    assets,
		Mods:       &mods.Resolved,
		Packages:   mods.Packages,
		AssetSetID: assetSetID,
	}
}

// advanceOffline advances the canonical local session in host-sized fixed steps.
func (fixture *realD2LegacyFixture) advanceOffline(t *testing.T, frames int) {
	t.Helper()

	for range frames {
		if _, err := fixture.app.offlineSession.AdvanceWithSource(
			time.Second/25,
			fixture.app.commandSource,
		); err != nil {
			t.Fatal(err)
		}
	}
}

// advanceGame advances the full application path in host-sized fixed steps.
func (fixture *realD2LegacyFixture) advanceGame(t *testing.T, frames int) {
	t.Helper()

	for range frames {
		if err := fixture.app.advanceGame(time.Second / 25); err != nil {
			t.Fatal(err)
		}
	}
}
