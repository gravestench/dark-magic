package clientapp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/platform/desktop"
)

// assembleModless starts only the engine-owned native shell. A mod profile may
// intentionally contain no packages; that state must remain recoverable so a
// broken or disabled game mod can never make the client itself unstartable.
func (app *application) assembleModless() error {
	slog.Info("starting client with no enabled mods")
	options := desktop.DefaultOptions()
	options.ViewportFit = app.options.ViewportFit
	options.BorderlessFullscreen = app.options.BorderlessFullscreen
	options.ShowSystemCursor = true
	options.Logger = slog.Default().With("component", "renderer")
	bundle, err := desktop.New(options)
	if err != nil {
		return fmt.Errorf("construct mod-neutral renderer: %w", err)
	}
	app.renderer = bundle.Renderer

	app.engineHost = host.New()
	if err := app.engineHost.Register(host.Definition{ID: "engine.renderer", Component: app.renderer}); err != nil {
		return fmt.Errorf("register mod-neutral renderer: %w", err)
	}
	if err := app.engineHost.Start(context.Background()); err != nil {
		return fmt.Errorf("start mod-neutral client shell: %w", err)
	}
	return nil
}
