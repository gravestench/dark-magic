package lua

import (
	"github.com/gravestench/dark-magic/pkg/services/config_file"
)

var _ config_file.HasDefaultConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "lua_environment.json"
}

func (s *Service) DefaultConfig() (cfg config_file.Config) {
	g := cfg.Group(s.Name())

	g.SetDefault("init script", "init.lua")

	return
}
