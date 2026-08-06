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

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/inputcore"
	"github.com/gravestench/dark-magic/internal/modruntime"
	"github.com/gravestench/dark-magic/internal/navigation"
	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/services/gameScene"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type englishLanguage struct{}

func (englishLanguage) GetSupportedLanguages() []string { return []string{"English"} }

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
	renderer := &raylibRenderer.Service{}
	renderer.SetLogger(slog.Default().With("component", "renderer"))
	renderer.Configure(raylibRenderer.DefaultConfig())
	inputService := input.New(renderer)
	inputService.SetLogger(slog.Default().With("component", "input"))
	worldConfig := gameScene.DefaultConfig()
	worldConfig.Source = ""
	world := gameScene.New(renderer, inputService, contentFS, englishLanguage{}, worldConfig)
	world.SetLogger(slog.Default().With("component", "world"))

	scripts := modruntime.New()
	composer := &rendercore.Composer{}
	scenes := modruntime.NewScenes(scripts, navigation.New())
	inputState := &inputcore.Store{}
	if err := scripts.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.VFSModule(contentFS)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.InputModule(inputState)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(modruntime.RenderModule(scripts, composer)); err != nil {
		return err
	}
	if err := scripts.RegisterModule(scenes.Module()); err != nil {
		return err
	}

	appHost := host.New()
	for _, definition := range []host.Definition{
		{ID: "engine.renderer", Component: renderer},
		{ID: "engine.input", DependsOn: []string{"engine.renderer"}, Component: inputService},
		{ID: "game.world.compatibility", DependsOn: []string{"engine.renderer", "engine.input"}, Component: world},
		{ID: "engine.lua", DependsOn: []string{"engine.renderer", "engine.input"}, Component: scripts},
	} {
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
	// Register composition draining after scene updates so Lua mutations are
	// visible to Raylib during the same frame.
	if err := renderer.AttachComposer(composer); err != nil {
		return err
	}

	components := host.NewManager()
	boot, err := modruntime.LoadDefinition(context.Background(), scripts, contentFS, "boot.lua")
	if err == nil {
		err = components.Register(boot.Managed())
	}
	if err == nil {
		err = components.Enable(context.Background(), boot.ID)
	}
	if err == nil {
		err = scenes.Flush(context.Background())
	}
	if err != nil {
		return err
	}

	runContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err = renderer.Run(runContext)
	stopSignals()
	stopSceneFrames()

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = errors.Join(err, scenes.Close(shutdown))
	err = errors.Join(err, components.DisableCascade(shutdown, boot.ID))
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
