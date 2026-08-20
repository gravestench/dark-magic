package world

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/gravestench/dark-magic/internal/game/worldgen"
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
type ZonePostprocessor func(*Map, *worldgen.Zone, []*TileCatalog) error

// Materializer incrementally joins generated stamp recipes into one immutable world map. Callers may run Step on a
// worker goroutine because it performs no native renderer work. The final Step error must always be honored: completion
// is recorded after placement and before final recipe/postprocessor work, preserving the established lifecycle.
type Materializer struct {
	source       fs.FS
	zone         *worldgen.Zone
	resolver     ObjectResolver
	stamps       []worldgen.Stamp
	load         stampLoader
	postprocess  ZonePostprocessor
	catalogs     map[string]*TileCatalog
	catalogOrder []*TileCatalog
	assembled    *Map
	next         int
	done         bool
}

// SetPostprocessor installs a mod-selected final recipe interpretation step.
// The engine owns decoded map assembly; a mod may translate its own recipe
// vocabulary into tile choices after every authored stamp is present.
func (materializer *Materializer) SetPostprocessor(postprocess ZonePostprocessor) {
	materializer.postprocess = postprocess
}

// NewMaterializer allocates the final map at generated zone bounds while Result withholds it until the final stamp has
// been placed. Callers must still honor each Step error, including errors from final recipe transformations.
func NewMaterializer(source fs.FS, zone *worldgen.Zone, resolvers ...ObjectResolver) (*Materializer, error) {
	if source == nil || zone == nil {
		return nil, errors.New("world: materializer requires an asset source and zone")
	}

	bounds := zone.Bounds()
	assembled := &Map{
		WidthTiles:     bounds.Width,
		HeightTiles:    bounds.Height,
		WidthSubtiles:  bounds.Width * SubtilesPerTile,
		HeightSubtiles: bounds.Height * SubtilesPerTile,
		Act:            int(zone.Request().Act),
	}
	assembled.flags = make([]Flags, assembled.WidthSubtiles*assembled.HeightSubtiles)

	result := &Materializer{
		source:    source,
		zone:      zone,
		stamps:    zone.Stamps(),
		catalogs:  make(map[string]*TileCatalog),
		assembled: assembled,
	}
	if len(resolvers) > 0 {
		// The variadic signature preserves Load's API convention; only the first resolver has historically been active.
		result.resolver = resolvers[0]
	}

	if len(result.stamps) == 0 {
		result.done = true
	}

	return result, nil
}

// Progress snapshots logical stamp progress without exposing the mutable assembled map. CurrentStamp and CurrentRole
// identify the next unit of work, so loading UI can describe what a subsequent Step will attempt.
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

	decoded, err := materializer.decodeMaterializationStamp(stamp)
	if err != nil {
		return fmt.Errorf("world: materialize stamp %d (%s): %w", stamp.ID, stamp.Role, err)
	}

	// DS1 dimensions include a shared terminal cell. A 24x24 LvlPrest room is
	// therefore commonly decoded as 25x25. The generated room footprint owns the
	// authoritative extent; the terminal row/column must not overlap its neighbor.
	if err := validateDecodedStampSize(stamp, decoded); err != nil {
		return err
	}

	if err := materializer.assembled.place(
		decoded,
		stamp.X,
		stamp.Y,
		stamp.Width,
		stamp.Height,
		stamp.Overlay,
	); err != nil {
		return fmt.Errorf("world: place stamp %d: %w", stamp.ID, err)
	}

	materializer.next++

	materializer.done = materializer.next == len(materializer.stamps)
	if !materializer.done {
		return nil
	}

	return materializer.finish()
}

// decodeMaterializationStamp honors the injected loader used by tests and tools, otherwise reuses DT1 catalogs across
// stamps. Resolver forwarding preserves the exact optional-argument shape expected by custom loaders.
func (materializer *Materializer) decodeMaterializationStamp(stamp worldgen.Stamp) (*Map, error) {
	if materializer.load == nil {
		return materializer.loadCached(stamp)
	}

	var resolvers []ObjectResolver
	if materializer.resolver != nil {
		resolvers = append(resolvers, materializer.resolver)
	}

	return materializer.load(materializer.source, stamp.DS1Path, stamp.TilePaths, resolvers...)
}

