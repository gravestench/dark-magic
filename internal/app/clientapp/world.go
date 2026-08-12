package clientapp

import (
	"context"
	"errors"
	"fmt"

	"github.com/gravestench/akara"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// buildEntryWorld generates and materializes both sides of the first playable
// zone seam. Maps publish together; a half-built wilderness is never active.
func (app *application) buildEntryWorld() error {
	townZone, moorZone, err := d2mapgen.GenerateEntryZones(app.ctx, app.options.Content, app.records, 1)
	if err != nil {
		return wrap("generate d2legacy entry world", err)
	}
	townMap, err := app.materializeZone(townZone)
	if err != nil {
		return wrap("materialize Act I town", err)
	}
	moorMap, err := app.materializeZone(moorZone)
	if err != nil {
		return wrap("materialize Blood Moor", err)
	}
	seam, err := gametransition.NewActOneTownMoorSeam(townZone, townMap, moorZone, moorMap)
	if err != nil {
		return wrap("join Act I town to Blood Moor", err)
	}
	app.transitionSeam = seam
	app.gameWorldZones = map[int]*worldgen.Zone{1: townZone, 2: moorZone}
	app.gameWorlds = map[int]*gameworld.Map{1: townMap, 2: moorMap}
	townSpawnX, townSpawnY, found := d2mapgen.ActOneTownEntry(townMap)
	if !found {
		return errors.New("Act I town has no campfire entry")
	}
	app.gameWorldSpawns, err = entryWorldSpawns(app.options.FixtureWorldSpawn, seam, townSpawnX, townSpawnY)
	if err != nil {
		return err
	}
	app.activeWorldLevel = app.options.FixtureWorldLevel
	if app.activeWorldLevel == 0 {
		app.activeWorldLevel = 1
	}
	if app.gameWorlds[app.activeWorldLevel] == nil {
		return fmt.Errorf("development fixture world level %d is unavailable", app.activeWorldLevel)
	}
	if app.options.FixturePointerMove {
		spawn := app.gameWorldSpawns[app.activeWorldLevel]
		app.pointerAcceptance = newPointerMovementAcceptance(
			app.gameWorlds[app.activeWorldLevel], spawn[0], spawn[1], app.profile.Width, app.profile.Height,
		)
	}
	return nil
}

// syncActiveWorldFromPlayer is a presentation adapter. Lua has already
// committed the authoritative level change; this only swaps client-side map
// caches and navigation inputs to match that fact.
func (app *application) syncActiveWorldFromPlayer() {
	controls, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.player_control")
	if !ok {
		return
	}
	locations, ok := akara.GetDynamicStore(app.entitySimulation.World(), "d2legacy.world.location")
	if !ok {
		return
	}
	for _, entity := range controls.Entities() {
		control, _ := controls.Get(entity)
		owner, _ := control.Get("player")
		if owner != "local-player" {
			continue
		}
		location, found := locations.Get(entity)
		if !found {
			return
		}
		level, _ := location.Get("level_id")
		if levelID := int(level.(int64)); levelID != app.activeWorldLevel {
			app.activateWorld(levelID)
		}
		return
	}
}

// entryWorldSpawns keeps the real admission rule and the screenshot fixture
// choice visibly separate. Players normally enter town at the campfire. A
// development capture may instead stand just inside either side of the seam.
func entryWorldSpawns(fixtureSpawn string, seam gametransition.Seam, townX, townY float64) (map[int][2]float64, error) {
	switch fixtureSpawn {
	case "", "entry":
		return map[int][2]float64{1: {townX, townY}, 2: {seam.Wilderness.ArrivalX, seam.Wilderness.ArrivalY}}, nil
	case "seam":
		return map[int][2]float64{1: {seam.Town.ArrivalX, seam.Town.ArrivalY}, 2: {seam.Wilderness.ArrivalX, seam.Wilderness.ArrivalY}}, nil
	default:
		return nil, fmt.Errorf("development fixture world spawn %q is unavailable", fixtureSpawn)
	}
}

// transitionBootstrapData exports collision-derived seam geometry without
// deciding what it means. The d2legacy mod owns level identities, trigger
// distance, arrival behavior, and the authoritative transition system.
func (app *application) transitionBootstrapData() map[string]any {
	endpoint := func(source, destination gametransition.SeamEndpoint) map[string]any {
		return map[string]any{
			"source_level": float64(source.LevelID), "destination_level": float64(destination.LevelID),
			"source_x": source.X, "source_y": source.Y,
			"arrival_x": destination.ArrivalX, "arrival_y": destination.ArrivalY,
			"world_width": destination.Width, "world_height": destination.Height,
		}
	}
	return map[string]any{"seams": []any{
		endpoint(app.transitionSeam.Town, app.transitionSeam.Wilderness),
		endpoint(app.transitionSeam.Wilderness, app.transitionSeam.Town),
	}}
}

func (app *application) materializeZone(zone *worldgen.Zone) (*gameworld.Map, error) {
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
}
