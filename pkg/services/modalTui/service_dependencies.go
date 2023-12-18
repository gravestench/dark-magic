package modalTui

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

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
