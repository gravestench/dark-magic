package input

import (
	"sync"
	"time"

	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
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

	mux sync.Mutex
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
