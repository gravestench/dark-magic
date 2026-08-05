package input

import (
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

	mux                 sync.Mutex
	callbackMux         sync.Mutex
	keyPressedCallbacks map[int32][]*lua.LFunction
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mesh = mesh
	s.keyStates = make(map[int32]InputState)
	s.keyModStates = make(map[int32]InputState)
	s.mouseButtonStates = make(map[int32]InputState)
	s.keyPressedCallbacks = make(map[int32][]*lua.LFunction)

	s.renderer.OnFrame(s.update)
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
	if s.renderer == nil || s.lua == nil {
		return false
	}

	return true
}
