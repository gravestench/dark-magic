package entryworld

import (
	"fmt"
	"math"
	"strconv"

	models "github.com/gravestench/dark-magic/internal/game/data/model"
	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	d2mapgen "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/mapgen"
)

// CaveWarpEndpoint is a directed, trusted warp. Arrival and dimensions belong
// to TargetLevel; X/Y belong to LevelID. IDs survive snapshot rematerialization.
type CaveWarpEndpoint struct {
	ID, PairedID         string
	LevelID, TargetLevel int
	X, Y                 float64
	ArrivalX, ArrivalY   float64
	Width, Height        float64
}

// CaveWarpPair is published only after both materialized sides validate.
type CaveWarpPair struct{ First, Second CaveWarpEndpoint }

type caveSide struct {
	source, arrival gameworld.Point
	world           *gameworld.Map
}

// resolveCaveWarp joins the serialized Lua relationship to typed authored
// records and unique full-identity DS1 markers. It never uses recipe x/y.
func resolveCaveWarp(
	records Records, zones map[int]*worldgen.Zone, worlds map[int]*gameworld.Map, spec d2mapgen.CaveWarpSpec,
) (CaveWarpPair, error) {
	if err := d2mapgen.ValidateCaveWarpSpec(zones, &spec); err != nil {
		return CaveWarpPair{}, err
	}
	first, second := worlds[spec.FirstLevel], worlds[spec.SecondLevel]
	if records == nil || first == nil || second == nil || first == second {
		return CaveWarpPair{}, fmt.Errorf("cave warp requires distinct materialized maps and records")
	}
	firstWarp, secondWarp, err := caveWarpRecords(records, spec)
	if err != nil {
		return CaveWarpPair{}, err
	}
	// Owned 1.14d definition 51 uses 10/6/24; previous-level cave chambers
	// use 10/0/21. The Vis slot selects the sub-index; orientation alone is
	// insufficient, and MainIndex is part of each authored identity.
	moorIdentity := gameworld.TileIdentity{Orientation: 10, MainIndex: 6, SubIndex: int32(21 + spec.FirstVisSlot)}
	moorSide, err := resolveCaveSide(first, moorIdentity, firstWarp)
	if err != nil {
		return CaveWarpPair{}, fmt.Errorf("Blood Moor cave entrance: %w", err)
	}
	denIdentity := gameworld.TileIdentity{Orientation: 10, MainIndex: 0, SubIndex: int32(21 + spec.SecondVisSlot)}
	denSide, err := resolveCaveSide(second, denIdentity, secondWarp)
	if err != nil {
		return CaveWarpPair{}, fmt.Errorf("Den cave exit: %w", err)
	}
	if err := validateOutdoorCaveInterior(moorSide); err != nil {
		return CaveWarpPair{}, err
	}
	if !denHasReachableInterior(denSide.world, denSide.arrival) {
		return CaveWarpPair{}, fmt.Errorf("Den cave exit has no player-sized route into the central half")
	}
	firstID := fmt.Sprintf("static-warp:%d:%d", spec.FirstLevel, spec.FirstZoneWarpID)
	secondID := fmt.Sprintf("static-warp:%d:%d", spec.SecondLevel, spec.SecondZoneWarpID)
	return CaveWarpPair{
		First:  caveDirectedEndpoint(firstID, secondID, spec.FirstLevel, spec.SecondLevel, moorSide, denSide),
		Second: caveDirectedEndpoint(secondID, firstID, spec.SecondLevel, spec.FirstLevel, denSide, moorSide),
	}, nil
}

// caveDirectedEndpoint binds each source marker to the opposite side's validated arrival and map bounds.
func caveDirectedEndpoint(id, paired string, level, target int, source, destination caveSide) CaveWarpEndpoint {
	return CaveWarpEndpoint{ID: id, PairedID: paired, LevelID: level, TargetLevel: target,
		X: source.source.X, Y: source.source.Y, ArrivalX: destination.arrival.X, ArrivalY: destination.arrival.Y,
		Width: float64(destination.world.WidthSubtiles), Height: float64(destination.world.HeightSubtiles)}
}

