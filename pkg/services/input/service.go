package input

import (
	"context"
	"sync"

	"github.com/gravestench/servicemesh"
	lua "github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/common"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type Service struct {
	common.Service
	mesh              servicemesh.Mesh
	renderer          raylibRenderer.Dependency
	lua               luaManager.Dependency
	keyStates         map[int32]InputState
	keyModStates      map[int32]InputState
	mouseButtonStates map[int32]InputState
	cursor            struct {
		X, Y int
	}

	mux                 sync.RWMutex
	callbackMux         sync.Mutex
	keyPressedCallbacks map[int32][]*lua.LFunction
	stopFrames          func()
}

// Start initializes input state and subscribes polling to the renderer frame.
func (s *Service) Start(context.Context) error {
	s.keyStates = make(map[int32]InputState)
	s.keyModStates = make(map[int32]InputState)
	s.mouseButtonStates = make(map[int32]InputState)
	s.keyPressedCallbacks = make(map[int32][]*lua.LFunction)
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

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mesh = mesh
	_ = s.Start(context.Background())
}

func (s *Service) update() {
	s.updateKeyboardState()
	s.updateKeyboardModifierState()
	s.updateMouseCursorState()
	s.updateMouseButtonState()
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
