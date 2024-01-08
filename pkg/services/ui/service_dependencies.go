package ui

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

func (s *Service) DependenciesResolved() bool {
	if s.renderer == nil {
		return false
	}

	if !s.renderer.IsInit() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case raylibRenderer.Dependency:
			s.renderer = candidate
		}
	}
}
