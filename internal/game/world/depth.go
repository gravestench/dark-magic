package world

// DepthKind identifies the presentation role of a world-space fact. The value
// returned by the helpers is an ordering key, not authoritative simulation
// state and not a renderer-specific layer number.
type DepthKind uint8

const (
	DepthFloor DepthKind = iota
	DepthShadow
	DepthLowerWall
	DepthEntity
	DepthUpperWall
	DepthRoof
)

const (
	depthTileStride = 100
	lowerWallDepth  = -1_000_002
	floorDepth      = -1_000_001
	shadowDepth     = -1_000_000
	roofDepth       = 1_000_000
)

// EntityDepth reproduces the legacy tile pass split used by OpenDiablo2 and
// Riiablo. An entity in subtile row/column zero belongs behind that tile's
// upper wall. Once both local coordinates enter the tile interior, it belongs
// in front. The small local sum keeps entities in one pass ordered by their
// feet without allowing them to escape that pass band.
func EntityDepth(subtileX, subtileY float64) float64 {
	cellX, cellY := CollisionCell(subtileX), CollisionCell(subtileY)
	tileX, tileY := cellX/SubtilesPerTile, cellY/SubtilesPerTile
	localX, localY := cellX%SubtilesPerTile, cellY%SubtilesPerTile
	baseline := (tileX + tileY + 1) * depthTileStride

	localOrder := (localX + localY) * 2
	if localX == 0 || localY == 0 {
		return float64(baseline - 20 + localOrder)
	}

	return float64(baseline + 1 + localOrder)
}

// TileDepth maps authored layers onto the legacy background/middleground/
// foreground passes. Only upper walls interleave with standing entities.
func TileDepth(layer TileLayer, tileX, tileY int) int {
	switch layer {
	case LayerLowerWall:
		return lowerWallDepth
	case LayerFloor:
		return floorDepth
	case LayerShadow:
		return shadowDepth
	case LayerUpperWall:
		return (tileX + tileY + 1) * depthTileStride
	case LayerRoof:
		return roofDepth
	default:
		return (tileX + tileY + 1) * depthTileStride
	}
}
