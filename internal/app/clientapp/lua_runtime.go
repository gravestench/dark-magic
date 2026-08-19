package clientapp

import (
	"fmt"
	"image"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gravestench/dark-magic/internal/audio"
	d2catalog "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/catalog"
	d2presentation "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/presentation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/video"
)

// registerLuaRuntime exposes host services, video, and presentation to scripts.
func (app *application) registerLuaRuntime() error {
	if err := app.registerHostOverrideModules(); err != nil {
		return err
	}

	if err := app.registerBaseLuaModules(); err != nil {
		return err
	}

	if err := app.registerVideoModule(); err != nil {
		return err
	}

	return app.registerPresentationModules()
}

// registerHostOverrideModules installs client-owned data and catalog adapters.
func (app *application) registerHostOverrideModules() error {
	for _, module := range app.hostOverrideLuaModules() {
		if err := app.scripts.RegisterModuleOverride(module); err != nil {
			return fmt.Errorf("register Lua host override %s: %w", module.Name, err)
		}
	}

	return nil
}

// registerBaseLuaModules installs the engine services shared by client scripts.
func (app *application) registerBaseLuaModules() error {
	for _, module := range app.baseLuaModules() {
		if err := app.scripts.RegisterModule(module); err != nil {
			return fmt.Errorf("register Lua module %s: %w", module.Name, err)
		}
	}

	return nil
}

// hostOverrideLuaModules returns adapters that replace distribution defaults.
func (app *application) hostOverrideLuaModules() []modruntime.Module {
	return []modruntime.Module{
		modruntime.DataModule(
			app.options.Content,
			d2presentation.ManifestTransforms(app.profile.ID),
		),
		d2catalog.QuestModule(app.questCatalog, app.locale),
		d2catalog.MapModule(app.questCatalog),
	}
}

// baseLuaModules returns narrow doors into engine-owned services.
func (app *application) baseLuaModules() []modruntime.Module {
	return []modruntime.Module{
		// Process and content services.
		modruntime.AppModule(BuildVersion(), app.stop),
		modruntime.ShellModule(app.shellSettings),
		modruntime.VFSModule(app.options.Content),
		modruntime.SessionWorldModule(
			app.options.Content,
			app.currentWorld,
			app.worldObjectResolver,
		),
		modruntime.InputModule(app.inputState),
		modruntime.DevModule(map[string]any{
			// Lua's deterministic random API expects a positive signed 31-bit seed.
			"random_seed": int(time.Now().UnixNano()%2147483646) + 1,
		}),

		// Presentation and user-facing services.
		modruntime.AudioModule(app.scripts, app.mixer, app.options.Content),
		modruntime.SettingsModule(app.gameSettings, app.mixer, app.renderer),
		modruntime.LocaleModule(app.locale),

		// Game-session control surfaces.
		d2save.Module(app.saves),
		modruntime.PlayerControlModule(app.playerControl),
		modruntime.CommandIntentModule(app.commandIntents),
		modruntime.LoadingModule(app.loading),
		modruntime.NetworkModule(app.network),
		modruntime.RealmModule(app.realm),
	}
}

// registerVideoModule installs embedded playback and viewport resize handling.
func (app *application) registerVideoModule() error {
	backend := newClientVideoBackend(app.composer, app.mixer, app.renderWindow)
	if resizable, ok := backend.(interface{ Resize(image.Point) error }); ok {
		app.subscribeVideoViewport(resizable)
	}

	module := modruntime.VideoModule(app.scripts, backend, app.options.Content)

	return app.scripts.RegisterModule(module)
}

// subscribeVideoViewport keeps cinematic output aligned with window size.
func (app *application) subscribeVideoViewport(resizable interface{ Resize(image.Point) error }) {
	// The renderer owns this callback for the same lifetime as the video backend.
	app.renderer.SubscribeViewport(func(width, height int) {
		if err := resizable.Resize(image.Pt(width, height)); err != nil {
			slog.Error("resizing cinematic viewport", "error", err)
		}
	})
}

// newClientVideoBackend constructs the embedded video implementation.
func newClientVideoBackend(
	composer *render.Composer,
	mixer *audio.Mixer,
	window image.Point,
) video.Backend {
	return video.NewEmbeddedBackend(composer, mixer, window)
}

// registerPresentationModules exposes composition and scene navigation to Lua.
func (app *application) registerPresentationModules() error {
	renderCapability := modruntime.NewRenderCapability(
		app.scripts,
		app.composer,
		app.options.Content,
	)
	app.renderCapability = renderCapability
	app.installProfileDiagnostics(renderCapability)

	// Capability registration precedes scenes because scenes consume its output.
	modules := []modruntime.Module{renderCapability.Module(), app.scenes.Module()}
	for _, module := range modules {
		if err := app.scripts.RegisterModule(module); err != nil {
			return fmt.Errorf("register Lua module %s: %w", module.Name, err)
		}
	}

	return nil
}

// installProfileDiagnostics supplies optional profiling with live backend state.
func (app *application) installProfileDiagnostics(capability *modruntime.RenderCapability) {
	if app.options.Profile == nil {
		return
	}
	// Build diagnostics on demand so captures observe current caches and timing.
	app.options.Profile.SetDiagnostics(func() any {
		return map[string]any{
			"composition": capability.Diagnostics(),
			"render_backend": map[string]any{
				"name":        app.renderer.Name(),
				"diagnostics": app.renderer.BackendDiagnostics(),
			},
			"texture_cache": app.renderer.CacheDiagnostics(),
			"frame_timing":  app.frameMetrics.Snapshot(),
		}
	})
}

// BuildVersion turns Go build metadata into a friendly script-facing string.
func BuildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "development"
	}

	return info.Main.Version
}
