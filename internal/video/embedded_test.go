package video

import (
	"context"
	"encoding/binary"
	"image"
	"io"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

type immediateMediaDecoder struct{}

// holdingMediaDecoder publishes one frame and chunk, then waits for cancellation
// so tests can inspect a playback that is guaranteed to remain active.
type holdingMediaDecoder struct {
	ready chan struct{}
	once  sync.Once
}

// Decode signals readiness after the frame enters the playback queue, proving
// the fixture is active before Resize and Stop exercise concurrent ownership.
func (d *holdingMediaDecoder) Decode(ctx context.Context, _ io.ReadSeeker, emit func(Frame) error) error {
	if err := emit(Frame{Image: image.NewRGBA(image.Rect(0, 0, 640, 480))}); err != nil {
		return err
	}

	// Audio and video share the fixture, so only the video path closes readiness.
	d.once.Do(func() { close(d.ready) })
	<-ctx.Done()

	return ctx.Err()
}

// DecodeAudio holds the audio consumer open until Stop cancels the shared pipeline.
func (*holdingMediaDecoder) DecodeAudio(ctx context.Context, _ io.ReadSeeker, emit func(AudioChunk) error) error {
	if err := emit(AudioChunk{PCM: []byte{0, 0, 0, 0}, SampleRate: 44100, Channels: 2}); err != nil {
		return err
	}

	<-ctx.Done()

	return ctx.Err()
}

// Decode emits one frame and completes immediately to exercise natural completion cleanup.
func (immediateMediaDecoder) Decode(_ context.Context, _ io.ReadSeeker, emit func(Frame) error) error {
	return emit(Frame{Image: image.NewRGBA(image.Rect(0, 0, 640, 480))})
}

// DecodeAudio emits one valid stereo sample and completes immediately with video.
func (immediateMediaDecoder) DecodeAudio(_ context.Context, _ io.ReadSeeker, emit func(AudioChunk) error) error {
	return emit(AudioChunk{PCM: []byte{0, 0, 0, 0}, SampleRate: 44100, Channels: 2})
}

// TestEmbeddedPlaybackCompletesAndReleasesOwnership verifies natural completion
// removes both retained-render and native-audio resources.
func TestEmbeddedPlaybackCompletesAndReleasesOwnership(t *testing.T) {
	data := minimalBIK()

	var (
		composer render.Composer
		mixer    audio.Mixer
	)

	decoder := immediateMediaDecoder{}
	backend := &Embedded{Composer: &composer, Mixer: &mixer, Viewport: image.Pt(640, 480), Video: decoder, Audio: decoder}

	playback, err := backend.Play(fstest.MapFS{"intro.bik": &fstest.MapFile{Data: data}}, "intro.bik")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		// Stop is idempotent after natural completion and prevents a failed
		// assertion from leaving decoder goroutines or retained resources live.
		if err := playback.Stop(); err != nil {
			t.Errorf("stop playback: %v", err)
		}
	})

	// Completion is asynchronous; the deadline bounds a failure without imposing
	// a scheduler-sensitive fixed sleep before checking ownership diagnostics.
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

// TestEmbeddedResizeRefitsActivePlayback verifies active sessions track viewport
// changes and unregister before later backend resizes.
func TestEmbeddedResizeRefitsActivePlayback(t *testing.T) {
	var (
		composer render.Composer
		mixer    audio.Mixer
	)

	decoder := &holdingMediaDecoder{ready: make(chan struct{})}
	backend := &Embedded{Composer: &composer, Mixer: &mixer, Viewport: image.Pt(640, 480), Video: decoder, Audio: decoder}

	playback, err := backend.Play(fstest.MapFS{"intro.bik": &fstest.MapFile{Data: minimalBIK()}}, "intro.bik")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		// The holding decoder exits only through cancellation, so every test
		// failure path must retain an explicit owner that stops it.
		if err := playback.Stop(); err != nil {
			t.Errorf("stop playback: %v", err)
		}
	})

	<-decoder.ready

	if err := backend.Resize(image.Pt(800, 600)); err != nil {
		t.Fatal(err)
	}

	nodes := composer.Snapshot()
	if len(nodes) != 1 || nodes[0].ScaleX != 1.25 || nodes[0].X != 400 || nodes[0].Y != 300 {
		t.Fatalf("resized cinematic node = %#v", nodes)
	}

	if err := playback.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := backend.Resize(image.Pt(1024, 768)); err != nil {
		t.Fatalf("completed playback remained registered: %v", err)
	}
}

// TestEmbeddedPlaybackPrefersReportedAudioClock verifies wall time is only a
// fallback until the mixer supplies an authoritative played-frame count.
func TestEmbeddedPlaybackPrefersReportedAudioClock(t *testing.T) {
	var mixer audio.Mixer

	id, err := mixer.OpenPCMStream(48000, 2)
	if err != nil {
		t.Fatal(err)
	}

	playback := &embeddedPlayback{mixer: &mixer, audioID: id}

	started := time.Now().Add(-time.Second)
	if elapsed, audio := playback.mediaTime(started); audio || elapsed < 900*time.Millisecond {
		t.Fatalf("initial media clock = %v, audio=%v", elapsed, audio)
	}

	if err := mixer.ReportPCMFrames(id, 4800); err != nil {
		t.Fatal(err)
	}

	if elapsed, audio := playback.mediaTime(started); !audio || elapsed != 100*time.Millisecond {
		t.Fatalf("reported media clock = %v, audio=%v", elapsed, audio)
	}
}

// minimalBIK returns the smallest valid metadata fixture needed by Embedded.Play;
// decoder fixtures ignore payload data, so no encoded frame bytes are required.
func minimalBIK() []byte {
	data := make([]byte, 60)
	copy(data, "BIKi")

	// These offsets mirror the BIK header fields consumed by assetdecode.BIK.
	put := func(offset int, value uint32) {
		binary.LittleEndian.PutUint32(data[offset:offset+4], value)
	}
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
