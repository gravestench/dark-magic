package lua

import (
	"fmt"

	"github.com/gravestench/runtime/examples/services/config_file"
)

var _ config_file.HasDefaultConfig = &Service{}

func (s *Service) ConfigFilePath() string {
	return "lua_environment.json"
}

func (s *Service) Config() (*config_file.Config, error) {
	if s.cfg == nil {
		return nil, fmt.Errorf("config manager is nil")
	}

	return s.cfg.GetConfig(s.ConfigFilePath())
}

func (s *Service) DefaultConfig() (cfg config_file.Config) {
	g := cfg.Group(s.Name())

	g.SetDefault("init script", "init.lua")

	return
}
