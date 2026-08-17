package clientapp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/gravestench/dark-magic/internal/app/networktrust"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	gameecs "github.com/gravestench/dark-magic/internal/game/ecs"
	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/inputstate"
	loadcore "github.com/gravestench/dark-magic/internal/loading"
	"github.com/gravestench/dark-magic/internal/localization"
	d2legacymod "github.com/gravestench/dark-magic/internal/mod/d2legacy"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	d2movement "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/movement"
	gameplayer "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/worldobjects"
	"github.com/gravestench/dark-magic/internal/mod/d2legacy/data/recovered"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	raylibinput "github.com/gravestench/dark-magic/internal/platform/raylib/input"
	raylibrenderer "github.com/gravestench/dark-magic/internal/platform/raylib/renderer"
	"github.com/gravestench/dark-magic/internal/preferences"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
)

func (app *application) loadSettings() error {
	path, err := darkpaths.ExpandHost(os.Getenv("DARK_MAGIC_SHELL_CONFIG"))
	if err != nil {
		return wrap("expand shell settings path", err)
	}
	app.shellSettings, err = shell.NewSettings(path)
	if err != nil {
		return wrap("load shell settings", err)
	}
	app.gameSettings, err = preferences.New(os.Getenv("DARK_MAGIC_PREFERENCES"))
	if err != nil {
		return wrap("load game preferences", err)
	}
	trustDirectory, err := networktrust.Directory(app.gameSettings.Path())
	if err != nil {
		return wrap("resolve network trust directory", err)
	}
	app.networkTrust, err = networktrust.New(trustDirectory)
	return wrap("create network trust store", err)
}

func (app *application) buildPresentationCore() error {
	profile, err := content.ResolvePresentationProfile(app.options.Content, app.options.PresentationProfileID)
	if err != nil {
		return err
	}
	app.profile = profile
	app.presentation, err = content.LoadPresentationBootstrap(app.options.Content)
	if err != nil {
		return err
	}

	// The renderer is the window. Input reads that window. Everything else talks
	// to backend-neutral stores so game code never needs to know about Raylib.
	app.renderer = &raylibrenderer.Service{}
	app.renderer.SetLogger(slog.Default().With("component", "renderer"))
	app.rendererConfig = raylibrenderer.DefaultConfig()
	app.rendererConfig.Resolution.Width = profile.Width
	app.rendererConfig.Resolution.Height = profile.Height
	app.rendererConfig.Resolution.Fit = app.options.ViewportFit
	app.rendererConfig.Window.Borderless = app.options.BorderlessFullscreen
	app.renderer.Configure(app.rendererConfig)
	if err := app.renderer.ConfigurePaletteQuantization(app.options.Content, app.options.OutputPalette); err != nil {
		return err
	}

	app.input = raylibinput.New(app.renderer)
	app.input.SetLogger(slog.Default().With("component", "input"))
	app.inputState = &inputstate.Store{}
	app.locale = localization.New(app.options.Content, "English")
	app.scripts = modruntime.New()
	app.composer = &render.Composer{}
	app.mixer = &audio.Mixer{}
	app.navigator = navigation.New()
	app.scenes = modruntime.NewScenes(app.scripts, app.navigator)
	app.scenes.SetInputStore(app.inputState)
	if app.options.Profile != nil {
		app.scenes.SetProfiler(app.options.Profile)
	}
	return nil
}

func (app *application) loadGameCatalogs() error {
	pinned, generation, err := recordstore.Pin(app.options.Content)
	if err != nil && !errors.Is(err, recordstore.ErrNoAuthoritativeTables) {
		return wrap("pin authoritative game data", err)
	}
	if err == nil {
		app.records = pinned
	} else {
		app.records = recordstore.New(app.options.Content)
	}
	app.records.SetLogger(slog.Default().With("component", "records"))
	app.questCatalog = recovered.New(app.options.Content)

	recoveredData, err := app.questCatalog.Snapshot()
	if err != nil {
		return wrap("load recovered game data", err)
	}
	app.worldObjectResolver, err = worldobjects.New(recoveredData, app.records)
	if err != nil {
		return err
	}
	slog.Info("loaded recovered d2legacy records",
		"game_data_generation_id", generation.ID,
		"quests", len(recoveredData.Quests), "speech", len(recoveredData.Speech),
		"map_objects", len(recoveredData.MapObjects))
	return nil
}

