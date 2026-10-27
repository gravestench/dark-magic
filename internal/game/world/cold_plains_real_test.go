package world_test

import (
	"math"
	"testing"

	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
)

// TestGeneratedColdPlainsMaterializesAConnectedEntry checks the structural zone against owned 1.14d assets.
// The fill and entry placement are Dark Magic policy; this does not assert patch-exact Cold Plains layout.
func TestGeneratedColdPlainsMaterializesAConnectedEntry(t *testing.T) {
	source := openRealWorldContent(t)
	records := recordstore.New(source)

	entryWorld, err := d2mapgen.GenerateEntryWorld(t.Context(), source, records, 17, 0)
	if err != nil {
		t.Fatal(err)
	}
	var sourceDirection string
	for _, warp := range entryWorld.Zones[2].Warps() {
		if warp.Role != "next-level-exit" || warp.DestinationLevel != 3 {
			continue
		}
		if sourceDirection != "" {
			t.Fatal("Blood Moor has more than one Cold Plains exit")
		}
		sourceDirection = warp.Direction
	}
	if sourceDirection == "" {
		t.Fatal("admitted Blood Moor has no Cold Plains exit")
	}

	runtime := openMapgenRuntime(t, source, records)
	zone, err := runtime.Generate(t.Context(), "cold_plains", float64(3), float64(17), sourceDirection, float64(0))
	if err != nil {
		t.Fatal(err)
	}
	worldMap := materializeRealZone(t, source, zone)
	if worldMap.WidthTiles != 80 || worldMap.HeightTiles != 80 || len(worldMap.Tiles) == 0 {
		t.Fatalf("Cold Plains dimensions/tiles = %dx%d/%d", worldMap.WidthTiles, worldMap.HeightTiles, len(worldMap.Tiles))
	}
	warps := zone.Warps()
	if len(warps) != 1 || warps[0].Role != "blood-moor-entry" || warps[0].DestinationLevel != 2 {
		t.Fatalf("Cold Plains entry warps = %#v", warps)
	}
	warp := warps[0]
	warpX := float64(warp.X*gameworld.SubtilesPerTile) + 2.5
	warpY := float64(warp.Y*gameworld.SubtilesPerTile) + 2.5
	arrivalX, arrivalY, found := worldMap.OpenPointNearSubtile(warpX, warpY)
	if !found || math.Hypot(arrivalX-warpX, arrivalY-warpY) > 10 {
		t.Fatalf("Cold Plains edge entry has no nearby open arrival: warp=(%.1f,%.1f), arrival=(%.1f,%.1f), found=%v",
			warpX, warpY, arrivalX, arrivalY, found)
	}
	flags, inside := worldMap.FlagsAt(int(arrivalX), int(arrivalY))
	if !inside || flags.Blocked() {
		t.Fatalf("Cold Plains arrival (%.1f,%.1f) is blocked or outside", arrivalX, arrivalY)
	}

	centerX := float64(worldMap.WidthSubtiles) / 2
	centerY := float64(worldMap.HeightSubtiles) / 2
	goalX, goalY, found := worldMap.OpenPointNearCenter()
	if !found || math.Hypot(goalX-centerX, goalY-centerY) > 10 {
		t.Fatalf("Cold Plains has no open center: center=(%.1f,%.1f), goal=(%.1f,%.1f), found=%v",
			centerX, centerY, goalX, goalY, found)
	}
	if math.Hypot(goalX-arrivalX, goalY-arrivalY) < 8*gameworld.SubtilesPerTile {
		t.Fatal("Cold Plains goal is not at least one room into the interior")
	}
	path, err := worldMap.FindPath(gameworld.PathRequest{
		Start: gameworld.Point{X: arrivalX, Y: arrivalY},
		Goal:  gameworld.Point{X: goalX, Y: goalY},
	})
	if err != nil {
		t.Fatalf("Cold Plains entry cannot reach its interior: %v", err)
	}
	if len(path) == 0 {
		t.Fatal("Cold Plains collision path is empty")
	}
}
