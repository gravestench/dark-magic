package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/faiface/mainthread"
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/host"
	"github.com/gravestench/dark-magic/internal/inputcore"
	"github.com/gravestench/dark-magic/internal/modruntime"
	"github.com/gravestench/dark-magic/internal/navigation"
	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/pkg/prettylog"
	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/cacheManager"
	"github.com/gravestench/dark-magic/pkg/services/configManager"
	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
	"github.com/gravestench/dark-magic/pkg/services/gameScene"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/locale"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
	"github.com/gravestench/dark-magic/pkg/services/luaModLoader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
	"github.com/gravestench/dark-magic/pkg/services/tweens"
	"github.com/gravestench/dark-magic/pkg/services/ui"
	"github.com/gravestench/dark-magic/pkg/services/webRouter"
	"github.com/gravestench/dark-magic/pkg/services/webServer"
)

const (
	projectName      = "Dark Magic"
	projectConfigDir = "~/.config/dark-magic"
)

func main() {
	app := servicemesh.New(projectName)

	app.SetLogHandler(prettylog.NewHandler(&slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// utility services
	//rt.Add(&modalTui.Service{})
	//app.Add(&goscript.Service{}) // WIP
	app.Add(&luaManager.Service{})
	app.Add(&cacheManager.Service{})
	app.Add(&fileLoader.Service{})
	app.Add(&fileWatcher.Service{})
	app.Add(&configManager.Service{RootDirectory: projectConfigDir})
	app.Add(&webServer.Service{})
	app.Add(&webRouter.Service{})
	app.Add(&tweens.Service{})

	// these all use the loaders and records
	app.Add(&assetLoader.Service{})
	app.Add(&recordManager.Service{})
	app.Add(&spriteManager.Service{})
	app.Add(&locale.Service{})
	//app.Add(&mapGenerator.Service{})
	//app.Add(&hero.Service{})

	// rendering-dependant services
	renderer := &raylibRenderer.Service{}
	app.Add(renderer)
	app.Add(&ui.Service{})
	inputService := &input.Service{}
	app.Add(inputService) // rendering backend also handles input
	app.Add(&gameScene.Service{})
	//app.Add(&backgroundMusic.Service{}) // rendering backend also handles audio
	//app.Add(&guiManager.Service{})
	//app.Add(&modalGameUI.Service{})
	//app.Add(&loading.Screen{})
	app.Add(&luaModLoader.Service{})

	// The new application host and script runtime intentionally coexist with
	// Service Mesh while native services are migrated incrementally. The
	// renderer still requires the process main thread.
	mainthread.Run(func() { run(app, renderer, inputService) })
}

func run(legacy servicemesh.Mesh, renderer *raylibRenderer.Service, inputService *input.Service) {
	contentFS, err := content.New(content.Layer{Name: "darkmagic", FS: content.Shim()})
	if err != nil {
		slog.Error("constructing content filesystem", "error", err)
		return
	}

	scripts := modruntime.New()
	composer := &rendercore.Composer{}
	scenes := modruntime.NewScenes(scripts, navigation.New())
	inputState := &inputcore.Store{}
	if err := scripts.RegisterInstaller(modruntime.ContentRequire(contentFS, "lua")); err != nil {
		slog.Error("registering Lua content loader", "error", err)
		return
	}
	if err := scripts.RegisterModule(modruntime.VFSModule(contentFS)); err != nil {
		slog.Error("registering Lua capability", "error", err)
		return
	}
	if err := scripts.RegisterModule(modruntime.InputModule(inputState)); err != nil {
		slog.Error("registering Lua input capability", "error", err)
		return
	}
	if err := scripts.RegisterModule(modruntime.RenderModule(scripts, composer)); err != nil {
		slog.Error("registering Lua render capability", "error", err)
		return
	}
	if err := scripts.RegisterModule(scenes.Module()); err != nil {
		slog.Error("registering Lua scene capability", "error", err)
		return
	}
	lastFrame := time.Now()
	renderer.OnFrame(func() {
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
		slog.Error("attaching render composition", "error", err)
		return
	}

	appHost := host.New()
	if err := appHost.Register(host.Definition{ID: "engine.lua", Component: scripts}); err != nil {
		slog.Error("registering application component", "error", err)
		return
	}
	if err := appHost.Start(context.Background()); err != nil {
		slog.Error("starting application host", "error", err)
		return
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
		slog.Error("starting Dark Magic shim", "error", err)
		stopHost(appHost)
		return
	}

	legacy.Run()

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := scenes.Close(shutdown); err != nil {
		slog.Error("stopping Dark Magic scenes", "error", err)
	}
	if err := components.DisableCascade(shutdown, boot.ID); err != nil {
		slog.Error("stopping Dark Magic shim", "error", err)
	}
	if err := appHost.Stop(shutdown); err != nil {
		slog.Error("stopping application host", "error", err)
	}
}

func stopHost(appHost *host.Host) {
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := appHost.Stop(shutdown); err != nil {
		slog.Error("stopping application host", "error", err)
	}
}
