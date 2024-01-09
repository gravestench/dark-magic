package raylibRenderer

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var _ configFile.HasConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "renderer.json"
}

const (
	groupKeyWindow      = "Window"
	keyWindowTitle      = "title"
	keyWindowWidth      = "width"
	keyWindowHeight     = "height"
	keyWindowFullscreen = "fullscreen"
	keyMonitor          = "monitor index"

	groupKeyResolution  = "Resolution"
	keyResolutionWidth  = "width"
	keyResolutionHeight = "height"

	groupKeyCache  = "Cache"
	keyCacheBudget = "budget (B)"
)

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	for group, kv := range map[string]map[string]any{
		groupKeyWindow: {
			keyWindowTitle:      "Dark Magic",
			keyWindowWidth:      1024,
			keyWindowHeight:     768,
			keyWindowFullscreen: false,
			keyMonitor:          0,
		},
		groupKeyResolution: {
			keyResolutionWidth:  1024,
			keyResolutionHeight: 768,
		},
		groupKeyCache: {
			keyCacheBudget: 100,
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
