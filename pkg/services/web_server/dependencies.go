package web_server

import (
	"github.com/gravestench/runtime/pkg"

	"github.com/gravestench/dark-magic/pkg/services/config_file"
	"github.com/gravestench/dark-magic/pkg/services/web_router"
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
		if router, ok := other.(web_router.Dependency); ok {
			if router.RouteRoot() != nil {
				s.router = router
			}
		}

		if cfg, ok := other.(config_file.Manager); ok && s.cfgManager == nil {
			s.cfgManager = cfg
		}
	}
}
