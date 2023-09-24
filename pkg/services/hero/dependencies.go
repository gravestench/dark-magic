package hero

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
)

// the following methods are invoked by the runtime
// automatically in an endless loop. As soon as the
// dependencies are resolved, the Init method is called.

func (s *Service) DependenciesResolved() bool {
	if s.config == nil {
		return false
	}

	if s.records == nil {
		return false
	}

	if !s.records.IsLoaded() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case configFile.Dependency:
			s.config = candidate
		case recordManager.Dependency:
			s.records = candidate
		}
	}
}
