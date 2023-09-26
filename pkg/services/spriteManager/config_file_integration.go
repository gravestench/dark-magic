package spriteManager

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

const (
	configKeySpriteCacheBudgetMB = "sprite cache size (MB)"
)

func (s *Service) ConfigFileName() string {
	return "sprite_manager.json"
}

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	g := cfg.Group(s.Name())

	g.Set(configKeySpriteCacheBudgetMB, 500)

	return
}
