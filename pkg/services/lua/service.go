package lua

import (
	"sync"

	ee "github.com/gravestench/eventemitter"
	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/pkg/events"
	"github.com/rs/zerolog"
	"github.com/yuin/gopher-lua"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	cfg           configFile.Dependency
	logger        *zerolog.Logger
	state         *lua.LState
	events        *ee.EventEmitter
	mux           sync.Mutex
	boundServices map[string]any
}

func (s *Service) Init(rt runtime.R) {
	s.boundServices = make(map[string]any)
	s.state = lua.NewState()
	s.bindLoggerToLuaEnvironment()

	rt.Events().On(events.EventServiceAdded, s.tryToExportToLuaEnvironment)

	for _, service := range rt.Services() {
		if candidate, ok := service.(runtime.HasDependencies); ok {
			if !candidate.DependenciesResolved() {
				continue
			}
		}

		if candidate, ok := service.(UsesLuaEnvironment); ok {
			s.tryToExportToLuaEnvironment(candidate)
		}
	}

	// wait for all siblings to be ready before we launch scripts
	for _, service := range rt.Services() {
		if candidate, ok := service.(runtime.HasDependencies); ok {
			if !candidate.DependenciesResolved() {
				continue
			}
		}

		break // all deps resolved for all siblings
	}

	cfg, err := s.cfg.GetConfigByFileName(s.ConfigFileName())
	if err != nil {
		s.logger.Fatal().Msgf("getting config: %v", err)
	}

	// TODO :: race condition where this script inits before other services
	//   have a chance to export their stuff to the lua state machine
	initScriptPath := cfg.Group(s.Name()).GetString("init script")

	if err = s.runScript(initScriptPath); err != nil {
		s.logger.Error().Msgf("running script: %v", err)
	}
}

func (s *Service) Name() string {
	return "Lua Environment"
}

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}
