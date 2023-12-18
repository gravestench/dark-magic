package raylibRenderer

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

// the following methods are invoked by the servicemesh
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {
	if s.cfg == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(mesh servicemesh.Mesh) {
	for _, service := range mesh.Services() {
		switch candidate := service.(type) {
		case configFile.Dependency:
			s.cfg = candidate
		}
	}
}
