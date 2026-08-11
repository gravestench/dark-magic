package mapgen

import (
	"fmt"
	"sort"

	gamedata "github.com/gravestench/dark-magic/internal/game/data/catalog"
	model "github.com/gravestench/dark-magic/internal/game/data/model"
)

const (
	connectionWest  uint8 = 1
	connectionEast  uint8 = 2
	connectionSouth uint8 = 4
	connectionNorth uint8 = 8
)

type mazeCell struct{ x, y int }
type mazeEdge struct{ a, b mazeCell }

// MazeGenerator implements the first deliberately narrow maze slice: ordinary
// Act I cave rooms (LevelType 3). Special quest/exit room replacement remains a
// later rule layer, but every ordinary chamber is selected by its exact authored
// W/E/S/N connection mask rather than by visual guesswork.
type MazeGenerator struct{ data gamedata.Snapshot }

func NewMazeGenerator(data gamedata.Snapshot) *MazeGenerator { return &MazeGenerator{data: data} }

func (generator *MazeGenerator) Generate(request Request) (*Zone, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	level, found := generator.data.LevelsByID[request.LevelID]
	if !found || level.Act+1 != int(request.Act) {
		return nil, fmt.Errorf("%w: level %d is absent or belongs to another act", ErrRequest, request.LevelID)
	}
	if level.DrlgType != 1 {
		return nil, fmt.Errorf("%w: level %d uses DRLG type %d, not maze type 1", ErrRequest, request.LevelID, level.DrlgType)
	}
	if level.LevelType != 3 {
		return nil, fmt.Errorf("%w: first maze slice supports Act I cave LevelType 3, got %d", ErrRequest, level.LevelType)
	}
	rules, found := generator.data.LevelMazeByLevel[request.LevelID]
	if !found || rules.SizeX <= 0 || rules.SizeY <= 0 {
		return nil, fmt.Errorf("%w: level %d has incomplete LvlMaze rules", ErrRequest, request.LevelID)
	}
	roomCount := mazeRoomCount(rules, request.Difficulty)
	if roomCount <= 0 {
		return nil, fmt.Errorf("%w: level %d requests no maze rooms", ErrRequest, request.LevelID)
	}
	cells, edges := growMaze(roomCount, NewStreams(request.Seed).For("maze-topology"))
	edges = mergeAdjacentCells(cells, edges, rules.Merge, NewStreams(request.Seed).For("maze-merge"))
	roles, err := specialRoomRoles(cells, edges, NewStreams(request.Seed).For("maze-special-rooms"))
	if err != nil {
		return nil, err
	}
	return generator.materialize(request, level, rules, cells, edges, roles)
}

func growMaze(count int, random Random) ([]mazeCell, []mazeEdge) {
	cells := []mazeCell{{}}
	var edges []mazeEdge
	occupied := map[mazeCell]bool{{}: true}
	for len(cells) < count {
		type frontier struct{ from, to mazeCell }
		var choices []frontier
		for _, from := range cells {
			for _, delta := range []mazeCell{{-1, 0}, {1, 0}, {0, 1}, {0, -1}} {
				to := mazeCell{from.x + delta.x, from.y + delta.y}
				if !occupied[to] {
					choices = append(choices, frontier{from: from, to: to})
				}
			}
		}
		picked := choices[random.Uint64n(uint64(len(choices)))]
		occupied[picked.to] = true
		cells = append(cells, picked.to)
		edges = append(edges, canonicalEdge(picked.from, picked.to))
	}
	return cells, edges
}

func mergeAdjacentCells(cells []mazeCell, edges []mazeEdge, chance int, random Random) []mazeEdge {
	existing := make(map[mazeEdge]bool, len(edges))
	for _, edge := range edges {
		existing[edge] = true
	}
	degrees := cellDegrees(edges)
	occupied := make(map[mazeCell]bool, len(cells))
	for _, cell := range cells {
		occupied[cell] = true
	}
	chance = min(max(chance, 0), 1000)
	for _, cell := range cells {
		for _, delta := range []mazeCell{{1, 0}, {0, 1}} {
			other := mazeCell{cell.x + delta.x, cell.y + delta.y}
			edge := canonicalEdge(cell, other)
			// Keep tree leaves available for authored one-door entrance/exit
			// chambers. Internal loops may merge; endpoints retain their role.
			if occupied[other] && !existing[edge] && degrees[cell] > 1 && degrees[other] > 1 && int(random.Uint64n(1000)) < chance {
				existing[edge] = true
				edges = append(edges, edge)
			}
		}
	}
	return edges
}

