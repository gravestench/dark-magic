package input

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type InputState int

const (
	StateUp InputState = iota
	StateDown
	StateReleased
	StatePressed
)

type Service struct {
	logger            *zerolog.Logger
	renderer          raylibRenderer.Dependency
	keyStates         map[int32]InputState
	keyModStates      map[int32]InputState
	mouseButtonStates map[int32]InputState
	cursor            struct {
		X, Y int
	}
}

func (s *Service) KeyboardState() {
	//TODO implement me
	panic("implement me")
}

func (s *Service) KeyboardModifierState() {
	//TODO implement me
	panic("implement me")
}

func (s *Service) MouseCursorState() (x, y int) {
	return s.cursor.X, s.cursor.Y
}

func (s *Service) MouseButtonState() {
	//TODO implement me
	panic("implement me")
}

func (s *Service) updateKeyboardState() {
	for _, key := range []int32{
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
	} {
		s.keyStates[key] = StateUp

		if rl.IsKeyPressed(key) {
			s.keyStates[key] = StatePressed
		} else if rl.IsKeyReleased(key) {
			s.keyStates[key] = StateReleased
		} else if rl.IsKeyDown(key) {
			s.keyStates[key] = StateDown
		}
	}
}

func (s *Service) updateKeyboardModifierState() {
	for _, key := range []int32{
		rl.KeyLeftShift,
		rl.KeyLeftControl,
		rl.KeyLeftAlt,
		rl.KeyLeftSuper,
		rl.KeyRightShift,
		rl.KeyRightControl,
		rl.KeyRightAlt,
		rl.KeyRightSuper,
	} {
		s.keyModStates[key] = StateUp

		if rl.IsKeyPressed(key) {
			s.keyModStates[key] = StatePressed
		} else if rl.IsKeyReleased(key) {
			s.keyModStates[key] = StateReleased
		} else if rl.IsKeyDown(key) {
			s.keyModStates[key] = StateDown
		}
	}
}

func (s *Service) updateMouseCursorState() {
	p := rl.GetMousePosition()
	s.cursor.X, s.cursor.Y = int(p.X), int(p.Y)
}

func (s *Service) updateMouseButtonState() {
	for _, key := range []int32{
		rl.MouseLeftButton,
		rl.MouseRightButton,
		rl.MouseMiddleButton,
		rl.MouseSideButton,
		rl.MouseExtraButton,
		rl.MouseForwardButton,
		rl.MouseBackButton,
	} {
		s.mouseButtonStates[key] = StateUp

		if rl.IsMouseButtonPressed(key) {
			s.mouseButtonStates[key] = StatePressed
		} else if rl.IsMouseButtonReleased(key) {
			s.mouseButtonStates[key] = StateReleased
		} else if rl.IsMouseButtonDown(key) {
			s.mouseButtonStates[key] = StateDown
		}
	}
}

func (s *Service) Init(rt runtime.Runtime) {
	s.keyStates = make(map[int32]InputState)
	s.keyModStates = make(map[int32]InputState)
	s.mouseButtonStates = make(map[int32]InputState)

	ticker := time.NewTicker(time.Second / 24)

	// Create a Goroutine to handle the ticker
	go func() {
		for {
			select {
			case <-ticker.C:
				// Call your function here
				s.updateKeyboardState()
				s.updateKeyboardModifierState()
				s.updateMouseCursorState()
				s.updateMouseButtonState()
			}
		}
	}()
}

func (s *Service) Name() string {
	return "Template"
}

// the following methods are boilerplate, but they are used
// by the runtime to enforce a standard logging format.

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) Foo() {
	// do stuff here
}
