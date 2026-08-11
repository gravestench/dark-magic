package maprender

import (
	"image"
	"image/color"

	"github.com/gravestench/dark-magic/internal/game/world"
)

// CollisionImage creates an explicitly diagnostic overlay in the same logical
// canvas as Compose. Production map rendering never calls this full-image path.
func CollisionImage(mapData *world.Map) *image.RGBA {
	width := (mapData.WidthTiles+mapData.HeightTiles)*world.TilePixelWidth/2 + world.PreviewMargin*2
	height := (mapData.WidthTiles+mapData.HeightTiles)*world.TilePixelHeight/2 + world.PreviewMargin*2
	result := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < mapData.HeightSubtiles; y++ {
		for x := 0; x < mapData.WidthSubtiles; x++ {
			flags, _ := mapData.FlagsAt(x, y)
			shade, visible := collisionColor(flags)
			if !visible {
				continue
			}
			centerX, centerY := mapData.SubtileToPixel(float64(x), float64(y))
			fillDiamond(result, int(centerX), int(centerY), world.TilePixelWidth/(world.SubtilesPerTile*2), world.TilePixelHeight/(world.SubtilesPerTile*2), shade)
		}
	}
	return result
}

// CollisionRegionImage creates only the requested subtile diagnostic window.
// It returns map-canvas pixel bounds so a retained node can share the exact
// transform used by DT1 placements without allocating a full outdoor canvas.
func CollisionRegionImage(mapData *world.Map, region image.Rectangle) (*image.RGBA, image.Rectangle) {
	region = region.Intersect(image.Rect(0, 0, mapData.WidthSubtiles, mapData.HeightSubtiles))
	if region.Empty() {
		return image.NewRGBA(image.Rect(0, 0, 1, 1)), image.Rect(0, 0, 1, 1)
	}
	points := []image.Point{}
	for _, corner := range []image.Point{region.Min, {X: region.Max.X, Y: region.Min.Y}, region.Max, {X: region.Min.X, Y: region.Max.Y}} {
		x, y := mapData.SubtileToPixel(float64(corner.X), float64(corner.Y))
		points = append(points, image.Pt(int(x), int(y)))
	}
	bounds := image.Rect(points[0].X, points[0].Y, points[0].X+1, points[0].Y+1)
	for _, point := range points[1:] {
		bounds = bounds.Union(image.Rect(point.X, point.Y, point.X+1, point.Y+1))
	}
	halfWidth := world.TilePixelWidth / (world.SubtilesPerTile * 2)
	halfHeight := world.TilePixelHeight / (world.SubtilesPerTile * 2)
	bounds = image.Rect(bounds.Min.X-halfWidth, bounds.Min.Y-halfHeight, bounds.Max.X+halfWidth+1, bounds.Max.Y+halfHeight+1)
	result := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			flags, _ := mapData.FlagsAt(x, y)
			shade, visible := collisionColor(flags)
			if !visible {
				continue
			}
			centerX, centerY := mapData.SubtileToPixel(float64(x), float64(y))
			fillDiamond(result, int(centerX)-bounds.Min.X, int(centerY)-bounds.Min.Y, halfWidth, halfHeight, shade)
		}
	}
	return result, bounds
}

// TileRegionImage draws the projection geometry which turns authoritative map
// coordinates into pixels. Cyan diamonds are DT1 tiles, faint blue diamonds
// are their 5x5 subtiles, and yellow crosses mark tile origins.
func TileRegionImage(mapData *world.Map, region image.Rectangle) (*image.RGBA, image.Rectangle) {
	region, result, bounds := diagnosticRegion(mapData, region)
	if region.Empty() {
		return result, bounds
	}
	subtileHalfWidth := world.TilePixelWidth / (world.SubtilesPerTile * 2)
	subtileHalfHeight := world.TilePixelHeight / (world.SubtilesPerTile * 2)
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			centerX, centerY := mapData.SubtileToPixel(float64(x), float64(y))
			drawDiamond(result, int(centerX)-bounds.Min.X, int(centerY)-bounds.Min.Y,
				subtileHalfWidth, subtileHalfHeight, color.RGBA{G: 128, B: 255, A: 70})
		}
	}
	firstTileX, firstTileY := region.Min.X/world.SubtilesPerTile, region.Min.Y/world.SubtilesPerTile
	lastTileX := (region.Max.X + world.SubtilesPerTile - 1) / world.SubtilesPerTile
	lastTileY := (region.Max.Y + world.SubtilesPerTile - 1) / world.SubtilesPerTile
	for tileY := firstTileY; tileY < lastTileY; tileY++ {
		for tileX := firstTileX; tileX < lastTileX; tileX++ {
			originX, originY := mapData.SubtileToPixel(float64(tileX*world.SubtilesPerTile), float64(tileY*world.SubtilesPerTile))
			x, y := int(originX)-bounds.Min.X, int(originY)-bounds.Min.Y
			drawDiamond(result, x, y, world.TilePixelWidth/2, world.TilePixelHeight/2, color.RGBA{G: 255, B: 255, A: 210})
			drawCross(result, x, y, 5, color.RGBA{R: 255, G: 220, A: 255})
		}
	}
	return result, bounds
}

