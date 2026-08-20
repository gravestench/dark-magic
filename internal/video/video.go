// Package video defines the backend-neutral contract for cinematic playback.
package video

import (
	"errors"
	"io/fs"
)

// ErrUnavailable reports an intentionally absent or unsupported playback
// backend; callers may apply authored skip/fallback policy rather than treating
// it as a corrupted cinematic.
var ErrUnavailable = errors.New("video playback is unavailable")

// State is the observable lifecycle of a playback session.
type State string

const (
	// Playing means decoding or presentation is still active.
	Playing State = "playing"
	// Complete means the media reached its natural end.
	Complete State = "complete"
	// Failed means playback stopped because of a retained error.
	Failed State = "failed"
	// Stopped means the owner explicitly cancelled playback.
	Stopped State = "stopped"
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

// Available is always false so capability checks select an authored fallback.
func (Unavailable) Available() bool { return false }

// Play returns the sentinel unavailable error without reading content or
// acquiring playback resources.
func (Unavailable) Play(fs.FS, string) (Playback, error) {
	return nil, ErrUnavailable
}
