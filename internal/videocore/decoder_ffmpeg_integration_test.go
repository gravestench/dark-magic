//go:build ffmpeg

package videocore_test

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/videocore"
)

func TestFFmpegDecoderReadsRealBIK(t *testing.T) {
	mpqPath := os.Getenv("DARK_MAGIC_TEST_VIDEO_MPQ")
	if mpqPath == "" {
		t.Skip("set DARK_MAGIC_TEST_VIDEO_MPQ to d2video.mpq")
	}
	archive, err := content.MPQ(mpqPath)
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(archive, "data/local/video/New_Bliz640x480.bik")
	if err != nil {
		t.Fatal(err)
	}
	stop := errors.New("enough frames decoded")
	frames := 0
	err = (videocore.FFmpegDecoder{}).Decode(context.Background(), bytes.NewReader(data), func(frame videocore.Frame) error {
		frames++
		if frame.Image.Bounds().Dx() != 640 || frame.Image.Bounds().Dy() != 480 {
			t.Fatalf("frame size = %v", frame.Image.Bounds().Size())
		}
		if frames == 3 {
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) {
		t.Fatalf("decode = %v", err)
	}
}

func TestFFmpegDecoderReadsRealBIKAudio(t *testing.T) {
	mpqPath := os.Getenv("DARK_MAGIC_TEST_VIDEO_MPQ")
	if mpqPath == "" {
		t.Skip("set DARK_MAGIC_TEST_VIDEO_MPQ to d2video.mpq")
	}
	archive, err := content.MPQ(mpqPath)
	if err != nil {
		t.Fatal(err)
	}
	data, err := fs.ReadFile(archive, "data/local/video/New_Bliz640x480.bik")
	if err != nil {
		t.Fatal(err)
	}
	stop := errors.New("enough audio decoded")
	chunks := 0
	err = (videocore.FFmpegDecoder{}).DecodeAudio(context.Background(), bytes.NewReader(data), func(chunk videocore.AudioChunk) error {
		chunks++
		if len(chunk.PCM) == 0 || chunk.SampleRate <= 0 || chunk.Channels != 2 {
			t.Fatalf("invalid PCM chunk: bytes=%d rate=%d channels=%d", len(chunk.PCM), chunk.SampleRate, chunk.Channels)
		}
		if chunks == 3 {
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) {
		t.Fatalf("decode audio = %v", err)
	}
}
