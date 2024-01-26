package bootstrap_game_client

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/screens/trademark"
	"github.com/gravestench/dark-magic/pkg/services/bootstrapBackend"
	"github.com/gravestench/dark-magic/pkg/services/bootstrapFrontend"
	"github.com/gravestench/dark-magic/pkg/services/screenManager"
)

type Service struct {
	logger   *slog.Logger
	backend  bootstrap_backend.Dependency
	frontend bootstrap_frontend.Dependency
}

func (s *Service) DependenciesResolved() bool {
	if s.backend == nil {
		return false
	}

	if s.frontend == nil {
		return false
	}

	if !s.backend.BackendReady() {
		return false
	}

	if !s.frontend.FrontendReady() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case bootstrap_backend.Dependency:
			s.backend = candidate
		case bootstrap_frontend.Dependency:
			s.frontend = candidate
		}
	}
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	mesh.Add(&screenManager.Service{})
	//mesh.Add(&loading.Screen{})
	mesh.Add(&trademark.Screen{})
	//mesh.Add(&tilesprite_test.Screen{})
}

func (s *Service) Name() string {
	return "Game Client Bootstrap"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
