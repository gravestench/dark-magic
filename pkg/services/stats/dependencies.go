package stats

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/locale"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
)

func (s *Service) DependenciesResolved() bool {
	if s.records == nil {
		return false
	}

	if s.tables == nil {
		return false
	}

	if s.locale == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.Runtime) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case recordManager.Dependency:
			s.records = candidate
		case tblLoader.Dependency:
			s.tables = candidate
		case locale.Dependency:
			s.locale = candidate
		}
	}
}
