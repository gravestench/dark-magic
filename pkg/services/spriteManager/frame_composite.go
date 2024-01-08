package spriteManager

import (
	"image"
)

type FrameComposite struct {
	frames   []image.Image
	GridSize int
	Layout   []int
}