func diagnosticRegion(mapData *world.Map, region image.Rectangle) (image.Rectangle, *image.RGBA, image.Rectangle) {
	region = region.Intersect(image.Rect(0, 0, mapData.WidthSubtiles, mapData.HeightSubtiles))
	if region.Empty() {
		return region, image.NewRGBA(image.Rect(0, 0, 1, 1)), image.Rect(0, 0, 1, 1)
	}
	points := []image.Point{}
	for _, corner := range []image.Point{region.Min, {X: region.Max.X, Y: region.Min.Y}, region.Max, {X: region.Min.X, Y: region.Max.Y}} {
		x, y := mapData.SubtileToPixel(float64(corner.X), float64(corner.Y))
		points = append(points, image.Pt(int(x), int(y)))
	}
	bounds := image.Rect(points[0].X, points[0].Y, points[0].X+1, points[0].Y+1)
	for _, point := range points[1:] {
		bounds = bounds.Union(image.Rect(point.X, point.Y, point.X+1, point.Y+1))
	}
	bounds = image.Rect(bounds.Min.X-world.TilePixelWidth/2-1, bounds.Min.Y-world.TilePixelHeight/2-1,
		bounds.Max.X+world.TilePixelWidth/2+2, bounds.Max.Y+world.TilePixelHeight/2+2)
	return region, image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy())), bounds
}

func collisionColor(flags world.Flags) (color.RGBA, bool) {
	switch {
	case flags.BlockWalk:
		return color.RGBA{R: 255, A: 150}, true
	case flags.BlockPlayerWalk:
		return color.RGBA{R: 255, G: 128, A: 150}, true
	case flags.BlockLOS:
		return color.RGBA{B: 255, A: 140}, true
	case flags.BlockJump:
		return color.RGBA{R: 255, B: 255, A: 140}, true
	case flags.BlockLight:
		return color.RGBA{R: 255, G: 255, A: 120}, true
	default:
		return color.RGBA{}, false
	}
}

func fillDiamond(target *image.RGBA, centerX, centerY, halfWidth, halfHeight int, shade color.RGBA) {
	for offsetY := -halfHeight; offsetY <= halfHeight; offsetY++ {
		span := halfWidth * (halfHeight - abs(offsetY)) / max(halfHeight, 1)
		for offsetX := -span; offsetX <= span; offsetX++ {
			point := image.Pt(centerX+offsetX, centerY+offsetY)
			if point.In(target.Bounds()) {
				target.SetRGBA(point.X, point.Y, shade)
			}
		}
	}
}

func drawDiamond(target *image.RGBA, centerX, centerY, halfWidth, halfHeight int, shade color.RGBA) {
	drawLine(target, centerX, centerY-halfHeight, centerX+halfWidth, centerY, shade)
	drawLine(target, centerX+halfWidth, centerY, centerX, centerY+halfHeight, shade)
	drawLine(target, centerX, centerY+halfHeight, centerX-halfWidth, centerY, shade)
	drawLine(target, centerX-halfWidth, centerY, centerX, centerY-halfHeight, shade)
}

func drawCross(target *image.RGBA, x, y, radius int, shade color.RGBA) {
	drawLine(target, x-radius, y, x+radius, y, shade)
	drawLine(target, x, y-radius, x, y+radius, shade)
}

func drawLine(target *image.RGBA, x0, y0, x1, y1 int, shade color.RGBA) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := -1, -1
	if x0 < x1 {
		sx = 1
	}
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		if image.Pt(x0, y0).In(target.Bounds()) {
			target.SetRGBA(x0, y0, shade)
		}
		if x0 == x1 && y0 == y1 {
			return
		}
		twice := 2 * err
		if twice >= dy {
			err += dy
			x0 += sx
		}
		if twice <= dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