// resolveCaveSide accepts only one authored marker and a traversable source/arrival pair.
func resolveCaveSide(world *gameworld.Map, identity gameworld.TileIdentity, warp models.LevelWarp) (caveSide, error) {
	var marker gameworld.SpecialTile
	count := 0
	for _, tile := range world.SpecialTiles {
		if tile.Orientation == identity.Orientation && tile.MainIndex == identity.MainIndex &&
			tile.SubIndex == identity.SubIndex {
			marker = tile
			count++
		}
	}
	if count != 1 {
		return caveSide{}, fmt.Errorf("expected one full-identity marker, found %d", count)
	}
	if marker.X < 0 || marker.Y < 0 || marker.X >= world.WidthTiles || marker.Y >= world.HeightTiles {
		return caveSide{}, fmt.Errorf("marker is outside map")
	}
	anchor := gameworld.Point{
		X: float64(marker.X*gameworld.SubtilesPerTile + warp.OffsetX),
		Y: float64(marker.Y*gameworld.SubtilesPerTile + warp.OffsetY),
	}
	source, err := caveOpenPoint(world, anchor)
	if err != nil {
		return caveSide{}, fmt.Errorf("source: %w", err)
	}
	// Offsets and ExitWalk are authored subtile units. The bounded radius-1
	// adjustment and immediate safe arrival are provisional Dark Magic policy;
	// no original-game seed or automatic-walk parity is claimed.
	arrival, err := caveOpenPoint(world, gameworld.Point{
		X: anchor.X + float64(warp.ExitWalkX), Y: anchor.Y + float64(warp.ExitWalkY),
	})
	if err != nil {
		return caveSide{}, fmt.Errorf("arrival: %w", err)
	}
	if math.Hypot(arrival.X-source.X, arrival.Y-source.Y) <= 2 {
		return caveSide{}, fmt.Errorf("arrival overlaps return interaction")
	}
	if _, err := world.FindPath(gameworld.PathRequest{Start: arrival, Goal: source, Radius: 1}); err != nil {
		return caveSide{}, fmt.Errorf("no player-sized source/arrival route: %w", err)
	}
	return caveSide{source: source, arrival: arrival, world: world}, nil
}

// The outdoor interior remains the existing center-area anchor. Preparation
// additionally proves connection to the trusted spawn from the town seam.
func validateOutdoorCaveInterior(side caveSide) error {
	world := side.world
	cx, cy, found := world.OpenPointNearSubtileForRadius(
		float64(world.WidthSubtiles)/2, float64(world.HeightSubtiles)/2, 1,
	)
	if !found || cx < float64(world.WidthSubtiles)/4 || cy < float64(world.HeightSubtiles)/4 ||
		cx >= float64(world.WidthSubtiles)*3/4 || cy >= float64(world.HeightSubtiles)*3/4 {
		return fmt.Errorf("Blood Moor cave entrance has no player-sized interior")
	}
	return validateCaveApproach(world, side.arrival, gameworld.Point{X: cx, Y: cy})
}

// validateCaveApproach rejects a marker whose player-sized route cannot reach the expected outdoor component.
func validateCaveApproach(world *gameworld.Map, source, destination gameworld.Point) error {
	if _, err := world.FindPath(gameworld.PathRequest{Start: source, Goal: destination, Radius: 1}); err != nil {
		return fmt.Errorf("cave entrance has no route to the established outdoor area: %w", err)
	}
	return nil
}

// Definition 85/File2 contains a disconnected nearest-center pocket. Prove
// that the arrival's own component reaches the central half instead. This is
// provisional Dark Magic admission policy, not a retail movement claim.
// The flood visits each subtile once and delegates all footprint/diagonal
// rules to WalkableStep. A 24x24 Den bounds it to 14,400 cells and eight edges
// per cell; no repeated whole-map path searches or persistent cache is needed.
func denHasReachableInterior(world *gameworld.Map, arrival gameworld.Point) bool {
	width, height := world.WidthSubtiles, world.HeightSubtiles
	startX, startY := gameworld.CollisionCell(arrival.X), gameworld.CollisionCell(arrival.Y)
	if !world.WalkableStep(arrival, arrival, 1) {
		return false
	}
	start := startY*width + startX
	visited := make([]bool, width*height)
	visited[start] = true
	queue := []int{start}
	for index := 0; index < len(queue); index++ {
		cell := queue[index]
		x, y := cell%width, cell/width
		if x >= width/4 && x < width*3/4 && y >= height/4 && y < height*3/4 {
			return true
		}
		from := gameworld.Point{X: float64(x), Y: float64(y)}
		for _, step := range [][2]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}, {-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
			nx, ny := x+step[0], y+step[1]
			if nx < 0 || ny < 0 || nx >= width || ny >= height {
				continue
			}
			next := ny*width + nx
			if visited[next] || !world.WalkableStep(from, gameworld.Point{X: float64(nx), Y: float64(ny)}, 1) {
				continue
			}
			visited[next] = true
			queue = append(queue, next)
		}
	}
	return false
}

