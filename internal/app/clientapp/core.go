package clientapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"

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
	return wrap("load game preferences", err)
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
	app.records = recordstore.New(app.options.Content)
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
		"quests", len(recoveredData.Quests), "speech", len(recoveredData.Speech),
		"map_objects", len(recoveredData.MapObjects))
	return nil
}

func (app *application) buildOfflineSession() error {
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
	trustDirectory, err := networktrust.Directory(app.gameSettings.Path())
	if err != nil {
		return wrap("resolve network trust directory", err)
	}
	app.networkTrust, err = networktrust.New(trustDirectory)
	if err != nil {
		return wrap("create network trust store", err)
	}
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
	initialData := map[string]any{
		"d2legacy.development_items": map[string]any{
			"enabled":                 app.options.FixtureCharacters > 0,
			"create_empty_containers": app.options.FixtureCharacters == 0,
		},
		"d2legacy.interactions":      app.interactionBootstrapData(),
		"d2legacy.world_transitions": app.transitionBootstrapData(),
	}
	identity, err := d2legacymod.Identity(app.options.Content, initialData)
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
	if err := d2legacymod.ConfigureRuntime(app.scripts, app.options.Content, app.records, app.entitySimulation, app.offlineSession,
		app.authoritativeState, app.authoritativeRandom, initialData); err != nil {
		return wrap("configure canonical d2legacy runtime", err)
	}
	if err := app.registerOfflineCommands(); err != nil {
		return err
	}
	return app.buildLoadingCoordinator()
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
	payload, err := json.Marshal(app.populationBootstrapData())
	if err != nil {
		return wrap("encode entry population geometry", err)
	}
	return wrap("queue d2legacy entry population", app.offlineSession.Submit(simulation.Command{
		Tick: 1, Player: "d2legacy.population", Authority: simulation.AuthoritySystem,
		Sequence: 1, Kind: "system.population.bootstrap", Payload: payload,
	}))
}

func (app *application) populationBootstrapData() map[string]any {
	zone, worldMap := app.gameWorldZones[2], app.gameWorlds[2]
	if zone == nil || worldMap == nil {
		return nil
	}
	request := zone.Request()
	populated := map[uint32]bool{}
	for _, stamp := range zone.Stamps() {
		populated[stamp.ID] = stamp.Populate
	}
	nearby := developmentScenes[app.options.StartScene].nearbyHostiles
	player := app.gameWorldSpawns[2]
	rooms := make([]any, 0, len(zone.Rooms()))
	for _, room := range zone.Rooms() {
		points := make([]any, 0, 8)
		anchors := [][2]float64{}
		if nearby > 0 {
			anchors = [][2]float64{{player[0] + 10, player[1]}, {player[0] + 7, player[1] + 7}, {player[0], player[1] + 10}, {player[0] - 7, player[1] + 7}}
		} else {
			centerX, centerY := float64((room.X+room.Width/2)*5)+2, float64((room.Y+room.Height/2)*5)+2
			anchors = [][2]float64{{centerX, centerY}, {centerX + 1, centerY}, {centerX, centerY + 1}, {centerX - 1, centerY}}
		}
		for _, anchor := range anchors {
			if x, y, ok := worldMap.OpenPointNearSubtile(anchor[0], anchor[1]); ok {
				points = append(points, map[string]any{"x": x, "y": y})
			}
		}
		rooms = append(rooms, map[string]any{"id": float64(room.ID), "populate": populated[room.StampID], "points": points})
	}
	return map[string]any{"seed": float64(request.Seed), "act": float64(request.Act), "level_id": float64(request.LevelID), "difficulty": float64(request.Difficulty), "rooms": rooms}
}

func (app *application) interactionBootstrapData() map[string]any {
	initial := ""
	if app.options.StartScene == "vendor" {
		initial = "act1-akara"
	}
	targets := []any{map[string]any{"id": "act1-akara", "npc": "Akara", "vendor": "Akara", "categories": "armo,misc,weap", "services": "", "x": float64(4096), "y": float64(4096), "radius": float64(160)}}
	for _, worldMap := range app.gameWorlds {
		objects := make(map[string]gameworld.Object, len(worldMap.Objects))
		for index, object := range worldMap.Objects {
			objects[fmt.Sprintf("ds1-object:%d:%d:%d", object.Type, object.ID, index)] = object
		}
		for _, selected := range worldMap.Selectables() {
			object := objects[selected.ID]
			name := strings.TrimSpace(object.Description)
			if name == "" {
				name = strings.TrimSpace(object.Class)
			}
			if name == "" {
				continue
			}
			targets = append(targets, map[string]any{"id": selected.ID, "npc": name, "vendor": "", "categories": "", "services": "", "x": selected.X, "y": selected.Y, "radius": float64(4)})
		}
	}
	return map[string]any{"owner": "local-player", "initial_target": initial, "targets": targets}
}

func (app *application) buildLoadingCoordinator() error {
	app.loading = loadcore.New(map[string]loadcore.Task{
		"selected_character": func(context.Context) error {
			if _, ok := app.saves.Selected(); !ok {
				return errors.New("no character is selected")
			}
			return nil
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
