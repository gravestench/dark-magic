package world

import (
	"sort"

	"github.com/gravestench/ds1"
)

// ReplaceAuthoredCell rematerializes one DS1 tile coordinate in place. It is
// intended for editor preview authorities: runtime maps remain immutable after
// ordinary loading. Collision is cleared and rebuilt only for the cell's 5x5
// subtile footprint.
func (m *Map) ReplaceAuthoredCell(catalog *TileCatalog, x, y int, record ds1.TileRecord) {
	if m == nil || catalog == nil || x < 0 || y < 0 || x >= m.WidthTiles || y >= m.HeightTiles {
		return
	}
	tiles := m.Tiles[:0]
	for _, tile := range m.Tiles {
		if tile.X != x || tile.Y != y {
			tiles = append(tiles, tile)
		}
	}
	m.Tiles = tiles
	special := m.SpecialTiles[:0]
	for _, tile := range m.SpecialTiles {
		if tile.X != x || tile.Y != y {
			special = append(special, tile)
		}
	}
	m.SpecialTiles = special
	m.clearCellFlags(x, y)
	m.appendFloorRecords(catalog, x, y, record.Floors)
	m.appendWallRecords(catalog, x, y, record.Walls)
	m.appendShadowRecords(catalog, x, y, record.Shadows)
	// Indexing traverses placements by presentation layer, then uses this
	// row-major order for equal-depth overlap. Stable sorting retains authored
	// layer/companion order within the replaced cell.
	sort.SliceStable(m.Tiles, func(left, right int) bool {
		if m.Tiles[left].Y != m.Tiles[right].Y {
			return m.Tiles[left].Y < m.Tiles[right].Y
		}
		return m.Tiles[left].X < m.Tiles[right].X
	})
}

// clearCellFlags removes only the 5×5 subtile collision footprint owned by one DS1 cell.
// Neighboring collision data stays intact for dirty-region preview updates.
func (m *Map) clearCellFlags(tileX, tileY int) {
	for y := tileY * SubtilesPerTile; y < (tileY+1)*SubtilesPerTile; y++ {
		for x := tileX * SubtilesPerTile; x < (tileX+1)*SubtilesPerTile; x++ {
			if x >= 0 && y >= 0 && x < m.WidthSubtiles && y < m.HeightSubtiles {
				m.flags[y*m.WidthSubtiles+x] = Flags{}
			}
		}
	}
}

// PresentationSnapshot copies mutable slices into a fresh map identity. Render
// caches can index the returned value without observing a later editor stroke.
func (m *Map) PresentationSnapshot() *Map {
	if m == nil {
		return nil
	}
	return &Map{
		WidthTiles: m.WidthTiles, HeightTiles: m.HeightTiles,
		WidthSubtiles: m.WidthSubtiles, HeightSubtiles: m.HeightSubtiles,
		Act: m.Act, Objects: append([]Object(nil), m.Objects...),
		Tiles:        append([]TilePlacement(nil), m.Tiles...),
		SpecialTiles: append([]SpecialTile(nil), m.SpecialTiles...),
		flags:        append([]Flags(nil), m.flags...),
	}
}
