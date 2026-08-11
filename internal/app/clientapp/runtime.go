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
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/video"
)

func (app *application) registerLuaRuntime() error {
	if err := app.scripts.RegisterInstaller(modruntime.ContentRequire(app.options.Content, "lua")); err != nil {
		return err
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

// baseLuaModules are small doors into engine-owned services. Lua receives the
// doors, not the service internals behind them.
func (app *application) baseLuaModules() []modruntime.Module {
	return []modruntime.Module{
		modruntime.AppModule(BuildVersion(), app.stop),
		modruntime.ShellModule(app.shellSettings),
		modruntime.VFSModule(app.options.Content),
		modruntime.DataModule(app.options.Content, app.profile.ID),
		modruntime.WorldModule(app.options.Content, app.worldObjectResolver),
		modruntime.InputModule(app.inputState),
		modruntime.DevModule(map[string]any{
			"random_seed":     int(time.Now().UnixNano()%2147483646) + 1,
			"composite_token": app.options.CompositeLab.Token, "composite_mode": app.options.CompositeLab.Mode,
			"composite_weapon": app.options.CompositeLab.WeaponClass, "composite_direction": app.options.CompositeLab.Direction,
			"composite_frame": app.options.CompositeLab.Frame, "composite_components": app.options.CompositeLab.Components,
			"composite_random": app.options.CompositeLab.Random,
			"dt1_path":         app.options.DT1Lab.Path, "dt1_palette": app.options.DT1Lab.Palette,
			"dt1_view": app.options.DT1Lab.View, "dt1_tile": app.options.DT1Lab.Tile,
			"ds1_path": app.options.DS1Lab.Path, "ds1_tiles": app.options.DS1Lab.Tiles,
			"ds1_palette": app.options.DS1Lab.Palette,
		}),
		modruntime.AudioModule(app.scripts, app.mixer, app.options.Content, app.gameData),
		modruntime.SettingsModule(app.gameSettings, app.mixer, app.renderer),
		modruntime.RecordsModule(app.gameData),
		modruntime.GameDataModule(app.gameData),
		modruntime.QuestCatalogModule(app.questCatalog, app.locale),
		modruntime.MapCatalogModule(app.questCatalog),
		modruntime.LocaleModule(app.locale),
		modruntime.SaveModule(app.saves),
		modruntime.PlayerControlModule(app.playerControl),
		modruntime.InteractionModule(app.interactionAuthority, app.interactionControl, "local-player"),
		modruntime.ItemModule(app.itemAuthority, app.itemControl, "local-player"),
		modruntime.NewECSCapability(app.scripts, app.entitySimulation).Module(),
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
