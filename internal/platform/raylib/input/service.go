package input

import (
	"context"
	"sync"

	"github.com/gravestench/dark-magic/internal/platform/raylib/common"
	raylibRenderer "github.com/gravestench/dark-magic/internal/platform/raylib/renderer"
)

type Service struct {
	common.Service
	renderer          raylibRenderer.Dependency
	keyStates         map[int32]InputState
	keyModStates      map[int32]InputState
	mouseButtonStates map[int32]InputState
	cursor            struct {
		X, Y int
	}

	mux        sync.RWMutex
	stopFrames func()
}

// Start initializes input state and subscribes polling to the renderer frame.
func (s *Service) Start(context.Context) error {
	s.keyStates = make(map[int32]InputState)
	s.keyModStates = make(map[int32]InputState)
	s.mouseButtonStates = make(map[int32]InputState)
	width, height := s.renderer.Resolution()
	s.cursor.X, s.cursor.Y = width/2, height/2
	s.stopFrames = s.renderer.SubscribeFrame(s.update)
	return nil
}

// Stop detaches frame polling.
func (s *Service) Stop(context.Context) error {
	if s.stopFrames != nil {
		s.stopFrames()
		s.stopFrames = nil
	}
	return nil
}

// New constructs input with its renderer dependency explicitly.
func New(renderer raylibRenderer.Dependency) *Service {
	return &Service{renderer: renderer}
}

func (s *Service) update() {
	s.updateKeyboardState()
	s.updateKeyboardModifierState()
	s.updateMouseCursorState()
	s.updateMouseButtonState()
}
