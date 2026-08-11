package world

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/mapgen"
)

var (
	ErrMaterializationIncomplete = errors.New("world: zone materialization is incomplete")
	ErrMaterializationComplete   = errors.New("world: zone materialization is already complete")
)

// MaterializationProgress is safe to show in loading UI. It describes decoded
// gameplay stamps, not renderer uploads or texture residency.
type MaterializationProgress struct {
	Completed, Total int
	CurrentStamp     uint32
	CurrentRole      string
}

type stampLoader func(fs.FS, string, []string, ...ObjectResolver) (*Map, error)

// Materializer incrementally joins generated stamp recipes into one immutable
// world map. Callers may run Step on a worker goroutine because it performs no
// native renderer work. Result becomes visible only after all steps succeed.
type Materializer struct {
	source    fs.FS
	zone      *mapgen.Zone
	resolver  ObjectResolver
	stamps    []mapgen.Stamp
	load      stampLoader
	catalogs  map[string]*TileCatalog
	assembled *Map
	next      int
	done      bool
}

func NewMaterializer(source fs.FS, zone *mapgen.Zone, resolvers ...ObjectResolver) (*Materializer, error) {
	if source == nil || zone == nil {
		return nil, errors.New("world: materializer requires an asset source and zone")
	}
	bounds := zone.Bounds()
	assembled := &Map{
		WidthTiles: bounds.Width, HeightTiles: bounds.Height,
		WidthSubtiles: bounds.Width * SubtilesPerTile, HeightSubtiles: bounds.Height * SubtilesPerTile,
		Act: int(zone.Request().Act),
	}
	assembled.flags = make([]Flags, assembled.WidthSubtiles*assembled.HeightSubtiles)
	result := &Materializer{source: source, zone: zone, stamps: zone.Stamps(), catalogs: make(map[string]*TileCatalog), assembled: assembled}
	if len(resolvers) > 0 {
		result.resolver = resolvers[0]
	}
	if len(result.stamps) == 0 {
		result.done = true
	}
	return result, nil
}

func (materializer *Materializer) Progress() MaterializationProgress {
	progress := MaterializationProgress{Completed: materializer.next, Total: len(materializer.stamps)}
	if materializer.next < len(materializer.stamps) {
		progress.CurrentStamp = materializer.stamps[materializer.next].ID
		progress.CurrentRole = materializer.stamps[materializer.next].Role
	}
	return progress
}

// Step decodes and joins exactly one authored stamp. Small steps keep loading
// responsive and make it straightforward for a coordinator to report progress.
func (materializer *Materializer) Step(ctx context.Context) error {
	if materializer.done {
		return ErrMaterializationComplete
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	stamp := materializer.stamps[materializer.next]
	var resolvers []ObjectResolver
	if materializer.resolver != nil {
		resolvers = append(resolvers, materializer.resolver)
	}
	var decoded *Map
	var err error
	if materializer.load != nil {
		decoded, err = materializer.load(materializer.source, stamp.DS1Path, stamp.TilePaths, resolvers...)
	} else {
		decoded, err = materializer.loadCached(stamp)
	}
	if err != nil {
		return fmt.Errorf("world: materialize stamp %d (%s): %w", stamp.ID, stamp.Role, err)
	}
	// DS1 dimensions include a shared terminal cell. A 24x24 LvlPrest room is
	// therefore commonly decoded as 25x25. The generated room footprint owns the
	// authoritative extent; the terminal row/column must not overlap its neighbor.
	if decoded.WidthTiles > stamp.Width+1 || decoded.HeightTiles > stamp.Height+1 {
		return fmt.Errorf("world: stamp %d decoded size %dx%d exceeds recipe %dx%d plus shared edge", stamp.ID, decoded.WidthTiles, decoded.HeightTiles, stamp.Width, stamp.Height)
	}
	if err := materializer.assembled.place(decoded, stamp.X, stamp.Y, stamp.Width, stamp.Height); err != nil {
		return fmt.Errorf("world: place stamp %d: %w", stamp.ID, err)
	}
	materializer.next++
	materializer.done = materializer.next == len(materializer.stamps)
	return nil
}

func (materializer *Materializer) loadCached(stamp mapgen.Stamp) (*Map, error) {
	key := strings.Join(stamp.TilePaths, "\x00")
	catalog := materializer.catalogs[key]
	if catalog == nil {
		var err error
		catalog, err = LoadTileCatalog(materializer.source, stamp.TilePaths)
		if err != nil {
			return nil, err
		}
		materializer.catalogs[key] = catalog
	}
	return loadStamp(materializer.source, stamp.DS1Path, catalog, materializer.resolver)
}

func (materializer *Materializer) Result() (*Map, error) {
	if !materializer.done {
		return nil, ErrMaterializationIncomplete
	}
	return materializer.assembled, nil
}

func (target *Map) place(source *Map, offsetX, offsetY, width, height int) error {
	if source == nil || offsetX < 0 || offsetY < 0 || width <= 0 || height <= 0 || width > source.WidthTiles || height > source.HeightTiles || offsetX+width > target.WidthTiles || offsetY+height > target.HeightTiles {
		return errors.New("stamp lies outside zone bounds")
	}
	for _, tile := range source.Tiles {
		if tile.X < 0 || tile.Y < 0 || tile.X >= width || tile.Y >= height {
			continue
		}
		tile.X += offsetX
		tile.Y += offsetY
		target.Tiles = append(target.Tiles, tile)
	}
	// DS1 object coordinates use the same 5x5 subtile grid as gameplay facts.
	objectOffsetX := int32(offsetX * SubtilesPerTile)
	objectOffsetY := int32(offsetY * SubtilesPerTile)
	for _, object := range source.Objects {
		if object.X < 0 || object.Y < 0 || int(object.X) >= width*SubtilesPerTile || int(object.Y) >= height*SubtilesPerTile {
			continue
		}
		object.X += objectOffsetX
		object.Y += objectOffsetY
		target.Objects = append(target.Objects, object)
	}
	for y := 0; y < height*SubtilesPerTile; y++ {
		for x := 0; x < width*SubtilesPerTile; x++ {
			flags, _ := source.FlagsAt(x, y)
			targetX := offsetX*SubtilesPerTile + x
			targetY := offsetY*SubtilesPerTile + y
			target.flags[targetY*target.WidthSubtiles+targetX] = flags
		}
	}
	return nil
}
