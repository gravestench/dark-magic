package raylibRenderer

import (
	"image"
	"image/color"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) render() {
	for s.cache == nil {
		time.Sleep(time.Second)
	}

	s.renderRecursively(s.rootNode)
}

func (s *Service) renderRecursively(node Renderable) {
	if node.IsEnabled() {
		s.renderNode(node)
	}

	for _, child := range node.Children() {
		s.renderRecursively(child)
	}
}

func (s *Service) renderNode(node Renderable) {
	x, y := node.Position()

	px := getAllPixelData(node.Image())
	if len(px) < 4 {
		return
	}

	tx := node.Texture()

	rl.UpdateTexture(tx, px)

	rl.DrawTextureEx(
		tx,
		rl.Vector2{X: float32(x), Y: float32(y)},
		node.Rotation(),
		node.Scale(),
		rl.NewColor(255, 255, 255, uint8(node.Opacity()*255)))
}

func getAllPixelData(img image.Image) []color.RGBA {
	// Get the dimensions of the image
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Convert the RGBA image to a slice of color.RGBA
	pixels := make([]color.RGBA, width*height)
	index := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.At(x, y).(color.RGBA)
			pixels[index] = pixel
			index++
		}
	}

	return pixels
}
