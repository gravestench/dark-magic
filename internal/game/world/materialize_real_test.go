package world_test

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/game/mapgen"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

func TestGeneratedActOneCaveMaterializesFromOwnedAssets(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to a Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	source, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := gamedata.New(recordstore.New(source)).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	zone, err := mapgen.NewMazeGenerator(snapshot).Generate(mapgen.Request{Version: mapgen.ContractVersion, Seed: 42, Act: 1, LevelID: 9})
	if err != nil {
		t.Fatal(err)
	}
	materializer, err := gameworld.NewMaterializer(source, zone)
	if err != nil {
		t.Fatal(err)
	}
	for materializer.Progress().Completed < materializer.Progress().Total {
		if err := materializer.Step(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	worldMap, err := materializer.Result()
	if err != nil {
		t.Fatal(err)
	}
	if worldMap.WidthTiles != zone.Bounds().Width || worldMap.HeightTiles != zone.Bounds().Height || len(worldMap.Tiles) == 0 {
		t.Fatalf("materialized map dimensions/tiles = %dx%d/%d", worldMap.WidthTiles, worldMap.HeightTiles, len(worldMap.Tiles))
	}
}

func TestGeneratedActOneTownMaterializesWithCampfireEntry(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to a Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	source, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := gamedata.New(recordstore.New(source)).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	zone, err := mapgen.NewPresetGenerator(snapshot).Generate(mapgen.Request{Version: mapgen.ContractVersion, Seed: 1, Act: 1, LevelID: 1})
	if err != nil {
		t.Fatal(err)
	}
	materializer, err := gameworld.NewMaterializer(source, zone)
	if err != nil {
		t.Fatal(err)
	}
	for materializer.Progress().Completed < materializer.Progress().Total {
		if err := materializer.Step(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	worldMap, err := materializer.Result()
	if err != nil {
		t.Fatal(err)
	}
	x, y, found := worldMap.ActOneTownEntry()
	if !found {
		t.Fatal("materialized town has no campfire-relative entry")
	}
	flags, inside := worldMap.FlagsAt(int(x), int(y))
	if !inside || flags.Blocked() {
		t.Fatalf("town entry (%v,%v) is not open", x, y)
	}
	if anchors := worldMap.AuthoredExitAnchors(); len(anchors) == 0 {
		t.Fatal("materialized town has no authored orientation-10/11 exit anchor")
	}
}

func TestGeneratedBloodMoorMaterializesFromTownExit(t *testing.T) {
	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to a Diablo II MPQ directory")
	}
	t.Setenv("MPQ_DIRECTORY", directory)
	source, err := content.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := gamedata.New(recordstore.New(source)).Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	town, err := mapgen.NewPresetGenerator(snapshot).Generate(mapgen.Request{Version: mapgen.ContractVersion, Seed: 17, Act: 1, LevelID: 1})
	if err != nil {
		t.Fatal(err)
	}
	moor, err := mapgen.NewActOneOutdoorGenerator(snapshot).GenerateFromTown(mapgen.Request{Version: mapgen.ContractVersion, Seed: 17, Act: 1, LevelID: 2}, town.Stamps()[0])
	if err != nil {
		t.Fatal(err)
	}
	materializer, err := gameworld.NewMaterializer(source, moor)
	if err != nil {
		t.Fatal(err)
	}
	for materializer.Progress().Completed < materializer.Progress().Total {
		if err := materializer.Step(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	worldMap, err := materializer.Result()
	if err != nil {
		t.Fatal(err)
	}
	if worldMap.WidthTiles != 80 || worldMap.HeightTiles != 80 || len(worldMap.Tiles) == 0 {
		t.Fatalf("Blood Moor = %dx%d with %d tiles", worldMap.WidthTiles, worldMap.HeightTiles, len(worldMap.Tiles))
	}
	warp := moor.Warps()[0]
	if _, inside := worldMap.FlagsAt(warp.X*gameworld.SubtilesPerTile, warp.Y*gameworld.SubtilesPerTile); !inside {
		t.Fatalf("town warp lies outside Blood Moor: %#v", warp)
	}
	townMaterializer, err := gameworld.NewMaterializer(source, town)
	if err != nil {
		t.Fatal(err)
	}
	for townMaterializer.Progress().Completed < townMaterializer.Progress().Total {
		if err := townMaterializer.Step(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	townMap, err := townMaterializer.Result()
	if err != nil {
		t.Fatal(err)
	}
	seam, err := gameworld.NewActOneTownMoorSeam(town, townMap, moor, worldMap)
	if err != nil {
		t.Fatal(err)
	}
	if seam.Town.LevelID != 1 || seam.Wilderness.LevelID != 2 {
		t.Fatalf("production seam = %#v", seam)
	}
}
