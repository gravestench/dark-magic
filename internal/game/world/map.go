// Package world decodes deterministic gameplay facts from DS1 stamps and DT1
// tilesets. It does not own presentation textures or native renderer state.
package world

import (
	"fmt"
	"io/fs"
	"sync"

	"github.com/gravestench/ds1"
)

// SubtilesPerTile is the fixed collision resolution encoded by DT1 tiles.
const SubtilesPerTile = 5

const (
	// Tile pixel dimensions describe isometric presentation space only; collision
	// and navigation use subtile coordinates.
	TilePixelWidth  = 160
	TilePixelHeight = 80
	PreviewMargin   = 160
)

const (
	// Object type values are authored DS1 record identities.
	ObjectTypeDynamic int32 = 1
	ObjectTypeStatic  int32 = 2
)

// Flags is the gameplay-relevant union of DT1 subtile collision bits.
type Flags struct {
	BlockWalk, BlockLOS, BlockJump, BlockPlayerWalk, BlockLight bool
}

// Blocked reports whether a player-sized point cannot walk through this
// subtile. BlockWalk is shared terrain collision; BlockPlayerWalk is the
// additional player-specific restriction encoded by DT1.
func (f Flags) Blocked() bool { return f.BlockWalk || f.BlockPlayerWalk }

// Object preserves authored DS1 placement plus optional catalog resolution.
// Loading identifies objects; authoritative systems decide whether to spawn them.
type Object struct {
	Type, ID, X, Y, Flags int32
	ObjectID              int
	Class                 string
	Description           string
	Resolved              bool
}

// ObjectResolver supplies act-local recovered identity joins without coupling
// world decoding to a concrete catalog generation.
type ObjectResolver interface {
	ResolveStaticObject(act, id int) (objectID int, description string, found bool)
	ResolveDynamicObject(act, id int) (class string, found bool)
}

// Map is an immutable decoded stamp in tile/subtile coordinates.
type Map struct {
	WidthTiles, HeightTiles       int
	WidthSubtiles, HeightSubtiles int
	Act                           int
	Objects                       []Object
	Tiles                         []TilePlacement
	SpecialTiles                  []SpecialTile
	flags                         []Flags
	selectorOnce                  sync.Once
	selector                      *Selector
	selectorErr                   error
}

// SpecialTile is an authored DS1 orientation-10/11 cell. These cells carry
// level-transition and other map semantics even when their visual tile is
// hidden. MainIndex and SubIndex are intentionally preserved: orientation
// alone cannot distinguish ordinary exits from map-entry, town-entry, corpse,
// or town-portal markers.
//
// This is a raw authored fact, not a resolved destination. Joining it to level
// and LvlWarp records belongs to map generation, not DS1 decoding.
type SpecialTile struct {
	X, Y                             int
	Orientation, MainIndex, SubIndex int32
	Hidden                           bool
}

// TileLayer is the stable global-pass order used by legacy map presentation.
// It is data, not a renderer Z value, so adapters remain free to batch chunks.
type TileLayer uint8

const (
	LayerFloor TileLayer = iota
	LayerLowerWall
	LayerShadow
	LayerUpperWall
	LayerRoof
)

func (l TileLayer) String() string {
	switch l {
	case LayerFloor:
		return "floor"
	case LayerLowerWall:
		return "lower-wall"
	case LayerShadow:
		return "shadow"
	case LayerUpperWall:
		return "upper-wall"
	case LayerRoof:
		return "roof"
	default:
		return "unknown"
	}
}

// TilePlacement is one deterministic DS1-cell-to-DT1 selection. Presentation
// decodes Reference.Path/Index; collision consumes only copied metadata.
type TilePlacement struct {
	X, Y      int
	Layer     TileLayer
	Identity  TileIdentity
	Reference TileReference
}

// Load joins one DS1 stamp with its DT1 collision definitions. It decodes no
// renderer textures and performs no entity spawning.
func Load(source fs.FS, stampPath string, tilePaths []string, resolvers ...ObjectResolver) (*Map, error) {
	// Preserve the public loader's useful error ordering: the requested stamp is
	// the primary resource, so report a missing DS1 before inspecting its DT1
	// dependencies. Generated-zone materialization uses loadStamp directly after
	// catalog lookup and therefore does not pay this existence probe per room.
	if _, err := fs.Stat(source, stampPath); err != nil {
		return nil, fmt.Errorf("world: open %q: %w", stampPath, err)
	}
	catalog, err := LoadTileCatalog(source, tilePaths)
	if err != nil {
		return nil, err
	}
	var resolver ObjectResolver
	if len(resolvers) > 0 {
		resolver = resolvers[0]
	}
	return loadStamp(source, stampPath, catalog, resolver)
}

