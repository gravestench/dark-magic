package video

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"sync"
)

// FFplay is a runtime Bink backend using FFmpeg's maintained Bink 1 decoders.
// It is intentionally optional: Available reports false when the executable is
// absent, allowing the Lua startup policy to skip cinematics cleanly.
type FFplay struct {
	Executable string
}

// executable resolves the configured player or the conventional ffplay name;
// lookup failure is retained so Play can wrap it with backend context.
func (b FFplay) executable() (string, error) {
	name := b.Executable
	if name == "" {
		name = "ffplay"
	}

	return exec.LookPath(name)
}

// Available performs the same executable lookup as Play without starting a
// process, allowing callers to choose a fallback before loading cinematic data.
func (b FFplay) Available() bool {
	_, err := b.executable()
	return err == nil
}

// Play copies the fs-backed asset to a host file required by ffplay, starts the
// child with inherited standard streams, and transfers file cleanup to Playback.
func (b FFplay) Play(source fs.FS, path string) (Playback, error) {
	executable, err := b.executable()
	if err != nil {
		return nil, fmt.Errorf("videocore: locate ffplay: %w", err)
	}

	input, err := source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("videocore: open %q: %w", path, err)
	}

	// The source is read-only and its close failure cannot invalidate a copied
	// temporary file or an already-started player.
	defer func() { _ = input.Close() }()

	temporary, err := os.CreateTemp("", "dark-magic-*.bik")
	if err != nil {
		return nil, fmt.Errorf("videocore: create temporary BIK: %w", err)
	}

	temporaryName := temporary.Name()

	if _, err := io.Copy(temporary, input); err != nil {
		discardTemporaryVideo(temporary)
		return nil, fmt.Errorf("videocore: extract %q: %w", path, err)
	}

	if err := temporary.Close(); err != nil {
		discardTemporaryVideo(temporary)
		return nil, fmt.Errorf("videocore: close temporary BIK: %w", err)
	}

	command := exec.Command(executable,
		"-autoexit", "-noborder", "-window_title", "Dark Magic Cinematic",
		"-loglevel", "error", temporaryName)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr

	playback := &ffplayPlayback{
		command: command, temporary: temporaryName,
		snapshot: Snapshot{State: Playing}, done: make(chan struct{}),
	}
	if err := command.Start(); err != nil {
		discardTemporaryVideo(temporary)
		return nil, fmt.Errorf("videocore: start ffplay: %w", err)
	}

	go playback.wait()

	return playback, nil
}

// discardTemporaryVideo closes before removing a failed extraction; both
// cleanup errors remain best-effort so the primary playback error is preserved.
func discardTemporaryVideo(temporary *os.File) {
	_ = temporary.Close()
	_ = os.Remove(temporary.Name())
}

type ffplayPlayback struct {
	mu        sync.RWMutex
	command   *exec.Cmd
	temporary string
	snapshot  Snapshot
	done      chan struct{}
	stopOnce  sync.Once
	cleanOnce sync.Once
}

// Snapshot copies child-process lifecycle state for lock-safe engine polling.
func (p *ffplayPlayback) Snapshot() Snapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.snapshot
}

// wait converts the child exit into Complete or Failed unless Stop already won
// the state transition, then removes the temporary file before signaling done.
func (p *ffplayPlayback) wait() {
	err := p.command.Wait()
	p.mu.Lock()
	if p.snapshot.State == Playing {
		if err == nil {
			p.snapshot = Snapshot{State: Complete}
		} else {
			p.snapshot = Snapshot{State: Failed, Error: err.Error()}
		}
	}
	p.mu.Unlock()
	p.cleanup()
	close(p.done)
}

// Stop records explicit cancellation once, requests child termination while
// holding the state lock, and waits for wait to reap the process and remove its input file.
func (p *ffplayPlayback) Stop() error {
	var stopErr error

	p.stopOnce.Do(func() {
		p.mu.Lock()
		if p.snapshot.State == Playing {
			p.snapshot = Snapshot{State: Stopped}
			if p.command.Process != nil {
				stopErr = p.command.Process.Kill()
			}
		}
		p.mu.Unlock()
	})
	<-p.done

	return stopErr
}

// cleanup removes the extracted BIK at most once so process exit and future
// cleanup callers cannot race the same filesystem operation.
func (p *ffplayPlayback) cleanup() {
	p.cleanOnce.Do(func() { _ = os.Remove(p.temporary) })
}
