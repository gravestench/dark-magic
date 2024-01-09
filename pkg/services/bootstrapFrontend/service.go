package bootstrap_frontend

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/bootstrapBackend"
	"github.com/gravestench/dark-magic/pkg/services/guiManager"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/modalGameUI"
	"github.com/gravestench/dark-magic/pkg/services/modalGameUI/ui/trademark"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type Service struct {
	logger  *slog.Logger
	backend bootstrap_backend.Dependency
}

func (s *Service) DependenciesResolved() bool {
	if s.backend == nil {
		return false
	}

	if s.backend.BackendReady() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case bootstrap_backend.Dependency:
			s.backend = candidate
		}
	}
}

func (s *Service) Init(mesh servicemesh.Mesh) {
	// rendering-dependant services
	mesh.Add(&raylibRenderer.Service{})
	mesh.Add(&input.Service{}) // rendering backend also handles input
	//mesh.Add(&backgroundMusic.Service{}) // rendering backend also handles audio

	mesh.Add(&guiManager.Service{})
	mesh.Add(&modalGameUI.Service{})
	//mesh.Add(&loading.Screen{})
	mesh.Add(&trademark.Screen{})
	//mesh.Add(&tilesprite_test.Screen{})
}

func (s *Service) Name() string {
	return "Frontend Bootstrap"
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
