package webServer

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/webRouter"
)

func (s *Service) DependenciesResolved() bool {
	if s.router == nil {
		return false
	}

	if s.config == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, other := range services {
		if router, ok := other.(webRouter.Dependency); ok {
			if router.RouteRoot() != nil {
				s.router = router
			}
		}
	}
}
