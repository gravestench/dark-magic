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
	transitions, err := worldMap.ResolveLevelTransitions(snapshot.LevelsByID[9], snapshot.LevelWarpsByID)
	if err != nil {
		t.Fatal(err)
	}
	if len(transitions) == 0 {
		t.Fatalf("materialized cave transitions = %#v", transitions)
	}
	first := transitions[0]
	if first.DestinationLevel != 3 || first.WarpID != 4 || first.Tile.X != 4 || first.Tile.Y != 29 || first.Tile.MainIndex != 0 || first.Tile.SubIndex != 21 {
		t.Fatalf("unexpected production cave transition = %#v", first)
	}
	wantGeometry := gameworld.WarpGeometry{
		CellOrigin: gameworld.SubtilePoint{X: 20, Y: 145}, EntityPosition: gameworld.SubtilePoint{X: 22, Y: 150},
		SelectionLocal: gameworld.LocalSelectionBounds{MinX: -30, MinY: -120, MaxX: 90, MaxY: 30},
		Arrival:        gameworld.SubtilePoint{X: 22, Y: 150}, ExitWalkTarget: gameworld.SubtilePoint{X: 25, Y: 155},
	}
	if got := first.Geometry(); got != wantGeometry {
		t.Fatalf("production cave geometry = %#v, want %#v", got, wantGeometry)
	}
	for _, transition := range transitions {
		geometry := transition.Geometry()
		if geometry.EntityPosition != geometry.Arrival || geometry.SelectionLocal.MaxX <= geometry.SelectionLocal.MinX || geometry.SelectionLocal.MaxY <= geometry.SelectionLocal.MinY {
			t.Fatalf("invalid authored warp geometry = %#v", geometry)
		}
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
	foundOpenBridge, foundBlockedRiver, foundBlockedCliff := false, false, false
	openBridgeTiles := make([]mapgen.PathTile, 0)
	for _, structure := range moor.Structures() {
		flags, inside := worldMap.FlagsAt(structure.X*gameworld.SubtilesPerTile+2, structure.Y*gameworld.SubtilesPerTile+2)
		if !inside {
			t.Fatalf("structure lies outside materialized map: %#v", structure)
		}
		if structure.Kind == "bridge" && !flags.Blocked() {
			tile := mapgen.PathTile{X: structure.X, Y: structure.Y}
			openBridgeTiles = append(openBridgeTiles, tile)
			foundOpenBridge = true
		}
		if structure.Kind == "river" && flags.Blocked() {
			foundBlockedRiver = true
		}
		if structure.Kind == "cliff" && flags.Blocked() {
			foundBlockedCliff = true
		}
	}
	if !foundOpenBridge || !foundBlockedRiver || !foundBlockedCliff {
		t.Fatalf("asset-backed structures: open bridge=%v blocked river=%v blocked cliff=%v; open bridge tile centers=%v", foundOpenBridge, foundBlockedRiver, foundBlockedCliff, openBridgeTiles)
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
