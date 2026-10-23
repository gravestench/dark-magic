package entryworld

import (
	"fmt"
	"math"

	gameworld "github.com/gravestench/dark-magic/internal/game/world"
	"github.com/gravestench/dark-magic/internal/game/worldgen"
	gametransition "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/transition"
)

// authoredStampEndpoint resolves only the authored exit inside the policy-selected chamber.
func authoredStampEndpoint(gameMap *gameworld.Map, zone *worldgen.Zone, stampID uint32) (gametransition.SeamEndpoint, error) {
	for _, stamp := range zone.Stamps() {
		if stamp.ID != stampID {
			continue
		}
		left, top := float64(stamp.X*gameworld.SubtilesPerTile), float64(stamp.Y*gameworld.SubtilesPerTile)
		width, height := float64(stamp.Width*gameworld.SubtilesPerTile), float64(stamp.Height*gameworld.SubtilesPerTile)
		for _, anchor := range gameMap.AuthoredExitAnchors() {
			if anchor.X < left || anchor.X >= left+width || anchor.Y < top || anchor.Y >= top+height {
				continue
			}

			// Radius-safe paths exclude isolated open pockets beside authored stairs.
			var path []gameworld.Point
			for _, offset := range [][2]float64{{0, 0}, {0.25, 0}, {-0.25, 0}, {0, 0.25}, {0, -0.25}} {
				centerX, centerY, found := gameMap.OpenPointNearSubtileForRadius(left+width*(0.5+offset[0]), top+height*(0.5+offset[1]), 1)
				if !found {
					continue
				}
				for distance := 2.0; distance <= 10; distance++ {
					path, _ = gameMap.FindPath(gameworld.PathRequest{
						Start: gameworld.Point{X: centerX, Y: centerY}, Goal: gameworld.Point{X: anchor.X, Y: anchor.Y},
						Radius: 1, StopRadius: distance,
					})
					if len(path) >= 8 {
						break
					}
				}
				if len(path) >= 8 {
					break
				}
			}
			if len(path) < 8 {
				continue
			}
			endpoint, arrival := path[len(path)-1], path[len(path)-8]
			if math.Hypot(arrival.X-endpoint.X, arrival.Y-endpoint.Y) <= 3 {
				continue
			}
			return gametransition.SeamEndpoint{LevelID: zone.Request().LevelID, X: endpoint.X, Y: endpoint.Y,
				ArrivalX: arrival.X, ArrivalY: arrival.Y, Width: float64(gameMap.WidthSubtiles), Height: float64(gameMap.HeightSubtiles)}, nil
		}
		return gametransition.SeamEndpoint{}, fmt.Errorf("d2legacy entry world: level %d stamp %d has no safe authored exit", zone.Request().LevelID, stampID)
	}
	return gametransition.SeamEndpoint{}, fmt.Errorf("d2legacy entry world: missing connection stamp %d", stampID)
}
