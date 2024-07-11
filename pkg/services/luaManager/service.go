package luaManager

import (
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
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
	isInit        bool
}

func (s *Service) Name() string {
	return "Lua Environment"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.mesh = mesh
	s.tasks = make(chan task)

	s.wg.Add(1)
	s.RebuildState()

	s.isInit = true

	for _, service := range s.mesh.Services() {
		go s.loadPlugin(service)
	}

	s.Logger().Info("setting up worker thread")

	go s.luaStateAccessSingletonWorker()

	mesh.Events().On(servicemesh.EventServiceAdded, func(args ...any) {
		if len(args) < 1 {
			return
		}

		service, ok := args[0].(servicemesh.Service)
		if !ok {
			return
		}

		s.loadPlugin(service)
	})

	s.Logger().Info("initialized")
}

func (s *Service) RebuildState() {
	s.Logger().Info("rebuilding state")
	s.state = s.newState()
}

func (s *Service) newState() *lua.LState {
	s.mux.Lock()
	defer s.mux.Lock()

	s.boundServices = make(map[string]any)

	s.Logger().Info("setting new state")
	s.state = lua.NewState()

	s.Logger().Info("importing OpenLibs")
	s.state.OpenLibs()

	bindLoggerToLuaEnvironment(s.Logger(), s.state)

	apiTable := s.state.NewTable()

	s.state.SetField(apiTable, "lua", createLuaTableFromStruct(s.state, s))
	s.state.SetGlobal(globalApiTableKey, apiTable)
	s.state.SetGlobal("sleep", s.state.NewFunction(s.luaSleep))

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

	eb := backoff.NewExponentialBackOff()

	for !s.isInit {
		time.Sleep(eb.NextBackOff())
	}

	go func() {
		s.tasks <- task{fn: fn, done: done}
	}()

	return <-done
}

func (s *Service) loadPlugin(service servicemesh.Service) {
	eb := backoff.NewExponentialBackOff()

	for !s.isInit {
		time.Sleep(eb.NextBackOff())
	}

	if candidate, ok := service.(servicemesh.HasDependencies); ok {
		for !candidate.DependenciesResolved() {
			s.Logger().Debug("waiting for candidate service to resolve dependencies", "target service", service.Name())
			time.Sleep(time.Second)
		}

		s.Logger().Debug("candidate service dependencies resolved", "target service", service.Name())
	}

	s.Logger().Info("loading plugin", "name", service.Name())
	go func() {
		s.mux.Lock()
		if _, exists := s.boundServices[service.Name()]; exists {
			return
		}

		apiTable := s.state.GetGlobal(globalApiTableKey)

		if s.boundServices == nil {
			s.boundServices = make(map[string]any)
		}

		s.boundServices[service.Name()] = service
		s.mux.Unlock()

		key := strings.ToLower(strings.ReplaceAll(service.Name(), " ", ""))

		s.Logger().Info("import successful", "target service", service.Name(), "table", fmt.Sprintf("api.%s", key))
		//go plugin.LuaPluginLoadIntoTable(L, apiTable.(*lua.LTable))
		s.state.SetField(apiTable, key, createLuaTableFromStruct(s.state, service))

		go s.WithState(func(L *lua.LState) error {
			if err := s.state.DoString(`
			function tree(obj, prefix)
				if not prefix then
					prefix = ""
				end
			
				for k,v in pairs(obj) do
					local valueType = type(v)
			
					if valueType == "table" then
						--print(prefix .. "." .. k)
						tree(v, prefix .. "." .. k)
					elseif valueType == "function" then
						print(prefix .. "." .. k .."()")
					end
				end
			end
			
			tree(api.` + key + ")"); err != nil {
				s.Logger().Error("foo", "error", err)
			}

			return nil
		})

		return
	}()
}

// WaitForGlobals blocks until all supplied globals exist
func (s *Service) WaitForGlobals(globals ...string) {
	eb := backoff.ExponentialBackOff{
		InitialInterval: time.Millisecond,
		Multiplier:      1.05,
		MaxInterval:     time.Second,
		MaxElapsedTime:  time.Second * 6,
	}

	for {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.WithState(func(L *lua.LState) error {
				if !s.globalsExist(L, globals...) {
					return fmt.Errorf("not found: %v", globals)
				}
				return nil
			})

			if err != nil {
				s.Logger().Warn("backoff", "errMsg", err)
				time.Sleep(eb.NextBackOff())
			}
		}()

		wg.Wait()

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

// sleep function that GopherLua can call using os.execute
func (s *Service) luaSleep(L *lua.LState) int {
	n := L.ToInt(1) // Get the number of seconds from the Lua script

	var cmd string
	switch runtime.GOOS {
	case "windows":
		cmd = fmt.Sprintf("timeout /T %d /NOBREAK", n)
	default: // Unix-like systems
		cmd = fmt.Sprintf("sleep %d", n)
	}

	result, err := runCommand(cmd)
	if err != nil {
		L.Push(lua.LString(result))
		return 1
	}

	return 0
}

func runCommand(cmd string) (string, error) {
	// Split the command into name and args
	command := exec.Command("sh", "-c", cmd)

	// Get the output of the command
	output, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
