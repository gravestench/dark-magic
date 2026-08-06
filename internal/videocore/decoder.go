package videocore

import (
	"context"
	"image"
	"io"
	"time"
)

// Frame is one decoded video image and its presentation timestamp.
type Frame struct {
	Image image.Image
	PTS   time.Duration
}

// Decoder converts a seekable media stream into timestamped video frames.
// Implementations decode only; presentation and native texture ownership stay
// behind Presenter and rendercore.
type Decoder interface {
	Decode(context.Context, io.ReadSeeker, func(Frame) error) error
}
