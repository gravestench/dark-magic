package mpqLoader

import (
	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

var _ configFile.HasConfig = &Service{}

func (s *Service) ConfigFileName() string {
	return "loaders.json"
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

func (s *Service) LoadConfig(config *configFile.Config) {
	s.config = config
}
