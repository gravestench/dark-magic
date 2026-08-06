package videocore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io/fs"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/audiocore"
	"github.com/gravestench/dark-magic/internal/rendercore"
	"github.com/gravestench/dark-magic/pkg/assetdecode"
)

const (
	decodedVideoQueue = 4
	decodedAudioQueue = 16
	mediaLead         = 80 * time.Millisecond
	lateFrameLimit    = 150 * time.Millisecond
	audioClockStall   = 2 * time.Second
)

// Embedded coordinates in-process decoding with retained presentation and PCM
// streaming. Both streams use one monotonic media clock; audio is never dropped
// and video frames that can no longer be displayed on time are discarded.
type Embedded struct {
	mu       sync.Mutex
	Composer *rendercore.Composer
	Mixer    *audiocore.Mixer
	Viewport image.Point
	Video    Decoder
	Audio    AudioDecoder
	active   map[*embeddedPlayback]struct{}
}

func (b *Embedded) Available() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Composer != nil && b.Mixer != nil && b.Video != nil && b.Audio != nil && b.Viewport.X > 0 && b.Viewport.Y > 0
}

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
	p := &embeddedPlayback{mixer: b.Mixer, presenter: presenter, cancel: cancel, snapshot: Snapshot{State: Playing}, done: make(chan struct{})}
	b.mu.Lock()
	if b.active == nil {
		b.active = make(map[*embeddedPlayback]struct{})
	}
	b.active[p] = struct{}{}
	p.onDone = func() {
		b.mu.Lock()
		delete(b.active, p)
		b.mu.Unlock()
	}
	b.mu.Unlock()
	go p.run(ctx, data, b.Video, b.Audio)
	return p, nil
}

// Resize refits every active cinematic to the new render viewport.
func (b *Embedded) Resize(viewport image.Point) error {
	if viewport.X <= 0 || viewport.Y <= 0 {
		return fmt.Errorf("videocore: invalid viewport %v", viewport)
	}
	b.mu.Lock()
	b.Viewport = viewport
	active := make([]*embeddedPlayback, 0, len(b.active))
	for playback := range b.active {
		active = append(active, playback)
	}
	b.mu.Unlock()
	var resizeErr error
	for _, playback := range active {
		resizeErr = errors.Join(resizeErr, playback.presenter.Resize(viewport))
	}
	return resizeErr
}

type embeddedPlayback struct {
	mu        sync.RWMutex
	mixer     *audiocore.Mixer
	presenter *Presenter
	cancel    context.CancelFunc
	audioID   audiocore.SoundID
	snapshot  Snapshot
	done      chan struct{}
	stopOnce  sync.Once
	onDone    func()
}

func (p *embeddedPlayback) Snapshot() Snapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.snapshot
}

func (p *embeddedPlayback) Stop() error {
	p.stopOnce.Do(func() {
		p.mu.Lock()
		if p.snapshot.State == Playing {
			p.snapshot = Snapshot{State: Stopped}
			p.cancel()
		}
		p.mu.Unlock()
	})
	<-p.done
	return nil
}

func (p *embeddedPlayback) run(ctx context.Context, data []byte, video Decoder, audio AudioDecoder) {
	defer close(p.done)
	defer func() {
		if p.onDone != nil {
			p.onDone()
		}
	}()
	videoFrames := make(chan Frame, decodedVideoQueue)
	audioChunks := make(chan AudioChunk, decodedAudioQueue)
	decodeErrors := make(chan error, 2)
	go func() {
		err := video.Decode(ctx, bytes.NewReader(data), func(frame Frame) error {
			select {
			case videoFrames <- frame:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		close(videoFrames)
		decodeErrors <- err
	}()
	go func() {
		err := audio.DecodeAudio(ctx, bytes.NewReader(data), func(chunk AudioChunk) error {
			select {
			case audioChunks <- chunk:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		close(audioChunks)
		decodeErrors <- err
	}()

	started := time.Now()
	presentationErrors := make(chan error, 2)
	go func() {
		var presentErr error
		for frame := range videoFrames {
			if err := p.waitForMediaTime(ctx, started, frame.PTS-mediaLead); err != nil {
				presentErr = err
				break
			}
			mediaTime, _ := p.mediaTime(started)
			if mediaTime-frame.PTS <= lateFrameLimit {
				if err := p.presenter.Present(frame.Image); err != nil {
					presentErr = err
					break
				}
			}
		}
		presentationErrors <- presentErr
	}()
	go func() {
		var streamErr error
		var audioEnd time.Duration
		var audioID audiocore.SoundID
		for chunk := range audioChunks {
			if err := p.waitForMediaTime(ctx, started, chunk.PTS-mediaLead); err != nil {
				streamErr = err
				break
			}
			if audioID == (audiocore.SoundID{}) {
				audioID, streamErr = p.mixer.OpenPCMStream(chunk.SampleRate, chunk.Channels)
				p.setAudioID(audioID)
			}
			if streamErr == nil {
				streamErr = p.mixer.WritePCM(audioID, chunk.PCM)
			}
			if streamErr != nil {
				break
			}
			bytesPerSecond := chunk.SampleRate * chunk.Channels * 2
			if bytesPerSecond > 0 {
				audioEnd = chunk.PTS + time.Duration(float64(len(chunk.PCM))/float64(bytesPerSecond)*float64(time.Second))
			}
		}
		if streamErr == nil && audioID != (audiocore.SoundID{}) {
			streamErr = p.waitForMediaTime(ctx, started, audioEnd)
		}
		presentationErrors <- streamErr
	}()

	var runErr error
	for range 2 {
		if err := <-presentationErrors; err != nil && !errors.Is(err, context.Canceled) {
			runErr = errors.Join(runErr, err)
			p.cancel()
		}
	}
	for range 2 {
		if err := <-decodeErrors; err != nil && !errors.Is(err, context.Canceled) {
			runErr = errors.Join(runErr, err)
		}
	}
	if audioID := p.getAudioID(); audioID != (audiocore.SoundID{}) {
		runErr = errors.Join(runErr, p.mixer.Stop(audioID))
	}
	runErr = errors.Join(runErr, p.presenter.Close())
	p.mu.Lock()
	if p.snapshot.State == Playing {
		if runErr != nil {
			p.snapshot = Snapshot{State: Failed, Error: runErr.Error()}
		} else {
			p.snapshot = Snapshot{State: Complete}
		}
	}
	p.mu.Unlock()
}

func (p *embeddedPlayback) setAudioID(id audiocore.SoundID) {
	p.mu.Lock()
	p.audioID = id
	p.mu.Unlock()
}

func (p *embeddedPlayback) getAudioID() audiocore.SoundID {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.audioID
}

func (p *embeddedPlayback) mediaTime(started time.Time) (time.Duration, bool) {
	if id := p.getAudioID(); id != (audiocore.SoundID{}) {
		if elapsed, available := p.mixer.PCMTime(id); available {
			return elapsed, true
		}
	}
	return time.Since(started), false
}

func (p *embeddedPlayback) waitForMediaTime(ctx context.Context, started time.Time, target time.Duration) error {
	if target <= 0 {
		return nil
	}
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	lastMedia := time.Duration(-1)
	lastAdvance := time.Now()
	for {
		current, audioClock := p.mediaTime(started)
		if current >= target {
			return nil
		}
		if current > lastMedia {
			lastMedia = current
			lastAdvance = time.Now()
		} else if audioClock && time.Since(lastAdvance) >= audioClockStall {
			return errors.New("videocore: cinematic audio device clock stalled")
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
