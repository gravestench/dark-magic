package world_test

import (
	"io/fs"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	mapgen "github.com/gravestench/dark-magic/internal/game/worldgen"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// TestGeneratedActOneCaveMaterializesFromOwnedAssets verifies production Lua generation and archive-backed loading meet
// at the renderer-neutral world contract.
func TestGeneratedActOneCaveMaterializesFromOwnedAssets(t *testing.T) {
	source := openRealWorldContent(t)
	records := recordstore.New(source)
	runtime := openMapgenRuntime(t, source, records)

	zone, err := runtime.Generate(t.Context(), "maze", float64(9), float64(42), float64(0))
	if err != nil {
		t.Fatal(err)
	}

	worldMap := materializeRealZone(t, source, zone)

	wrongDimensions := worldMap.WidthTiles != zone.Bounds().Width || worldMap.HeightTiles != zone.Bounds().Height
	if wrongDimensions || len(worldMap.Tiles) == 0 {
		t.Fatalf(
			"materialized map dimensions/tiles = %dx%d/%d",
			worldMap.WidthTiles,
			worldMap.HeightTiles,
			len(worldMap.Tiles),
		)
	}
}

// TestGeneratedActOneTownMaterializesWithCampfireEntry verifies the production town policy resolves a traversable entry
// and retains authored exits after full map assembly.
func TestGeneratedActOneTownMaterializesWithCampfireEntry(t *testing.T) {
	source := openRealWorldContent(t)
	records := recordstore.New(source)
	runtime := openMapgenRuntime(t, source, records)

	zone, err := runtime.Generate(t.Context(), "preset", float64(1), float64(1), float64(0))
	if err != nil {
		t.Fatal(err)
	}

	worldMap := materializeRealZone(t, source, zone)

	x, y, found := runtime.TownEntry(t.Context(), worldMap)
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

// TestGeneratedBloodMoorMaterializesFromTownExit verifies production outdoor structures, warp bounds, and the final
// town/wilderness seam against archive-backed assets.
func TestGeneratedBloodMoorMaterializesFromTownExit(t *testing.T) {
	source := openRealWorldContent(t)
	records := recordstore.New(source)

	entryWorld, err := d2mapgen.GenerateEntryWorld(t.Context(), source, records, 17, 0)
	if err != nil {
		t.Fatal(err)
	}

	town, moor := entryWorld.Town, entryWorld.Wilderness

	worldMap := materializeRealZone(t, source, moor)

	if worldMap.WidthTiles != 80 || worldMap.HeightTiles != 80 || len(worldMap.Tiles) == 0 {
		t.Fatalf("Blood Moor = %dx%d with %d tiles", worldMap.WidthTiles, worldMap.HeightTiles, len(worldMap.Tiles))
	}

	foundOpenBridge, foundBlockedRiver, foundBlockedCliff := false, false, false
	openBridgeTiles := make([]mapgen.PathTile, 0)

	for _, structure := range moor.Structures() {
		flags, inside := worldMap.FlagsAt(
			structure.X*gameworld.SubtilesPerTile+2,
			structure.Y*gameworld.SubtilesPerTile+2,
		)
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
		t.Fatalf(
			"asset-backed structures: open bridge=%v blocked river=%v blocked cliff=%v; centers=%v",
			foundOpenBridge,
			foundBlockedRiver,
			foundBlockedCliff,
			openBridgeTiles,
		)
	}

	warp := moor.Warps()[0]
	if _, inside := worldMap.FlagsAt(warp.X*gameworld.SubtilesPerTile, warp.Y*gameworld.SubtilesPerTile); !inside {
		t.Fatalf("town warp lies outside Blood Moor: %#v", warp)
	}

	townMap := materializeRealZone(t, source, town)

	seam, err := gametransition.ResolveSeam(entryWorld.Seam, townMap, worldMap)
	if err != nil {
		t.Fatal(err)
	}

	if seam.Town.LevelID != 1 || seam.Wilderness.LevelID != 2 {
		t.Fatalf("production seam = %#v", seam)
	}
}

// openRealWorldContent opens the optional MPQ fixture and binds its archive handles to the test lifecycle. Centralized
// setup keeps every integration scenario on the same layered-content contract.
func openRealWorldContent(t *testing.T) *content.FS {
	t.Helper()

	directory := os.Getenv("DARK_MAGIC_TEST_MPQ_DIRECTORY")
	if directory == "" {
		t.Skip("set DARK_MAGIC_TEST_MPQ_DIRECTORY to a Diablo II MPQ directory")
	}

	t.Setenv("MPQ_DIRECTORY", directory)

	source, err := content.FromEnvironment(content.Layer{Name: "d2legacy", FS: content.D2Legacy()})
	if err != nil {
		t.Fatal(err)
	}
	// Cleanup reports close failures because leaked archive handles make later real-asset tests unreliable.
	t.Cleanup(func() {
		if err := source.Close(); err != nil {
			t.Errorf("close content: %v", err)
		}
	})

	return source
}

// openMapgenRuntime starts the production Lua policy host and guarantees it stops even after a materialization failure.
func openMapgenRuntime(
	t *testing.T,
	source fs.FS,
	records *recordstore.Store,
) *d2mapgen.Runtime {
	t.Helper()

	runtime, err := d2mapgen.NewRuntime(t.Context(), source, records)
	if err != nil {
		t.Fatal(err)
	}
	// Cleanup owns the Lua runtime independently from the content source registered by openRealWorldContent.
	t.Cleanup(func() {
		if err := runtime.Close(t.Context()); err != nil {
			t.Errorf("close mapgen runtime: %v", err)
		}
	})

	return runtime
}

// materializeRealZone drives incremental loading to completion and returns only the published map. Keeping the loop in
// one helper makes each integration test describe its gameplay contract instead of materializer mechanics.
func materializeRealZone(t *testing.T, source fs.FS, zone *mapgen.Zone) *gameworld.Map {
	t.Helper()

	materializer, err := gameworld.NewMaterializer(source, zone)
	if err != nil {
		t.Fatal(err)
	}

	for materializer.Progress().Completed < materializer.Progress().Total {
		if err := materializer.Step(t.Context()); err != nil {
			t.Fatal(err)
		}
	}

	result, err := materializer.Result()
	if err != nil {
		t.Fatal(err)
	}

	return result
}
