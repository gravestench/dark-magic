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

// sessionInitialData serializes native policy and generated geometry into the
// deterministic input hashed/consumed by the d2legacy authoritative runtime.
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

// queueEntryPopulation runs only after authoritative component handlers start;
// submitting earlier would acknowledge a bootstrap command no system can consume.
func (app *application) queueEntryPopulation() error {
	command, err := app.populationBootstrapCommand()
	if err != nil {
		return err
	}

	return wrap("queue d2legacy entry population", app.offlineSession.Submit(command))
}

// populationBootstrapCommand delegates entity construction to the entry-world
// adapter while applying only the development scene's explicit hostile policy.
func (app *application) populationBootstrapCommand() (simulation.Command, error) {
	nearby := developmentScenes[app.options.StartScene].nearbyHostiles
	return app.preparedEntryWorld().PopulationCommand(nearby)
}

// interactionBootstrapData keeps synthetic vendor startup isolated to the vendor
// fixture; normal sessions begin without an implicitly opened interaction.
func (app *application) interactionBootstrapData() map[string]any {
	initial := ""
	if app.options.StartScene == "vendor" {
		initial = "act1-akara"
	}

	return entryworld.InteractionData(app.gameWorlds, app.gameWorldZones, "local-player", initial)
}

// preparedEntryWorld presents already published maps as one coherent value to
// bootstrap helpers, avoiding a second generation or divergent spawn calculation.
func (app *application) preparedEntryWorld() *entryworld.Prepared {
	return &entryworld.Prepared{
		Worlds: app.gameWorlds,
		Zones:  app.gameWorldZones,
		Spawns: app.gameWorldSpawns,
		Seam:   app.transitionSeam,
	}
}

// buildLoadingCoordinator makes scene admission depend on character, assets, and
// world readiness through named tasks that the loading UI can report independently.
func (app *application) buildLoadingCoordinator() error {
	app.loading = loadcore.New(map[string]loadcore.Task{
		"selected_character": app.requireSelectedCharacter,
		"loading_assets":     app.requireLoadingAssets,
		"world":              loadingWorldReady,
	})

	return nil
}

// requireSelectedCharacter recognizes the two trusted identity paths: a native
// selected save or a Realm-authenticated admission. Presentation state is insufficient.
func (app *application) requireSelectedCharacter(context.Context) error {
	if _, ok := app.saves.Selected(); ok {
		return nil
	}

	if app.network != nil && app.network.hasSelectedCharacter() {
		return nil
	}

	return errors.New("no character is selected")
}

// requireLoadingAssets fails before scene entry if the selected presentation
// profile references content absent from the pinned filesystem.
func (app *application) requireLoadingAssets(context.Context) error {
	for _, name := range app.presentation.LoadingAssets {
		if _, err := fs.Stat(app.options.Content, name); err != nil {
			return fmt.Errorf("load dependency %q: %w", name, err)
		}
	}

	return nil
}

// loadingWorldReady documents that world generation completed during assembly;
// the named no-op still exposes that prerequisite to loading orchestration.
func loadingWorldReady(context.Context) error {
	return nil
}

// fixtureNeedsSelection is the allowlist for bypassing frontend navigation while
// still requiring a deterministic selected character for player-dependent scenes.
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
