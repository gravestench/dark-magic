package mapgen

import (
	"fmt"
	"strings"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	model "github.com/gravestench/dark-magic/internal/game/data/model"
)

const actOneOutdoorCellSize = 8

// ActOneOutdoorGenerator is the first intentionally narrow outdoor strategy.
// It builds Blood Moor's authored 8x8 coarse grid and preserves the edge that
// joins Rogue Encampment. Structural facts are emitted separately from authored
// fill stamps so simulation can validate rivers, cliffs, and bridge openings
// before presentation selects legacy DT1 graphics.
type ActOneOutdoorGenerator struct{ data gamedata.Snapshot }

func NewActOneOutdoorGenerator(data gamedata.Snapshot) *ActOneOutdoorGenerator {
	return &ActOneOutdoorGenerator{data: data}
}

// GenerateFromTown ties Blood Moor's town-facing edge to the selected town
// layout. The town role is part of the immutable recipe, never inferred from a
// renderer node or the camera.
func (generator *ActOneOutdoorGenerator) GenerateFromTown(request Request, town Stamp) (*Zone, error) {
	direction, err := townExitDirection(town.Role)
	if err != nil {
		return nil, err
	}
	return generator.generate(request, direction)
}

func (generator *ActOneOutdoorGenerator) Generate(request Request) (*Zone, error) {
	return generator.generate(request, "south")
}

func (generator *ActOneOutdoorGenerator) generate(request Request, townDirection string) (*Zone, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	level, found := generator.data.LevelsByID[request.LevelID]
	if !found || request.Act != 1 || level.Act != 0 {
		return nil, fmt.Errorf("%w: Blood Moor request belongs to Act I", ErrRequest)
	}
	if request.LevelID != 2 || level.DrlgType != 3 || level.LevelType != 2 {
		return nil, fmt.Errorf("%w: first outdoor strategy supports Blood Moor level 2", ErrRequest)
	}
	width, height := levelSize(level, request.Difficulty)
	if width <= 0 || height <= 0 || width%actOneOutdoorCellSize != 0 || height%actOneOutdoorCellSize != 0 {
		return nil, fmt.Errorf("%w: Blood Moor dimensions %dx%d are not an 8-tile grid", ErrZone, width, height)
	}
	columns, rows := width/actOneOutdoorCellSize, height/actOneOutdoorCellSize
	entry := townEdgeWarp(width, height, townDirection)
	exit := nextLevelEdgeWarp(width, height, entry.Direction)
	route := outdoorRoute(request.Seed, columns, rows, entry.Direction)
	path := outdoorPathTiles(route, columns, rows, entry, exit)
	structures, structuralStamps, path, err := generator.outdoorStructures(request, level, columns, rows, entry.Direction, path)
	if err != nil {
		return nil, err
	}
	stamps := make([]Stamp, 0, columns*rows+len(structuralStamps))
	rooms := make([]Room, 0, columns*rows)
	links := make([]Link, 0, (columns-1)*rows+(rows-1)*columns)
	for y := 0; y < rows; y++ {
		for x := 0; x < columns; x++ {
			id := uint32(y*columns + x + 1)
			presetDef := []int{29, 30, 35}[NewStreams(request.Seed).For(fmt.Sprintf("outdoor-cell-%d-preset", id)).Uint64n(3)]
			preset, found := generator.data.LevelPresetByDef[presetDef]
			if !found || preset.SizeX != actOneOutdoorCellSize || preset.SizeY != actOneOutdoorCellSize {
				return nil, fmt.Errorf("%w: Blood Moor fill preset %d is unavailable", ErrZone, presetDef)
			}
			stamp, err := generator.outdoorStamp(request, level, preset, id, x*actOneOutdoorCellSize, y*actOneOutdoorCellSize)
			if err != nil {
				return nil, err
			}
			if route[[2]int{x, y}] {
				stamp.Role = "blood-moor-route"
			}
			stamps = append(stamps, stamp)
			rooms = append(rooms, Room{ID: id, X: stamp.X, Y: stamp.Y, Width: stamp.Width, Height: stamp.Height, StampID: id})
			if x > 0 {
				links = append(links, Link{From: id - 1, To: id})
			}
			if y > 0 {
				links = append(links, Link{From: id - uint32(columns), To: id})
			}
		}
	}
	stamps = append(stamps, structuralStamps...)
	return NewZone(Definition{Request: request, Kind: Outdoor, Bounds: Bounds{Width: width, Height: height}, Stamps: stamps, Rooms: rooms, Links: links, Warps: []Warp{entry, exit}, Paths: path, Structures: structures, Trace: []string{
		fmt.Sprintf("Levels[%d] selected Act I outdoor strategy on a %dx%d coarse grid", request.LevelID, columns, rows),
		"authored 8x8 Blood Moor fill presets selected by independent cell streams",
		fmt.Sprintf("Rogue Encampment joins the %s Blood Moor edge", oppositeDirection(townDirection)),
		fmt.Sprintf("a deterministic %d-cell route joins town to the opposite next-level edge", len(route)),
		"a continuous river and cliff ridge preserve explicit passable crossings on the primary route",
	}})
}

