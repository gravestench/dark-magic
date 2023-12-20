package tsvLoader

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

const (
	configKeySpriteCacheBudgetMB = "cache size (MB)"
)

var _ configFile.HasConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "loaders.json"
}

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	g := cfg.Group(s.Name())

	g.Set(configKeySpriteCacheBudgetMB, 50)

	return
}

func (s *Service) LoadConfig(config *configFile.Config) {
	s.config = config
}
