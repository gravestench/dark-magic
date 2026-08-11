package clientapp

import (
	"context"
	"errors"

	"github.com/gravestench/dark-magic/internal/game/mapgen"
	gametransition "github.com/gravestench/dark-magic/internal/game/transition"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// buildEntryWorld generates and materializes both sides of the first playable
// zone seam. Maps publish together; a half-built wilderness is never active.
func (app *application) buildEntryWorld() error {
	snapshot, err := app.gameData.Snapshot()
	if err != nil {
		return wrap("load entry-world records", err)
	}
	townZone, err := mapgen.NewPresetGenerator(snapshot).Generate(mapgen.Request{
		Version: mapgen.ContractVersion, Seed: 1, Act: 1, LevelID: 1, Difficulty: mapgen.Normal,
	})
	if err != nil {
		return wrap("generate Act I town", err)
	}
	moorZone, err := mapgen.NewActOneOutdoorGenerator(snapshot).GenerateFromTown(mapgen.Request{Version: mapgen.ContractVersion, Seed: 1, Act: 1, LevelID: 2, Difficulty: mapgen.Normal}, townZone.Stamps()[0])
	if err != nil {
		return wrap("generate Blood Moor", err)
	}
	townMap, err := app.materializeZone(townZone)
	if err != nil {
		return wrap("materialize Act I town", err)
	}
	moorMap, err := app.materializeZone(moorZone)
	if err != nil {
		return wrap("materialize Blood Moor", err)
	}
	seam, err := gameworld.NewActOneTownMoorSeam(townZone, townMap, moorZone, moorMap)
	if err != nil {
		return wrap("join Act I town to Blood Moor", err)
	}
	app.transitionAuthority, err = gametransition.NewAuthority(seam)
	if err != nil {
		return wrap("create zone transition authority", err)
	}
	app.gameWorldZones = map[int]*mapgen.Zone{1: townZone, 2: moorZone}
	app.gameWorlds = map[int]*gameworld.Map{1: townMap, 2: moorMap}
	app.activeWorldLevel = 1
	app.transitionAuthority.SetObserver(app.activateWorld)
	return nil
}

func (app *application) materializeZone(zone *mapgen.Zone) (*gameworld.Map, error) {
	materializer, err := gameworld.NewMaterializer(app.options.Content, zone, app.worldObjectResolver)
	if err != nil {
		return nil, err
	}
	for {
		err = materializer.Step(context.Background())
		if errors.Is(err, gameworld.ErrMaterializationComplete) {
			break
		}
		if err != nil {
			return nil, err
		}
		if materializer.Progress().Completed == materializer.Progress().Total {
			break
		}
	}
	return materializer.Result()
}

func (app *application) currentWorld() modruntime.CurrentWorld {
	app.worldMu.RLock()
	defer app.worldMu.RUnlock()
	worldMap, zone := app.gameWorlds[app.activeWorldLevel], app.gameWorldZones[app.activeWorldLevel]
	if worldMap == nil || zone == nil || len(zone.Stamps()) == 0 {
		return modruntime.CurrentWorld{}
	}
	stamp := zone.Stamps()[0]
	return modruntime.CurrentWorld{Map: worldMap, DS1: stamp.DS1Path, DT1: stamp.TilePaths, LevelID: app.activeWorldLevel}
}

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
	if app.interactionAuthority != nil {
		app.interactionAuthority.ConfigureWorld(worldMap)
	}
}
