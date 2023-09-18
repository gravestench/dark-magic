package recordManager

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/tsvLoader"
)

func (s *Service) DependenciesResolved() bool {
	if s.cfg == nil {
		return false
	}

	if s.tsv == nil {
		return false
	}

	if !s.tsv.(runtime.HasDependencies).DependenciesResolved() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(runtime runtime.R) {
	for _, service := range runtime.Services() {
		switch candidate := service.(type) {
		case configFile.Dependency:
			s.cfg = candidate
		case tsvLoader.Dependency:
			s.tsv = candidate
		}
	}
}
