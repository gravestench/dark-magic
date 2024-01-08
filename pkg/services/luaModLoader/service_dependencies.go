package luaModLoader

import (
	"github.com/gravestench/servicemesh"

	"github.com/gravestench/dark-magic/pkg/services/fileLoader"
	"github.com/gravestench/dark-magic/pkg/services/luaManager"
)

func (s *Service) DependenciesResolved() bool {
	if s.Config == nil {
		return false
	}

	if s.loader == nil {
		return false
	}

	if s.lua == nil {
		return false
	}

	if !s.lua.Ready() {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(services []servicemesh.Service) {
	for _, service := range services {
		switch candidate := service.(type) {
		case luaManager.Dependency:
			s.lua = candidate
		case fileLoader.Dependency:
			s.loader = candidate
		}
	}
}
