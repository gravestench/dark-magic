package webRouter

import (
	"github.com/gravestench/runtime"

	"github.com/gravestench/dark-magic/pkg/services/configFile"
)

func (s *Service) DependenciesResolved() bool {
	if s.cfgManager == nil {
		return false
	}

	return true
}

func (s *Service) ResolveDependencies(rt runtime.R) {
	for _, other := range rt.Services() {
		if cfg, ok := other.(configFile.Manager); ok {
			s.cfgManager = cfg
		}
	}
}
