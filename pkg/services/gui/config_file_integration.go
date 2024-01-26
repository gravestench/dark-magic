package gui

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var _ configFile.HasConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "gui.json"
}

const (
	groupGUI = "GUI"
	keyFPS
)

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	for group, kv := range map[string]map[string]any{
		groupGUI: {
			keyFPS: 24,
		},
	} {
		for key, val := range kv {
			cfg.Group(group).Set(key, val)
		}
	}

	return cfg
}

func (s *Service) LoadConfig(config *configFile.Config) {
	s.config = config
}
