package service_template

import (
	"github.com/gravestench/runtime"
)

// the following methods are invoked by the runtime
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {
	// NOTE: not everything here needs to be
	// runtime services. We can check here if
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

func (s *Service) ResolveDependencies(rt runtime.R) {
	type barDependency any // example
	type bazDependency any // example

	// it is up to us to iterate over our sibling services
	// and resolve our own dependencies.

	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case barDependency:
			s.bar = candidate
		case bazDependency:
			s.baz = candidate
		}
	}
}
