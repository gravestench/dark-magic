package recordManager

import (
	"github.com/gravestench/servicemesh"

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

	if !s.tsv.(servicemesh.HasDependencies).DependenciesResolved() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case configFile.Dependency:
			s.cfg = candidate
		case tsvLoader.Dependency:
			s.tsv = candidate
		}
	}
}
