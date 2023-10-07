package raylibRenderer

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

func (s *Service) GetTexture(uuid uuid.UUID, img image.Image) (texture rl.Texture2D, isNew bool) {
	key := uuid.String()

	bounds := img.Bounds()
	numBytes := bounds.Dx() * bounds.Dy() * 4 // RGBA

	cached, exists := s.cache.Retrieve(key)
	if !exists {
		cached = rl.LoadTextureFromImage(rl.NewImageFromImage(img))
		if err := s.cache.Insert(key, cached, numBytes); err != nil {
			s.logger.Error().Msgf("[%s] caching texture: %v", key, err)
		}

		return cached.(rl.Texture2D), true
	}

	return cached.(rl.Texture2D), false
}
