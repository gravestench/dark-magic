package luaManager

import (
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	ee "github.com/gravestench/eventemitter"
	"github.com/gravestench/servicemesh"
	"github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/common"
)

const (
	globalApiTableKey = "api"
)

var ErrStateClosed = errors.New("Lua state is closed")

type Service struct {
	common.Service
	mesh          servicemesh.Mesh
	state         *lua.LState
	closed        bool
	events        *ee.EventEmitter
	mux           sync.Mutex
	stateMux      sync.Mutex
	boundServices map[string]any
	wg            sync.WaitGroup
}

// WithState serializes access to the shared Lua state. Callers must not retain
// the state after fn returns.
func (s *Service) WithState(fn func(state *lua.LState) error) error {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()

	if s.closed {
		return ErrStateClosed
	}
	if s.state == nil {
		s.state = s.newState()
	}

	return fn(s.state)
}

func (s *Service) Name() string {
	return "Lua Environment"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) OnShutdown() {
	s.stateMux.Lock()
	defer s.stateMux.Unlock()
	if s.state != nil {
		s.state.Close()
		s.state = nil
	}
	s.closed = true
}

func (s *Service) Init(fromMesh servicemesh.Mesh) {
	s.mesh = fromMesh
	s.RebuildState()
}

func (s *Service) RebuildState() {
	s.stateMux.Lock()
	if s.state != nil {
		s.state.Close()
	}
	s.state = s.newState()
	s.closed = false
	s.stateMux.Unlock()

	s.mux.Lock()
	s.boundServices = make(map[string]any)
	s.mux.Unlock()

	if s.mesh != nil {
		for _, service := range s.mesh.Services() {
			s.exportToLuaEnvironment(service)
		}
	}
}

func (s *Service) newState() *lua.LState {
	state := lua.NewState()
	state.OpenLibs()

	bindLoggerToLuaEnvironment(s.Logger(), state)

	apiTable := state.NewTable()
	state.SetGlobal(globalApiTableKey, apiTable)

	return state
}

func bindLoggerToLuaEnvironment(logger *slog.Logger, state *lua.LState) {
	fnPrint := state.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			return 0
		}

		arg0 := L.CheckString(1)
		args := make([]any, 0, numArgs-1)
		for idx := 1; idx < numArgs; idx++ {
			args = append(args, L.CheckAny(idx+1))
		}

		logger.Info(arg0, args...)

		return 0
	})

	state.SetGlobal("print", fnPrint)
}

func (s *Service) exportToLuaEnvironment(service servicemesh.Service) {
	plugin, ok := service.(LuaPlugin)
	if !ok {
		return
	}

	s.mux.Lock()
	if _, exists := s.boundServices[service.Name()]; exists {
		s.mux.Unlock()
		return
	}
	s.boundServices[service.Name()] = service
	s.mux.Unlock()

	if err := s.WithState(func(state *lua.LState) error {
		apiTable := state.GetGlobal(globalApiTableKey)
		if apiTable == lua.LNil {
			apiTable = state.NewTable()
		}
		state.SetGlobal(globalApiTableKey, apiTable)
		plugin.ExportToLua(state, apiTable.(*lua.LTable))
		return nil
	}); err != nil {
		s.Logger().Error("exporting service to Lua", "service", service.Name(), "error", err)
	}
}

// WaitForGlobals blocks until all supplied globals exist
func (s *Service) WaitForGlobals(globals ...string) {
	eb := backoff.NewExponentialBackOff()
	for {
		if !s.GlobalsExist(globals...) {
			s.Logger().Warn("waiting for Lua globals", "globals", globals)
			time.Sleep(eb.NextBackOff())
			continue
		}

		break
	}
}

func (s *Service) GlobalsExist(globals ...string) bool {
	exists := false
	_ = s.WithState(func(state *lua.LState) error {
		exists = s.globalsExist(state, globals...)
		return nil
	})
	return exists
}

func (s *Service) globalsExist(L *lua.LState, globals ...string) bool {
	for _, globalKey := range globals {
		if !luaExists(L, globalKey) {
			return false
		}
	}

	return true
}

func luaExists(L *lua.LState, path string) bool {
	keys := strings.Split(path, ".")
	value := L.GetGlobal(keys[0])
	if value == lua.LNil {
		return false
	}

	for _, key := range keys[1:] {
		if tbl, ok := value.(*lua.LTable); ok {
			if L.GetField(tbl, key) == lua.LNil {
				return false
			}
		} else {
			return false
		}
	}

	return true
}
