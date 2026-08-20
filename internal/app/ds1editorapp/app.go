// Package ds1editorapp assembles the standalone DS1 authoring tool. It is
// intentionally separate from clientapp: maps are content artifacts, not a
// running game session.
package ds1editorapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/inputstate"
	"github.com/gravestench/dark-magic/internal/mapeditor"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
	"github.com/gravestench/dark-magic/internal/platform/desktop"
	"github.com/gravestench/dark-magic/internal/presentation/navigation"
	"github.com/gravestench/dark-magic/internal/presentation/render"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// Options names the authoring-only policy supplied by tools/ds1-editor.
// Content remains mounted read-only; OutputDirectory is the only write root.
type Options struct {
	Content            *content.FS
	OutputDirectory    string
	ReadOnlyRoots      []string
	WindowWidth        int
	WindowHeight       int
	Borderless         bool
	OutputPalette      string
	DisableNativeAudio bool
}

// Run owns a single editor process and never initializes client sessions,
// records, network transport, saves, ECS, or gameplay systems.
func Run(options Options) error {
	app, err := newApplication(options)
	if err != nil {
		return err
	}
	defer app.stop()
	defer app.shutdown()

	if err := app.assemble(); err != nil {
		return err
	}

	runErr := app.renderer.Run(app.ctx)
	select {
	case sceneErr := <-app.sceneError:
		return errors.Join(runErr, sceneErr)
	default:
		return runErr
	}
}

type application struct {
	options Options
	ctx     context.Context
	stop    context.CancelFunc

	renderer   desktop.Renderer
	input      desktop.Input
	scripts    *modruntime.Runtime
	scenes     *modruntime.Scenes
	navigator  *navigation.Manager
	inputState *inputstate.Store
	composer   *render.Composer
	host       *host.Host
	components *host.Manager
	stopFrame  func()
	lastFrame  time.Time
	sceneError chan error
}

