package static_assets

import (
	"github.com/gravestench/servicemesh"
)

func (m *Middleware) DependenciesResolved() bool {
	if m.router == nil {
		return false
	}

	return true
}

func (m *Middleware) ResolveDependencies(services []servicemesh.Service) {
	for _, candidate := range services {
		if router, ok := candidate.(IsWebRouter); ok {
			if router.RouteRoot() == nil {
				return
			}

			m.router = router
		}
	}
}
