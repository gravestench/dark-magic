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
)

// Embedded coordinates in-process decoding with retained presentation and PCM
// streaming. Both streams use one monotonic media clock; audio is never dropped
// and video frames that can no longer be displayed on time are discarded.
type Embedded struct {
	Composer *rendercore.Composer
	Mixer    *audiocore.Mixer
	Viewport image.Point
	Video    Decoder
	Audio    AudioDecoder
}

func (b *Embedded) Available() bool {
	return b != nil && b.Composer != nil && b.Mixer != nil && b.Video != nil && b.Audio != nil && b.Viewport.X > 0 && b.Viewport.Y > 0
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
	presenter, err := NewPresenter(b.Composer, image.Pt(int(metadata.Width), int(metadata.Height)), b.Viewport)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &embeddedPlayback{mixer: b.Mixer, presenter: presenter, cancel: cancel, snapshot: Snapshot{State: Playing}, done: make(chan struct{})}
	go p.run(ctx, data, b.Video, b.Audio)
	return p, nil
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
	videoOpen, audioOpen := true, true
	var audioEnd time.Duration
	var runErr error
	for videoOpen || audioOpen {
		select {
		case <-ctx.Done():
			videoOpen, audioOpen = false, false
		case frame, ok := <-videoFrames:
			if !ok {
				videoOpen = false
				continue
			}
			if err := waitForMediaTime(ctx, started, frame.PTS-mediaLead); err != nil {
				videoOpen, audioOpen = false, false
				continue
			}
			if time.Since(started)-frame.PTS <= lateFrameLimit {
				runErr = errors.Join(runErr, p.presenter.Present(frame.Image))
			}
		case chunk, ok := <-audioChunks:
			if !ok {
				audioOpen = false
				continue
			}
			if err := waitForMediaTime(ctx, started, chunk.PTS-mediaLead); err != nil {
				videoOpen, audioOpen = false, false
				continue
			}
			if p.audioID == (audiocore.SoundID{}) {
				p.audioID, runErr = p.mixer.OpenPCMStream(chunk.SampleRate, chunk.Channels)
			}
			if runErr == nil {
				runErr = p.mixer.WritePCM(p.audioID, chunk.PCM)
			}
			bytesPerSecond := chunk.SampleRate * chunk.Channels * 2
			if bytesPerSecond > 0 {
				audioEnd = chunk.PTS + time.Duration(float64(len(chunk.PCM))/float64(bytesPerSecond)*float64(time.Second))
			}
		}
		if runErr != nil {
			p.cancel()
			videoOpen, audioOpen = false, false
		}
	}
	for range 2 {
		if err := <-decodeErrors; err != nil && !errors.Is(err, context.Canceled) {
			runErr = errors.Join(runErr, err)
		}
	}
	if runErr == nil && p.audioID != (audiocore.SoundID{}) {
		if err := waitForMediaTime(ctx, started, audioEnd); err != nil && !errors.Is(err, context.Canceled) {
			runErr = errors.Join(runErr, err)
		}
	}
	if p.audioID != (audiocore.SoundID{}) {
		runErr = errors.Join(runErr, p.mixer.Stop(p.audioID))
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

func waitForMediaTime(ctx context.Context, started time.Time, target time.Duration) error {
	wait := target - time.Since(started)
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
