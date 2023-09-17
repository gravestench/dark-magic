package record_manager

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/config_file"
	"github.com/gravestench/dark-magic/pkg/services/loaders/tsvLoader"
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
		case config_file.Dependency:
			s.cfg = candidate
		case tsvLoader.Dependency:
			s.tsv = candidate
		}
	}
}