func loadStamp(source fs.FS, stampPath string, catalog *TileCatalog, resolver ObjectResolver) (*Map, error) {
	stampFile, err := source.Open(stampPath)
	if err != nil {
		return nil, fmt.Errorf("world: open %q: %w", stampPath, err)
	}
	stamp, err := ds1.FromReader(stampFile)
	closeErr := stampFile.Close()
	if err != nil {
		return nil, fmt.Errorf("world: decode DS1 %q: %w", stampPath, err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("world: close DS1 %q: %w", stampPath, closeErr)
	}
	result := &Map{
		WidthTiles: int(stamp.Width), HeightTiles: int(stamp.Height), Act: int(stamp.Act),
		WidthSubtiles: int(stamp.Width) * SubtilesPerTile, HeightSubtiles: int(stamp.Height) * SubtilesPerTile,
	}
	result.flags = make([]Flags, result.WidthSubtiles*result.HeightSubtiles)
	for _, object := range stamp.Objects {
		decoded := resolveObject(result.Act, object.Type, object.ID, object.X, object.Y, object.Flags, resolver)
		result.Objects = append(result.Objects, decoded)
	}
	for tileY, row := range stamp.Tiles {
		for tileX, record := range row {
			for _, floor := range record.Floors {
				if !floor.Hidden && floor.Prop1 != 0 {
					result.addTile(catalog, tileX, tileY, LayerFloor, TileIdentity{MainIndex: int32(floor.Style), SubIndex: int32(floor.Sequence)})
				}
			}
			for _, wall := range record.Walls {
				identity := TileIdentity{Orientation: int32(wall.Type), MainIndex: int32(wall.Style), SubIndex: int32(wall.Sequence)}
				if identity.Orientation == 10 || identity.Orientation == 11 {
					result.SpecialTiles = append(result.SpecialTiles, SpecialTile{
						X: tileX, Y: tileY,
						Orientation: identity.Orientation, MainIndex: identity.MainIndex, SubIndex: identity.SubIndex,
						Hidden: wall.Hidden,
					})
				}
				if !wall.Hidden && wall.Prop1 != 0 {
					layer := LayerUpperWall
					if identity.Orientation >= 16 && identity.Orientation <= 19 {
						layer = LayerLowerWall
					} else if identity.Orientation == 15 {
						layer = LayerRoof
					}
					result.addTile(catalog, tileX, tileY, layer, identity)
					// A north corner is authored as one DS1 orientation but drawn
					// from paired type-3 and type-4 DT1 records on one baseline.
					if identity.Orientation == 3 {
						companion := identity
						companion.Orientation = 4
						result.addTile(catalog, tileX, tileY, layer, companion)
					}
				}
			}
			for _, shadow := range record.Shadows {
				if !shadow.Hidden && shadow.Prop1 != 0 {
					result.addTile(catalog, tileX, tileY, LayerShadow, TileIdentity{Orientation: 13, MainIndex: int32(shadow.Style), SubIndex: int32(shadow.Sequence)})
				}
			}
		}
	}
	return result, nil
}

func resolveObject(act int, objectType, id, x, y, flags int32, resolver ObjectResolver) Object {
	result := Object{Type: objectType, ID: id, X: x, Y: y, Flags: flags}
	if resolver == nil {
		return result
	}
	switch objectType {
	case ObjectTypeStatic:
		result.ObjectID, result.Description, result.Resolved = resolver.ResolveStaticObject(act, int(id))
	case ObjectTypeDynamic:
		result.Class, result.Resolved = resolver.ResolveDynamicObject(act, int(id))
	}
	return result
}

func (m *Map) addTile(catalog *TileCatalog, tileX, tileY int, layer TileLayer, identity TileIdentity) {
	reference, found := catalog.Select(identity, tileX, tileY, 0)
	if !found {
		return
	}
	m.Tiles = append(m.Tiles, TilePlacement{X: tileX, Y: tileY, Layer: layer, Identity: identity, Reference: reference})
	if layer != LayerShadow && layer != LayerRoof {
		m.apply(tileX, tileY, reference)
	}
}

func (m *Map) apply(tileX, tileY int, tile TileReference) {
	for index, source := range tile.SubTileFlags {
		x := tileX*SubtilesPerTile + index%SubtilesPerTile
		// DT1 stores its five subtile rows bottom-to-top. World coordinates grow
		// top-to-bottom, so reading the bytes as ordinary row-major data mirrors
		// collision vertically inside every tile.
		y := tileY*SubtilesPerTile + (SubtilesPerTile - 1 - index/SubtilesPerTile)
		if x < 0 || y < 0 || x >= m.WidthSubtiles || y >= m.HeightSubtiles {
			continue
		}
		target := &m.flags[y*m.WidthSubtiles+x]
		target.BlockWalk = target.BlockWalk || source.BlockWalk
		target.BlockLOS = target.BlockLOS || source.BlockLOS
		target.BlockJump = target.BlockJump || source.BlockJump
		target.BlockPlayerWalk = target.BlockPlayerWalk || source.BlockPlayerWalk
		target.BlockLight = target.BlockLight || source.BlockLight
	}
}

func (m *Map) FlagsAt(x, y int) (Flags, bool) {
	if x < 0 || y < 0 || x >= m.WidthSubtiles || y >= m.HeightSubtiles {
		return Flags{}, false
	}
	index := y*m.WidthSubtiles + x
	if index < 0 || index >= len(m.flags) {
		return Flags{}, false
	}
	return m.flags[index], true
}

// OpenPointNearCenter returns a deterministic unblocked subtile for fixture
// entry. The expanding square search prevents an authored wall at the exact
// center from trapping a newly admitted player.
func (m *Map) OpenPointNearCenter() (float64, float64, bool) {
	return m.openPointNear(m.WidthSubtiles/2, m.HeightSubtiles/2, 0)
}

// OpenPointNearSubtile is the exported collision-aware anchor resolver used by
// generated warp/seam assembly. Inputs and outputs are gameplay subtiles.
func (m *Map) OpenPointNearSubtile(x, y float64) (float64, float64, bool) {
	return m.openPointNear(CollisionCell(x), CollisionCell(y), 0)
}

// ActOneTownEntry returns a deterministic open point near the authored Rogue
// Encampment bonfire. The bonfire is the stable town landmark shared by all
// four cardinal layouts; using it avoids tying session entry to screen pixels
// or to whichever shape happens to surround the map's numeric center.
func (m *Map) ActOneTownEntry() (float64, float64, bool) {
	for _, object := range m.Objects {
		if object.Type == ObjectTypeStatic && object.ID == 2 { // act-local RogueBonfire
			return m.openPointNear(int(object.X), int(object.Y), 4)
		}
	}
	return 0, 0, false
}

func (m *Map) openPointNear(centerX, centerY, firstRadius int) (float64, float64, bool) {
	limit := max(m.WidthSubtiles, m.HeightSubtiles)
	for radius := firstRadius; radius <= limit; radius++ {
		for y := centerY - radius; y <= centerY+radius; y++ {
			for x := centerX - radius; x <= centerX+radius; x++ {
				if radius > 0 && x != centerX-radius && x != centerX+radius && y != centerY-radius && y != centerY+radius {
					continue
				}
				flags, inside := m.FlagsAt(x, y)
				if inside && !flags.Blocked() {
					return float64(x), float64(y), true
				}
			}
		}
	}
	return 0, 0, false
}

// SubtileToPixel projects continuous DS1 subtile coordinates into the same
// image-space diamond centers used by TexturedDS1Preview.
func (m *Map) SubtileToPixel(x, y float64) (float64, float64) {
	point := m.Coordinates().SubtileToWorldPixel(Point{X: x, Y: y})
	return point.X, point.Y
}

// PixelToSubtile reverses SubtileToPixel. Fractional values are preserved so
// callers can choose their own collision sampling policy.
func (m *Map) PixelToSubtile(x, y float64) (float64, float64) {
	point := m.Coordinates().WorldPixelToSubtile(Point{X: x, Y: y})
	return point.X, point.Y
}
