package raylibRenderer

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

// the following methods are invoked by the runtime
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {
	if s.cfg == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case configFile.Dependency:
			s.cfg = candidate
		}
	}
}