func (app *application) gameDataGenerationID() string {
	if app.records != nil && app.records.GenerationID() != "" {
		return app.records.GenerationID()
	}
	return simulation.GameDataGenerationIDForAssetSet(app.options.AssetSetID)
}

func (app *application) buildOfflineSession() error {
	app.configuredMods = cloneRuntimePackages(app.options.Packages)
	fixtures := DevelopmentCharacters(app.options.FixtureCharacters)
	profilePath, err := darkpaths.ExpandHost(app.options.PlayerProfilePath)
	if err != nil {
		return wrap("expand player profile path", err)
	}
	app.saves, app.playerProfilePath, err = loadPlayerProfile(profilePath, fixtures)
	if err != nil {
		return err
	}
	app.network = newNetworkController(app)
	app.realm = newRealmController(app)
	if len(fixtures) > 0 && fixtureNeedsSelection(app.options.StartScene) {
		if err := app.saves.Select(fixtures[0].ID); err != nil {
			return wrap("select development fixture", err)
		}
	}

	app.entitySimulation = gameecs.New()
	if err := app.buildEntryWorld(); err != nil {
		return err
	}
	session, err := gamesession.New(app.entitySimulation, gamesession.Config{})
	if err != nil {
		_ = app.entitySimulation.Close()
		return wrap("create offline game session", err)
	}
	app.offlineSession = session
	app.authoritativeState = simulation.NewStateStore()
	app.authoritativeRandom, err = d2legacymod.NewRandomStreams(0)
	if err != nil {
		return wrap("register d2legacy random streams", err)
	}
	initialData := app.sessionInitialData()
	d2legacySource, err := app.modSource("d2legacy")
	if err != nil {
		return wrap("resolve d2legacy package", err)
	}
	identity, err := d2legacymod.IdentityForPackagesAndData(d2legacySource, app.options.Packages,
		app.options.AssetSetID, app.gameDataGenerationID(), initialData)
	if err != nil {
		return wrap("identify d2legacy mod", err)
	}
	if err := session.RegisterAuthoritativeRuntime(identity, app.authoritativeState, app.authoritativeRandom); err != nil {
		return wrap("register d2legacy authoritative runtime", err)
	}
	app.commandIntents = &gamesession.IntentController{}
	app.commandIntentSource, err = gamesession.NewIntentSource(app.commandIntents, "local-player")
	if err != nil {
		return wrap("create local command intent source", err)
	}
	app.ecsCapability = modruntime.NewECSCapability(app.scripts, app.entitySimulation)
	if app.options.Mods != nil {
		packages := app.options.Mods.Packages()
		packageIDs := make([]string, len(packages))
		for index, pkg := range packages {
			packageIDs[index] = pkg.Manifest.ID
		}
		app.packageRegistry = modruntime.NewPackageRegistry(packageIDs)
		if err := app.scripts.RegisterInstaller(modruntime.PackageRequireRegistry(app.options.Content, app.packageRegistry)); err != nil {
			return wrap("register package Lua namespaces", err)
		}
		app.packageDigests = make(map[string]string, len(packages))
		for _, pkg := range packages {
			app.packageDigests[pkg.Manifest.ID] = pkg.Descriptor.Digest
		}
	}
	if err := d2legacymod.ConfigureRuntime(app.scripts, d2legacySource, app.records, app.entitySimulation, app.offlineSession,
		app.authoritativeState, app.authoritativeRandom, initialData, app.ecsCapability); err != nil {
		return wrap("configure canonical d2legacy runtime", err)
	}
	if err := app.registerOfflineCommands(); err != nil {
		return err
	}
	return app.buildLoadingCoordinator()
}

func (app *application) sessionInitialData() map[string]any {
	return map[string]any{
		"engine.game_data_generation_id": app.gameDataGenerationID(),
		"d2legacy.game_rules": map[string]any{
			"target": "lod-1.14d", "expansion": true, "difficulty": 0,
			"hardcore": false, "ladder": false, "maximum_players": 8,
		},
		"d2legacy.development_items": map[string]any{
			"enabled":                 app.options.FixtureCharacters > 0,
			"create_empty_containers": app.options.FixtureCharacters == 0,
		},
		"d2legacy.interactions":      app.interactionBootstrapData(),
		"d2legacy.world_transitions": app.transitionBootstrapData(),
	}
}

