//go:build ffmpeg

package videocore

import (
	"image"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/rendercore"
)

func NewEmbeddedBackend(composer *rendercore.Composer, mixer *audiocore.Mixer, viewport image.Point) Backend {
	decoder := FFmpegDecoder{}
	return &Embedded{Composer: composer, Mixer: mixer, Viewport: viewport, Video: decoder, Audio: decoder}
}
