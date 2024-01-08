package recordManager

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
)

func (s *Service) DependenciesResolved() bool {
	if s.assets == nil {
		return false
	}

	if !s.assets.(servicemesh.HasDependencies).DependenciesResolved() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case assetLoader.Dependency:
			s.assets = candidate
		}
	}
}
