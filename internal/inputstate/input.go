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
	Actions    map[string]ActionState
	Text       string
	CursorX    float64
	CursorY    float64
	ScrollX    float64
	ScrollY    float64
	Owner      FocusOwner
	Gameplay   bool
	WorldView  string
	WorldSplit float64
}

// Route assigns one focus owner. A capturing debug surface receives the raw
// frame before Route; downstream scene consumers retain pointer position for
// presentation but cannot observe actions or entered text.
func Route(frame Frame, owner FocusOwner, captured, gameplay bool, worldView string) Frame {
	frame.Owner = owner
	frame.Gameplay = gameplay && !captured
	frame.WorldView = worldView
	if captured {
		frame.Actions = make(map[string]ActionState)
		frame.Text = ""
		frame.ScrollX, frame.ScrollY = 0, 0
	}
	return frame
}

// Store publishes and reads cloned immutable frame snapshots.
type Store struct {
	current      atomic.Value
	suppressed   atomic.Int32
	gameplayOnly atomic.Int32
	pointerOnly  atomic.Int32
	allPointer   atomic.Int32
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
		frame.ScrollX, frame.ScrollY = 0, 0
		return frame
	}
	if s.gameplayOnly.Load() > 0 {
		frame.Actions = gameplayActions(frame.Actions, frame.CursorX, frame.WorldView, frame.WorldSplit, s.allPointer.Load() > 0)
		frame.Text = ""
		return frame
	}
	if s.pointerOnly.Load() > 0 {
		frame.Actions = pointerActions(frame.Actions)
		frame.Text = ""
		return frame
	}
	frame.Actions = cloneActions(frame.Actions)
	return frame
}

// Text returns the UTF-8 text entered during the current frame.
func (s *Store) Text() string {
	if s.suppressed.Load() > 0 || s.gameplayOnly.Load() > 0 || s.pointerOnly.Load() > 0 {
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
	if s.gameplayOnly.Load() > 0 {
		if !isGameplayAction(name) || name == "pointer_primary" && s.allPointer.Load() == 0 && !pointerInWorld(value.(Frame)) {
			return ActionState{}
		}
	} else if s.pointerOnly.Load() > 0 && name != "pointer_primary" {
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

// Scroll reads this frame's high-resolution pointer scroll delta. Values stay
// fractional so trackpads are not degraded into simulated wheel clicks.
func (s *Store) Scroll() (float64, float64) {
	if s.suppressed.Load() > 0 || s.gameplayOnly.Load() > 0 {
		return 0, 0
	}
	value := s.current.Load()
	if value == nil {
		return 0, 0
	}
	frame := value.(Frame)
	return frame.ScrollX, frame.ScrollY
}

// Owner returns the current frame's explicit focus owner.
func (s *Store) Owner() FocusOwner {
	value := s.current.Load()
	if value == nil {
		return FocusOwner{Domain: FocusNone}
	}
	return value.(Frame).Owner
}

// Gameplay reports whether the current UI owner permits authoritative gameplay
// input to pass through to the world.
func (s *Store) Gameplay() bool {
	value := s.current.Load()
	return value != nil && value.(Frame).Gameplay
}

// Suppress prevents a nonfocused scene callback from reading actions, text, or
// pointer coordinates while preserving the published frame for its owner.
func (s *Store) Suppress(callback func() error) error {
	s.suppressed.Add(1)
	defer s.suppressed.Add(-1)
	return callback()
}

// GameplayOnly lets an underlying world callback observe only gameplay actions
// and pointer coordinates. Overlay toggles, menu navigation, and text remain
// exclusive to the focused UI owner.
func (s *Store) GameplayOnly(callback func() error) error {
	s.gameplayOnly.Add(1)
	defer s.gameplayOnly.Add(-1)
	return callback()
}

// GameplayAndPointer exposes gameplay actions plus pointer input everywhere.
// The persistent game-world scene uses this to keep HUD controls interactive;
// authoritative world consumers still use Frame.WorldView to reject covered
// regions.
func (s *Store) GameplayAndPointer(callback func() error) error {
	s.gameplayOnly.Add(1)
	s.allPointer.Add(1)
	defer s.gameplayOnly.Add(-1)
	defer s.allPointer.Add(-1)
	return callback()
}

// PointerOnly exposes cursor state and pointer actions without keyboard,
// gamepad, text, or gameplay actions to a nonfocused visible UI surface.
func (s *Store) PointerOnly(callback func() error) error {
	s.pointerOnly.Add(1)
	defer s.pointerOnly.Add(-1)
	return callback()
}

func isGameplayAction(name string) bool {
	switch name {
	case "left", "right", "up", "down", "pointer_primary", "interact", "skill_primary", "skill_secondary", "toggle_run":
		return true
	default:
		return false
	}
}

func gameplayActions(actions map[string]ActionState, cursorX float64, worldView string, worldSplit float64, allPointer bool) map[string]ActionState {
	result := make(map[string]ActionState)
	for name, state := range actions {
		if isGameplayAction(name) && (name != "pointer_primary" || allPointer || pointerInView(cursorX, worldView, worldSplit)) {
			result[name] = state
		}
	}
	return result
}

func pointerActions(actions map[string]ActionState) map[string]ActionState {
	result := make(map[string]ActionState)
	if state, ok := actions["pointer_primary"]; ok {
		result["pointer_primary"] = state
	}
	return result
}

func pointerInWorld(frame Frame) bool {
	return pointerInView(frame.CursorX, frame.WorldView, frame.WorldSplit)
}

func pointerInView(x float64, view string, split float64) bool {
	if split <= 0 {
		split = 400
	}
	switch view {
	case "none":
		return false
	case "left":
		return x < split
	case "right":
		return x >= split
	default:
		return true
	}
}

func cloneActions(actions map[string]ActionState) map[string]ActionState {
	result := make(map[string]ActionState, len(actions))
	for name, state := range actions {
		result[name] = state
	}
	return result
}
