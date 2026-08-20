//go:build ffmpeg

package video

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/asticode/go-astiav"
)

// DecodeAudio decodes the preferred audio stream to interleaved stereo S16
// PCM. Keeping this normalization here lets the renderer audio backend remain
// independent of FFmpeg sample formats and channel layouts.
func (FFmpegDecoder) DecodeAudio(ctx context.Context, input io.ReadSeeker, emit func(AudioChunk) error) error {
	if input == nil || emit == nil {
		return errors.New("videocore: decoder input and audio callback are required")
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

	stream, codec, err := format.FindBestStream(astiav.MediaTypeAudio, -1, -1)
	if err != nil {
		return fmt.Errorf("videocore: find audio stream: %w", err)
	}

	decoder := astiav.AllocCodecContext(codec)
	if decoder == nil {
		return errors.New("videocore: allocate FFmpeg audio decoder")
	}
	defer decoder.Free()

	if err := stream.CodecParameters().ToCodecContext(decoder); err != nil {
		return fmt.Errorf("videocore: configure audio decoder: %w", err)
	}

	if err := decoder.Open(codec, nil); err != nil {
		return fmt.Errorf("videocore: open audio decoder: %w", err)
	}

	packet, decoded := astiav.AllocPacket(), astiav.AllocFrame()

	resampler := astiav.AllocSoftwareResampleContext()
	if packet == nil || decoded == nil || resampler == nil {
		if packet != nil {
			packet.Free()
		}

		if decoded != nil {
			decoded.Free()
		}

		if resampler != nil {
			resampler.Free()
		}

		return errors.New("videocore: allocate FFmpeg audio state")
	}
	defer packet.Free()
	defer decoded.Free()
	defer resampler.Free()

	receiver := audioFrameReceiver{
		ctx:       ctx,
		decoder:   decoder,
		decoded:   decoded,
		resampler: resampler,
		timeBase:  stream.TimeBase().Float64(),
		emit:      emit,
	}

	return decodeAudioPackets(ctx, format, stream, decoder, packet, receiver.receive)
}

// audioFrameReceiver retains timestamp fallback state across decoded frames so
// streams without packet PTS remain continuous after stereo S16 normalization.
type audioFrameReceiver struct {
	ctx       context.Context
	decoder   *astiav.CodecContext
	decoded   *astiav.Frame
	resampler *astiav.SoftwareResampleContext
	timeBase  float64
	nextPTS   time.Duration
	emit      func(AudioChunk) error
}

// receive drains available decoded frames and treats EAGAIN and EOF as normal
// packet boundaries while preserving synchronous callback backpressure.
func (r *audioFrameReceiver) receive() error {
	for {
		r.decoded.Unref()

		if err := r.decoder.ReceiveFrame(r.decoded); err != nil {
			if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
				return nil
			}

			return fmt.Errorf("videocore: receive audio frame: %w", err)
		}

		if err := r.ctx.Err(); err != nil {
			return err
		}

		chunk, err := r.convertFrame()
		if err != nil {
			return err
		}

		if err := r.emit(chunk); err != nil {
			return err
		}
	}
}

// convertFrame copies normalized native PCM into Go-owned storage, then advances
// the fallback timestamp by the exact emitted byte duration.
func (r *audioFrameReceiver) convertFrame() (AudioChunk, error) {
	output := astiav.AllocFrame()
	if output == nil {
		return AudioChunk{}, errors.New("videocore: allocate PCM frame")
	}

	output.SetChannelLayout(astiav.ChannelLayoutStereo)
	output.SetSampleFormat(astiav.SampleFormatS16)
	output.SetSampleRate(r.decoded.SampleRate())

	if err := r.resampler.ConvertFrame(r.decoded, output); err != nil {
		output.Free()

		return AudioChunk{}, fmt.Errorf("videocore: convert audio frame: %w", err)
	}

	size, err := output.SamplesBufferSize(1)
	if err != nil {
		output.Free()

		return AudioChunk{}, fmt.Errorf("videocore: size PCM frame: %w", err)
	}

	pcm := make([]byte, size)
	written, err := output.SamplesCopyToBuffer(pcm, 1)
	output.Free()

	if err != nil {
		return AudioChunk{}, fmt.Errorf("videocore: copy PCM frame: %w", err)
	}

	pts := r.nextPTS
	if r.decoded.Pts() != astiav.NoPtsValue {
		pts = time.Duration(float64(r.decoded.Pts()) * r.timeBase * float64(time.Second))
	}

	duration := time.Duration(
		float64(written) / float64(r.decoded.SampleRate()*2*2) * float64(time.Second),
	)
	r.nextPTS = pts + duration

	return AudioChunk{
		PCM:        pcm[:written],
		PTS:        pts,
		SampleRate: r.decoded.SampleRate(),
		Channels:   2,
	}, nil
}

// decodeAudioPackets filters the selected stream, drains after each packet, and
// flushes delayed samples at EOF without changing FFmpeg error wrapping.
func decodeAudioPackets(
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
				return fmt.Errorf("videocore: flush audio decoder: %w", err)
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
			return fmt.Errorf("videocore: send audio packet: %w", err)
		}

		if err := receive(); err != nil {
			return err
		}
	}
}
