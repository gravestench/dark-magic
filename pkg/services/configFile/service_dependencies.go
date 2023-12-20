package configFile

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/fileWatcher"
)

func (s *Service) DependenciesResolved() bool {
	if s.fileWatcher == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case fileWatcher.Dependency:
			s.fileWatcher = candidate
		}
	}
}
