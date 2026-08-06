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

// AudioChunk is interleaved signed 16-bit little-endian PCM and its media PTS.
type AudioChunk struct {
	PCM        []byte
	PTS        time.Duration
	SampleRate int
	Channels   int
}

// AudioDecoder normalizes a media stream into timestamped engine PCM.
type AudioDecoder interface {
	DecodeAudio(context.Context, io.ReadSeeker, func(AudioChunk) error) error
}
