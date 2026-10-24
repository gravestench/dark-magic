package entryworld

import (
	"reflect"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
)

// Level 4 must not reuse player anchors outside a room, or project several
// anchors onto the same open cell.
func TestPopulationDataLevelFourRoomLocalPoints(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		name := "open"
		if blocked {
			name = "blocked"
		}
		t.Run(name, func(t *testing.T) {
			prepared := populationPreparedFixture(t)
			if blocked {
				// Block the whole room while leaving the surrounding map open.
				// Projecting an anchor outside the room must not rescue it.
				for y := 5; y < 10; y++ {
					for x := 5; x < 10; x++ {
						tile := gameworld.TilePlacement{X: x, Y: y, Layer: gameworld.LayerFloor}
						tile.Reference.Path = "synthetic"
						for i := range tile.Reference.SubTileFlags {
							tile.Reference.SubTileFlags[i].BlockWalk = true
						}
						prepared.Worlds[4].Tiles = append(prepared.Worlds[4].Tiles, tile)
					}
				}
				prepared.Worlds[4].RebuildFlags()
			}
			for _, nearby := range []int{0, 1} {
				first, err := prepared.PopulationData(nearby)
				if err != nil {
					t.Fatal(err)
				}
				second, err := prepared.PopulationData(nearby)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(first, second) {
					t.Fatal("population geometry changed between calls")
				}
				room := first["zones"].([]any)[2].(map[string]any)["rooms"].([]any)[0].(map[string]any)
				points := room["points"].([]any)
				if blocked && len(points) != 0 {
					t.Fatalf("blocked room points=%#v", points)
				}
				if !blocked && len(points) != 4 {
					t.Fatalf("open room points=%#v, want four", points)
				}
				assertPopulationRoomPoints(t, prepared.Worlds[4], room)
			}
		})
	}
}

// Candidate points remain distinct, room-local, and open so geometry cannot admit an unsafe center.
func assertPopulationRoomPoints(t *testing.T, gameMap *gameworld.Map, room map[string]any) {
	t.Helper()
	points := room["points"].([]any)
	if len(points) > 4 {
		t.Fatalf("room %v has %d points, want at most four", room["id"], len(points))
	}
	seen := map[[2]float64]bool{}
	for _, raw := range points {
		point := raw.(map[string]any)
		xy := [2]float64{point["x"].(float64), point["y"].(float64)}
		x, y := room["x"].(float64), room["y"].(float64)
		if seen[xy] || xy[0] < x || xy[0] >= x+room["width"].(float64) ||
			xy[1] < y || xy[1] >= y+room["height"].(float64) {
			t.Fatalf("room %v has duplicate or nonlocal point %v", room["id"], xy)
		}
		flags, ok := gameMap.FlagsAt(gameworld.CollisionCell(xy[0]), gameworld.CollisionCell(xy[1]))
		if !ok || flags.Blocked() {
			t.Fatalf("room %v has blocked point %v", room["id"], xy)
		}
		seen[xy] = true
	}
}
