package goscript

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var _ configFile.HasDefaultConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "goscript_environment.json"
}

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	g := cfg.Group(s.Name())

	g.SetDefault("init script", "init.go")

	return
}
