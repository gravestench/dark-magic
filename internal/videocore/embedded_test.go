package videocore

import (
	"context"
	"encoding/binary"
	"image"
	"io"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/rendercore"
)

type immediateMediaDecoder struct{}

func (immediateMediaDecoder) Decode(_ context.Context, _ io.ReadSeeker, emit func(Frame) error) error {
	return emit(Frame{Image: image.NewRGBA(image.Rect(0, 0, 640, 480))})
}

func (immediateMediaDecoder) DecodeAudio(_ context.Context, _ io.ReadSeeker, emit func(AudioChunk) error) error {
	return emit(AudioChunk{PCM: []byte{0, 0, 0, 0}, SampleRate: 44100, Channels: 2})
}

func TestEmbeddedPlaybackCompletesAndReleasesOwnership(t *testing.T) {
	data := minimalBIK()
	var composer rendercore.Composer
	var mixer audiocore.Mixer
	decoder := immediateMediaDecoder{}
	backend := &Embedded{Composer: &composer, Mixer: &mixer, Viewport: image.Pt(640, 480), Video: decoder, Audio: decoder}
	playback, err := backend.Play(fstest.MapFS{"intro.bik": &fstest.MapFile{Data: data}}, "intro.bik")
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for playback.Snapshot().State == Playing && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := playback.Snapshot(); got.State != Complete {
		t.Fatalf("playback = %#v", got)
	}
	if got := composer.Diagnostics(); got.ActiveNodes != 0 || got.ActiveResources != 0 {
		t.Fatalf("render ownership leaked: %#v", got)
	}
	if got := mixer.Diagnostics(); got.Active != 0 {
		t.Fatalf("audio ownership leaked: %#v", got)
	}
}

func minimalBIK() []byte {
	data := make([]byte, 60)
	copy(data, "BIKi")
	put := func(offset int, value uint32) { binary.LittleEndian.PutUint32(data[offset:offset+4], value) }
	put(4, uint32(len(data)-8))
	put(8, 1)
	put(12, 16)
	put(16, 1)
	put(20, 640)
	put(24, 480)
	put(28, 24)
	put(32, 1)
	put(40, 1)
	put(44, 4096)
	binary.LittleEndian.PutUint16(data[48:50], 44100)
	binary.LittleEndian.PutUint16(data[50:52], 0xe000)
	put(52, 7)
	return data
}
