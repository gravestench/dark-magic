package input

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// InputState describes a key or button during one sampled renderer frame.
type InputState int

const (
	// StateUp means the control is inactive and had no transition this frame.
	StateUp InputState = iota
	// StateDown means the control remains held after an earlier press.
	StateDown
	// StateReleased means the control transitioned to inactive this frame.
	StateReleased
	// StatePressed means the control transitioned to active or emitted an allowed held-key repeat this frame.
	StatePressed
)

// KeyboardState returns a defensive copy of the most recently sampled ordinary keys. Callers may retain or mutate the
// map without racing the next renderer-thread update.
func (s *Service) KeyboardState() map[int32]InputState {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return cloneStates(s.keyStates)
}

// KeyState returns one ordinary key from a consistent sampled frame; unknown keys intentionally read as StateUp.
func (s *Service) KeyState(key int32) InputState {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.keyStates[key]
}

// KeyboardModifierState returns a defensive copy so consumer-side chord detection cannot mutate service ownership.
func (s *Service) KeyboardModifierState() map[int32]InputState {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return cloneStates(s.keyModStates)
}

// ModifierKeyState returns one modifier from the latest sample while sharing the same lock as bulk snapshots.
func (s *Service) ModifierKeyState(key int32) InputState {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.keyModStates[key]
}

// MouseCursorState returns the last valid logical coordinate, which remains stable while the native pointer is outside
// the game viewport.
func (s *Service) MouseCursorState() (x, y int) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.cursor.X, s.cursor.Y
}

// MouseButtonState returns a defensive copy of every tracked button so callers cannot modify the next frame's state.
func (s *Service) MouseButtonState() map[int32]InputState {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return cloneStates(s.mouseButtonStates)
}

// MouseWheelState returns the two-axis delta sampled for the current frame rather than accumulating prior motion.
func (s *Service) MouseWheelState() (x, y float32) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.scroll.X, s.scroll.Y
}

// cloneStates preserves Service ownership of mutable maps across frame publication.
func cloneStates(states map[int32]InputState) map[int32]InputState {
	result := make(map[int32]InputState, len(states))
	for key, state := range states {
		result[key] = state
	}

	return result
}

// updateKeyboardState samples the complete ordinary-key set while holding the publication lock, preventing readers
// from combining key states from two native frames.
func (s *Service) updateKeyboardState() {
	s.mux.Lock()
	for _, key := range s.normalKeyCodes() {
		s.keyStates[key] = currentKeyState(key, repeatsWhenHeld(key))
	}
	s.mux.Unlock()
}

// updateKeyboardModifierState samples modifiers separately because they never use held-key repeat semantics.
func (s *Service) updateKeyboardModifierState() {
	s.mux.Lock()
	for _, key := range s.modifierKeyCodes() {
		s.keyModStates[key] = currentKeyState(key, false)
	}
	s.mux.Unlock()
}

// updateMouseCursorState converts the native position through the renderer viewport and retains the previous logical
// point when conversion reports letterboxing or an out-of-window coordinate.
func (s *Service) updateMouseCursorState() {
	s.mux.Lock()
	p := rl.GetMousePosition()
	x, y, inside := s.renderer.ScreenToGame(int(p.X), int(p.Y))
	s.cursor.X, s.cursor.Y = retainLogicalCursor(s.cursor.X, s.cursor.Y, x, y, inside)
	s.mux.Unlock()
}

// updateMouseWheelState replaces both axes with Raylib's per-frame delta so scroll input is consumed exactly once.
func (s *Service) updateMouseWheelState() {
	s.mux.Lock()
	delta := rl.GetMouseWheelMoveV()
	s.scroll.X, s.scroll.Y = delta.X, delta.Y
	s.mux.Unlock()
}

// retainLogicalCursor prevents a software pointer from flashing at the logical
// origin when the native pointer enters letterboxing or leaves the window.
// Pointer actions are still rejected outside the viewport by ScreenToGame;
// presentation simply remains at its last meaningful coordinate.
func retainLogicalCursor(previousX, previousY, x, y int, inside bool) (int, int) {
	if !inside {
		return previousX, previousY
	}

	return x, y
}

// updateMouseButtonState samples every supported Raylib button under one lock and gives edge transitions precedence
// over the held state for the same frame.
func (s *Service) updateMouseButtonState() {
	s.mux.Lock()
	for _, key := range s.mouseButtonCodes() {
		code := int32(key)
		s.mouseButtonStates[code] = StateUp

		if rl.IsMouseButtonPressed(key) {
			s.mouseButtonStates[code] = StatePressed
		} else if rl.IsMouseButtonReleased(key) {
			s.mouseButtonStates[code] = StateReleased
		} else if rl.IsMouseButtonDown(key) {
			s.mouseButtonStates[code] = StateDown
		}
	}
	s.mux.Unlock()
}

