package raylibRenderer

import (
	"image"
	"image/color"
	"sort"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) render() {
	sort.Slice(s.renderables, func(i, j int) bool {
		return s.renderables[i].ZIndex() > s.renderables[j].ZIndex()
	})

	for s.cache == nil {
		time.Sleep(time.Second)
	}

	for _, r := range s.renderables {
		layerID := r.UUID().String()

		if tx, exists := s.cache.Retrieve(layerID); !exists {
			img := r.Image()
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
		rl.UpdateTexture(tx, getAllPixelData(r.Image()))

		x, y := r.Position()
		rl.DrawTextureEx(tx, rl.Vector2{float32(x), float32(y)}, r.Rotation(), r.Scale(), rl.NewColor(255, 255, 255, uint8(r.Opacity()*255)))
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
