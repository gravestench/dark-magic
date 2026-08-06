package assetcatalog

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	dc6 "github.com/gravestench/dc6/pkg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// DC6ContactSheet renders every direction/frame with a label and checkerboard
// transparency. Cells share a size so offsets and dimension changes stand out.
func DC6ContactSheet(asset *dc6.DC6) ([]byte, error) {
	if asset == nil {
		return nil, fmt.Errorf("asset catalog: nil DC6")
	}
	frames := 0
	maxWidth, maxHeight := 1, 1
	for _, direction := range asset.Directions {
		frames += len(direction.Frames)
		for _, frame := range direction.Frames {
			if int(frame.Width) > maxWidth {
				maxWidth = int(frame.Width)
			}
			if int(frame.Height) > maxHeight {
				maxHeight = int(frame.Height)
			}
		}
	}
	if frames == 0 {
		return nil, fmt.Errorf("asset catalog: DC6 has no frames")
	}
	const padding, labelHeight, maxColumns = 8, 18, 8
	columns := int(math.Ceil(math.Sqrt(float64(frames))))
	if columns > maxColumns {
		columns = maxColumns
	}
	rows := (frames + columns - 1) / columns
	cellWidth := maxWidth + padding*2
	cellHeight := maxHeight + labelHeight + padding*2
	canvas := image.NewRGBA(image.Rect(0, 0, columns*cellWidth, rows*cellHeight))
	fillChecker(canvas)

	index := 0
	for directionIndex, direction := range asset.Directions {
		for frameIndex, frame := range direction.Frames {
			column, row := index%columns, index/columns
			origin := image.Pt(column*cellWidth+padding, row*cellHeight+labelHeight+padding)
			frameImage, err := assetdecode.FrameImage(asset, frame)
			if err != nil {
				return nil, err
			}
			x := origin.X + (maxWidth-frameImage.Bounds().Dx())/2
			y := origin.Y + maxHeight - frameImage.Bounds().Dy()
			draw.Draw(canvas, frameImage.Bounds().Add(image.Pt(x, y)), frameImage, frameImage.Bounds().Min, draw.Over)
			drawLabel(canvas, column*cellWidth+padding, row*cellHeight+14,
				fmt.Sprintf("d%d f%d %dx%d %+d,%+d", directionIndex, frameIndex, frame.Width, frame.Height, frame.OffsetX, frame.OffsetY))
			index++
		}
	}
	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		return nil, fmt.Errorf("asset catalog: encode contact sheet: %w", err)
	}
	return output.Bytes(), nil
}

func fillChecker(target *image.RGBA) {
	colors := [2]color.RGBA{{R: 34, G: 36, B: 42, A: 255}, {R: 51, G: 54, B: 62, A: 255}}
	const size = 12
	for y := target.Bounds().Min.Y; y < target.Bounds().Max.Y; y++ {
		for x := target.Bounds().Min.X; x < target.Bounds().Max.X; x++ {
			target.SetRGBA(x, y, colors[(x/size+y/size)&1])
		}
	}
}

func drawLabel(target draw.Image, x, y int, label string) {
	drawer := font.Drawer{Dst: target, Src: image.NewUniform(color.White), Face: basicfont.Face7x13, Dot: fixed.P(x, y)}
	drawer.DrawString(label)
}
