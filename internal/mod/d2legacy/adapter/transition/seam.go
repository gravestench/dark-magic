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
	} else if _, err := firstMap.FindPath(gameworld.PathRequest{
		Start:      gameworld.Point{X: firstArrivalX, Y: firstArrivalY},
		Goal:       gameworld.Point{X: firstX, Y: firstY},
		Radius:     1,
		StopRadius: float64(seamTriggerRadius(spec)),
	}); err != nil {
		return Seam{}, fmt.Errorf("world: authored seam has no route to return trigger: %w", err)
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

	return Seam{First: first, Second: second, TriggerRadius: float64(seamTriggerRadius(spec))}, nil
}

// seamTriggerRadius applies the default used by both route validation and the published trigger.
func seamTriggerRadius(spec d2mapgen.SeamSpec) int {
	if spec.TriggerRadius == 0 {
		return 2
	}
	return spec.TriggerRadius
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

// resolveAuthoredEdge projects the authored guide onto a traversable floor corridor at the named stamp edge.
func resolveAuthoredEdge(spec d2mapgen.SeamSpec, world *gameworld.Map) (float64, float64, error) {
	anchor, found := cardinalAnchor(world.AuthoredExitAnchors(), spec.FirstDirection)
	if !found {
		return 0, 0, fmt.Errorf("world: first map has no authored seam anchor")
	}

	guideX, guideY, found := world.OpenPointNearExit(anchor)
	if !found {
		return 0, 0, fmt.Errorf("world: first seam edge is blocked")
	}
	line, found := newAuthoredEdgeLine(world, spec.FirstDirection)
	if !found {
		return 0, 0, fmt.Errorf("world: first map has no authored edge")
	}

	floors := make(map[[2]int]bool)
	for _, tile := range world.Tiles {
		if tile.Layer == gameworld.LayerFloor && tile.Reference.Path != "" {
			floors[[2]int{tile.X, tile.Y}] = true
		}
	}

	guide := gameworld.Point{X: guideX, Y: guideY}
	var corridors [][2]int
	for index := 1; index < line.length-1; index++ {
		point := line.point(index, line.inner)
		if !authoredFloorAt(floors, point) || !authoredOutwardLane(world, floors, line, index) {
			continue
		}
		if _, err := world.FindPath(gameworld.PathRequest{Start: guide, Goal: point, Radius: 1}); err != nil {
			continue
		}
		if len(corridors) == 0 || corridors[len(corridors)-1][1] != index-1 {
			corridors = append(corridors, [2]int{index, index})
		} else {
			corridors[len(corridors)-1][1] = index
		}
	}
	if len(corridors) == 0 {
		return 0, 0, fmt.Errorf("world: first map has no reachable authored exit corridor")
	}

	guideIndex := gameworld.CollisionCell(guideY)
	if !line.vertical {
		guideIndex = gameworld.CollisionCell(guideX)
	}
	best := corridors[0]
	for _, corridor := range corridors[1:] {
		if corridorDistance(corridor, guideIndex) < corridorDistance(best, guideIndex) {
			best = corridor
		}
	}
	selected := guideIndex
	if selected < best[0] || selected > best[1] {
		selected = (best[0] + best[1]) / 2
	}
	point := line.point(selected, line.inner)
	return point.X, point.Y, nil
}

type authoredEdgeLine struct {
	vertical           bool
	inner, outer, step int
	length             int
}

// The endpoint is one tile inside the stamp edge; the outer cell leaves one collision cell for a radius-1 footprint.
func newAuthoredEdgeLine(world *gameworld.Map, direction string) (authoredEdgeLine, bool) {
	switch direction {
	case "north":
		return authoredEdgeLine{
			inner: gameworld.SubtilesPerTile, outer: 1, step: -1, length: world.WidthSubtiles,
		}, world.HeightSubtiles > 2*gameworld.SubtilesPerTile
	case "east":
		return authoredEdgeLine{vertical: true, inner: world.WidthSubtiles - gameworld.SubtilesPerTile,
				outer: world.WidthSubtiles - 2, step: 1, length: world.HeightSubtiles},
			world.WidthSubtiles > 2*gameworld.SubtilesPerTile
	case "south":
		return authoredEdgeLine{inner: world.HeightSubtiles - gameworld.SubtilesPerTile,
				outer: world.HeightSubtiles - 2, step: 1, length: world.WidthSubtiles},
			world.HeightSubtiles > 2*gameworld.SubtilesPerTile
	case "west":
		return authoredEdgeLine{vertical: true, inner: gameworld.SubtilesPerTile,
			outer: 1, step: -1, length: world.HeightSubtiles}, world.WidthSubtiles > 2*gameworld.SubtilesPerTile
	default:
		return authoredEdgeLine{}, false
	}
}

// point converts one corridor index and inward/outward position to subtile coordinates for either axis.
func (line authoredEdgeLine) point(index, along int) gameworld.Point {
	if line.vertical {
		return gameworld.Point{X: float64(along), Y: float64(index)}
	}
	return gameworld.Point{X: float64(index), Y: float64(along)}
}

// authoredFloorAt rejects open collision cells that have no visible authored floor placement.
func authoredFloorAt(floors map[[2]int]bool, point gameworld.Point) bool {
	return floors[[2]int{gameworld.CollisionCell(point.X) / gameworld.SubtilesPerTile,
		gameworld.CollisionCell(point.Y) / gameworld.SubtilesPerTile}]
}

// authoredOutwardLane requires a continuous player-sized floor walk from the inset to the outer safe cell.
func authoredOutwardLane(world *gameworld.Map, floors map[[2]int]bool, line authoredEdgeLine, index int) bool {
	previous := line.point(index, line.inner)
	for along := line.inner + line.step; ; along += line.step {
		point := line.point(index, along)
		if !authoredFloorAt(floors, point) || !world.WalkableStep(previous, point, 1) {
			return false
		}
		if along == line.outer {
			return true
		}
		previous = point
	}
}

// corridorDistance ranks edge openings by the authored guide without favoring one side of a wide corridor.
func corridorDistance(corridor [2]int, guide int) int {
	if guide < corridor[0] {
		return corridor[0] - guide
	}
	if guide > corridor[1] {
		return guide - corridor[1]
	}
	return 0
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
