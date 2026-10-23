package entryworld

import (
	"math"
	"testing"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
)

func TestDungeonCandidatesRespectFootprintsAndPackSpacing(t *testing.T) {
	gameMap, err := gameworld.NewOpenMap(50, 50)
	if err != nil {
		t.Fatal(err)
	}
	reachable := reachablePopulationCells(gameMap, [2]float64{5, 5})
	points, count := mazePopulationPoints(gameMap, worldgen.Room{Width: 10, Height: 10}, reachable)
	if count != 100 || len(points) != 4 {
		t.Fatalf("eligible tiles/candidates=%d/%d", count, len(points))
	}
	selected := []gameworld.Point{}
	for _, value := range points {
		point := value.(map[string]any)
		position := gameworld.Point{X: point["x"].(float64), Y: point["y"].(float64)}
		if !reachable[position] || !gameMap.WalkableStep(position, position, point["clearance"].(float64)) {
			t.Fatal("unsafe candidate footprint")
		}
		for _, previous := range selected {
			if math.Hypot(position.X-previous.X, position.Y-previous.Y) < 6 {
				t.Fatal("overlapping pack candidates")
			}
		}
		selected = append(selected, position)
	}
	if len(reachablePopulationCells(gameMap, [2]float64{0, 0})) != 0 {
		t.Fatal("blocked admission footprint exposed candidates")
	}
	points, count = mazePopulationPoints(gameMap, worldgen.Room{Width: 10, Height: 10}, map[gameworld.Point]bool{})
	if len(points) != 0 || count != 0 {
		t.Fatal("unreachable room exposed candidates")
	}
}
