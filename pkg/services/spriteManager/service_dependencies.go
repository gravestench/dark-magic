package spriteManager

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/assetLoader"
)

// the following methods are invoked by the servicemesh
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {
	if s.assets == nil {
		return false
	}

	if s.config == nil {
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