// caveOpenPoint bounds the adjustment from an authored subtile anchor to a safe player-sized point.
func caveOpenPoint(world *gameworld.Map, anchor gameworld.Point) (gameworld.Point, error) {
	if anchor.X < 0 || anchor.Y < 0 || anchor.X >= float64(world.WidthSubtiles) ||
		anchor.Y >= float64(world.HeightSubtiles) {
		return gameworld.Point{}, fmt.Errorf("authored anchor outside map")
	}
	x, y, found := world.OpenPointNearSubtileForRadius(anchor.X, anchor.Y, 1)
	if !found || math.Hypot(x-anchor.X, y-anchor.Y) > 10 {
		return gameworld.Point{}, fmt.Errorf("no player-sized point within ten subtiles of authored anchor")
	}
	return gameworld.Point{X: x, Y: y}, nil
}

// The Records boundary supplies generic rows. Bind the exact used columns to
// the existing typed schemas, rejecting absent, malformed, and duplicate keys.
func caveWarpRecords(records Records, spec d2mapgen.CaveWarpSpec) (models.LevelWarp, models.LevelWarp, error) {
	levels, err := records.Load("data/global/excel/Levels.txt")
	if err != nil {
		return models.LevelWarp{}, models.LevelWarp{}, err
	}
	moor, err := uniqueCaveRecord(levels, spec.FirstLevel)
	if err != nil {
		return models.LevelWarp{}, models.LevelWarp{}, err
	}
	den, err := uniqueCaveRecord(levels, spec.SecondLevel)
	if err != nil {
		return models.LevelWarp{}, models.LevelWarp{}, err
	}
	var a, b models.LevelData
	for _, field := range []struct {
		row    map[string]string
		key    string
		target *int
	}{
		{moor, "Id", &a.Id}, {moor, "Vis3", &a.Vis3}, {moor, "Warp3", &a.Warp3},
		{den, "Id", &b.Id}, {den, "Vis0", &b.Vis0}, {den, "Warp0", &b.Warp0},
	} {
		value, err := caveRecordNumber(field.row, field.key)
		if err != nil {
			return models.LevelWarp{}, models.LevelWarp{}, err
		}
		*field.target = value
	}
	if a.Vis3 != b.Id || b.Vis0 != a.Id || a.Warp3 != spec.FirstLvlWarpID || b.Warp0 != spec.SecondLvlWarpID {
		return models.LevelWarp{}, models.LevelWarp{}, fmt.Errorf("Levels cave Vis/Warp relationship is not reciprocal")
	}
	rows, err := records.Load("data/global/excel/LvlWarp.txt")
	if err != nil {
		return models.LevelWarp{}, models.LevelWarp{}, err
	}
	first, err := bindCaveWarp(rows, a.Warp3)
	if err != nil {
		return models.LevelWarp{}, models.LevelWarp{}, err
	}
	second, err := bindCaveWarp(rows, b.Warp0)
	return first, second, err
}

// uniqueCaveRecord prevents an ambiguous authored row from selecting a different warp on restore.
func uniqueCaveRecord(rows []map[string]string, id int) (map[string]string, error) {
	var found map[string]string
	for _, row := range rows {
		if row["Id"] != strconv.Itoa(id) {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("duplicate cave record %d", id)
		}
		found = row
	}
	if found == nil {
		return nil, fmt.Errorf("missing cave record %d", id)
	}
	return found, nil
}

// caveRecordNumber rejects missing or malformed numeric fields before publishing endpoints.
func caveRecordNumber(row map[string]string, key string) (int, error) {
	value, err := strconv.Atoi(row[key])
	if err != nil {
		return 0, fmt.Errorf("invalid cave record %s: %w", key, err)
	}
	return value, nil
}

// bindCaveWarp reads only the offset fields needed for the selected interactive LvlWarp row.
func bindCaveWarp(rows []map[string]string, id int) (models.LevelWarp, error) {
	row, err := uniqueCaveRecord(rows, id)
	if err != nil {
		return models.LevelWarp{}, err
	}
	warp := models.LevelWarp{Id: id}
	for _, field := range []struct {
		key    string
		target *int
	}{
		{"OffsetX", &warp.OffsetX}, {"OffsetY", &warp.OffsetY},
		{"ExitWalkX", &warp.ExitWalkX}, {"ExitWalkY", &warp.ExitWalkY},
	} {
		value, err := caveRecordNumber(row, field.key)
		if err != nil {
			return models.LevelWarp{}, err
		}
		*field.target = value
	}
	if row["NoInteract"] != "" {
		warp.NoInteract, err = caveRecordNumber(row, "NoInteract")
		if err != nil {
			return models.LevelWarp{}, err
		}
	}
	if warp.NoInteract != 0 {
		return models.LevelWarp{}, fmt.Errorf("cave record %d forbids interaction", id)
	}
	return warp, nil
}
