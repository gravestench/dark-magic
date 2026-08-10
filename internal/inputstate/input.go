// Package inputstate defines backend-neutral, frame-stable input snapshots.
package inputstate

import "sync/atomic"

// ActionState is the state of one logical game action for a frame.
type ActionState struct {
	Down     bool
	Pressed  bool
	Released bool
}

// FocusOwner identifies the one consumer entitled to act on a frame.
type FocusOwner struct {
	Domain string
	ID     string
}

const (
	FocusNone  = "none"
	FocusScene = "scene"
	FocusDebug = "debug"
)

// Frame is an immutable input snapshot published by a native backend.
type Frame struct {
	Actions map[string]ActionState
	Text    string
	CursorX float64
	CursorY float64
	Owner   FocusOwner
}

// Route assigns one focus owner. A capturing debug surface receives the raw
// frame before Route; downstream scene consumers retain pointer position for
// presentation but cannot observe actions or entered text.
func Route(frame Frame, owner FocusOwner, captured bool) Frame {
	frame.Owner = owner
	if captured {
		frame.Actions = make(map[string]ActionState)
		frame.Text = ""
	}
	return frame
}

// Store publishes and reads cloned immutable frame snapshots.
type Store struct {
	current    atomic.Value
	suppressed atomic.Int32
}

// Publish replaces the current frame.
func (s *Store) Publish(frame Frame) {
	if frame.Owner.Domain == "" {
		frame.Owner.Domain = FocusNone
	}
	frame.Actions = cloneActions(frame.Actions)
	s.current.Store(frame)
}

// Snapshot returns a defensive copy of the latest frame.
func (s *Store) Snapshot() Frame {
	value := s.current.Load()
	if value == nil {
		return Frame{Actions: make(map[string]ActionState)}
	}
	frame := value.(Frame)
	if s.suppressed.Load() > 0 {
		frame.Actions = make(map[string]ActionState)
		frame.Text = ""
		frame.CursorX, frame.CursorY = 0, 0
		return frame
	}
	frame.Actions = cloneActions(frame.Actions)
	return frame
}

// Text returns the UTF-8 text entered during the current frame.
func (s *Store) Text() string {
	if s.suppressed.Load() > 0 {
		return ""
	}
	value := s.current.Load()
	if value == nil {
		return ""
	}
	return value.(Frame).Text
}

// Action reads one immutable logical action without cloning the frame map.
func (s *Store) Action(name string) ActionState {
	if s.suppressed.Load() > 0 {
		return ActionState{}
	}
	value := s.current.Load()
	if value == nil {
		return ActionState{}
	}
	return value.(Frame).Actions[name]
}

// Cursor reads the immutable current cursor coordinates without allocation.
func (s *Store) Cursor() (float64, float64) {
	if s.suppressed.Load() > 0 {
		return 0, 0
	}
	value := s.current.Load()
	if value == nil {
		return 0, 0
	}
	frame := value.(Frame)
	return frame.CursorX, frame.CursorY
}

// Owner returns the current frame's explicit focus owner.
func (s *Store) Owner() FocusOwner {
	value := s.current.Load()
	if value == nil {
		return FocusOwner{Domain: FocusNone}
	}
	return value.(Frame).Owner
}

// Suppress prevents a nonfocused scene callback from reading actions, text, or
// pointer coordinates while preserving the published frame for its owner.
func (s *Store) Suppress(callback func() error) error {
	s.suppressed.Add(1)
	defer s.suppressed.Add(-1)
	return callback()
}

func cloneActions(actions map[string]ActionState) map[string]ActionState {
	result := make(map[string]ActionState, len(actions))
	for name, state := range actions {
		result[name] = state
	}
	return result
}
