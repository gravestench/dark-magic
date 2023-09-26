package spriteManager

import (
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/dccLoader"
	"github.com/gravestench/dark-magic/pkg/services/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
)

type Service struct {
	logger *zerolog.Logger
	config configFile.Dependency

	mpq mpqLoader.Dependency
	pl2 pl2Loader.Dependency
	dc6 dc6Loader.Dependency
	dcc dccLoader.Dependency
	dt1 dt1Loader.Dependency
	ds1 ds1Loader.Dependency

	spriteCache *cache.Cache
}

func (s *Service) Init(rt runtime.R) {
	s.initSpriteCache()
}

func (s *Service) initSpriteCache() {
	const (
		kb = 1024
		mb = 1024 * kb
	)

	cfg, err := s.config.GetConfigByFileName(s.ConfigFileName())
	if err != nil {
		s.spriteCache = cache.New(500 * mb)
		return
	}

	budget := cfg.Group(s.Name()).GetInt(configKeySpriteCacheBudgetMB)
	if budget == 0 {
		budget = 500
	}

	s.spriteCache = cache.New(budget * mb)
}

func (s *Service) Name() string {
	return "Sprite Manager"
}

// the following methods are boilerplate, but they are used
// by the runtime to enforce a standard logging format.

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}
