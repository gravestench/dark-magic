package clientapp

import (
	"context"
	"fmt"
	"image"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/runtimeapi"
	d2catalog "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/catalog"
	d2presentation "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/presentation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/video"
)

func (app *application) registerLuaRuntime() error {
	for _, module := range app.hostOverrideLuaModules() {
		if err := app.scripts.RegisterModuleOverride(module); err != nil {
			return fmt.Errorf("register Lua host override %s: %w", module.Name, err)
		}
	}
	for _, module := range app.baseLuaModules() {
		if err := app.scripts.RegisterModule(module); err != nil {
			return fmt.Errorf("register Lua module %s: %w", module.Name, err)
		}
	}
	if err := app.registerVideoModule(); err != nil {
		return err
	}
	return app.registerPresentationModules()
}

func (app *application) hostOverrideLuaModules() []modruntime.Module {
	return []modruntime.Module{
		d2catalog.QuestModule(app.questCatalog, app.locale),
		d2catalog.MapModule(app.questCatalog),
	}
}

// baseLuaModules are small doors into engine-owned services. Lua receives the
// doors, not the service internals behind them.
func (app *application) baseLuaModules() []modruntime.Module {
	return []modruntime.Module{
		modruntime.AppModule(BuildVersion(), app.stop),
		modruntime.ShellModule(app.shellSettings),
		modruntime.VFSModule(app.options.Content),
		modruntime.DataModule(app.options.Content, d2presentation.ManifestTransforms(app.profile.ID)),
		modruntime.SessionWorldModule(app.options.Content, app.currentWorld, app.worldObjectResolver),
		modruntime.InputModule(app.inputState),
		modruntime.DevModule(map[string]any{
			"random_seed": int(time.Now().UnixNano()%2147483646) + 1,
		}),
		modruntime.AudioModule(app.scripts, app.mixer, app.options.Content),
		modruntime.SettingsModule(app.gameSettings, app.mixer, app.renderer),
		modruntime.LocaleModule(app.locale),
		d2save.Module(app.saves),
		modruntime.PlayerControlModule(app.playerControl),
		modruntime.CommandIntentModule(app.commandIntents),
		modruntime.LoadingModule(app.loading),
	}
}

func (app *application) registerVideoModule() error {
	window := image.Pt(app.rendererConfig.Window.Width, app.rendererConfig.Window.Height)
	backend := video.Backend(video.NewEmbeddedBackend(app.composer, app.mixer, window))
	if !backend.Available() {
		backend = video.FFplay{}
	}
	if resizable, ok := backend.(interface{ Resize(image.Point) error }); ok {
		app.renderer.SubscribeViewport(func(width, height int) {
			if err := resizable.Resize(image.Pt(width, height)); err != nil {
				slog.Error("resizing cinematic viewport", "error", err)
			}
		})
	}
	return app.scripts.RegisterModule(modruntime.VideoModule(app.scripts, backend, app.options.Content))
}

func (app *application) registerPresentationModules() error {
	renderCapability := modruntime.NewRenderCapability(app.scripts, app.composer, app.options.Content)
	app.renderCapability = renderCapability
	if app.options.Profile != nil {
		app.options.Profile.SetDiagnostics(func() any {
			return map[string]any{
				"composition":    renderCapability.Diagnostics(),
				"raylib_backend": app.renderer.BackendDiagnostics(),
				"texture_cache":  app.renderer.CacheDiagnostics(),
			}
		})
	}
	for _, module := range []modruntime.Module{renderCapability.Module(), app.scenes.Module()} {
		if err := app.scripts.RegisterModule(module); err != nil {
			return fmt.Errorf("register Lua module %s: %w", module.Name, err)
		}
	}
	return nil
}

func (app *application) startEngineHost() error {
	app.components = host.NewManager()
	app.engineHost = host.New()
	definitions := []host.Definition{
		{ID: "engine.renderer", Component: app.renderer},
		{ID: "engine.input", DependsOn: []string{"engine.renderer"}, Component: app.input},
		{ID: "engine.lua", DependsOn: []string{"engine.renderer", "engine.input"}, Component: app.scripts},
	}
	if address := os.Getenv("DARK_MAGIC_DEBUG_ADDR"); address != "" {
		definitions = append(definitions, host.Definition{ID: "engine.runtime-api", Component: runtimeapi.New(address, app.components)})
	}
	for _, definition := range definitions {
		if err := app.engineHost.Register(definition); err != nil {
			return fmt.Errorf("register host component %s: %w", definition.ID, err)
		}
	}
	return app.engineHost.Start(context.Background())
}

// BuildVersion turns Go's build metadata into a friendly string for scripts.
func BuildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "development"
	}
	return info.Main.Version
}
