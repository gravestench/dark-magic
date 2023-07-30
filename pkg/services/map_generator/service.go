package map_generator

import (
	"github.com/gravestench/runtime"
	"github.com/rs/zerolog"

	"github.com/gravestench/dark-magic/pkg/models"
	"github.com/gravestench/dark-magic/pkg/services/d2_asset_loader"
)

type Service struct {
	logger *zerolog.Logger
	assets d2_asset_loader.Dependency

	seed uint64
}

func (s *Service) BindLogger(logger *zerolog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *zerolog.Logger {
	return s.logger
}

func (s *Service) Init(rt runtime.R) {

}

func (s *Service) Name() string {
	return "Diablo2Map Generator"
}

func (s *Service) DependenciesResolved() bool {
	if s.assets == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		if candidate, ok := service.(d2_asset_loader.Dependency); ok {
			s.assets = candidate
		}
	}
}

func (s *Service) Seed() uint64 {
	return s.seed
}

func (s *Service) SetSeed(u uint64) {
	s.seed = u
}

func (s *Service) GenerateMap(act, difficulty uint) (models.Diablo2Map, error) {
	return models.Diablo2Map{}, nil
}
