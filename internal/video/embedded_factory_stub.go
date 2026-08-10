//go:build !ffmpeg

package video

import (
	"image"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// NewEmbeddedBackend returns the explicit unavailable backend when FFmpeg was
// not selected at build time, keeping capability detection portable.
func NewEmbeddedBackend(*render.Composer, *audio.Mixer, image.Point) Backend {
	return Unavailable{}
}
