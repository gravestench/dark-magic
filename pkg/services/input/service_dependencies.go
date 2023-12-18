package input

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
)

// the following methods are invoked by the servicemesh
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {
	if s.renderer == nil {
		return false
	}

	if !s.renderer.IsInit() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(mesh servicemesh.Mesh) {
	for _, service := range mesh.Services() {
		switch candidate := service.(type) {
		case raylibRenderer.Dependency:
			s.renderer = candidate
		}
	}
}
