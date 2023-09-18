package locale

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/tblLoader"
)

func (s *Service) DependenciesResolved() bool {
	if s.tbl == nil {
		return false
	}

	if s.mpq == nil {
		return false
	}

	if !s.mpq.RequiredArchivesLoaded() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.Runtime) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case tblLoader.Dependency:
			s.tbl = candidate
		case mpqLoader.Dependency:
			s.mpq = candidate
		}
	}
}
