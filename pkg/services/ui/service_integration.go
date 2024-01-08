package ui

import (
	"image/color"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type CreatesShapes interface {
	CreatesRectangles
}

type CreatesRectangles interface {
	FillRect(x, y, w, h, strokeWidth int, fill, stroke color.Color) raylibRenderer.Renderable
}
