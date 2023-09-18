package web_router

import (
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/config_file"
)

func (s *Service) ConfigFileName() string {
	return "web_router.json"
}

func (s *Service) Config() (*config_file.Config, error) {
	if s.cfgManager == nil {
		return nil, fmt.Errorf("no config manager service bound")
	}

	return s.cfgManager.GetConfigByFileName(s.ConfigFileName())
}

func (s *Service) DefaultConfig() (cfg config_file.Config) {
	g := cfg.Group("Gin Route Handler")

	g.SetDefault("debug", true)

	return
}
