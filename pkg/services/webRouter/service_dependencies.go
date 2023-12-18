package webRouter

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
	for _, other := range mesh.Services() {
		if cfg, ok := other.(configFile.Manager); ok {
			s.cfgManager = cfg
		}
	}
}
