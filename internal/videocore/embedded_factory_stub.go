//go:build !ffmpeg

package videocore

import (
	"image"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/rendercore"
)

func NewEmbeddedBackend(*rendercore.Composer, *audiocore.Mixer, image.Point) Backend {
	return Unavailable{}
}
