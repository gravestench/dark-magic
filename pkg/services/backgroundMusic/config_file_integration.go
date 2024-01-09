package backgroundMusic

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var _ configFile.HasConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "audio.json"
}

const (
	groupBGM = "Background Music"
	keyBGMVolume
)

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	for group, kv := range map[string]map[string]any{
		groupBGM: {
			keyBGMVolume: 0.5,
		},
	} {
		for key, val := range kv {
			cfg.Group(group).Set(key, val)
		}
	}

	return cfg
}

func (s *Service) LoadConfig(config *configFile.Config) {
	volume := config.Group(groupBGM).GetFloat(keyBGMVolume)
	if volume < 0 || volume > 1.0 {
		volume = 0.5
	}

	s.config = config
}
