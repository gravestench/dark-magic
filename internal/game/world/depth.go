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
	depthScale = 10
	floorDepth = -1_000_000
)

// EntityDepth sorts things that stand in the world by their projected
// baseline. On an isometric map, x+y increases toward the viewer.
func EntityDepth(subtileX, subtileY float64) float64 {
	return (subtileX + subtileY) * depthScale
}

// TileDepth gives an authored tile a baseline compatible with entity depth.
// One tile advances five subtiles; the extra tile-sized step is the near edge
// of the diamond where standing entities cross in front of the tile.
func TileDepth(layer TileLayer, tileX, tileY int) int {
	if layer == LayerFloor {
		return floorDepth
	}
	baseline := (tileX + tileY + 1) * SubtilesPerTile * depthScale
	switch layer {
	case LayerShadow:
		return baseline - 2
	case LayerLowerWall:
		return baseline - 1
	case LayerUpperWall:
		return baseline + 1
	case LayerRoof:
		return baseline + 2
	default:
		return baseline
	}
}