// currentKeyState preserves transition precedence: press (including allowed repeat), release, held, then up. This
// ordering ensures a transition cannot be hidden by Raylib also reporting the key as down.
func currentKeyState(key int32, repeat bool) InputState {
	if rl.IsKeyPressed(key) || repeat && rl.IsKeyPressedRepeat(key) {
		return StatePressed
	}

	if rl.IsKeyReleased(key) {
		return StateReleased
	}

	if rl.IsKeyDown(key) {
		return StateDown
	}

	return StateUp
}

// repeatsWhenHeld follows ordinary desktop behavior: navigation and editing
// keys repeat after Raylib's platform-style initial delay, while activation,
// cancellation, toggles, and function keys remain one-shot.
func repeatsWhenHeld(key int32) bool {
	switch key {
	case rl.KeyBackspace, rl.KeyDelete, rl.KeyInsert,
		rl.KeyRight, rl.KeyLeft, rl.KeyDown, rl.KeyUp,
		rl.KeyPageUp, rl.KeyPageDown, rl.KeyHome, rl.KeyEnd:
		return true
	default:
		return false
	}
}

// normalKeyCodes defines the stable keyboard surface published by Service; keys outside this list read as StateUp.
func (*Service) normalKeyCodes() []int32 {
	return []int32{
		rl.KeySpace,
		rl.KeyEscape,
		rl.KeyEnter,
		rl.KeyTab,
		rl.KeyBackspace,
		rl.KeyInsert,
		rl.KeyDelete,
		rl.KeyRight,
		rl.KeyLeft,
		rl.KeyDown,
		rl.KeyUp,
		rl.KeyPageUp,
		rl.KeyPageDown,
		rl.KeyHome,
		rl.KeyEnd,
		rl.KeyCapsLock,
		rl.KeyScrollLock,
		rl.KeyNumLock,
		rl.KeyPrintScreen,
		rl.KeyPause,
		rl.KeyF1,
		rl.KeyF2,
		rl.KeyF3,
		rl.KeyF4,
		rl.KeyF5,
		rl.KeyF6,
		rl.KeyF7,
		rl.KeyF8,
		rl.KeyF9,
		rl.KeyF10,
		rl.KeyF11,
		rl.KeyF12,
		rl.KeyKbMenu,
		rl.KeyLeftBracket,
		rl.KeyBackSlash,
		rl.KeyRightBracket,
		rl.KeyGrave,
		rl.KeyKp0,
		rl.KeyKp1,
		rl.KeyKp2,
		rl.KeyKp3,
		rl.KeyKp4,
		rl.KeyKp5,
		rl.KeyKp6,
		rl.KeyKp7,
		rl.KeyKp8,
		rl.KeyKp9,
		rl.KeyKpDecimal,
		rl.KeyKpDivide,
		rl.KeyKpMultiply,
		rl.KeyKpSubtract,
		rl.KeyKpAdd,
		rl.KeyKpEnter,
		rl.KeyKpEqual,
		rl.KeyApostrophe,
		rl.KeyComma,
		rl.KeyMinus,
		rl.KeyPeriod,
		rl.KeySlash,
		rl.KeyZero,
		rl.KeyOne,
		rl.KeyTwo,
		rl.KeyThree,
		rl.KeyFour,
		rl.KeyFive,
		rl.KeySix,
		rl.KeySeven,
		rl.KeyEight,
		rl.KeyNine,
		rl.KeySemicolon,
		rl.KeyEqual,
		rl.KeyA,
		rl.KeyB,
		rl.KeyC,
		rl.KeyD,
		rl.KeyE,
		rl.KeyF,
		rl.KeyG,
		rl.KeyH,
		rl.KeyI,
		rl.KeyJ,
		rl.KeyK,
		rl.KeyL,
		rl.KeyM,
		rl.KeyN,
		rl.KeyO,
		rl.KeyP,
		rl.KeyQ,
		rl.KeyR,
		rl.KeyS,
		rl.KeyT,
		rl.KeyU,
		rl.KeyV,
		rl.KeyW,
		rl.KeyX,
		rl.KeyY,
		rl.KeyZ,
		rl.KeyBack,
		rl.KeyMenu,
		rl.KeyVolumeUp,
		rl.KeyVolumeDown,
	}
}

// modifierKeyCodes keeps left and right modifiers distinct so callers can implement exact chord policy.
func (*Service) modifierKeyCodes() []int32 {
	return []int32{
		rl.KeyLeftShift,
		rl.KeyLeftControl,
		rl.KeyLeftAlt,
		rl.KeyLeftSuper,
		rl.KeyRightShift,
		rl.KeyRightControl,
		rl.KeyRightAlt,
		rl.KeyRightSuper,
	}
}

// mouseButtonCodes lists every Raylib mouse button sampled into a published frame.
func (*Service) mouseButtonCodes() []rl.MouseButton {
	return []rl.MouseButton{
		rl.MouseButtonLeft,
		rl.MouseButtonRight,
		rl.MouseButtonMiddle,
		rl.MouseButtonSide,
		rl.MouseButtonExtra,
		rl.MouseButtonForward,
		rl.MouseButtonBack,
	}
}
