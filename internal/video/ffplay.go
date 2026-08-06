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

func (b FFplay) executable() (string, error) {
	name := b.Executable
	if name == "" {
		name = "ffplay"
	}
	return exec.LookPath(name)
}

func (b FFplay) Available() bool {
	_, err := b.executable()
	return err == nil
}

func (b FFplay) Play(source fs.FS, path string) (Playback, error) {
	executable, err := b.executable()
	if err != nil {
		return nil, fmt.Errorf("videocore: locate ffplay: %w", err)
	}
	input, err := source.Open(path)
	if err != nil {
		return nil, fmt.Errorf("videocore: open %q: %w", path, err)
	}
	defer input.Close()
	temporary, err := os.CreateTemp("", "dark-magic-*.bik")
	if err != nil {
		return nil, fmt.Errorf("videocore: create temporary BIK: %w", err)
	}
	temporaryName := temporary.Name()
	cleanup := func() { _ = temporary.Close(); _ = os.Remove(temporaryName) }
	if _, err := io.Copy(temporary, input); err != nil {
		cleanup()
		return nil, fmt.Errorf("videocore: extract %q: %w", path, err)
	}
	if err := temporary.Close(); err != nil {
		cleanup()
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
		cleanup()
		return nil, fmt.Errorf("videocore: start ffplay: %w", err)
	}
	go playback.wait()
	return playback, nil
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

func (p *ffplayPlayback) Snapshot() Snapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.snapshot
}

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

func (p *ffplayPlayback) cleanup() {
	p.cleanOnce.Do(func() { _ = os.Remove(p.temporary) })
}