func (generator *MazeGenerator) materialize(request Request, level model.LevelData, rules model.LevelMazeData, cells []mazeCell, edges []mazeEdge, roles map[mazeCell]string) (*Zone, error) {
	minX, minY, maxX, maxY := cells[0].x, cells[0].y, cells[0].x, cells[0].y
	for _, cell := range cells[1:] {
		minX, minY, maxX, maxY = min(minX, cell.x), min(minY, cell.y), max(maxX, cell.x), max(maxY, cell.y)
	}
	sort.Slice(cells, func(i, j int) bool {
		if cells[i].y == cells[j].y {
			return cells[i].x < cells[j].x
		}
		return cells[i].y < cells[j].y
	})
	ids := make(map[mazeCell]uint32, len(cells))
	for index, cell := range cells {
		ids[cell] = uint32(index + 1)
	}
	links := make([]Link, 0, len(edges))
	masks := make(map[mazeCell]uint8, len(cells))
	for _, edge := range edges {
		from, to := ids[edge.a], ids[edge.b]
		links = append(links, Link{From: from, To: to})
		applyConnection(masks, edge.a, edge.b)
	}
	rooms := make([]Room, 0, len(cells))
	stamps := make([]Stamp, 0, len(cells))
	for _, cell := range cells {
		id, mask := ids[cell], masks[cell]
		presetDef := ordinaryCavePreset(mask)
		role := roles[cell]
		if role != "" {
			presetDef = specialCavePreset(role, mask)
		}
		preset, found := generator.data.LevelPresetByDef[presetDef]
		if !found || preset.SizeX != rules.SizeX || preset.SizeY != rules.SizeY {
			return nil, fmt.Errorf("%w: cave connection mask %#x has no compatible LvlPrest %d", ErrZone, mask, presetDef)
		}
		variants := presetFiles(preset)
		if len(variants) == 0 {
			return nil, fmt.Errorf("%w: cave LvlPrest %d has no DS1 variants", ErrZone, presetDef)
		}
		variant := int(NewStreams(request.Seed).For(fmt.Sprintf("maze-room-%d-variant", id)).Uint64n(uint64(len(variants))))
		tiles, err := maskedTilePaths(generator.data.LevelTypes, level.LevelType, preset.Dt1Mask)
		if err != nil {
			return nil, err
		}
		x, y := (cell.x-minX)*rules.SizeX, (cell.y-minY)*rules.SizeY
		stamps = append(stamps, Stamp{ID: id, PresetDef: presetDef, Role: role, X: x, Y: y, Width: rules.SizeX, Height: rules.SizeY, DS1Path: assetPath(variants[variant]), TilePaths: tiles, Variant: variant, Populate: preset.Populate != 0, LogicalWalls: preset.Logicals != 0})
		rooms = append(rooms, Room{ID: id, X: x, Y: y, Width: rules.SizeX, Height: rules.SizeY, StampID: id})
	}
	return NewZone(Definition{Request: request, Kind: Maze, Bounds: Bounds{Width: (maxX - minX + 1) * rules.SizeX, Height: (maxY - minY + 1) * rules.SizeY}, Stamps: stamps, Rooms: rooms, Links: links, Trace: []string{
		fmt.Sprintf("LvlMaze[%d] requested %d rooms of %dx%d", request.LevelID, len(rooms), rules.SizeX, rules.SizeY),
		fmt.Sprintf("topology created %d canonical room links", len(links)),
		"Act I Cave chamber definitions selected by exact W/E/S/N masks",
		"distinct leaf chambers assigned previous-level and next-level roles",
	}})
}

func ordinaryCavePreset(mask uint8) int { return 52 + int(mask) }

func specialCavePreset(role string, mask uint8) int {
	ordinal := map[uint8]int{connectionWest: 0, connectionEast: 1, connectionSouth: 2, connectionNorth: 3}[mask]
	switch role {
	case "previous-level":
		return 83 + ordinal
	case "next-level":
		return 87 + ordinal
	default:
		return ordinaryCavePreset(mask)
	}
}

func specialRoomRoles(cells []mazeCell, edges []mazeEdge, random Random) (map[mazeCell]string, error) {
	degrees := cellDegrees(edges)
	var leaves []mazeCell
	for _, cell := range cells {
		if degrees[cell] == 1 {
			leaves = append(leaves, cell)
		}
	}
	if len(cells) == 1 {
		return map[mazeCell]string{cells[0]: "previous-level"}, nil
	}
	if len(leaves) < 2 {
		return nil, fmt.Errorf("%w: maze has fewer than two endpoint rooms", ErrZone)
	}
	sort.Slice(leaves, func(i, j int) bool {
		if leaves[i].y == leaves[j].y {
			return leaves[i].x < leaves[j].x
		}
		return leaves[i].y < leaves[j].y
	})
	entrance := leaves[random.Uint64n(uint64(len(leaves)))]
	distances := cellDistances(entrance, edges)
	exit := leaves[0]
	for _, candidate := range leaves {
		if candidate != entrance && (exit == entrance || distances[candidate] > distances[exit]) {
			exit = candidate
		}
	}
	return map[mazeCell]string{entrance: "previous-level", exit: "next-level"}, nil
}

func cellDegrees(edges []mazeEdge) map[mazeCell]int {
	result := make(map[mazeCell]int)
	for _, edge := range edges {
		result[edge.a]++
		result[edge.b]++
	}
	return result
}

func cellDistances(start mazeCell, edges []mazeEdge) map[mazeCell]int {
	adjacent := make(map[mazeCell][]mazeCell)
	for _, edge := range edges {
		adjacent[edge.a] = append(adjacent[edge.a], edge.b)
		adjacent[edge.b] = append(adjacent[edge.b], edge.a)
	}
	result := map[mazeCell]int{start: 0}
	queue := []mazeCell{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range adjacent[current] {
			if _, found := result[next]; !found {
				result[next] = result[current] + 1
				queue = append(queue, next)
			}
		}
	}
	return result
}

func mazeRoomCount(record model.LevelMazeData, difficulty Difficulty) int {
	switch difficulty {
	case Nightmare:
		return record.RoomsN
	case Hell:
		return record.RoomsH
	default:
		return record.Rooms
	}
}

func canonicalEdge(a, b mazeCell) mazeEdge {
	if b.y < a.y || (b.y == a.y && b.x < a.x) {
		a, b = b, a
	}
	return mazeEdge{a: a, b: b}
}

func applyConnection(masks map[mazeCell]uint8, a, b mazeCell) {
	switch {
	case b.x == a.x-1:
		masks[a] |= connectionWest
		masks[b] |= connectionEast
	case b.x == a.x+1:
		masks[a] |= connectionEast
		masks[b] |= connectionWest
	case b.y == a.y+1:
		masks[a] |= connectionSouth
		masks[b] |= connectionNorth
	case b.y == a.y-1:
		masks[a] |= connectionNorth
		masks[b] |= connectionSouth
	}
}
