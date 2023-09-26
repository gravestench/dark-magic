package tsvLoader

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
)

func (s *Service) DependenciesResolved() bool {
	if s.mpq == nil {
		return false
	}

	if !s.mpq.RequiredArchivesLoaded() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, service := range rt.Services() {
		switch candidate := service.(type) {
		case mpqLoader.Dependency:
			s.mpq = candidate
		}
	}
}
