package mpqLoader

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

func (s *Service) DependenciesResolved() bool {
	if s.cfgManager == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(mesh servicemesh.Mesh) {
	for _, service := range mesh.Services() {
		if candidate, ok := service.(configFile.Manager); ok {
			s.cfgManager = candidate
		}
	}
}
