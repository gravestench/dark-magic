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
type Seam struct {
	Town       SeamEndpoint
	Wilderness SeamEndpoint
}

// ResolveSeam applies a mod-authored seam specification to two materialized
// collision maps. It knows cardinal geometry, but no level IDs or role names.
func ResolveSeam(spec d2mapgen.SeamSpec, firstMap, secondMap *gameworld.Map) (Seam, error) {
	if firstMap == nil || secondMap == nil {
		return Seam{}, fmt.Errorf("world: seam requires both materialized maps")
	}

	firstX, firstY, err := resolveAuthoredEdge(spec, firstMap)
	if err != nil {
		return Seam{}, err
	}

	secondX, secondY, err := resolveGeneratedEdge(spec, secondMap)
	if err != nil {
		return Seam{}, err
	}

	// Validate arrivals only after both edge points, preserving the established error precedence for malformed seams.
	firstArrivalX, firstArrivalY, found := insetArrival(firstMap, firstX, firstY, spec.FirstDirection)
	if !found {
		return Seam{}, fmt.Errorf("world: first seam arrival is blocked")
	}

	secondArrivalX, secondArrivalY, found := insetArrival(secondMap, secondX, secondY, spec.SecondDirection)
	if !found {
		return Seam{}, fmt.Errorf("world: second seam arrival is blocked")
	}

	first := newSeamEndpoint(
		spec.FirstLevel,
		firstX,
		firstY,
		firstArrivalX,
		firstArrivalY,
		spec.FirstDirection,
		firstMap,
	)
	second := newSeamEndpoint(
		spec.SecondLevel,
		secondX,
		secondY,
		secondArrivalX,
		secondArrivalY,
		spec.SecondDirection,
		secondMap,
	)

	return Seam{Town: first, Wilderness: second}, nil
}

// resolveAuthoredEdge anchors the first zone to the extreme authored exit in the requested cardinal direction.
func resolveAuthoredEdge(spec d2mapgen.SeamSpec, world *gameworld.Map) (float64, float64, error) {
	anchor, found := cardinalAnchor(world.AuthoredExitAnchors(), spec.FirstDirection)
	if !found {
		return 0, 0, fmt.Errorf("world: first map has no authored seam anchor")
	}

	x, y, found := world.OpenPointNearExit(anchor)
	if !found {
		return 0, 0, fmt.Errorf("world: first seam edge is blocked")
	}

	return x, y, nil
}

// resolveGeneratedEdge converts the wilderness tile coordinate to its center subtile before collision lookup.
func resolveGeneratedEdge(spec d2mapgen.SeamSpec, world *gameworld.Map) (float64, float64, error) {
	x := float64(spec.SecondTileX*gameworld.SubtilesPerTile) + 2.5
	y := float64(spec.SecondTileY*gameworld.SubtilesPerTile) + 2.5

	x, y, found := world.OpenPointNearSubtile(x, y)
	if !found {
		return 0, 0, fmt.Errorf("world: second seam edge is blocked")
	}

	return x, y, nil
}

// newSeamEndpoint records map dimensions with collision-space coordinates so transition consumers need no map access.
func newSeamEndpoint(
	levelID int,
	x float64,
	y float64,
	arrivalX float64,
	arrivalY float64,
	direction string,
	world *gameworld.Map,
) SeamEndpoint {
	return SeamEndpoint{
		LevelID:   levelID,
		X:         x,
		Y:         y,
		ArrivalX:  arrivalX,
		ArrivalY:  arrivalY,
		Width:     float64(world.WidthSubtiles),
		Height:    float64(world.HeightSubtiles),
		Direction: direction,
	}
}

// insetArrival moves six subtiles into the destination so arrivals do not immediately retrigger the seam.
func insetArrival(world *gameworld.Map, x, y float64, edge string) (float64, float64, bool) {
	delta := map[string][2]float64{"north": {0, 6}, "east": {-6, 0}, "south": {0, -6}, "west": {6, 0}}[edge]
	return world.OpenPointNearSubtile(x+delta[0], y+delta[1])
}

// cardinalAnchor chooses the extreme authored exit in the requested direction while preserving authored tie order.
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
