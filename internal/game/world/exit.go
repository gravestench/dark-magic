package world

import "sort"

// ExitAnchor is an authored DS1 level-exit tile expressed at its gameplay
// subtile center. Orientations 10 and 11 are the legacy left/right exit tile
// semantics; hidden appearance does not erase their gameplay meaning.
type ExitAnchor struct {
	X, Y        float64
	Orientation int32
}

// AuthoredExitAnchors returns a stable copy of semantic exit anchors. It never
// guesses from transparent pixels, map bounds, or apparently open collision.
func (m *Map) AuthoredExitAnchors() []ExitAnchor {
	if m == nil {
		return nil
	}
	result := make([]ExitAnchor, 0)
	for _, tile := range m.Tiles {
		if tile.Identity.Orientation != 10 && tile.Identity.Orientation != 11 {
			continue
		}
		result = append(result, ExitAnchor{X: float64(tile.X*SubtilesPerTile) + 2.5, Y: float64(tile.Y*SubtilesPerTile) + 2.5, Orientation: tile.Identity.Orientation})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Y == result[j].Y {
			return result[i].X < result[j].X
		}
		return result[i].Y < result[j].Y
	})
	return result
}

// OpenPointNearExit moves an authored exit anchor to the nearest walkable
// subtile while retaining deterministic scan order.
func (m *Map) OpenPointNearExit(anchor ExitAnchor) (float64, float64, bool) {
	return m.openPointNear(CollisionCell(anchor.X), CollisionCell(anchor.Y), 0)
}
