package tblLoader

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/cache"
	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
)

type Service struct {
	logger *slog.Logger
	mpq    mpqLoader.Dependency
	config *configFile.Config
	cache  *cache.Cache
}

func (s *Service) Init(mesh servicemesh.Mesh) {
}

func (s *Service) Name() string {
	return "TBL Loader"
}

func (s *Service) Ready() bool {
	if s.mpq == nil {
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

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
