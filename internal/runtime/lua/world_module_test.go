package modruntime

import (
	"context"
	"testing"
	"testing/fstest"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// TestWorldModuleReportsAssetErrorsToLua protects the world module reports asset errors to lua contract, including
// its observable ordering and failure behavior.
func TestWorldModuleReportsAssetErrorsToLua(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(WorldModule(fstest.MapFS{})); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local world = require("engine.world/v1")
local decoded, err = world.load("missing.ds1", {"missing.dt1"})
assert(decoded == nil)
assert(string.find(err, "missing.ds1"))
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

// TestSessionWorldModuleReturnsCurrentMapAndCopiedRecipe protects the session world module returns current map and
// copied recipe contract, including its observable ordering and failure behavior.
func TestSessionWorldModuleReturnsCurrentMapAndCopiedRecipe(t *testing.T) {
	runtime := New()
	current := CurrentWorld{
		Map:     &gameworld.Map{WidthTiles: 56, HeightTiles: 40},
		DS1:     "town.ds1",
		DT1:     []string{"floor.dt1"},
		LevelID: 7,
	}

	if err := runtime.RegisterModule(
		SessionWorldModule(fstest.MapFS{}, func() CurrentWorld { return current }),
	); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local world = require("engine.world/v1")
local map, recipe = world.current()
assert(map:dimensions().width_tiles == 56)
assert(map:line_clear(1, 1, 2, 2))
assert(map:barrier_clear(1, 1, 2, 2))
assert(recipe.ds1 == "town.ds1" and recipe.dt1[1] == "floor.dt1")
assert(recipe.level_id == 7 and world.current_level() == 7)
recipe.dt1[1] = "changed"
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}

	if current.DT1[0] != "floor.dt1" {
		t.Fatal("Lua mutated the authoritative recipe")
	}
}

// TestWorldModuleRejectsNonStringTilePaths protects the world module rejects non string tile paths contract,
// including its observable ordering and failure behavior.
func TestWorldModuleRejectsNonStringTilePaths(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(WorldModule(fstest.MapFS{})); err != nil {
		t.Fatal(err)
	}

	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()

	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local world = require("engine.world/v1")
local ok, err = pcall(world.load, "missing.ds1", {42})
assert(not ok and string.find(err, "sequence of strings"))
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}

func TestWorldPresentationCopiesSortedPreparedMaps(t *testing.T) {
	value := PresentationWorld{ActiveLevel: 2, Maps: []CurrentWorld{
		{LevelID: 3, Map: &gameworld.Map{WidthTiles: 30}, DS1: "plains.ds1", DT1: []string{"plains.dt1"}},
		{LevelID: 1, Map: &gameworld.Map{WidthTiles: 10}, DS1: "town.ds1", DT1: []string{"town.dt1"}},
	}}
	runtime := New()
	if err := runtime.RegisterModule(SessionWorldModuleWithPresentation(fstest.MapFS{}, nil, func() PresentationWorld { return value })); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()
	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local world = require("engine.world/v1")
local snapshot = world.presentation()
assert(snapshot.active_level == 2 and #snapshot.maps == 2)
assert(snapshot.maps[1].level_id == 1 and snapshot.maps[2].level_id == 3)
assert(snapshot.maps[1].map:dimensions().width_tiles == 10)
assert(snapshot.maps[2].map:dimensions().width_tiles == 30)
snapshot.maps[1].dt1[1] = "changed"
snapshot.maps[1].ds1 = "changed"
snapshot.maps[1] = nil
local again = world.presentation()
assert(again.maps[1].ds1 == "town.ds1" and again.maps[1].dt1[1] == "town.dt1")
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
	if value.Maps[0].LevelID != 3 || value.Maps[1].DT1[0] != "town.dt1" {
		t.Fatal("presentation mutated provider")
	}
}

func TestWorldPresentationWithoutSessionIsEmpty(t *testing.T) {
	runtime := New()
	if err := runtime.RegisterModule(WorldModule(fstest.MapFS{})); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = runtime.Stop(context.Background()) }()
	script := fstest.MapFS{"test.lua": &fstest.MapFile{Data: []byte(`
local world = require("engine.world/v1")
local snapshot = world.presentation()
assert(snapshot.active_level == 0 and #snapshot.maps == 0)
assert(world.current() == nil and world.current_level() == nil)
`)}}
	if err := runtime.Execute(context.Background(), script, "test.lua"); err != nil {
		t.Fatal(err)
	}
}
