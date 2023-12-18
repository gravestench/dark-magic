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

func (m *Middleware) ResolveDependencies(mesh servicemesh.Mesh) {
	for _, candidate := range mesh.Services() {
		if router, ok := candidate.(IsWebRouter); ok {
			if router.RouteRoot() == nil {
				return
			}

			m.router = router
		}
	}
}
