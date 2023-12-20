package webRouter

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var _ configFile.HasConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "web_router.json"
}

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	g := cfg.Group("Gin Route Handler")

	g.SetDefault("debug", true)

	return
}

func (s *Service) LoadConfig(config *configFile.Config) {
	s.config = config
}
