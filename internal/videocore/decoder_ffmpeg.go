//go:build ffmpeg

package videocore

import (
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"time"

	"github.com/asticode/go-astiav"
)

// FFmpegDecoder decodes Bink and other libav-supported video streams in the
// Dark Magic process. Build with -tags ffmpeg and link the libav development
// libraries advertised through pkg-config.
type FFmpegDecoder struct{}

func (FFmpegDecoder) Decode(ctx context.Context, input io.ReadSeeker, emit func(Frame) error) error {
	if input == nil || emit == nil {
		return errors.New("videocore: decoder input and frame callback are required")
	}
	format := astiav.AllocFormatContext()
	if format == nil {
		return errors.New("videocore: allocate FFmpeg format context")
	}
	defer format.Free()
	ioContext, err := astiav.AllocIOContext(32*1024, false, input.Read, input.Seek, nil)
	if err != nil {
		return fmt.Errorf("videocore: allocate FFmpeg IO: %w", err)
	}
	defer ioContext.Free()
	format.SetPb(ioContext)
	if err := format.OpenInput("", nil, nil); err != nil {
		return fmt.Errorf("videocore: open media: %w", err)
	}
	defer format.CloseInput()
	if err := format.FindStreamInfo(nil); err != nil {
		return fmt.Errorf("videocore: inspect media streams: %w", err)
	}
	stream, codec, err := format.FindBestStream(astiav.MediaTypeVideo, -1, -1)
	if err != nil {
		return fmt.Errorf("videocore: find video stream: %w", err)
	}
	decoder := astiav.AllocCodecContext(codec)
	if decoder == nil {
		return errors.New("videocore: allocate FFmpeg decoder")
	}
	defer decoder.Free()
	if err := stream.CodecParameters().ToCodecContext(decoder); err != nil {
		return fmt.Errorf("videocore: configure decoder: %w", err)
	}
	if err := decoder.Open(codec, nil); err != nil {
		return fmt.Errorf("videocore: open decoder: %w", err)
	}

	packet := astiav.AllocPacket()
	decoded := astiav.AllocFrame()
	if packet == nil || decoded == nil {
		if packet != nil {
			packet.Free()
		}
		if decoded != nil {
			decoded.Free()
		}
		return errors.New("videocore: allocate FFmpeg packet/frame")
	}
	defer packet.Free()
	defer decoded.Free()

	var scaler *astiav.SoftwareScaleContext
	defer func() {
		if scaler != nil {
			scaler.Free()
		}
	}()
	timeBase := stream.TimeBase().Float64()
	receive := func() error {
		for {
			decoded.Unref()
			if err := decoder.ReceiveFrame(decoded); err != nil {
				if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
					return nil
				}
				return fmt.Errorf("videocore: receive video frame: %w", err)
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if scaler == nil {
				var err error
				scaler, err = astiav.CreateSoftwareScaleContext(decoded.Width(), decoded.Height(), decoded.PixelFormat(), decoded.Width(), decoded.Height(), astiav.PixelFormatRgba, astiav.NewSoftwareScaleContextFlags(astiav.SoftwareScaleContextFlagBilinear))
				if err != nil {
					return fmt.Errorf("videocore: create RGBA converter: %w", err)
				}
			}
			rgba := astiav.AllocFrame()
			if rgba == nil {
				return errors.New("videocore: allocate RGBA frame")
			}
			if err := scaler.ScaleFrame(decoded, rgba); err != nil {
				rgba.Free()
				return fmt.Errorf("videocore: convert video frame: %w", err)
			}
			output := image.NewNRGBA(image.Rect(0, 0, rgba.Width(), rgba.Height()))
			if err := rgba.Data().ToImage(output); err != nil {
				rgba.Free()
				return fmt.Errorf("videocore: copy video frame: %w", err)
			}
			rgba.Free()
			if err := emit(Frame{Image: output, PTS: time.Duration(float64(decoded.Pts()) * timeBase * float64(time.Second))}); err != nil {
				return err
			}
		}
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		packet.Unref()
		err := format.ReadFrame(packet)
		if errors.Is(err, astiav.ErrEof) {
			if err := decoder.SendPacket(nil); err != nil && !errors.Is(err, astiav.ErrEof) {
				return fmt.Errorf("videocore: flush decoder: %w", err)
			}
			return receive()
		}
		if err != nil {
			return fmt.Errorf("videocore: read media packet: %w", err)
		}
		if packet.StreamIndex() != stream.Index() {
			continue
		}
		if err := decoder.SendPacket(packet); err != nil {
			return fmt.Errorf("videocore: send video packet: %w", err)
		}
		if err := receive(); err != nil {
			return err
		}
	}
}
