package mapGenerator

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
)

func (s *Service) DependenciesResolved() bool {
	if s.assets == nil {
		return false
	}

	if s.records == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case assetLoader.Dependency:
			s.assets = candidate
		case recordManager.Dependency:
			s.records = candidate
		}
	}
}
