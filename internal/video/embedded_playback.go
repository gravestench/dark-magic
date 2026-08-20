package video

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gravestench/dark-magic/internal/audio"
)

const (
	decodedVideoQueue = 4
	decodedAudioQueue = 16
	mediaLead         = 80 * time.Millisecond
	lateFrameLimit    = 150 * time.Millisecond
	audioClockStall   = 2 * time.Second
)

// embeddedPlayback owns one decoding pipeline and publishes a lock-protected
// lifecycle snapshot while Stop and the worker race to establish the final state.
type embeddedPlayback struct {
	mu        sync.RWMutex
	mixer     *audio.Mixer
	presenter *Presenter
	cancel    context.CancelFunc
	audioID   audio.SoundID
	snapshot  Snapshot
	done      chan struct{}
	stopOnce  sync.Once
	onDone    func()
}

// Snapshot copies the current lifecycle state so polling never observes a
// partially updated failure description.
func (p *embeddedPlayback) Snapshot() Snapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.snapshot
}

// Stop establishes the explicit Stopped state once, cancels decoding while
// holding the state lock, and waits until native audio/render ownership is gone.
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

// run starts video and audio decoding before their consumers, joins presentation
// first, then decoder completion, and releases audio before renderer resources.
func (p *embeddedPlayback) run(ctx context.Context, data []byte, video Decoder, audioDecoder AudioDecoder) {
	defer close(p.done)
	defer p.notifyDone()

	videoFrames := make(chan Frame, decodedVideoQueue)
	audioChunks := make(chan AudioChunk, decodedAudioQueue)
	decodeErrors := make(chan error, 2)

	go decodeVideo(ctx, data, video, videoFrames, decodeErrors)
	go decodeAudio(ctx, data, audioDecoder, audioChunks, decodeErrors)

	started := time.Now()
	presentationErrors := make(chan error, 2)

	go p.presentVideo(ctx, started, videoFrames, presentationErrors)
	go p.streamAudio(ctx, started, audioChunks, presentationErrors)

	runErr := p.joinPresentationErrors(presentationErrors)
	runErr = errors.Join(runErr, joinDecodeErrors(decodeErrors))
	runErr = errors.Join(runErr, p.releaseResources())

	p.publishRunResult(runErr)
}

// notifyDone unregisters the playback before run closes done, preserving the
// guarantee that Stop returns only after backend ownership is released.
func (p *embeddedPlayback) notifyDone() {
	if p.onDone != nil {
		p.onDone()
	}
}

