package webServer

import (
	"github.com/gravestench/runtime/pkg"

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

func (s *Service) ResolveDependencies(runtime pkg.IsRuntime) {
	for _, other := range runtime.Services() {
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
