package mpqLoader

import (
	"fmt"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

func (s *Service) ConfigFileName() string {
	return "loaders.json"
}

func (s *Service) Config() (*configFile.Config, error) {
	if s.cfgManager == nil {
		return nil, fmt.Errorf("config manager is nil")
	}

	return s.cfgManager.GetConfigByFileName(s.ConfigFileName())
}

func (s *Service) DefaultConfig() (cfg configFile.Config) {
	g := cfg.Group(s.Name())

	g.SetDefault("load order", []string{
		"patch_d2.mpq",
		"d2exp.mpq",
		"d2xmusic.mpq",
		"d2xtalk.mpq",
		"d2xvideo.mpq",
		"d2data.mpq",
		"d2char.mpq",
		"d2music.mpq",
		"d2sfx.mpq",
		"d2video.mpq",
		"d2speech.mpq",
	})

	g.SetDefault("directory", "~/mpq/")

	return
}