// outdoorStructures creates an authored vertical river band plus a semantic
// cliff ridge. East/west routes cross the bridge; north/south routes keep the
// river near the opposite edge. The cliff always leaves a route opening.
func (generator *ActOneOutdoorGenerator) outdoorStructures(request Request, level model.LevelData, columns, rows int, entryDirection string, path []PathTile) ([]StructureTile, []Stamp, []PathTile, error) {
	plan := planActOneRiver(request.Seed, columns, rows, entryDirection, path)
	if plan.crossesRoute {
		path = pathThroughBridge(path, plan.column*actOneOutdoorCellSize, plan.row*actOneOutdoorCellSize+actOneOutdoorCellSize/2)
	}
	structures := riverStructureTiles(plan, rows)
	structures = append(structures, cliffStructureTiles(request.Seed, columns, rows, path, plan)...)
	stamps, err := generator.outdoorStructureStamps(request, level, columns, rows, plan)
	if err != nil {
		return nil, nil, nil, err
	}
	return structures, stamps, path, nil
}

type actOneRiverPlan struct {
	column, row  int
	crossesRoute bool
}

func planActOneRiver(seed uint64, columns, rows int, entryDirection string, path []PathTile) actOneRiverPlan {
	plan := actOneRiverPlan{column: columns/2 - 1, row: rows / 2, crossesRoute: entryDirection == "west" || entryDirection == "east"}
	if !plan.crossesRoute {
		// Act I's authored river presets are vertical in tile space. For a
		// north/south primary route, keep the band near the opposite edge.
		plan.column = columns - 2
		plan.row = int(NewStreams(seed).For("blood-moor-bridge-row").Uint64n(uint64(rows)))
		return plan
	}
	for _, tile := range path {
		if tile.X >= plan.column*actOneOutdoorCellSize && tile.X < (plan.column+2)*actOneOutdoorCellSize {
			plan.row = tile.Y / actOneOutdoorCellSize
			break
		}
	}
	return plan
}

func riverStructureTiles(plan actOneRiverPlan, rows int) []StructureTile {
	result := make([]StructureTile, 0, rows*actOneOutdoorCellSize*2)
	for y := 0; y < rows*actOneOutdoorCellSize; y++ {
		for x := plan.column * actOneOutdoorCellSize; x < (plan.column+2)*actOneOutdoorCellSize; x++ {
			kind, passable := "river", false
			if y/actOneOutdoorCellSize == plan.row && y%actOneOutdoorCellSize >= 2 && y%actOneOutdoorCellSize <= 5 {
				kind = "bridge"
				localX, localY := x-plan.column*actOneOutdoorCellSize, y%actOneOutdoorCellSize
				passable = localY == 3 || localY == 4 || (localY == 2 || localY == 5) && (localX < 4 || localX >= 12)
			}
			result = append(result, StructureTile{X: x, Y: y, Kind: kind, Passable: passable})
		}
	}
	return result
}

func cliffStructureTiles(seed uint64, columns, rows int, path []PathTile, river actOneRiverPlan) []StructureTile {
	y := rows * actOneOutdoorCellSize / 4
	if NewStreams(seed).For("blood-moor-cliff-side").Uint64n(2) == 1 {
		y = rows*actOneOutdoorCellSize - 1 - y
	}
	pathSet := make(map[PathTile]bool, len(path))
	for _, tile := range path {
		pathSet[tile] = true
	}
	result := make([]StructureTile, 0, columns*actOneOutdoorCellSize)
	for x := 0; x < columns*actOneOutdoorCellSize; x++ {
		if x >= river.column*actOneOutdoorCellSize && x < (river.column+2)*actOneOutdoorCellSize {
			continue
		}
		open := false
		for offset := -1; offset <= 1; offset++ {
			open = open || pathSet[PathTile{X: x, Y: y + offset}]
		}
		if !open {
			result = append(result, StructureTile{X: x, Y: y, Kind: "cliff"})
		}
	}
	return result
}

