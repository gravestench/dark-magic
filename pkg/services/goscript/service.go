package goscript

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	cfg    configFile.Dependency
	logger *slog.Logger
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	cfg, err := s.cfg.GetConfigByFileName(s.ConfigFileName())
	if err != nil {
		s.logger.Error("getting config", "error", err)
		panic(err)
	}

	initScriptPath := cfg.Group(s.Name()).GetString("init script")

	s.runScript(initScriptPath)

	// TODO: add file watcher for re-running init script.
}

func (s *Service) Name() string {
	return "Goscript"
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
