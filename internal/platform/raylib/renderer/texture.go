package raylibRenderer

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) getTexture(key string, img image.Image) (texture rl.Texture2D, isNew bool) {

	bounds := img.Bounds()
	numBytes := bounds.Dx() * bounds.Dy() * 4 // RGBA

	cached, exists := s.cache.Retrieve(key)
	if !exists {
		cached = rl.LoadTextureFromImage(rl.NewImageFromImage(img))
		if err := s.cache.Insert(key, cached, numBytes); err != nil {
			s.logger.Error("caching texture", "key", key, "error", err)
		}

		return cached.(rl.Texture2D), true
	}

	return cached.(rl.Texture2D), false
}
