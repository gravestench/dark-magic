package raylibRenderer

import (
	"image"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

func (s *Service) GetTexture(uuid uuid.UUID, img image.Image) (texture rl.Texture2D, isNew bool) {
	key := uuid.String()

	tx, exists := s.cache.Retrieve(key)
	if !exists {
		bounds := img.Bounds()
		numBytes := bounds.Dx() * bounds.Dy() * 4 // RGBA

		tx = rl.LoadTextureFromImage(rl.NewImageFromImage(img))

		if err := s.cache.Insert(key, tx, numBytes); err != nil {
			s.logger.Error().Msgf("[%s] caching texture: %v", key, err)
		}

		return tx.(rl.Texture2D), true
	}

	return tx.(rl.Texture2D), false
}
