package input

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type InputState int

const (
	StateUp InputState = iota
	StateDown
	StateReleased
	StatePressed
)

func (s *Service) KeyboardState() map[int32]InputState {
	s.mux.Lock()
	defer s.mux.Unlock()

	return cloneStates(s.keyStates)
}

func (s *Service) KeyboardModifierState() map[int32]InputState {
	s.mux.Lock()
	defer s.mux.Unlock()

	return cloneStates(s.keyModStates)
}

func (s *Service) MouseCursorState() (x, y int) {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.cursor.X, s.cursor.Y
}

func (s *Service) MouseButtonState() map[int32]InputState {
	s.mux.Lock()
	defer s.mux.Unlock()

	return cloneStates(s.mouseButtonStates)
}

func cloneStates(states map[int32]InputState) map[int32]InputState {
	result := make(map[int32]InputState, len(states))
	for key, state := range states {
		result[key] = state
	}
	return result
}

func (s *Service) updateKeyboardState() {
	s.mux.Lock()
	var shouldEmitEvent bool
	var pressed []int32

	for _, key := range s.normalKeyCodes() {
		beforeChange := s.keyStates[key]
		s.keyStates[key] = currentKeyState(key)

		if beforeChange != s.keyStates[key] {
			shouldEmitEvent = true
		}
		if s.keyStates[key] == StatePressed {
			pressed = append(pressed, key)
		}
	}
	snapshot := cloneStates(s.keyStates)
	s.mux.Unlock()
	if shouldEmitEvent {
		s.mesh.Events().Emit("KeyboardKeyStateChange", snapshot)
	}
	for _, key := range pressed {
		s.dispatchKeyPressed(key)
	}
}

func (s *Service) updateKeyboardModifierState() {
	s.mux.Lock()
	var shouldEmitEvent bool

	for _, key := range s.modifierKeyCodes() {
		beforeChange := s.keyModStates[key]

		s.keyModStates[key] = currentKeyState(key)

		if beforeChange != s.keyModStates[key] {
			shouldEmitEvent = true
		}
	}
	snapshot := cloneStates(s.keyModStates)
	s.mux.Unlock()
	if shouldEmitEvent {
		s.mesh.Events().Emit("KeyboardModKeyStateChange", snapshot)
	}
}

func (s *Service) updateMouseCursorState() {
	s.mux.Lock()
	beforeX, beforeY := s.cursor.X, s.cursor.Y

	p := rl.GetMousePosition()
	s.cursor.X, s.cursor.Y = int(p.X), int(p.Y)
	cursor := s.cursor
	s.mux.Unlock()

	if beforeX != cursor.X || beforeY != cursor.Y {
		s.mesh.Events().Emit("MouseCursorStateChange", cursor)
	}
}

func (s *Service) updateMouseButtonState() {
	s.mux.Lock()
	var shouldEmitEvent bool

	for _, key := range s.mouseButtonKeyCodess() {
		beforeChange := s.mouseButtonStates[key]
		s.mouseButtonStates[key] = StateUp

		if rl.IsMouseButtonPressed(key) {
			s.mouseButtonStates[key] = StatePressed
		} else if rl.IsMouseButtonReleased(key) {
			s.mouseButtonStates[key] = StateReleased
		} else if rl.IsMouseButtonDown(key) {
			s.mouseButtonStates[key] = StateDown
		}

		if beforeChange != s.mouseButtonStates[key] {
			shouldEmitEvent = true
		}
	}
	snapshot := cloneStates(s.mouseButtonStates)
	s.mux.Unlock()
	if shouldEmitEvent {
		s.mesh.Events().Emit("MouseButtonStateChange", snapshot)
	}
}

func currentKeyState(key int32) InputState {
	if rl.IsKeyPressed(key) {
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

func (*Service) mouseButtonKeyCodess() []int32 {
	return []int32{
		rl.MouseLeftButton,
		rl.MouseRightButton,
		rl.MouseMiddleButton,
		rl.MouseSideButton,
		rl.MouseExtraButton,
		rl.MouseForwardButton,
		rl.MouseBackButton,
	}
}
