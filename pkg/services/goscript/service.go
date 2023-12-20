package goscript

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

type Service struct {
	config *configFile.Config
	logger *slog.Logger
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	initScriptPath := s.config.Group(s.Name()).GetString("init script")

	s.runScript(initScriptPath)

	// TODO: add file watcher for re-running init script.
}

func (s *Service) Name() string {
	return "Goscript"
}

func (s *Service) Ready() bool {
	if s.config == nil {
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
