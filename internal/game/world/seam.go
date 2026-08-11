package world

import (
	"fmt"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/mapgen"
)

// SeamEndpoint is one authoritative side of a level transition. Coordinates
// are collision-space subtiles, not renderer pixels.
type SeamEndpoint struct {
	LevelID            int
	X, Y               float64
	ArrivalX, ArrivalY float64
	Width, Height      float64
	Direction          string
}

// Seam is a verified bidirectional relationship between two materialized
// zones. Transition commands consume this value; presentation only observes it.
type Seam struct{ Town, Wilderness SeamEndpoint }

func NewActOneTownMoorSeam(townZone *mapgen.Zone, townMap *Map, moorZone *mapgen.Zone, moorMap *Map) (Seam, error) {
	if townZone == nil || townMap == nil || moorZone == nil || moorMap == nil {
		return Seam{}, fmt.Errorf("world: town/Blood Moor seam requires both zones and maps")
	}
	if townZone.Request().LevelID != 1 || moorZone.Request().LevelID != 2 {
		return Seam{}, fmt.Errorf("world: Act I seam requires Rogue Encampment and Blood Moor")
	}
	townStamps, warps := townZone.Stamps(), moorZone.Warps()
	if len(townStamps) != 1 || len(warps) != 1 || warps[0].DestinationLevel != 1 {
		return Seam{}, fmt.Errorf("world: Act I seam recipes are incomplete")
	}
	townDirection := strings.TrimPrefix(townStamps[0].Role, "act1-town:exit-")
	if oppositeCardinal(townDirection) != warps[0].Direction {
		return Seam{}, fmt.Errorf("world: town exit %q does not meet wilderness edge %q", townDirection, warps[0].Direction)
	}
	anchor, found := cardinalTownAnchor(townMap.AuthoredExitAnchors(), townDirection)
	if !found {
		return Seam{}, fmt.Errorf("world: town has no authored exit anchor")
	}
	townX, townY, found := townMap.OpenPointNearExit(anchor)
	if !found {
		return Seam{}, fmt.Errorf("world: town exit is blocked")
	}
	moorX, moorY, found := moorMap.OpenPointNearSubtile(float64(warps[0].X*SubtilesPerTile)+2.5, float64(warps[0].Y*SubtilesPerTile)+2.5)
	if !found {
		return Seam{}, fmt.Errorf("world: Blood Moor town edge is blocked")
	}
	townArrivalX, townArrivalY, found := insetArrival(townMap, townX, townY, townDirection)
	if !found {
		return Seam{}, fmt.Errorf("world: town arrival is blocked")
	}
	moorArrivalX, moorArrivalY, found := insetArrival(moorMap, moorX, moorY, warps[0].Direction)
	if !found {
		return Seam{}, fmt.Errorf("world: Blood Moor arrival is blocked")
	}
	return Seam{
		Town:       SeamEndpoint{LevelID: 1, X: townX, Y: townY, ArrivalX: townArrivalX, ArrivalY: townArrivalY, Width: float64(townMap.WidthSubtiles), Height: float64(townMap.HeightSubtiles), Direction: townDirection},
		Wilderness: SeamEndpoint{LevelID: 2, X: moorX, Y: moorY, ArrivalX: moorArrivalX, ArrivalY: moorArrivalY, Width: float64(moorMap.WidthSubtiles), Height: float64(moorMap.HeightSubtiles), Direction: warps[0].Direction},
	}, nil
}

func insetArrival(world *Map, x, y float64, edge string) (float64, float64, bool) {
	delta := map[string][2]float64{"north": {0, 6}, "east": {-6, 0}, "south": {0, -6}, "west": {6, 0}}[edge]
	return world.OpenPointNearSubtile(x+delta[0], y+delta[1])
}

func cardinalTownAnchor(anchors []ExitAnchor, direction string) (ExitAnchor, bool) {
	if len(anchors) == 0 {
		return ExitAnchor{}, false
	}
	best := anchors[0]
	for _, candidate := range anchors[1:] {
		switch direction {
		case "north":
			if candidate.Y < best.Y {
				best = candidate
			}
		case "east":
			if candidate.X > best.X {
				best = candidate
			}
		case "south":
			if candidate.Y > best.Y {
				best = candidate
			}
		case "west":
			if candidate.X < best.X {
				best = candidate
			}
		default:
			return ExitAnchor{}, false
		}
	}
	return best, true
}

func oppositeCardinal(direction string) string {
	return map[string]string{"north": "south", "east": "west", "south": "north", "west": "east"}[direction]
}
