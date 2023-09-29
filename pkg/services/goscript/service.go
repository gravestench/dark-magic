package goscript

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"
)

type Service struct {
	cfg    configFile.Dependency
	logger *zerolog.Logger
}

func (s *Service) Init(rt runtime.Runtime) {
	cfg, err := s.cfg.GetConfigByFileName(s.ConfigFileName())
	if err != nil {
		s.logger.Fatal().Msgf("getting config: %v", err)
	}

	initScriptPath := cfg.Group(s.Name()).GetString("init script")

	s.runScript(initScriptPath)

	// TODO: add file watcher for re-running init script.
}

func (s *Service) Name() string {
	return "Goscript"
}

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}
