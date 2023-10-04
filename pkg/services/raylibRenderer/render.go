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

	for _, r := range s.renderables {
		x, y := r.Position()
		tx := r.Texture()
		rl.UpdateTexture(tx, getAllPixelData(r.Image()))

		rl.DrawTextureEx(
			tx,
			rl.Vector2{X: float32(x), Y: float32(y)},
			r.Rotation(),
			r.Scale(),
			rl.NewColor(255, 255, 255, uint8(r.Opacity()*255)))
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
