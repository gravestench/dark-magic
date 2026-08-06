//go:build ffmpeg

package videocore_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/content"
	"github.com/gravestench/dark-magic/internal/rendercore"
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
	var previousPTS time.Duration
	err = (videocore.FFmpegDecoder{}).Decode(context.Background(), bytes.NewReader(data), func(frame videocore.Frame) error {
		frames++
		if frame.PTS < 0 {
			t.Fatalf("negative frame PTS %s", frame.PTS)
		}
		if frame.Image.Bounds().Dx() != 640 || frame.Image.Bounds().Dy() != 480 {
			t.Fatalf("frame size = %v", frame.Image.Bounds().Size())
		}
		if frames > 1 && frame.PTS <= previousPTS {
			t.Fatalf("frame PTS %s did not advance after %s", frame.PTS, previousPTS)
		}
		previousPTS = frame.PTS
		if frames == 3 {
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) {
		t.Fatalf("decode = %v", err)
	}
}

func TestEmbeddedBackendPresentsRealBIKFrame(t *testing.T) {
	mpqPath := os.Getenv("DARK_MAGIC_TEST_VIDEO_MPQ")
	if mpqPath == "" {
		t.Skip("set DARK_MAGIC_TEST_VIDEO_MPQ to d2video.mpq")
	}
	archive, err := content.MPQ(mpqPath)
	if err != nil {
		t.Fatal(err)
	}
	var composer rendercore.Composer
	var mixer audiocore.Mixer
	decoder := videocore.FFmpegDecoder{}
	backend := &videocore.Embedded{Composer: &composer, Mixer: &mixer, Viewport: image.Pt(640, 480), Video: decoder, Audio: decoder}
	playback, err := backend.Play(archive, "data/local/video/New_Bliz640x480.bik")
	if err != nil {
		t.Fatal(err)
	}
	defer playback.Stop()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		nodes := composer.Snapshot()
		if len(nodes) == 1 {
			resource, err := composer.ResourceSnapshot(nodes[0].Resource)
			if err == nil {
				frame := resource.Payload.(image.Image)
				for y := 0; y < frame.Bounds().Dy(); y += 40 {
					for x := 0; x < frame.Bounds().Dx(); x += 40 {
						r, g, b, _ := frame.At(x, y).RGBA()
						if r != 0 || g != 0 || b != 0 {
							return
						}
					}
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("embedded backend did not present a non-black frame: %#v, playback %#v", composer.Diagnostics(), playback.Snapshot())
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
	var previousPTS time.Duration
	err = (videocore.FFmpegDecoder{}).DecodeAudio(context.Background(), bytes.NewReader(data), func(chunk videocore.AudioChunk) error {
		chunks++
		if len(chunk.PCM) == 0 || chunk.SampleRate <= 0 || chunk.Channels != 2 {
			t.Fatalf("invalid PCM chunk: bytes=%d rate=%d channels=%d", len(chunk.PCM), chunk.SampleRate, chunk.Channels)
		}
		if chunk.PTS < 0 || chunks > 1 && (chunk.PTS <= previousPTS || chunk.PTS-previousPTS > time.Second) {
			t.Fatalf("audio PTS %s did not advance after %s", chunk.PTS, previousPTS)
		}
		previousPTS = chunk.PTS
		if chunks == 100 {
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) {
		t.Fatalf("decode audio = %v", err)
	}
}
