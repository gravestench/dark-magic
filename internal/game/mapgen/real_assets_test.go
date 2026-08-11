package mapgen_test

import (
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	"github.com/gravestench/dark-magic/internal/game/data/store"
	"github.com/gravestench/dark-magic/internal/game/mapgen"
)

func TestActOnePresetRecipeAgainstOwnedAssets(t *testing.T) {
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
	zone, err := mapgen.NewPresetGenerator(snapshot).Generate(mapgen.Request{
		Version: mapgen.ContractVersion, Seed: 1, Act: 1, LevelID: 38,
	})
	if err != nil {
		t.Fatal(err)
	}
	stamp := zone.Stamps()[0]
	stampFile, err := source.Open(stamp.DS1Path)
	if err != nil {
		t.Fatalf("generated DS1 %q is unavailable: %v", stamp.DS1Path, err)
	}
	_ = stampFile.Close()
	for _, tile := range stamp.TilePaths {
		file, err := source.Open(tile)
		if err != nil {
			t.Fatalf("generated DT1 %q is unavailable: %v", tile, err)
		}
		_ = file.Close()
	}
	maze, err := mapgen.NewMazeGenerator(snapshot).Generate(mapgen.Request{
		Version: mapgen.ContractVersion, Seed: 42, Act: 1, LevelID: 9,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(maze.Rooms()) != 4 || len(maze.Links()) < 3 {
		t.Fatalf("real Cave Level 1 topology has %d rooms and %d links", len(maze.Rooms()), len(maze.Links()))
	}
	for _, chamber := range maze.Stamps() {
		file, err := source.Open(chamber.DS1Path)
		if err != nil {
			t.Fatalf("generated cave DS1 %q is unavailable: %v", chamber.DS1Path, err)
		}
		_ = file.Close()
	}
}
