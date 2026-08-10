package raylibRenderer

import (
	"image"
	"runtime"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Service) getTexture(key string, img image.Image) (texture rl.Texture2D, isNew bool) {

	bounds := img.Bounds()
	numBytes := bounds.Dx() * bounds.Dy() * 4 // RGBA

	cached, exists := s.cache.Retrieve(key)
	if !exists {
		cached = loadTexture(img)
		s.textureUploads.Add(1)
		s.textureUploadBytes.Add(uint64(numBytes))
		if err := s.cache.Insert(key, cached, numBytes); err != nil {
			s.logger.Error("caching texture", "key", key, "error", err)
		}

		return cached.(rl.Texture2D), true
	}

	return cached.(rl.Texture2D), false
}

// loadTexture takes the zero-copy path used by decoded engine assets. The
// raylib-go convenience converter draws one pixel through cgo at a time, which
// is extremely expensive for animation frames. LoadTextureFromImage consumes
// the bytes synchronously, so the Go slice only needs to remain alive until the
// call returns.
func loadTexture(img image.Image) rl.Texture2D {
	bounds := img.Bounds()
	if pixels, ok := contiguousRGBA(img); ok {
		native := rl.NewImage(pixels, int32(bounds.Dx()), int32(bounds.Dy()), 1, rl.UncompressedR8g8b8a8)
		texture := rl.LoadTextureFromImage(native)
		runtime.KeepAlive(pixels)
		return texture
	}

	// Unusual image implementations still get a safe conversion path. Unlike
	// the former call site, release the temporary C image after GPU upload.
	native := rl.NewImageFromImage(img)
	defer rl.UnloadImage(native)
	return rl.LoadTextureFromImage(native)
}
