//go:build !ffmpeg

package video

import (
	"image"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

func NewEmbeddedBackend(*render.Composer, *audio.Mixer, image.Point) Backend {
	return Unavailable{}
}
