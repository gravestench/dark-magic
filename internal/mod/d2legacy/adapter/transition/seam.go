package transition

import (
	"fmt"
	"math"

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
	First         SeamEndpoint
	Second        SeamEndpoint
	TriggerRadius float64
}

// ResolveSeam applies a mod-authored seam specification to two materialized
// collision maps. It knows cardinal geometry, but no level IDs or role names.
func ResolveSeam(spec d2mapgen.SeamSpec, firstMap, secondMap *gameworld.Map) (Seam, error) {
	if firstMap == nil || secondMap == nil {
		return Seam{}, fmt.Errorf("world: seam requires both materialized maps")
	}

	var firstX, firstY float64
	var err error
	if spec.FirstGenerated {
		firstX, firstY, err = resolveGeneratedTile(firstMap, spec.FirstTileX, spec.FirstTileY)
	} else {
		firstX, firstY, err = resolveAuthoredEdge(spec, firstMap)
	}
	if err != nil {
		return Seam{}, err
	}

	secondX, secondY, err := resolveGeneratedTile(secondMap, spec.SecondTileX, spec.SecondTileY)
	if err != nil {
		return Seam{}, err
	}

	// Validate arrivals only after both edge points, preserving the established error precedence for malformed seams.
	inset := spec.ArrivalInset
	if inset == 0 {
		inset = 6
	}
	firstArrivalX, firstArrivalY, found := insetArrival(firstMap, firstX, firstY, spec.FirstDirection, inset)
	if !found {
		return Seam{}, fmt.Errorf("world: first seam arrival is blocked")
	}

	secondArrivalX, secondArrivalY, found := insetArrival(secondMap, secondX, secondY, spec.SecondDirection, inset)
	if !found {
		return Seam{}, fmt.Errorf("world: second seam arrival is blocked")
	}
	if spec.FirstGenerated {
		for _, side := range []struct {
			world              *gameworld.Map
			tileX, tileY       int
			edgeX, edgeY       float64
			arrivalX, arrivalY float64
		}{
			{firstMap, spec.FirstTileX, spec.FirstTileY, firstX, firstY, firstArrivalX, firstArrivalY},
			{secondMap, spec.SecondTileX, spec.SecondTileY, secondX, secondY, secondArrivalX, secondArrivalY},
		} {
			if err := generatedArrivalSafe(spec, side.world, side.tileX, side.tileY,
				side.edgeX, side.edgeY, side.arrivalX, side.arrivalY); err != nil {
				return Seam{}, err
			}
		}
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

	radius := spec.TriggerRadius
	if radius == 0 {
		radius = 2
	}
	return Seam{First: first, Second: second, TriggerRadius: float64(radius)}, nil
}

// Generated edges have no authored DS1 anchor. Bound collision search and require a walkable route to both the
// interior and return trigger before advertising a reciprocal transition.
func generatedArrivalSafe(
	spec d2mapgen.SeamSpec, world *gameworld.Map, tileX, tileY int,
	edgeX, edgeY, arrivalX, arrivalY float64,
) error {
	anchorX := float64(tileX*gameworld.SubtilesPerTile) + 2.5
	anchorY := float64(tileY*gameworld.SubtilesPerTile) + 2.5
	limit := float64(spec.MaxArrivalDistance)
	if math.Hypot(edgeX-anchorX, edgeY-anchorY) > limit ||
		math.Hypot(arrivalX-edgeX, arrivalY-edgeY) > limit {
		return fmt.Errorf("world: generated seam exceeds collision search bound")
	}
	if math.Hypot(arrivalX-edgeX, arrivalY-edgeY) <= float64(spec.TriggerRadius) {
		return fmt.Errorf("world: generated seam arrival is inside return trigger")
	}
	interiorX, interiorY, found := world.OpenPointNearCenter()
	if !found || !insideCentralHalf(world, interiorX, interiorY) ||
		math.Hypot(interiorX-edgeX, interiorY-edgeY) <= float64(spec.ArrivalInset+spec.TriggerRadius) {
		return fmt.Errorf("world: generated seam has no open interior")
	}
	if _, err := world.FindPath(gameworld.PathRequest{
		Start: gameworld.Point{X: arrivalX, Y: arrivalY},
		Goal:  gameworld.Point{X: interiorX, Y: interiorY},
	}); err != nil {
		return fmt.Errorf("world: generated seam has no route inward: %w", err)
	}
	if _, err := world.FindPath(gameworld.PathRequest{
		Start:      gameworld.Point{X: arrivalX, Y: arrivalY},
		Goal:       gameworld.Point{X: edgeX, Y: edgeY},
		StopRadius: float64(spec.TriggerRadius),
	}); err != nil {
		return fmt.Errorf("world: generated seam has no route to return trigger: %w", err)
	}
	return nil
}

// insideCentralHalf prevents a center search from treating a narrow open border strip as playable interior.
func insideCentralHalf(world *gameworld.Map, x, y float64) bool {
	return x >= float64(world.WidthSubtiles)/4 && x < float64(world.WidthSubtiles)*3/4 &&
		y >= float64(world.HeightSubtiles)/4 && y < float64(world.HeightSubtiles)*3/4
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

// resolveGeneratedTile converts a generated warp tile coordinate to its center subtile before collision lookup.
func resolveGeneratedTile(world *gameworld.Map, tileX, tileY int) (float64, float64, error) {
	x := float64(tileX*gameworld.SubtilesPerTile) + 2.5
	y := float64(tileY*gameworld.SubtilesPerTile) + 2.5

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

// insetArrival moves the policy-specified distance inward so arrivals do not immediately retrigger the seam.
func insetArrival(world *gameworld.Map, x, y float64, edge string, inset int) (float64, float64, bool) {
	step := float64(inset)
	delta := map[string][2]float64{"north": {0, step}, "east": {-step, 0}, "south": {0, -step}, "west": {step, 0}}[edge]
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
