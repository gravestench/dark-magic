// Package inputstate defines backend-neutral, frame-stable input snapshots.
package inputstate

import "sync/atomic"

// ActionState is the state of one logical game action for a frame.
type ActionState struct {
	Down     bool
	Pressed  bool
	Released bool
}

// Frame is an immutable input snapshot published by a native backend.
type Frame struct {
	Actions map[string]ActionState
	Text    string
	CursorX float64
	CursorY float64
}

// Text returns the UTF-8 text entered during the current frame.
func (s *Store) Text() string {
	value := s.current.Load()
	if value == nil {
		return ""
	}
	return value.(Frame).Text
}

// Store publishes and reads cloned immutable frame snapshots.
type Store struct{ current atomic.Value }

// Publish replaces the current frame.
func (s *Store) Publish(frame Frame) {
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
	frame.Actions = cloneActions(frame.Actions)
	return frame
}

// Action reads one immutable logical action without cloning the frame map.
func (s *Store) Action(name string) ActionState {
	value := s.current.Load()
	if value == nil {
		return ActionState{}
	}
	return value.(Frame).Actions[name]
}

// Cursor reads the immutable current cursor coordinates without allocation.
func (s *Store) Cursor() (float64, float64) {
	value := s.current.Load()
	if value == nil {
		return 0, 0
	}
	frame := value.(Frame)
	return frame.CursorX, frame.CursorY
}

func cloneActions(actions map[string]ActionState) map[string]ActionState {
	result := make(map[string]ActionState, len(actions))
	for name, state := range actions {
		result[name] = state
	}
	return result
}
