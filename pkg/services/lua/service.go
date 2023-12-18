package lua

import (
	"log/slog"
	"sync"

	ee "github.com/gravestench/eventemitter"
	"github.com/gravestench/servicemesh"
	"github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	cfg           configFile.Dependency
	logger        *slog.Logger
	state         *lua.LState
	events        *ee.EventEmitter
	mux           sync.Mutex
	boundServices map[string]any
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	s.boundServices = make(map[string]any)
	s.state = lua.NewState()
	s.bindLoggerToLuaEnvironment()

	mesh.Events().On(servicemesh.EventServiceAdded, s.tryToExportToLuaEnvironment)

	for _, service := range mesh.Services() {
		if candidate, ok := service.(servicemesh.HasDependencies); ok {
			if !candidate.DependenciesResolved() {
				continue
			}
		}

		if candidate, ok := service.(UsesLuaEnvironment); ok {
			s.tryToExportToLuaEnvironment(candidate)
		}
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

	cfg, err := s.cfg.GetConfigByFileName(s.ConfigFileName())
	if err != nil {
		s.logger.Error("getting config", "error", err)
		panic(err)
	}

	// TODO :: race condition where this script inits before other services
	//   have a chance to export their stuff to the lua state machine
	initScriptPath := cfg.Group(s.Name()).GetString("init script")

	if err = s.runScript(initScriptPath); err != nil {
		s.logger.Error("running script", "error", err)
	}
}

func (s *Service) Name() string {
	return "Lua Environment"
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
