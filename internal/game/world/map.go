// Package world decodes deterministic gameplay facts from DS1 stamps and DT1
// tilesets. It does not own presentation textures or native renderer state.
package world

import (
	"fmt"
	"math"
	"sync"
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

// NewOpenMap creates an empty traversable collision map. It is useful to
// generic tools, generated-world assembly, and tests that need geometry
// without first decoding a DS1 stamp.
func NewOpenMap(widthSubtiles, heightSubtiles int) (*Map, error) {
	if widthSubtiles <= 0 || heightSubtiles <= 0 {
		return nil, fmt.Errorf("world: open map dimensions must be positive")
	}

	return &Map{
		WidthSubtiles: widthSubtiles, HeightSubtiles: heightSubtiles,
		WidthTiles: widthSubtiles / SubtilesPerTile, HeightTiles: heightSubtiles / SubtilesPerTile,
		flags: make([]Flags, widthSubtiles*heightSubtiles),
	}, nil
}

// Blocked reports whether a player-sized point cannot walk through this
// subtile. BlockWalk is shared terrain collision; BlockPlayerWalk is the
// additional player-specific restriction encoded by DT1.
func (f Flags) Blocked() bool { return f.BlockWalk || f.BlockPlayerWalk }

// ReplaceFloor swaps one floor placement while preserving every other layer.
// Which identity to choose is policy supplied by the caller.
func (m *Map) ReplaceFloor(x, y int, identity TileIdentity, reference TileReference) {
	for index := range m.Tiles {
		if m.Tiles[index].X == x && m.Tiles[index].Y == y && m.Tiles[index].Layer == LayerFloor {
			m.Tiles[index].Identity = identity
			m.Tiles[index].Reference = reference

			return
		}
	}

	m.Tiles = append(m.Tiles, TilePlacement{X: x, Y: y, Layer: LayerFloor, Identity: identity, Reference: reference})
}

// RebuildFlags recomputes collision after a trusted assembly postprocessor
// changes tile references.
func (m *Map) RebuildFlags() {
	m.flags = make([]Flags, m.WidthSubtiles*m.HeightSubtiles)
	for _, tile := range m.Tiles {
		if tile.Layer != LayerShadow && tile.Layer != LayerRoof && tile.Reference.Path != "" {
			m.apply(tile.X, tile.Y, tile.Reference)
		}
	}
}

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

// String returns the stable presentation name used by diagnostics and adapters. Unknown values remain explicit instead
// of leaking numeric layer values into user-facing output.
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

// apply merges one selected DT1 tile into authoritative collision. Multiple visible layers contribute flags with OR so
// a later decorative layer can never erase a blocker established by an earlier layer.
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

// FlagsAt returns collision for an integer subtile and rejects both geometric and backing-slice overflow. The second
// check protects partially constructed maps used by diagnostics and tests from panicking.
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

// OpenPointNearSubtileForRadius resolves an anchor whose complete circular
// footprint is traversable. Relocation destinations must use this rather than
// validating only their center cell, or a player can arrive unable to take the
// first locomotion step.
func (m *Map) OpenPointNearSubtileForRadius(x, y, radius float64) (float64, float64, bool) {
	if radius < 0 || math.IsNaN(radius) || math.IsInf(radius, 0) {
		return 0, 0, false
	}

	return m.openFootprintNear(CollisionCell(x), CollisionCell(y), 0, radius)
}

// OpenPointNear resolves a traversable point at or outside a caller-selected
// radius. The caller owns why that inset is appropriate.
func (m *Map) OpenPointNear(x, y float64, firstRadius int) (float64, float64, bool) {
	if firstRadius < 0 {
		return 0, 0, false
	}

	return m.openPointNear(CollisionCell(x), CollisionCell(y), firstRadius)
}

// openPointNear adapts point-sized callers to the shared footprint search without duplicating its deterministic order.
func (m *Map) openPointNear(centerX, centerY, firstRadius int) (float64, float64, bool) {
	return m.openFootprintNear(centerX, centerY, firstRadius, 0)
}

// openFootprintNear scans expanding square perimeters from top-left to bottom-right. That authored-coordinate order is
// part of deterministic spawn and relocation selection, so changing it can alter replay outcomes.
func (m *Map) openFootprintNear(
	centerX, centerY, firstRadius int,
	footprintRadius float64,
) (float64, float64, bool) {
	limit := max(m.WidthSubtiles, m.HeightSubtiles)
	for radius := firstRadius; radius <= limit; radius++ {
		for y := centerY - radius; y <= centerY+radius; y++ {
			for x := centerX - radius; x <= centerX+radius; x++ {
				insideSquare := x != centerX-radius && x != centerX+radius && y != centerY-radius && y != centerY+radius
				if radius > 0 && insideSquare {
					continue
				}

				if m.walkableCell(navCell{x: x, y: y}, footprintRadius) {
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