// decodeVideo gives the decoder its own reader and closes the frame stream
// before reporting completion, distinguishing normal stream exhaustion from errors.
func decodeVideo(
	ctx context.Context,
	data []byte,
	decoder Decoder,
	frames chan<- Frame,
	decodeErrors chan<- error,
) {
	err := decoder.Decode(ctx, bytes.NewReader(data), func(frame Frame) error {
		select {
		case frames <- frame:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	close(frames)

	decodeErrors <- err
}

// decodeAudio gives audio an independent reader and closes the chunk stream
// before reporting completion; the bounded channel preserves decoder backpressure.
func decodeAudio(
	ctx context.Context,
	data []byte,
	decoder AudioDecoder,
	chunks chan<- AudioChunk,
	decodeErrors chan<- error,
) {
	err := decoder.DecodeAudio(ctx, bytes.NewReader(data), func(chunk AudioChunk) error {
		select {
		case chunks <- chunk:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	close(chunks)

	decodeErrors <- err
}

// presentVideo schedules frames slightly ahead of media time and discards only
// frames beyond the established lateness threshold, keeping playback responsive.
func (p *embeddedPlayback) presentVideo(
	ctx context.Context,
	started time.Time,
	frames <-chan Frame,
	presentationErrors chan<- error,
) {
	var presentErr error

	for frame := range frames {
		if err := p.waitForMediaTime(ctx, started, frame.PTS-mediaLead); err != nil {
			presentErr = err
			break
		}

		mediaTime, _ := p.mediaTime(started)

		// Video may be dropped when late; audio cannot, because it is the media
		// clock and dropping PCM would change duration and synchronization.
		if mediaTime-frame.PTS > lateFrameLimit {
			continue
		}

		if err := p.presenter.Present(frame.Image); err != nil {
			presentErr = err
			break
		}
	}

	presentationErrors <- presentErr
}

// streamAudio opens the PCM stream lazily from decoded metadata, writes chunks
// in order until completion or failure, then waits for the final accepted sample.
func (p *embeddedPlayback) streamAudio(
	ctx context.Context,
	started time.Time,
	chunks <-chan AudioChunk,
	presentationErrors chan<- error,
) {
	var (
		streamErr error
		audioEnd  time.Duration
		audioID   audio.SoundID
	)

	for chunk := range chunks {
		if err := p.waitForMediaTime(ctx, started, chunk.PTS-mediaLead); err != nil {
			streamErr = err
			break
		}

		if audioID == (audio.SoundID{}) {
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
			audioEnd = chunk.PTS + pcmDuration(len(chunk.PCM), bytesPerSecond)
		}
	}

	if streamErr == nil && audioID != (audio.SoundID{}) {
		streamErr = p.waitForMediaTime(ctx, started, audioEnd)
	}

	presentationErrors <- streamErr
}

// pcmDuration converts interleaved S16 byte length to media duration with the
// same floating-point calculation used by the scheduler.
func pcmDuration(byteCount, bytesPerSecond int) time.Duration {
	return time.Duration(float64(byteCount) / float64(bytesPerSecond) * float64(time.Second))
}

// joinPresentationErrors waits for both consumers and cancels the shared context
// after any substantive error so blocked decoder callbacks can unwind.
func (p *embeddedPlayback) joinPresentationErrors(results <-chan error) error {
	var runErr error

	for range 2 {
		if err := <-results; err != nil && !errors.Is(err, context.Canceled) {
			runErr = errors.Join(runErr, err)

			p.cancel()
		}
	}

	return runErr
}

// joinDecodeErrors waits for both decoders after consumers have finished and
// excludes cancellation caused by Stop or a presentation failure.
func joinDecodeErrors(results <-chan error) error {
	var runErr error

	for range 2 {
		if err := <-results; err != nil && !errors.Is(err, context.Canceled) {
			runErr = errors.Join(runErr, err)
		}
	}

	return runErr
}

// releaseResources stops native audio before closing the retained presenter,
// retaining the original cleanup order and joining both failures.
func (p *embeddedPlayback) releaseResources() error {
	var runErr error

	if audioID := p.getAudioID(); audioID != (audio.SoundID{}) {
		runErr = errors.Join(runErr, p.mixer.Stop(audioID))
	}

	return errors.Join(runErr, p.presenter.Close())
}

// publishRunResult records natural completion or failure only when Stop has not
// already established the externally visible final state.
func (p *embeddedPlayback) publishRunResult(runErr error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.snapshot.State != Playing {
		return
	}

	if runErr != nil {
		p.snapshot = Snapshot{State: Failed, Error: runErr.Error()}

		return
	}

	p.snapshot = Snapshot{State: Complete}
}

// setAudioID publishes native audio ownership so Stop and final cleanup observe
// the stream only after OpenPCMStream returns.
func (p *embeddedPlayback) setAudioID(id audio.SoundID) {
	p.mu.Lock()
	p.audioID = id
	p.mu.Unlock()
}

// getAudioID reads the current native stream handle without exposing mutable
// playback state to scheduling and cleanup code.
func (p *embeddedPlayback) getAudioID() audio.SoundID {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.audioID
}

// mediaTime prefers the mixer-reported PCM clock once available and falls
// back to wall time before audio opens or when the device cannot report progress.
func (p *embeddedPlayback) mediaTime(started time.Time) (time.Duration, bool) {
	if id := p.getAudioID(); id != (audio.SoundID{}) {
		if elapsed, available := p.mixer.PCMTime(id); available {
			return elapsed, true
		}
	}

	return time.Since(started), false
}

// waitForMediaTime polls at the established five-millisecond cadence. Once the
// audio clock is selected, failing to exceed the last observed media time for
// two seconds is treated as a stalled playback device.
func (p *embeddedPlayback) waitForMediaTime(
	ctx context.Context,
	started time.Time,
	target time.Duration,
) error {
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
