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

const (
	contactSheetPadding     = 8
	contactSheetLabelHeight = 18
	contactSheetMaxColumns  = 8
)

// contactSheetLayout centralizes dimensions shared by measurement and rendering so neither phase can drift in how it
// positions cells or allocates the canvas.
type contactSheetLayout struct {
	frameWidth  int
	frameHeight int
	cellWidth   int
	cellHeight  int
	columns     int
	rows        int
}

// DC6ContactSheet renders every direction and frame into uniformly sized cells. The shared dimensions, labels, and
// checkerboard make placement offsets, dimension changes, and transparency visible without proprietary source bytes.
func DC6ContactSheet(asset *dc6.DC6) ([]byte, error) {
	if asset == nil {
		return nil, fmt.Errorf("asset catalog: nil DC6")
	}

	layout, err := measureContactSheet(asset)
	if err != nil {
		return nil, err
	}

	canvas := image.NewRGBA(image.Rect(0, 0, layout.columns*layout.cellWidth, layout.rows*layout.cellHeight))
	fillChecker(canvas)

	if err := renderContactSheetFrames(canvas, asset, layout); err != nil {
		return nil, err
	}

	return encodeContactSheet(canvas)
}

// measureContactSheet derives one cell size from the largest frame so every frame remains directly comparable. It
// rejects empty decoder output before column and row calculations could divide by zero.
func measureContactSheet(asset *dc6.DC6) (contactSheetLayout, error) {
	frameCount := 0
	maxWidth, maxHeight := 1, 1

	for _, direction := range asset.Directions {
		frameCount += len(direction.Frames)

		for _, frame := range direction.Frames {
			if int(frame.Width) > maxWidth {
				maxWidth = int(frame.Width)
			}

			if int(frame.Height) > maxHeight {
				maxHeight = int(frame.Height)
			}
		}
	}

	if frameCount == 0 {
		return contactSheetLayout{}, fmt.Errorf("asset catalog: DC6 has no frames")
	}

	columns := int(math.Ceil(math.Sqrt(float64(frameCount))))
	if columns > contactSheetMaxColumns {
		columns = contactSheetMaxColumns
	}

	return contactSheetLayout{
		frameWidth:  maxWidth,
		frameHeight: maxHeight,
		cellWidth:   maxWidth + contactSheetPadding*2,
		cellHeight:  maxHeight + contactSheetLabelHeight + contactSheetPadding*2,
		columns:     columns,
		rows:        (frameCount + columns - 1) / columns,
	}, nil
}

// renderContactSheetFrames preserves decoder direction/frame order so labels and visual placement correspond to the
// metadata emitted by Verify. Rendering stops at the first invalid frame because a partial sheet would be misleading.
func renderContactSheetFrames(canvas draw.Image, asset *dc6.DC6, layout contactSheetLayout) error {
	frameIndexInSheet := 0

	for directionIndex, direction := range asset.Directions {
		for frameIndex, frame := range direction.Frames {
			column := frameIndexInSheet % layout.columns
			row := frameIndexInSheet / layout.columns
			origin := image.Pt(
				column*layout.cellWidth+contactSheetPadding,
				row*layout.cellHeight+contactSheetLabelHeight+contactSheetPadding,
			)

			frameImage, err := assetdecode.FrameImage(asset, frame)
			if err != nil {
				return err
			}

			// Center horizontally and align to the bottom so dimensions and placement offsets remain easy to compare.
			x := origin.X + (layout.frameWidth-frameImage.Bounds().Dx())/2
			y := origin.Y + layout.frameHeight - frameImage.Bounds().Dy()
			draw.Draw(canvas, frameImage.Bounds().Add(image.Pt(x, y)), frameImage, frameImage.Bounds().Min, draw.Over)

			label := fmt.Sprintf(
				"d%d f%d %dx%d %+d,%+d",
				directionIndex,
				frameIndex,
				frame.Width,
				frame.Height,
				frame.OffsetX,
				frame.OffsetY,
			)
			drawLabel(
				canvas,
				column*layout.cellWidth+contactSheetPadding,
				row*layout.cellHeight+14,
				label,
			)

			frameIndexInSheet++
		}
	}

	return nil
}

// encodeContactSheet wraps PNG failures with catalog context while returning an owned byte slice to the caller.
func encodeContactSheet(canvas image.Image) ([]byte, error) {
	var output bytes.Buffer
	if err := png.Encode(&output, canvas); err != nil {
		return nil, fmt.Errorf("asset catalog: encode contact sheet: %w", err)
	}

	return output.Bytes(), nil
}

// fillChecker paints opaque alternating tiles beneath every frame so transparent pixels remain distinguishable from
// the sheet background across both dark and light sprite colors.
func fillChecker(target *image.RGBA) {
	colors := [2]color.RGBA{{R: 34, G: 36, B: 42, A: 255}, {R: 51, G: 54, B: 62, A: 255}}

	const size = 12

	for y := target.Bounds().Min.Y; y < target.Bounds().Max.Y; y++ {
		for x := target.Bounds().Min.X; x < target.Bounds().Max.X; x++ {
			target.SetRGBA(x, y, colors[(x/size+y/size)&1])
		}
	}
}

// drawLabel uses a fixed bitmap face so labels render identically across hosts and do not depend on installed fonts.
func drawLabel(target draw.Image, x, y int, label string) {
	drawer := font.Drawer{Dst: target, Src: image.NewUniform(color.White), Face: basicfont.Face7x13, Dot: fixed.P(x, y)}
	drawer.DrawString(label)
}
