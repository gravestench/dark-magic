package cofLoader

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
)

type Service struct {
	logger *slog.Logger
	mpq    mpqLoader.Dependency
}

func (s *Service) DependenciesResolved() bool {
	if s.mpq == nil {
		return false
	}

	if !s.mpq.RequiredArchivesLoaded() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case mpqLoader.Dependency:
			s.mpq = candidate
		}
	}
}

func (s *Service) Init(mesh servicemesh.Mesh) {

}

func (s *Service) Name() string {
	return "COF Loader"
}

func (s *Service) Ready() bool {
	for _, dependency := range []any{
		s.mpq,
	} {
		if dependency == nil {
			return false
		}
	}

	return true
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
