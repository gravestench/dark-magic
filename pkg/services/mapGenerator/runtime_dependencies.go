package mapGenerator

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/ds1Loader"
	"github.com/gravestench/dark-magic/pkg/services/dt1Loader"
	"github.com/gravestench/dark-magic/pkg/services/recordManager"
)

func (s *Service) DependenciesResolved() bool {
	if s.dt1 == nil {
		return false
	}

	if s.ds1 == nil {
		return false
	}

	if s.records == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case dt1Loader.Dependency:
			s.dt1 = candidate
		case ds1Loader.Dependency:
			s.ds1 = candidate
		case recordManager.Dependency:
			s.records = candidate
		}
	}
}
