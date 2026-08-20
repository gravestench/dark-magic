package input

import (
	"context"
	"sync"

	"github.com/gravestench/dark-magic/internal/platform/raylib/common"
	raylibRenderer "github.com/gravestench/dark-magic/internal/platform/raylib/renderer"
)

// Service samples Raylib on the renderer owner thread and publishes copied
// logical input snapshots. It does not decide scene focus or gameplay authority.
type Service struct {
	common.Service
	renderer          raylibRenderer.Dependency
	keyStates         map[int32]InputState
	keyModStates      map[int32]InputState
	mouseButtonStates map[int32]InputState
	cursor            struct {
		X, Y int
	}
	scroll struct {
		X, Y float32
	}

	mux        sync.RWMutex
	stopFrames func()
}

// Start initializes published state before subscribing update to the renderer owner thread. The initial cursor is the
// logical center so software-pointer consumers have a meaningful coordinate before the first native sample.
func (s *Service) Start(context.Context) error {
	s.keyStates = make(map[int32]InputState)
	s.keyModStates = make(map[int32]InputState)
	s.mouseButtonStates = make(map[int32]InputState)
	width, height := s.renderer.Resolution()
	s.cursor.X, s.cursor.Y = width/2, height/2
	s.stopFrames = s.renderer.SubscribeFrame(s.update)

	return nil
}

// Stop detaches frame polling exactly once so subsequent renderer frames cannot mutate a stopped service.
func (s *Service) Stop(context.Context) error {
	if s.stopFrames != nil {
		s.stopFrames()
		s.stopFrames = nil
	}

	return nil
}

// New records the renderer dependency without starting native sampling; lifecycle ownership remains with the caller.
func New(renderer raylibRenderer.Dependency) *Service {
	return &Service{renderer: renderer}
}

// update samples keyboard, modifiers, cursor, wheel, and buttons in stable order on every renderer frame. Each domain
// publishes under its own lock interval, matching the established reader-visible timing.
func (s *Service) update() {
	s.updateKeyboardState()
	s.updateKeyboardModifierState()
	s.updateMouseCursorState()
	s.updateMouseWheelState()
	s.updateMouseButtonState()
}