// validateDecodedStampSize allows DS1's shared terminal row and column but rejects any larger mismatch before mutating
// assembled state. This keeps a failed Step retryable at the same progress index.
func validateDecodedStampSize(stamp worldgen.Stamp, decoded *Map) error {
	if decoded.WidthTiles <= stamp.Width+1 && decoded.HeightTiles <= stamp.Height+1 {
		return nil
	}

	return fmt.Errorf(
		"world: stamp %d decoded size %dx%d exceeds recipe %dx%d plus shared edge",
		stamp.ID,
		decoded.WidthTiles,
		decoded.HeightTiles,
		stamp.Width,
		stamp.Height,
	)
}

// finish applies engine-owned floor recipes before the optional mod postprocessor. The ordering ensures mods see final
// collision from recipe floors and can intentionally replace it, matching the established materialization contract.
func (materializer *Materializer) finish() error {
	if err := materializeRecipeFloors(
		materializer.assembled,
		materializer.zone,
		materializer.catalogOrder,
	); err != nil {
		return err
	}

	if materializer.postprocess == nil {
		return nil
	}

	// The slice copy prevents a postprocessor from reordering the materializer's cache precedence for later inspection.
	catalogs := append([]*TileCatalog(nil), materializer.catalogOrder...)

	return materializer.postprocess(materializer.assembled, materializer.zone, catalogs)
}

// materializeRecipeFloors applies opaque floor identities selected by a mod's
// admitted world recipe. The engine performs only catalog lookup and placement;
// it does not know why a route cell uses a particular DT1 identity.
func materializeRecipeFloors(world *Map, zone *worldgen.Zone, catalogs []*TileCatalog) error {
	changed := false

	for _, tile := range zone.Paths() {
		if tile.MainIndex == 0 && tile.SubIndex == 0 {
			continue
		}

		identity := TileIdentity{MainIndex: tile.MainIndex, SubIndex: tile.SubIndex}

		var (
			reference TileReference
			found     bool
		)
		for _, catalog := range catalogs {
			if reference, found = catalog.Select(identity, tile.X, tile.Y, 0); found {
				break
			}
		}

		if !found {
			return fmt.Errorf(
				"world: recipe floor (%d,%d) is unavailable at %d,%d",
				identity.MainIndex,
				identity.SubIndex,
				tile.X,
				tile.Y,
			)
		}

		world.ReplaceFloor(tile.X, tile.Y, identity, reference)

		changed = true
	}

	if changed {
		world.RebuildFlags()
	}

	return nil
}

// loadCached shares a TileCatalog for stamps with the same ordered path list. The NUL separator makes the key
// unambiguous without sorting, because authored path order controls deterministic tile candidate order.
func (materializer *Materializer) loadCached(stamp worldgen.Stamp) (*Map, error) {
	key := strings.Join(stamp.TilePaths, "\x00")

	catalog := materializer.catalogs[key]
	if catalog == nil {
		var err error

		catalog, err = LoadTileCatalog(materializer.source, stamp.TilePaths)
		if err != nil {
			return nil, err
		}

		materializer.catalogs[key] = catalog
		materializer.catalogOrder = append(materializer.catalogOrder, catalog)
	}

	return loadStamp(materializer.source, stamp.DS1Path, catalog, materializer.resolver)
}

// Result returns assembled state after the last stamp advances completion. The caller must honor the final Step result
// because the historical lifecycle records completion immediately before final recipe and postprocessor work.
func (materializer *Materializer) Result() (*Map, error) {
	if !materializer.done {
		return nil, ErrMaterializationIncomplete
	}

	return materializer.assembled, nil
}

// place clips one decoded stamp to its generated footprint and shifts every fact into zone coordinates. Overlay mode
// replaces matching tile layers and special cells while objects and collision retain established append/copy rules.
func (target *Map) place(source *Map, offsetX, offsetY, width, height int, overlay bool) error {
	if !target.canPlace(source, offsetX, offsetY, width, height) {
		return errors.New("stamp lies outside zone bounds")
	}

	target.placeTiles(source, offsetX, offsetY, width, height, overlay)
	target.placeSpecialTiles(source, offsetX, offsetY, width, height, overlay)
	target.placeObjects(source, offsetX, offsetY, width, height)
	target.placeCollision(source, offsetX, offsetY, width, height)

	return nil
}

