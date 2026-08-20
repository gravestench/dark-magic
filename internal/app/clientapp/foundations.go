package clientapp

import (
	"errors"
	"image"
	"log/slog"
	"os"

	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/localization"
	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/worldobjects"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	"github.com/gravestench/dark-magic/internal/platform/desktop"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
)

// loadSettings establishes user policy before any backend or network client is
// created, ensuring all later capabilities derive paths and trust from one snapshot.
func (app *application) loadSettings() error {
	if err := app.loadShellSettings(); err != nil {
		return err
	}

	if err := app.loadGameSettings(); err != nil {
		return err
	}

	return app.loadNetworkTrust()
}

// loadShellSettings expands host syntax at the boundary and selects transient
// settings when persistence was not requested, rather than inventing a path.
func (app *application) loadShellSettings() error {
	path, err := darkpaths.ExpandHost(os.Getenv("DARK_MAGIC_SHELL_CONFIG"))
	if err != nil {
		return wrap("expand shell settings path", err)
	}

	app.shellSettings, err = shell.NewSettings(path)

	return wrap("load shell settings", err)
}

// loadGameSettings loads the durable preferences shared by presentation and
// connection setup; later services receive this store instead of rereading disk.
func (app *application) loadGameSettings() error {
	settings, err := preferences.New(os.Getenv("DARK_MAGIC_PREFERENCES"))
	app.gameSettings = settings

	return wrap("load game preferences", err)
}

// loadNetworkTrust derives persistent certificate pinning from the preferences
// directory so gateway trust survives restarts alongside the setting that names it.
func (app *application) loadNetworkTrust() error {
	trustDirectory, err := networktrust.Directory(app.gameSettings.Path())
	if err != nil {
		return wrap("resolve network trust directory", err)
	}

	app.networkTrust, err = networktrust.New(trustDirectory)

	return wrap("create network trust store", err)
}

// buildPresentationCore creates the native backend before backend-neutral stores
// are wired to it. Keeping this phase separate makes native capability ownership explicit.
func (app *application) buildPresentationCore() error {
	if err := app.loadPresentationConfiguration(); err != nil {
		return err
	}

	// Native details stop at desktop.Bundle; downstream services stay backend-neutral.
	backendOptions := app.desktopOptions()

	bundle, err := desktop.New(backendOptions)
	if err != nil {
		return err
	}

	app.installPresentationServices(bundle, backendOptions)

	return nil
}

// loadPresentationConfiguration pins one manifest-owned profile and validates its
// bootstrap assets before scenes can request presentation data.
func (app *application) loadPresentationConfiguration() error {
	profile, err := content.ResolvePresentationProfile(
		app.options.Content,
		app.options.PresentationProfileID,
	)
	if err != nil {
		return err
	}

	app.profile = profile

	app.presentation, err = content.LoadPresentationBootstrap(app.options.Content)

	return err
}

// desktopOptions contains the only translation from external process policy to
// desktop backend configuration, preventing scenes from interpreting CLI choices.
func (app *application) desktopOptions() desktop.Options {
	options := desktop.DefaultOptions()
	options.Content = app.options.Content
	options.PalettePath = app.options.OutputPalette
	options.LogicalWidth = app.profile.Width
	options.LogicalHeight = app.profile.Height
	options.ViewportFit = app.options.ViewportFit
	options.BorderlessFullscreen = app.options.BorderlessFullscreen
	options.NativeAudio = !app.options.DisableNativeAudio
	options.Logger = slog.Default().With("component", "renderer")

	return options
}

// installPresentationServices installs stable façades after backend creation.
// Consumers hold these façades so a backend choice does not leak through the application.
func (app *application) installPresentationServices(bundle *desktop.Bundle, options desktop.Options) {
	// Native adapters and the window geometry form the outer presentation edge.
	app.renderer, app.input = bundle.Renderer, bundle.Input
	app.renderWindow = image.Pt(options.WindowWidth, options.WindowHeight)

	// Input and localization are backend-neutral state read by scripts and scenes.
	app.inputState = &inputstate.Store{}
	app.locale = localization.New(app.options.Content, "English")

	// Lua owns presentation composition but not native renderer implementation.
	app.scripts = modruntime.New()
	app.composer = &render.Composer{}
	app.mixer = &audio.Mixer{}
	app.navigator = navigation.New()
	app.scenes = modruntime.NewScenes(app.scripts, app.navigator)
	app.scenes.SetInputStore(app.inputState)

	if app.options.Profile != nil {
		app.scenes.SetProfiler(app.options.Profile)
	}
}

// loadGameCatalogs pins all record-derived behavior to one content generation.
// Simulation, prediction, and world generation must never mix catalogs from different mounts.
func (app *application) loadGameCatalogs() error {
	pinned, generation, err := recordstore.Pin(app.options.Content)
	if err != nil && !errors.Is(err, recordstore.ErrNoAuthoritativeTables) {
		return wrap("pin authoritative game data", err)
	}

	if err == nil {
		app.records = pinned
	} else {
		// Modless and test content still needs a readable, initially empty store.
		app.records = recordstore.New(app.options.Content)
	}

	app.records.SetLogger(slog.Default().With("component", "records"))

	// Movement rates come from authoritative records, not recovered metadata.
	app.movementCatalog, err = d2movement.LoadCatalog(app.records)
	if err != nil {
		return wrap("load class movement rates", err)
	}

	// Quest and map-object relationships share one recovered-data snapshot.
	app.questCatalog = recovered.New(app.options.Content)

	recoveredData, err := app.loadRecoveredCatalogs()
	if err != nil {
		return err
	}

	slog.Info(
		"loaded recovered d2legacy records",
		"game_data_generation_id", generation.ID,
		"quests", len(recoveredData.Quests),
		"speech", len(recoveredData.Speech),
		"map_objects", len(recoveredData.MapObjects),
	)

	return nil
}

// loadRecoveredCatalogs takes one immutable recovered-data snapshot so world
// object resolution cannot change underneath already generated maps.
func (app *application) loadRecoveredCatalogs() (recovered.Snapshot, error) {
	recoveredData, err := app.questCatalog.Snapshot()
	if err != nil {
		return recovered.Snapshot{}, wrap("load recovered game data", err)
	}

	app.worldObjectResolver, err = worldobjects.New(recoveredData, app.records)

	return recoveredData, err
}

// gameDataGenerationID prefers the record store's content identity and falls back
// only before records exist; the value participates in multiplayer compatibility.
func (app *application) gameDataGenerationID() string {
	if app.records != nil && app.records.GenerationID() != "" {
		return app.records.GenerationID()
	}

	return simulation.GameDataGenerationIDForAssetSet(app.options.AssetSetID)
}
