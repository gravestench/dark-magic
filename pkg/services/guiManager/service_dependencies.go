package guiManager

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/dc6Loader"
	"github.com/gravestench/dark-magic/pkg/services/input"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/raylibRenderer"
	"github.com/gravestench/dark-magic/pkg/services/spriteManager"
)

// the following methods are invoked by the servicemesh
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {
	if s.sprite == nil {
		return false
	}

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

	if s.mpq == nil {
		return false
	}

	if !s.mpq.RequiredArchivesLoaded() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(mesh servicemesh.Mesh) {
	for _, service := range mesh.Services() {
		switch candidate := service.(type) {
		case spriteManager.Dependency:
			s.sprite = candidate
		case raylibRenderer.Dependency:
			s.renderer = candidate
		case input.Dependency:
			s.input = candidate
		case dc6Loader.Dependency:
			s.dc6 = candidate
		case mpqLoader.Dependency:
			s.mpq = candidate
		}
	}
}
