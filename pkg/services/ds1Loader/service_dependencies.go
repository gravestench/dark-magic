package ds1Loader

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
	"github.com/gravestench/dark-magic/pkg/services/mpqLoader"
)

func (s *Service) DependenciesResolved() bool {
	if s.mpq == nil {
		return false
	}

	if !s.mpq.RequiredArchivesLoaded() {
		return false
	}

	if s.config == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(mesh servicemesh.Mesh) {
	for _, service := range mesh.Services() {
		switch candidate := service.(type) {
		case mpqLoader.Dependency:
			s.mpq = candidate
		case configFile.Dependency:
			s.config = candidate
		}
	}
}
