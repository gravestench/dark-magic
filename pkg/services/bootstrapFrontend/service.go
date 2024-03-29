package bootstrap_frontend

import (
	"log/slog"

	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/bootstrapBackend"
	"github.com/gravestench/dark-magic/pkg/services/gui"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

type Service struct {
	logger  *slog.Logger
	backend bootstrap_backend.Dependency

	renderer   raylibRenderer.Service
	input      input.Service
	guiManager gui.Service
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
	mesh.Add(&s.renderer)
	mesh.Add(&s.input)
	mesh.Add(&s.guiManager)
}

func (s *Service) Name() string {
	return "Frontend Bootstrap"
}

func (s *Service) Ready() bool {
	return true
}

func (s *Service) FrontendReady() bool {
	return s.Ready()
}

func (s *Service) SetLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Logger() *slog.Logger {
	return s.logger
}
