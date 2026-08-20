//go:build ffmpeg

package video

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

// Decode opens the preferred video stream, preserves FFmpeg allocation cleanup
// order, and emits RGBA frames synchronously so callback backpressure is honored.
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

	receiver := videoFrameReceiver{
		ctx:      ctx,
		decoder:  decoder,
		decoded:  decoded,
		timeBase: stream.TimeBase().Float64(),
		emit:     emit,
	}
	defer receiver.close()

	return decodeVideoPackets(ctx, format, stream, decoder, packet, receiver.receive)
}

// videoFrameReceiver retains the lazily allocated RGBA scaler across frames;
// avoiding per-frame scaler construction keeps the decode hot path unchanged.
type videoFrameReceiver struct {
	ctx      context.Context
	decoder  *astiav.CodecContext
	decoded  *astiav.Frame
	scaler   *astiav.SoftwareScaleContext
	timeBase float64
	emit     func(Frame) error
}

// close frees the optional scaler before Decode's frame, packet, decoder, and
// input defers run, preserving native dependency cleanup order.
func (r *videoFrameReceiver) close() {
	if r.scaler != nil {
		r.scaler.Free()
	}
}

// receive drains every frame currently available from the decoder and treats
// EAGAIN and EOF as normal packet-boundary conditions rather than failures.
func (r *videoFrameReceiver) receive() error {
	for {
		r.decoded.Unref()

		if err := r.decoder.ReceiveFrame(r.decoded); err != nil {
			if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
				return nil
			}

			return fmt.Errorf("videocore: receive video frame: %w", err)
		}

		if err := r.ctx.Err(); err != nil {
			return err
		}

		if err := r.ensureScaler(); err != nil {
			return err
		}

		output, err := r.convertFrame()
		if err != nil {
			return err
		}

		pts := time.Duration(float64(r.decoded.Pts()) * r.timeBase * float64(time.Second))
		if err := r.emit(Frame{Image: output, PTS: pts}); err != nil {
			return err
		}
	}
}

// ensureScaler binds the converter to the first decoded frame's geometry and
// pixel format, matching FFmpeg's established stream-level conversion behavior.
func (r *videoFrameReceiver) ensureScaler() error {
	if r.scaler != nil {
		return nil
	}

	flags := astiav.NewSoftwareScaleContextFlags(astiav.SoftwareScaleContextFlagBilinear)

	scaler, err := astiav.CreateSoftwareScaleContext(
		r.decoded.Width(),
		r.decoded.Height(),
		r.decoded.PixelFormat(),
		r.decoded.Width(),
		r.decoded.Height(),
		astiav.PixelFormatRgba,
		flags,
	)
	if err != nil {
		return fmt.Errorf("videocore: create RGBA converter: %w", err)
	}

	r.scaler = scaler

	return nil
}

// convertFrame copies one native RGBA frame into Go-owned image storage before
// freeing the native frame, allowing emit callbacks to retain the image safely.
func (r *videoFrameReceiver) convertFrame() (image.Image, error) {
	rgba := astiav.AllocFrame()
	if rgba == nil {
		return nil, errors.New("videocore: allocate RGBA frame")
	}

	if err := r.scaler.ScaleFrame(r.decoded, rgba); err != nil {
		rgba.Free()

		return nil, fmt.Errorf("videocore: convert video frame: %w", err)
	}

	output := image.NewNRGBA(image.Rect(0, 0, rgba.Width(), rgba.Height()))
	if err := rgba.Data().ToImage(output); err != nil {
		rgba.Free()

		return nil, fmt.Errorf("videocore: copy video frame: %w", err)
	}

	rgba.Free()

	return output, nil
}

// decodeVideoPackets filters the container to the selected stream, drains after
// each accepted packet, and flushes delayed frames once container EOF is reached.
func decodeVideoPackets(
	ctx context.Context,
	format *astiav.FormatContext,
	stream *astiav.Stream,
	decoder *astiav.CodecContext,
	packet *astiav.Packet,
	receive func() error,
) error {
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
