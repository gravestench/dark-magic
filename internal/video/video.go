// Package video defines the backend-neutral contract for cinematic playback.
package video

import (
	"errors"
	"io/fs"
)

var ErrUnavailable = errors.New("video playback is unavailable")

// State is the observable lifecycle of a playback session.
type State string

const (
	Playing  State = "playing"
	Complete State = "complete"
	Failed   State = "failed"
	Stopped  State = "stopped"
)

// Snapshot is safe to poll from the engine update thread.
type Snapshot struct {
	State State
	Error string
}

// Playback is one independently stoppable cinematic.
type Playback interface {
	Snapshot() Snapshot
	Stop() error
}

// Backend owns decoding, presentation, and synchronized cinematic audio.
// Content is passed as an fs.FS so implementations can choose streaming or
// buffered access without forcing large BIK payloads through Lua.
type Backend interface {
	Available() bool
	Play(source fs.FS, path string) (Playback, error)
}

// Unavailable is the portable fallback used when no native video backend was
// compiled or configured. Lua can detect it and apply explicit failure policy.
type Unavailable struct{}

func (Unavailable) Available() bool                      { return false }
func (Unavailable) Play(fs.FS, string) (Playback, error) { return nil, ErrUnavailable }
