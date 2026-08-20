package clientapp

import (
	"fmt"
	"image"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/mapeditor"
	d2catalog "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/catalog"
	d2presentation "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/presentation"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
	"github.com/gravestench/dark-magic/internal/video"
)

// registerLuaRuntime installs capabilities in dependency order before scripts
// start. A failed registration aborts composition so Lua never sees a partial host API.
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

// registerHostOverrideModules replaces distribution defaults with adapters tied
// to this application's pinned records and state; duplicates are therefore intentional.
func (app *application) registerHostOverrideModules() error {
	for _, module := range app.hostOverrideLuaModules() {
		if err := app.scripts.RegisterModuleOverride(module); err != nil {
			return fmt.Errorf("register Lua host override %s: %w", module.Name, err)
		}
	}

	return nil
}

// registerBaseLuaModules exposes narrow engine capabilities rather than the
// application object, limiting script authority to explicitly reviewed doors.
func (app *application) registerBaseLuaModules() error {
	var storage *mapeditor.Storage
	if app.options.MapEditorOutput != "" {
		var err error
		storage, err = mapeditor.NewStorage(app.options.MapEditorOutput, app.options.MapEditorReadOnlyRoots...)
		if err != nil {
			return fmt.Errorf("configure map editor output: %w", err)
		}
	}
	for _, module := range app.baseLuaModules(storage) {
		if err := app.scripts.RegisterModule(module); err != nil {
			return fmt.Errorf("register Lua module %s: %w", module.Name, err)
		}
	}

	return nil
}

// hostOverrideLuaModules lists modules whose behavior must follow live client
// composition instead of package-provided defaults, such as current world and saves.
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

// baseLuaModules declares the stable script-facing capability surface. Ordering
// here is registration order, not an invitation for modules to depend implicitly.
func (app *application) baseLuaModules(mapEditorStorage *mapeditor.Storage) []modruntime.Module {
	return []modruntime.Module{
		// Process and content services.
		modruntime.AppModule(BuildVersion(), app.stop),
		modruntime.DisplayModule(app.renderer.Resolution),
		modruntime.ShellModule(app.shellSettings),
		modruntime.VFSModule(app.options.Content),
		modruntime.MapEditorModule(app.options.Content, mapEditorStorage, app.worldObjectResolver),
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

// registerVideoModule selects one native decoder and attaches resize handling only
// after Lua registration succeeds, avoiding callbacks into an unavailable module.
func (app *application) registerVideoModule() error {
	backend := newClientVideoBackend(app.composer, app.mixer, app.renderWindow)
	if resizable, ok := backend.(interface{ Resize(image.Point) error }); ok {
		app.subscribeVideoViewport(resizable)
	}

	module := modruntime.VideoModule(app.scripts, backend, app.options.Content)

	return app.scripts.RegisterModule(module)
}

// subscribeVideoViewport stores the unsubscribe callback on application so resize
// delivery is detached before the video backend closes during shutdown.
func (app *application) subscribeVideoViewport(resizable interface{ Resize(image.Point) error }) {
	// The renderer owns this callback for the same lifetime as the video backend.
	app.renderer.SubscribeViewport(func(width, height int) {
		if err := resizable.Resize(image.Pt(width, height)); err != nil {
			slog.Error("resizing cinematic viewport", "error", err)
		}
	})
}

// newClientVideoBackend contains the concrete decoder choice at composition time,
// keeping modruntime dependent only on its video capability contract.
func newClientVideoBackend(
	composer *render.Composer,
	mixer *audio.Mixer,
	window image.Point,
) video.Backend {
	return video.NewEmbeddedBackend(composer, mixer, window)
}

// registerPresentationModules gives Lua presentation control but not simulation
// authority; navigation and composition operate on views produced by native state.
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

// installProfileDiagnostics adds renderer diagnostics only when profiling exists,
// keeping production startup free of diagnostic probes and their retention costs.
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

// BuildVersion reports module version plus source revision when available, giving
// scripts and diagnostics a reproducible identity without shelling out to Git.
func BuildVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "development"
	}

	return info.Main.Version
}
