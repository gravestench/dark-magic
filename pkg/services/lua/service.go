package lua

import (
	"log/slog"
	"sync"
	"time"

	ee "github.com/gravestench/eventemitter"
	"github.com/gravestench/servicemesh"
	"github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	config        *configFile.Config
	logger        *slog.Logger
	state         *lua.LState
	events        *ee.EventEmitter
	mux           sync.Mutex
	boundServices map[string]any
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	if s.boundServices == nil {
		s.boundServices = make(map[string]any)
	}

	s.state = lua.NewState()
	s.bindLoggerToLuaEnvironment()

	for _, service := range mesh.Services() {
		s.tryToExportToLuaEnvironment(service)
	}

	// wait for all siblings to be ready before we launch scripts
	for _, service := range mesh.Services() {
		if candidate, ok := service.(servicemesh.HasDependencies); ok {
			if !candidate.DependenciesResolved() {
				continue
			}
		}

		break // all deps resolved for all siblings
	}

	for {
		// TODO :: race condition where this script inits before other services
		//   have a chance to export their stuff to the lua state machine
		initScriptPath := s.config.Group(s.Name()).GetString("init script")

		if err := s.runScript(initScriptPath); err != nil {
			s.logger.With("script", initScriptPath).Error("running script", "error", err)
			time.Sleep(time.Second * 1)
			continue
		}

		break
	}
}

func (s *Service) Name() string {
	return "Lua Environment"
}

func (s *Service) Ready() bool {
	if s.config == nil {
		return false
	}

	if s.state == nil {
		return false
	}

	return true
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
