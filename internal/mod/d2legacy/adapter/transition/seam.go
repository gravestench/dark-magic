package transition

import (
	"fmt"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
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

// ResolveSeam applies a mod-authored seam specification to two materialized
// collision maps. It knows cardinal geometry, but no level IDs or role names.
func ResolveSeam(spec d2mapgen.SeamSpec, firstMap, secondMap *gameworld.Map) (Seam, error) {
	if firstMap == nil || secondMap == nil {
		return Seam{}, fmt.Errorf("world: seam requires both materialized maps")
	}
	anchor, found := cardinalAnchor(firstMap.AuthoredExitAnchors(), spec.FirstDirection)
	if !found {
		return Seam{}, fmt.Errorf("world: first map has no authored seam anchor")
	}
	firstX, firstY, found := firstMap.OpenPointNearExit(anchor)
	if !found {
		return Seam{}, fmt.Errorf("world: first seam edge is blocked")
	}
	secondX, secondY, found := secondMap.OpenPointNearSubtile(float64(spec.SecondTileX*gameworld.SubtilesPerTile)+2.5, float64(spec.SecondTileY*gameworld.SubtilesPerTile)+2.5)
	if !found {
		return Seam{}, fmt.Errorf("world: second seam edge is blocked")
	}
	firstArrivalX, firstArrivalY, found := insetArrival(firstMap, firstX, firstY, spec.FirstDirection)
	if !found {
		return Seam{}, fmt.Errorf("world: first seam arrival is blocked")
	}
	secondArrivalX, secondArrivalY, found := insetArrival(secondMap, secondX, secondY, spec.SecondDirection)
	if !found {
		return Seam{}, fmt.Errorf("world: second seam arrival is blocked")
	}
	return Seam{
		Town:       SeamEndpoint{LevelID: spec.FirstLevel, X: firstX, Y: firstY, ArrivalX: firstArrivalX, ArrivalY: firstArrivalY, Width: float64(firstMap.WidthSubtiles), Height: float64(firstMap.HeightSubtiles), Direction: spec.FirstDirection},
		Wilderness: SeamEndpoint{LevelID: spec.SecondLevel, X: secondX, Y: secondY, ArrivalX: secondArrivalX, ArrivalY: secondArrivalY, Width: float64(secondMap.WidthSubtiles), Height: float64(secondMap.HeightSubtiles), Direction: spec.SecondDirection},
	}, nil
}

func insetArrival(world *gameworld.Map, x, y float64, edge string) (float64, float64, bool) {
	delta := map[string][2]float64{"north": {0, 6}, "east": {-6, 0}, "south": {0, -6}, "west": {6, 0}}[edge]
	return world.OpenPointNearSubtile(x+delta[0], y+delta[1])
}

func cardinalAnchor(anchors []gameworld.ExitAnchor, direction string) (gameworld.ExitAnchor, bool) {
	if len(anchors) == 0 {
		return gameworld.ExitAnchor{}, false
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
			return gameworld.ExitAnchor{}, false
		}
	}
	return best, true
}
