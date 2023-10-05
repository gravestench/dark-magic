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

	for _, child := range s.rootNode.(hasChildren).Children() {
		x, y := child.Position()

		px := getAllPixelData(child.Image())
		if len(px) < 4 {
			continue
		}

		tx := child.Texture()

		rl.UpdateTexture(tx, px)

		rl.DrawTextureEx(
			tx,
			rl.Vector2{X: float32(x), Y: float32(y)},
			child.Rotation(),
			child.Scale(),
			rl.NewColor(255, 255, 255, uint8(child.Opacity()*255)))
	}
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
