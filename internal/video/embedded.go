package video

import (
	"context"
	"errors"
	"fmt"
	"image"
	"io/fs"
	"sync"

	"github.com/gravestench/dark-magic/internal/assets/decode"
	"github.com/gravestench/dark-magic/internal/audio"
	"github.com/gravestench/dark-magic/internal/presentation/render"
)

// Embedded coordinates in-process decoding with retained presentation and PCM
// streaming. Both streams use the same media-time selection; audio chunks remain
// ordered while video frames that can no longer be displayed on time are discarded.
type Embedded struct {
	mu       sync.Mutex
	Composer *render.Composer
	Mixer    *audio.Mixer
	Viewport image.Point
	Video    Decoder
	Audio    AudioDecoder
	active   map[*embeddedPlayback]struct{}
}

// Available reports whether every in-process playback dependency and the current
// viewport are valid; callers can use false as a non-fatal capability decision.
func (b *Embedded) Available() bool {
	if b == nil {
		return false
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	return b.Composer != nil &&
		b.Mixer != nil &&
		b.Video != nil &&
		b.Audio != nil &&
		b.Viewport.X > 0 &&
		b.Viewport.Y > 0
}

// Play validates the BIK header before acquiring presentation resources, then
// registers the asynchronous playback so later viewport changes can refit it.
func (b *Embedded) Play(source fs.FS, path string) (Playback, error) {
	if !b.Available() {
		return nil, ErrUnavailable
	}

	data, err := fs.ReadFile(source, path)
	if err != nil {
		return nil, fmt.Errorf("videocore: read %q: %w", path, err)
	}

	metadata, err := assetdecode.BIK(data)
	if err != nil {
		return nil, err
	}

	b.mu.Lock()
	viewport := b.Viewport
	b.mu.Unlock()

	presenter, err := NewPresenter(b.Composer, image.Pt(int(metadata.Width), int(metadata.Height)), viewport)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	playback := &embeddedPlayback{
		mixer:     b.Mixer,
		presenter: presenter,
		cancel:    cancel,
		snapshot:  Snapshot{State: Playing},
		done:      make(chan struct{}),
	}

	b.registerPlayback(playback)

	// Registration precedes launch so an immediate Resize cannot miss a live
	// playback whose presenter has already entered the retained composition.
	go playback.run(ctx, data, b.Video, b.Audio)

	return playback, nil
}

// Resize refits every active cinematic to the new render viewport.
func (b *Embedded) Resize(viewport image.Point) error {
	if viewport.X <= 0 || viewport.Y <= 0 {
		return fmt.Errorf("videocore: invalid viewport %v", viewport)
	}

	active := b.setViewportAndSnapshotPlaybacks(viewport)

	var resizeErr error
	for _, playback := range active {
		resizeErr = errors.Join(resizeErr, playback.presenter.Resize(viewport))
	}

	return resizeErr
}

// registerPlayback publishes a playback under the same lock used by Resize and
// installs removal before launch, preventing a completed session from leaking.
func (b *Embedded) registerPlayback(playback *embeddedPlayback) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.active == nil {
		b.active = make(map[*embeddedPlayback]struct{})
	}

	b.active[playback] = struct{}{}
	playback.onDone = func() {
		b.mu.Lock()
		delete(b.active, playback)
		b.mu.Unlock()
	}
}

// setViewportAndSnapshotPlaybacks updates the viewport atomically with a copy
// of active ownership; Presenter calls remain outside the backend lock.
func (b *Embedded) setViewportAndSnapshotPlaybacks(viewport image.Point) []*embeddedPlayback {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Viewport = viewport
	active := make([]*embeddedPlayback, 0, len(b.active))

	for playback := range b.active {
		active = append(active, playback)
	}

	return active
}
