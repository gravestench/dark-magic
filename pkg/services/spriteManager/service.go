package spriteManager

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

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
	logger *slog.Logger
	config *configFile.Config

	mpq mpqLoader.Dependency
	pl2 pl2Loader.Dependency
	dc6 dc6Loader.Dependency
	dcc dccLoader.Dependency
	dt1 dt1Loader.Dependency
	ds1 ds1Loader.Dependency

	spriteCache *cache.Cache
}

func (s *Service) Init(mesh servicemesh.Mesh) {

}

func (s *Service) Name() string {
	return "Sprite Manager"
}

func (s *Service) Ready() bool {
	if s.mpq == nil {
		return false
	}

	if s.pl2 == nil {
		return false
	}

	if s.dc6 == nil {
		return false
	}

	if s.dcc == nil {
		return false
	}

	if s.dt1 == nil {
		return false
	}

	if s.ds1 == nil {
		return false
	}

	if s.config == nil {
		return false
	}

	if !s.mpq.RequiredArchivesLoaded() {
		return false
	}

	return true
}

// the following methods are boilerplate, but they are used
// by the servicemesh to enforce a standard logging format.

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
