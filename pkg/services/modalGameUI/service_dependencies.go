package modalGameUI

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/pl2Loader"
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

	if s.input == nil {
		return false
	}

	if s.dc6 == nil {
		return false
	}

	if s.pl2 == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(mesh servicemesh.Mesh) {
	for _, service := range mesh.Services() {
		switch candidate := service.(type) {
		case raylibRenderer.Dependency:
			s.renderer = candidate
		case input.Dependency:
			s.input = candidate
		case dc6Loader.Dependency:
			s.dc6 = candidate
		case pl2Loader.Dependency:
			s.pl2 = candidate
		}
	}
}
