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
}