func (generator *ActOneOutdoorGenerator) outdoorStructureStamps(request Request, level model.LevelData, columns, rows int, plan actOneRiverPlan) ([]Stamp, error) {
	stamps := make([]Stamp, 0, rows*2)
	for row := 0; row < rows; row++ {
		for side, presetDef := range []int{26, 27} {
			// D2MOO's nPickedFile is the authored one-based file number. Stamp
			// variants are zero-based, so river file 3 and bridge files 1/3 map
			// to indices 2 and 0/2 respectively.
			variant := 2
			if row == plan.row {
				presetDef = 28
				variant = []int{0, 2}[side]
			}
			preset, found := generator.data.LevelPresetByDef[presetDef]
			if !found {
				return nil, fmt.Errorf("%w: Act I structural preset %d is unavailable", ErrZone, presetDef)
			}
			stamp, err := generator.outdoorStructuralStamp(request, level, preset, variant,
				uint32(columns*rows+row*2+side+1), (plan.column+side)*actOneOutdoorCellSize, row*actOneOutdoorCellSize)
			if err != nil {
				return nil, err
			}
			stamps = append(stamps, stamp)
		}
	}
	return stamps, nil
}

// pathThroughBridge bends only the portion of the semantic route adjacent to
// the river. Old path cells inside the water band are removed, then both banks
// are connected to the authored bridge center. This keeps the route continuous
// without changing either endpoint or unrelated outdoor topology.
func pathThroughBridge(path []PathTile, riverX, bridgeY int) []PathTile {
	const riverWidth = 2 * actOneOutdoorCellSize
	left, right := PathTile{X: riverX - 1, Y: bridgeY}, PathTile{X: riverX + riverWidth, Y: bridgeY}
	leftSource, rightSource := left, right
	leftDistance, rightDistance := int(^uint(0)>>1), int(^uint(0)>>1)
	result := make(map[PathTile]bool, len(path)+riverWidth)
	for _, tile := range path {
		if tile.X < riverX || tile.X >= riverX+riverWidth {
			result[tile] = true
		}
		if tile.X == left.X && absInt(tile.Y-bridgeY) < leftDistance {
			leftSource, leftDistance = tile, absInt(tile.Y-bridgeY)
		}
		if tile.X == right.X && absInt(tile.Y-bridgeY) < rightDistance {
			rightSource, rightDistance = tile, absInt(tile.Y-bridgeY)
		}
	}
	rasterPath(leftSource, left, result)
	rasterPath(left, right, result)
	rasterPath(right, rightSource, result)
	connected := make([]PathTile, 0, len(result))
	for tile := range result {
		connected = append(connected, tile)
	}
	return connected
}

func (generator *ActOneOutdoorGenerator) outdoorStructuralStamp(request Request, level model.LevelData, preset model.LevelPreset, variant int, id uint32, x, y int) (Stamp, error) {
	variants := presetFiles(preset)
	if variant < 0 || variant >= len(variants) {
		return Stamp{}, fmt.Errorf("%w: structural preset %d lacks variant %d", ErrZone, preset.Def, variant)
	}
	tiles, err := maskedTilePaths(generator.data.LevelTypes, level.LevelType, preset.Dt1Mask)
	if err != nil {
		return Stamp{}, err
	}
	return Stamp{ID: id, PresetDef: preset.Def, Role: "blood-moor-structure", X: x, Y: y,
		Width: preset.SizeX, Height: preset.SizeY, DS1Path: assetPath(variants[variant]), TilePaths: tiles,
		Variant: variant, Populate: false, LogicalWalls: preset.Logicals != 0, Overlay: true}, nil
}

func outdoorPathTiles(route map[[2]int]bool, columns, rows int, entry, exit Warp) []PathTile {
	points := []PathTile{{X: entry.X, Y: entry.Y}}
	horizontal := entry.Direction == "west" || entry.Direction == "east"
	forward := entry.Direction == "west" || entry.Direction == "north"
	length := rows
	if horizontal {
		length = columns
	}
	for step := 0; step < length; step++ {
		axis := step
		if !forward {
			axis = length - step - 1
		}
		for cell := range route {
			cellAxis := cell[1]
			if horizontal {
				cellAxis = cell[0]
			}
			if cellAxis == axis {
				points = append(points, PathTile{X: cell[0]*actOneOutdoorCellSize + actOneOutdoorCellSize/2, Y: cell[1]*actOneOutdoorCellSize + actOneOutdoorCellSize/2})
				break
			}
		}
	}
	points = append(points, PathTile{X: exit.X, Y: exit.Y})
	seen := map[PathTile]bool{}
	for index := 1; index < len(points); index++ {
		rasterPath(points[index-1], points[index], seen)
	}
	result := make([]PathTile, 0, len(seen))
	for tile := range seen {
		result = append(result, tile)
	}
	return result
}

