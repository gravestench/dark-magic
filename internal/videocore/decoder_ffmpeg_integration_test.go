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