// newApplication validates process-wide policy and creates the signal-owned lifetime.
// Resource construction remains deferred so a failed option cannot leak a desktop backend.
func newApplication(options Options) (*application, error) {
	if options.Content == nil {
		return nil, errors.New("DS1 editor: content filesystem is required")
	}
	if options.OutputDirectory == "" {
		return nil, errors.New("DS1 editor: output directory is required")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	return &application{
		options:    options,
		ctx:        ctx,
		stop:       stop,
		stopFrame:  func() {},
		sceneError: make(chan error, 1),
	}, nil
}

// assemble constructs the tool from the platform boundary inward.
// The order ensures Lua never observes a renderer or input service that failed to start.
func (app *application) assemble() error {
	if err := app.buildDesktop(); err != nil {
		return err
	}
	if err := app.configureLua(); err != nil {
		return err
	}
	if err := app.startHost(); err != nil {
		return err
	}
	return app.attachFrameLoop()
}

// buildDesktop creates the renderer-facing services without starting them.
// Native resolution is mandatory because editor chrome must track the actual workspace size.
func (app *application) buildDesktop() error {
	options := desktop.DefaultOptions()
	options.Content = app.options.Content
	options.WindowTitle = "Dark Magic — DS1 Editor"
	options.WindowWidth = app.options.WindowWidth
	options.WindowHeight = app.options.WindowHeight
	options.NativeResolution = true
	options.ViewportFit = "contain"
	options.BorderlessFullscreen = app.options.Borderless
	options.ShowSystemCursor = false
	options.NativeAudio = !app.options.DisableNativeAudio
	options.PalettePath = app.options.OutputPalette
	options.Logger = slog.Default().With("component", "ds1-editor")
	if options.WindowWidth <= 0 {
		options.WindowWidth = 1280
	}
	if options.WindowHeight <= 0 {
		options.WindowHeight = 760
	}

	bundle, err := desktop.New(options)
	if err != nil {
		return fmt.Errorf("create DS1 editor desktop: %w", err)
	}
	app.renderer, app.input = bundle.Renderer, bundle.Input
	app.inputState = &inputstate.Store{}
	app.composer = &render.Composer{}
	app.scripts = modruntime.New()
	app.navigator = navigation.New()
	app.scenes = modruntime.NewScenes(app.scripts, app.navigator)
	app.scenes.SetInputStore(app.inputState)

	return nil
}

// configureLua exposes only the capabilities needed by the authoring scene.
// PackageRequire restricts imports to the dedicated editor namespace instead of the game mod runtime.
func (app *application) configureLua() error {
	storage, err := mapeditor.NewStorage(app.options.OutputDirectory, app.options.ReadOnlyRoots...)
	if err != nil {
		return fmt.Errorf("configure DS1 editor output: %w", err)
	}
	installer := modruntime.PackageRequire(app.options.Content, []string{"ds1editor"})
	if err := app.scripts.RegisterInstaller(installer); err != nil {
		return fmt.Errorf("register DS1 editor Lua namespace: %w", err)
	}
	renderCapability := modruntime.NewRenderCapability(app.scripts, app.composer, app.options.Content)
	for _, module := range []modruntime.Module{
		modruntime.DisplayModule(app.renderer.Resolution),
		modruntime.DataModule(app.options.Content),
		modruntime.VFSModule(app.options.Content),
		modruntime.InputModule(app.inputState),
		modruntime.MapEditorModule(app.options.Content, storage),
		renderCapability.Module(),
		app.scenes.Module(),
	} {
		if err := app.scripts.RegisterModule(module); err != nil {
			return fmt.Errorf("register DS1 editor Lua module %s: %w", module.Name, err)
		}
	}
	return nil
}

// startHost starts the small platform dependency graph and attaches its retained composer.
// Tool-specific component IDs keep diagnostics distinct from the game application's host.
func (app *application) startHost() error {
	app.host = host.New()
	for _, definition := range []host.Definition{
		{ID: "tool.renderer", Component: app.renderer},
		{ID: "tool.input", DependsOn: []string{"tool.renderer"}, Component: app.input},
		{ID: "tool.lua", DependsOn: []string{"tool.renderer", "tool.input"}, Component: app.scripts},
	} {
		if err := app.host.Register(definition); err != nil {
			return err
		}
	}
	if err := app.host.Start(app.ctx); err != nil {
		return err
	}
	return app.renderer.AttachComposer(app.composer)
}

// startEditorScene loads the one synthetic bootstrap that enters the namespaced authoring screen.
// Deferring this until the first frame gives resize and high-DPI state time to settle.
func (app *application) startEditorScene() error {
	definition, err := modruntime.LoadDefinition(app.ctx, app.scripts, fstest.MapFS{
		"editor.lua": &fstest.MapFile{Data: []byte(editorBootstrap)},
	}, "editor.lua")
	if err != nil {
		return fmt.Errorf("load DS1 editor scene: %w", err)
	}
	app.components = host.NewManager()
	if err := app.components.Register(definition.Managed()); err != nil {
		return err
	}
	if err := app.components.ApplyDesired(app.ctx, map[string]bool{definition.ID: true}); err != nil {
		return fmt.Errorf("start DS1 editor scene: %w", err)
	}
	return app.scenes.Flush(app.ctx)
}

// attachFrameLoop publishes input and advances the authoring scene once per renderer frame.
// Scene errors cancel the process and cross the callback boundary through a single buffered result.
func (app *application) attachFrameLoop() error {
	app.lastFrame = time.Now()
	app.stopFrame = app.renderer.SubscribeFrame(func() {
		// Raylib finalizes a resized/high-DPI drawable immediately before frame
		// subscribers run. Deferring scene construction until here prevents the
		// editor chrome from being retained at the provisional launch size and
		// needing a user maximize/resize to look correct.
		if app.components == nil {
			if err := app.startEditorScene(); err != nil {
				app.reportError(err)
			}
			return
		}
		frame := app.input.Snapshot()
		frame.Owner = inputstate.FocusOwner{Domain: inputstate.FocusScene, ID: "map_editor"}
		app.inputState.Publish(frame)
		frameContext := app.scenes.FrameContext(app.ctx)
		elapsed := time.Since(app.lastFrame)
		app.lastFrame = time.Now()
		if err := app.scenes.Update(frameContext, elapsed); err != nil {
			app.reportError(fmt.Errorf("updating DS1 editor: %w", err))
			return
		}
		if err := app.scenes.Render(frameContext); err != nil {
			app.reportError(fmt.Errorf("rendering DS1 editor: %w", err))
		}
	})
	return nil
}

// reportError publishes the first frame-callback failure and initiates orderly shutdown.
// Later errors are ignored because the first failure is the useful causal signal.
func (app *application) reportError(err error) {
	select {
	case app.sceneError <- err:
		app.stop()
	default:
	}
}

// shutdown closes scene, component, and platform owners in dependency order.
// Nil checks make cleanup safe after any partially completed assembly phase.
func (app *application) shutdown() {
	app.stopFrame()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if app.scenes != nil {
		_ = app.scenes.Close(ctx)
	}
	if app.components != nil {
		_ = app.components.ApplyDesired(ctx, map[string]bool{})
	}
	if app.host != nil {
		_ = app.host.Stop(ctx)
	}
}

// editorBootstrap deliberately registers just one authoring scene. It does not
// load d2legacy.boot or its scene registry, which would bring game menus and
// gameplay presentation into this standalone tool.
const editorBootstrap = `
local scenes = require("engine.scene/v1")
local cursor = require("ds1editor.ui.cursor")
local editor = require("ds1editor.screens.map_editor")
return {
  id = "tools.ds1_editor",
  start = function()
    scenes.register("map_editor", cursor.wrap(editor))
    scenes.replace("map_editor")
  end,
}
`

// DefaultOutputDirectory provides a user-owned location outside mounted game
// data when the command has not received an explicit --output path.
func DefaultOutputDirectory() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "dark-magic", "maps"), nil
}

// MountedRoots resolves comma-separated source mounts for write protection.
func MountedRoots(value string) ([]string, error) {
	if value == "" {
		return nil, nil
	}
	var roots []string
	for _, entry := range strings.Split(value, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		expanded, err := darkpaths.ExpandHost(entry)
		if err != nil {
			return nil, fmt.Errorf("expand mounted game-data root %q: %w", entry, err)
		}
		resolved, err := filepath.EvalSymlinks(expanded)
		if err != nil {
			return nil, fmt.Errorf("resolve mounted game-data root %q: %w", entry, err)
		}
		roots = append(roots, resolved)
	}
	return roots, nil
}
