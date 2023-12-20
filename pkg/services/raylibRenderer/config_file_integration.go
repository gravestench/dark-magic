package raylibRenderer

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var _ configFile.HasConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "raylib_renderer.json"
}

const (
	groupKeyWindow      = "Window"
	keyWindowTitle      = "title"
	keyWindowWidth      = "width"
	keyWindowHeight     = "height"
	keyWindowFullscreen = "fullscreen"

	groupKeyResolution  = "Resolution"
	keyResolutionWidth  = "width"
	keyResolutionHeight = "height"

	groupKeyCache  = "Cache"
	keyCacheBudget = "budget (B)"
)

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	for key, val := range map[string]any{
		keyWindowTitle:      "MTG",
		keyWindowWidth:      800,
		keyWindowHeight:     600,
		keyWindowFullscreen: false,
	} {
		cfg.Group(groupKeyWindow).Set(key, val)
	}

	for key, val := range map[string]any{
		keyResolutionWidth:  800,
		keyResolutionHeight: 600,
	} {
		cfg.Group(groupKeyResolution).Set(key, val)
	}

	for key, val := range map[string]any{
		keyCacheBudget: 100,
	} {
		cfg.Group(groupKeyCache).Set(key, val)
	}

	return cfg
}

func (s *Service) LoadConfig(config *configFile.Config) {
	s.config = config
}