// canPlace validates every geometry relationship before place mutates the target. Rejecting invalid dimensions up front
// prevents partial tile/object publication from a malformed generated recipe.
func (target *Map) canPlace(source *Map, offsetX, offsetY, width, height int) bool {
	return source != nil &&
		offsetX >= 0 && offsetY >= 0 &&
		width > 0 && height > 0 &&
		width <= source.WidthTiles && height <= source.HeightTiles &&
		offsetX+width <= target.WidthTiles && offsetY+height <= target.HeightTiles
}

// placeTiles clips authored placements, shifts copies into zone coordinates, and replaces only exact cell/layer keys
// for overlays. Preserving append order keeps global presentation passes deterministic.
func (target *Map) placeTiles(source *Map, offsetX, offsetY, width, height int, overlay bool) {
	placements := make([]TilePlacement, 0, len(source.Tiles))

	replaced := make(map[[3]int]struct{}, len(source.Tiles))
	for _, tile := range source.Tiles {
		if tile.X < 0 || tile.Y < 0 || tile.X >= width || tile.Y >= height {
			continue
		}

		tile.X += offsetX

		tile.Y += offsetY
		if overlay {
			replaced[[3]int{tile.X, tile.Y, int(tile.Layer)}] = struct{}{}
		}

		placements = append(placements, tile)
	}

	if overlay {
		kept := target.Tiles[:0]
		for _, tile := range target.Tiles {
			if _, replace := replaced[[3]int{tile.X, tile.Y, int(tile.Layer)}]; !replace {
				kept = append(kept, tile)
			}
		}

		target.Tiles = kept
	}

	target.Tiles = append(target.Tiles, placements...)
}

// placeSpecialTiles applies overlay replacement by cell rather than by orientation. A generated overlay owns all exit
// semantics at each covered cell, including hidden markers.
func (target *Map) placeSpecialTiles(source *Map, offsetX, offsetY, width, height int, overlay bool) {
	specials := make([]SpecialTile, 0, len(source.SpecialTiles))

	replacedCells := make(map[[2]int]struct{}, len(source.SpecialTiles))
	for _, special := range source.SpecialTiles {
		if special.X < 0 || special.Y < 0 || special.X >= width || special.Y >= height {
			continue
		}

		special.X += offsetX

		special.Y += offsetY
		if overlay {
			replacedCells[[2]int{special.X, special.Y}] = struct{}{}
		}

		specials = append(specials, special)
	}

	if overlay {
		kept := target.SpecialTiles[:0]
		for _, special := range target.SpecialTiles {
			if _, replace := replacedCells[[2]int{special.X, special.Y}]; !replace {
				kept = append(kept, special)
			}
		}

		target.SpecialTiles = kept
	}

	target.SpecialTiles = append(target.SpecialTiles, specials...)
}

// placeObjects clips in subtile space and shifts copies by the tile-to-subtile scale. Objects append even for overlays,
// preserving the historical rule that overlays replace terrain semantics but do not erase authored entity records.
func (target *Map) placeObjects(source *Map, offsetX, offsetY, width, height int) {
	objectOffsetX := int32(offsetX * SubtilesPerTile)

	objectOffsetY := int32(offsetY * SubtilesPerTile)
	for _, object := range source.Objects {
		outside := object.X < 0 || object.Y < 0 ||
			int(object.X) >= width*SubtilesPerTile || int(object.Y) >= height*SubtilesPerTile
		if outside {
			continue
		}

		object.X += objectOffsetX
		object.Y += objectOffsetY
		target.Objects = append(target.Objects, object)
	}
}

// placeCollision copies the clipped subtile rectangle directly into final storage. Unlike layered tile metadata,
// collision is already the source stamp's resolved union and overlays therefore replace every covered cell.
func (target *Map) placeCollision(source *Map, offsetX, offsetY, width, height int) {
	for y := 0; y < height*SubtilesPerTile; y++ {
		for x := 0; x < width*SubtilesPerTile; x++ {
			flags, _ := source.FlagsAt(x, y)
			targetX := offsetX*SubtilesPerTile + x
			targetY := offsetY*SubtilesPerTile + y
			target.flags[targetY*target.WidthSubtiles+targetX] = flags
		}
	}
}
