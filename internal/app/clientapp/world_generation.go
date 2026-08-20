package clientapp

import (
	"fmt"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	entryworld "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/entryworld"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// buildEntryWorld prepares town and wilderness as one unit, validates the requested fixture level,
// and publishes only a complete result. A failed generation cannot replace the currently usable
// world with partial maps.
func (app *application) buildEntryWorld() error {
	prepared, err := app.prepareEntryWorld()
	if err != nil {
		return err
	}

	spawns, err := app.resolveEntryWorldSpawns(prepared)
	if err != nil {
		return err
	}

	activeLevel := app.fixtureActiveLevel()

	activeWorld := prepared.Worlds[activeLevel]
	if activeWorld == nil {
		return fmt.Errorf("development fixture world level %d is unavailable", activeLevel)
	}

	pointerAcceptance := app.newFixturePointerAcceptance(activeWorld, spawns[activeLevel])
	app.publishEntryWorld(prepared, spawns, activeLevel, pointerAcceptance)

	if app.movementSource != nil {
		app.movementSource.SetNavigation(activeWorld)
	}

	return nil
}

// prepareEntryWorld delegates coupled map, seam, zone, and spawn generation to one adapter call so
// every returned structure describes the same deterministic world generation.
func (app *application) prepareEntryWorld() (*entryworld.Prepared, error) {
	d2legacySource, err := app.modSource("d2legacy")
	if err != nil {
		return nil, wrap("resolve d2legacy package", err)
	}

	prepared, err := entryworld.Build(
		app.ctx,
		app.options.Content,
		d2legacySource,
		app.records,
		app.worldObjectResolver,
		1,
		0,
	)
	if err != nil {
		return nil, wrap("prepare d2legacy entry world", err)
	}

	return prepared, nil
}

// resolveEntryWorldSpawns derives development-fixture spawns from generated anchors. Production
// admission and seam screenshot positioning remain separate policies.
func (app *application) resolveEntryWorldSpawns(
	prepared *entryworld.Prepared,
) (map[int][2]float64, error) {
	townSpawn := prepared.Spawns[prepared.Seam.Town.LevelID]

	return entryWorldSpawns(
		app.options.FixtureWorldSpawn,
		prepared.Seam,
		townSpawn[0],
		townSpawn[1],
	)
}

// fixtureActiveLevel confines direct world selection to development fixtures and defaults ordinary
// startup to Rogue Encampment.
func (app *application) fixtureActiveLevel() int {
	if app.options.FixtureWorldLevel != 0 {
		return app.options.FixtureWorldLevel
	}

	return 1
}

// newFixturePointerAcceptance enables the renderer acceptance probe only when requested. Keeping it
// nil otherwise avoids production movement and capture state gaining test-only behavior.
func (app *application) newFixturePointerAcceptance(
	worldMap *gameworld.Map,
	spawn [2]float64,
) *pointerMovementAcceptance {
	if !app.options.FixturePointerMove {
		return nil
	}

	return newPointerMovementAcceptance(
		worldMap,
		spawn[0],
		spawn[1],
		app.profile.Width,
		app.profile.Height,
	)
}

// publishEntryWorld swaps seam, zones, maps, spawns, active level, and optional probe in one critical
// section. Readers can never pair geometry from one generation with metadata from another.
func (app *application) publishEntryWorld(
	prepared *entryworld.Prepared,
	spawns map[int][2]float64,
	activeLevel int,
	pointerAcceptance *pointerMovementAcceptance,
) {
	app.worldMu.Lock()
	app.transitionSeam = prepared.Seam
	app.gameWorldZones = prepared.Zones
	app.gameWorlds = prepared.Worlds
	app.gameWorldSpawns = spawns
	app.activeWorldLevel = activeLevel
	app.pointerAcceptance = pointerAcceptance
	app.worldMu.Unlock()
}

// entryWorldSpawns preserves the normal town entry anchor unless the explicit seam fixture is
// selected. Wilderness admission always uses its generated arrival side of the transition.
func entryWorldSpawns(
	fixtureSpawn string,
	seam gametransition.Seam,
	townX float64,
	townY float64,
) (map[int][2]float64, error) {
	switch fixtureSpawn {
	case "", "entry":
		return map[int][2]float64{
			1: {townX, townY},
			2: {seam.Wilderness.ArrivalX, seam.Wilderness.ArrivalY},
		}, nil
	case "seam":
		return map[int][2]float64{
			1: {seam.Town.ArrivalX, seam.Town.ArrivalY},
			2: {seam.Wilderness.ArrivalX, seam.Wilderness.ArrivalY},
		}, nil
	default:
		return nil, fmt.Errorf("development fixture world spawn %q is unavailable", fixtureSpawn)
	}
}

// currentWorld snapshots a complete active map and its source stamp for Lua. Missing preparation
// returns an empty value instead of exposing a partially initialized map.
func (app *application) currentWorld() modruntime.CurrentWorld {
	app.worldMu.RLock()
	defer app.worldMu.RUnlock()

	worldMap := app.gameWorlds[app.activeWorldLevel]

	zone := app.gameWorldZones[app.activeWorldLevel]
	if worldMap == nil || zone == nil || len(zone.Stamps()) == 0 {
		return modruntime.CurrentWorld{}
	}

	stamp := zone.Stamps()[0]

	return modruntime.CurrentWorld{
		Map:     worldMap,
		DS1:     stamp.DS1Path,
		DT1:     stamp.TilePaths,
		LevelID: app.activeWorldLevel,
	}
}

// activateWorld changes active level only when the map exists, then updates navigation outside the
// world lock. Same-level activation still refreshes navigation after connected recomposition.
func (app *application) activateWorld(levelID int) {
	app.worldMu.Lock()

	worldMap := app.gameWorlds[levelID]
	if worldMap != nil {
		app.activeWorldLevel = levelID
	}
	app.worldMu.Unlock()

	if worldMap == nil {
		return
	}

	if app.movementSource != nil {
		app.movementSource.SetNavigation(worldMap)
	}
}