// rasterPath is an integer Bresenham walk. Diagonal steps are legal because
// the legacy dirt-path realization selects art from all eight neighbors.
func rasterPath(from, to PathTile, result map[PathTile]bool) {
	x, y := from.X, from.Y
	dx, dy := absInt(to.X-x), -absInt(to.Y-y)
	sx, sy := -1, -1
	if x < to.X {
		sx = 1
	}
	if y < to.Y {
		sy = 1
	}
	err := dx + dy
	for {
		result[PathTile{X: x, Y: y}] = true
		if x == to.X && y == to.Y {
			return
		}
		twice := 2 * err
		if twice >= dy {
			err += dy
			x += sx
		}
		if twice <= dx {
			err += dx
			y += sy
		}
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func (generator *ActOneOutdoorGenerator) outdoorStamp(request Request, level model.LevelData, preset model.LevelPreset, id uint32, x, y int) (Stamp, error) {
	variants := presetFiles(preset)
	if len(variants) == 0 {
		return Stamp{}, fmt.Errorf("%w: outdoor preset %d has no DS1 variants", ErrZone, preset.Def)
	}
	variant := int(NewStreams(request.Seed).For(fmt.Sprintf("outdoor-cell-%d-variant", id)).Uint64n(uint64(len(variants))))
	tiles, err := maskedTilePaths(generator.data.LevelTypes, level.LevelType, preset.Dt1Mask)
	if err != nil {
		return Stamp{}, err
	}
	return Stamp{ID: id, PresetDef: preset.Def, Role: "blood-moor-fill", X: x, Y: y, Width: preset.SizeX, Height: preset.SizeY, DS1Path: assetPath(variants[variant]), TilePaths: tiles, Variant: variant, Populate: preset.Populate != 0, LogicalWalls: preset.Logicals != 0}, nil
}

func townExitDirection(role string) (string, error) {
	const prefix = "act1-town:exit-"
	if !strings.HasPrefix(role, prefix) {
		return "", fmt.Errorf("%w: town stamp has no cardinal exit role", ErrRequest)
	}
	direction := strings.TrimPrefix(role, prefix)
	if direction != "north" && direction != "east" && direction != "south" && direction != "west" {
		return "", fmt.Errorf("%w: unknown town exit %q", ErrRequest, direction)
	}
	return direction, nil
}

func townEdgeWarp(width, height int, townDirection string) Warp {
	direction := oppositeDirection(townDirection)
	x, y := width/2, height/2
	switch direction {
	case "north":
		y = 0
	case "east":
		x = width - 1
	case "south":
		y = height - 1
	case "west":
		x = 0
	}
	return Warp{ID: 1, Role: "town-entry", Direction: direction, X: x, Y: y, DestinationLevel: 1}
}

func nextLevelEdgeWarp(width, height int, entryDirection string) Warp {
	direction := oppositeDirection(entryDirection)
	warp := townEdgeWarp(width, height, oppositeDirection(direction))
	warp.ID = 2
	warp.Role = "next-level-exit"
	warp.Direction = direction
	warp.DestinationLevel = 3
	return warp
}

// outdoorRoute reserves one coarse-cell-wide semantic corridor. Each step
// advances toward the far edge and may move sideways by one cell, so the path
// is contiguous, deterministic, and cannot fold back on itself. It does not
// pretend that a fill DS1 is road artwork; later path/cliff/river layers use
// this recipe when choosing their authored pieces.
func outdoorRoute(seed uint64, columns, rows int, entryDirection string) map[[2]int]bool {
	horizontal := entryDirection == "west" || entryDirection == "east"
	length, crossSize := rows, columns
	if horizontal {
		length, crossSize = columns, rows
	}
	forward := entryDirection == "west" || entryDirection == "north"
	cross, target := crossSize/2, crossSize/2
	result := make(map[[2]int]bool, length)
	stream := NewStreams(seed).For("blood-moor-primary-route")
	for step := 0; step < length; step++ {
		axis := step
		if !forward {
			axis = length - step - 1
		}
		x, y := cross, axis
		if horizontal {
			x, y = axis, cross
		}
		result[[2]int{x, y}] = true
		remaining := length - step - 1
		if remaining == 0 {
			continue
		}
		delta := int(stream.Uint64n(3)) - 1
		next := max(0, min(crossSize-1, cross+delta))
		if next-target > remaining-1 {
			next = target + remaining - 1
		} else if target-next > remaining-1 {
			next = target - remaining + 1
		}
		cross = next
	}
	return result
}

func oppositeDirection(direction string) string {
	return map[string]string{"north": "south", "east": "west", "south": "north", "west": "east"}[direction]
}
