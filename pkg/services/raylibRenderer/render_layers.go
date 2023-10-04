package raylibRenderer

import (
	"image"
	"image/color"
	"image/draw"
	"sort"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) drawRenderableLayers() {
	sort.Slice(s.renderableLayers, func(i, j int) bool {
		return s.renderableLayers[i].LayerIndex() > s.renderableLayers[j].LayerIndex()
	})

	for s.cache == nil {
		time.Sleep(time.Second)
	}

	for _, layer := range s.renderableLayers {
		layerID := layer.UUID().String()

		if tx, exists := s.cache.Retrieve(layerID); !exists {
			img := layer.LayerImage()
			tx = rl.LoadTextureFromImage(rl.NewImageFromImage(img))
			bounds := img.Bounds()
			numBytes := bounds.Dx() * bounds.Dy() * 4 // RGBA
			err := s.cache.Insert(layerID, tx, numBytes)
			if err != nil {
				s.logger.Error().Msgf("caching texture: %v", err)
			}
		}

		cached, exists := s.cache.Retrieve(layerID)
		if !exists {
			s.logger.Error().Msg("texture not in cache")
			continue
		}

		tx := cached.(rl.Texture2D)
		rl.UpdateTexture(tx, getAllPixelData(layer.LayerImage()))

		x, y := layer.Position()
		rl.DrawTextureEx(tx, rl.Vector2{float32(x), float32(y)}, layer.Rotation(), layer.Scale(), rl.NewColor(255, 255, 255, uint8(layer.Opacity()*255)))
	}
}

func getAllPixelData(img image.Image) []color.RGBA {
	// Get the dimensions of the image
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new image.RGBA with the same dimensions as the original image
	rgba := image.NewRGBA(bounds)

	// Draw the original image onto the new RGBA image
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Convert the RGBA image to a slice of color.RGBA
	pixels := make([]color.RGBA, width*height)
	index := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := rgba.At(x, y).(color.RGBA)
			pixels[index] = pixel
			index++
		}
	}

	return pixels
}
