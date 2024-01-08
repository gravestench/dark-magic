package ui

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

func (s *Service) FillRect(x, y, w, h, strokeWidth int, fill, stroke color.Color) raylibRenderer.Renderable {
	rect := image.Rect(0, 0, w, h)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill the background with a transparent color
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Transparent}, image.Point{}, draw.Src)

	// Fill the rectangle with the fill color
	draw.Draw(img, rect, &image.Uniform{fill}, image.Point{}, draw.Src)

	// Draw the stroke
	var (
		p  image.Point
		u  = &image.Uniform{stroke}
		op = draw.Src
	)

	for i := 0; i < strokeWidth; i++ {
		// Top border
		if x1, y1, x2, y2 := 0, 0, w, strokeWidth; true {
			draw.Draw(img, image.Rect(x1, y1, x2, y2), u, p, op)
		}

		// Bottom border
		if x1, y1, x2, y2 := 0, h-strokeWidth, w, h; true {
			draw.Draw(img, image.Rect(x1, y1, x2, y2), u, p, op)
		}

		// Left border
		if x1, y1, x2, y2 := 0, 0, strokeWidth, h; true {
			draw.Draw(img, image.Rect(x1, y1, x2, y2), u, p, op)
		}

		// Right border
		if x1, y1, x2, y2 := w-strokeWidth, 0, w, h; true {
			draw.Draw(img, image.Rect(x1, y1, x2, y2), u, p, op)
		}
	}

	rootNode := s.renderer.NewRenderable()
	rootNode.SetImage(img)
	rootNode.SetPosition(float32(x), float32(y))

	return rootNode
}
