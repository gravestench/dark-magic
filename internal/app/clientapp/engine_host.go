package clientapp

import (
	"context"
	"fmt"
	"os"

	"github.com/gravestench/dark-magic/internal/app/host"
	"github.com/gravestench/dark-magic/internal/app/runtimeapi"
)

// startEngineHost delegates lifecycle ordering to host after every definition is
// known. No engine service may start before dependencies can be validated as a graph.
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

// engineHostDefinitions is the executable dependency graph for native services.
// The edges also determine safe reverse shutdown order, so changes have lifecycle implications.
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
