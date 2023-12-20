package service_template

import (
	"github.com/gravestench/servicemesh"
)

// the following methods are invoked by the servicemesh
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {
	// NOTE: not everything here needs to be
	// services. We can check here if
	// files are loaded, if other service "ready"
	// methods yield true, etc.

	if s.bar == nil {
		return false
	}

	if s.baz == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(mesh servicemesh.Mesh) {
	type barDependency any // example
	type bazDependency any // example

	// it is up to us to iterate over our sibling services
	// and resolve our own dependencies.

	for _, service := range mesh.Services() {
		switch candidate := service.(type) {
		case barDependency:
			s.bar = candidate
		case bazDependency:
			s.baz = candidate
		}
	}
}
