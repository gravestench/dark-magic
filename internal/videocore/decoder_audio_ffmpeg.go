//go:build ffmpeg

package videocore

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
	timeBase := stream.TimeBase().Float64()
	receive := func() error {
		for {
			decoded.Unref()
			if err := decoder.ReceiveFrame(decoded); err != nil {
				if errors.Is(err, astiav.ErrEagain) || errors.Is(err, astiav.ErrEof) {
					return nil
				}
				return fmt.Errorf("videocore: receive audio frame: %w", err)
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			output := astiav.AllocFrame()
			if output == nil {
				return errors.New("videocore: allocate PCM frame")
			}
			output.SetChannelLayout(astiav.ChannelLayoutStereo)
			output.SetSampleFormat(astiav.SampleFormatS16)
			output.SetSampleRate(decoded.SampleRate())
			if err := resampler.ConvertFrame(decoded, output); err != nil {
				output.Free()
				return fmt.Errorf("videocore: convert audio frame: %w", err)
			}
			size, err := output.SamplesBufferSize(1)
			if err != nil {
				output.Free()
				return fmt.Errorf("videocore: size PCM frame: %w", err)
			}
			pcm := make([]byte, size)
			written, err := output.SamplesCopyToBuffer(pcm, 1)
			output.Free()
			if err != nil {
				return fmt.Errorf("videocore: copy PCM frame: %w", err)
			}
			if err := emit(AudioChunk{PCM: pcm[:written], PTS: time.Duration(float64(decoded.Pts()) * timeBase * float64(time.Second)), SampleRate: decoded.SampleRate(), Channels: 2}); err != nil {
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
