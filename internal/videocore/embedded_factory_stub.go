//go:build !ffmpeg

package videocore

import (
	"image"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/rendercore"
)

func NewEmbeddedBackend(*rendercore.Composer, *audio.Mixer, image.Point) Backend {
	return Unavailable{}
}
