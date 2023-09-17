package locale

import (
	"time"

	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/loaders/mpqLoader"
	"github.com/gravestench/dark-magic/pkg/services/loaders/tblLoader"
)

func (s *Service) DependenciesResolved() bool {
	if s.tbl == nil {
		return false
	}

	if s.mpq == nil {
		return false
	}

	if len(s.mpq.Archives()) != 11 {
		time.Sleep(time.Second)
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
