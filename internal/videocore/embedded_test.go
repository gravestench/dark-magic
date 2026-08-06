package videocore

import (
	"context"
	"encoding/binary"
	"image"
	"io"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/rendercore"
)

type immediateMediaDecoder struct{}

type holdingMediaDecoder struct {
	ready chan struct{}
	once  sync.Once
}

func (d *holdingMediaDecoder) Decode(ctx context.Context, _ io.ReadSeeker, emit func(Frame) error) error {
	if err := emit(Frame{Image: image.NewRGBA(image.Rect(0, 0, 640, 480))}); err != nil {
		return err
	}
	d.once.Do(func() { close(d.ready) })
	<-ctx.Done()
	return ctx.Err()
}

func (*holdingMediaDecoder) DecodeAudio(ctx context.Context, _ io.ReadSeeker, emit func(AudioChunk) error) error {
	if err := emit(AudioChunk{PCM: []byte{0, 0, 0, 0}, SampleRate: 44100, Channels: 2}); err != nil {
		return err
	}
	<-ctx.Done()
	return ctx.Err()
}

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

func TestEmbeddedResizeRefitsActivePlayback(t *testing.T) {
	var composer rendercore.Composer
	var mixer audiocore.Mixer
	decoder := &holdingMediaDecoder{ready: make(chan struct{})}
	backend := &Embedded{Composer: &composer, Mixer: &mixer, Viewport: image.Pt(640, 480), Video: decoder, Audio: decoder}
	playback, err := backend.Play(fstest.MapFS{"intro.bik": &fstest.MapFile{Data: minimalBIK()}}, "intro.bik")
	if err != nil {
		t.Fatal(err)
	}
	<-decoder.ready
	if err := backend.Resize(image.Pt(800, 600)); err != nil {
		t.Fatal(err)
	}
	nodes := composer.Snapshot()
	if len(nodes) != 1 || nodes[0].ScaleX != 1.25 || nodes[0].X != 0 || nodes[0].Y != 0 {
		t.Fatalf("resized cinematic node = %#v", nodes)
	}
	if err := playback.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := backend.Resize(image.Pt(1024, 768)); err != nil {
		t.Fatalf("completed playback remained registered: %v", err)
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
