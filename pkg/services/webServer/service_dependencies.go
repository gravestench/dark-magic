package webServer

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/webRouter"
)

func (s *Service) DependenciesResolved() bool {
	if s.router == nil {
		return false
	}

	if s.cfgManager == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(mesh servicemesh.Mesh) {
	for _, other := range mesh.Services() {
		if router, ok := other.(webRouter.Dependency); ok {
			if router.RouteRoot() != nil {
				s.router = router
			}
		}

		if cfg, ok := other.(configFile.Manager); ok && s.cfgManager == nil {
			s.cfgManager = cfg
		}
	}
}
