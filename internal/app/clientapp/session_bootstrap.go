package clientapp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/gravestench/dark-magic/internal/game/simulation"
	loadcore "github.com/gravestench/dark-magic/internal/loading"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
)

// sessionInitialData returns deterministic bootstrap data consumed by d2legacy.
func (app *application) sessionInitialData() map[string]any {
	return map[string]any{
		"engine.game_data_generation_id": app.gameDataGenerationID(),
		"d2legacy.game_rules": map[string]any{
			"target":          "lod-1.14d",
			"expansion":       true,
			"difficulty":      0,
			"hardcore":        false,
			"ladder":          false,
			"maximum_players": 8,
		},
		"d2legacy.development_items": map[string]any{
			"enabled":                 app.options.FixtureCharacters > 0,
			"create_empty_containers": app.options.FixtureCharacters == 0,
		},
		"d2legacy.development_skills": app.developmentSkillsBootstrapData(),
		"d2legacy.interactions":       app.interactionBootstrapData(),
		"d2legacy.world_transitions":  app.transitionBootstrapData(),
		"d2legacy.world_warps":        app.warpBootstrapData(),
	}
}

// queueEntryPopulation submits initial world population after handlers start.
func (app *application) queueEntryPopulation() error {
	command, err := app.populationBootstrapCommand()
	if err != nil {
		return err
	}

	return wrap("queue d2legacy entry population", app.offlineSession.Submit(command))
}

// populationBootstrapCommand creates the authoritative entry-world command.
func (app *application) populationBootstrapCommand() (simulation.Command, error) {
	nearby := developmentScenes[app.options.StartScene].nearbyHostiles
	return app.preparedEntryWorld().PopulationCommand(nearby)
}

// interactionBootstrapData selects the optional direct-start NPC interaction.
func (app *application) interactionBootstrapData() map[string]any {
	initial := ""
	if app.options.StartScene == "vendor" {
		initial = "act1-akara"
	}

	return entryworld.InteractionData(app.gameWorlds, app.gameWorldZones, "local-player", initial)
}

// preparedEntryWorld gathers the maps and admission positions for bootstrapping.
func (app *application) preparedEntryWorld() *entryworld.Prepared {
	return &entryworld.Prepared{
		Worlds: app.gameWorlds,
		Zones:  app.gameWorldZones,
		Spawns: app.gameWorldSpawns,
		Seam:   app.transitionSeam,
	}
}

// buildLoadingCoordinator declares the prerequisites for entering game scenes.
func (app *application) buildLoadingCoordinator() error {
	app.loading = loadcore.New(map[string]loadcore.Task{
		"selected_character": app.requireSelectedCharacter,
		"loading_assets":     app.requireLoadingAssets,
		"world":              loadingWorldReady,
	})

	return nil
}

// requireSelectedCharacter accepts either local or authenticated realm selection.
func (app *application) requireSelectedCharacter(context.Context) error {
	if _, ok := app.saves.Selected(); ok {
		return nil
	}

	if app.network != nil && app.network.hasSelectedCharacter() {
		return nil
	}

	return errors.New("no character is selected")
}

// requireLoadingAssets verifies every presentation bootstrap dependency.
func (app *application) requireLoadingAssets(context.Context) error {
	for _, name := range app.presentation.LoadingAssets {
		if _, err := fs.Stat(app.options.Content, name); err != nil {
			return fmt.Errorf("load dependency %q: %w", name, err)
		}
	}

	return nil
}

// loadingWorldReady marks world preparation as synchronous at composition time.
func loadingWorldReady(context.Context) error {
	return nil
}

// fixtureNeedsSelection reports whether direct scene entry requires a player.
func fixtureNeedsSelection(scene string) bool {
	switch scene {
	case "game_world",
		"game_loading",
		"combat_lab",
		"spell_lab",
		"warp_lab",
		"inventory",
		"character",
		"skills",
		"vendor":
		return true
	default:
		return false
	}
}
