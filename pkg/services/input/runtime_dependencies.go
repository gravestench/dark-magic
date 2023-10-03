package input

import (
	"github.com/gravestench/runtime"
)

// the following methods are invoked by the runtime
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {
	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case barDependency:
			s.bar = candidate
		case bazDependency:
			s.baz = candidate
		}
	}
}
