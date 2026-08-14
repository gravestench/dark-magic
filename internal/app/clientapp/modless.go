package clientapp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gravestench/dark-magic/internal/app/host"
	raylibrenderer "github.com/gravestench/dark-magic/internal/platform/raylib/renderer"
)

// assembleModless starts only the engine-owned native shell. A mod profile may
// intentionally contain no packages; that state must remain recoverable so a
// broken or disabled game mod can never make the client itself unstartable.
func (app *application) assembleModless() error {
	slog.Info("starting client with no enabled mods")
	app.renderer = &raylibrenderer.Service{}
	app.renderer.SetLogger(slog.Default().With("component", "renderer"))
	app.rendererConfig = raylibrenderer.DefaultConfig()
	app.rendererConfig.Resolution.Fit = app.options.ViewportFit
	app.rendererConfig.Window.Borderless = app.options.BorderlessFullscreen
	app.rendererConfig.Window.ShowSystemCursor = true
	app.renderer.Configure(app.rendererConfig)

	app.engineHost = host.New()
	if err := app.engineHost.Register(host.Definition{ID: "engine.renderer", Component: app.renderer}); err != nil {
		return fmt.Errorf("register mod-neutral renderer: %w", err)
	}
	if err := app.engineHost.Start(context.Background()); err != nil {
		return fmt.Errorf("start mod-neutral client shell: %w", err)
	}
	return nil
}
