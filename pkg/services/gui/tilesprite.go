package gui

import (
	"fmt"
	"image"

	dc6 "github.com/gravestench/dc6/pkg"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

var _ managesTileSprites = &Service{}

func (s *Service) NewTileSprite(dc6Path, pl2Path string, gridWidth, gridHeight int) (ts *TileSprite, err error) {
	img, err := s.loadDc6WithPalette(dc6Path, pl2Path)
	if err != nil {
		return nil, fmt.Errorf("loading DC6 with palette transform: %v", err)
	}

	if gridWidth < 1 || gridHeight < 1 {
		gridWidth, gridHeight = 1, 1
	}

	ts = &TileSprite{
		service:    s,
		root:       s.renderer.NewRenderable(),
		DC6:        img,
		gridWidth:  gridWidth,
		gridHeight: gridHeight,
	}

	ts.setupChildNodes()

	return
}

type TileSprite struct {
	service               *Service
	DC6                   *dc6.DC6
	root                  raylibRenderer.Renderable
	gridWidth, gridHeight int
}

func (ts *TileSprite) Renderable() raylibRenderer.Renderable {
	return ts.root
}

func (ts *TileSprite) arrangeFramesInGrid() []image.Point {
	frames := ts.DC6.Directions[0].Frames

	if len(frames) == 0 {
		return nil
	}

	numFrames := len(frames)

	// Calculate the maximum height for each row and the maximum width for each column
	rowHeights := make([]int, ts.gridHeight)
	colWidths := make([]int, ts.gridWidth)

	for i, frame := range frames {
		row := i % ts.gridHeight
		col := i / ts.gridHeight

		frameWidth := frame.Bounds().Dx()
		frameHeight := frame.Bounds().Dy()

		if frameWidth > colWidths[col] {
			colWidths[col] = frameWidth
		}

		if frameHeight > rowHeights[row] {
			rowHeights[row] = frameHeight
		}
	}

	// Calculate the top-left origin coordinates
	var origins []image.Point
	x, y := 0, 0
	for i := 0; i < numFrames; i++ {
		origins = append(origins, image.Pt(x, y))

		x += colWidths[i/ts.gridHeight]
		if x >= colWidths[i/ts.gridHeight]*ts.gridWidth {
			x = 0
			y += rowHeights[i%ts.gridHeight]
		}
	}

	return origins
}

func (ts *TileSprite) setupChildNodes() {
	for idx, tcfg := range ts.arrangeFramesInGrid() {
		tile := ts.DC6.Directions[0].Frames[idx]

		tNode := ts.service.renderer.NewRenderable()
		tNode.SetOrigin(0, 0)
		tNode.SetImage(tile)

		tNode.SetPosition(float32(tcfg.X), float32(tcfg.Y))
		tNode.SetParent(ts.root)
	}
}
