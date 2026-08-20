package assetdecode

import (
	"fmt"
	"image"
	"image/draw"

	dc6 "github.com/gravestench/dc6/pkg"
)

const dc6PageSize = 256

// dc6PageGrid records the derived tile and canvas geometry shared by every
// logical page in one DC6 direction.
type dc6PageGrid struct {
	columns int
	rows    int
	width   int
	height  int
}

// CombinedDC6Pages reconstructs tiled logical images using Diablo II's 256-pixel
// page convention, where one authored image may span several codec frames.
func CombinedDC6Pages(asset *dc6.DC6, direction int) ([]image.Image, error) {
	if asset == nil || direction < 0 || direction >= len(asset.Directions) {
		return nil, fmt.Errorf("DC6 combined direction %d is out of range", direction)
	}

	frames := asset.Directions[direction].Frames
	if len(frames) == 0 {
		return nil, fmt.Errorf("DC6 combined direction has no frames")
	}

	grid := measureDC6PageGrid(frames)

	framesPerPage := grid.rows * grid.columns
	if framesPerPage <= 0 || len(frames)%framesPerPage != 0 {
		return nil, fmt.Errorf(
			"DC6 combined grid %dx%d does not divide %d frames",
			grid.columns,
			grid.rows,
			len(frames),
		)
	}

	pageCount := len(frames) / framesPerPage

	pages := make([]image.Image, 0, pageCount)
	for page := 0; page < pageCount; page++ {
		canvas, err := drawDC6Page(asset, frames, page*framesPerPage, grid)
		if err != nil {
			return nil, err
		}

		pages = append(pages, canvas)
	}

	return pages, nil
}

// measureDC6PageGrid derives the tile geometry from the first short frame in
// each axis, which is the format's marker for the logical image edge.
func measureDC6PageGrid(frames []*dc6.Frame) dc6PageGrid {
	grid := dc6PageGrid{}
	for _, frame := range frames {
		grid.columns++

		grid.width += int(frame.Width)
		if frame.Width < dc6PageSize {
			break
		}
	}

	for index := 0; index < len(frames); index += grid.columns {
		grid.rows++

		grid.height += int(frames[index].Height)
		if frames[index].Height < dc6PageSize {
			break
		}
	}

	return grid
}

// drawDC6Page decodes and places one contiguous frame group in row-major order,
// preserving the archive's authored page ordering.
func drawDC6Page(
	asset *dc6.DC6,
	frames []*dc6.Frame,
	firstFrame int,
	grid dc6PageGrid,
) (*image.RGBA, error) {
	canvas := image.NewRGBA(image.Rect(0, 0, grid.width, grid.height))
	frameIndex := firstFrame
	y := 0

	for row := 0; row < grid.rows; row++ {
		x := 0

		for column := 0; column < grid.columns; column++ {
			frame := frames[frameIndex]
			frameIndex++

			decoded, err := FrameImage(asset, frame)
			if err != nil {
				return nil, err
			}

			destination := decoded.Bounds().Add(image.Pt(x, y))
			draw.Draw(canvas, destination, decoded, decoded.Bounds().Min, draw.Over)

			x += int(frame.Width)
		}

		// Full pages advance by the format's fixed tile size, even when the
		// final row contributes a shorter logical height to the canvas.
		y += dc6PageSize
	}

	return canvas, nil
}
