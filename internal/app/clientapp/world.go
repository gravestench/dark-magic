package clientapp

import (
	"context"
	"errors"

	"github.com/gravestench/dark-magic/internal/game/mapgen"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	modruntime "github.com/gravestench/dark-magic/internal/runtime/lua"
)

// buildEntryWorld turns the session's deterministic town recipe into immutable
// gameplay facts. Even though Act I town is one stamp today, stepping the
// materializer keeps this composition root ready for multi-room generated maps.
func (app *application) buildEntryWorld() error {
	snapshot, err := app.gameData.Snapshot()
	if err != nil {
		return wrap("load entry-world records", err)
	}
	zone, err := mapgen.NewPresetGenerator(snapshot).Generate(mapgen.Request{
		Version: mapgen.ContractVersion, Seed: 1, Act: 1, LevelID: 1, Difficulty: mapgen.Normal,
	})
	if err != nil {
		return wrap("generate Act I town", err)
	}
	materializer, err := gameworld.NewMaterializer(app.options.Content, zone, app.worldObjectResolver)
	if err != nil {
		return wrap("create Act I town materializer", err)
	}
	for {
		err = materializer.Step(context.Background())
		if errors.Is(err, gameworld.ErrMaterializationComplete) {
			break
		}
		if err != nil {
			return wrap("materialize Act I town", err)
		}
		if materializer.Progress().Completed == materializer.Progress().Total {
			break
		}
	}
	worldMap, err := materializer.Result()
	if err != nil {
		return wrap("publish Act I town", err)
	}
	app.gameWorldZone, app.gameWorld = zone, worldMap
	return nil
}

func (app *application) currentWorld() modruntime.CurrentWorld {
	if app.gameWorld == nil || app.gameWorldZone == nil || len(app.gameWorldZone.Stamps()) == 0 {
		return modruntime.CurrentWorld{}
	}
	stamp := app.gameWorldZone.Stamps()[0]
	return modruntime.CurrentWorld{Map: app.gameWorld, DS1: stamp.DS1Path, DT1: stamp.TilePaths}
}
