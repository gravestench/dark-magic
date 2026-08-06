package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/faiface/mainthread"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/filewatch"
	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/hotreload"
	"github.com/gravestench/dark-magic/internal/inputcore"
	"github.com/gravestench/dark-magic/internal/localecore"
	"github.com/gravestench/dark-magic/internal/modruntime"
	"github.com/gravestench/dark-magic/internal/navigation"
	input "github.com/gravestench/dark-magic/internal/raylib/input"
	raylibRenderer "github.com/gravestench/dark-magic/internal/raylib/renderer"
	gameScene "github.com/gravestench/dark-magic/internal/raylib/world"
	"github.com/gravestench/dark-magic/internal/recordstore"
	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/internal/runtimeapi"
	"github.com/gravestench/dark-magic/internal/savecore"
	darkpaths "github.com/gravestench/dark-magic/pkg/paths"
	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/scene"
)

func main() {
	slog.SetDefault(slog.New(prettylog.NewHandler(&slog.HandlerOptions{Level: slog.LevelDebug})))
	contentFS, err := content.FromEnvironment()
	if err != nil {
		slog.Error("constructing content filesystem", "error", err)
		return
	}
	mainthread.Run(func() {
		if err := run(contentFS); err != nil {
			slog.Error("running Dark Magic", "error", err)
		}
	})
}

func run(contentFS *content.FS) error {
	runContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	renderer := &raylibRenderer.Service{}
	renderer.SetLogger(slog.Default().With("component", "renderer"))
	renderer.Configure(raylibRenderer.DefaultConfig())
	inputService := input.New(renderer)
	inputService.SetLogger(slog.Default().With("component", "input"))
	locale := localecore.New(contentFS, "English")
	worldConfig := gameScene.DefaultConfig()
	worldConfig.Source = ""
	world := gameScene.New(renderer, inputService, contentFS, locale, worldConfig)
	world.SetLogger(slog.Default().With("component", "world"))

	scripts := modruntime.New()
	composer := &rendercore.Composer{}
	mixer := &audiocore.Mixer{}
	scenes := modruntime.NewScenes(scripts, navigation.New())
	inputState := &inputcore.Store{}
	records := recordstore.New(contentFS)
	saves := savecore.New(savecore.Character{ID: "default-amazon", Name: "Dark Wanderer", Class: "Amazon", Level: 1})
	simulation := modruntime.NewSimulation(scene.New(1, 4096, 4096))
	components := host.NewManager()
	if err := scripts.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.AppModule("development", stopSignals)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.VFSModule(contentFS)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.DataModule(contentFS)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.InputModule(inputState)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.AudioModule(scripts, mixer, contentFS)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.RecordsModule(records)); err != nil {
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
	if err := scripts.RegisterModule(modruntime.RenderModuleWithAssets(scripts, composer, contentFS)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(scenes.Module()); err != nil {
		return err
	}

	appHost := host.New()
	staticDefinitions := []host.Definition{
		{ID: "engine.renderer", Component: renderer},
		{ID: "engine.input", DependsOn: []string{"engine.renderer"}, Component: inputService},
		{ID: "game.world.compatibility", DependsOn: []string{"engine.renderer", "engine.input"}, Component: world},
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

	lastFrame := time.Now()
	stopSceneFrames := renderer.SubscribeFrame(func() {
		inputState.Publish(inputService.Snapshot())
		now := time.Now()
		elapsed := now.Sub(lastFrame)
		lastFrame = now
		if err := scenes.Update(context.Background(), elapsed); err != nil {
			slog.Error("updating Lua scenes", "error", err)
		}
		if err := scenes.Render(context.Background()); err != nil {
			slog.Error("rendering Lua scenes", "error", err)
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
		coordinator := hotreload.New(contentFS, scripts, components, records, definitions)
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
	if err != nil {
		return err
	}

	err = renderer.Run(runContext)
	stopSceneFrames()

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = errors.Join(err, scenes.Close(shutdown))
	err = errors.Join(err, components.ApplyDesired(shutdown, map[string]bool{}))
	err = errors.Join(err, appHost.Stop(shutdown))
	stopped = true
	return err
}

func stopHost(appHost *host.Host) {
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := appHost.Stop(shutdown); err != nil {
		slog.Error("stopping application host", "error", err)
	}
}