func (app *application) registerOfflineCommands() error {
	bloodMoor := app.gameWorlds[2]
	if bloodMoor == nil {
		return errors.New("register hostile simulation: Blood Moor world is unavailable")
	}
	if err := gameworld.RegisterVelocityMovement(app.entitySimulation, bloodMoor, gameworld.VelocityComponents{
		Position: "d2legacy.world.position", Velocity: "d2legacy.world.velocity", Collider: "d2legacy.world.collider",
	}); err != nil {
		return wrap("register generic velocity movement", err)
	}
	movement := &d2movement.MovementController{}
	movementSource, err := d2movement.NewMovementSource(app.entitySimulation, app.inputState, "local-player", "game_world", movement)
	if err != nil {
		return wrap("create offline movement source", err)
	}
	app.movementSource = movementSource
	entryLevel := app.activeWorldLevel
	worldMap := app.gameWorlds[entryLevel]
	if worldMap == nil {
		return errors.New("load offline entry map: session world is unavailable")
	}
	movementSource.SetNavigation(worldMap)
	spawn, found := app.gameWorldSpawns[entryLevel]
	if !found {
		return errors.New("create offline player entry source: world has no trusted spawn subtile")
	}
	spawnX, spawnY := spawn[0], spawn[1]
	request := app.gameWorldZones[entryLevel].Request()
	destination, err := gameplayer.NewDestination(spawnX, spawnY, float64(worldMap.WidthSubtiles), float64(worldMap.HeightSubtiles), int64(request.Act), int64(request.LevelID))
	if err != nil {
		return wrap("create Act I town admission destination", err)
	}
	entry, err := gameplayer.NewEntrySourceForDestination(app.entitySimulation, app.saves, "local-player", destination)
	if err != nil {
		return wrap("create offline player entry source", err)
	}
	sequencer := simulation.NewLocalSequencer()
	app.commandSource = func(tick uint64) []simulation.Command {
		commands := entry.Commands(tick)
		commands = append(commands, movementSource.Commands(tick)...)
		commands = append(commands, app.commandIntentSource.Commands(tick)...)
		return sequencer.Assign(commands)
	}
	app.playerControl = movement
	return nil
}

func (app *application) queueEntryPopulation() error {
	command, err := app.populationBootstrapCommand()
	if err != nil {
		return err
	}
	return wrap("queue d2legacy entry population", app.offlineSession.Submit(command))
}

func (app *application) populationBootstrapCommand() (simulation.Command, error) {
	nearby := developmentScenes[app.options.StartScene].nearbyHostiles
	return app.preparedEntryWorld().PopulationCommand(nearby)
}

func (app *application) populationBootstrapData() map[string]any {
	nearby := developmentScenes[app.options.StartScene].nearbyHostiles
	return app.preparedEntryWorld().PopulationData(nearby)
}

func (app *application) interactionBootstrapData() map[string]any {
	initial := ""
	if app.options.StartScene == "vendor" {
		initial = "act1-akara"
	}
	return entryworld.InteractionData(app.gameWorlds, "local-player", initial)
}

func (app *application) preparedEntryWorld() *entryworld.Prepared {
	return &entryworld.Prepared{Worlds: app.gameWorlds, Zones: app.gameWorldZones, Spawns: app.gameWorldSpawns, Seam: app.transitionSeam}
}

func (app *application) buildLoadingCoordinator() error {
	app.loading = loadcore.New(map[string]loadcore.Task{
		"selected_character": func(context.Context) error {
			if _, ok := app.saves.Selected(); ok {
				return nil
			}
			if app.network != nil && app.network.hasSelectedCharacter() {
				return nil
			}
			return errors.New("no character is selected")
		},
		"loading_assets": func(context.Context) error {
			for _, name := range app.presentation.LoadingAssets {
				if _, err := fs.Stat(app.options.Content, name); err != nil {
					return fmt.Errorf("load dependency %q: %w", name, err)
				}
			}
			return nil
		},
		"world": func(context.Context) error { return nil },
	})
	return nil
}

func fixtureNeedsSelection(scene string) bool {
	switch scene {
	case "game_world", "game_loading", "combat_lab", "inventory", "character", "skills", "vendor":
		return true
	default:
		return false
	}
}
