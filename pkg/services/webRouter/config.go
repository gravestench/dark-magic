package webRouter

import (
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

func (s *Service) ConfigFileName() string {
	return "web_router.json"
}

func (s *Service) Config() (*configFile.Config, error) {
	if s.cfgManager == nil {
		return nil, fmt.Errorf("no config manager service bound")
	}

	return s.cfgManager.GetConfigByFileName(s.ConfigFileName())
}

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	g := cfg.Group("Gin Route Handler")

	g.SetDefault("debug", true)

	return
}
