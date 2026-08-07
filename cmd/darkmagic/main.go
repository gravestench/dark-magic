package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/gravestench/dark-magic/internal/app/filewatch"
	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/hotreload"
	"github.com/gravestench/dark-magic/internal/app/runtimeapi"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/dev/capture"
	"github.com/gravestench/dark-magic/internal/dev/profiling"
	"github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/loading"
	"github.com/gravestench/dark-magic/internal/localization"
	"github.com/gravestench/dark-magic/internal/logging"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	"github.com/gravestench/dark-magic/internal/persistence"
	"github.com/gravestench/dark-magic/internal/platform/raylib/input"
	raylibRenderer "github.com/gravestench/dark-magic/internal/platform/raylib/renderer"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	"github.com/gravestench/dark-magic/internal/presentation/scene"
	"github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/shell"
	"github.com/gravestench/dark-magic/internal/shell/luashell"
	raylibShell "github.com/gravestench/dark-magic/internal/shell/raylib"
	"github.com/gravestench/dark-magic/internal/video"
)

func main() {
	// Cocoa and GLFW must be initialized and pumped from the process's original
	// main thread. Keep the entire renderer lifecycle on that thread.
	runtime.LockOSThread()
	logLevelFlag := flag.String("log-level", environmentDefault("DARK_MAGIC_LOG_LEVEL", "info"), "log verbosity: debug, info, warn, or error")
	profileDirectory := flag.String("profile-dir", os.Getenv("DARK_MAGIC_PROFILE_DIR"), "capture CPU and heap profiles plus PDF reports in this directory")
	profileScenes := flag.String("profile-scenes", os.Getenv("DARK_MAGIC_PROFILE_SCENES"), "comma-separated scene IDs (or all) for per-scene CPU and heap reports")
	captureDirectoryFlag := flag.String("capture-dir", os.Getenv("DARK_MAGIC_CAPTURE_DIR"), "write local scene screenshots and report.json to this directory")
	captureScenes := flag.String("capture-scenes", os.Getenv("DARK_MAGIC_CAPTURE_SCENES"), "comma-separated scene IDs to capture (defaults to loading,title)")
	captureSettle := flag.Int("capture-settle-frames", 10, "stable frames to wait before capturing a scene")
	startScene := flag.String("start-scene", os.Getenv("DARK_MAGIC_START_SCENE"), "development-only scene ID to enter after boot")
	fixtureCharacters := flag.Int("fixture-characters", 0, "development-only number of in-memory characters to create")
	outputPalette := flag.String("output-palette", os.Getenv("DARK_MAGIC_OUTPUT_PALETTE"), "quantize the final display through this mounted pal.dat asset")
	viewportFit := flag.String("viewport-fit", environmentDefault("DARK_MAGIC_VIEWPORT_FIT", "contain"), "game viewport fit: contain or stretch")
	flag.Parse()
	logLevel, err := parseLogLevel(*logLevelFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	shellLogs := shell.NewLogBuffer(1000)
	logHandler := logging.NewHandlerWithObserver(&slog.HandlerOptions{Level: logLevel}, func(record logging.Record) {
		shellLogs.Append(shell.LogEntry{
			At: record.At, Level: record.Level.String(), Message: record.Message, Attributes: record.Attributes,
		})
	})
	slog.SetDefault(slog.New(logHandler))
	var profile *profiling.Session
	if *profileDirectory != "" {
		var err error
		profile, err = profiling.Start(*profileDirectory, true)
		if err != nil {
			slog.Error("starting profiler", "error", err)
			return
		}
		profile.ConfigureScenes(*profileScenes)
		defer func() {
			if err := profile.Stop(); err != nil {
				slog.Error("finishing profiler", "error", err)
			}
		}()
	}
	contentFS, err := content.FromEnvironment()
	if err != nil {
		slog.Error("constructing content filesystem", "error", err)
		return
	}
	if err := content.ValidateClientAssets(contentFS); err != nil {
		slog.Error("validating client content", "error", err)
		return
	}
	captureDirectory, err := darkpaths.ExpandHost(*captureDirectoryFlag)
	if err != nil {
		slog.Error("expanding capture directory", "error", err)
		return
	}
	if captureDirectory != "" && *captureScenes == "" {
		*captureScenes = "loading,title"
	}
	if err := run(contentFS, profile, captureDirectory, *captureScenes, *captureSettle, *startScene, *fixtureCharacters, *outputPalette, *viewportFit, shellLogs); err != nil {
		slog.Error("running Dark Magic", "error", err)
	}
}

func environmentDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func parseLogLevel(value string) (slog.Level, error) {
	return logging.ParseLevel(value)
}

func run(contentFS *content.FS, profile *profiling.Session, captureDirectory, captureScenes string, captureSettle int, startScene string, fixtureCharacters int, outputPalette, viewportFit string, shellLogs *shell.LogBuffer) error {
	runContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	shellSettingsPath, err := darkpaths.ExpandHost(os.Getenv("DARK_MAGIC_SHELL_CONFIG"))
	if err != nil {
		return fmt.Errorf("expand shell settings path: %w", err)
	}
	shellSettings, err := shell.NewSettings(shellSettingsPath)
	if err != nil {
		return err
	}
	sceneErrors := make(chan error, 1)
	reportSceneError := func(err error) {
		select {
		case sceneErrors <- err:
			stopSignals()
		default:
		}
	}
	renderer := &raylibRenderer.Service{}
	renderer.SetLogger(slog.Default().With("component", "renderer"))
	rendererConfig := raylibRenderer.DefaultConfig()
	rendererConfig.Resolution.Fit = viewportFit
	renderer.Configure(rendererConfig)
	if err := renderer.ConfigurePaletteQuantization(contentFS, outputPalette); err != nil {
		return err
	}
	inputService := input.New(renderer)
	inputService.SetLogger(slog.Default().With("component", "input"))
	locale := localization.New(contentFS, "English")
	scripts := modruntime.New()
	composer := &render.Composer{}
	mixer := &audio.Mixer{}
	navigator := navigation.New()
	scenes := modruntime.NewScenes(scripts, navigator)
	if profile != nil {
		scenes.SetProfiler(profile)
	}
	inputState := &inputstate.Store{}
	records := recordstore.New(contentFS)
	records.SetLogger(slog.Default().With("component", "records"))
	gameData := gamedata.New(records)
	presentation, err := content.LoadPresentationBootstrap(contentFS)
	if err != nil {
		return err
	}
	typedRecords, err := gameData.Snapshot()
	if err != nil {
		return fmt.Errorf("load typed game data: %w", err)
	}
	slog.Info("loaded typed game-data catalog", "issues", len(typedRecords.Issues))
	fixtureEntries := developmentCharacters(fixtureCharacters)
	saves := persistence.New(fixtureEntries...)
	if len(fixtureEntries) > 0 && (startScene == "game_world" || startScene == "inventory" || startScene == "character") {
		if err := saves.Select(fixtureEntries[0].ID); err != nil {
			return fmt.Errorf("select development fixture: %w", err)
		}
	}
	simulation := modruntime.NewSimulation(scene.New(1, 4096, 4096))
	loading := loading.New(map[string]loading.Task{
		"selected_character": func(context.Context) error {
			if _, ok := saves.Selected(); !ok {
				return errors.New("no character is selected")
			}
			return nil
		},
		"loading_assets": func(_ context.Context) error {
			for _, name := range presentation.LoadingAssets {
				if _, err := fs.Stat(contentFS, name); err != nil {
					return fmt.Errorf("load dependency %q: %w", name, err)
				}
			}
			return nil
		},
		"world": func(context.Context) error {
			_ = simulation.Snapshot()
			return nil
		},
	})
	defer loading.Close()
	components := host.NewManager()
	if err := scripts.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.AppModule(buildVersion(), stopSignals)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.ShellModule(shellSettings)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.VFSModule(contentFS)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.DataModule(contentFS)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.WorldModule(contentFS)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.InputModule(inputState)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.AudioModule(scripts, mixer, contentFS, gameData)); err != nil {
		return err
	}
	videoBackend := video.NewEmbeddedBackend(composer, mixer, image.Pt(rendererConfig.Window.Width, rendererConfig.Window.Height))
	if !videoBackend.Available() {
		videoBackend = video.FFplay{}
	}
	if resizable, ok := videoBackend.(interface{ Resize(image.Point) error }); ok {
		renderer.SubscribeViewport(func(width, height int) {
			if err := resizable.Resize(image.Pt(width, height)); err != nil {
				slog.Error("resizing cinematic viewport", "error", err)
			}
		})
	}
	if err := scripts.RegisterModule(modruntime.VideoModule(scripts, videoBackend, contentFS)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.RecordsModule(gameData)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.GameDataModule(gameData)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.LocaleModule(locale)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.SaveModule(saves)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.SimulationModule(simulation)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.LoadingModule(loading)); err != nil {
		return err
	}
	renderCapability := modruntime.NewRenderCapability(scripts, composer, contentFS)
	if profile != nil {
		profile.SetDiagnostics(func() any { return renderCapability.Diagnostics() })
	}
	if err := scripts.RegisterModule(renderCapability.Module()); err != nil {
		return err
	}
	if err := scripts.RegisterModule(scenes.Module()); err != nil {
		return err
	}

	appHost := host.New()
	staticDefinitions := []host.Definition{
		{ID: "engine.renderer", Component: renderer},
		{ID: "engine.input", DependsOn: []string{"engine.renderer"}, Component: inputService},
		{ID: "engine.lua", DependsOn: []string{"engine.renderer", "engine.input"}, Component: scripts},
	}
	if address := os.Getenv("DARK_MAGIC_DEBUG_ADDR"); address != "" {
		staticDefinitions = append(staticDefinitions, host.Definition{ID: "engine.runtime-api", Component: runtimeapi.New(address, components)})
	}
	for _, definition := range staticDefinitions {
		if err := appHost.Register(definition); err != nil {
			return err
		}
	}
	if err := appHost.Start(context.Background()); err != nil {
		return err
	}
	stopped := false
	defer func() {
		if !stopped {
			stopHost(appHost)
		}
	}()
	shellPolicy := shell.Policy{
		Name:         "local-developer",
		Capabilities: scripts.ModuleNames(),
		Mutable:      true,
	}
	shellEvaluator, err := luashell.NewForPolicy(scripts, shellPolicy)
	if err != nil {
		return err
	}
	shellSession, err := shell.NewSession("client-local", "client", shellPolicy, shellEvaluator)
	if err != nil {
		return err
	}
	shellSession.AttachLogs(shellLogs)
	shellSession.AttachSettings(shellSettings)
	defer shellSession.Close()
	console := raylibShell.New(shellSession, shellSettings)
	if err := console.LoadFont(); err != nil {
		return err
	}
	defer console.Close()

	lastFrame := time.Now()
	stopSceneFrames := renderer.SubscribeFrame(func() {
		frameContext := scenes.FrameContext(context.Background())
		pprof.SetGoroutineLabels(frameContext)
		inputFrame := inputService.Snapshot()
		if console.Handle(runContext, inputFrame) {
			inputFrame = inputstate.Frame{
				Actions: make(map[string]inputstate.ActionState),
				CursorX: inputFrame.CursorX,
				CursorY: inputFrame.CursorY,
			}
		}
		inputState.Publish(inputFrame)
		now := time.Now()
		elapsed := now.Sub(lastFrame)
		lastFrame = now
		if err := scenes.Update(frameContext, elapsed); err != nil {
			reportSceneError(fmt.Errorf("updating Lua scenes: %w", err))
			return
		}
		// Updating can replace the focused scene. Refresh the persistent label so
		// composer draining and native frame work are charged to the new owner.
		frameContext = scenes.FrameContext(context.Background())
		pprof.SetGoroutineLabels(frameContext)
		if err := scenes.Render(frameContext); err != nil {
			reportSceneError(fmt.Errorf("rendering Lua scenes: %w", err))
		}
	})
	if err := renderer.AttachAudio(mixer); err != nil {
		return err
	}
	// Register composition draining after scene updates so Lua mutations are
	// visible to Raylib during the same frame.
	if err := renderer.AttachComposer(composer); err != nil {
		return err
	}
	stopShellOverlay := renderer.SubscribeOverlay(func() {
		width, height := renderer.WindowSize()
		console.Draw(width, height)
	})

	definitions, err := modruntime.DiscoverDefinitions(context.Background(), scripts, contentFS)
	for _, definition := range definitions {
		if err == nil {
			err = components.Register(definition.Managed())
		}
	}
	modDirectory := os.Getenv("DARK_MAGIC_MOD_DIRECTORY")
	if modDirectory != "" {
		modDirectory, err = darkpaths.ExpandHost(modDirectory)
	}
	if err == nil && modDirectory != "" {
		coordinator := hotreload.New(contentFS, scripts, components, gameData, definitions)
		err = components.Register(host.ManagedDefinition{
			ID: "engine.hot-reload",
			New: func(context.Context) (host.Component, error) {
				return filewatch.New(modDirectory, 250*time.Millisecond, coordinator.Reload), nil
			},
		})
	}
	desired, desiredErr := host.ParseDesired(os.Getenv("DARK_MAGIC_ENABLED_COMPONENTS"), "darkmagic.boot")
	if modDirectory != "" && desired != nil {
		desired["engine.hot-reload"] = true
	}
	if err == nil {
		err = desiredErr
	}
	if err == nil {
		err = components.ApplyDesired(context.Background(), desired)
	}
	if err == nil {
		err = scenes.Flush(context.Background())
	}
	if err == nil && startScene != "" {
		err = navigator.Replace(context.Background(), startScene)
	}
	if err != nil {
		return err
	}
	var captureSession *capture.Session
	stopCaptureFrames := func() {}
	if captureDirectory != "" {
		captureSession, err = capture.New(captureDirectory, captureScenes, captureSettle, renderer)
		if err != nil {
			return err
		}
		stopCaptureFrames = renderer.SubscribePostFrame(func() {
			captureSession.Observe(navigator.Stack())
			if captureSession.Complete() {
				stopSignals()
			}
		})
	}

	err = renderer.Run(runContext)
	select {
	case sceneErr := <-sceneErrors:
		err = errors.Join(err, sceneErr)
	default:
	}
	stopCaptureFrames()
	if captureSession != nil {
		err = errors.Join(err, captureSession.Close())
	}
	stopSceneFrames()
	stopShellOverlay()
	console.Close()

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = errors.Join(err, scenes.Close(shutdown))
	err = errors.Join(err, shellSession.Close())
	err = errors.Join(err, components.ApplyDesired(shutdown, map[string]bool{}))
	err = errors.Join(err, appHost.Stop(shutdown))
	stopped = true
	return err
}

func developmentCharacters(count int) []persistence.Character {
	if count <= 0 {
		return nil
	}
	classes := []string{"Amazon", "Sorceress", "Necromancer", "Paladin", "Barbarian", "Assassin", "Druid"}
	result := make([]persistence.Character, 0, count)
	for index := 0; index < count; index++ {
		class := classes[index%len(classes)]
		result = append(result, persistence.Character{
			ID:        fmt.Sprintf("fixture-%02d", index+1),
			Name:      fmt.Sprintf("Hero%02d", index+1),
			Class:     class,
			Level:     index + 1,
			Expansion: true,
			Hardcore:  index%3 == 2,
			Stats: &persistence.Stats{
				Experience: 1200, NextLevelExperience: 2250,
				Strength: 25, Dexterity: 20, Vitality: 25, Energy: 15,
				Defense: 42, Health: 70, MaxHealth: 70, Mana: 30, MaxMana: 30,
				Stamina: 84, MaxStamina: 84,
			},
		})
	}
	return result
}

func buildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "development"
	}
	return info.Main.Version
}

func stopHost(appHost *host.Host) {
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := appHost.Stop(shutdown); err != nil {
		slog.Error("stopping application host", "error", err)
	}
}
