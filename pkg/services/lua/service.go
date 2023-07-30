package lua

import (
	ee "github.com/gravestench/eventemitter"
	"github.com/gravestench/runtime"
	"github.com/gravestench/runtime/examples/services/config_file"
	"github.com/gravestench/runtime/pkg/events"
	"github.com/rs/zerolog"
	"github.com/yuin/gopher-lua"
)

type Service struct {
	cfg    config_file.Manager
	logger *zerolog.Logger
	state  *lua.LState
	events *ee.EventEmitter
}

func (s *Service) Init(rt runtime.R) {
	s.state = lua.NewState()
	s.bindLoggerToLuaEnvironment()

	s.events.On(events.EventServiceAdded, s.tryToExportToLuaEnvironment)

	for _, service := range rt.Services() {
		if candidate, ok := service.(UsesLuaEnvironment); ok {
			s.tryToExportToLuaEnvironment(candidate)
		}
	}

	cfg, err := s.Config()
	if err != nil {
		s.logger.Fatal().Msgf("getting config: %v", err)
	}

	initScript := cfg.Group(s.Name()).GetString("init script")

	if err = s.state.DoFile(initScript); err != nil {
		s.logger.Warn().Msgf("executing init script '%s': %v", initScript, err)
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
