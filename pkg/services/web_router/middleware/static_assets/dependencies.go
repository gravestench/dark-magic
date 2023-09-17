package static_assets

import (
	"github.com/gravestench/runtime"
)

func (m *Middleware) DependenciesResolved() bool {
	if m.router == nil {
		return false
	}

	return true
}

func (m *Middleware) ResolveDependencies(rt runtime.R) {
	for _, candidate := range rt.Services() {
		if router, ok := candidate.(IsWebRouter); ok {
			if router.RouteRoot() == nil {
				return
			}

			m.router = router
		}
	}
}
