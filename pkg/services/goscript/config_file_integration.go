package goscript

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var _ configFile.HasDefaultConfig = &Service{}

var _ configFile.HasConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "goscript_environment.json"
}

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	g := cfg.Group(s.Name())

	g.SetDefault("init script", "init.go")

	return
}

func (s *Service) LoadConfig(config *configFile.Config) {
	s.config = config
}
