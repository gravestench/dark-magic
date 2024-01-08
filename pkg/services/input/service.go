package input

import (
	"sync"
	"time"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/common"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type Service struct {
	common.Service
	mesh              servicemesh.Mesh
	renderer          raylibRenderer.Dependency
	keyStates         map[int32]InputState
	keyModStates      map[int32]InputState
	mouseButtonStates map[int32]InputState
	cursor            struct {
		X, Y int
	}

	mux sync.Mutex

	debug bool
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.debug = true

	s.mesh = mesh
	s.keyStates = make(map[int32]InputState)
	s.keyModStates = make(map[int32]InputState)
	s.mouseButtonStates = make(map[int32]InputState)

	go s.updateNonBlocking()
}

func (s *Service) updateNonBlocking() {
	ticker := time.NewTicker(time.Second / 24)

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
	return "Input"
}

func (s *Service) Ready() bool {
	if s.renderer == nil {
		return false
	}

	return true
}

func (s *Service) Foo() {
	// do stuff here
}
