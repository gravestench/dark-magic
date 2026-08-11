package world

import "math"

// Point is a coordinate pair. Method names state which coordinate space a
// point belongs to; the small value type prevents projection code from growing
// renderer or ECS dependencies.
type Point struct{ X, Y float64 }

// Coordinates owns every conversion between authored tile/subtile space and
// the map's isometric pixel canvas. Camera conversion is included here so UI
// code never has to reproduce the projection equations.
type Coordinates struct {
	HeightTiles int
}

func (c Coordinates) TileToSubtile(point Point) Point {
	return Point{X: point.X * SubtilesPerTile, Y: point.Y * SubtilesPerTile}
}

func (c Coordinates) SubtileToTile(point Point) Point {
	return Point{X: point.X / SubtilesPerTile, Y: point.Y / SubtilesPerTile}
}

func (c Coordinates) SubtileToWorldPixel(point Point) Point {
	originX := float64(c.HeightTiles*TilePixelWidth/2 + PreviewMargin)
	// The tile origin is its top isometric vertex, while an integer gameplay
	// position names the CENTER of that subtile's 16x8 occupancy diamond. Riiablo
	// applies this same half-subtile height when drawing filled walkable cells.
	originY := float64(PreviewMargin + TilePixelHeight/(2*SubtilesPerTile))
	return Point{
		X: originX + (point.X-point.Y)*TilePixelWidth/(2*SubtilesPerTile),
		Y: originY + (point.X+point.Y)*TilePixelHeight/(2*SubtilesPerTile),
	}
}

func (c Coordinates) WorldPixelToSubtile(point Point) Point {
	originX := float64(c.HeightTiles*TilePixelWidth/2 + PreviewMargin)
	originY := float64(PreviewMargin + TilePixelHeight/(2*SubtilesPerTile))
	difference := (point.X - originX) * (2 * SubtilesPerTile) / TilePixelWidth
	sum := (point.Y - originY) * (2 * SubtilesPerTile) / TilePixelHeight
	return Point{X: (difference + sum) / 2, Y: (sum - difference) / 2}
}

// WorldPixelToScreen anchors the camera at a chosen screen point. That anchor
// moves when a side overlay changes the unobscured part of the viewport.
func (Coordinates) WorldPixelToScreen(point, camera, anchor Point) Point {
	return Point{X: anchor.X + point.X - camera.X, Y: anchor.Y + point.Y - camera.Y}
}

func (Coordinates) ScreenToWorldPixel(point, camera, anchor Point) Point {
	return Point{X: camera.X + point.X - anchor.X, Y: camera.Y + point.Y - anchor.Y}
}

func (c Coordinates) SubtileToScreen(point, cameraSubtile, anchor Point) Point {
	return c.WorldPixelToScreen(c.SubtileToWorldPixel(point), c.SubtileToWorldPixel(cameraSubtile), anchor)
}

func (c Coordinates) ScreenToSubtile(point, cameraSubtile, anchor Point) Point {
	worldPixel := c.ScreenToWorldPixel(point, c.SubtileToWorldPixel(cameraSubtile), anchor)
	return c.WorldPixelToSubtile(worldPixel)
}

// CollisionCell defines the single sampling rule for continuous authoritative
// positions. A position names a subtile center, so half values enter the next
// cell instead of relying on language-specific integer conversion behavior.
func CollisionCell(value float64) int { return int(math.Floor(value + 0.5)) }

func (m *Map) Coordinates() Coordinates { return Coordinates{HeightTiles: m.HeightTiles} }

func (m *Map) FlagsAtPosition(x, y float64) (Flags, bool) {
	return m.FlagsAt(CollisionCell(x), CollisionCell(y))
}
