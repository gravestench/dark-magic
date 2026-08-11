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

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
