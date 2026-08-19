package clientapp

import (
	"context"
	"fmt"
	"os"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/runtimeapi"
)

// startEngineHost registers and starts renderer, input, Lua, and debug services.
func (app *application) startEngineHost() error {
	app.components = host.NewManager()
	app.engineHost = host.New()

	for _, definition := range app.engineHostDefinitions() {
		if err := app.engineHost.Register(definition); err != nil {
			return fmt.Errorf("register host component %s: %w", definition.ID, err)
		}
	}

	return app.engineHost.Start(context.Background())
}

// engineHostDefinitions declares startup dependencies between engine services.
func (app *application) engineHostDefinitions() []host.Definition {
	definitions := []host.Definition{
		{ID: "engine.renderer", Component: app.renderer},
		{
			ID:        "engine.input",
			DependsOn: []string{"engine.renderer"},
			Component: app.input,
		},
		{
			ID:        "engine.lua",
			DependsOn: []string{"engine.renderer", "engine.input"},
			Component: app.scripts,
		},
	}

	if address := os.Getenv("DARK_MAGIC_DEBUG_ADDR"); address != "" {
		definitions = append(definitions, host.Definition{
			ID:        "engine.runtime-api",
			Component: runtimeapi.New(address, app.components),
		})
	}

	return definitions
}
