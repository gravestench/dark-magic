package world_test

import (
	"math"
	"testing"

	recordstore "github.com/gravestench/dark-magic/internal/game/data/store"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// TestGeneratedStonyFieldMaterializesAConnectedEntry checks the structural zone against owned 1.14d assets.
// The layout and explicit source direction are provisional Dark Magic policy.
// This does not assert patch-exact layout or an existing Cold Plains exit.
func TestGeneratedStonyFieldMaterializesAConnectedEntry(t *testing.T) {
	source := openRealWorldContent(t)
	records := recordstore.New(source)

	runtime := openMapgenRuntime(t, source, records)
	zone, err := runtime.Generate(t.Context(), "stony_field", float64(4), float64(17), "east", float64(0))
	if err != nil {
		t.Fatal(err)
	}
	worldMap := materializeRealZone(t, source, zone)
	if worldMap.WidthTiles != 80 || worldMap.HeightTiles != 80 || len(worldMap.Tiles) == 0 {
		t.Fatalf("Stony Field dimensions/tiles = %dx%d/%d", worldMap.WidthTiles, worldMap.HeightTiles, len(worldMap.Tiles))
	}
	warps := zone.Warps()
	if len(warps) != 1 || warps[0].Role != "cold-plains-entry" || warps[0].DestinationLevel != 3 || warps[0].Direction != "west" {
		t.Fatalf("Stony Field entry warps = %#v", warps)
	}
	warp := warps[0]
	warpX := float64(warp.X*gameworld.SubtilesPerTile) + 2.5
	warpY := float64(warp.Y*gameworld.SubtilesPerTile) + 2.5
	arrivalX, arrivalY, found := worldMap.OpenPointNearSubtile(warpX, warpY)
	if !found || math.Hypot(arrivalX-warpX, arrivalY-warpY) > 10 {
		t.Fatalf("Stony Field edge entry has no nearby open arrival: warp=(%.1f,%.1f), arrival=(%.1f,%.1f), found=%v",
			warpX, warpY, arrivalX, arrivalY, found)
	}
	flags, inside := worldMap.FlagsAt(int(arrivalX), int(arrivalY))
	if !inside || flags.Blocked() {
		t.Fatalf("Stony Field arrival (%.1f,%.1f) is blocked or outside", arrivalX, arrivalY)
	}

	centerX := float64(worldMap.WidthSubtiles) / 2
	centerY := float64(worldMap.HeightSubtiles) / 2
	goalX, goalY, found := worldMap.OpenPointNearCenter()
	if !found || math.Hypot(goalX-centerX, goalY-centerY) > 10 {
		t.Fatalf("Stony Field has no open center: center=(%.1f,%.1f), goal=(%.1f,%.1f), found=%v",
			centerX, centerY, goalX, goalY, found)
	}
	if math.Hypot(goalX-arrivalX, goalY-arrivalY) < 8*gameworld.SubtilesPerTile {
		t.Fatal("Stony Field goal is not at least one room into the interior")
	}
	path, err := worldMap.FindPath(gameworld.PathRequest{
		Start: gameworld.Point{X: arrivalX, Y: arrivalY},
		Goal:  gameworld.Point{X: goalX, Y: goalY},
	})
	if err != nil {
		t.Fatalf("Stony Field entry cannot reach its interior: %v", err)
	}
	if len(path) == 0 {
		t.Fatal("Stony Field collision path is empty")
	}
}
