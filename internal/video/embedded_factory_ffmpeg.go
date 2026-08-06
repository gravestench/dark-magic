//go:build ffmpeg

package video

import (
	"image"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

func NewEmbeddedBackend(composer *render.Composer, mixer *audio.Mixer, viewport image.Point) Backend {
	decoder := FFmpegDecoder{}
	return &Embedded{Composer: composer, Mixer: mixer, Viewport: viewport, Video: decoder, Audio: decoder}
}
