package raylibRenderer

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

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
)

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	for key, val := range map[string]any{
		keyWindowTitle:      "Dark Magic",
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

	return cfg
}
