package luaManager

import (
	"fmt"
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

type Service struct {
	common.Service
	mesh          servicemesh.Mesh
	state         *lua.LState
	events        *ee.EventEmitter
	mux           sync.Mutex
	boundServices map[string]any
	wg            sync.WaitGroup
	tasks         chan task
}

func (s *Service) Name() string {
	return "Lua Environment"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) Init(fromMesh servicemesh.Mesh) {
	s.mesh = fromMesh
	s.tasks = make(chan task)
	s.wg.Add(1)
	go s.luaStateAccessSingletonWorker()
	s.RebuildState()
}

func (s *Service) RebuildState() {
	s.state = s.newState()
}

func (s *Service) newState() *lua.LState {
	s.mux.Lock()
	defer s.mux.Lock()

	s.boundServices = make(map[string]any)

	s.state = lua.NewState()
	s.state.OpenLibs()

	bindLoggerToLuaEnvironment(s.Logger(), s.state)

	apiTable := s.state.NewTable()
	s.state.SetGlobal(globalApiTableKey, apiTable)

	for _, service := range s.mesh.Services() {
		s.exportToLuaEnvironment(service)
	}

	return s.state
}

func bindLoggerToLuaEnvironment(logger *slog.Logger, state *lua.LState) {
	fnPrint := state.NewFunction(func(L *lua.LState) int {
		numArgs := L.GetTop()

		if numArgs < 1 {
			return 0
		}

		arg0 := L.CheckString(1)
		args := make([]any, numArgs)
		for idx := 1; idx < numArgs; idx++ {
			args = append(args, L.CheckAny(idx))
		}

		logger.Info(arg0, args[1:]...)

		return 0
	})

	state.SetGlobal("print", fnPrint)
}

// WithState execs the given function with the singleton lua state
func (s *Service) WithState(fn func(state *lua.LState) error) (err error) {
	done := make(chan error)
	s.tasks <- task{fn: fn, done: done}
	return <-done
}

func (s *Service) exportToLuaEnvironment(service servicemesh.Service) {
	plugin, ok := service.(LuaPlugin)
	if !ok {
		return
	}

	if candidate, ok := service.(servicemesh.HasDependencies); ok {
		for !candidate.DependenciesResolved() {
			s.Logger().Warn("waiting for service to resolve dependencies", "target", service.Name())
			time.Sleep(time.Second)
		}
	}

	if err := s.WithState(func(L *lua.LState) error {
		if _, exists := s.boundServices[service.Name()]; exists {
			return nil
		}

		apiTable := L.GetGlobal(globalApiTableKey)

		switch apiTable.(type) {
		case *lua.LNilType:
			apiTable = L.NewTable()
		}

		L.SetGlobal(globalApiTableKey, apiTable)

		if s.boundServices == nil {
			s.boundServices = make(map[string]any)
		}

		s.boundServices[service.Name()] = service

		// this needs to be done in a goroutine!
		go plugin.ExportToLua(L, apiTable.(*lua.LTable))

		return nil
	}); err != nil {
		s.Logger().Error("exporting to lua", "target service", service.Name(), "error", err)
	}
}

// WaitForGlobals blocks until all supplied globals exist
func (s *Service) WaitForGlobals(globals ...string) {
	eb := backoff.NewExponentialBackOff()
	for {
		err := s.WithState(func(L *lua.LState) error {
			if !s.globalsExist(L, globals...) {
				return fmt.Errorf("not found: %v", globals)
			}
			return nil
		})

		if err != nil {
			s.Logger().Warn("backoff", "errMsg", err)
			time.Sleep(eb.NextBackOff())
			continue
		}

		break
	}
}

func (s *Service) GlobalsExist(globals ...string) bool {
	for _, globalKey := range globals {
		if !luaExists(s.state, globalKey) {
			return false
		}
	}
	return true
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
