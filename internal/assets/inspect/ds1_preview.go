package assetinspect

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"io/fs"

	"github.com/gravestench/ds1"
)

// DS1Preview renders the structural floor and wall layout of a map stamp. It
// avoids DT1 dependencies so placement data can be diagnosed in isolation.
func DS1Preview(source fs.FS, path string) ([]byte, error) {
	file, err := source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening DS1 asset %q: %w", path, err)
	}
	defer closeFileWithoutReporting(file)

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading DS1 asset %q: %w", path, err)
	}

	stamp, err := ds1.FromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("decoding DS1 asset %q: %w", path, err)
	}

	const tileWidth, tileHeight, margin = 64, 32, 48

	width := (int(stamp.Width+stamp.Height) * tileWidth / 2) + margin*2
	height := (int(stamp.Width+stamp.Height) * tileHeight / 2) + margin*2
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	fill(canvas, color.RGBA{R: 17, G: 18, B: 22, A: 255})

	originX := int(stamp.Height)*tileWidth/2 + margin
	drawStructuralDS1Stamp(canvas, stamp, originX, tileWidth, tileHeight, margin)

	return encodePreviewPNG(canvas, "encoding DS1 preview")
}

// drawStructuralDS1Stamp maps DS1 grid order onto an isometric canvas. Using
// source order keeps overlapping wall markers deterministic for diagnostics.
func drawStructuralDS1Stamp(
	canvas *image.RGBA,
	stamp *ds1.DS1,
	originX int,
	tileWidth int,
	tileHeight int,
	margin int,
) {
	for y, row := range stamp.Tiles {
		for x, tile := range row {
			centerX := originX + (x-y)*tileWidth/2
			centerY := margin + (x+y)*tileHeight/2

			shade := uint8(55)
			if len(tile.Floors) > 0 {
				shade += uint8((int(tile.Floors[0].Style) + int(tile.Floors[0].Sequence)) % 90)
			}

			fillDiamond(
				canvas,
				centerX,
				centerY,
				tileWidth,
				tileHeight,
				color.RGBA{R: shade, G: shade + 18, B: shade, A: 255},
			)

			if len(tile.Walls) > 0 && tile.Walls[0].Prop1 != 0 {
				drawDiamond(
					canvas,
					centerX,
					centerY,
					tileWidth,
					tileHeight,
					color.RGBA{R: 205, G: 175, B: 95, A: 255},
				)
			}
		}
	}
}

// fill paints the full RGBA canvas, including transparent pixels that later
// compositing passes may leave untouched.
func fill(img *image.RGBA, c color.Color) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

// fillDiamond paints an isometric cell row by row so structural previews and
// missing-texture placeholders share exactly the same footprint.
func fillDiamond(img *image.RGBA, centerX, centerY, width, height int, c color.Color) {
	for deltaY := -height / 2; deltaY <= height/2; deltaY++ {
		half := (width / 2) * (height/2 - abs(deltaY)) / (height / 2)
		for x := centerX - half; x <= centerX+half; x++ {
			img.Set(x, centerY+deltaY, c)
		}
	}
}

// drawDiamond outlines occupied wall cells without hiding their floor shading,
// keeping both structural signals visible in the diagnostic image.
func drawDiamond(img *image.RGBA, centerX, centerY, width, height int, c color.Color) {
	for deltaY := -height / 2; deltaY <= height/2; deltaY++ {
		half := (width / 2) * (height/2 - abs(deltaY)) / (height / 2)
		img.Set(centerX-half, centerY+deltaY, c)
		img.Set(centerX+half, centerY+deltaY, c)
	}
}

// abs keeps diamond row widths symmetric around the center without converting
// signed scan-line offsets to an unsigned coordinate type.
func abs(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
